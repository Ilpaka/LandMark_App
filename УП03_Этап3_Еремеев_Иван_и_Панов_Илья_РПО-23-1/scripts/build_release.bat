@echo off
cd /d "%~dp0\.."

echo ============================================================
echo  LandMark - mobile client release build (Flutter)
echo ============================================================
echo.
if "%API_BASE_URL%"=="" set API_BASE_URL=http://localhost:8080
echo Backend base URL for build: %API_BASE_URL%

cd /d "%~dp0\..\landmark_app"

echo [1/3] flutter pub get...
call flutter pub get
if errorlevel 1 exit /b 1

echo.
echo [2/3] Building web release (or use: flutter build apk --release)...
call flutter build web --release --dart-define=API_BASE_URL=%API_BASE_URL%
if errorlevel 1 exit /b 1

echo.
echo [3/3] Packaging into release\project_release.zip...
cd /d "%~dp0\.."
if not exist "release" mkdir "release"
powershell -NoProfile -Command "Compress-Archive -Path 'landmark_app/build/web/*' -DestinationPath 'release/project_release.zip' -Force"

echo.
echo Done. Artifacts: landmark_app\build\web  and  release\project_release.zip
echo Serve locally: cd landmark_app\build\web ^&^& python -m http.server 8090