@echo off
chcp 65001 > nul
cd /d "%~dp0\.."

echo ============================================================
echo  LandMark — production/demo развертывание (Docker)
echo ============================================================

if not exist ".env.production" (
  echo [i] Файл .env.production не найден.
  echo     Копирую .env.demo.example -^> .env.production для демо-стенда.
  copy ".env.demo.example" ".env.production" > nul
  echo [i] Для реального production отредактируйте .env.production
  echo     и подставьте настоящие секреты, затем запустите deploy.bat снова.
)

echo.
echo [1/2] Сборка и запуск контейнеров...
docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d
if errorlevel 1 (
  echo [!] Развертывание завершилось с ошибкой. Смотрите логи: scripts\check_deploy.bat
  exit /b 1
)

echo.
echo [2/2] Статус сервисов:
docker compose --env-file .env.production -f docker-compose.prod.yml ps

echo.
echo Готово. API доступен на http://localhost:8080  (проверка: /healthz)
echo Grafana: http://localhost:3000   Mailpit: http://localhost:8025
