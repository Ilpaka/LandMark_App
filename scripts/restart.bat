@echo off
cd /d "%~dp0\.."

echo ============================================================
echo  LandMark - restart production/demo stand
echo ============================================================

echo [1/2] Stopping containers...
docker compose --env-file .env.production -f docker-compose.prod.yml down

echo.
echo [2/2] Starting again...
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d

echo.
docker compose --env-file .env.production -f docker-compose.prod.yml ps
echo Restart complete. API: http://localhost:8080