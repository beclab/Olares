#!/usr/bin/env bash
# Verify EG data-plane Linkerd mesh (release path). Fails on missing linkerd-proxy or namespace inject.
set -euo pipefail

NS="${NS:-app-gateway}"
GW_NAME="${GW_NAME:-app-gateway}"
ENVOY_PROXY_NAME="${ENVOY_PROXY_NAME:-app-gateway-envoy-proxy}"
MESH_LINKERD_ENABLED="${MESH_LINKERD_ENABLED:-true}"
LINKERD_NS="${LINKERD_NS:-linkerd}"

fail() { echo "FAIL: $*" >&2; exit 1; }
warn() { echo "WARN: $*" >&2; }

echo "== Namespace (must not have linkerd.io/inject) =="
ns_inject="$(kubectl get ns "${NS}" -o jsonpath='{.metadata.annotations.linkerd\.io/inject}' 2>/dev/null || true)"
if [[ -n "${ns_inject}" ]]; then
  fail "namespace ${NS} has linkerd.io/inject=${ns_inject} (EG mesh uses EnvoyProxy pod template only)"
fi
echo "OK: no namespace-level linkerd.io/inject"

echo "== EnvoyProxy =="
if ! kubectl -n "${NS}" get envoyproxy "${ENVOY_PROXY_NAME}" >/dev/null 2>&1; then
  kubectl -n "${NS}" get envoyproxy -o name 2>/dev/null || fail "EnvoyProxy ${ENVOY_PROXY_NAME} not found in ${NS}"
else
  kubectl -n "${NS}" get envoyproxy "${ENVOY_PROXY_NAME}" -o name
fi

if [[ "${MESH_LINKERD_ENABLED}" == "true" ]]; then
  inject="$(kubectl -n "${NS}" get envoyproxy "${ENVOY_PROXY_NAME}" -o jsonpath='{.spec.provider.kubernetes.envoyDeployment.pod.annotations.linkerd\.io/inject}' 2>/dev/null || true)"
  if [[ "${inject}" != "enabled" ]]; then
    fail "EnvoyProxy ${ENVOY_PROXY_NAME} linkerd.io/inject=${inject:-<unset>} (want enabled)"
  fi
  echo "OK: EnvoyProxy pod inject enabled"
fi

echo "== EG data-plane pods (expect envoy + linkerd-proxy when mesh enabled) =="
mapfile -t dp_pods < <(kubectl -n "${NS}" get pods -l "gateway.envoyproxy.io/owning-gateway-name=${GW_NAME}" -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')
[[ ${#dp_pods[@]} -gt 0 ]] || fail "no EG data-plane pods for gateway ${GW_NAME} in ${NS}"

kubectl -n "${NS}" get pods -l "gateway.envoyproxy.io/owning-gateway-name=${GW_NAME}" \
  -o custom-columns=NAME:.metadata.name,CONTAINERS:.spec.containers[*].name,READY:.status.containerStatuses[*].ready

for pod in "${dp_pods[@]}"; do
  [[ -n "${pod}" ]] || continue
  phase="$(kubectl -n "${NS}" get pod "${pod}" -o jsonpath='{.status.phase}')"
  [[ "${phase}" != "Failed" ]] || fail "pod ${pod} phase=Failed"

  if [[ "${MESH_LINKERD_ENABLED}" == "true" ]]; then
    has_proxy="$(kubectl -n "${NS}" get pod "${pod}" -o jsonpath='{.status.containerStatuses[?(@.name=="linkerd-proxy")].name}' 2>/dev/null || true)"
    [[ -n "${has_proxy}" ]] || fail "pod ${pod} missing linkerd-proxy container"
    proxy_ready="$(kubectl -n "${NS}" get pod "${pod}" -o jsonpath='{.status.containerStatuses[?(@.name=="linkerd-proxy")].ready}' 2>/dev/null || true)"
    [[ "${proxy_ready}" == "true" ]] || fail "pod ${pod} linkerd-proxy not Ready"
  fi

  has_envoy="$(kubectl -n "${NS}" get pod "${pod}" -o jsonpath='{.status.containerStatuses[?(@.name=="envoy")].name}' 2>/dev/null || true)"
  [[ -n "${has_envoy}" ]] || fail "pod ${pod} missing envoy container"
  envoy_ready="$(kubectl -n "${NS}" get pod "${pod}" -o jsonpath='{.status.containerStatuses[?(@.name=="envoy")].ready}' 2>/dev/null || true)"
  [[ "${envoy_ready}" == "true" ]] || fail "pod ${pod} envoy container not Ready"
done
echo "OK: ${#dp_pods[@]} data-plane pod(s) mesh-ready"

echo "== linkerd viz stat (optional) =="
if command -v linkerd >/dev/null 2>&1; then
  if command -v timeout >/dev/null 2>&1; then
    timeout 15s linkerd viz stat deploy -n "${NS}" 2>/dev/null \
      || timeout 15s linkerd viz stat pods -n "${NS}" 2>/dev/null \
      || warn "linkerd viz stat unavailable (timeout or no metrics)"
  else
    linkerd viz stat deploy -n "${NS}" 2>/dev/null || linkerd viz stat pods -n "${NS}" 2>/dev/null || warn "linkerd viz stat unavailable"
  fi
else
  warn "linkerd CLI not installed; skip viz stat"
fi

echo "== mesh NetworkPolicy (bootstrap / app-service) =="
kubectl -n "${NS}" get networkpolicy app-gateway-mesh-np -o name 2>/dev/null || warn "app-gateway-mesh-np missing in ${NS}"
kubectl -n "${LINKERD_NS}" get networkpolicy app-gateway-mesh-np -o name 2>/dev/null || warn "app-gateway-mesh-np missing in ${LINKERD_NS}"

echo "== stable data-plane Service (F-1: app-gateway-data) =="
DATA_SVC="${DATA_SVC:-app-gateway-data}"
if ! kubectl -n "${NS}" get svc "${DATA_SVC}" >/dev/null 2>&1; then
  fail "Service ${NS}/${DATA_SVC} missing (PR-1: data-plane-svc.yaml not applied)"
fi
data_target_port="$(kubectl -n "${NS}" get svc "${DATA_SVC}" -o jsonpath='{.spec.ports[?(@.port==80)].targetPort}')"
[[ "${data_target_port}" == "10080" ]] || fail "Service ${DATA_SVC} port 80 targetPort=${data_target_port:-<unset>} (want 10080)"
mapfile -t data_eps < <(kubectl -n "${NS}" get endpoints "${DATA_SVC}" -o jsonpath='{range .subsets[*].addresses[*]}{.ip}{"\n"}{end}' | sort -u)
[[ ${#data_eps[@]} -gt 0 ]] || fail "Endpoints ${DATA_SVC} has no addresses (selector mismatch or EG data-plane not Ready)"
mapfile -t envoy_eps < <(kubectl -n "${NS}" get endpoints -l gateway.envoyproxy.io/owning-gateway-name="${GW_NAME}" -o jsonpath='{range .items[*].subsets[*].addresses[*]}{.ip}{"\n"}{end}' | sort -u)
[[ ${#envoy_eps[@]} -gt 0 ]] || fail "EG data-plane endpoints (label gateway.envoyproxy.io/owning-gateway-name=${GW_NAME}) empty"
diff_lines="$(comm -3 <(printf '%s\n' "${data_eps[@]}") <(printf '%s\n' "${envoy_eps[@]}") || true)"
[[ -z "${diff_lines}" ]] || fail "Endpoints ${DATA_SVC} differ from EG data-plane: ${diff_lines}"
echo "OK: ${DATA_SVC} endpoints (${#data_eps[@]}) match EG data-plane"

echo "PASS: verify-app-gateway-mesh (${NS}, gateway=${GW_NAME}, mesh=${MESH_LINKERD_ENABLED})"
