#!/usr/bin/env bash
# LandMark — production/demo развертывание (Docker), вариант для Linux/VPS.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "============================================================"
echo " LandMark — production/demo развертывание (Docker)"
echo "============================================================"

if [ ! -f .env.production ]; then
  echo "[i] .env.production не найден — копирую .env.demo.example -> .env.production"
  cp .env.demo.example .env.production
  echo "[i] Для реального production отредактируйте .env.production и задайте секреты."
fi

echo "[1/2] Сборка и запуск контейнеров..."
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d

echo "[2/2] Статус сервисов:"
docker compose --env-file .env.production -f docker-compose.prod.yml ps

echo "Готово. API: http://localhost:8080 (проверка: /healthz)"
