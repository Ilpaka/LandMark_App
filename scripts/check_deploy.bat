@echo off
cd /d "%~dp0\.."

echo ============================================================
echo  LandMark - deployment check
echo ============================================================

echo [1/3] Container status:
docker compose --env-file .env.production -f docker-compose.prod.yml ps

echo.
echo [2/3] API Gateway availability (/healthz):
curl -s -o nul -w "HTTP %%{http_code}\n" http://localhost:8080/healthz
if errorlevel 1 echo [!] API Gateway not reachable at http://localhost:8080

echo.
echo [3/3] Recent log lines:
docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=80