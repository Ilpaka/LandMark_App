@echo off
REM LandMark stage 4 - API test matrix (success + error scenarios) via curl.
REM Requires the demo stand running (docker compose ... up -d) and bash (Git Bash).
setlocal
set "SUB=%~dp0.."
set "BASE=%BASE%"
if "%BASE%"=="" set "BASE=http://localhost:8080"

echo ============================================================
echo  LandMark - API tests against %BASE%
echo ============================================================
if not exist "%SUB%\reports" mkdir "%SUB%\reports"
bash "%SUB%\tests\api\run_api_tests.sh" > "%SUB%\reports\api_test_run.txt" 2>&1
type "%SUB%\reports\api_test_run.txt"
echo.
echo Report saved to %SUB%\reports\api_test_run.txt
echo Done. Save screenshots 05_api_success_request.png and 06_api_error_request_handled.png.
endlocal
