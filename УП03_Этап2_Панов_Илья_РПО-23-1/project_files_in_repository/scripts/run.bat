@echo off
chcp 65001 > nul
cd /d "%~dp0\.."
echo ========================================
echo  Запуск backend (Docker Compose)
echo ========================================
docker compose up --build
