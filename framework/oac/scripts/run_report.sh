#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/.."
go test -run TestAppsTestdataReport -v -timeout 600s .
