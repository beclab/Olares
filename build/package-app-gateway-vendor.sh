#!/usr/bin/env bash
# Copy bundled app-gateway vendor charts into installer dist (called from package.sh).
set -euo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST="${DIST_PATH:-${BASE_DIR}/../.dist}"
SRC="${BASE_DIR}/../framework/app-gateway/vendor-charts"
VENDOR_SCRIPT="${BASE_DIR}/../framework/app-gateway/build/bundle-vendor-charts.sh"

if [[ ! -d "${SRC}/envoy-gateway-helm" ]]; then
  echo "app-gateway vendor-charts missing; running bundle-vendor-charts.sh ..."
  bash "${VENDOR_SCRIPT}"
fi

DEST="${DIST}/wizard/config/app-gateway-vendor"
mkdir -p "${DEST}"
rm -rf "${DEST:?}"/*
cp -a "${SRC}/." "${DEST}/"
echo "packaged app-gateway-vendor -> ${DEST}"
