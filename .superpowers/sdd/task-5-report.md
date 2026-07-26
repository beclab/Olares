# Task 5 Report

## Status
DONE_WITH_CONCERNS

Commits: `refactor(preinstall): reuse secure materialization primitives`; `fix(preinstall): tighten shared filesystem policies`.

## Result and invariants
- Bundle and HF share static-bundle open, verified regular-file copy, trusted staging, marker I/O, and entry validation while retaining their separate publish protocols.
- Copies reject traversal, symlink, hardlink, special files, size/digest changes, and non-exclusive output; mode and fsync guarantees remain.
- Staging trust requires a known matching euid, accepted mode, and stable inode. Bundle cleanup accepts active `0700`, legacy `0755`, and sealed `0555` roots; every opened root is closed before removal or return.
- HF alone rejects its reserved stage/completion marker paths through copy policy. The generic copy primitive and bundle path do not know HF filenames.
- Bundle rollback/fsync and HF no-replace, marker ordering, ownership, and interrupted-publish recovery remain unchanged.
- The minimal `contract-constants.json` mirror is retained because Market and Olares CI cannot consume one generated artifact across repositories. It covers schema/source/file names, JSON/app/chart/manifest/entry/artifact limits, artifact kind, JSON tag, and scopes; golden bundles do not cover numeric drift.

## TDD and line delta
- RED covered sealed `0555` cleanup, unknown-owner fail-closed policy, and bundle/HF reserved-marker contrast; shared copy tests cover modes, size, digest, limits, TOCTOU, and Unix hardlinks.
- Production: `+369/-375`, net `-6`; tests/fixtures: `+241/-278`, net `-37`; repository total: `+640/-653`, net `-13`.

## Verification
- `go test ./pkg/preinstall -count=1` and `go test -race ./pkg/preinstall -count=1`: pass.
- `go test -vet=off ./pkg/terminus ./pkg/common -count=1`, `go vet ./pkg/preinstall`, `go build ./...`, `git diff --check`, and IDE diagnostics: pass.

## Concern
Default terminus/common tests and broad vet remain blocked by pre-existing findings in `kube_runtime.go`, `natgateway.go`, and `tasks.go`.
