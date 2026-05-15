#!/usr/bin/env bash
# Maintainer/build-time: bundle third-party charts into Olares installer tree (read-only copy from 3rd/gateway).
# Does NOT modify 3rd/gateway source.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_GATEWAY_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
AIDEV_ROOT="$(cd "${APP_GATEWAY_ROOT}/../../.." && pwd)"
GATEWAY_3RD="${AIDEV_ROOT}/3rd/gateway/charts"
OUT="${APP_GATEWAY_ROOT}/vendor-charts"

LINKERD_CHART_VERSION="${LINKERD_CHART_VERSION:-1.16.11}"

rm -rf "${OUT}"
mkdir -p "${OUT}"

echo "==> Copy Envoy Gateway charts from ${GATEWAY_3RD}"
[[ -d "${GATEWAY_3RD}/gateway-crds-helm" ]] || { echo "missing ${GATEWAY_3RD}/gateway-crds-helm"; exit 1; }
[[ -d "${GATEWAY_3RD}/gateway-helm" ]] || { echo "missing ${GATEWAY_3RD}/gateway-helm"; exit 1; }
cp -a "${GATEWAY_3RD}/gateway-crds-helm" "${OUT}/envoy-gateway-crds-helm"
cp -a "${GATEWAY_3RD}/gateway-helm" "${OUT}/envoy-gateway-helm"

echo "==> Pull Linkerd control-plane chart (helm pull, build-time only)"
if command -v helm >/dev/null 2>&1; then
  helm repo add linkerd https://helm.linkerd.io/stable 2>/dev/null || true
  helm repo update linkerd 2>/dev/null || helm repo update || true
  helm pull linkerd/linkerd-control-plane --version "${LINKERD_CHART_VERSION}" \
    --destination "${OUT}" --untar
  mv "${OUT}/linkerd-control-plane" "${OUT}/linkerd-control-plane-chart" 2>/dev/null || true
else
  echo "WARN: helm not found; skip linkerd chart pull. Run with helm before release build." >&2
  mkdir -p "${OUT}/linkerd-control-plane-chart"
  echo "# run: helm pull linkerd/linkerd-control-plane --version ${LINKERD_CHART_VERSION}" > "${OUT}/linkerd-control-plane-chart/README"
fi

cp -f "${APP_GATEWAY_ROOT}/vendor-charts-values/envoy-gateway-values.yaml" "${OUT}/envoy-gateway-values.yaml" 2>/dev/null || \
  cp -f "${AIDEV_ROOT}/devops/dev/platform-gateway/values/envoy-gateway-values.yaml" "${OUT}/envoy-gateway-values.yaml"

cp -f "${APP_GATEWAY_ROOT}/vendor-charts-values/linkerd-values.yaml" "${OUT}/linkerd-values.yaml" 2>/dev/null || \
  cat > "${OUT}/linkerd-values.yaml" <<'EOF'
# Linkerd control-plane minimal PoC values (install via Olares cli helm SDK)
identity:
  issuer:
    scheme: kubernetes.io/tls
EOF

echo "OK: vendor charts at ${OUT}"
