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

docker compose pull
docker compose up -d --remove-orphans

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

if [ -L "$LEGACY_SITE" ] && grep -q "server_name ${DOMAIN}" "$LEGACY_SITE"; then
  LEGACY_TARGET="$(readlink -f "$LEGACY_SITE")"
  cp -a "$LEGACY_TARGET" "${LEGACY_TARGET}.disabled-$(date +%Y%m%d%H%M%S)"
  rm -f "$LEGACY_SITE"
fi

install -m 0644 "nginx-${DOMAIN}.conf" "$NGINX_CONF"
nginx -t
systemctl reload nginx

docker compose ps
