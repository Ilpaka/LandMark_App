#!/usr/bin/env bash
# Smoke-сценарий регистрации + верификации + логина через API Gateway.
# Использует только dev_code из ответа /register (логи docker не нужны).
# Запуск из корня репозитория:  ./scripts/test-auth-flow.sh
set -euo pipefail

GW="${GW:-http://localhost:8080}"
SUFFIX="$(date +%s)"
EMAIL="auth-smoke-${SUFFIX}@local.test"
PASS="StrongPass123"

echo "Gateway:     $GW"
echo "Test email:  $EMAIL"
echo

echo "Health:"
curl -sf "$GW/healthz"; echo
echo

echo "[1/3] register"
REG=$(curl -s -X POST "$GW/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Smoke User\",\"email\":\"${EMAIL}\",\"password\":\"${PASS}\"}")
echo "  response: $REG"

VID=$(printf '%s' "$REG" | python3 -c 'import sys,json;print(json.load(sys.stdin)["verification_id"])')
CODE=$(printf '%s' "$REG" | python3 -c 'import sys,json;print(json.load(sys.stdin)["dev_code"])')
echo "  verification_id = $VID"
echo "  dev_code        = $CODE"
echo

echo "[2/3] verify-email"
VRF=$(curl -s -X POST "$GW/v1/auth/verify-email" \
  -H "Content-Type: application/json" \
  -d "{\"verification_id\":\"${VID}\",\"code\":\"${CODE}\"}")
echo "  response: $VRF"
echo

echo "[3/3] login"
LOG=$(curl -s -X POST "$GW/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASS}\"}")
echo "  response: $LOG"
echo

ACCESS=$(printf '%s' "$LOG" | python3 -c 'import sys,json;d=json.load(sys.stdin);print(d.get("access_token",""))')
REFRESH=$(printf '%s' "$LOG" | python3 -c 'import sys,json;d=json.load(sys.stdin);print(d.get("refresh_token",""))')
if [ -n "$ACCESS" ] && [ -n "$REFRESH" ]; then
  echo "  access_token  : ${ACCESS:0:32}... (len=${#ACCESS})"
  echo "  refresh_token : ${REFRESH:0:32}... (len=${#REFRESH})"
  echo
  echo "OK"
else
  echo "FAIL: access/refresh tokens missing"
  exit 1
fi
