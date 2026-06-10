@echo off
REM LandMark stage 5 - dependency vulnerability check (Go). ASCII-only.
setlocal
set "REPO=%~dp0..\.."
pushd "%REPO%\backend"

echo ============================================================
echo  LandMark - dependency vulnerability scan (govulncheck)
echo ============================================================
where govulncheck >nul 2>&1
if errorlevel 1 (
  echo govulncheck not installed. Install once:
  echo   go install golang.org/x/vuln/cmd/govulncheck@latest
  goto :done
)
for %%m in (auth-service api-gateway) do (
  echo.
  echo --- module %%m ---
  pushd "%%m"
  govulncheck ./...
  popd
)
:done
echo.
echo Outdated modules (optional): run inside a module:  go list -u -m all
echo Dart client deps: run "flutter pub outdated" where the Flutter SDK is available.
popd
endlocal
