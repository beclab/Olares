#!/usr/bin/env bash
# Optional: copy Olares-owned values into the committed vendor tree.
# Does NOT fetch charts — chart directories must already exist under
# .olares/config/app-gateway-vendor/ (same as infrastructure/gpu/.olares/config/gpu/).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AGW_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
OUT="${AGW_ROOT}/.olares/config/app-gateway-vendor"
VALS_SRC="${AGW_ROOT}/vendor-charts-values"

for f in envoy-gateway-values.yaml linkerd-values.yaml linkerd-crds-values.yaml; do
  cp -f "${VALS_SRC}/${f}" "${OUT}/${f}"
done

echo "OK: synced values into ${OUT}"
