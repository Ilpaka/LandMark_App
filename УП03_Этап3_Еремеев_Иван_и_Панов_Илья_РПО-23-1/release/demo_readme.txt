LandMark (Wanderlog) — demo / release
======================================

Это краткая инструкция к release/demo-варианту проекта.

A. Backend как demo-стенд (Docker)
----------------------------------
1. Установить Docker Desktop / Docker Engine + Compose v2.
2. В корне репозитория:
     cp .env.demo.example .env.production
     docker compose --env-file .env.production -f docker-compose.prod.yml up --build -d
   (Windows: scripts\deploy.bat)
3. Открыть http://localhost:8080/healthz — должен вернуться HTTP 200.
4. Проверить сценарий по DEMO_GUIDE.md.

B. Клиент как release-билд (Flutter web)
----------------------------------------
1. Собрать: scripts\build_release.bat
   (получится landmark_app\build\web и release\project_release.zip).
2. Запустить локально без IDE, например:
     cd landmark_app\build\web
     python -m http.server 8090
   и открыть http://localhost:8090
3. Клиент обращается к backend по адресу API_BASE_URL (по умолчанию
   http://localhost:8080).

Тестовые аккаунты — см. demo_accounts.txt
