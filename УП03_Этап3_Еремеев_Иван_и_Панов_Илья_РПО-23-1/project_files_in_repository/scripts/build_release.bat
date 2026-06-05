@echo off
chcp 65001 > nul
cd /d "%~dp0\.."

echo ============================================================
echo  LandMark — сборка release-билда мобильного клиента (Flutter)
echo ============================================================
echo.
echo Базовый адрес backend для сборки: %API_BASE_URL%
if "%API_BASE_URL%"=="" set API_BASE_URL=http://localhost:8080

cd /d "%~dp0\..\landmark_app"

echo [1/3] flutter pub get...
call flutter pub get
if errorlevel 1 exit /b 1

echo.
echo [2/3] Сборка web-релиза (можно заменить на: flutter build apk --release)...
call flutter build web --release --dart-define=API_BASE_URL=%API_BASE_URL%
if errorlevel 1 exit /b 1

echo.
echo [3/3] Упаковка в release\project_release.zip...
cd /d "%~dp0\.."
if not exist "release" mkdir "release"
powershell -NoProfile -Command "Compress-Archive -Path 'landmark_app/build/web/*' -DestinationPath 'release/project_release.zip' -Force"

echo.
echo Готово. Артефакт: landmark_app\build\web  и  release\project_release.zip
echo Открыть локально: cd landmark_app\build\web ^&^& python -m http.server 8090
