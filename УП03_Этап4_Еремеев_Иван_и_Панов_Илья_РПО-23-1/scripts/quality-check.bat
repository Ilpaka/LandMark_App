@echo off
REM LandMark stage 4 - one command that runs the key quality checks.
setlocal
echo ========================================
echo  LandMark - overall quality check
echo ========================================

call "%~dp0test.bat"
call "%~dp0api-test.bat"
call "%~dp0logs-check.bat"

echo.
echo ========================================
echo  Quality check completed.
echo  See reports\ for go_test_report.txt, api_test_run.txt, logs_tail.txt
echo  Save a screenshot of the result.
echo ========================================
endlocal
