@echo off
REM LandMark stage 5 - open ports and running services. ASCII-only.
setlocal
set "REPO=%~dp0..\.."
pushd "%REPO%"

echo ============================================================
echo  LandMark - containers and published ports
echo ============================================================
docker compose --env-file .env.production -f docker-compose.prod.yml ps
popd

echo.
echo === Windows LISTENING ports (stand: 8080/3000/8025/9090/9000/9001) ===
netstat -ano -p tcp | findstr LISTENING | findstr ":8080 :3000 :8025 :9090 :9000 :9001 :5432 :6379"

echo.
echo Expected: only api-gateway(8080) + operator UIs are published.
echo PostgreSQL(5432) and Redis(6379) must NOT be published to the host.
endlocal
