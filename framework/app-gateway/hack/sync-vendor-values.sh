#!/usr/bin/env bash
# Optional: copy Olares-owned values into the committed vendor tree.
# Does NOT fetch charts — chart directories must already exist under
# .olares/config/app-gateway-vendor/ (same as infrastructure/gpu/.olares/config/gpu/).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AGW_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
OUT="${AGW_ROOT}/.olares/config/app-gateway-vendor"
VALS_SRC="${AGW_ROOT}/vendor-charts-values"

for f in envoy-gateway-values.yaml envoy-gateway-crds-values.yaml linkerd-values.yaml linkerd-crds-values.yaml linkerd-viz-values.yaml; do
  cp -f "${VALS_SRC}/${f}" "${OUT}/${f}"
done

# Installer-only assets (PKI script + bootstrap mesh NP); charts stay unchanged.
cp -f "${SCRIPT_DIR}/generate-linkerd-identity-certs.sh" "${OUT}/generate-linkerd-identity-certs.sh"
chmod 755 "${OUT}/generate-linkerd-identity-certs.sh"
mkdir -p "${OUT}/network-policies"
cp -f "${AGW_ROOT}/deploy/network-policies/linkerd-mesh-ingress.yaml" \
  "${OUT}/network-policies/linkerd-mesh-ingress.yaml"
mkdir -p "${OUT}/deploy/linkerd"
cp -f "${AGW_ROOT}/deploy/linkerd/prometheus-pod-proxy-rbac.yaml" \
  "${OUT}/deploy/linkerd/prometheus-pod-proxy-rbac.yaml"

echo "OK: synced values and installer assets into ${OUT}"
