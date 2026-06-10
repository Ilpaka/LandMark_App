@echo off
REM LandMark stage 5 - restore a backup into a TEST database and verify. ASCII-only.
REM Usage: restore.bat [path\to\backup.sql]   (default: newest in backups\)
setlocal
set "REPO=%~dp0..\.."
set "SUB=%~dp0.."
set "BK=%~1"
if "%BK%"=="" for /f "delims=" %%f in ('dir /b /o-d "%SUB%\backups\backup_*.sql" 2^>nul') do if not defined BK set "BK=%SUB%\backups\%%f"
if "%BK%"=="" ( echo No backup file found in backups\ - run backup.bat first. & exit /b 1 )

echo ============================================================
echo  LandMark - restore + verify (test DB landmark_restore_test)
echo ============================================================
echo Restoring: %BK%
pushd "%REPO%"
set "DC=docker compose --env-file .env.production -f docker-compose.prod.yml exec -T postgres"
%DC% psql -U landmark -d landmark -c "DROP DATABASE IF EXISTS landmark_restore_test;"
%DC% psql -U landmark -d landmark -c "CREATE DATABASE landmark_restore_test;"
%DC% psql -U landmark -d landmark_restore_test < "%BK%"
echo.
echo --- row counts in restored DB (compare with original) ---
%DC% psql -U landmark -d landmark_restore_test -c "SELECT 'auth_accounts' AS t, count(*) FROM auth.auth_accounts UNION ALL SELECT 'trips', count(*) FROM trips.trips_trips UNION ALL SELECT 'sessions', count(*) FROM auth.auth_sessions;"
popd
echo.
echo If counts match the original DB, the backup is valid.
echo (Drop the test DB afterwards: psql -c "DROP DATABASE landmark_restore_test;")
endlocal
