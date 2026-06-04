@echo off
chcp 65001 > nul
cd /d "%~dp0\.."
echo ========================================
echo  Поднять backend (Docker, detached)
echo ========================================
docker compose up --build -d
docker compose ps
