#!/usr/bin/env bash
# LandMark — перезапуск production/demo-стенда (Linux/VPS).
set -euo pipefail
cd "$(dirname "$0")/.."

echo "[1/2] Остановка контейнеров..."
docker compose --env-file .env.production -f docker-compose.prod.yml down

echo "[2/2] Повторный запуск..."
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d
docker compose --env-file .env.production -f docker-compose.prod.yml ps
echo "Перезапуск завершён. API: http://localhost:8080"
