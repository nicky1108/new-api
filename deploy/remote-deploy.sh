#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${APP_DIR:-/opt/new-api}"
DOMAIN="${DOMAIN:-openhubs.xyz}"
PORT="${NEW_API_PORT:-3000}"
NGINX_CONF="/etc/nginx/conf.d/${DOMAIN}.conf"
LEGACY_SITE="/etc/nginx/sites-enabled/agenthub-openai-gateway"

cd "$APP_DIR"

mkdir -p data logs
chmod 700 "$APP_DIR"
install -d -m 0755 /var/www/openhubs
install -m 0644 favicon.svg /var/www/openhubs/favicon.svg
install -d -m 0755 /var/www/openhubs/docs/user-guide
install -m 0644 docs/user-guide/index.html /var/www/openhubs/docs/user-guide/index.html

docker compose pull
docker compose up -d --remove-orphans redis new-api

for attempt in $(seq 1 60); do
  if curl -fsS "http://127.0.0.1:${PORT}/api/status" | grep -q '"success"[[:space:]]*:[[:space:]]*true'; then
    break
  fi

  if [ "$attempt" -eq 60 ]; then
    docker compose ps
    docker compose logs --tail=120 new-api
    exit 1
  fi

  sleep 2
done

APP_SQL_DSN="$(awk -F= '$1 == "SQL_DSN" { sub(/^[^=]*=/, ""); print; exit }' .env || true)"
if [ -n "$APP_SQL_DSN" ] && ! printf '%s' "$APP_SQL_DSN" | grep -q '@postgres:5432/'; then
  docker compose stop postgres || true
fi

if [ -L "$LEGACY_SITE" ] && grep -q "server_name ${DOMAIN}" "$LEGACY_SITE"; then
  LEGACY_TARGET="$(readlink -f "$LEGACY_SITE")"
  cp -a "$LEGACY_TARGET" "${LEGACY_TARGET}.disabled-$(date +%Y%m%d%H%M%S)"
  rm -f "$LEGACY_SITE"
fi

install -m 0644 "nginx-${DOMAIN}.conf" "$NGINX_CONF"
nginx -t
systemctl reload nginx

docker compose ps -a
