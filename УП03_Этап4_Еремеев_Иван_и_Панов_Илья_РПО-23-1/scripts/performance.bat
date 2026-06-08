@echo off
REM LandMark stage 4 - performance measurement.
REM Uses k6 if installed, otherwise falls back to a curl-based timing harness.
setlocal
set "SUB=%~dp0.."
set "BASE=%BASE%"
if "%BASE%"=="" set "BASE=http://localhost:8080"
if not exist "%SUB%\reports" mkdir "%SUB%\reports"

echo ============================================================
echo  LandMark - performance check
echo ============================================================

where k6 >nul 2>&1
if %errorlevel%==0 (
  echo [k6] running tests\load\basic_load.js ...
  k6 run "%SUB%\tests\load\basic_load.js"
) else (
  echo [i] k6 not found - using curl timing harness instead.
  bash "%SUB%\tests\load\perf_curl.sh" > "%SUB%\reports\perf_curl_summary.txt" 2>&1
  type "%SUB%\reports\perf_curl_summary.txt"
)
echo.
echo For a web Lighthouse report run:
echo   npx lighthouse http://localhost:3000/login --output html --output-path "%SUB%\reports\lighthouse_report.html" --chrome-flags="--headless=new"
echo Done. Save screenshot 04_lighthouse_or_performance.png.
endlocal
