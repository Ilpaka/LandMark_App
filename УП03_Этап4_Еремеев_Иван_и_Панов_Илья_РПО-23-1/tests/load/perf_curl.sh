#!/usr/bin/env bash
# LandMark - lightweight performance measurement (no k6 required).
# Measures response time of key endpoints over N iterations and prints min/avg/max.
# Usage: BASE=http://localhost:8080 N=30 bash perf_curl.sh
set -u
BASE="${BASE:-http://localhost:8080}"
N="${N:-30}"

measure() {
  local name="$1" method="$2" path="$3" data="${4:-}"
  local times=() i out code t
  for ((i=0;i<N;i++)); do
    if [ -n "$data" ]; then
      out="$(curl -s -o /dev/null -w '%{http_code} %{time_total}' -X "$method" -H 'Content-Type: application/json' -d "$data" "$BASE$path")"
    else
      out="$(curl -s -o /dev/null -w '%{http_code} %{time_total}' -X "$method" "$BASE$path")"
    fi
    code="${out%% *}"; t="${out##* }"; times+=("$t")
  done
  printf '%s' "${times[@]/%/$'\n'}" | awk -v n="$name" -v code="$code" '
    { v=$1*1000; a[NR]=v; s+=v; if(min==""||v<min)min=v; if(v>max)max=v }
    END { m=s/NR; asort(a); p95=a[int(NR*0.95)?int(NR*0.95):1];
      printf "%-26s last=%s  min=%6.1f ms  avg=%6.1f ms  p95=%6.1f ms  max=%6.1f ms  (n=%d)\n", n, code, min, m, p95, max, NR }'
}

echo "LandMark - API performance measurement"
echo "Base URL: $BASE   iterations per endpoint: $N"
echo "------------------------------------------------------------------------------------"
measure "GET /healthz"        GET  "/healthz"
measure "GET /v1/places"      GET  "/v1/places"
measure "GET /v1/places?bad"  GET  "/v1/places/00000000-0000-0000-0000-000000000000"
measure "POST /v1/auth/login" POST "/v1/auth/login" '{"email":"nobody@example.com","password":"WrongPass123"}'
echo "------------------------------------------------------------------------------------"
echo "Note: measured on the local Docker demo stand; absolute numbers depend on the host."
