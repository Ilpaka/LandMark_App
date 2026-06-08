@echo off
REM LandMark stage 4 - run Go unit tests across all backend modules.
REM ASCII-only on purpose (cmd.exe parses .bat in the console code page).
setlocal
set "SUB=%~dp0.."
set "REPO=%~dp0..\.."
set "OUT=%SUB%\reports\go_test_report.txt"

echo ============================================================
echo  LandMark - Go test suite
echo ============================================================
if not exist "%SUB%\reports" mkdir "%SUB%\reports"
> "%OUT%" echo LandMark - Go test suite
for /d %%m in ("%REPO%\backend\*") do (
  if exist "%%m\go.mod" (
    echo.
    echo ^>^>^> module: %%~nxm
    echo. >> "%OUT%"
    echo ^>^>^> module: %%~nxm >> "%OUT%"
    pushd "%%m"
    go test ./... -count=1 >> "%OUT%" 2>&1
    go test ./... -count=1
    popd
  )
)
echo.
echo Report saved to %OUT%
echo Done. Save a screenshot of this output (07_tests_success.png).
endlocal
