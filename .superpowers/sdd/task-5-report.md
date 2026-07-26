# Task 5 Report

## Status

DONE_WITH_CONCERNS

Commit: `refactor(preinstall): reuse secure materialization primitives`

## Result

- Bundle publish and HF materialization now share secure static-bundle open, verified regular-file copy, trusted staging creation/open, sealed marker write/read, and single-entry rename validation.
- Bundle retains replace/backup/rollback. HF retains atomic no-replace publish, stage/completion markers, ownership transfer, and interrupted-publish recovery.
- Removed the local `contract-constants.json` mirror and its decoder test. Golden runtime bundle, deployment contract, and typed contract tests remain the drift boundary.

## Security invariants

- Root-relative traversal and symlink components fail closed; regular-file copies also reject hardlinks and special files.
- Copy validates inode identity, type, size limit, exact size, metadata stability, SHA-256, exclusive destination creation, final mode, sync, and close.
- HF marker names are reserved from payload copy. Marker writes are exclusive, synced, and removed on incomplete completion-marker persistence.
- Staging names use cryptographic tokens. Cleanup opens only same-inode directories with expected mode and trusted euid ownership; untrusted HF staging remains untouched.
- Bundle replace rollback and directory fsync windows are unchanged. HF no-replace, marker ordering, parent/tree fsync, and stale/interrupted recovery are unchanged.

## TDD and line delta

- RED: shared copy policy tests failed on missing `verifiedCopy` / `copyVerifiedRegularFile`; GREEN covers bundle/HF modes, size, digest, limit, reserved marker, and Unix hardlink rejection.
- Existing crash, TOCTOU, symlink, forged/replaced staging, no-replace, ownership, and interrupted-publish tests stayed green during refactoring.
- Production: `+358/-362`, net `-4`.
- Tests/fixtures: `+98/-166`, net `-68`.
- Code total: `+456/-528`, net `-72`.

## Verification

- `go test ./pkg/preinstall -count=1`: pass.
- `go test -race ./pkg/preinstall -count=1`: pass; existing macOS linker warning only.
- `go test -vet=off ./pkg/terminus ./pkg/common -count=1`: pass (`common` has no tests).
- `go vet ./pkg/preinstall`: pass.
- `go build ./...`: pass from `cli/`.
- `git diff --check` and IDE diagnostics: pass.

## Concern

Default `go test ./pkg/terminus ./pkg/common` and broad `go vet` remain blocked by pre-existing vet findings in `kube_runtime.go`, `natgateway.go`, and `tasks.go`; none is touched by this task.
