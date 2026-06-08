#!/usr/bin/env bash
# LandMark - API test matrix (stage 4)
# Runs success + error scenarios against the API Gateway and asserts HTTP codes.
# Usage: BASE=http://localhost:8080 bash run_api_tests.sh
set -u
BASE="${BASE:-http://localhost:8080}"
PASS=0; FAIL=0
RUN_ID="$RANDOM$RANDOM"
EMAIL="qa.user.${RUN_ID}@example.com"
PASSWORD="QaPass1234567"

line() { printf '%s\n' "------------------------------------------------------------"; }

# check NAME EXPECTED METHOD PATH [data] [authHeader]
check() {
  local name="$1" expected="$2" method="$3" path="$4" data="${5:-}" auth="${6:-}"
  local args=(-s -o /tmp/body.$$ -w '%{http_code} %{time_total}' -X "$method" "$BASE$path")
  [ -n "$data" ] && args+=(-H "Content-Type: application/json" -d "$data")
  [ -n "$auth" ] && args+=(-H "Authorization: Bearer $auth")
  local out code time
  out="$(curl "${args[@]}")"; code="${out%% *}"; time="${out##* }"
  local body; body="$(head -c 200 /tmp/body.$$ 2>/dev/null)"; rm -f /tmp/body.$$
  if [ "$code" = "$expected" ]; then
    printf '[PASS] %-34s %s %-26s -> %s (%ss)\n' "$name" "$method" "$path" "$code" "$time"; PASS=$((PASS+1))
  else
    printf '[FAIL] %-34s %s %-26s -> %s (expected %s)\n' "$name" "$method" "$path" "$code" "$expected"; FAIL=$((FAIL+1))
  fi
  [ -n "$body" ] && printf '       body: %s\n' "$body"
}

echo "LandMark API test matrix"
echo "Base URL: $BASE"
echo "Run id:   $RUN_ID"
line

echo "[1] Health & public reads"
check "health endpoint"            200 GET  "/healthz"
check "public list places"         200 GET  "/v1/places"
line

echo "[2] Auth flow (registration -> verify -> login)"
REG="$(curl -s -X POST "$BASE/v1/auth/register" -H "Content-Type: application/json" \
       -d "{\"name\":\"QA User\",\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")"
echo "       register response: $(printf '%s' "$REG" | head -c 200)"
VID="$(printf '%s' "$REG" | sed -n 's/.*"verification_id":"\([^"]*\)".*/\1/p')"
CODE="$(printf '%s' "$REG" | sed -n 's/.*"dev_code":"\([^"]*\)".*/\1/p')"
check "verify email"               200 POST "/v1/auth/verify-email" "{\"verification_id\":\"$VID\",\"code\":\"$CODE\"}"
LOGIN="$(curl -s -X POST "$BASE/v1/auth/login" -H "Content-Type: application/json" \
        -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")"
TOKEN="$(printf '%s' "$LOGIN" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')"
if [ -n "$TOKEN" ]; then echo "       login OK, token acquired (len=${#TOKEN})"; PASS=$((PASS+1));
else echo "       login FAILED (no token)"; FAIL=$((FAIL+1)); fi
line

echo "[3] Authenticated requests + create (DB change)"
check "profile/me WITH token"      200 GET  "/v1/profile/me" "" "$TOKEN"
check "create trip (POST)"         201 POST "/v1/trips" "{\"title\":\"QA trip $RUN_ID\"}" "$TOKEN"
# DB-change verification: the created trip must now appear in the user's list
TRIPS="$(curl -s "$BASE/v1/trips" -H "Authorization: Bearer $TOKEN")"
if printf '%s' "$TRIPS" | grep -q "QA trip $RUN_ID"; then
  echo "       [PASS] DB change verified: created trip is listed in GET /v1/trips"; PASS=$((PASS+1));
else
  echo "       [FAIL] created trip NOT found in GET /v1/trips"; FAIL=$((FAIL+1)); fi
line

echo "[4] Error scenarios (must be handled, not 500)"
check "login wrong password"       401 POST "/v1/auth/login" "{\"email\":\"$EMAIL\",\"password\":\"WrongPass99999\"}"
check "register invalid email"     400 POST "/v1/auth/register" "{\"name\":\"x\",\"email\":\"not-an-email\",\"password\":\"$PASSWORD\"}"
check "register short password"    400 POST "/v1/auth/register" "{\"name\":\"x\",\"email\":\"a@b.io\",\"password\":\"short\"}"
check "profile WITHOUT token"      401 GET  "/v1/profile"
check "place not found (uuid)"     404 GET  "/v1/places/00000000-0000-0000-0000-000000000000"
line

printf 'RESULT: %s passed, %s failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ] && echo "ALL CHECKS PASSED" || echo "SOME CHECKS FAILED (see DEFECT_LOG.md)"
