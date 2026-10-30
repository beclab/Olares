#!/bin/sh
# Local verification with sibling API/proxylistener checkouts until the API is
# published. No go.mod replacements or service/cluster mutations are made.
set -eu
repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
api_dir=${ENTRANCE_API_DIR:-"$repo_dir/../api"}
proxy_dir=${ENTRANCE_PROXY_DIR:-"$repo_dir/../proxylistener"}
api_dir=$(CDPATH= cd -- "$api_dir" && pwd)
proxy_dir=$(CDPATH= cd -- "$proxy_dir" && pwd)
output_dir=${ENTRANCE_BUILD_OUTPUT:-$(mktemp -d /tmp/entrance-ids-build.XXXXXX)}
mkdir -p "$output_dir"
output_dir=$(CDPATH= cd -- "$output_dir" && pwd)
work_dir=$(mktemp -d /tmp/entrance-ids-work.XXXXXX)
export GOCACHE=${GOCACHE:-"$work_dir/cache"}
# macOS does not provide C.UTF-8, which breaks legacy shell-output assertions.
export LC_ALL=C LANG=C
for kind in backend cli proxy; do
  work_file="$work_dir/$kind.work"
  printf 'go 1.25.7\n\nuse (\n' > "$work_file"
  printf ' "%s"\n' "$api_dir" >> "$work_file"
  case "$kind" in
    backend)
      for component in app-service bfl l4-bfl-proxy system-server; do
        printf ' "%s"\n' "$repo_dir/framework/$component" >> "$work_file"
      done ;;
    cli) printf ' "%s"\n' "$repo_dir/cli" >> "$work_file" ;;
    proxy) printf ' "%s"\n' "$proxy_dir" >> "$work_file" ;;
  esac
  printf ')\n' >> "$work_file"
done
build_component() (
  cd "$1"
  export GOWORK="$work_dir/$2.work"
  go build -o "$output_dir/$3" "$4"
)
if [ "${ENTRANCE_RUN_TESTS:-0}" = 1 ]; then
  (cd "$api_dir"; GOWORK=off go test ./api/app.bytetrade.io/v1alpha1 -run 'TestPersistentEntranceAllocation|TestSharedAndDevIDs|TestApplicationIDsAndRoundTrip|TestGlobalOrdinaryClaims'; GOWORK=off go test ./pkg/entranceids)
  (cd "$repo_dir/framework/app-service"; export GOWORK="$work_dir/backend.work"; go test ./pkg/appcfg ./pkg/gateway ./pkg/apiserver ./pkg/appinstaller; go test ./controllers -run 'TestAssignIDsUsesOriginalDomainUser|TestExistingMissingIDsAreNotReconciled')
  (cd "$repo_dir/framework/bfl"; GOWORK="$work_dir/backend.work" go test ./pkg/app_service/v1)
  (cd "$repo_dir/framework/l4-bfl-proxy"; GOWORK="$work_dir/backend.work" go test ./internal/translator ./internal/provider)
  (cd "$repo_dir/framework/system-server"; GOWORK="$work_dir/backend.work" go test ./pkg/providerregistry/v2alpha1)
  (cd "$repo_dir/cli"; GOWORK="$work_dir/cli.work" go test ./pkg/upgrade/entranceids ./pkg/upgrade)
  (cd "$proxy_dir"; GOWORK="$work_dir/proxy.work" go test ./...)
fi
build_component "$repo_dir/framework/app-service" backend app-service ./cmd/app-service
build_component "$repo_dir/framework/bfl" backend bfl ./cmd/apiserver
build_component "$repo_dir/framework/l4-bfl-proxy" backend l4-bfl-proxy ./cmd
build_component "$repo_dir/framework/system-server" backend system-server ./cmd/server
build_component "$repo_dir/cli" cli olares-cli ./cmd
build_component "$proxy_dir" proxy proxylistener ./cmd/proxylistener
printf 'Built binaries: %s\nTemporary workspaces: %s\n' "$output_dir" "$work_dir"
