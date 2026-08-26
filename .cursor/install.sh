#!/usr/bin/env bash
#
# Cloud Agent bootstrap for the Olares monorepo.
#
# Installs the toolchains the repository needs (Go 1.25 for the Go modules,
# plus the CGO system libraries the olaresd daemon links against), warms the
# module/npm caches, and builds the two flagship Go applications so a fresh
# agent starts from a known-good state. Safe to run repeatedly.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# Go version required by cli/go.mod and daemon/go.mod (the `go` directive).
GO_VERSION="1.25.0"
ARCH="$(dpkg --print-architecture)"

echo "==> Ensuring Go ${GO_VERSION} is installed"
if ! /usr/local/go/bin/go version 2>/dev/null | grep -q "go${GO_VERSION} "; then
  curl -fsSL -o /tmp/go.tgz "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf /tmp/go.tgz
  rm -f /tmp/go.tgz
fi
# Make the toolchain available on the default PATH (/usr/local/bin is already there).
sudo ln -sf /usr/local/go/bin/go /usr/local/bin/go
sudo ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
go version

echo "==> Installing system build dependencies (CGO libs for olaresd)"
sudo apt-get update -qq
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq libudev-dev libpcap-dev

echo "==> Building olares-cli"
make -C cli build

echo "==> Building olaresd daemon"
( cd daemon && CGO_ENABLED=1 go build -o bin/olaresd ./cmd/terminusd/main.go )

echo "==> Installing frontend (apps) workspace dependencies"
if command -v npm >/dev/null 2>&1; then
  ( cd apps && npm install )
else
  echo "npm not found; skipping frontend dependency install"
fi

echo "==> Bootstrap complete"
