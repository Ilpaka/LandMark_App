@echo off
REM LandMark stage 5 - PostgreSQL backup via pg_dump. ASCII-only.
setlocal
set "REPO=%~dp0..\.."
set "SUB=%~dp0.."
if not exist "%SUB%\backups" mkdir "%SUB%\backups"
for /f %%t in ('powershell -NoProfile -Command "Get-Date -Format yyyyMMdd_HHmmss"') do set "TS=%%t"
set "BK=%SUB%\backups\backup_%TS%.sql"

echo ============================================================
echo  LandMark - database backup
echo ============================================================
pushd "%REPO%"
docker compose --env-file .env.production -f docker-compose.prod.yml exec -T postgres pg_dump -U landmark landmark > "%BK%"
popd

if exist "%BK%" (
  echo Backup created: backups\backup_%TS%.sql
  dir "%SUB%\backups"
) else (
  echo [!] Backup failed - is the stand running?
)
echo Note: backups\ is in .gitignore and is NOT committed.
endlocal
