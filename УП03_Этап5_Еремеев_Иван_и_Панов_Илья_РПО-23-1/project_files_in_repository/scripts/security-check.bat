@echo off
REM LandMark stage 5 - secret / git hygiene check. ASCII-only.
setlocal
set "REPO=%~dp0..\.."
pushd "%REPO%"

echo ============================================================
echo  LandMark - secret scan and git hygiene
echo ============================================================
echo.
echo === Suspicious keywords in code/config (excluding docs/diagrams/examples/tests) ===
git grep -n -i -E "password|secret|token|api_key|apikey|jwt|smtp|database_url" -- "." ":!*.md" ":!WIKI/" ":!*.drawio" ":!*.html" ":!*.example" ":!*_test.go"

echo.
echo === .env files tracked by git (must be only *.example) ===
git ls-files | findstr /I ".env"

echo.
echo === Git status (nothing secret should be staged) ===
git status --short

popd
echo.
echo If real passwords/tokens appear above - remove them and use .env.example placeholders.
endlocal
