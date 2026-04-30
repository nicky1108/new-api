#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${APP_DIR:-/opt/new-api}"
PORT="${NEW_API_PORT:-3000}"
TARGET_CLIENT_IMAGE="${POSTGRES_TARGET_CLIENT_IMAGE:-postgres:18-alpine}"

if [ "${1:-}" = "--dsn-from-stdin" ]; then
  IFS= read -r RDS_SQL_DSN
fi

: "${RDS_SQL_DSN:?RDS_SQL_DSN is required}"

cd "$APP_DIR"

if [ ! -f .env ]; then
  echo "Missing $APP_DIR/.env"
  exit 1
fi

env_value() {
  local key="$1"
  awk -F= -v key="$key" '$1 == key { sub(/^[^=]*=/, ""); print; exit }' .env
}

POSTGRES_USER="$(env_value POSTGRES_USER)"
POSTGRES_DB="$(env_value POSTGRES_DB)"
POSTGRES_PASSWORD="$(env_value POSTGRES_PASSWORD)"

: "${POSTGRES_USER:?POSTGRES_USER is required in .env}"
: "${POSTGRES_DB:?POSTGRES_DB is required in .env}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required in .env}"

timestamp="$(date +%Y%m%d%H%M%S)"
backup_dir="$APP_DIR/backups/rds-migration-$timestamp"
env_backup="$backup_dir/env-before-rds"
cutover_confirmed=0

rollback_on_error() {
  local exit_code=$?
  if [ "$cutover_confirmed" -ne 1 ]; then
    echo "Migration failed before confirmed cutover; restoring application to the original database."
    if [ -f "$env_backup" ]; then
      cp -a "$env_backup" .env
    fi
    docker compose up -d new-api || true
  fi
  exit "$exit_code"
}

trap rollback_on_error ERR

mkdir -p "$backup_dir"
chmod 700 "$backup_dir"
cp -a .env "$env_backup"

query=""
base_dsn="$RDS_SQL_DSN"
if [[ "$base_dsn" == *\?* ]]; then
  query="?${base_dsn#*\?}"
  base_dsn="${base_dsn%%\?*}"
fi

target_db="${base_dsn##*/}"
admin_dsn="${base_dsn%/*}/postgres${query}"

if [ -z "$target_db" ] || [ "$target_db" = "$base_dsn" ]; then
  echo "Unable to parse target database from RDS_SQL_DSN"
  exit 1
fi

if [[ ! "$target_db" =~ ^[A-Za-z0-9_]+$ ]]; then
  echo "Target database name contains unsupported characters: $target_db"
  exit 1
fi

echo "Checking source PostgreSQL and target RDS connectivity."
docker compose exec -T postgres pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null

docker run --rm \
  -e ADMIN_DSN="$admin_dsn" \
  -e TARGET_DB="$target_db" \
  "$TARGET_CLIENT_IMAGE" \
  sh -eu -c '
    exists="$(psql "$ADMIN_DSN" -At -v ON_ERROR_STOP=1 -c "SELECT 1 FROM pg_database WHERE datname = '\''$TARGET_DB'\'';")"
    if [ "$exists" != "1" ]; then
      psql "$ADMIN_DSN" -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"$TARGET_DB\";"
    fi
  '

docker run --rm \
  -e RDS_SQL_DSN="$RDS_SQL_DSN" \
  "$TARGET_CLIENT_IMAGE" \
  psql "$RDS_SQL_DSN" -v ON_ERROR_STOP=1 -c "select 1" >/dev/null

echo "Backing up the current target RDS database before replacement."
docker run --rm \
  -e RDS_SQL_DSN="$RDS_SQL_DSN" \
  -v "$backup_dir:/backup" \
  "$TARGET_CLIENT_IMAGE" \
  pg_dump --format=custom --no-owner --no-privileges --file=/backup/target-before-rds.dump "$RDS_SQL_DSN"

echo "Stopping application writes for the final consistent source dump."
docker compose stop new-api

echo "Writing exact source table counts."
docker compose exec -T postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At -F $'\t' <<'SQL' \
  | awk -F '\t' '$1 ~ /^public[.]/ { print }' > "$backup_dir/source-counts.tsv"
SELECT format('SELECT %L, count(*) FROM %I.%I;', schemaname || '.' || tablename, schemaname, tablename)
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY schemaname, tablename;
\gexec
SQL

echo "Creating final source dump."
docker compose exec -T postgres pg_dump \
  -U "$POSTGRES_USER" \
  -d "$POSTGRES_DB" \
  --format=custom \
  --no-owner \
  --no-privileges \
  > "$backup_dir/source-final.dump"

sha256sum "$backup_dir/source-final.dump" > "$backup_dir/source-final.dump.sha256"

echo "Replacing RDS public schema with the final source dump."
docker run --rm \
  -e RDS_SQL_DSN="$RDS_SQL_DSN" \
  -v "$backup_dir:/backup" \
  "$TARGET_CLIENT_IMAGE" \
  sh -eu -c '
    psql "$RDS_SQL_DSN" -v ON_ERROR_STOP=1 -c "DROP SCHEMA IF EXISTS public CASCADE;"
    psql "$RDS_SQL_DSN" -v ON_ERROR_STOP=1 -c "CREATE SCHEMA public;"
    pg_restore --exit-on-error --no-owner --no-privileges -d "$RDS_SQL_DSN" /backup/source-final.dump
  '

echo "Writing exact target table counts."
docker run --rm \
  -i \
  -e RDS_SQL_DSN="$RDS_SQL_DSN" \
  "$TARGET_CLIENT_IMAGE" \
  psql "$RDS_SQL_DSN" -At -F $'\t' <<'SQL' \
  | awk -F '\t' '$1 ~ /^public[.]/ { print }' > "$backup_dir/target-counts.tsv"
SELECT format('SELECT %L, count(*) FROM %I.%I;', schemaname || '.' || tablename, schemaname, tablename)
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY schemaname, tablename;
\gexec
SQL

if ! diff -u "$backup_dir/source-counts.tsv" "$backup_dir/target-counts.tsv" > "$backup_dir/count-diff.txt"; then
  echo "Source and target row counts do not match. See $backup_dir/count-diff.txt"
  exit 1
fi

source_tables="$(wc -l < "$backup_dir/source-counts.tsv" | tr -d ' ')"
source_rows="$(awk -F '\t' '{ total += $2 } END { print total + 0 }' "$backup_dir/source-counts.tsv")"

echo "Switching application SQL_DSN to RDS."
awk '!/^SQL_DSN=/{ print } END { print "SQL_DSN=" ENVIRON["RDS_SQL_DSN"] }' .env > "$backup_dir/env-new"
chmod 600 "$backup_dir/env-new"
cp -a "$backup_dir/env-new" .env

docker compose up -d new-api

echo "Waiting for application health check after RDS cutover."
for attempt in $(seq 1 90); do
  if curl -fsS "http://127.0.0.1:${PORT}/api/status" > "$backup_dir/status-after-cutover.json"; then
    if grep -q '"success"[[:space:]]*:[[:space:]]*true' "$backup_dir/status-after-cutover.json"; then
      cutover_confirmed=1
      break
    fi
  fi
  sleep 2
done

if [ "$cutover_confirmed" -ne 1 ]; then
  echo "Application did not become healthy on RDS; restoring original .env."
  cp -a "$env_backup" .env
  docker compose up -d new-api
  exit 1
fi

{
  echo "backup_dir=$backup_dir"
  echo "source_database=$POSTGRES_DB"
  echo "target_database=$target_db"
  echo "source_tables=$source_tables"
  echo "source_rows=$source_rows"
  echo "source_dump=$backup_dir/source-final.dump"
  echo "source_dump_sha256=$(cut -d ' ' -f 1 "$backup_dir/source-final.dump.sha256")"
  echo "env_backup=$env_backup"
  echo "status_after_cutover=$backup_dir/status-after-cutover.json"
} > "$backup_dir/report.txt"

echo "Migration complete."
cat "$backup_dir/report.txt"
docker compose ps
