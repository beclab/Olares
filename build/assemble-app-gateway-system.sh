#!/usr/bin/env bash
#
# Assemble the app-gateway-system umbrella chart for the Olares install-wizard.
#
# Envoy Gateway is NOT vendored into this repository. This script pulls it from
# the official OCI registry at the version pinned in UPSTREAM.lock.yaml and lays
# it into the (gitignored) charts/ and crds/ sub-directories of the umbrella
# chart, exactly as app-service charts are assembled into the wizard bundle.
# No service mesh (Linkerd) is involved.

set -euo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${BASE_DIR}/.." && pwd)"

AGW_DIR="${REPO_ROOT}/framework/app-gateway"
UPSTREAM_LOCK="${AGW_DIR}/upstream-charts/UPSTREAM.lock.yaml"
SYSTEM_DIR="${AGW_DIR}/.olares/config/app-gateway-system"
VENDOR_VALS="${AGW_DIR}/.olares/config/app-gateway-vendor"
CHARTS_DIR="${SYSTEM_DIR}/charts"
CRDS_DIR="${SYSTEM_DIR}/crds"

OCI_REGISTRY="${APP_GATEWAY_EG_OCI:-oci://docker.io/envoyproxy}"

required_paths=(
  "${UPSTREAM_LOCK}"
  "${SYSTEM_DIR}/Chart.yaml"
  "${SYSTEM_DIR}/values.yaml"
  "${VENDOR_VALS}/envoy-gateway-crds-values.yaml"
)
for path in "${required_paths[@]}"; do
  if [ ! -e "${path}" ]; then
    echo "missing required path: ${path}" >&2
    exit 1
  fi
done

if ! command -v helm >/dev/null 2>&1; then
  echo "helm is required on PATH" >&2
  exit 1
fi

eg_version="$(awk -F'"' '/^envoy_gateway:/ {print $2; exit}' "${UPSTREAM_LOCK}")"
if [ -z "${eg_version}" ]; then
  echo "failed to parse envoy_gateway version from ${UPSTREAM_LOCK}" >&2
  exit 1
fi

# Guard: the umbrella dependency version must match the pinned lock version.
dependency_version() {
  local alias_name="$1"
  awk -v alias_name="${alias_name}" '
    $1 == "-" && $2 == "name:" {in_dep=1; alias=""; version=""; next}
    in_dep && $1 == "alias:" {alias=$2; next}
    in_dep && $1 == "version:" {
      version=$2
      if (alias == alias_name) { gsub(/"/, "", version); print version; exit }
      next
    }
  ' "${SYSTEM_DIR}/Chart.yaml"
}
dep_eg_version="$(dependency_version "envoy-gateway")"
if [ "${dep_eg_version}" != "${eg_version}" ]; then
  echo "app-gateway-system dependency envoy-gateway version mismatch: got ${dep_eg_version}, want ${eg_version}" >&2
  exit 1
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

echo "pulling Envoy Gateway ${eg_version} from ${OCI_REGISTRY} (control plane + CRDs) ..."
helm pull "${OCI_REGISTRY}/gateway-helm" --version "${eg_version}" --untar --untardir "${tmp_dir}/cp"
helm pull "${OCI_REGISTRY}/gateway-crds-helm" --version "${eg_version}" --untar --untardir "${tmp_dir}/crds"

rm -rf "${CHARTS_DIR}" "${CRDS_DIR}"
mkdir -p "${CHARTS_DIR}" "${CRDS_DIR}/envoy-gateway-crds"

# Control-plane subchart for the umbrella Helm release (installed with --skip-crds).
cp -a "${tmp_dir}/cp/gateway-helm" "${CHARTS_DIR}/envoy-gateway-helm"

# CRDs are applied separately via server-side apply (large CRD set > Helm 1MiB limit).
cp -a "${tmp_dir}/crds/gateway-crds-helm" "${CRDS_DIR}/envoy-gateway-crds/chart"
cp -f "${VENDOR_VALS}/envoy-gateway-crds-values.yaml" "${CRDS_DIR}/envoy-gateway-crds/values.yaml"

echo "assembled app-gateway-system with Envoy Gateway ${eg_version} (no service mesh)"
