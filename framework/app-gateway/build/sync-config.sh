#!/usr/bin/env bash
# Sync config/defaults.yaml -> Helm values + DevOps env snippet (single source of truth).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULTS="${ROOT}/config/defaults.yaml"
VALUES="${ROOT}/.olares/config/user/helm-charts/app-gateway/values.yaml"
DEVOPS_ENV="${ROOT}/../../../devops/dev/platform-gateway/generated/app-gateway.env"

if ! command -v yq >/dev/null 2>&1; then
  echo "ERROR: yq required (https://github.com/mikefarah/yq)" >&2
  exit 1
fi

NS="$(yq -r '.namespace' "${DEFAULTS}")"
GW_NAME="$(yq -r '.gateway.name' "${DEFAULTS}")"
GW_CLASS="$(yq -r '.gateway.gatewayClassName' "${DEFAULTS}")"
DEMO_HOST="$(yq -r '.demo.host' "${DEFAULTS}")"
DEMO_ENABLED="$(yq -r '.demo.enabled' "${DEFAULTS}")"
LINKERD_NS="$(yq -r '.vendor.linkerdNamespace' "${DEFAULTS}")"

cat > "${VALUES}" <<EOF
# AUTO-GENERATED from config/defaults.yaml — do not edit namespace here; run: make sync-config
namespace: ${NS}

gateway:
  name: ${GW_NAME}
  gatewayClassName: ${GW_CLASS}

demo:
  enabled: ${DEMO_ENABLED}
  host: ${DEMO_HOST}
EOF

mkdir -p "$(dirname "${DEVOPS_ENV}")"
cat > "${DEVOPS_ENV}" <<EOF
# AUTO-GENERATED from Olares/framework/app-gateway/config/defaults.yaml
APP_GATEWAY_NAMESPACE=${NS}
GATEWAY_NS=${NS}
GATEWAY_NAME=${GW_NAME}
AGW_DEMO_NS=${NS}
DEMO_HOST=${DEMO_HOST}
ENVOY_GATEWAY_NAMESPACE=${NS}
LINKERD_NAMESPACE=${LINKERD_NS}
EOF

echo "OK: synced defaults -> ${VALUES}"
echo "OK: synced defaults -> ${DEVOPS_ENV}"
