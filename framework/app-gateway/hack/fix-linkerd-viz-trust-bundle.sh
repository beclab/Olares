#!/usr/bin/env bash
# Fix linkerd-viz Dashboard 500 after Linkerd control-plane reinstall / PKI trust-bundle change.
# Root cause: viz data-plane proxies keep old trust bundle → BadSignature to policy:8090 / dst:8086.
set -euo pipefail

LINKERD_NS="${LINKERD_NS:-linkerd}"
VIZ_NS="${VIZ_NS:-linkerd-viz}"
APP_GATEWAY_NS="${APP_GATEWAY_NS:-app-gateway}"
INSTALLER_DIR="${OLARES_INSTALLER_DIR:-}"
SKIP_MAINTAIN_PKI="${SKIP_MAINTAIN_PKI:-0}"
SKIP_APP_GATEWAY_RESTART="${SKIP_APP_GATEWAY_RESTART:-0}"

echo "==> Label namespaces for Olares mesh NP"
kubectl label ns "${LINKERD_NS}" bytetrade.io/ns-type=system --overwrite
kubectl label ns "${VIZ_NS}" bytetrade.io/ns-type=system --overwrite 2>/dev/null || true

if [[ "${SKIP_MAINTAIN_PKI}" != "1" && -n "${INSTALLER_DIR}" ]]; then
  if command -v olares-cli >/dev/null 2>&1; then
    echo "==> olares-cli maintain-linkerd-pki"
    olares-cli maintain-linkerd-pki --installer-dir "${INSTALLER_DIR}"
  else
    echo "WARN: olares-cli not in PATH; skip maintain-linkerd-pki"
  fi
fi

echo "==> Rollout restart linkerd-viz (load current trust bundle)"
kubectl -n "${VIZ_NS}" rollout restart deploy
kubectl -n "${VIZ_NS}" rollout status deploy/metrics-api --timeout=300s
kubectl -n "${VIZ_NS}" rollout status deploy/web --timeout=300s
kubectl -n "${VIZ_NS}" rollout status deploy/tap --timeout=300s
kubectl -n "${VIZ_NS}" rollout status deploy/tap-injector --timeout=300s

if [[ "${SKIP_APP_GATEWAY_RESTART}" != "1" ]]; then
  if kubectl get ns "${APP_GATEWAY_NS}" >/dev/null 2>&1; then
    echo "==> Rollout restart ${APP_GATEWAY_NS} deployments"
    kubectl -n "${APP_GATEWAY_NS}" rollout restart deploy 2>/dev/null || true
    kubectl -n "${APP_GATEWAY_NS}" rollout status deploy --timeout=300s 2>/dev/null || true
  fi
fi

echo "==> Verify"
if command -v linkerd >/dev/null 2>&1; then
  linkerd viz check || true
  linkerd viz stat deploy -n "${VIZ_NS}" --time-window 1m || true
fi
kubectl -n "${VIZ_NS}" logs deploy/metrics-api -c linkerd-proxy --tail=15 || true

echo "OK: fix-linkerd-viz-trust-bundle finished. Port-forward: kubectl -n ${VIZ_NS} port-forward svc/web 8084:8084"
