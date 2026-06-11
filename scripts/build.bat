@echo off
chcp 65001 > /dev/null
cd /d "%~dp0\.."
echo ========================================
echo  Сборка release-бинарников backend
echo ========================================
if not exist dist mkdir dist
for %%s in (api-gateway auth-service favorites-service journal-service media-service moderation-service notifications-service places-service profile-service trips-service) do (
  echo ^>^> go build: %%s
  pushd backend\%%s
  set CGO_ENABLED=0
  go build -o ..\..\dist\%%s.exe ./cmd/server || goto :err
  popd
)
echo.
echo Бинарники собраны в dist\.
goto :eof
:err
echo ОШИБКА: сборка не удалась.
exit /b 1
