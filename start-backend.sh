#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR/backend"

echo "🚀 Запуск бэкенда LandMark..."

# Проверяем что Docker запущен
if ! docker info &>/dev/null; then
  echo "❌ Docker не запущен. Запусти Docker Desktop и повтори."
  exit 1
fi

# context: .. в docker-compose.yml указывает на корень LandMark_App/
docker compose up -d --build

echo ""
echo "✅ Бэкенд запущен!"
echo ""
echo "  API Gateway:  http://localhost:8080"
echo "  MailHog:      http://localhost:8025  (входящие письма)"
echo "  MinIO:        http://localhost:9001  (хранилище медиа)"
echo "  Grafana:      http://localhost:3000  (мониторинг)"
echo ""
echo "Для остановки: cd backend && docker compose down"
