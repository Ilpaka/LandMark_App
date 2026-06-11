@echo off
chcp 65001 > /dev/null
cd /d "%~dp0\.."
if "%1"=="" (
  echo Использование: scripts\create-release.bat vX.Y.Z
  echo Пример:        scripts\create-release.bat v0.3.1
  exit /b 1
)
set VERSION=%1
echo ========================================
echo  Подготовка релиза %VERSION%
echo ========================================
call scripts\release-check.bat || goto :err
if not exist release mkdir release
echo ^>^> Архив исходников release\landmark-%VERSION%-src.zip
git archive --format=zip -o release\landmark-%VERSION%-src.zip HEAD || goto :err
echo ^>^> Тег %VERSION%
git tag -a %VERSION% -m "Release %VERSION%" || goto :err
echo.
echo Готово. Осталось: git push origin %VERSION%
echo (по тегу GitHub Actions соберёт релиз автоматически — .github/workflows/release.yml)
goto :eof
:err
echo ОШИБКА: релиз не подготовлен.
exit /b 1
