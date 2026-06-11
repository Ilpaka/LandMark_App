@echo off
chcp 65001 > /dev/null
cd /d "%~dp0\.."
echo ========================================
echo  Тесты проекта (go test + flutter test)
echo ========================================
for %%s in (api-gateway auth-service favorites-service journal-service media-service moderation-service notifications-service places-service profile-service trips-service) do (
  echo ^>^> go test: %%s
  pushd backend\%%s
  go test ./... -count=1 -short || goto :err
  popd
)
echo ^>^> flutter test
pushd landmark_app
call flutter test || goto :err
popd
echo.
echo Тесты пройдены.
goto :eof
:err
echo ОШИБКА: тесты не пройдены.
exit /b 1
