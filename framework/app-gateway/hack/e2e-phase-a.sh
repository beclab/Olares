#!/usr/bin/env bash
# Shared external access Phase-A end-to-end gate (G1-G8): L4 -> EG -> ext_authz -> HTTPRoute -> backend.
#
# Inputs (env vars; required unless marked optional):
#   EDGE_IP          edge node IP exposing the L4 proxy (NodePort/LoadBalancer)
#   SHARED_HOST      lowercase no-port shared host produced by GenSharedEntranceURL
#   PILOT_NS         {app}-shared namespace of the v3 pilot (e.g. ollamav2-shared)
#   PILOT_APP        manifest metadata.name of the pilot app (default: derived from PILOT_NS)
#   APP_GATEWAY_NS   app-gateway namespace (default: app-gateway)
#   PILOT_PATH       HTTP path to probe (default: /)
#   PILOT_HTTPS      "1" to curl https on EDGE_IP, "0" for http (default: 1)
#   EG_LABEL         label selector for EG data-plane pods
#                    (default: gateway.envoyproxy.io/owning-gateway-name=app-gateway)
#   AUTHZ_NS         namespace of app-service (default: os-framework)
#   AUTHZ_STS        StatefulSet name (default: app-service)
#
# Exit code 0 iff every gate passes; prints "PASS phase-a" as last line.

set -u
set -o pipefail

EDGE_IP="${EDGE_IP:-}"
SHARED_HOST="${SHARED_HOST:-}"
PILOT_NS="${PILOT_NS:-}"
PILOT_APP="${PILOT_APP:-${PILOT_NS%-shared}}"
APP_GATEWAY_NS="${APP_GATEWAY_NS:-app-gateway}"
PILOT_PATH="${PILOT_PATH:-/}"
PILOT_HTTPS="${PILOT_HTTPS:-1}"
EG_LABEL="${EG_LABEL:-gateway.envoyproxy.io/owning-gateway-name=app-gateway}"
AUTHZ_NS="${AUTHZ_NS:-os-framework}"
AUTHZ_STS="${AUTHZ_STS:-app-service}"

PASS=0
FAIL=0
RESULTS=()

red()   { printf '\033[31m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }
hdr()   { printf '\n=== %s ===\n' "$*"; }

record() {
  local id="$1" status="$2" detail="$3"
  RESULTS+=("${id}|${status}|${detail}")
  if [[ "${status}" == "PASS" ]]; then
    PASS=$((PASS + 1))
    green "[${id}] PASS: ${detail}"
  else
    FAIL=$((FAIL + 1))
    red   "[${id}] FAIL: ${detail}"
  fi
}

require_env() {
  local missing=()
  for v in "$@"; do
    if [[ -z "${!v:-}" ]]; then missing+=("$v"); fi
  done
  if ((${#missing[@]} > 0)); then
    red "missing required env: ${missing[*]}"
    exit 2
  fi
}

require_env EDGE_IP SHARED_HOST PILOT_NS

hdr "G1 app-gateway-data endpoints"
if EP_DATA=$(kubectl -n "${APP_GATEWAY_NS}" get endpoints app-gateway-data \
      -o jsonpath='{.subsets[0].addresses[0].ip}' 2>/dev/null) \
   && [[ -n "${EP_DATA}" ]]; then
  if EP_ENVOY=$(kubectl -n "${APP_GATEWAY_NS}" get endpoints -l "${EG_LABEL}" \
        -o jsonpath='{.items[0].subsets[0].addresses[0].ip}' 2>/dev/null) \
     && [[ -n "${EP_ENVOY}" ]]; then
    if [[ "${EP_DATA}" == "${EP_ENVOY}" ]]; then
      record G1 PASS "app-gateway-data EP=${EP_DATA} == EG EP=${EP_ENVOY}"
    else
      record G1 FAIL "app-gateway-data EP=${EP_DATA} != EG EP=${EP_ENVOY}"
    fi
  else
    record G1 FAIL "EG data-plane endpoints not found (label=${EG_LABEL})"
  fi
else
  record G1 FAIL "Service/Endpoints app-gateway-data missing in ${APP_GATEWAY_NS}"
fi

hdr "G2 SharedRouteRegistry"
SRR_LIST=$(kubectl -n "${PILOT_NS}" get sharedrouteregistry -o name 2>/dev/null || true)
if [[ -z "${SRR_LIST}" ]]; then
  record G2 FAIL "no SharedRouteRegistry in ${PILOT_NS}"
else
  SRR_NAME=$(printf '%s' "${SRR_LIST}" | head -n1 | awk -F/ '{print $2}')
  HOSTS=$(kubectl -n "${PILOT_NS}" get srr "${SRR_NAME}" \
    -o jsonpath='{.spec.hostPatterns}' 2>/dev/null || true)
  if printf '%s' "${HOSTS}" | grep -qi -F "${SHARED_HOST}"; then
    record G2 PASS "SRR ${SRR_NAME} hostPatterns include ${SHARED_HOST}"
  else
    record G2 FAIL "SRR ${SRR_NAME} hostPatterns=${HOSTS} missing ${SHARED_HOST}"
  fi
fi

hdr "G3 HTTPRoute in pilot namespace"
HR_LIST=$(kubectl -n "${PILOT_NS}" get httproute -o name 2>/dev/null || true)
if [[ -z "${HR_LIST}" ]]; then
  record G3 FAIL "no HTTPRoute in ${PILOT_NS}"
else
  HR_HOSTS=$(kubectl -n "${PILOT_NS}" get httproute -o jsonpath='{range .items[*]}{.spec.hostnames}{"\n"}{end}' 2>/dev/null)
  if printf '%s' "${HR_HOSTS}" | grep -qi -F "${SHARED_HOST}"; then
    record G3 PASS "HTTPRoute hostnames include ${SHARED_HOST}"
  else
    record G3 FAIL "HTTPRoute hostnames=${HR_HOSTS} missing ${SHARED_HOST}"
  fi
fi

hdr "G4 NetworkPolicy for shared ingress"
if kubectl -n "${PILOT_NS}" get networkpolicy app-gateway-shared-ingress-np >/dev/null 2>&1; then
  record G4 PASS "NetworkPolicy app-gateway-shared-ingress-np present in ${PILOT_NS}"
else
  record G4 FAIL "NetworkPolicy app-gateway-shared-ingress-np missing in ${PILOT_NS}"
fi

hdr "G5 external curl"
SCHEME="https"; CURL_FLAGS=(-sk)
if [[ "${PILOT_HTTPS}" == "0" ]]; then SCHEME="http"; CURL_FLAGS=(-s); fi
URL="${SCHEME}://${EDGE_IP}${PILOT_PATH}"
HTTP_CODE=$(curl "${CURL_FLAGS[@]}" -o /tmp/e2e-phase-a-body -w '%{http_code}' \
              -H "Host: ${SHARED_HOST}" "${URL}" 2>/dev/null)
HTTP_CODE="${HTTP_CODE:-000}"
if [[ "${HTTP_CODE}" =~ ^2[0-9][0-9]$ ]]; then
  record G5 PASS "curl ${URL} -> HTTP ${HTTP_CODE}"
else
  record G5 FAIL "curl ${URL} -> HTTP ${HTTP_CODE} (expected 2xx)"
fi

hdr "G6 EG access log shows host"
if kubectl -n "${APP_GATEWAY_NS}" logs -l "${EG_LABEL}" --tail=200 2>/dev/null \
     | grep -qi -F "${SHARED_HOST}"; then
  record G6 PASS "EG log contains ${SHARED_HOST}"
else
  record G6 FAIL "EG log missing ${SHARED_HOST} (tail=200)"
fi

hdr "G7 app-service ext_authz allow log (optional)"
if kubectl -n "${AUTHZ_NS}" get sts "${AUTHZ_STS}" >/dev/null 2>&1; then
  if kubectl -n "${AUTHZ_NS}" logs "sts/${AUTHZ_STS}" -c app-service --tail=100 2>/dev/null \
       | grep -qi 'allow'; then
    record G7 PASS "${AUTHZ_STS} log shows allow decision"
  else
    record G7 PASS "${AUTHZ_STS} present but no recent allow log (non-fatal)"
  fi
else
  record G7 FAIL "sts/${AUTHZ_STS} missing in ${AUTHZ_NS}"
fi

hdr "summary"
printf '%-4s %-6s %s\n' "GATE" "STATUS" "DETAIL"
for line in "${RESULTS[@]}"; do
  id="${line%%|*}"
  rest="${line#*|}"
  status="${rest%%|*}"
  detail="${rest#*|}"
  printf '%-4s %-6s %s\n' "${id}" "${status}" "${detail}"
done

if (( FAIL > 0 )); then
  red "FAIL phase-a (${FAIL} gate(s) failed, ${PASS} passed)"
  exit 1
fi

green "G8 all gates green"
green "PASS phase-a"
exit 0
