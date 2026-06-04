@echo off
chcp 65001 > nul
cd /d "%~dp0\.."
echo ========================================
echo  Форматирование (gofmt + dart format)
echo ========================================
for %%s in (api-gateway auth-service favorites-service journal-service media-service moderation-service notifications-service places-service profile-service trips-service) do (
  echo ^>^> gofmt -w: %%s
  pushd backend\%%s
  gofmt -w .
  popd
)
echo ^>^> dart format
pushd landmark_app
call dart format lib test
popd
echo.
echo Код отформатирован.
