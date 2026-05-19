#!/usr/bin/env bash
# Optional linkerd-viz install (no bundled Prometheus). Prefer: olares-cli install-linkerd-viz
set -euo pipefail

PROMETHEUS_URL="${OLARES_LINKERD_PROMETHEUS_URL:-http://prometheus-k8s.kubesphere-monitoring-system.svc.cluster.local:9090}"
INSTALLER_DIR="${OLARES_INSTALLER_DIR:-}"

args=(install-linkerd-viz --prometheus-url "${PROMETHEUS_URL}")
if [[ -n "${INSTALLER_DIR}" ]]; then
  args+=(--installer-dir "${INSTALLER_DIR}")
fi

if command -v olares-cli >/dev/null 2>&1; then
  exec olares-cli "${args[@]}"
fi

echo "olares-cli not found; build from Olares/cli or set PATH" >&2
exit 1
