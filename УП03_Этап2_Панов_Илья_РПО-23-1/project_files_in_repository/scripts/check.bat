@echo off
chcp 65001 > nul
cd /d "%~dp0\.."
echo ========================================
echo  Проверка качества (go vet + flutter analyze)
echo ========================================
for %%s in (api-gateway auth-service favorites-service journal-service media-service moderation-service notifications-service places-service profile-service trips-service) do (
  echo ^>^> go vet: %%s
  pushd backend\%%s
  go vet ./... || goto :err
  popd
)
echo ^>^> flutter analyze
pushd landmark_app
call flutter analyze || goto :err
popd
echo.
echo Проверки пройдены.
goto :eof
:err
echo ОШИБКА: проверки не пройдены.
exit /b 1
