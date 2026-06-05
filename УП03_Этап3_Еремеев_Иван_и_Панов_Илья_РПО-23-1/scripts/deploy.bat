@echo off
cd /d "%~dp0\.."

echo ============================================================
echo  LandMark - production/demo deployment (Docker)
echo ============================================================

if not exist ".env.production" (
  echo [i] File .env.production not found.
  echo     Copying .env.demo.example to .env.production for the demo stand.
  copy ".env.demo.example" ".env.production" > nul
  echo [i] For real production edit .env.production with real secrets,
  echo     then run deploy.bat again.
)

echo.
echo [1/2] Building and starting containers...
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d
if errorlevel 1 (
  echo [!] Deployment failed. Check logs: scripts\check_deploy.bat
  exit /b 1
)

echo.
echo [2/2] Service status:
docker compose --env-file .env.production -f docker-compose.prod.yml ps

echo.
echo Done. API available at http://localhost:8080  (health: /healthz)
echo Grafana: http://localhost:3000   Mailpit: http://localhost:8025