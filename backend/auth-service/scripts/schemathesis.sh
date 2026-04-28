#!/usr/bin/env bash
set -euo pipefail
# Run against a live base URL (e.g. http://127.0.0.1:8080) with OpenAPI spec.
# Usage: BASE_URL=http://127.0.0.1:8080 SPEC=./api/openapi/auth.yaml ./scripts/schemathesis.sh
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SPEC="${SPEC:-$ROOT/api/openapi/auth.yaml}"
BASE_URL="${BASE_URL:?set BASE_URL to running auth API}"
python3 -m venv "${ROOT}/.venv-schemathesis"
# shellcheck disable=SC1091
source "${ROOT}/.venv-schemathesis/bin/activate"
pip install -q schemathesis
schemathesis run "$SPEC" --base-url="$BASE_URL" --hypothesis-max-examples=30 --stateful=links
