@echo off
chcp 65001 > nul
cd /d "%~dp0\.."

echo ============================================================
echo  LandMark — проверка развертывания
echo ============================================================

echo [1/3] Статус контейнеров:
docker compose --env-file .env.production -f docker-compose.prod.yml ps

echo.
echo [2/3] Проверка доступности API Gateway (/healthz):
curl -s -o nul -w "HTTP %%{http_code}\n" http://localhost:8080/healthz
if errorlevel 1 echo [!] API Gateway недоступен на http://localhost:8080

echo.
echo [3/3] Последние строки логов:
docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=80
