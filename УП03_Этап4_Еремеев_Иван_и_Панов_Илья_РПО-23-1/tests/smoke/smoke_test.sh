#!/usr/bin/env bash
# LandMark - smoke test: quick "is the deployed stand alive and usable?" check.
# Exit code 0 = healthy. Usage: BASE=http://localhost:8080 bash smoke_test.sh
set -u
BASE="${BASE:-http://localhost:8080}"
fail=0

step() { # name expected url
  local code; code="$(curl -s -o /dev/null -w '%{http_code}' "$2")"
  if [ "$code" = "$1" ]; then echo "[OK]   $3 -> $code"; else echo "[FAIL] $3 -> $code (expected $1)"; fail=1; fi
}

echo "LandMark smoke test @ $BASE"
step 200 "$BASE/healthz"        "health"
step 200 "$BASE/v1/places"      "public places list"
step 401 "$BASE/v1/profile"     "protected route rejects anonymous"
# main scenario: register -> verify -> login must yield a token
RID="$RANDOM$RANDOM"; EM="smoke.$RID@example.com"; PW="SmokePass12345"
REG="$(curl -s -X POST "$BASE/v1/auth/register" -H 'Content-Type: application/json' -d "{\"name\":\"S\",\"email\":\"$EM\",\"password\":\"$PW\"}")"
VID="$(printf '%s' "$REG" | sed -n 's/.*"verification_id":"\([^"]*\)".*/\1/p')"
CODE="$(printf '%s' "$REG" | sed -n 's/.*"dev_code":"\([^"]*\)".*/\1/p')"
curl -s -o /dev/null -X POST "$BASE/v1/auth/verify-email" -H 'Content-Type: application/json' -d "{\"verification_id\":\"$VID\",\"code\":\"$CODE\"}"
TOK="$(curl -s -X POST "$BASE/v1/auth/login" -H 'Content-Type: application/json' -d "{\"email\":\"$EM\",\"password\":\"$PW\"}" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')"
if [ -n "$TOK" ]; then echo "[OK]   main scenario: register/verify/login -> token"; else echo "[FAIL] main scenario: no token"; fail=1; fi

[ "$fail" -eq 0 ] && { echo "SMOKE: PASS"; exit 0; } || { echo "SMOKE: FAIL"; exit 1; }
