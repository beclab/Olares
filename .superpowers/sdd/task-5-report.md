# Task 5 — generic recovery, unknown-module handling, panic containment

Status: **COMPLETE**
Baseline: `659fe791188a4d98600e36e2c18efd405a1ed153`
Commit: see "Commit" below
Package: `daemon/pkg/cluster/clusterop`

## What changed, in one paragraph

Recovery at daemon startup is now decided the same way for every module a
registry holds, not by a manager method that only knew about reboots.
`recovery.go` adds `MarkInterrupted` (extracted from the old
`Manager.markInterrupted`), `Manager.recoverLoadedOperations` (the load-time
dispatch: unknown type → settle `unsupported_operation`; no
`RecoverableModule` and non-terminal → `MarkInterrupted`; a
`RecoverableModule` → call `Recover` first and only fall back to
`MarkInterrupted` if that leaves the record still moving), and panic
boundaries — `safeRun`, `safeRecover`, and a package-level `ExecuteNode` — so
a module panicking in `Run`, `Recover`, or `ExecuteNode` ends that one
operation or that one request, logs a stack to `klog`, and never reaches the
persisted record or an HTTP caller. Six Task 4 review minors that are
specifically about recovery and settlement are fixed alongside it.

## TDD cycles

Each cycle was RED first: the production file was written to a scratch path
outside the tree, the tests were written and run to fail (compile failure,
since the symbols did not exist), then the production file was restored and
the tests re-run to GREEN.

### Cycle 1 — unknown historical modules and module-driven recovery dispatch

RED (`go vet` — the tests could not compile without `MarkInterrupted`,
`ErrRecoveryCannotExtendDeadline`, `ExecuteNode`, `safeRun`, `safeRecover`):

```
pkg/cluster/clusterop/panic_test.go:212:12: undefined: ExecuteNode
```

GREEN — `recovery.go`; `manager.go`'s loading loop restructured to load every
record's bookkeeping first, then call `recoverLoadedOperations` once, then
compute `activeID` from the (possibly now-settled) final state, then start
async resume for outstanding `command_issued` records exactly as before.

Tests: `TestUnknownHistoricalModuleSettlesFailed`,
`TestUnknownHistoricalModuleRetainsTerminalRecord` (succeeded / failed /
partially_failed / command_issued, all four unaffected by a missing module),
`TestManagerCallsModuleRecovery`,
`TestManagerFallsBackToMarkInterruptedWhenRecoveryLeavesItRunning`,
`TestManagerRecoveryFallbackDoesNotOverwriteWhatTheModuleSettled`.

Existing tests that exercise the same dispatch through the built-in modules —
`TestAnOperationInterruptedByARestartIsReportedFailed` (a `running` reboot,
now routed through `rebootModule.Recover` synchronously before falling back
to `MarkInterrupted`, because `rebootModule` implements `RecoverableModule`),
`TestARecoverableModuleIsHandedItsUnfinishedCommand`,
`TestAModuleWithoutRecoveryLeavesItsCommandAlone`,
`TestARebootIsConfirmedAfterItsGraceDeadlineHasLongPassed` — all still pass
unmodified.

### Cycle 2 — panic containment for Run, Recover, ExecuteNode

RED — same compile failure as above (`safeRun`/`safeRecover`/`ExecuteNode`
did not exist).

GREEN — `orchestrate.go`'s `run` calls `safeRun(module, ctx, rt, req)`
instead of `module.Run` directly; `Manager.resume` and
`recoverLoadedOperations` both call `safeRecover`, which wraps a module's
`Recover` in the same `defer`/`recover` shape used elsewhere in this package.
`ExecuteNode(ctx, *ModuleRegistry, NodeRequest) error` is a new
package-level function: registry lookup, `NodeOperationModule` assertion,
then a panic boundary around the module's own `ExecuteNode`. Both refusals
(unknown type, no node capability) and the panic fallback return the same
`*PowerError` shape `PowerHost` already uses, so a caller mapping one error
type maps both.

Tests: `TestRunPanicSettlesTheOperationFailed`,
`TestRunPanicReleasesTheOperationLock`,
`TestRecoverPanicDuringLoadFallsBackToMarkInterrupted`,
`TestRecoverPanicDuringResumeLeavesTheRecordIntact`,
`TestExecuteNodeDispatchesToTheRegisteredModule`,
`TestExecuteNodeRefusesAnUnknownType`,
`TestExecuteNodeRefusesAModuleWithNoNodeCapability`,
`TestExecuteNodeReturnsTheModulesOwnError`,
`TestExecuteNodePanicSettlesModuleFailed`.

Every panic test asserts the panic's own text (`"boom…"`) never reaches the
persisted `Error` / returned error string, only the stable code
(`module_failed`) and the reviewed sentence do; the two `Run`-panic tests
also assert the operation reaches `failed` and the cluster lock is free
again afterward (a second `Create` succeeds and is itself awaited to
completion so its goroutine cannot race the test's own `t.TempDir()`
cleanup).

### Cycle 3 — the six Task 4 review minors this task owns

Each was RED via a targeted test first (see below), then fixed in place. No
production behavior outside its own fix changed.

1. **Typed-nil `OperationStore` refused.** `Deps.validate()` compared
   `d.Store == nil`, which is false for an interface holding a nil `*Store`.
   `storeIsNil` (the same `reflect.Ptr`/`IsNil` check `ModuleRegistry.Register`
   already uses for a module) replaces the plain comparison.
   RED: `TestNewManagerRefusesTypedNilStore` — `NewManager` accepted the
   typed-nil store before the fix.
2. **`settle` no longer warns a second time about a persistence failure
   `applyLocked` already logged at error level.** `settlementFailedToPersist`
   distinguishes "this Complete call's own write just failed"
   (`errStatePersistenceFailed`) and "a *previous* call already forced this
   record into `persistFailed`" (`ErrOperationTerminal` while
   `m.persistFailed[id]` is true) from an ordinary "already terminal"
   refusal, which still warns. Not independently testable: this package has
   no klog-capture convention (checked — none of the existing test files
   assert on log output), so this is verified by reading the code path and
   by every persistence-failure test (`TestAConfirmationThatCannotBePersistedConfirmsNothing`,
   `TestAnUnusableOutcomeStillSettlesTheOperation`) still passing with the
   guard in place.
3. **Recovery cannot re-arm an expired `CommandIssuedUntil`.**
   `SetCommandIssuedUntil` on a recovery runtime now refuses a future
   deadline for a record whose own deadline has already passed
   (`ErrRecoveryCannotExtendDeadline`); clearing it is still always allowed,
   and an ordinary run's runtime is unaffected.
   RED: `TestRecoveryCannotExtendAnExpiredDeadline` — the call succeeded and
   silently re-armed the lock before the fix.
   GREEN also covered by `TestRecoveryCanClearAnExpiredDeadline` and
   `TestRunRuntimeMaySetAFutureDeadlineRegardless` (no regression for a run).
4. **`recoveryRuntime` assertion in `rebootModule.Recover` documented.** The
   comment now says the assertion always succeeds in production (every
   caller builds the runtime with `newRecoveryRuntime`) and exists to refuse
   a foreign `Runtime` rather than improvise. Documentation only; no
   behavior change, so no new test.
5. **Detail logged when a trusted `reason` and a module's `Error` coexist.**
   `persistedReason` now logs (at warning level) when both are set, while
   still persisting only the trusted `reason` — never blending or
   overriding it. RED/GREEN: `TestPersistedReasonKeepsTheTrustedReasonEvenWithDetailPresent`
   asserts the persisted text is unaffected; the log line itself is not
   asserted, for the same klog-testability reason as item 2.
6. **Deterministic recovery write-failure test setup.**
   `TestAConfirmationThatCannotBePersistedConfirmsNothing` called
   `store.refuseWrites()` *after* `NewManager`, racing the reboot-confirmation
   goroutine that `NewManager` starts. It now calls `refuseWrites()` before
   `NewManager`, so the very first write attempt — regardless of the
   goroutine's timing — fails. Test still passes; the flake window is gone.

## Migration map

| Was (baseline `659fe7911`) | Is now |
| --- | --- |
| `Manager.markInterrupted` (method) | `MarkInterrupted(*Operation, time.Time)` (package function), same status/code/step/node/timestamp transitions |
| `NewManagerWithRegistry`'s inline `switch` over `!Terminal()` / `== StatusCommandIssued` | `Manager.recoverLoadedOperations()`, dispatching through the registry per operation |
| (nothing — an unknown type could not previously appear at load, since `ParseType`/`Lookup` gated `Create`) | `settleUnknownModule` — a historical record of a type no longer registered settles `failed`/`unsupported_operation` instead of being left exactly as stored |
| `resume`'s bare `go recoverable.Recover(...)` | `go m.safeRecover(...)` — same call, now panic-contained |
| `orchestrate.go` `run`'s bare `module.Run(...)` | `safeRun(module, ...)` — same call, now panic-contained |
| (nothing — `NodeOperationModule.ExecuteNode` existed but nothing routed to it) | package-level `ExecuteNode(ctx, *ModuleRegistry, NodeRequest) error`, for task 6's two HTTP handlers to share |
| `Deps.validate()`'s `d.Store == nil` | `storeIsNil(d.Store)` |
| `settle`'s unconditional `klog.Warningf` on any `Complete` failure | conditional on `!settlementFailedToPersist(id, err)` |
| `SetCommandIssuedUntil`'s unconditional write | validated against `ErrRecoveryCannotExtendDeadline` when `rt.recovery` |
| `persistedReason`'s silent `return o.reason` | logs `o.Error` first when both are set, still returns `o.reason` unchanged |

Preserved unchanged: the `daemon_restarted` status/code/step/node/timestamp
shape `MarkInterrupted` produces; `rejectSettledDuringRecovery`'s existing
grace-window semantics for a run vs. a recovery; `rebootModule`'s and
`shutdownModule`'s own `Run`/`Recover`/`ExecuteNode` bodies (only their doc
comments changed); every stable code and reviewed sentence; the one-write
atomic confirm (`settlement.go`, untouched).

## Files

New

- `pkg/cluster/clusterop/recovery.go`
- `pkg/cluster/clusterop/panic_test.go`

Changed

- `pkg/cluster/clusterop/manager.go` (loading loop, typed-nil `Store` check, `markInterrupted` removed, `resume` panic-contained)
- `pkg/cluster/clusterop/orchestrate.go` (`safeRun`, `settlementFailedToPersist`)
- `pkg/cluster/clusterop/runtime.go` (`ErrRecoveryCannotExtendDeadline`, `operationRuntime.recovery`, `SetCommandIssuedUntil` validation)
- `pkg/cluster/clusterop/reason.go` (`persistedReason` logs detail alongside a trusted reason)
- `pkg/cluster/clusterop/module_reboot.go`, `pkg/cluster/clusterop/module_shutdown.go` (doc comments only)
- `pkg/cluster/clusterop/recovery_test.go`, `pkg/cluster/clusterop/runtime_test.go`, `pkg/cluster/clusterop/reason_test.go`, `pkg/cluster/clusterop/manager_test.go`

## Tests

- `go build ./pkg/cluster/clusterop/...` — clean.
- `go vet ./pkg/cluster/clusterop/...` and `go vet ./pkg/cluster/... ./internel/apiserver/...` — clean.
- `gofmt -l ./pkg/cluster/clusterop/` — clean.
- Focused run (every new/changed test, 33 tests, `-v -count=1`) — all PASS;
  panic tests show the expected single `klog.Errorf` stack-trace line per
  panic and nothing else.
- `go test ./pkg/cluster/clusterop/... -count=1` — ok (13.8s), 232 total test
  functions in the package.
- `go test ./pkg/cluster/clusterop/... -race -count=1` — ok (16.2s), no data
  races.
- `go test ./pkg/cluster/clusterop/... -race -count=3` — ok (35.9s), run
  three times to rule out the flake noted during planning
  (`TestCreatePersistsParamsDigestWithoutRawParams` / `t.TempDir` cleanup);
  did not reproduce in this round.
- `go test ./pkg/cluster/... ./internel/apiserver/... -count=1` — all ok
  except `pkg/cluster/state` `TestCurrentState`
  (`route ip+net: operation not permitted`), which fails the same way at
  baseline (network interface enumeration, forbidden in this sandbox) —
  unrelated to this change.
- `go build ./...` — fails only in `internel/watcher/system`
  (`utils.ListenNetworkCarrierChanges` undefined on darwin), pre-existing and
  platform-specific, same as at baseline.

## Self-review

- Every panic boundary is narrow: `safeRun` wraps only `module.Run(...)`,
  `safeRecover` wraps only `recoverable.Recover(...)`, `ExecuteNode`'s
  `defer` wraps only `nodeModule.ExecuteNode(...)`. No manager bookkeeping
  runs inside any of the three `defer`/`recover` blocks, so a bug in
  `applyLocked` or `checkedUpdate` is never silently swallowed as if it were
  a module panic.
- A panicking `Run` still settles through the same `Manager.settle` path an
  ordinary unusable outcome does (`failedWith(CodeModuleFailed, ...)`), so
  there is exactly one framework-owned failure mode for "the module
  stopped without saying what it did," not two.
- A panicking `Recover` leaves the record exactly as the module's own
  `Runtime` calls last committed it — `safeRecover` does not force any
  status — so a module that panics after confirming a step but before
  completing the operation does not lose that step. Verified directly by
  `TestRecoverPanicDuringResumeLeavesTheRecordIntact`.
- The synchronous recovery path (`recoverLoadedOperations`) still applies
  its own `MarkInterrupted` fallback after a panicking `Recover`, because it
  checks the operation's status after `safeRecover` returns regardless of
  whether that return was normal or via a recovered panic — there is no
  special case for "the module panicked" in that check.
- `recoverLoadedOperations` reads and mutates `m.ops`/`m.order` before any
  goroutine exists (construction is still single-threaded until `resume`
  starts one), so no new lock is needed for it; `m.update` inside it still
  takes `m.mu` like every other caller.
- `ExecuteNode`'s two refusal paths (unknown type, no node capability) and
  its panic fallback all return the same `*PowerError` shape `power.go`
  already uses, so a task-6 handler written against `PowerHost`'s error
  shape needs no second error type.
- Hardcoding scan: `recovery.go` holds no `reboot`/`shutdown` literal, and
  neither does the code added to `manager.go`, `orchestrate.go`, `runtime.go`,
  or `reason.go` — the only occurrences left in the package are the existing
  ones in `module_reboot.go`/`module_shutdown.go`/`power_sequence.go`
  documented as unchanged by Task 4.

## Concerns / follow-ups

1. `settlementFailedToPersist`'s and `persistedReason`'s log-suppression /
   log-addition behavior is not directly asserted by a test, because this
   package has no existing convention for capturing `klog` output. Both are
   exercised indirectly (existing persistence-failure tests still pass with
   the new code path; `TestPersistedReasonKeepsTheTrustedReasonEvenWithDetailPresent`
   asserts the persisted text is unaffected) but a reviewer who wants log-line
   assertions will need to add a capture seam first — out of scope here.
2. `recoverLoadedOperations` still calls a `RecoverableModule`'s `Recover`
   synchronously, in the constructor, for any non-terminal record (this was
   already true for `command_issued` records via the old `resume`, and is
   now also true for `running`/`pending` ones). Every module built into this
   package returns immediately for anything but a `command_issued` record it
   can prove happened, so this does not block `NewManagerWithRegistry` today
   — but a third-party module whose `Recover` does real work for a `running`
   record would now do that work on the daemon's startup path. This was
   flagged as a design tension in the original planning discussion and
   accepted as intentional for this task; worth revisiting if a
   `RecoverableModule` more expensive than `rebootModule`'s ever lands.
3. `ExecuteNode`'s panic fallback and its "unknown type / no node capability"
   refusal share one stable code family (`unsupported_operation`,
   `module_failed`) with the rest of this package, but task 6 has not yet
   wired an HTTP status to either — that mapping is out of scope here.
4. `ModuleRegistry.Freeze` (task 7) remains untouched, as instructed.

## Commit

`refactor(clusterop): delegate recovery to operation modules` — local commit
only, not pushed, per this task's authorization. See `git log -1` in the
working tree for the hash.
