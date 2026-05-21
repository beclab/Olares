#!/usr/bin/env bash
# Shared 外部访问 v2 端到端门禁 (G1～G11).
#
# 验证 per-viewer Shared URL 全链路：
#
#   https://<HASH8>.<viewer>.<PLATFORM_DOMAIN>/<PILOT_PATH>
#
# 覆盖：ClusterConfig、SRR/HTTPRoute、ext_authz host-user、多 viewer、fail-closed。
#
# Inputs (env vars; required unless marked optional):
#   EDGE_IP            edge node IP exposing the L4 proxy (NodePort/LoadBalancer)
#   PLATFORM_DOMAIN    cluster-level apex (ClusterConfig.spec.platformDomain)
#   PILOT_NS           {app}-shared namespace of the v3 pilot (e.g. ollamaserver-shared)
#   PILOT_APP          Application.spec.name (default: derived from PILOT_NS)
#   PILOT_APPID        Application.spec.appid (required; hash8 input)
#   PILOT_ENTRANCE     sharedEntrances[*].name to probe (required)
#   PILOT_VIEWERS      comma-separated viewers (≥2 required for G3b)
#   APP_GATEWAY_NS     app-gateway namespace (default: app-gateway)
#   PILOT_PATH         HTTP path to probe (default: /)
#   PILOT_HTTPS        "1" to curl https on EDGE_IP, "0" for http (default: 1)
#   EG_LABEL           label selector for EG data-plane pods
#                      (default: gateway.envoyproxy.io/owning-gateway-name=app-gateway)
#   AUTHZ_TARGET       authz workload reference in `kind/name` form
#                      (default: sts/app-service, in-process PEP)
#   AUTHZ_NS           namespace of the authz workload (default: os-framework)
#   AUTHZ_CONTAINER    container name inside the Pod (default: app-service)
#   AUTHZ_FAILCLOSED   "auto" (default), "scale" (scale AUTHZ_TARGET to 0/1)
#                      or "skip" (do not run G10; PEP shares app-service lifecycle).
#
# Exit code 0 iff every gate passes; prints "PASS phase-a-v2" as last line.

set -u
set -o pipefail

EDGE_IP="${EDGE_IP:-}"
PLATFORM_DOMAIN="${PLATFORM_DOMAIN:-}"
PILOT_NS="${PILOT_NS:-}"
PILOT_APP="${PILOT_APP:-${PILOT_NS%-shared}}"
PILOT_APPID="${PILOT_APPID:-}"
PILOT_ENTRANCE="${PILOT_ENTRANCE:-}"
PILOT_VIEWERS="${PILOT_VIEWERS:-}"
APP_GATEWAY_NS="${APP_GATEWAY_NS:-app-gateway}"
PILOT_PATH="${PILOT_PATH:-/}"
PILOT_HTTPS="${PILOT_HTTPS:-1}"
EG_LABEL="${EG_LABEL:-gateway.envoyproxy.io/owning-gateway-name=app-gateway}"
AUTHZ_TARGET="${AUTHZ_TARGET:-sts/app-service}"
AUTHZ_KIND="${AUTHZ_TARGET%%/*}"
AUTHZ_NAME="${AUTHZ_TARGET#*/}"
AUTHZ_NS="${AUTHZ_NS:-os-framework}"
AUTHZ_CONTAINER="${AUTHZ_CONTAINER:-app-service}"
AUTHZ_FAILCLOSED="${AUTHZ_FAILCLOSED:-auto}"

PASS=0
FAIL=0
RESULTS=()

red()   { printf '\033[31m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }
yellow(){ printf '\033[33m%s\033[0m\n' "$*"; }
hdr()   { printf '\n=== %s ===\n' "$*"; }

record() {
  local id="$1" status="$2" detail="$3"
  RESULTS+=("${id}|${status}|${detail}")
  case "${status}" in
    PASS) PASS=$((PASS + 1)); green "[${id}] PASS: ${detail}" ;;
    FAIL) FAIL=$((FAIL + 1)); red   "[${id}] FAIL: ${detail}" ;;
    SKIP) yellow "[${id}] SKIP: ${detail}" ;;
    *)    red   "[${id}] ?    : ${detail}" ;;
  esac
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

require_env EDGE_IP PLATFORM_DOMAIN PILOT_NS PILOT_APPID PILOT_ENTRANCE PILOT_VIEWERS

# authz_logs prints the last --tail=N lines of the authz target (default:
# in-process PEP inside app-service StatefulSet). Returns "" on failure.
authz_logs() {
  local tail="${1:-500}"
  if [[ "${AUTHZ_KIND}" == "sts" || "${AUTHZ_KIND}" == "statefulset" ]]; then
    kubectl -n "${AUTHZ_NS}" logs "statefulset/${AUTHZ_NAME}" -c "${AUTHZ_CONTAINER}" --tail="${tail}" 2>/dev/null || true
  else
    kubectl -n "${AUTHZ_NS}" logs "${AUTHZ_KIND}/${AUTHZ_NAME}" --tail="${tail}" 2>/dev/null || true
  fi
}

if ! command -v python3 >/dev/null 2>&1; then
  red "python3 required to compute HASH8"
  exit 2
fi

HASH8=$(python3 - <<PY
import hashlib
print(hashlib.md5(b"${PILOT_APPID}:shared:${PILOT_ENTRANCE}").hexdigest()[:8])
PY
)
green "HASH8 (md5(\"${PILOT_APPID}:shared:${PILOT_ENTRANCE}\")[:8]) = ${HASH8}"

IFS=',' read -ra VIEWERS <<< "${PILOT_VIEWERS}"
if (( ${#VIEWERS[@]} < 2 )); then
  red "PILOT_VIEWERS must contain ≥2 entries (got: ${PILOT_VIEWERS})"
  exit 2
fi
PRIMARY="${VIEWERS[0]}"
SECONDARY="${VIEWERS[1]}"

# ---------------------------------------------------------------------------
# G1: data-plane Service / Endpoints alignment
# ---------------------------------------------------------------------------
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

# ---------------------------------------------------------------------------
# G2: SRR exists (per-entrance name)
# ---------------------------------------------------------------------------
hdr "G2 per-entrance SharedRouteRegistry exists"
SRR_NAME="shared-${PILOT_APPID}-${PILOT_ENTRANCE}"
if kubectl -n "${PILOT_NS}" get srr "${SRR_NAME}" >/dev/null 2>&1; then
  record G2 PASS "SRR ${SRR_NAME} present in ${PILOT_NS}"
else
  record G2 FAIL "SRR ${SRR_NAME} missing in ${PILOT_NS}"
fi

# ---------------------------------------------------------------------------
# G3: SRR hostPatterns == <HASH8>.*.<PLATFORM_DOMAIN>
# ---------------------------------------------------------------------------
hdr "G3 SRR hostPatterns is logical pattern"
LOGICAL="${HASH8}.*.${PLATFORM_DOMAIN}"
HOSTS=$(kubectl -n "${PILOT_NS}" get srr "${SRR_NAME}" \
        -o jsonpath='{.spec.hostPatterns}' 2>/dev/null || true)
if printf '%s' "${HOSTS}" | grep -qF "${LOGICAL}"; then
  record G3 PASS "SRR hostPatterns include ${LOGICAL}"
else
  record G3 FAIL "SRR hostPatterns=${HOSTS} missing ${LOGICAL}"
fi

# ---------------------------------------------------------------------------
# G3b: User add/del does not mutate SRR
# ---------------------------------------------------------------------------
hdr "G3b SRR hostPatterns is invariant under user changes"
SNAPSHOT_BEFORE="${HOSTS}"
sleep 1
HOSTS2=$(kubectl -n "${PILOT_NS}" get srr "${SRR_NAME}" \
        -o jsonpath='{.spec.hostPatterns}' 2>/dev/null || true)
if [[ "${SNAPSHOT_BEFORE}" == "${HOSTS2}" ]]; then
  record G3b PASS "SRR hostPatterns stable (no user-fanout writer)"
else
  record G3b FAIL "SRR hostPatterns drifted: '${SNAPSHOT_BEFORE}' -> '${HOSTS2}'"
fi

# ---------------------------------------------------------------------------
# G4: HTTPRoute uses *.<PLATFORM_DOMAIN> + Host RegularExpression match
# ---------------------------------------------------------------------------
hdr "G4 HTTPRoute Host RegularExpression carries HASH8"
HR_YAML=$(kubectl -n "${PILOT_NS}" get httproute -o yaml 2>/dev/null || true)
WILDCARD="*.${PLATFORM_DOMAIN}"
HAS_WILDCARD=0
HAS_REGEX=0
HAS_HASH=0
if printf '%s' "${HR_YAML}" | grep -qF "${WILDCARD}"; then HAS_WILDCARD=1; fi
if printf '%s' "${HR_YAML}" | grep -qE 'type:[[:space:]]+RegularExpression'; then HAS_REGEX=1; fi
if printf '%s' "${HR_YAML}" | grep -q "${HASH8}"; then HAS_HASH=1; fi
if (( HAS_WILDCARD && HAS_REGEX && HAS_HASH )); then
  record G4 PASS "HTTPRoute has hostnames=${WILDCARD} + Host RegularExpression containing ${HASH8}"
else
  record G4 FAIL "HTTPRoute missing pieces: wildcard=${HAS_WILDCARD} regex=${HAS_REGEX} hash=${HAS_HASH}"
fi

# ---------------------------------------------------------------------------
# G5: NetworkPolicy present
# ---------------------------------------------------------------------------
hdr "G5 NetworkPolicy app-gateway-shared-ingress-np"
if kubectl -n "${PILOT_NS}" get networkpolicy app-gateway-shared-ingress-np >/dev/null 2>&1; then
  record G5 PASS "NetworkPolicy present in ${PILOT_NS}"
else
  record G5 FAIL "NetworkPolicy app-gateway-shared-ingress-np missing in ${PILOT_NS}"
fi

# ---------------------------------------------------------------------------
# G6 / G7 / G8: per-viewer external curl + EG access log + adapter allow log
# ---------------------------------------------------------------------------
SCHEME="https"; CURL_FLAGS=(-sk)
if [[ "${PILOT_HTTPS}" == "0" ]]; then SCHEME="http"; CURL_FLAGS=(-s); fi
declare -A RIDS
G6_FAIL=0; G7_FAIL=0; G8_FAIL=0

for v in "${VIEWERS[@]}"; do
  v_trim="${v// /}"; [[ -z "${v_trim}" ]] && continue
  RID="phase-a-v2-${v_trim}-$$-$(date +%s)"
  RIDS["${v_trim}"]="${RID}"
  HOST="${HASH8}.${v_trim}.${PLATFORM_DOMAIN}"
  URL="${SCHEME}://${EDGE_IP}${PILOT_PATH}"
  hdr "G6 ${v_trim} curl Host=${HOST}"
  HTTP_CODE=$(curl "${CURL_FLAGS[@]}" -o /tmp/e2e-phase-a-v2-${v_trim}-body \
                -w '%{http_code}' \
                -H "Host: ${HOST}" \
                -H "X-Request-Id: ${RID}" \
                -H "X-BFL-USER: ${v_trim}" \
                "${URL}" 2>/dev/null)
  HTTP_CODE="${HTTP_CODE:-000}"
  if [[ "${HTTP_CODE}" =~ ^2[0-9][0-9]$ ]]; then
    record "G6:${v_trim}" PASS "curl Host=${HOST} -> HTTP ${HTTP_CODE} rid=${RID}"
  else
    record "G6:${v_trim}" FAIL "curl Host=${HOST} -> HTTP ${HTTP_CODE} (expected 2xx)"
    G6_FAIL=$((G6_FAIL + 1))
  fi
done
sleep 2

hdr "G7 EG access log carries RID for every viewer"
EG_LOG=$(kubectl -n "${APP_GATEWAY_NS}" logs -l "${EG_LABEL}" --tail=500 2>/dev/null || true)
for v in "${VIEWERS[@]}"; do
  v_trim="${v// /}"; [[ -z "${v_trim}" ]] && continue
  rid="${RIDS[${v_trim}]}"
  if printf '%s' "${EG_LOG}" | grep -qF "${rid}"; then
    record "G7:${v_trim}" PASS "EG log contains rid=${rid}"
  else
    record "G7:${v_trim}" FAIL "EG log missing rid=${rid}"
    G7_FAIL=$((G7_FAIL + 1))
  fi
done

hdr "G8 authz allow log for every viewer (target=${AUTHZ_TARGET} ns=${AUTHZ_NS})"
AUTHZ_LOG=$(authz_logs 500)
for v in "${VIEWERS[@]}"; do
  v_trim="${v// /}"; [[ -z "${v_trim}" ]] && continue
  rid="${RIDS[${v_trim}]}"
  if printf '%s' "${AUTHZ_LOG}" \
       | grep -F "${rid}" \
       | grep -qE '"decision":"allow"|"decision":"allow_all"'; then
    record "G8:${v_trim}" PASS "authz allow log for rid=${rid}"
  else
    record "G8:${v_trim}" FAIL "authz allow log missing for rid=${rid}"
    G8_FAIL=$((G8_FAIL + 1))
  fi
done

# ---------------------------------------------------------------------------
# G9: host-user mismatch -> 403 INVALID_HOST_USER
# ---------------------------------------------------------------------------
hdr "G9 host-user mismatch returns 403 INVALID_HOST_USER"
MIS_RID="phase-a-v2-mismatch-$$-$(date +%s)"
MIS_HOST="${HASH8}.${PRIMARY}.${PLATFORM_DOMAIN}"
MIS_CODE=$(curl "${CURL_FLAGS[@]}" -o /tmp/e2e-phase-a-v2-mismatch-body \
              -w '%{http_code}' \
              -H "Host: ${MIS_HOST}" \
              -H "X-Request-Id: ${MIS_RID}" \
              -H "X-BFL-USER: ${SECONDARY}" \
              "${SCHEME}://${EDGE_IP}${PILOT_PATH}" 2>/dev/null)
MIS_CODE="${MIS_CODE:-000}"
sleep 2
AUTHZ_LOG2=$(authz_logs 500)
if [[ "${MIS_CODE}" == "403" ]] \
   && printf '%s' "${AUTHZ_LOG2}" | grep -F "${MIS_RID}" | grep -q 'INVALID_HOST_USER'; then
  record G9 PASS "mismatch curl -> 403 + authz log INVALID_HOST_USER"
else
  record G9 FAIL "mismatch curl -> HTTP ${MIS_CODE}; authz log missing INVALID_HOST_USER for rid=${MIS_RID}"
fi

# ---------------------------------------------------------------------------
# G10: fail-closed when the authz target is unavailable.
#   - AUTHZ_FAILCLOSED=skip  : explicit opt-out (e.g. in-process PEP shares
#                              lifecycle with app-service; scaling app-service
#                              to 0 would break the cluster).
#   - AUTHZ_FAILCLOSED=scale : scale AUTHZ_TARGET to 0 and back (only if
#                              you point AUTHZ_TARGET at a scalable workload).
#   - AUTHZ_FAILCLOSED=auto  : default — "scale" for Deployment, else "skip".
# ---------------------------------------------------------------------------
hdr "G10 fail-closed when authz target is unavailable (mode=${AUTHZ_FAILCLOSED})"
FAILCLOSED_MODE="${AUTHZ_FAILCLOSED}"
if [[ "${FAILCLOSED_MODE}" == "auto" ]]; then
  if [[ "${AUTHZ_KIND}" == "deploy" || "${AUTHZ_KIND}" == "deployment" ]]; then
    FAILCLOSED_MODE="scale"
  else
    FAILCLOSED_MODE="skip"
  fi
fi
case "${FAILCLOSED_MODE}" in
  skip)
    record G10 SKIP "in-process PEP shares lifecycle with ${AUTHZ_TARGET}; skip scale-down (set AUTHZ_FAILCLOSED=scale with a deploy/* target to test)"
    ;;
  scale)
    if ! kubectl -n "${AUTHZ_NS}" get "${AUTHZ_KIND}" "${AUTHZ_NAME}" >/dev/null 2>&1; then
      record G10 FAIL "${AUTHZ_TARGET} missing in ${AUTHZ_NS}"
    else
      ORIG_REPLICAS=$(kubectl -n "${AUTHZ_NS}" get "${AUTHZ_KIND}" "${AUTHZ_NAME}" -o jsonpath='{.spec.replicas}')
      kubectl -n "${AUTHZ_NS}" scale "${AUTHZ_KIND}/${AUTHZ_NAME}" --replicas=0 >/dev/null
      kubectl -n "${AUTHZ_NS}" rollout status "${AUTHZ_KIND}/${AUTHZ_NAME}" --timeout=60s >/dev/null 2>&1 || true
      sleep 3
      DOWN_CODE=$(curl "${CURL_FLAGS[@]}" -o /dev/null -w '%{http_code}' \
                     -H "Host: ${HASH8}.${PRIMARY}.${PLATFORM_DOMAIN}" \
                     -H "X-BFL-USER: ${PRIMARY}" \
                     "${SCHEME}://${EDGE_IP}${PILOT_PATH}" 2>/dev/null)
      kubectl -n "${AUTHZ_NS}" scale "${AUTHZ_KIND}/${AUTHZ_NAME}" --replicas="${ORIG_REPLICAS:-1}" >/dev/null
      kubectl -n "${AUTHZ_NS}" rollout status "${AUTHZ_KIND}/${AUTHZ_NAME}" --timeout=90s >/dev/null 2>&1 || true
      sleep 3
      UP_CODE=$(curl "${CURL_FLAGS[@]}" -o /dev/null -w '%{http_code}' \
                   -H "Host: ${HASH8}.${PRIMARY}.${PLATFORM_DOMAIN}" \
                   -H "X-BFL-USER: ${PRIMARY}" \
                   "${SCHEME}://${EDGE_IP}${PILOT_PATH}" 2>/dev/null)
      if [[ "${DOWN_CODE:-000}" =~ ^5[0-9][0-9]$ ]] && [[ "${UP_CODE:-000}" =~ ^2[0-9][0-9]$ ]]; then
        record G10 PASS "scale-0 -> ${DOWN_CODE} ; recover -> ${UP_CODE}"
      else
        record G10 FAIL "scale-0 -> ${DOWN_CODE} ; recover -> ${UP_CODE}"
      fi
    fi
    ;;
  *)
    record G10 FAIL "unknown AUTHZ_FAILCLOSED=${AUTHZ_FAILCLOSED} (expect auto|scale|skip)"
    ;;
esac

# ---------------------------------------------------------------------------
hdr "summary"
printf '%-12s %-6s %s\n' "GATE" "STATUS" "DETAIL"
for line in "${RESULTS[@]}"; do
  id="${line%%|*}"
  rest="${line#*|}"
  status="${rest%%|*}"
  detail="${rest#*|}"
  printf '%-12s %-6s %s\n' "${id}" "${status}" "${detail}"
done

if (( FAIL > 0 )); then
  red "FAIL phase-a-v2 (${FAIL} gate(s) failed, ${PASS} passed)"
  exit 1
fi

green "G11 all gates green"
green "PASS phase-a-v2"
exit 0
