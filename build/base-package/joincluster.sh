#!/usr/bin/env bash

# Bootstrap a worker node into an existing Olares cluster.
#
# This script only has one job: install the olares-cli matching the cluster's
# Olares version, then hand over to "olares-cli node join", which owns the
# actual flow (validating the master, preparing this machine and joining).
#
# It is normally not run by hand: "olares-cli node join-command" on the
# master prints a ready-to-paste invocation with every value already filled in.
#
# Environment:
#   VERSION                    Olares version to join, must match the master.
#   MASTER_AUTH_INFO           Master address and SSH credentials, as generated
#                              by "olares-cli node join-command".
#   OLARES_SYSTEM_CDN_SERVICE  CDN endpoint to download from; defaults to the
#                              global endpoint. Regions with their own endpoint
#                              must set it, and the generated command does.
#   MASTER_HOST, MASTER_SSH_USER, MASTER_SSH_PASSWORD,
#   MASTER_SSH_PRIVATE_KEY_PATH, MASTER_SSH_PORT
#                              Master connection, for use without a payload.
#                              They also override individual payload fields.

set -o pipefail
set -e

DEFAULT_CDN_SERVICE="https://cdn.olares.com"

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

fail() {
    echo "error: $*" >&2
    exit 1
}

# read_tty prompts on the controlling terminal rather than stdin, because this
# script is normally piped into bash and stdin therefore carries the script.
# Without a usable terminal (CI, "ssh -T", cloud-init) prompting can never
# succeed, so report what to set instead of failing on a redirection error.
read_tty() {
    local label="$1" variable="$2" no_tty_hint="$3" value
    if ! { : >/dev/tty; } 2>/dev/null; then
        fail "$no_tty_hint"
    fi
    printf "%s" "$label" >/dev/tty
    IFS= read -r value </dev/tty || fail "$no_tty_hint"
    printf -v "$variable" "%s" "$value"
}

run_root() {
    if [[ "$(id -u)" == "0" ]]; then
        "$@"
        return
    fi
    command_exists sudo || fail "sudo is required to install and run olares-cli"
    sudo -E "$@"
}

# detect_repo_path picks the download path for this hardware. It is normally
# replaced at build time; the /etc/machine.info fallback exists so a script
# fetched straight from the repository still finds the Olares One artifacts.
detect_repo_path() {
    if [[ -n "$REPO_PATH" ]]; then
        return
    fi
    REPO_PATH="#__REPO_PATH__"
    if [[ -z "$REPO_PATH" || "${REPO_PATH:3}" == "REPO_PATH__" ]]; then
        local machine_info=""
        if [[ -r /etc/machine.info ]]; then
            machine_info="$(tr -d '[:space:]' </etc/machine.info)"
        fi
        shopt -s nocasematch
        if [[ "$machine_info" == "olaresone" ]]; then
            REPO_PATH="/olares-one/"
        else
            REPO_PATH="/"
        fi
        shopt -u nocasematch
    fi
    export REPO_PATH
}

resolve_version() {
    if [[ -z "$VERSION" ]]; then
        VERSION="#__VERSION__"
    fi
    if [[ -z "$VERSION" || "${VERSION:3}" == "VERSION__" ]]; then
        VERSION=""
        while [[ -z "$VERSION" ]]; do
            read_tty "Olares version to join (for example 1.12.7): " VERSION \
                "the Olares version is unknown; set VERSION to the version running on the master (VERSION=1.12.7), or use the command printed by 'olares-cli node join-command' on the master"
        done
    fi
    if [[ ! "$VERSION" =~ ^[A-Za-z0-9._-]+$ ]]; then
        fail "invalid Olares version '$VERSION'"
    fi
    export VERSION
}

# detect_arch sets ARCH. It assigns a global rather than echoing, so that a
# failure here exits the script instead of only the command substitution.
detect_arch() {
    local os_arch
    [[ "$(uname -s)" == "Linux" ]] || fail "only Linux machines can join an Olares cluster"
    os_arch="$(uname -m)"
    case "$os_arch" in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) fail "unsupported architecture '$os_arch'" ;;
    esac
}

# expected_vendor mirrors the vendor an olares-cli built for this download path
# reports, so a CLI left over from a different build is never reused.
expected_vendor() {
    if [[ "$(basename "$REPO_PATH")" == "olares-one" ]]; then
        echo "OlaresOne"
    else
        echo "main"
    fi
}

cli_is_current() {
    local cli="$1"
    command_exists "$cli" || return 1
    [[ "$("$cli" -v 2>/dev/null | awk 'NR==1{print $3}')" == "$VERSION" ]] || return 1
    [[ "$("$cli" --vendor 2>/dev/null)" == "$(expected_vendor)" ]] || return 1
}

install_cli() {
    local target="$1" cdn_url="$2"
    local release_suffix="" cli_file cli_url

    if [[ -n "$RELEASE_ID" ]]; then
        release_suffix=".$RELEASE_ID"
    fi

    cli_file="olares-cli-v${VERSION}_linux_${ARCH}${release_suffix}.tar.gz"
    cli_url="${cdn_url}${REPO_PATH}${cli_file}"

    WORK_DIR="$(mktemp -d)"
    trap 'rm -rf "$WORK_DIR"' EXIT

    echo "Downloading olares-cli ${VERSION} from ${cli_url} ..."
    curl --fail --show-error --location --retry 3 -o "$WORK_DIR/$cli_file" "$cli_url" ||
        fail "failed to download olares-cli from $cli_url"
    gzip -t "$WORK_DIR/$cli_file" 2>/dev/null || fail "the downloaded olares-cli archive is invalid"
    tar -zxf "$WORK_DIR/$cli_file" -C "$WORK_DIR" olares-cli || fail "failed to unpack olares-cli"
    run_root install -m 0755 "$WORK_DIR/olares-cli" "$target"
}

main() {
    local cdn_url target base_dir

    detect_arch
    command_exists curl || fail "curl is required to download olares-cli"
    command_exists tar || fail "tar is required to unpack olares-cli"
    command_exists gzip || fail "gzip is required to verify olares-cli"

    detect_repo_path
    resolve_version

    cdn_url="${OLARES_SYSTEM_CDN_SERVICE:-$DEFAULT_CDN_SERVICE}"
    cdn_url="${cdn_url%/}"
    export OLARES_SYSTEM_CDN_SERVICE="$cdn_url"

    if [[ -z "$RELEASE_ID" ]]; then
        RELEASE_ID="#__RELEASE_ID__"
        if [[ "${RELEASE_ID:3}" == "RELEASE_ID__" ]]; then
            RELEASE_ID=""
        fi
    fi
    export RELEASE_ID

    target="/usr/local/bin/olares-cli"
    if cli_is_current "$target"; then
        echo "olares-cli ${VERSION} is already installed, skipping the download"
    else
        install_cli "$target" "$cdn_url"
    fi

    # Export the base directory rather than passing --base-dir, so that it stays
    # a default rather than an override.
    #
    # The CLI would otherwise fall back to the passwd home of the effective user,
    # which is root's under sudo. But it must not outrank OLARES_BASE_DIR from
    # /etc/olares/release either: that file records where this machine's Olares
    # actually lives, and it has to win. Otherwise a node prepared or installed
    # by root, joined by a regular user, would be inspected at the wrong path and
    # look untouched -- and the refusal to join a node that already runs Olares
    # would silently never trigger. An operator who really does need another
    # location can still pass --base-dir to `olares-cli node join`.
    export BASE_DIR="$HOME/.olares"
    mkdir -p "$BASE_DIR"

    run_root "$target" node join
}

main "$@"
