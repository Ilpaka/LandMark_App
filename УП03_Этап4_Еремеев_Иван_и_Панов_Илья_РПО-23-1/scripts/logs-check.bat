@echo off
REM LandMark stage 4 - collect container logs and scan for critical errors.
setlocal
set "SUB=%~dp0.."
set "REPO=%~dp0..\.."
set "OUT=%SUB%\reports\logs_tail.txt"
if not exist "%SUB%\reports" mkdir "%SUB%\reports"

echo ============================================================
echo  LandMark - logs check
echo ============================================================
pushd "%REPO%"
docker compose --env-file .env.production -f docker-compose.prod.yml logs --tail=120 --no-color > "%OUT%" 2>&1
popd

echo Saved %OUT%
echo.
echo --- scanning for critical errors (panic / fatal / 5xx) in services ---
findstr /I /C:"panic" /C:"fatal" /C:" 500 " "%OUT%"
if %errorlevel%==0 (
  echo [!] Potential critical entries found above - review them.
) else (
  echo [OK] No panic/fatal/500 entries in service logs.
)
echo Done. Save screenshot 08_logs_without_critical_errors.png.
endlocal
