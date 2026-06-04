@echo off
chcp 65001 > nul
cd /d "%~dp0\.."
echo ========================================
echo  Запуск мобильного клиента (Flutter)
echo ========================================
cd landmark_app
flutter run
