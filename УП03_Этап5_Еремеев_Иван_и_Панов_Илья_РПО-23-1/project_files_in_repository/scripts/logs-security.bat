@echo off
REM LandMark stage 5 - collect logs and scan for critical errors. ASCII-only.
setlocal
set "REPO=%~dp0..\.."
set "SUB=%~dp0.."
set "OUT=%SUB%\reports\logs_security.txt"
if not exist "%SUB%\reports" mkdir "%SUB%\reports"

echo ============================================================
echo  LandMark - logs security scan
echo ============================================================
pushd "%REPO%"
docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=150 --no-color > "%OUT%" 2>&1
popd
echo Saved %OUT%

echo.
echo === critical entries in application services (panic/fatal) ===
findstr /I /C:"panic" /C:"fatal" "%OUT%" | findstr /V /I "grafana postgres"
if errorlevel 1 (
  echo [OK] No panic/fatal in application services.
) else (
  echo [!] Review the entries above.
)
echo (postgres "starting up" and restore-test messages are benign noise.)
endlocal
