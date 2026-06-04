@echo off
chcp 65001 > nul
cd /d "%~dp0\.."
echo ========================================
echo  Установка зависимостей (Flutter + Go)
echo ========================================
echo.
echo [1/2] flutter pub get
pushd landmark_app
call flutter pub get || goto :err
popd
echo.
echo [2/2] go mod download (по сервисам)
for %%s in (api-gateway auth-service favorites-service journal-service media-service moderation-service notifications-service places-service profile-service trips-service) do (
  echo   ^>^> %%s
  pushd backend\%%s
  go mod download || goto :err
  popd
)
echo.
echo Зависимости установлены.
goto :eof
:err
echo ОШИБКА: не удалось установить зависимости.
exit /b 1
