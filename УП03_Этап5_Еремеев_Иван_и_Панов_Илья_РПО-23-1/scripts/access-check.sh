#!/usr/bin/env bash
# LandMark stage 5 - access-control checks: roles (RBAC) + foreign-data access (IDOR).
# Requires the demo stand running. Usage: BASE=http://localhost:8080 bash access-check.sh
set -u
BASE="${BASE:-http://localhost:8080}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@apple.com}"
ADMIN_PASS="${ADMIN_PASS:-AdminPass123456}"
PASS=0; FAIL=0
ok(){ echo "[PASS] $1"; PASS=$((PASS+1)); }
no(){ echo "[FAIL] $1"; FAIL=$((FAIL+1)); }

code(){ # METHOD PATH TOKEN [data] -> echoes http code
  local m="$1" p="$2" tok="${3:-}" data="${4:-}"
  local a=(-s -o /dev/null -w '%{http_code}' -X "$m" "$BASE$p")
  [ -n "$tok" ] && a+=(-H "Authorization: Bearer $tok")
  [ -n "$data" ] && a+=(-H 'Content-Type: application/json' -d "$data")
  curl "${a[@]}"
}
mktoken(){ # registers+verifies+logs in a fresh user, echoes access token
  local em="$1" pw="UserPass123456"
  local reg vid c
  reg="$(curl -s -X POST "$BASE/v1/auth/register" -H 'Content-Type: application/json' -d "{\"name\":\"U\",\"email\":\"$em\",\"password\":\"$pw\"}")"
  vid="$(printf '%s' "$reg" | sed -n 's/.*"verification_id":"\([^"]*\)".*/\1/p')"
  c="$(printf '%s' "$reg" | sed -n 's/.*"dev_code":"\([^"]*\)".*/\1/p')"
  curl -s -o /dev/null -X POST "$BASE/v1/auth/verify-email" -H 'Content-Type: application/json' -d "{\"verification_id\":\"$vid\",\"code\":\"$c\"}"
  curl -s -X POST "$BASE/v1/auth/login" -H 'Content-Type: application/json' -d "{\"email\":\"$em\",\"password\":\"$pw\"}" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p'
}

echo "LandMark stage 5 - access-control checks @ $BASE"
echo "======================================================================"
RID="$RANDOM$RANDOM"
TOKA="$(mktoken "userA.$RID@example.com")"
TOKB="$(mktoken "userB.$RID@example.com")"
TOKADM="$(curl -s -X POST "$BASE/v1/auth/login" -H 'Content-Type: application/json' -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\"}" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')"
echo "tokens: userA=${TOKA:+ok} userB=${TOKB:+ok} admin=${TOKADM:+ok}"
echo

echo "[1] RBAC - admin-only endpoint /v1/admin/users"
c=$(code GET /v1/admin/users "");      [ "$c" = "401" ] && ok "anonymous -> $c (denied)" || no "anonymous -> $c (expected 401)"
c=$(code GET /v1/admin/users "$TOKA"); { [ "$c" = "403" ] || [ "$c" = "401" ]; } && ok "regular user -> $c (forbidden)" || no "regular user -> $c (expected 403)"
c=$(code GET /v1/admin/users "$TOKADM"); [ "$c" = "200" ] && ok "admin -> $c (allowed)" || no "admin -> $c (expected 200)"
echo

echo "[2] IDOR - user B must not access user A's trip"
TRIP="$(curl -s -X POST "$BASE/v1/trips" -H "Authorization: Bearer $TOKA" -H 'Content-Type: application/json' -d "{\"title\":\"Private trip A $RID\"}")"
TID="$(printf '%s' "$TRIP" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')"
echo "userA created trip id=$TID"
c=$(code GET "/v1/trips/$TID" "$TOKA");  [ "$c" = "200" ] && ok "owner (A) reads own trip -> $c" || no "owner (A) -> $c (expected 200)"
c=$(code GET "/v1/trips/$TID" "$TOKB");  { [ "$c" = "403" ] || [ "$c" = "404" ]; } && ok "other user (B) reads A's trip -> $c (denied)" || no "user B -> $c (EXPECTED 403/404 - possible IDOR!)"
c=$(code PATCH "/v1/trips/$TID" "$TOKB" '{"title":"hacked"}'); { [ "$c" = "403" ] || [ "$c" = "404" ]; } && ok "user B edits A's trip -> $c (denied)" || no "user B PATCH -> $c (EXPECTED 403/404 - possible IDOR!)"
c=$(code DELETE "/v1/trips/$TID" "$TOKB"); { [ "$c" = "403" ] || [ "$c" = "404" ]; } && ok "user B deletes A's trip -> $c (denied)" || no "user B DELETE -> $c (EXPECTED 403/404 - possible IDOR!)"
echo

echo "======================================================================"
printf 'RESULT: %s passed, %s failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ] && echo "ACCESS CONTROL: OK" || echo "ACCESS CONTROL: ISSUES FOUND (see RISK_REGISTER.md)"
