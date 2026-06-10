@echo off
chcp 65001 > /dev/null
cd /d "%~dp0\.."
echo ========================================
echo  Финальная проверка перед релизом
echo ========================================
call scripts\check.bat || goto :err
call scripts\test.bat || goto :err
call scripts\build.bat || goto :err
echo.
echo Проект готов к релизу.
goto :eof
:err
echo ОШИБКА: релизная проверка не пройдена.
exit /b 1
