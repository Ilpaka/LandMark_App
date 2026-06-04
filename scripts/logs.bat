@echo off
chcp 65001 > nul
cd /d "%~dp0\.."
echo ========================================
echo  Логи контейнеров (Ctrl+C для выхода)
echo ========================================
docker compose logs -f
