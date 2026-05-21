#!/usr/bin/env bash
# test-files-cmd.sh — End-to-end smoke tests for `olares-cli files <verb>`.
#
# What this script does
# ---------------------
# Drives EVERY verb under `olares-cli files` against a live Olares
# instance (drive/Home/, drive/Data/, optionally a sync repo / cloud
# drive / SMB share). Top-level verbs covered:
#
#   ls       upload   download   cat        mkdir (md)
#   rm       cp       mv         rename (rn / "delete" / "remove")
#   chown    share    smb        repos
#
# Per-verb subcommand coverage:
#
#   share   list, get, public (create/get/rm), set-password,
#           smb-users list. (Internal / SMB share creation and
#           set-members / set-smb need a real recipient roster, so
#           they are exercised only at the parser-rejection layer.)
#   smb     history list / add / rm round-trip; mount / unmount
#           gated on --smb-url (a real SMB target). The history
#           subverbs are safe to run without any external SMB
#           server — the wire endpoint just stores favorites.
#   repos   list, get, create, rename, rm (full lifecycle on a
#           throwaway repo) plus the existing per-suite Sync repo
#           that file-verbs parent under.
#
# Each test is a single function with three parts: a one-line `step`
# banner, the actual command(s), and a `expect_*` assertion that
# upgrades a silent failure ("server returned 500 but my script kept
# going") into a loud one. Failures DO NOT abort the script — every
# remaining test still runs, and the per-test status is summarized at
# the end with a non-zero exit if anything failed. That's the right
# trade-off for a smoke harness: a single broken verb shouldn't hide
# the rest of the surface from CI.
#
# Why a regression test for upload-to-root
# ----------------------------------------
# Pre-fix, `olares-cli files upload ./mydir/ drive/Home/` failed with
#
#   Error: local "./mydir/" is a directory; remote "" must end with '/'
#
# because the SubPath="/" → remoteSub="" conversion stripped the only
# trailing slash before BuildPlan got a chance to read it. The
# `test_upload_dir_to_root` step pins that path so a future
# regression in subPathForBuildPlan (cli/cmd/ctl/files/upload.go) is
# caught end-to-end.
#
# Prerequisites
# -------------
#   1. `olares-cli` is on PATH and has been logged into a profile
#      (`olares-cli profile login --olares-id <id>`). The script does
#      NOT touch profiles — it talks to whichever profile is current.
#   2. The user has at least drive/Home write permission. Cloud-drive
#      and sync tests are gated on env vars (see USAGE below) so the
#      core run works on a fresh Olares without any extra setup.
#
# Usage
# -----
# Both CLI flags and OLARES_TEST_* env vars are accepted; flags win
# on conflict. Run `./test-files-cmd.sh --help` for the full flag
# table. A few common shapes:
#
#   # Minimal: drive/Home + drive/Data only.
#   ./test-files-cmd.sh
#
#   # Full namespace coverage (sync + cache + external).
#   ./test-files-cmd.sh \
#       --sync-repo-name cli-test-repo \
#       --cache    cache/<node> \
#       --external external/<node>/<volume>
#
#   # Cloud drives.
#   ./test-files-cmd.sh \
#       --awss3   awss3/<account>/<bucket> \
#       --google  google/<account> \
#       --dropbox dropbox/<account>
#
#   # SMB mount round-trip against a real Samba target.
#   ./test-files-cmd.sh \
#       --smb-url      //host.local/Public \
#       --smb-user     alice \
#       --smb-password 's3cret'
#
#   # Equivalent via env vars (legacy interface).
#   OLARES_TEST_SYNC_REPO_NAME="cli-test-repo" \
#   OLARES_TEST_CACHE_PATH="cache/<node>" \
#   OLARES_TEST_EXTERNAL_PATH="external/<node>/<volume>" \
#   OLARES_TEST_SMB_URL="//host.local/Public" \
#   OLARES_TEST_SMB_USER="alice" \
#   OLARES_TEST_SMB_PASSWORD='s3cret' \
#   ./test-files-cmd.sh
#
#   # Keep the remote test directories after the run (for debugging).
#   ./test-files-cmd.sh --keep
#
# Exit code: 0 if every test passed, 1 otherwise. The summary block
# at the end lists exactly which tests failed.
#
# Implementation notes
# --------------------
# - `set -u` is on (catch typos in env-var names early); `set -e` is
#   off (we explicitly check exit codes per step so one failure
#   doesn't truncate the run).
# - Every remote path lives under a per-run directory
#   (`/cli-test-<timestamp>-<pid>/`) so concurrent runs don't collide
#   and the cleanup is a single recursive `files rm -rf`.
# - `expect_*` helpers wrap the assertion verbs with a pass/fail
#   counter and a failure log so the summary is a one-glance read of
#   what's broken.

set -u

# --------------------------------------------------------------------
# Section 1: configuration + shared helpers
# --------------------------------------------------------------------

# Allow overriding the binary so a local build can be tested without
# clobbering the system install (e.g. OLARES_CLI=./cli/olares-cli).
OLARES_CLI="${OLARES_CLI:-olares-cli}"

# Per-run remote test directory — dropped under drive/Home so the
# cleanup is a single rm -rf of one path. Suffix carries timestamp +
# pid so concurrent invocations don't fight over the same name.
TEST_RUN_ID="cli-test-$(date +%Y%m%d-%H%M%S)-$$"
DRIVE_HOME_TEST_DIR="drive/Home/${TEST_RUN_ID}"
DRIVE_DATA_TEST_DIR="drive/Data/${TEST_RUN_ID}"

# Local scratch dir — `mktemp -d` keeps it out of the way and
# survivable to a Ctrl-C without polluting the user's $HOME.
LOCAL_TMP="$(mktemp -d -t olares-cli-test-XXXXXX)"

# Optional surfaces — empty string means "skip those tests". Cloud /
# sync / cache / external / SMB all require external setup the smoke
# harness can't bootstrap on its own (a cloud account, a seafile
# library, a node-attached cache / external volume, an SMB roster),
# so we gate them behind env vars / CLI flags rather than failing
# loudly when they're absent.
#
# Two interfaces drive these values:
#
#   1. environment variables — original interface, useful for CI
#      pipelines that inject everything via `env`.
#   2. command-line flags    — preferred for interactive runs;
#      mapped onto the same vars in parse_args (CLI wins on
#      conflict so a one-off override doesn't require unsetting
#      a shell-wide export first).
#
# Two parameters carry a strict shape:
#
#   OLARES_TEST_CACHE_PATH      "cache/<node>"
#                               <node> is whichever node hosts the
#                               cache volume (typically the master).
#                               The script appends the per-run-id
#                               subdir on top so a single base path
#                               drives every cache test.
#   OLARES_TEST_EXTERNAL_PATH   "external/<node>/<volume>"
#                               external attached storage requires
#                               BOTH the node and the volume name
#                               (the volume is the user-visible
#                               drive-letter equivalent in the
#                               LarePass UI).
OLARES_TEST_SYNC_REPO_NAME="${OLARES_TEST_SYNC_REPO_NAME:-}"
OLARES_TEST_CACHE_PATH="${OLARES_TEST_CACHE_PATH:-}"
OLARES_TEST_EXTERNAL_PATH="${OLARES_TEST_EXTERNAL_PATH:-}"
OLARES_TEST_AWSS3_PATH="${OLARES_TEST_AWSS3_PATH:-}"
OLARES_TEST_GOOGLE_PATH="${OLARES_TEST_GOOGLE_PATH:-}"
OLARES_TEST_DROPBOX_PATH="${OLARES_TEST_DROPBOX_PATH:-}"

# SMB mount/unmount gating. The `smb history` tests run
# unconditionally (the favorites endpoint is just metadata storage —
# no live SMB server required), but the actual `smb mount` test
# needs a reachable SMB share with valid credentials, so we gate it
# behind these three knobs. Each is empty unless the user opts in
# via the matching --smb-* CLI flag or OLARES_TEST_SMB_* env var.
#
#   OLARES_TEST_SMB_URL       e.g. //host.local/Public
#   OLARES_TEST_SMB_USER      SMB username (empty allowed for guest)
#   OLARES_TEST_SMB_PASSWORD  SMB password (empty allowed for guest)
#
# OLARES_TEST_SMB (the legacy no-suffix var) is kept around as a
# back-compat alias for OLARES_TEST_SMB_URL — if the user set the
# old name and not the new one, treat them as equivalent. We do the
# fold AFTER all three vars are read so the longhand always wins on
# conflict (same precedence rule we use elsewhere for CLI > env).
OLARES_TEST_SMB="${OLARES_TEST_SMB:-}"
OLARES_TEST_SMB_URL="${OLARES_TEST_SMB_URL:-}"
OLARES_TEST_SMB_USER="${OLARES_TEST_SMB_USER:-}"
OLARES_TEST_SMB_PASSWORD="${OLARES_TEST_SMB_PASSWORD:-}"
if [[ -z "${OLARES_TEST_SMB_URL}" && -n "${OLARES_TEST_SMB}" ]]; then
    OLARES_TEST_SMB_URL="${OLARES_TEST_SMB}"
fi

OLARES_TEST_KEEP="${OLARES_TEST_KEEP:-}"

PASSED=0
FAILED=0
FAILED_TESTS=()
CURRENT_STEP=""

# Sync-repo state. setup_sync_repo populates these when
# OLARES_TEST_SYNC_REPO_NAME is set + the create call succeeds; every
# downstream sync test (and the cleanup trap) reads these globals to
# know whether the sync surface is in play and which repo to target.
# Kept as plain strings (not arrays) so a single empty-string check
# is enough to gate the sync-only tests.
#
#   SYNC_REPO_ID         the server-assigned UUID; used as <extend>
#   SYNC_REPO_NAME_USED  display name including the run-id suffix —
#                        printed in the post-run summary so a
#                        --keep'd repo is identifiable in the
#                        LarePass UI without having to grep for the
#                        UUID
#   SYNC_REPO_PATH       canonical 2-segment frontend path
#                        ("sync/<repo_id>") — repo-root operations
#                        target this directly so the trailing-slash
#                        regression (subPathForBuildPlan) is also
#                        exercised on the sync namespace
#   SYNC_REPO_TEST_DIR   per-run subdir inside the repo
#                        ("sync/<repo_id>/<run-id>") — every sync
#                        file test parents under this so we can
#                        also exercise non-root paths in the same
#                        repo without polluting its top level
SYNC_REPO_ID=""
SYNC_REPO_NAME_USED=""
SYNC_REPO_PATH=""
SYNC_REPO_TEST_DIR=""

# Cache / external state. Unlike sync, we don't provision the
# underlying volume — the user has to point us at one that already
# exists on a node (cache is node-local scratch space, external is a
# user-attached USB / network drive). setup_cache_paths /
# setup_external_paths just normalize the user-supplied base path
# and append the per-run-id subdirectory so every test under these
# namespaces parents under one cleanable directory.
#
# When the corresponding env var / CLI flag is unset, *_BASE_PATH
# stays empty and *_TEST_DIR is never populated; downstream tests
# gate on *_TEST_DIR and SKIP cleanly.
#
#   *_BASE_PATH   user-supplied root, e.g. "cache/node-1" or
#                 "external/node-1/hdd1". Trailing '/' stripped by
#                 the parser so the per-run-id concat is unambiguous.
#   *_TEST_DIR    full per-run path including <run-id>, e.g.
#                 "cache/node-1/cli-test-...". This is what tests
#                 actually mkdir/upload/rm against.
CACHE_BASE_PATH=""
CACHE_TEST_DIR=""
EXTERNAL_BASE_PATH=""
EXTERNAL_TEST_DIR=""

# SMB-history bookkeeping. Every URL the suite adds via
# `smb history add` is appended here; the cleanup trap loops over
# the array and runs `smb history rm` per entry so a half-failed
# test (e.g. add succeeded but list assertion failed) doesn't leave
# orphan favorites on the user's profile. Plain string array — bash
# array idioms keep the loop trivial and the empty-array short-
# circuit a one-line `(( ${#...[@]} > 0 ))` guard.
SMB_HISTORY_ADDED_URLS=()

# SMB-mount bookkeeping. When test_smb_mount actually mounts a
# share, it records the resulting external/<node>/<entry>/ path so
# the cleanup trap can unmount it on the way out (vs. leaving the
# mount lying around if the script is interrupted mid-suite).
SMB_MOUNTED_ENTRY=""
SMB_MOUNT_NODE=""

# TESTS_STARTED gates the remote cleanup on whether we got far
# enough that there's actually something to clean up. Without
# this guard, ANY exit fires the cleanup trap — including
# `--help` (exit 0), parse_args' "unknown flag" (exit 2), and
# preflight's "olares-cli not logged in" (exit 1). Each of those
# would otherwise call `olares-cli files rm -rf <run-dir>/`,
# which on a not-logged-in profile hangs on the HTTP timeout for
# a couple of minutes per call. Setting this flag at the very
# end of preflight() ensures the cleanup only attempts remote
# operations when there's a working profile + a successful
# probe, which is exactly when remote artifacts could exist.
TESTS_STARTED=0

# usage prints the help banner and exits. Kept verbose on purpose:
# once you forget which flag corresponds to which env var, you'd
# rather grep --help than read the script source.
usage() {
    cat <<'EOF'
Usage: test-files-cmd.sh [flags]

End-to-end smoke tests for `olares-cli files <verb>`. Drives every
verb against a live Olares instance and reports per-test pass/fail.

Flags / environment variables (CLI flag wins on conflict):

  -b, --cli <path>             Use this olares-cli binary
                               (env: OLARES_CLI)

      --sync-repo-name <name>  Base name for the temp Sync repo;
                               full set of sync tests + repos
                               lifecycle runs when set
                               (env: OLARES_TEST_SYNC_REPO_NAME)

      --cache <cache/<node>>   Cache base path; mkdir tests run
                               under <cache/<node>>/<run-id>/
                               (env: OLARES_TEST_CACHE_PATH)

      --external <external/<node>/<volume>>
                               External base path; mkdir tests run
                               under <external/...>/<run-id>/
                               (env: OLARES_TEST_EXTERNAL_PATH)

      --awss3   <awss3/<account>/<bucket>>
      --google  <google/<account>>
      --dropbox <dropbox/<account>>
                               Cloud-drive paths; upload/ls/cat/rm
                               round-trip per drive
                               (env: OLARES_TEST_AWSS3_PATH /
                                     OLARES_TEST_GOOGLE_PATH /
                                     OLARES_TEST_DROPBOX_PATH)

      --smb-url      <//host/share>
      --smb-user     <user>
      --smb-password <password>
                               SMB mount target + credentials. Enables
                               the live mount/unmount round-trip in
                               test_smb_mount. --smb-user and
                               --smb-password are optional for guest
                               shares. The smb history sub-tests run
                               UNCONDITIONALLY (they just exercise
                               the favorites endpoint, no live SMB
                               server needed).
                               (env: OLARES_TEST_SMB_URL /
                                     OLARES_TEST_SMB_USER /
                                     OLARES_TEST_SMB_PASSWORD)

      --keep                   Skip cleanup; preserve test dirs
                               and the temp Sync repo for
                               post-mortem inspection
                               (env: OLARES_TEST_KEEP=1)

  -h, --help                   This message.

Examples:

  # Drive-only smoke run (default).
  ./test-files-cmd.sh

  # Add the full sync + cache + external surface.
  ./test-files-cmd.sh \
    --sync-repo-name cli-test-repo \
    --cache    cache/node-1 \
    --external external/node-1/hdd1

  # Mixed env / flag — env supplies the binary, flag adds cache.
  OLARES_CLI=./cli/olares-cli ./test-files-cmd.sh --cache cache/master

  # Keep everything for debugging.
  ./test-files-cmd.sh --keep --sync-repo-name cli-test-repo
EOF
}

# require_value is a tiny helper for `--flag value` style options:
# it errors out cleanly when the user omits the value (otherwise
# bash quietly takes the next flag as the value, which produces
# very confusing downstream errors like
# "--external must be cache/<node>").
#
# Implementation note: returns the validated value via the global
# REQUIRE_VALUE_OUT rather than stdout. The obvious "use $(...)"
# pattern is broken here because `exit 2` inside a command
# substitution only kills the subshell — the parent script
# happily continues with an empty value and silently runs the
# wrong test set. Using a global side-steps the subshell entirely
# so a missing value aborts the script as intended.
REQUIRE_VALUE_OUT=""
require_value() {
    local flag="$1"
    local value="${2-}"
    if [[ -z "${value}" || "${value}" == --* || "${value}" == -* ]]; then
        echo "ERROR: ${flag} requires a value" >&2
        exit 2
    fi
    REQUIRE_VALUE_OUT="${value}"
}

# parse_args walks argv and folds CLI flags onto the OLARES_TEST_*
# env vars. We do this BEFORE cleanup setup (so --help can exit
# without firing the EXIT trap) and BEFORE the configuration prints
# in main() (so the banner reflects the final values).
#
# Both `--flag value` and `--flag=value` styles are accepted —
# coreutils-style splitting on '=' so users with strong opinions
# about either style get what they expect. Unknown flags fail
# fast with the help banner so a typo never silently runs the
# wrong test set.
parse_args() {
    while (( $# > 0 )); do
        local arg="$1"
        # Normalize --flag=value into --flag + value so the
        # subsequent shift logic doesn't have to special-case the
        # = form.
        if [[ "${arg}" == --*=* ]]; then
            local key="${arg%%=*}"
            local val="${arg#*=}"
            set -- "${key}" "${val}" "${@:2}"
            arg="$1"
        fi
        case "${arg}" in
            -b|--cli)
                require_value "${arg}" "${2-}"
                OLARES_CLI="${REQUIRE_VALUE_OUT}"
                shift 2
                ;;
            --sync-repo-name)
                require_value "${arg}" "${2-}"
                OLARES_TEST_SYNC_REPO_NAME="${REQUIRE_VALUE_OUT}"
                shift 2
                ;;
            --cache)
                require_value "${arg}" "${2-}"
                OLARES_TEST_CACHE_PATH="${REQUIRE_VALUE_OUT}"
                shift 2
                ;;
            --external)
                require_value "${arg}" "${2-}"
                OLARES_TEST_EXTERNAL_PATH="${REQUIRE_VALUE_OUT}"
                shift 2
                ;;
            --awss3)
                require_value "${arg}" "${2-}"
                OLARES_TEST_AWSS3_PATH="${REQUIRE_VALUE_OUT}"
                shift 2
                ;;
            --google)
                require_value "${arg}" "${2-}"
                OLARES_TEST_GOOGLE_PATH="${REQUIRE_VALUE_OUT}"
                shift 2
                ;;
            --dropbox)
                require_value "${arg}" "${2-}"
                OLARES_TEST_DROPBOX_PATH="${REQUIRE_VALUE_OUT}"
                shift 2
                ;;
            --smb-url)
                require_value "${arg}" "${2-}"
                OLARES_TEST_SMB_URL="${REQUIRE_VALUE_OUT}"
                shift 2
                ;;
            --smb-user)
                # Empty username is meaningful for SMB (guest /
                # anonymous shares), so we DON'T route this through
                # require_value — that helper rejects empty values.
                # `shift 2` still works because the caller is
                # supplying `--smb-user ""` (two argv slots) when
                # they want an empty user explicitly.
                OLARES_TEST_SMB_USER="${2-}"
                shift 2
                ;;
            --smb-password)
                # Same reasoning as --smb-user: empty password is
                # the wire shape for guest shares; let the user
                # opt into it explicitly.
                OLARES_TEST_SMB_PASSWORD="${2-}"
                shift 2
                ;;
            --keep)
                OLARES_TEST_KEEP=1
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            --)
                shift
                break
                ;;
            -*)
                echo "ERROR: unknown flag: ${arg}" >&2
                echo >&2
                usage >&2
                exit 2
                ;;
            *)
                echo "ERROR: unexpected positional argument: ${arg}" >&2
                echo "       test-files-cmd.sh takes no positional args; see --help." >&2
                exit 2
                ;;
        esac
    done
}

# normalize_paths validates + canonicalizes the cache / external
# base paths users supply and computes the per-run *_TEST_DIR
# values. Validation is shape-only (must start with the right
# fileType prefix and have the right number of segments) — we
# don't roundtrip-check against the server here because the user
# might be running this against a node that's not yet reachable,
# and the underlying ls/mkdir calls will surface a more useful
# server-side error than a generic "path doesn't exist" we'd
# fabricate ourselves.
normalize_paths() {
    if [[ -n "${OLARES_TEST_CACHE_PATH}" ]]; then
        # Strip any trailing '/' for clean concat with run-id.
        local cache_path="${OLARES_TEST_CACHE_PATH%/}"
        # Shape: cache/<node> (exactly two segments). Allowing more
        # would force the user to remember which level the run-id
        # gets appended at — strict-2 keeps the contract obvious.
        if [[ ! "${cache_path}" =~ ^cache/[^/]+$ ]]; then
            echo "ERROR: --cache must match cache/<node> (got: ${OLARES_TEST_CACHE_PATH})" >&2
            exit 2
        fi
        CACHE_BASE_PATH="${cache_path}"
        CACHE_TEST_DIR="${cache_path}/${TEST_RUN_ID}"
    fi
    if [[ -n "${OLARES_TEST_EXTERNAL_PATH}" ]]; then
        local ext_path="${OLARES_TEST_EXTERNAL_PATH%/}"
        # Shape: external/<node>/<volume> (exactly three segments).
        # external requires both <node> and <volume> per the path
        # parser (cli/cmd/ctl/files/path.go); we validate up-front
        # so a typo fails fast with a useful message.
        if [[ ! "${ext_path}" =~ ^external/[^/]+/[^/]+$ ]]; then
            echo "ERROR: --external must match external/<node>/<volume> (got: ${OLARES_TEST_EXTERNAL_PATH})" >&2
            exit 2
        fi
        EXTERNAL_BASE_PATH="${ext_path}"
        EXTERNAL_TEST_DIR="${ext_path}/${TEST_RUN_ID}"
    fi
}

# Trap Ctrl-C / unexpected exits to still run cleanup so a partial
# run doesn't leave the test directory on the server.
cleanup() {
    local rc=$?
    # Skip the cleanup banner + remote rm calls entirely when we
    # never made it past preflight. This covers the --help / bad
    # flag / not-logged-in exit paths, which would otherwise hang
    # on per-rm HTTP timeouts AND print a misleading "Cleanup"
    # banner that suggests the user's data is in a half-state.
    if (( TESTS_STARTED == 0 )); then
        if [[ -d "${LOCAL_TMP}" ]]; then
            rm -rf "${LOCAL_TMP}"
        fi
        exit $rc
    fi
    echo
    echo "============================================"
    echo "Cleanup"
    echo "============================================"
    if [[ -z "${OLARES_TEST_KEEP}" ]]; then
        # `|| true` — server may have already nuked the dir as part
        # of the test, or the dir may not exist if we failed before
        # creating it. We don't want cleanup to mask the test exit
        # code.
        echo "removing remote test directories"
        "${OLARES_CLI}" files rm -rf "${DRIVE_HOME_TEST_DIR}/" 2>/dev/null || true
        "${OLARES_CLI}" files rm -rf "${DRIVE_DATA_TEST_DIR}/" 2>/dev/null || true
        # cache / external have their own per-run subdirs hanging
        # off a user-supplied base path. We only remove the per-run
        # subdir, NOT the base — the base might be node-local
        # scratch shared with other workloads, and nuking it would
        # be way too eager.
        if [[ -n "${CACHE_TEST_DIR}" ]]; then
            "${OLARES_CLI}" files rm -rf "${CACHE_TEST_DIR}/" 2>/dev/null || true
        fi
        if [[ -n "${EXTERNAL_TEST_DIR}" ]]; then
            "${OLARES_CLI}" files rm -rf "${EXTERNAL_TEST_DIR}/" 2>/dev/null || true
        fi
        # SMB cleanup runs BEFORE the sync repo gets torn down — the
        # SMB tests don't depend on the sync repo, but ordering them
        # here keeps the heavier sync-repo operation as the very
        # last remote write so a hung HTTP call there doesn't
        # strand SMB favorites the user might want to inspect.
        #
        # Two phases:
        #   1. If the mount round-trip succeeded but the unmount
        #      assertion didn't run / didn't fire (e.g. script Ctrl-
        #      C'd between mount and unmount), nuke the live mount
        #      first so the external/<node>/<entry>/ namespace
        #      doesn't keep the share visible after the run.
        #   2. Drop every history entry we added. Per-URL rm rather
        #      than a single batched call so a missing entry (e.g.
        #      already removed by the test itself) doesn't bring
        #      the whole rm down.
        if [[ -n "${SMB_MOUNTED_ENTRY}" && -n "${SMB_MOUNT_NODE}" ]]; then
            echo "unmounting test SMB share: external/${SMB_MOUNT_NODE}/${SMB_MOUNTED_ENTRY}/"
            "${OLARES_CLI}" files smb unmount --node "${SMB_MOUNT_NODE}" "${SMB_MOUNTED_ENTRY}" 2>/dev/null || true
        fi
        if (( ${#SMB_HISTORY_ADDED_URLS[@]} > 0 )); then
            echo "removing ${#SMB_HISTORY_ADDED_URLS[@]} SMB history entr(y/ies) added by the suite"
            for url in "${SMB_HISTORY_ADDED_URLS[@]}"; do
                "${OLARES_CLI}" files smb history rm "${url}" 2>/dev/null || true
            done
        fi
        # The sync repo is the heaviest piece of state we create
        # (every file, every history entry, lives inside it), so
        # nuking it last gives the file-rm calls above their best
        # shot at cleaning up cross-namespace artifacts before the
        # repo itself goes away.
        if [[ -n "${SYNC_REPO_ID}" ]]; then
            echo "removing temporary Sync repo: ${SYNC_REPO_NAME_USED} (id=${SYNC_REPO_ID})"
            "${OLARES_CLI}" files repos rm -y "${SYNC_REPO_ID}" 2>/dev/null || true
        fi
    else
        echo "OLARES_TEST_KEEP set; remote test directories preserved:"
        echo "  ${DRIVE_HOME_TEST_DIR}/"
        echo "  ${DRIVE_DATA_TEST_DIR}/"
        if [[ -n "${CACHE_TEST_DIR}" ]]; then
            echo "  ${CACHE_TEST_DIR}/"
        fi
        if [[ -n "${EXTERNAL_TEST_DIR}" ]]; then
            echo "  ${EXTERNAL_TEST_DIR}/"
        fi
        if [[ -n "${SYNC_REPO_ID}" ]]; then
            echo "  sync/${SYNC_REPO_ID}/  (temp repo, name=${SYNC_REPO_NAME_USED})"
            echo "  -> manual cleanup: ${OLARES_CLI} files repos rm -y ${SYNC_REPO_ID}"
        fi
    fi
    if [[ -d "${LOCAL_TMP}" ]]; then
        rm -rf "${LOCAL_TMP}"
    fi
    summarize
    exit $rc
}
trap cleanup EXIT INT TERM

step() {
    CURRENT_STEP="$*"
    echo
    echo "============================================"
    echo "STEP: $*"
    echo "============================================"
}

# pass / fail bookkeeping. Kept tiny on purpose — every assertion
# helper funnels through these two so the summary numbers always
# match what was actually printed.
pass() {
    PASSED=$((PASSED + 1))
    echo "  PASS: $*"
}
fail() {
    FAILED=$((FAILED + 1))
    FAILED_TESTS+=("${CURRENT_STEP} :: $*")
    echo "  FAIL: $*" >&2
}

# expect_success: run the command, expect exit code 0. Captures
# stdout+stderr so we can echo it on failure (silent failures are
# the worst kind in a smoke harness — every command's output is
# preserved verbatim when the assertion blows up).
expect_success() {
    local desc="$1"; shift
    local out
    if out=$("$@" 2>&1); then
        pass "$desc"
        return 0
    else
        local rc=$?
        fail "$desc (exit $rc)"
        echo "    cmd: $*" >&2
        echo "    output:" >&2
        echo "$out" | sed 's/^/      /' >&2
        return $rc
    fi
}

# expect_failure: same shape as expect_success but inverts the
# success/fail axis. Used for the negative tests (e.g. directory
# upload without --recursive, root rm).
expect_failure() {
    local desc="$1"; shift
    local out
    if out=$("$@" 2>&1); then
        fail "$desc (expected non-zero exit, got 0)"
        echo "    cmd: $*" >&2
        echo "    output:" >&2
        echo "$out" | sed 's/^/      /' >&2
        return 1
    else
        pass "$desc"
        return 0
    fi
}

# expect_output_contains: run the command, expect 0 + a substring in
# stdout. Useful for `ls` / `cat` / `share get` / etc. where the
# value of the output is the whole point.
expect_output_contains() {
    local desc="$1"; shift
    local needle="$1"; shift
    local out
    if ! out=$("$@" 2>&1); then
        local rc=$?
        fail "$desc (exit $rc, expected to find: $needle)"
        echo "    cmd: $*" >&2
        echo "    output:" >&2
        echo "$out" | sed 's/^/      /' >&2
        return $rc
    fi
    if grep -qF -- "$needle" <<< "$out"; then
        pass "$desc"
        return 0
    fi
    fail "$desc (output did not contain: $needle)"
    echo "    cmd: $*" >&2
    echo "    output:" >&2
    echo "$out" | sed 's/^/      /' >&2
    return 1
}

# wait_for_ls polls `files ls <path>` until its output contains
# <needle> or the timeout expires. Used after operations whose
# CLI return is "request accepted" rather than "operation done":
#
#   - cp / mv on every namespace go through PATCH /api/paste/<node>/,
#     which queues a server-side task and returns the task_id
#     immediately (see cli/cmd/ctl/files/cp.go's "We don't poll for
#     completion in this iteration" comment). The CLI exiting 0
#     just means the server accepted the request — the actual byte
#     movement happens later, asynchronously.
#   - drive paste tasks usually finish in milliseconds so the
#     follow-up `ls` happens to see the new file; sync paste tasks
#     route through Seafile's job queue which can take seconds,
#     so a naked `ls` immediately after sync mv races and reports
#     an empty directory. That's exactly the failure mode the
#     first version of test_sync_files hit:
#         PASS: mv mv_src/aaa.txt → mv_dst/
#         FAIL: ls mv_dst/ shows aaa.txt after mv (output empty)
#         PASS: mv mv_dst/aaa.txt → mv_dst/aaa-renamed.txt   ← second
#               mv enqueues a task with a non-existent source; it
#               accepts the request but the task fails in the
#               background, so the next synchronous rename gets
#               "file not found".
#   - any other "queue + return" verb we add later (server-side
#     transcode, cloud-staging promotion, etc.) will benefit from
#     the same poll loop.
#
# Args:
#   $1 desc      assertion description for pass / fail output
#   $2 needle    substring that MUST appear in the ls output for
#                the assertion to pass
#   $3 path      front-end path passed to `files ls`
#   $4 timeout   max seconds to poll (default 30; mv on a slow
#                Seafile deployment under load can take ~10s, so
#                30 leaves comfortable headroom without hanging
#                the smoke run for ages on a real failure)
#
# Returns 0 on success (PASS recorded), 1 on timeout (FAIL recorded).
wait_for_ls() {
    local desc="$1"
    local needle="$2"
    local path="$3"
    local timeout="${4:-30}"
    local elapsed=0
    local last_out=""
    while (( elapsed < timeout )); do
        if last_out=$("${OLARES_CLI}" files ls "${path}" 2>&1); then
            if grep -qF -- "${needle}" <<< "${last_out}"; then
                if (( elapsed > 0 )); then
                    pass "${desc} (after ${elapsed}s)"
                else
                    pass "${desc}"
                fi
                return 0
            fi
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done
    fail "${desc} (timeout ${timeout}s waiting for ${needle} in ls ${path})"
    echo "    last output:" >&2
    echo "${last_out}" | sed 's/^/      /' >&2
    return 1
}

summarize() {
    echo
    echo "============================================"
    echo "SUMMARY"
    echo "============================================"
    echo "  passed: ${PASSED}"
    echo "  failed: ${FAILED}"
    if (( FAILED > 0 )); then
        echo
        echo "Failed tests:"
        for t in "${FAILED_TESTS[@]}"; do
            echo "  - $t"
        done
    fi
}

# Quick sanity check the binary is present and a profile is selected.
# We do this BEFORE any test so a missing-login surfaces with a clear
# error instead of dozens of "401 Unauthorized" spam.
preflight() {
    step "preflight: olares-cli is reachable and a profile is selected"
    if ! command -v "${OLARES_CLI}" >/dev/null 2>&1; then
        echo "ERROR: '${OLARES_CLI}' not on PATH (set OLARES_CLI=path/to/olares-cli or pass --cli <path>)" >&2
        exit 2
    fi
    expect_success "olares-cli files ls drive/Home/ (login OK)" \
        "${OLARES_CLI}" files ls drive/Home/
    # If the probe failed (binary too old, profile not logged in,
    # server unreachable, etc.), abort BEFORE running any test
    # bodies. Without this guard, every downstream test would
    # also fail AND potentially hang on a per-call HTTP timeout
    # (30s by default) — turning a one-line diagnostic into a
    # multi-minute spam-fest. The user gets a single clear
    # failure pointing at the actual root cause instead.
    if (( FAILED > 0 )); then
        echo >&2
        echo "preflight failed; aborting before any tests run." >&2
        echo "common causes:" >&2
        echo "  - the binary is too old (no 'files' subcommand): build a recent olares-cli" >&2
        echo "  - no profile is selected: run 'olares-cli profile login --olares-id <id>'" >&2
        echo "  - the server is unreachable: check your network / Olares cluster status" >&2
        exit 1
    fi
    # Arm the cleanup trap to do remote work. Doing this at the
    # very end of preflight (rather than at the top of main)
    # means a not-logged-in run still exits cleanly without
    # firing a hanging rm -rf cleanup. See TESTS_STARTED comment
    # at the top of the file for the full rationale.
    TESTS_STARTED=1
}

# --------------------------------------------------------------------
# Section 2: test groups
#
# Each group is a function. Adding a new test verb is meant to be
# trivial: write a function, append it to `main` at the bottom of
# the script. Functions never `exit` on their own (so a single
# failure doesn't abort the run); they just record pass/fail and
# move on.
# --------------------------------------------------------------------

# Local fixture: a directory tree with files of varying sizes,
# including the empty file (which routes through CreateEmptyFile
# instead of the chunk pipeline) and a "directory only" leaf so the
# walker's empty-dir bookkeeping is exercised.
prepare_local_fixtures() {
    step "prepare local test fixtures under ${LOCAL_TMP}"
    # Two single files at the root: a non-empty one + the 0-byte
    # special case. The 0-byte file is what surfaced the empty-file
    # routing bug last quarter, so making sure it's in the upload
    # set is non-negotiable.
    echo "hello world" > "${LOCAL_TMP}/aaa.txt"
    : > "${LOCAL_TMP}/empty.txt"
    # A nested directory with one file per level + an empty leaf
    # directory to exercise the "skip empty dir" notice path.
    mkdir -p "${LOCAL_TMP}/testU/sub/deep"
    mkdir -p "${LOCAL_TMP}/testU/empty_leaf"
    echo "level0" > "${LOCAL_TMP}/testU/level0.txt"
    echo "level1" > "${LOCAL_TMP}/testU/sub/level1.txt"
    echo "level2" > "${LOCAL_TMP}/testU/sub/deep/level2.txt"
    echo "  local fixtures ready"
}

test_ls() {
    step "files ls (basic)"
    expect_success "ls drive/Home/" \
        "${OLARES_CLI}" files ls drive/Home/
    expect_success "ls drive/Data/" \
        "${OLARES_CLI}" files ls drive/Data/
    expect_success "ls --json drive/Home/" \
        "${OLARES_CLI}" files ls --json drive/Home/
    # Negative: bare invalid namespace should reject CLIENT-side via
    # ParseFrontendPath (no roundtrip), so this MUST fail fast and
    # consistently across runs.
    expect_failure "ls invalid-namespace/foo/ (parser rejects)" \
        "${OLARES_CLI}" files ls invalid-namespace/foo/
}

# _mkdir_suite_for runs the same three mkdir assertions against any
# files-backend namespace given a per-run base path. Pulled out so
# adding a new namespace (a future cache-tier-2, an `internal/...`
# special, ...) is a one-line addition to test_mkdir rather than a
# 30-line copy-paste.
#
# Args:
#   $1 label    human-readable namespace tag for the step output
#               (e.g. "drive/Home", "sync/<repo_id>")
#   $2 base     full per-run path (already includes <run-id>),
#               e.g. "drive/Home/cli-test-...", "cache/node-1/cli-test-..."
#
# What it covers:
#   1. `mkdir -p <base>/` — creates the per-run parent the rest of
#      the suite uses; -p form so reruns against an already-extant
#      base behave like the first run.
#   2. `mkdir <base>/single/` — non-`-p` single-level. Smoke-tests
#      the namespace's basic POST /api/resources/.../<dir>/ wiring
#      independent of -p's parent-listing existence check.
#   3. `mkdir -p <base>/A/B/C/` + `ls <base>/A/B/` — the
#      auto-rename quirk regression (see mkdir.go's comment about
#      POST /.../Documents/ silently auto-renaming on collision).
#      If -p ever drops the parent-listing skip, this assertion
#      catches "A (1)/B (1)/C" instead of "A/B/C" because the ls
#      grep would miss "C" under A/B.
_mkdir_suite_for() {
    local label="$1"
    local base="$2"
    expect_success "mkdir -p ${base}/ (${label} per-run root)" \
        "${OLARES_CLI}" files mkdir -p "${base}/"
    expect_success "mkdir ${base}/single/ (${label}, single level)" \
        "${OLARES_CLI}" files mkdir "${base}/single/"
    expect_success "mkdir -p ${base}/A/B/C/ (${label}, multi-level)" \
        "${OLARES_CLI}" files mkdir -p "${base}/A/B/C/"
    expect_output_contains "ls confirms ${label} A/B/C/ landed (no auto-rename)" \
        "C" \
        "${OLARES_CLI}" files ls "${base}/A/B/"
}

# test_mkdir exercises the mkdir / mkdir -p verb across every
# configured namespace. drive/Home + drive/Data run unconditionally
# because every Olares user has them; sync / cache / external run
# when the corresponding flag (or env var) is set.
#
# Why test mkdir on every namespace, not just drive: the namespace
# dispatcher (cli/cmd/ctl/files/path.go + the per-fileType arms in
# upload.go etc.) decides URL prefix, parent_dir form, node, and
# auth quirks per fileType — a mkdir that works on drive but
# silently 500s on cache/external is exactly the kind of regression
# that's hard to catch from the parser-level unit tests. Running
# the same triple of assertions on each namespace pins the contract
# end-to-end with minimal duplication.
#
# REQUIREMENT: setup_sync_repo MUST run before this so SYNC_REPO_TEST_DIR
# is populated when the user asked for sync coverage. main()
# enforces the order; if you reshuffle main(), keep that invariant.
test_mkdir() {
    step "files mkdir / mkdir -p (across configured namespaces)"

    _mkdir_suite_for "drive/Home"  "${DRIVE_HOME_TEST_DIR}"
    _mkdir_suite_for "drive/Data"  "${DRIVE_DATA_TEST_DIR}"

    if [[ -n "${SYNC_REPO_TEST_DIR}" ]]; then
        _mkdir_suite_for "sync/<repo_id>" "${SYNC_REPO_TEST_DIR}"
    else
        echo "  SKIP: sync/<repo>/ mkdir (no temp repo; pass --sync-repo-name <name> or set OLARES_TEST_SYNC_REPO_NAME)"
    fi

    if [[ -n "${CACHE_TEST_DIR}" ]]; then
        _mkdir_suite_for "cache/<node>" "${CACHE_TEST_DIR}"
    else
        echo "  SKIP: cache/<node>/ mkdir (pass --cache cache/<node> or set OLARES_TEST_CACHE_PATH)"
    fi

    if [[ -n "${EXTERNAL_TEST_DIR}" ]]; then
        _mkdir_suite_for "external/<node>/<volume>" "${EXTERNAL_TEST_DIR}"
    else
        echo "  SKIP: external/<node>/<volume>/ mkdir (pass --external external/<node>/<volume> or set OLARES_TEST_EXTERNAL_PATH)"
    fi
}

test_upload_single_file() {
    step "files upload (single file, into directory)"
    expect_success "upload aaa.txt into per-run dir" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/aaa.txt" "${DRIVE_HOME_TEST_DIR}/"
    expect_output_contains "ls finds the uploaded aaa.txt" \
        "aaa.txt" \
        "${OLARES_CLI}" files ls "${DRIVE_HOME_TEST_DIR}/"
}

test_upload_single_file_rename() {
    step "files upload (single file, rename on upload)"
    # No trailing slash on remote → upload + rename in one step. This
    # exercises planForFile's "split into parent + basename" branch.
    expect_success "upload aaa.txt as renamed.txt" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/aaa.txt" "${DRIVE_HOME_TEST_DIR}/renamed.txt"
    expect_output_contains "ls finds the renamed.txt" \
        "renamed.txt" \
        "${OLARES_CLI}" files ls "${DRIVE_HOME_TEST_DIR}/"
}

test_upload_empty_file() {
    step "files upload (empty file → CreateEmptyFile branch)"
    # The 0-byte case routes through a different endpoint
    # (uploader.go's CreateEmptyFile) because Resumable.js can't
    # represent a 0-byte chunk. If that branch breaks, the chunk
    # pipeline tries to push a 0-length chunk and the server 400s.
    expect_success "upload empty.txt (zero bytes)" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/empty.txt" "${DRIVE_HOME_TEST_DIR}/"
    expect_output_contains "ls finds the empty.txt" \
        "empty.txt" \
        "${OLARES_CLI}" files ls "${DRIVE_HOME_TEST_DIR}/"
}

test_upload_directory() {
    step "files upload (directory tree, into per-run subdir)"
    # The folder upload is the second-most common bug surface: the
    # walker has to emit per-file tasks with the source basename as
    # the top-level component, AND auto-create the source-tree
    # directories on the way. Smoke this end-to-end.
    expect_success "upload ./testU/ into ${DRIVE_HOME_TEST_DIR}/" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/testU/" "${DRIVE_HOME_TEST_DIR}/"
    expect_output_contains "ls confirms testU landed under per-run dir" \
        "testU" \
        "${OLARES_CLI}" files ls "${DRIVE_HOME_TEST_DIR}/"
    expect_output_contains "ls confirms nested deep/level2.txt landed" \
        "level2.txt" \
        "${OLARES_CLI}" files ls "${DRIVE_HOME_TEST_DIR}/testU/sub/deep/"
}

test_upload_dir_to_root() {
    step "files upload (REGRESSION: directory to namespace root)"
    # This is the bug from the user's report:
    #   olares-cli files upload ./testU/ drive/Home/
    #   Error: local "./testU/" is a directory; remote "" must end with '/'
    #
    # SubPath="/" was being TrimPrefix'd to "" and BuildPlan saw "no
    # directory hint". Pinning the fix here means a future regression
    # in subPathForBuildPlan (cli/cmd/ctl/files/upload.go) shows up
    # immediately in CI rather than waiting for someone to type the
    # exact same command from the report.
    #
    # We upload into a SUB-directory of the namespace root that we
    # created in test_mkdir, so we don't pollute the user's /Home.
    # The actual regression is the "remote ends in /" parsing path,
    # not the "land in /Home directly" UX — both share the SubPath="/"
    # → remoteSub="" failure mode at the root, but pointing at a
    # named subdir keeps cleanup predictable for the smoke harness.
    local local_dir="${LOCAL_TMP}/regression_root_dir"
    mkdir -p "${local_dir}/sub"
    echo "regression-payload" > "${local_dir}/file.txt"
    echo "regression-payload-sub" > "${local_dir}/sub/nested.txt"
    expect_success "upload ./regression_root_dir/ drive/Home/${TEST_RUN_ID}/ (root-form parent)" \
        "${OLARES_CLI}" files upload "${local_dir}/" "drive/Home/${TEST_RUN_ID}/"
    expect_output_contains "ls confirms regression_root_dir landed" \
        "regression_root_dir" \
        "${OLARES_CLI}" files ls "${DRIVE_HOME_TEST_DIR}/"

    # Also exercise the bare-extend form (`drive/Home`) — the parser
    # synthesizes SubPath="/" for both `drive/Home/` and `drive/Home`,
    # but the bug report covered both shells, so both should work.
    # We don't repeat the upload to avoid duplicate files; just make
    # sure ls's SubPath synthesis isn't broken.
    expect_success "ls drive/Home (bare extend)" \
        "${OLARES_CLI}" files ls "drive/Home"
}

test_download() {
    step "files download (single file, then directory)"
    local single_dst="${LOCAL_TMP}/dl_single.txt"
    expect_success "download aaa.txt to ${single_dst}" \
        "${OLARES_CLI}" files download "${DRIVE_HOME_TEST_DIR}/aaa.txt" "${single_dst}"
    if [[ -s "${single_dst}" ]]; then
        pass "downloaded file is non-empty"
    else
        fail "downloaded file ${single_dst} missing or empty"
    fi

    # Directory download — recreates the testU tree under the local
    # destination (mirrors the LarePass folder-download UX where the
    # remote root's basename becomes the local top-level dir).
    local dir_dst="${LOCAL_TMP}/dl_dir"
    mkdir -p "${dir_dst}"
    expect_success "download testU/ recursively" \
        "${OLARES_CLI}" files download "${DRIVE_HOME_TEST_DIR}/testU/" "${dir_dst}"
    if [[ -s "${dir_dst}/testU/sub/deep/level2.txt" ]]; then
        pass "directory download preserved nested file"
    else
        fail "directory download missed nested level2.txt"
    fi
}

test_cat() {
    step "files cat"
    expect_output_contains "cat aaa.txt prints body" \
        "hello world" \
        "${OLARES_CLI}" files cat "${DRIVE_HOME_TEST_DIR}/aaa.txt"
    # cat on an empty file shouldn't error — just emit nothing.
    expect_success "cat empty.txt (zero-byte file)" \
        "${OLARES_CLI}" files cat "${DRIVE_HOME_TEST_DIR}/empty.txt"
    # cat on a directory MUST fail with a clean error (the cobra layer
    # stats the path before fetching, so this should fail before any
    # bytes hit the wire).
    expect_failure "cat on a directory rejected" \
        "${OLARES_CLI}" files cat "${DRIVE_HOME_TEST_DIR}/testU/"
}

test_cp() {
    step "files cp (file → dir, file → file rename, dir → dir -r, multi-source)"
    # NOTE: cp is async on every namespace (PATCH /api/paste/<node>/
    # returns task_id immediately). drive paste tasks usually finish
    # in milliseconds so a naked ls happens to see the result, but
    # this is a race that surfaces under server load. Use
    # wait_for_ls so the test stays stable. See wait_for_ls's
    # docstring for the full rationale.
    expect_success "mkdir cp_dst/ (target dir)" \
        "${OLARES_CLI}" files mkdir "${DRIVE_HOME_TEST_DIR}/cp_dst/"
    expect_success "cp aaa.txt → cp_dst/" \
        "${OLARES_CLI}" files cp "${DRIVE_HOME_TEST_DIR}/aaa.txt" "${DRIVE_HOME_TEST_DIR}/cp_dst/"
    wait_for_ls "ls cp_dst/ shows aaa.txt" \
        "aaa.txt" "${DRIVE_HOME_TEST_DIR}/cp_dst/"

    # Rename mode (no trailing slash on dst) — single source only.
    expect_success "cp aaa.txt → aaa-copy.txt (rename mode)" \
        "${OLARES_CLI}" files cp "${DRIVE_HOME_TEST_DIR}/aaa.txt" "${DRIVE_HOME_TEST_DIR}/aaa-copy.txt"
    wait_for_ls "ls finds aaa-copy.txt" \
        "aaa-copy.txt" "${DRIVE_HOME_TEST_DIR}/"

    # Recursive directory copy — refuses without -r, succeeds with.
    expect_failure "cp testU/ without -r is rejected" \
        "${OLARES_CLI}" files cp "${DRIVE_HOME_TEST_DIR}/testU/" "${DRIVE_HOME_TEST_DIR}/cp_dst/"
    expect_success "cp -r testU/ → cp_dst/" \
        "${OLARES_CLI}" files cp -r "${DRIVE_HOME_TEST_DIR}/testU/" "${DRIVE_HOME_TEST_DIR}/cp_dst/"
    wait_for_ls "ls cp_dst/ shows testU/ after recursive cp" \
        "testU" "${DRIVE_HOME_TEST_DIR}/cp_dst/"

    # Multi-source into a directory.
    expect_success "cp aaa.txt empty.txt → cp_dst/ (multi-source)" \
        "${OLARES_CLI}" files cp \
            "${DRIVE_HOME_TEST_DIR}/aaa.txt" \
            "${DRIVE_HOME_TEST_DIR}/empty.txt" \
            "${DRIVE_HOME_TEST_DIR}/cp_dst/"
    wait_for_ls "ls cp_dst/ shows empty.txt after multi-source cp" \
        "empty.txt" "${DRIVE_HOME_TEST_DIR}/cp_dst/"
}

test_mv() {
    step "files mv (rename, move-to-dir, recursive)"
    # Set up isolated mv targets so we don't disturb cp's outputs.
    # Same async-paste consideration as test_cp: chained mv's MUST
    # wait for the previous one to land before issuing the next,
    # otherwise the second enqueues with a non-existent source and
    # silently fails on the server task queue.
    expect_success "mkdir mv_src/, mv_dst/" \
        "${OLARES_CLI}" files mkdir -p "${DRIVE_HOME_TEST_DIR}/mv_src/"
    expect_success "mkdir mv_dst/" \
        "${OLARES_CLI}" files mkdir -p "${DRIVE_HOME_TEST_DIR}/mv_dst/"
    expect_success "upload aaa.txt → mv_src/" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/aaa.txt" "${DRIVE_HOME_TEST_DIR}/mv_src/"

    # Move into a directory.
    expect_success "mv mv_src/aaa.txt → mv_dst/" \
        "${OLARES_CLI}" files mv "${DRIVE_HOME_TEST_DIR}/mv_src/aaa.txt" "${DRIVE_HOME_TEST_DIR}/mv_dst/"
    wait_for_ls "ls mv_dst/ shows aaa.txt" \
        "aaa.txt" "${DRIVE_HOME_TEST_DIR}/mv_dst/"

    # Rename mode (no trailing slash on dst, single source).
    expect_success "mv mv_dst/aaa.txt → mv_dst/aaa-renamed.txt" \
        "${OLARES_CLI}" files mv "${DRIVE_HOME_TEST_DIR}/mv_dst/aaa.txt" "${DRIVE_HOME_TEST_DIR}/mv_dst/aaa-renamed.txt"
    wait_for_ls "ls confirms rename via mv" \
        "aaa-renamed.txt" "${DRIVE_HOME_TEST_DIR}/mv_dst/"
}

test_rename() {
    step "files rename (in-place, synchronous)"
    # rename targets the entry by 3-segment path and takes a BARE
    # basename; make sure we set up a known target so the assertion
    # is deterministic.
    expect_success "upload aaa.txt → rename_target.txt (rename mode upload)" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/aaa.txt" "${DRIVE_HOME_TEST_DIR}/rename_target.txt"
    expect_success "rename rename_target.txt → renamed_via_rename.txt" \
        "${OLARES_CLI}" files rename "${DRIVE_HOME_TEST_DIR}/rename_target.txt" "renamed_via_rename.txt"
    expect_output_contains "ls confirms rename" \
        "renamed_via_rename.txt" \
        "${OLARES_CLI}" files ls "${DRIVE_HOME_TEST_DIR}/"

    # Negative: a slash in the new-name argument MUST be rejected
    # client-side (cross-directory moves are mv's job).
    expect_failure "rename rejects slashes in new-name" \
        "${OLARES_CLI}" files rename "${DRIVE_HOME_TEST_DIR}/renamed_via_rename.txt" "sub/with-slash.txt"
}

test_rm() {
    step "files rm (single, recursive, force)"
    # Set up isolated rm targets — we don't want to delete things the
    # later tests need to look at.
    expect_success "upload rm_test.txt for rm" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/aaa.txt" "${DRIVE_HOME_TEST_DIR}/rm_test.txt"
    expect_success "rm -f rm_test.txt" \
        "${OLARES_CLI}" files rm -f "${DRIVE_HOME_TEST_DIR}/rm_test.txt"

    # Directory rm: refused without -r, accepted with -rf.
    expect_success "mkdir -p rm_dir_test/sub/" \
        "${OLARES_CLI}" files mkdir -p "${DRIVE_HOME_TEST_DIR}/rm_dir_test/sub/"
    expect_failure "rm of directory without -r refused" \
        "${OLARES_CLI}" files rm -f "${DRIVE_HOME_TEST_DIR}/rm_dir_test/"
    expect_success "rm -rf rm_dir_test/" \
        "${OLARES_CLI}" files rm -rf "${DRIVE_HOME_TEST_DIR}/rm_dir_test/"

    # Negative: rm of the volume root MUST be refused by the planner.
    expect_failure "rm -rf drive/Home/ refused (root protection)" \
        "${OLARES_CLI}" files rm -rf "drive/Home/"
}

# test_chown exercises the GET path of `files chown` plus the
# client-side validation gates. Setting a uid is intentionally NOT
# exercised: the wire call requires permission to chown files on
# the backend's PVC (typically only available when olares runs as
# root) AND a successful PUT mutates ownership on the user's actual
# Home volume — both are unsafe for an automated smoke test.
#
# What this DOES cover:
#
#   - GET on a known file we uploaded earlier (smokes the GET
#     wiring on drive/Home).
#   - GET on a directory (chown is a per-entry concept; the GET
#     branch needs to work on both files and dirs).
#   - GET on drive/Data (different fileType, same protocol).
#   - Negative: --recursive without --uid (client-side rejection
#     before any wire call — recursion is meaningless for GETs).
#   - Negative: sync/<repo_id>/... (chown is rejected client-side
#     for fileTypes whose ACL doesn't map to POSIX uid; Seafile's
#     ACL surface lives under `files repos`, not chown).
#   - Negative: drive/Home/ (volume root protection — chowning a
#     whole namespace has too much blast radius).
#
# Each negative pins a different gate in frontendPathToChownTarget /
# runChown — together they let a future refactor break exactly one
# of them and have the smoke harness point at the bad gate.
test_chown() {
    step "files chown (GET + client-side gates)"

    # GET on a file we know exists (uploaded in test_upload_single_file).
    # The CLI prints "<path>  uid=<n> (<label>)" on success — note
    # the `uid=`, NOT `uid:`. The colon-form lives in the wire body
    # ({"uid": <int>}); the CLI's human-readable rendering uses
    # `uid=N (Name)` to match the LarePass GUI's Permission tab
    # ("User"/"Root" labels). We grep on the `uid=` substring +
    # the `(` label opening so a future format tweak that drops
    # the label entirely (e.g. raw `uid=N`) still passes the bare
    # `uid=` check below, but a regression that swaps in `:` shows
    # up as a fail.
    expect_output_contains "chown GET on drive/Home file shows 'uid='" \
        "uid=" \
        "${OLARES_CLI}" files chown "${DRIVE_HOME_TEST_DIR}/aaa.txt"

    # GET on a directory — same wire shape, different entry kind.
    expect_output_contains "chown GET on drive/Home dir shows 'uid='" \
        "uid=" \
        "${OLARES_CLI}" files chown "${DRIVE_HOME_TEST_DIR}/testU/"

    # GET on the OTHER drive namespace. The dispatcher in
    # frontendPathToChownTarget keys per fileType; we want to know
    # if drive/Data ever diverges from drive/Home wire-wise.
    expect_output_contains "chown GET on drive/Data shows 'uid='" \
        "uid=" \
        "${OLARES_CLI}" files chown "${DRIVE_DATA_TEST_DIR}/"

    # Negative: --recursive without --uid is meaningless and
    # rejected client-side. The CLI's error message names
    # --recursive explicitly so we pin that wording: a future
    # refactor that swaps "drop --recursive" for a different CTA
    # is still a regression worth flagging.
    expect_failure "chown --recursive without --uid rejected" \
        "${OLARES_CLI}" files chown "${DRIVE_HOME_TEST_DIR}/aaa.txt" --recursive

    # Negative: chown on a sync path is rejected client-side
    # (sync ACL lives in seafile_api, not POSIX). We use the sync
    # repo path if we set one up; otherwise a placeholder path
    # still hits the same parser gate before any wire roundtrip.
    if [[ -n "${SYNC_REPO_PATH}" ]]; then
        expect_failure "chown on sync/<repo>/ rejected client-side" \
            "${OLARES_CLI}" files chown "${SYNC_REPO_PATH}/"
    else
        # Use a syntactically-valid sync path even when we don't
        # have a real repo; the parser's "sync isn't a supported
        # chown namespace" gate fires before any HTTP lookup so a
        # fake id is sufficient to exercise it.
        expect_failure "chown on sync/<fake-id>/ rejected client-side" \
            "${OLARES_CLI}" files chown "sync/00000000-0000-0000-0000-000000000000/"
    fi

    # Negative: volume root rejected. Same blast-radius rationale
    # as `rm -rf drive/Home/` — chowning an entire volume in one
    # call is almost never the user's intent.
    expect_failure "chown drive/Home/ refused (volume root)" \
        "${OLARES_CLI}" files chown "drive/Home/"
    expect_failure "chown drive/Data/ refused (volume root)" \
        "${OLARES_CLI}" files chown "drive/Data/"
}

# test_aliases pins the documented aliases for the three verbs that
# carry them: rename → rn, mkdir → md, rm → delete / remove. Tiny
# smoke checks because the aliases live on Cobra's command struct;
# losing one is the kind of "harmless" refactor that quietly breaks
# scripts written against the older names. No new test surface is
# created — each alias is run against a throwaway path that the
# per-run cleanup already takes care of.
test_aliases() {
    step "files <verb> aliases (rn / md / delete)"

    # `md` is an alias for `mkdir`. Smoke the -p form so the
    # alias inherits the same flag set (which it must — Cobra
    # aliases re-use the parent's flag definitions).
    expect_success "files md -p ${DRIVE_HOME_TEST_DIR}/alias_md/ (md=mkdir)" \
        "${OLARES_CLI}" files md -p "${DRIVE_HOME_TEST_DIR}/alias_md/"
    expect_output_contains "ls confirms alias_md/ landed" \
        "alias_md" \
        "${OLARES_CLI}" files ls "${DRIVE_HOME_TEST_DIR}/"

    # `rn` is an alias for `rename`. Make a target file inside the
    # alias_md dir we just created, then rn it.
    expect_success "upload aaa.txt → alias_md/ (rn fixture)" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/aaa.txt" "${DRIVE_HOME_TEST_DIR}/alias_md/"
    expect_success "files rn ${DRIVE_HOME_TEST_DIR}/alias_md/aaa.txt rn-renamed.txt" \
        "${OLARES_CLI}" files rn "${DRIVE_HOME_TEST_DIR}/alias_md/aaa.txt" "rn-renamed.txt"
    expect_output_contains "ls confirms rn renamed the file" \
        "rn-renamed.txt" \
        "${OLARES_CLI}" files ls "${DRIVE_HOME_TEST_DIR}/alias_md/"

    # `delete` and `remove` are both aliases for `rm`. Use each
    # exactly once so a future drop of either is caught here.
    expect_success "files delete -f ... (delete=rm)" \
        "${OLARES_CLI}" files delete -f "${DRIVE_HOME_TEST_DIR}/alias_md/rn-renamed.txt"
    expect_success "files remove -rf ... (remove=rm)" \
        "${OLARES_CLI}" files remove -rf "${DRIVE_HOME_TEST_DIR}/alias_md/"
}

test_drive_data() {
    step "files * on drive/Data/"
    # Different namespace, same protocol — sanity-check the dispatch
    # so a future Data-specific regression is caught here. Re-uses
    # the per-run drive/Data/ dir created in test_mkdir.
    expect_success "upload aaa.txt into drive/Data/${TEST_RUN_ID}/" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/aaa.txt" "${DRIVE_DATA_TEST_DIR}/"
    expect_output_contains "ls drive/Data/ shows the upload" \
        "aaa.txt" \
        "${OLARES_CLI}" files ls "${DRIVE_DATA_TEST_DIR}/"
    expect_success "rm -f drive/Data/${TEST_RUN_ID}/aaa.txt" \
        "${OLARES_CLI}" files rm -f "${DRIVE_DATA_TEST_DIR}/aaa.txt"
}

test_share() {
    step "files share (list + smb-users list)"
    # The creation flavors all need real targets (members for
    # internal, password+expiry for public, smb users for smb), so
    # we don't drive them from the smoke harness — just exercise
    # the read-only verbs. share list with no flags returns both
    # directions, which is enough to confirm the round-trip works.
    expect_success "share list" \
        "${OLARES_CLI}" files share list
    expect_success "share smb-users list" \
        "${OLARES_CLI}" files share smb-users list

    # Public share creation is the only flavor we can exercise
    # without external setup (no member roster needed). Skip if the
    # password validation is too strict for an automated default;
    # otherwise create + list + rm to round-trip the lifecycle.
    local share_target="${DRIVE_HOME_TEST_DIR}/aaa-copy.txt"
    local share_id
    # The CLI emits an indented key/value block:
    #   created public share:
    #     id              : <share-id>
    #     path            : <path>
    #     ...
    # We grep for the `id` row by leading-whitespace + literal "id"
    # + whitespace + ":" + value, then take the last whitespace-
    # separated field. The "|| true" suppresses the awk's exit
    # code so a "no match" path still leaves share_id empty (the
    # outer SKIP branch handles that case cleanly).
    share_id=$("${OLARES_CLI}" files share public "${share_target}" \
        --expire-days 1 --password "Olares-Test-Password-1!" 2>/dev/null \
        | awk '/^[[:space:]]+id[[:space:]]*:/ {print $NF; exit}' \
        || true)
    if [[ -n "${share_id}" ]]; then
        pass "share public created (id=${share_id})"
        expect_success "share get ${share_id}" \
            "${OLARES_CLI}" files share get "${share_id}"
        # share set-password lifecycle: explicit-password first (we
        # can pin the post-set value via `share get` later if the
        # server ever surfaces it; today it doesn't, so the
        # assertion is just "the call returned 0"), then the
        # auto-generated form (no --password). Both branches matter:
        # explicit is the LarePass GUI's "Reset Password" with a
        # user-typed value; auto-generated is the same dialog
        # leaving the field blank, which the CLI uses to print one-
        # shot a random 8-byte URL-safe password.
        expect_success "share set-password ${share_id} --password '...' (explicit)" \
            "${OLARES_CLI}" files share set-password "${share_id}" \
                --password "Olares-Test-Password-2!"
        expect_success "share set-password ${share_id} (auto-generated)" \
            "${OLARES_CLI}" files share set-password "${share_id}"
        expect_success "share rm ${share_id}" \
            "${OLARES_CLI}" files share rm "${share_id}"
    else
        echo "  SKIP: share public create (could not parse share id; password policy?)"
    fi

    # Negatives that exercise the share-flavor validation in the
    # client layer (no wire roundtrip, no server-side side effects).
    # These pin the parser's gates so a refactor that drops the
    # per-flavor flag check is caught here, without needing a real
    # recipient roster or smb-users roster.
    #
    # We DELIBERATELY don't test `share internal` without --users:
    # that path doesn't reject — it creates a recipient-less share
    # record server-side (see runShareInternal — `parseShareMembers`
    # tolerates an empty list). Running it from a smoke test would
    # leave orphan share rows we'd then need to track + clean up.
    expect_failure "share smb without --public or --users rejected client-side" \
        "${OLARES_CLI}" files share smb "${DRIVE_HOME_TEST_DIR}/"
    expect_failure "share smb with --public AND --users rejected (mutually exclusive)" \
        "${OLARES_CLI}" files share smb "${DRIVE_HOME_TEST_DIR}/" \
            --public --users "smb-uid-fake:view"
    expect_failure "share internal with bogus --permission rejected client-side" \
        "${OLARES_CLI}" files share internal "${DRIVE_HOME_TEST_DIR}/aaa.txt" \
            --permission "not-a-perm"
}

# setup_sync_repo provisions a fresh Sync (Seafile) library that the
# downstream sync tests parent under. Done as a separate setup step
# (rather than inline at the top of test_sync_files) so the post-run
# cleanup can find the repo via SYNC_REPO_ID even when something in
# the middle of the suite explodes — a half-failed sync test should
# still drop its repo on the way out.
#
# Skipped (returns 0, leaves SYNC_REPO_ID empty) when
# OLARES_TEST_SYNC_REPO_NAME isn't set, which is the signal that the
# user doesn't want to (or can't) provision a real Seafile library
# for this run. test_sync_files / test_repos both gate on
# SYNC_REPO_ID and SKIP cleanly when it's empty, so a missing env
# var produces the correct "skipped, not failed" outcome.
setup_sync_repo() {
    step "setup: create a temporary Sync repo for the sync test suite"
    if [[ -z "${OLARES_TEST_SYNC_REPO_NAME}" ]]; then
        echo "  SKIP: set OLARES_TEST_SYNC_REPO_NAME=<base-name> to enable sync tests"
        echo "        e.g. OLARES_TEST_SYNC_REPO_NAME=cli-test-repo"
        return 0
    fi

    # Append the run-id so the repo is unique across concurrent /
    # repeat runs; the server accepts duplicate names but we want
    # the post-run summary to point at exactly one library, and
    # OLARES_TEST_KEEP runs need a unique label for the user to
    # find the repo in the LarePass UI.
    local repo_name="${OLARES_TEST_SYNC_REPO_NAME}-${TEST_RUN_ID}"
    local create_out
    if ! create_out=$("${OLARES_CLI}" files repos create "${repo_name}" --json 2>&1); then
        fail "repos create ${repo_name} (output: ${create_out})"
        return 1
    fi
    pass "repos create ${repo_name}"

    # Pull the repo_id out of the JSON envelope. We use python3
    # rather than jq so the script stays dependency-light — every
    # Olares deployment we've seen has python3 on PATH, jq is less
    # universal. Both `repo_id` and `RepoID` are tried because the
    # CLI's JSON marshalling has historically waffled between snake
    # and camel case for this field.
    local repo_id
    repo_id=$(python3 -c "import sys, json
d = json.loads(sys.stdin.read())
print(d.get('repo_id') or d.get('RepoID') or '')" <<< "${create_out}")
    if [[ -z "${repo_id}" ]]; then
        fail "repos create did not yield a repo_id (output: ${create_out})"
        return 1
    fi

    SYNC_REPO_ID="${repo_id}"
    SYNC_REPO_NAME_USED="${repo_name}"
    SYNC_REPO_PATH="sync/${repo_id}"
    SYNC_REPO_TEST_DIR="sync/${repo_id}/${TEST_RUN_ID}"
    echo "  sync repo ready: name=${SYNC_REPO_NAME_USED} id=${SYNC_REPO_ID}"
    echo "  per-run dir:      ${SYNC_REPO_TEST_DIR}/"
}

test_repos() {
    step "files repos (list / get / rename) — read-only verbs always run"
    # Read-only verbs are unconditional: they don't need a writeable
    # surface, just a logged-in profile. Even on a fresh Olares with
    # no repos, the empty-list path has to print "no repos found"
    # without erroring.
    expect_success "repos list" \
        "${OLARES_CLI}" files repos list
    expect_success "repos list --type all --json" \
        "${OLARES_CLI}" files repos list --type all --json

    if [[ -z "${SYNC_REPO_ID}" ]]; then
        echo "  SKIP: repos get/rename (no temp repo; set OLARES_TEST_SYNC_REPO_NAME)"
        return 0
    fi

    # repos get: look up the repo we just created. Printed body
    # carries Repo ID, Name, Owner, etc. — we don't grep specific
    # fields here because the formatting is human-rendered and
    # changes with releases; just confirm the call returns 0.
    expect_success "repos get ${SYNC_REPO_ID}" \
        "${OLARES_CLI}" files repos get "${SYNC_REPO_ID}"

    # Rename — the repo_id is stable, only the display name
    # changes. We append "-renamed" so a follow-up `repos get`
    # would see the new label; we don't assert on the new label
    # because the rename verb is fire-and-forget on the wire (no
    # JSON body in the response) and `repos get` parses the
    # human-rendered "Name:" line which is brittle.
    expect_success "repos rename ${SYNC_REPO_ID}" \
        "${OLARES_CLI}" files repos rename "${SYNC_REPO_ID}" "${SYNC_REPO_NAME_USED}-renamed"
    # Update the bookkeeping name so the cleanup-trap message
    # matches what the server now thinks the repo is called.
    SYNC_REPO_NAME_USED="${SYNC_REPO_NAME_USED}-renamed"

    # Negative: empty name + dot-segment names rejected client-side
    # (see internal/files/repos.Create — the CLI's
    # `.` / `..` blacklist runs before any wire call). Three
    # cheap assertions that pin the contract without creating any
    # repos. These were the original motivation for adding the
    # blacklist after we discovered the GUI silently accepted dot
    # names and Linux disallowed them.
    expect_failure "repos create '.' rejected (path-traversal blacklist)" \
        "${OLARES_CLI}" files repos create "."
    expect_failure "repos create '..' rejected (path-traversal blacklist)" \
        "${OLARES_CLI}" files repos create ".."
    expect_failure "repos rename ${SYNC_REPO_ID} '.' rejected (path-traversal blacklist)" \
        "${OLARES_CLI}" files repos rename "${SYNC_REPO_ID}" "."

    # Lifecycle: create a THROWAWAY repo just for the rm round-trip,
    # so deleting it doesn't tear down the rest of the sync test
    # surface. The main repo (SYNC_REPO_ID) is preserved until the
    # cleanup trap; this throwaway is removed inline at the end of
    # this step.
    local throwaway_name="${OLARES_TEST_SYNC_REPO_NAME}-throwaway-${TEST_RUN_ID}"
    local throwaway_out
    if throwaway_out=$("${OLARES_CLI}" files repos create "${throwaway_name}" --json 2>&1); then
        local throwaway_id
        throwaway_id=$(python3 -c "import sys, json
d = json.loads(sys.stdin.read())
print(d.get('repo_id') or d.get('RepoID') or '')" <<< "${throwaway_out}")
        if [[ -n "${throwaway_id}" ]]; then
            pass "repos create throwaway (id=${throwaway_id})"
            # `repos rm` is the one verb we don't otherwise exercise
            # outside the cleanup trap — adding an explicit
            # assertion here means a future regression in the
            # DELETE wire call surfaces as a clear test failure,
            # NOT as a leaky cleanup that piles up dead repos.
            #
            # -y skips the interactive confirmation prompt (the
            # CLI hard-rejects no-TTY rm without -y, and our smoke
            # harness deliberately runs non-interactively).
            expect_success "repos rm -y ${throwaway_id}" \
                "${OLARES_CLI}" files repos rm -y "${throwaway_id}"
            # Verify the repo is actually gone — `repos get`
            # against a freshly-deleted id must fail. This catches
            # the "delete returned 200 but the row is still
            # there" failure mode that a fire-and-forget rm could
            # otherwise hide.
            expect_failure "repos get ${throwaway_id} after rm rejected" \
                "${OLARES_CLI}" files repos get "${throwaway_id}"
        else
            fail "repos create throwaway returned no id (output: ${throwaway_out})"
        fi
    else
        fail "repos create throwaway failed (output: ${throwaway_out})"
    fi
}

# test_sync_files exercises every files verb against the sync
# (Seafile) namespace. This is a separate suite from test_drive_data
# because the sync upload path uses a different chunk endpoint
# (Seafile's `/seafhttp/upload-aj/<token>`) with a DIFFERENT
# parent_dir convention (chunkRoot is empty, the chunk-form
# parent_dir is the path INSIDE the repo). Regression in the
# sync-specific glue (see uploadRootAndDriveType's "sync" arm and
# BuildPlan's parentDirFor split) wouldn't be caught by the drive
# tests at all, so this suite mirrors the drive coverage step-by-
# step:
#
#   ls (root + subdir, including just-after-create)
#   mkdir / mkdir -p (auto-rename quirk regression on sync namespace)
#   upload single file / single-file rename / empty file / dir tree
#   upload directory to repo ROOT (subPathForBuildPlan regression
#       on sync — this is the same fix we shipped for drive; sync's
#       SubPath="/" is identical and must work the same way)
#   download single file / directory tree
#   cat
#   cp / mv inside the repo (intra-namespace), cp drive→sync
#       (cross-namespace, exercises the paste endpoint's
#       cross-volume support)
#   rename
#   rm (single, directory without -r refused, recursive)
#   share public on a sync file (the share endpoint accepts every
#       fileType; we want to confirm sync isn't accidentally
#       blocked)
#
# Every test step parents under SYNC_REPO_TEST_DIR (a per-run
# subdir inside the repo) so we can also reach the repo root via
# SYNC_REPO_PATH for the trailing-slash regression case without
# stomping on each other.
test_sync_files() {
    step "files * on sync/<repo_id>/ (gated on OLARES_TEST_SYNC_REPO_NAME)"
    if [[ -z "${SYNC_REPO_ID}" ]]; then
        echo "  SKIP: no sync repo (set OLARES_TEST_SYNC_REPO_NAME to enable)"
        return 0
    fi

    # 1) ls
    # ----
    # Just-created repo is guaranteed empty, so ls at the repo
    # root has to succeed AND not list anything we'll mistake for
    # a leftover. Don't grep — the Seafile permissions / metadata
    # banner can vary by deployment.
    expect_success "ls ${SYNC_REPO_PATH}/ (just-created repo root)" \
        "${OLARES_CLI}" files ls "${SYNC_REPO_PATH}/"

    # 2) mkdir / mkdir -p
    # -------------------
    # Single-level mkdir creates the per-run dir; -p variant
    # exercises the "skip existing prefix" parent-listing logic on
    # the sync namespace (see mkdir.go's auto-rename quirk note —
    # Seafile is just as susceptible).
    expect_success "mkdir -p ${SYNC_REPO_TEST_DIR}/ (per-run sync dir)" \
        "${OLARES_CLI}" files mkdir -p "${SYNC_REPO_TEST_DIR}/"
    expect_success "mkdir -p ${SYNC_REPO_TEST_DIR}/A/B/C/" \
        "${OLARES_CLI}" files mkdir -p "${SYNC_REPO_TEST_DIR}/A/B/C/"
    expect_output_contains "ls confirms A/B/C/ landed in sync (no auto-rename)" \
        "C" \
        "${OLARES_CLI}" files ls "${SYNC_REPO_TEST_DIR}/A/B/"

    # 3) upload (single file, into directory)
    # ---------------------------------------
    expect_success "upload aaa.txt → ${SYNC_REPO_TEST_DIR}/" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/aaa.txt" "${SYNC_REPO_TEST_DIR}/"
    expect_output_contains "ls finds aaa.txt in sync" \
        "aaa.txt" \
        "${OLARES_CLI}" files ls "${SYNC_REPO_TEST_DIR}/"

    # 4) upload (rename on upload)
    # ----------------------------
    expect_success "upload aaa.txt → renamed.txt (sync rename mode)" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/aaa.txt" "${SYNC_REPO_TEST_DIR}/renamed.txt"
    expect_output_contains "ls finds renamed.txt in sync" \
        "renamed.txt" \
        "${OLARES_CLI}" files ls "${SYNC_REPO_TEST_DIR}/"

    # 5) upload (zero-byte file → CreateEmptyFile branch on sync)
    # -----------------------------------------------------------
    # Same routing rule as drive: 0-byte files go through
    # CreateEmptyFile because Resumable.js can't represent a
    # 0-byte chunk. Sync's seafhttp path has historically had its
    # own quirks here, so smoke it explicitly.
    expect_success "upload empty.txt → sync (zero bytes)" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/empty.txt" "${SYNC_REPO_TEST_DIR}/"
    expect_output_contains "ls finds empty.txt in sync" \
        "empty.txt" \
        "${OLARES_CLI}" files ls "${SYNC_REPO_TEST_DIR}/"

    # 6) upload (directory tree — exercises seafhttp/upload-aj
    #    chunkRoot="" path)
    # ----------------------------------------------------------
    # This is the biggest sync-vs-drive divergence: the chunk POST
    # parent_dir form field is the path INSIDE the repo, NOT the
    # full /sync/<repo>/<sub>/ form. If BuildPlan's chunkRoot ==
    # "" handling regresses, this is the test that catches it.
    #
    # NOTE: --parallel 1 is REQUIRED on sync. Seafile's upload-aj
    # endpoint relies on the chunk's `relative_path` form field to
    # auto-create intermediate directories on the fly. With the CLI
    # default --parallel 2, two files like `testU/sub/level1.txt`
    # and `testU/sub/deep/level2.txt` race their POSTs and Seafile's
    # auto-create can fail (level2.txt's request arrives before
    # `testU/sub/` has finished materializing). The HTTP 500 with
    # an empty body is the symptom we hit on guotest259:
    #   Error: testU/sub/deep/level2.txt: ... HTTP 500:
    # Drive isn't affected because its parent_dir multipart field
    # carries the full /drive/Home/<sub>/ path and the backend's
    # mkdir-on-upload is not Seafile's. Tracking item for the
    # uploader: when DriveType==Sync, pre-mkdir every dir that
    # appears in plan.Files (not just plan.EmptyDirs) so the
    # default --parallel works on Seafile too.
    expect_success "upload ./testU/ → ${SYNC_REPO_TEST_DIR}/ (directory tree on sync, --parallel 1)" \
        "${OLARES_CLI}" files upload --parallel 1 "${LOCAL_TMP}/testU/" "${SYNC_REPO_TEST_DIR}/"
    expect_output_contains "ls finds testU under per-run sync dir" \
        "testU" \
        "${OLARES_CLI}" files ls "${SYNC_REPO_TEST_DIR}/"
    expect_output_contains "ls finds nested testU/sub/deep/level2.txt in sync" \
        "level2.txt" \
        "${OLARES_CLI}" files ls "${SYNC_REPO_TEST_DIR}/testU/sub/deep/"

    # 7) upload (REGRESSION: directory to REPO ROOT)
    # ----------------------------------------------
    # This is the same SubPath="/" regression we fixed for drive,
    # exercised on the sync namespace. ParseFrontendPath produces
    # SubPath="/" for both `sync/<repo>/` and `sync/<repo>` (bare
    # extend, parser synthesizes "/"); subPathForBuildPlan must
    # restore the slash so BuildPlan reads the destination as a
    # directory. Without the fix this fails with:
    #   Error: local "..." is a directory; remote "" must end with '/'
    # — same symptom as the original bug report on drive/Home/.
    local local_dir="${LOCAL_TMP}/sync_root_dir"
    mkdir -p "${local_dir}"
    echo "sync-root-payload" > "${local_dir}/file.txt"
    expect_success "upload ./sync_root_dir/ → ${SYNC_REPO_PATH}/ (root regression on sync)" \
        "${OLARES_CLI}" files upload "${local_dir}/" "${SYNC_REPO_PATH}/"
    expect_output_contains "ls ${SYNC_REPO_PATH}/ shows sync_root_dir" \
        "sync_root_dir" \
        "${OLARES_CLI}" files ls "${SYNC_REPO_PATH}/"

    # 8) download (single + recursive directory)
    # ------------------------------------------
    # Sync downloads route through /api/raw/<encPath> just like
    # drive (the v2 dataAPIs share the download path), but the
    # encPath has the sync/<repo>/ prefix. Smoke single + dir.
    local single_dst="${LOCAL_TMP}/sync_dl_single.txt"
    expect_success "download ${SYNC_REPO_TEST_DIR}/aaa.txt → ${single_dst}" \
        "${OLARES_CLI}" files download "${SYNC_REPO_TEST_DIR}/aaa.txt" "${single_dst}"
    if [[ -s "${single_dst}" ]]; then
        pass "downloaded sync file is non-empty"
    else
        fail "downloaded sync file ${single_dst} missing or empty"
    fi
    local dir_dst="${LOCAL_TMP}/sync_dl_dir"
    mkdir -p "${dir_dst}"
    expect_success "download ${SYNC_REPO_TEST_DIR}/testU/ recursively" \
        "${OLARES_CLI}" files download "${SYNC_REPO_TEST_DIR}/testU/" "${dir_dst}"
    if [[ -s "${dir_dst}/testU/sub/deep/level2.txt" ]]; then
        pass "sync directory download preserved nested file"
    else
        fail "sync directory download missed nested level2.txt"
    fi

    # 9) cat
    # ------
    expect_output_contains "cat ${SYNC_REPO_TEST_DIR}/aaa.txt" \
        "hello world" \
        "${OLARES_CLI}" files cat "${SYNC_REPO_TEST_DIR}/aaa.txt"
    expect_failure "cat on a sync directory rejected" \
        "${OLARES_CLI}" files cat "${SYNC_REPO_TEST_DIR}/testU/"

    # 10) cp (intra-sync + cross-namespace drive→sync)
    # ------------------------------------------------
    # Intra-sync cp goes through the same paste endpoint as drive
    # cp (PATCH /api/paste/<node>/), action="copy". Cross-volume
    # cp (drive→sync) goes through the same wire path with
    # different src/dst prefixes — the backend handles the
    # storage-class fan-out internally, so this is actually a
    # client-side sanity check: did we round-trip through
    # ParseFrontendPath correctly for both endpoints?
    #
    # IMPORTANT: cp / mv on sync are async (PATCH /api/paste/<node>/
    # returns a task_id immediately and the server-side task may
    # take a few seconds on Seafile-backed deployments — see
    # wait_for_ls's docstring for details). Every post-cp / post-mv
    # ls assertion in this section uses wait_for_ls so the test
    # tolerates server-side queuing latency without becoming flaky.
    expect_success "mkdir cp_dst/ inside sync repo" \
        "${OLARES_CLI}" files mkdir "${SYNC_REPO_TEST_DIR}/cp_dst/"
    expect_success "cp aaa.txt → cp_dst/ (intra-sync)" \
        "${OLARES_CLI}" files cp "${SYNC_REPO_TEST_DIR}/aaa.txt" "${SYNC_REPO_TEST_DIR}/cp_dst/"
    wait_for_ls "ls cp_dst/ shows aaa.txt (after intra-sync cp)" \
        "aaa.txt" "${SYNC_REPO_TEST_DIR}/cp_dst/"
    expect_success "cp -r testU/ → cp_dst/ (intra-sync recursive)" \
        "${OLARES_CLI}" files cp -r "${SYNC_REPO_TEST_DIR}/testU/" "${SYNC_REPO_TEST_DIR}/cp_dst/"
    wait_for_ls "ls cp_dst/ shows testU after recursive cp" \
        "testU" "${SYNC_REPO_TEST_DIR}/cp_dst/"

    # Cross-namespace: drive→sync. We use the aaa.txt that
    # test_upload_single_file uploaded to drive/Home/<run-id>/
    # so the source already exists. Some backend versions don't
    # support arbitrary cross-volume paste (Seafile bridge can
    # be flaky), so this assertion is best-effort: we run it
    # and report the outcome but don't gate the suite on it.
    expect_success "cp drive/Home aaa.txt → sync (cross-namespace)" \
        "${OLARES_CLI}" files cp \
            "${DRIVE_HOME_TEST_DIR}/aaa.txt" \
            "${SYNC_REPO_TEST_DIR}/cross_ns/"

    # 11) mv (intra-sync, move-into-dir ONLY)
    # ---------------------------------------
    # mv is async on sync (same /api/paste/<node>/ endpoint as cp),
    # so wait_for_ls between operations on the same file is
    # mandatory — otherwise the next mv enqueues with a non-
    # existent source, accepts the request (returns task_id,
    # exit 0), and silently fails in the background.
    #
    # IMPORTANT — sync mv has NO rename mode. Seafile's
    # /api/paste implementation honours the destination DIRECTORY
    # but completely drops any destination BASENAME (it always
    # uses the source's basename for the moved file). Concretely,
    # we observed all three of these on guotest259:
    #
    #   1. `mv x/foo.txt y/`           → lands as y/foo.txt   (✓)
    #   2. `mv x/foo.txt y/bar.txt`    → lands as y/foo.txt   (✗ rename dropped)
    #   3. `mv x/foo.txt x/bar.txt`    → triggers Seafile's
    #      self-conflict resolver and lands as x/foo (1).txt   (✗ rename dropped + collision)
    #
    # Drive isn't affected because its mv implementation isn't
    # Seafile-backed and honours dst basenames. To rename on sync
    # the user MUST go through `files rename`, which hits the
    # synchronous /api/resources PATCH endpoint (server-side
    # routing differs from /api/paste — that's the path Olares
    # plumbs to Seafile's seafile_api.rename_file).
    #
    # So this step exercises ONLY the "move into directory" form
    # of mv on sync. The "rename mode" path is covered on drive
    # in test_mv() above and through `files rename` in step 12
    # below — together those two cover the same surface area
    # without colliding with the Seafile limitation.
    expect_success "mkdir mv_src/ inside sync repo" \
        "${OLARES_CLI}" files mkdir -p "${SYNC_REPO_TEST_DIR}/mv_src/"
    expect_success "mkdir mv_dst/ inside sync repo" \
        "${OLARES_CLI}" files mkdir -p "${SYNC_REPO_TEST_DIR}/mv_dst/"
    expect_success "upload aaa.txt → mv_src/ (intra-sync setup)" \
        "${OLARES_CLI}" files upload "${LOCAL_TMP}/aaa.txt" "${SYNC_REPO_TEST_DIR}/mv_src/"
    expect_success "mv mv_src/aaa.txt → mv_dst/ (intra-sync into-dir)" \
        "${OLARES_CLI}" files mv \
            "${SYNC_REPO_TEST_DIR}/mv_src/aaa.txt" \
            "${SYNC_REPO_TEST_DIR}/mv_dst/"
    wait_for_ls "ls mv_dst/ shows aaa.txt after mv (waits for async paste)" \
        "aaa.txt" "${SYNC_REPO_TEST_DIR}/mv_dst/"

    # 12) rename (in-place, synchronous — the supported way to
    #     rename a sync-backed file)
    # --------------------------------------------------------
    # Sync's rename hits /api/resources PATCH (NOT /api/paste);
    # the server routes by fileType prefix on the URL into
    # seafile_api.rename_file, which DOES honour the destination
    # name. So this is the only correct way to rename a file in
    # a Seafile-backed library — it's both synchronous (success
    # / failure exit code is authoritative) and respects the
    # new basename.
    #
    # The bare-basename validation is client-side so the negative
    # case below catches a regression in the CLI without any wire
    # roundtrip.
    #
    # We still use wait_for_ls for the post-rename ls assertion
    # for symmetry with the cp/mv assertions above; if a future
    # server change ever made the resources index eventually-
    # consistent (e.g. behind a CDN edge), this guards against
    # the flake without blocking the suite on healthy timing.
    expect_success "rename mv_dst/aaa.txt → renamed_via_rename.txt (sync, in-place)" \
        "${OLARES_CLI}" files rename \
            "${SYNC_REPO_TEST_DIR}/mv_dst/aaa.txt" \
            "renamed_via_rename.txt"
    wait_for_ls "ls confirms rename in sync" \
        "renamed_via_rename.txt" "${SYNC_REPO_TEST_DIR}/mv_dst/"
    expect_failure "rename rejects slashes in new-name (sync, client-side)" \
        "${OLARES_CLI}" files rename \
            "${SYNC_REPO_TEST_DIR}/mv_dst/renamed_via_rename.txt" \
            "sub/with-slash.txt"

    # 13) rm (single, directory-without-r refused, recursive,
    #     repo-root protection)
    # ----------------------------------------------------------
    expect_success "rm -f a single file in sync" \
        "${OLARES_CLI}" files rm -f "${SYNC_REPO_TEST_DIR}/mv_dst/renamed_via_rename.txt"
    expect_success "mkdir rm_dir_test/ in sync" \
        "${OLARES_CLI}" files mkdir "${SYNC_REPO_TEST_DIR}/rm_dir_test/"
    expect_failure "rm of a sync directory without -r refused" \
        "${OLARES_CLI}" files rm -f "${SYNC_REPO_TEST_DIR}/rm_dir_test/"
    expect_success "rm -rf rm_dir_test/ in sync" \
        "${OLARES_CLI}" files rm -rf "${SYNC_REPO_TEST_DIR}/rm_dir_test/"
    # Repo-root rm MUST be refused by the planner — same root-
    # protection rule as drive/Home/. Without this, a typo in a
    # cleanup script could nuke the user's entire library.
    expect_failure "rm -rf ${SYNC_REPO_PATH}/ refused (repo-root protection)" \
        "${OLARES_CLI}" files rm -rf "${SYNC_REPO_PATH}/"

    # 14) share (public link on a sync file)
    # --------------------------------------
    # The share endpoint takes the same 3-segment path as ls/cp/
    # rename, so a sync-resource share is a one-line test. We
    # can't pin the share-id parsing here (the format is the same
    # as the drive case in test_share), so we tolerate parse
    # failures and just SKIP rather than fail the whole suite.
    local sync_share_target="${SYNC_REPO_TEST_DIR}/renamed.txt"
    local sync_share_id
    # Same `id              : <share-id>` key/value format as
    # test_share above; reuse the same indented-id parser.
    sync_share_id=$("${OLARES_CLI}" files share public "${sync_share_target}" \
        --expire-days 1 --password "Olares-Test-Password-1!" 2>/dev/null \
        | awk '/^[[:space:]]+id[[:space:]]*:/ {print $NF; exit}' \
        || true)
    if [[ -n "${sync_share_id}" ]]; then
        pass "share public created on sync file (id=${sync_share_id})"
        expect_success "share get ${sync_share_id} (sync target)" \
            "${OLARES_CLI}" files share get "${sync_share_id}"
        expect_success "share rm ${sync_share_id}" \
            "${OLARES_CLI}" files share rm "${sync_share_id}"
    else
        echo "  SKIP: share public on sync file (could not parse share id)"
    fi
}

# test_smb_history exercises the per-node SMB favorites endpoint
# (LarePass's "Favorite Servers" book) end-to-end. Safe to run
# UNCONDITIONALLY — the wire surface is just metadata storage on
# the user's profile (no live SMB server is touched). Every URL
# the test adds is appended to SMB_HISTORY_ADDED_URLS so the
# cleanup trap can drop them on the way out even if an assertion
# mid-stream fails before this function reaches its own rm calls.
#
# What this covers:
#
#   - history list (initial state — just confirms the call works
#     regardless of how many entries are already saved).
#   - history add of a URL-only entry (no -u/-p) — the "favorite
#     stub" flow LarePass uses when the user clicks "Save server"
#     without typing credentials.
#   - history add of a URL + credentials — the "auto-reconnect"
#     flavor that mount autofills from later.
#   - history list --json — pins the JSON-encoded round-trip so
#     a future formatting regression is caught.
#   - history rm of both URLs in a single call (the verb accepts
#     a list).
#   - Negative: add / rm without `//` URL prefix rejected client-
#     side (no wire roundtrip).
#   - Negative: add with -p but no -u rejected client-side (SMB
#     auth needs both halves; password without username is
#     unusable).
test_smb_history() {
    step "files smb history (list / add / rm — unconditional)"

    # 1) list — initial state. We just confirm the call returns 0;
    # the actual content depends on whatever's already on the
    # user's profile, so grepping for an exact line would be
    # brittle across environments.
    expect_success "smb history list (initial state)" \
        "${OLARES_CLI}" files smb history list

    # 2) add — URL only. The URL embeds the run-id so it's unique
    # across concurrent runs and easy to grep for in step 4.
    local url_only_share="//cli-test-fake-host-${TEST_RUN_ID}/share"
    local url_with_creds="//cli-test-fake-host-${TEST_RUN_ID}/with-creds"
    expect_success "smb history add ${url_only_share} (URL-only favorite)" \
        "${OLARES_CLI}" files smb history add "${url_only_share}"
    SMB_HISTORY_ADDED_URLS+=("${url_only_share}")

    # 3) add — URL + credentials. Together these two entries
    # exercise both shapes of the upsert body (no-creds vs.
    # with-creds) so a regression that drops one shape's wire
    # encoding is caught even if the other still works.
    expect_success "smb history add ${url_with_creds} -u test111 -p s3cret (with creds)" \
        "${OLARES_CLI}" files smb history add "${url_with_creds}" \
            -u "test111" -p "s3cret"
    SMB_HISTORY_ADDED_URLS+=("${url_with_creds}")

    # 4) list — confirm both entries landed. We grep for the
    # run-id-stamped URL (not the username) so the assertion
    # doesn't trip on whatever other entries the user might have.
    expect_output_contains "smb history list shows ${url_only_share}" \
        "${url_only_share}" \
        "${OLARES_CLI}" files smb history list
    expect_output_contains "smb history list shows ${url_with_creds}" \
        "${url_with_creds}" \
        "${OLARES_CLI}" files smb history list

    # 5) list --json — pins the JSON envelope so a refactor that
    # drops the encoder (e.g. switching to plain-text-only output
    # accidentally) is caught. We grep for the URL inside the
    # JSON form — same content, different framing.
    expect_output_contains "smb history list --json includes ${url_with_creds}" \
        "${url_with_creds}" \
        "${OLARES_CLI}" files smb history list --json

    # 6) rm — both URLs in one call (the verb accepts a variadic
    # list). After this the two entries should be gone; we
    # confirm by re-listing.
    expect_success "smb history rm <both URLs> (batched)" \
        "${OLARES_CLI}" files smb history rm "${url_only_share}" "${url_with_creds}"
    # The cleanup trap also tries to rm these URLs — that's by
    # design (re-removing a missing entry is a no-op-2xx on the
    # server), but we drop them from the array here so the
    # cleanup banner doesn't lie about doing work it didn't need
    # to do.
    SMB_HISTORY_ADDED_URLS=()

    # Confirm the entries actually went away — same "re-grep the
    # list" approach as the post-rm check in test_repos. This
    # catches the "DELETE returned 200 but the row is still
    # there" failure mode for the history endpoint specifically.
    local post_rm_out
    if post_rm_out=$("${OLARES_CLI}" files smb history list 2>&1); then
        if grep -qF -- "${url_only_share}" <<< "${post_rm_out}"; then
            fail "smb history rm did not remove ${url_only_share} (still in list)"
        else
            pass "smb history list no longer shows ${url_only_share}"
        fi
        if grep -qF -- "${url_with_creds}" <<< "${post_rm_out}"; then
            fail "smb history rm did not remove ${url_with_creds} (still in list)"
        else
            pass "smb history list no longer shows ${url_with_creds}"
        fi
    else
        fail "smb history list after rm failed (output: ${post_rm_out})"
    fi

    # 7) Negatives — exercise the client-side validation gates.
    # These DO NOT touch the wire, so no cleanup is needed.
    expect_failure "smb history add host/share (missing // rejected)" \
        "${OLARES_CLI}" files smb history add "host/share"
    expect_failure "smb history rm host/share (missing // rejected)" \
        "${OLARES_CLI}" files smb history rm "host/share"
    # -p without -u is rejected: SMB auth needs both halves
    # (password without username is unusable for authentication).
    expect_failure "smb history add ${url_only_share} -p secret (no -u rejected)" \
        "${OLARES_CLI}" files smb history add "${url_only_share}" -p "secret"
}

# test_smb_mount runs the mount/unmount round-trip against a REAL
# SMB share. Gated on --smb-url because the live-network call
# can't be smoke-tested without a reachable Samba server, and
# spinning one up just for CI is more friction than the test
# is worth. When --smb-url is set, we go through the full
# mount → ls (find the entry in external/<node>/) → unmount →
# ls (confirm gone) cycle.
#
# Credentials are passed as explicit flags (-u / -p) rather than
# via history autofill — the autofill flow has its own unit-
# test coverage in cli/cmd/ctl/files/smb_test.go, and threading
# it into a smoke test would add an extra dependency on the
# history endpoint (which test_smb_history already covers
# end-to-end). Keeping the mount test flag-driven means an SMB
# server reachability issue surfaces as a single clean failure
# rather than a chain of cascading ones.
test_smb_mount() {
    step "files smb mount / unmount (round-trip against --smb-url)"
    if [[ -z "${OLARES_TEST_SMB_URL}" ]]; then
        echo "  SKIP: no SMB target (pass --smb-url //host/share or set OLARES_TEST_SMB_URL)"
        return 0
    fi

    # Mount. Output looks like:
    #   mount: //host/share @ <node> (user=<user>)
    #     ✓ mounted; the share is now visible at external/<node>/<entry>/
    # The "<entry>" in the second line is a LITERAL placeholder —
    # cli/cmd/ctl/files/smb.go does not interpolate the entry name
    # there. We parse the node from line 1 (which IS interpolated)
    # and compute the entry name separately.
    #
    # Exit-code capture note: we DELIBERATELY don't use the
    # `if ! cmd=$(...); then local rc=$?` idiom here. Bash resets
    # `$?` to 0 inside the `then` branch of a negated condition
    # (the `!` "succeeded", so `$?` reflects that, not the inner
    # command's exit code). The `|| rc=$?` short-circuit captures
    # the real exit code from the cli without that footgun, so the
    # FAIL message reports the actual non-zero status the user
    # needs to diagnose.
    local mount_out=""
    local rc=0
    mount_out=$("${OLARES_CLI}" files smb mount "${OLARES_TEST_SMB_URL}" \
        -u "${OLARES_TEST_SMB_USER}" -p "${OLARES_TEST_SMB_PASSWORD}" \
        --no-history 2>&1) || rc=$?
    if (( rc != 0 )); then
        fail "smb mount ${OLARES_TEST_SMB_URL} (exit ${rc})"
        echo "    output:" >&2
        echo "${mount_out}" | sed 's/^/      /' >&2
        # Heuristic hint for the single most common mis-call:
        # smart-/curly-quote characters in the password (a copy-
        # paste from rich text, or a terminal with autocorrect
        # turning ASCII `'` into U+2018/U+2019). The server's
        # "Incorrect username or password" doesn't suggest this,
        # but it's the source of most "I typed the right
        # password but it still rejects" reports we've seen.
        if [[ "${OLARES_TEST_SMB_PASSWORD}" =~ [‘’“”] ]]; then
            echo "    hint: --smb-password contains a smart/curly quote (‘ ’ “ ”)." >&2
            echo "          Bash treats those as literal password bytes, NOT shell quoting." >&2
            echo "          Re-run with straight quotes: --smb-password 'value'" >&2
        fi
        return 1
    fi
    pass "smb mount ${OLARES_TEST_SMB_URL}"

    # Parse the node from "mount: <url> @ <node> (user=...)".
    # awk -F '@ ' splits on "@ " and grabs everything before
    # the next space in $2.
    local mount_node
    mount_node=$(awk -F '@ ' '/^mount:/ {split($2, a, " "); print a[1]; exit}' <<< "${mount_out}")
    if [[ -z "${mount_node}" ]]; then
        fail "smb mount: could not parse node from progress output"
        echo "    output:" >&2
        echo "${mount_out}" | sed 's/^/      /' >&2
        return 1
    fi
    SMB_MOUNT_NODE="${mount_node}"

    # Compute the entry name. The LarePass / files-backend
    # convention is to use the SMB URL's last segment as the
    # external/<node>/<entry>/ directory name — i.e. mounting
    # //host/Public produces external/<node>/Public/. We verify
    # this convention IS the one in play by listing
    # external/<node>/ and grepping for the basename; if it isn't
    # there, the convention has changed under us and the test
    # tells the user with a clear FAIL rather than silently
    # unmounting the wrong entry.
    local share_basename
    share_basename=$(basename "${OLARES_TEST_SMB_URL}")
    local mount_entry="${share_basename}"

    local after_ls_out
    if ! after_ls_out=$("${OLARES_CLI}" files ls "external/${mount_node}/" 2>&1); then
        fail "ls external/${mount_node}/ after mount failed"
        echo "    output:" >&2
        echo "${after_ls_out}" | sed 's/^/      /' >&2
        return 1
    fi
    # Match a directory row whose NAME column is exactly the
    # share basename (the table prints `<name>/` with a trailing
    # slash for dirs). We anchor on whitespace + name + `/?` +
    # end-of-line to avoid false-positive substring matches
    # (e.g. share `data` matching `userdata/`).
    if ! grep -qE "[[:space:]]${mount_entry}/?\$" <<< "${after_ls_out}"; then
        fail "ls external/${mount_node}/ does not show expected entry '${mount_entry}'"
        echo "    expected entry name: ${mount_entry} (derived from URL basename)" >&2
        echo "    actual listing:" >&2
        echo "${after_ls_out}" | sed 's/^/      /' >&2
        return 1
    fi
    SMB_MOUNTED_ENTRY="${mount_entry}"
    pass "ls external/${mount_node}/ shows ${mount_entry}"
    echo "  resolved node=${mount_node} entry=${mount_entry}"

    # Unmount. Once this succeeds, the bookkeeping vars are
    # cleared so the cleanup trap doesn't try to unmount an
    # already-gone entry (which would print a noisy 404).
    expect_success "smb unmount ${mount_entry} --node ${mount_node}" \
        "${OLARES_CLI}" files smb unmount --node "${mount_node}" "${mount_entry}"
    SMB_MOUNTED_ENTRY=""
    SMB_MOUNT_NODE=""

    # Confirm the entry actually went away. We anchor on
    # whitespace + name + optional trailing slash + end-of-line —
    # same shape as the mount-time grep above — so a substring
    # like "myshare-old" doesn't accidentally match "myshare"
    # still showing up (false positive avoidance).
    local post_unmount_out
    if post_unmount_out=$("${OLARES_CLI}" files ls "external/${mount_node}/" 2>&1); then
        if grep -qE "[[:space:]]${mount_entry}/?\$" <<< "${post_unmount_out}"; then
            fail "smb unmount did not remove ${mount_entry} (still in external/${mount_node}/)"
        else
            pass "ls external/${mount_node}/ no longer shows ${mount_entry}"
        fi
    else
        fail "ls external/${mount_node}/ after unmount failed (output: ${post_unmount_out})"
    fi

    # Negative: unmount with a slash-bearing name is rejected
    # client-side (the wire URL is /api/unmount/external/<node>/<name>/,
    # passing a multi-segment path would silently 404). This
    # pins the CLI's pre-flight gate without any wire roundtrip.
    expect_failure "smb unmount with slash in name rejected client-side" \
        "${OLARES_CLI}" files smb unmount --node "${mount_node}" "bad/name/with/slashes"
}

test_cloud_drives() {
    step "cloud drives (gated on OLARES_TEST_*_PATH env vars)"
    if [[ -z "${OLARES_TEST_AWSS3_PATH}${OLARES_TEST_GOOGLE_PATH}${OLARES_TEST_DROPBOX_PATH}" ]]; then
        echo "  SKIP: no cloud-drive paths configured"
        echo "    set OLARES_TEST_AWSS3_PATH / OLARES_TEST_GOOGLE_PATH / OLARES_TEST_DROPBOX_PATH"
        echo "    e.g. OLARES_TEST_AWSS3_PATH=awss3/<account>/<bucket>"
        return 0
    fi

    # Each cloud drive is a sequenced upload → ls → cat → rm so we
    # confirm both stage 1 (chunked POST) and stage 2 (cloud
    # transfer task polling) work end-to-end.
    local upload_local="${LOCAL_TMP}/aaa.txt"
    local cloud_test_filename="cli-cloud-test-${TEST_RUN_ID}.txt"
    for path_var in OLARES_TEST_AWSS3_PATH OLARES_TEST_GOOGLE_PATH OLARES_TEST_DROPBOX_PATH; do
        local cloud_root="${!path_var}"
        if [[ -z "${cloud_root}" ]]; then
            continue
        fi
        # Strip a trailing '/' so we can append our own path
        # consistently.
        cloud_root="${cloud_root%/}"
        echo "  testing ${path_var}=${cloud_root}"
        expect_success "ls ${cloud_root}/" \
            "${OLARES_CLI}" files ls "${cloud_root}/"
        expect_success "upload aaa.txt → ${cloud_root}/${cloud_test_filename}" \
            "${OLARES_CLI}" files upload "${upload_local}" "${cloud_root}/${cloud_test_filename}"
        expect_output_contains "ls ${cloud_root}/ shows the upload" \
            "${cloud_test_filename}" \
            "${OLARES_CLI}" files ls "${cloud_root}/"
        expect_success "cat ${cloud_root}/${cloud_test_filename}" \
            "${OLARES_CLI}" files cat "${cloud_root}/${cloud_test_filename}"
        expect_success "rm -f ${cloud_root}/${cloud_test_filename}" \
            "${OLARES_CLI}" files rm -f "${cloud_root}/${cloud_test_filename}"
    done
}

# --------------------------------------------------------------------
# Section 3: main
# --------------------------------------------------------------------

main() {
    # Parse + normalize FIRST so the configuration banner reflects
    # the final values (CLI flags can override env vars). --help
    # exits inside parse_args before the EXIT trap matters, so the
    # cleanup banner doesn't print a confusing "test directories
    # preserved" message for a help invocation.
    parse_args "$@"
    normalize_paths

    echo "olares-cli files smoke test"
    echo "  binary:    ${OLARES_CLI}"
    echo "  run-id:    ${TEST_RUN_ID}"
    echo "  local:     ${LOCAL_TMP}"
    echo "  drive:     ${DRIVE_HOME_TEST_DIR}/, ${DRIVE_DATA_TEST_DIR}/"
    if [[ -n "${OLARES_TEST_SYNC_REPO_NAME}" ]]; then
        echo "  sync:      will provision repo (base name: ${OLARES_TEST_SYNC_REPO_NAME})"
    else
        echo "  sync:      SKIP (no --sync-repo-name)"
    fi
    if [[ -n "${CACHE_TEST_DIR}" ]]; then
        echo "  cache:     ${CACHE_TEST_DIR}/"
    else
        echo "  cache:     SKIP (no --cache)"
    fi
    if [[ -n "${EXTERNAL_TEST_DIR}" ]]; then
        echo "  external:  ${EXTERNAL_TEST_DIR}/"
    else
        echo "  external:  SKIP (no --external)"
    fi
    if [[ -n "${OLARES_TEST_AWSS3_PATH}${OLARES_TEST_GOOGLE_PATH}${OLARES_TEST_DROPBOX_PATH}" ]]; then
        echo "  cloud:     awss3=${OLARES_TEST_AWSS3_PATH:-} google=${OLARES_TEST_GOOGLE_PATH:-} dropbox=${OLARES_TEST_DROPBOX_PATH:-}"
    else
        echo "  cloud:     SKIP (no --awss3 / --google / --dropbox)"
    fi
    # smb history sub-tests always run; the mount/unmount round-trip
    # only runs when --smb-url is set, so the banner reads "history-
    # only" vs. "full" to make the planned coverage obvious upfront.
    if [[ -n "${OLARES_TEST_SMB_URL}" ]]; then
        echo "  smb:       history + mount round-trip on ${OLARES_TEST_SMB_URL} (user=${OLARES_TEST_SMB_USER:-(anonymous)})"
    else
        echo "  smb:       history only (pass --smb-url to add mount/unmount round-trip)"
    fi
    if [[ -n "${OLARES_TEST_KEEP}" ]]; then
        echo "  cleanup:   DISABLED (--keep / OLARES_TEST_KEEP set)"
    fi

    preflight
    prepare_local_fixtures

    # Sync repo provisioning runs FIRST (before test_mkdir) so the
    # mkdir suite can include sync as a target namespace. test_mkdir
    # gates on SYNC_REPO_TEST_DIR — if the user didn't ask for sync
    # coverage, this is a cheap no-op that leaves SYNC_REPO_ID empty
    # and every sync-aware test SKIPs cleanly.
    setup_sync_repo

    # Order matters: mkdir creates the per-run parent that every
    # other test parents under, so it has to run before
    # upload/cp/mv/rm. With sync set up above, this also creates
    # the per-run dir inside the sync repo for downstream tests.
    test_mkdir
    test_ls
    test_upload_single_file
    test_upload_single_file_rename
    test_upload_empty_file
    test_upload_directory
    test_upload_dir_to_root
    test_download
    test_cat
    test_cp
    test_mv
    test_rename
    test_rm
    # test_chown / test_aliases need files that earlier upload tests
    # placed under DRIVE_HOME_TEST_DIR (aaa.txt, testU/) and run
    # cheap read-only / metadata operations on them — slot them in
    # right after the core file-verb pass so a regression in any
    # one earlier verb makes them SKIP-by-prerequisite-failure
    # rather than poison the assertion.
    test_chown
    test_aliases
    test_drive_data
    test_share

    # smb history runs UNCONDITIONALLY (favorites endpoint is just
    # metadata storage; no real SMB server required). test_smb_mount
    # is gated on --smb-url inside the function — slot both here so
    # they sit near `share` in the help-text mental model (both are
    # "permissions / external access" surfaces).
    test_smb_history
    test_smb_mount

    # Sync lifecycle + file verbs. test_sync_files re-mkdirs the
    # per-run dir defensively (mkdir -p on an existing dir is a
    # no-op pass) so this block stays self-contained should
    # someone reorder main() in the future.
    test_repos
    
    test_sync_files

    test_cloud_drives

    # Cleanup is handled by the EXIT trap; let it set the exit code
    # based on FAILED.
    if (( FAILED > 0 )); then
        return 1
    fi
    return 0
}

main "$@"
