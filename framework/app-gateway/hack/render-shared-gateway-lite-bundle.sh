#!/usr/bin/env bash
# Render a single YAML bundle for Shared HTTPS-Lite gateway (EG + app-gateway-system, no Linkerd).
# Vendor dev only; human applies the output. No :latest image tags are injected here.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AGW_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
OLARES_ROOT="${OLARES_ROOT:-$(cd "${AGW_ROOT}/../.." && pwd)}"
VENDOR_VALS="${AGW_ROOT}/vendor-charts-values"
SYSTEM_CHART="${AGW_ROOT}/.olares/config/app-gateway-system"
VALUES_LITE="${SYSTEM_CHART}/values-lite.yaml"
EG_CRDS_CHART="${AGW_ROOT}/upstream-charts/envoy-gateway-crds-helm"
EG_CP_CHART="${AGW_ROOT}/upstream-charts/envoy-gateway-helm"
RELEASE_NS="${APP_GATEWAY_NAMESPACE:-app-gateway}"

if ! command -v helm >/dev/null 2>&1; then
  echo "helm is required on PATH" >&2
  exit 1
fi
if [[ ! -f "${VALUES_LITE}" ]]; then
  echo "missing values-lite overlay: ${VALUES_LITE}" >&2
  exit 1
fi

emit() {
  helm template "$@"
}

# 1) Envoy Gateway CRDs
emit eg-crds "${EG_CRDS_CHART}" \
  --namespace "${RELEASE_NS}" \
  -f "${VENDOR_VALS}/envoy-gateway-crds-values.yaml"

echo "---"

# 2) Envoy Gateway control plane
emit eg "${EG_CP_CHART}" \
  --namespace "${RELEASE_NS}" \
  -f "${VENDOR_VALS}/envoy-gateway-values.yaml" \
  --set createNamespace=false

echo "---"

# 3) app-gateway-system (lite values; omit Linkerd subchart and mesh NP templates)
SYSTEM_SHOW_ONLY=(
  -s templates/gateway.yaml
  -s templates/gatewayclass.yaml
  -s templates/envoyproxy.yaml
  -s templates/data-plane-svc.yaml
  -s templates/securitypolicy-ext-authz.yaml
  -s templates/securitypolicy-ext-authz-referencegrant.yaml
  -s charts/envoy-gateway/templates/certgen-rbac.yaml
  -s charts/envoy-gateway/templates/certgen.yaml
  -s charts/envoy-gateway/templates/envoy-gateway-config.yaml
  -s charts/envoy-gateway/templates/envoy-gateway-deployment.yaml
  -s charts/envoy-gateway/templates/envoy-gateway-rbac.yaml
  -s charts/envoy-gateway/templates/envoy-gateway-serviceaccount.yaml
  -s charts/envoy-gateway/templates/envoy-gateway-service.yaml
  -s charts/envoy-gateway/templates/envoy-proxy-topology-injector-webhook.yaml
  -s charts/envoy-gateway/templates/infra-manager-rbac.yaml
  -s charts/envoy-gateway/templates/leader-election-rbac.yaml
)

emit app-gateway-system "${SYSTEM_CHART}" \
  --namespace "${RELEASE_NS}" \
  -f "${VALUES_LITE}" \
  "${SYSTEM_SHOW_ONLY[@]}"
