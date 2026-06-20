#!/usr/bin/env bash
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-${BASE_URL:-http://127.0.0.1:5678}}"
API_BASE_URL="${API_BASE_URL%/}"
DEVICE_ID="${DEVICE_ID:-smoke-test-device}"
RESPONSE_FILE="$(mktemp)"

cleanup() {
  rm -f "$RESPONSE_FILE"
}
trap cleanup EXIT

check() {
  local method="$1"
  local path="$2"
  local expected="${3:-200}"
  local url="${API_BASE_URL}${path}"

  local status
  if ! status="$(curl -sS -o "$RESPONSE_FILE" -w "%{http_code}" \
      --connect-timeout 5 \
      --max-time 30 \
      -X "$method" \
      -H "X-Device-ID: ${DEVICE_ID}" \
      -H "Content-Type: application/json" \
      "$url")"; then
    echo "FAIL [network] ${method} ${path}: cannot connect to ${API_BASE_URL}"
    return 1
  fi

  if [[ "$status" != "$expected" ]]; then
    echo "FAIL [${status}] ${method} ${path}: expected HTTP ${expected}"
    sed -n '1,20p' "$RESPONSE_FILE"
    return 1
  fi

  echo "PASS [${status}] ${method} ${path}"
}

extract_first_project_id() {
  if command -v jq >/dev/null 2>&1; then
    jq -r '.data.items[0].id // empty' "$RESPONSE_FILE"
    return
  fi
	if command -v node >/dev/null 2>&1; then
		node -e 'const fs=require("fs");const value=JSON.parse(fs.readFileSync(process.argv[1],"utf8"));process.stdout.write(String(value?.data?.items?.[0]?.id||""))' "$RESPONSE_FILE"
		return
	fi
  grep -o '"id"[[:space:]]*:[[:space:]]*[0-9]\+' "$RESPONSE_FILE" | head -n 1 | grep -o '[0-9]\+' || true
}

echo "Digital Silk Road smoke test"
echo "API_BASE_URL=${API_BASE_URL}"
echo "DEVICE_ID=${DEVICE_ID}"

check GET /health
check GET /api/v1/workbench/summary
check GET /api/v1/analytics/summary
check GET "/api/v1/agent/history?limit=1"
check GET "/api/v1/digital-humans?page=1&page_size=1"
check GET "/api/v1/images?page=1&page_size=1"
check GET "/api/v1/videos?page=1&page_size=1"
check GET "/api/v1/projects?page=1&page_size=1"

PROJECT_ID="$(extract_first_project_id)"
if [[ -n "$PROJECT_ID" ]]; then
	check GET "/api/v1/projects/${PROJECT_ID}"
	check GET "/api/v1/projects/${PROJECT_ID}/script"
	check GET "/api/v1/projects/${PROJECT_ID}/tasks"
	check GET "/api/v1/projects/${PROJECT_ID}/assets"
	check GET "/api/v1/projects/${PROJECT_ID}/timeline"
else
	echo "SKIP [no project] GET /api/v1/projects/:id"
	echo "SKIP [no project] GET /api/v1/projects/:id/script"
	echo "SKIP [no project] GET /api/v1/projects/:id/tasks"
	echo "SKIP [no project] GET /api/v1/projects/:id/assets"
	echo "SKIP [no project] GET /api/v1/projects/:id/timeline"
fi

echo "PASS All required smoke checks completed."
