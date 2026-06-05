@echo off
chcp 65001 > nul
cd /d "%~dp0\.."

echo ============================================================
echo  LandMark — перезапуск production/demo-стенда
echo ============================================================

echo [1/2] Остановка контейнеров...
docker compose --env-file .env.production -f docker-compose.prod.yml down

echo.
echo [2/2] Повторный запуск...
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d

echo.
docker compose --env-file .env.production -f docker-compose.prod.yml ps
echo Перезапуск завершён. API: http://localhost:8080
