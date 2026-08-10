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
only, not pushed, per this task's authorization.
Hash: `ab7b4ca7b13ed5f93c5a5a73a4828e7d19d71b3d`.

---

# Review fix round — I1 (recovery deadline guard), I2 (Run panic overwrite), and five minors

Status: **COMPLETE**
Base: `ab7b4ca7b13ed5f93c5a5a73a4828e7d19d71b3d` (the commit above)
Commit: see "Commit" at the end of this section.

## What changed, in one paragraph

I1 and I2 were both the same shape of bug: a check that existed on one
call path but not on the others that reach the same state. I1's
`ErrRecoveryCannotExtendDeadline` guard lived only in
`SetCommandIssuedUntil`; `Complete` and `Settle` — the two paths a
`RecoverableModule` actually uses to move a `command_issued` record forward
— wrote `CommandIssuedUntil` with no check at all. I2's `safeRun` treated
every panic identically, settling `failed/module_failed` even when the
module had already committed a real `command_issued` handoff or terminal
outcome through its own `Runtime` moments before it panicked. Both are
fixed by moving the decision to one place each path calls through
(`checkCommandIssuedUntilTransition`, `recoverFromRunPanic`) rather than by
patching each call site separately. Five minors from the same review round
are fixed alongside them: the zero-value bug that `checkCommandIssuedUntilTransition`
would otherwise have inherited from the original `SetCommandIssuedUntil`
check, a nil-registry panic in `ExecuteNode`, a busy-poll in
`TestRecoverPanicDuringResumeLeavesTheRecordIntact` replaced with a
deterministic signal, and a double-`Now()`-call bug in the settlement
primitive `MarkInterrupted` (and `recoverFromRunPanic`'s own settlement)
share, which could stamp `UpdatedAt` a moment apart from `FinishedAt` and
every step/node `FinishedAt` it closes.

## TDD cycles

Each cycle was RED first: the four touched production files
(`manager.go`, `recovery.go`, `runtime.go`, `settlement.go`) were reverted
to the pre-fix commit above, the new tests were added, and each was run to
fail for the reason described — then the fixed files were restored and the
same tests re-run to GREEN. Two failure modes showed up: a genuine
assertion failure (the bug), and, for `ExecuteNode`'s nil-registry case, an
actual unrecovered panic that crashed the test binary — both are RED in
the sense the fix removes them.

### I1 — recovery deadline guard was one call site, not one rule

RED:

```
--- FAIL: TestRecoveryCannotExtendAnExpiredDeadlineViaComplete
    runtime_test.go:720: Complete() = <nil>, want ErrRecoveryCannotExtendDeadline
--- FAIL: TestRecoveryCannotExtendAnExpiredDeadlineViaSettle
    runtime_test.go:736: Settle() = <nil>, want ErrRecoveryCannotExtendDeadline
--- FAIL: TestRecoveryMaySetAFirstDeadlineFromZeroViaSetCommandIssuedUntil
    runtime_test.go:759: SetCommandIssuedUntil() = recovery cannot extend an
    expired command deadline, want nil for a first-time deadline from zero
```

The third failure is the zero-value minor: the original check compared
`!op.CommandIssuedUntil.After(now)` with no `IsZero()` guard, so a record
that had *never* held a deadline (the zero value, always "not after now")
was treated exactly like one whose deadline had expired, and recovery could
not give a `running` record its own first `command_issued` grace period.

GREEN — `runtime.go` adds `checkCommandIssuedUntilTransition(current, next
time.Time) error` on `*operationRuntime`: a no-op for an ordinary run, a
no-op for clearing (`next.IsZero()`) or for extending a still-active
deadline, and `ErrRecoveryCannotExtendDeadline` only when `next` names a
future time and `current` is *both* non-zero *and* already expired.
`SetCommandIssuedUntil`, `Complete` (`runtime.go`), and `Settle`
(`settlement.go`) all call through it — `Complete`/`Settle` build a
composite `validate` that runs `rt.settled` first and this check second,
only when `rt.recovery` is true, so an ordinary run's `Complete`/`Settle`
take no new code path at all.

Tests: `TestRecoveryCannotExtendAnExpiredDeadlineViaComplete`,
`TestRecoveryCannotExtendAnExpiredDeadlineViaSettle`,
`TestRecoveryMaySetAFirstDeadlineFromZeroViaSetCommandIssuedUntil`,
`TestRecoveryMaySettleCommandIssuedFromRunningWithAFreshDeadline`,
`TestRecoveryMaySettleWithZeroDeadlineEvenWhenExpired` (the last is the
direct `Settle`-level regression guard for late reboot confirmation — zero
deadline, already-expired record, must still succeed). The existing
`TestARebootIsConfirmedAfterItsGraceDeadlineHasLongPassed` (end-to-end
through `rebootModule`/`confirmReboot`) and
`TestRecoveryCannotExtendAnExpiredDeadline` /
`TestRecoveryCanClearAnExpiredDeadline` /
`TestRunRuntimeMaySetAFutureDeadlineRegardless` (the original
`SetCommandIssuedUntil`-only tests from the first round) all still pass
unmodified.

### I2 — a panicking Run could overwrite state the module already committed

RED:

```
--- FAIL: TestRunPanicAfterCommandIssuedLeavesItIntact
    panic_test.go:123: status = "failed", want command_issued left exactly
    as the module committed it
--- FAIL: TestRunPanicAfterAStartedStepClosesItAndSettlesFailed
    panic_test.go:185: Steps[0] = [{Name:work Status:running
    FinishedAt:<nil> ...}], want the started step closed failed rather than
    left running
```

(`TestRunPanicAfterSucceededLeavesItIntact` did not fail against the old
code — a `Succeeded` record is already `Terminal()`/non-`command_issued`,
so the old code's *unconditional* `rejectSettled` already refused the
follow-up `Complete`, just noisily, with a `klog.Warningf`. `command_issued`
is the one status `rejectSettled` treats as still "active" while its
deadline holds, which is exactly why only that case corrupted the record.)

GREEN — `recovery.go`'s `safeRun` now calls a new `recoverFromRunPanic(rt
Runtime) Outcome` from its `recover()` instead of building
`failedWith(CodeModuleFailed, ...)` directly. It re-reads the operation
through `rt`: if the status is no longer `running` — the module already
called `Complete` or `Settle` before it panicked — it returns
`Outcome{}.alreadyRecorded()` so `Manager.settle` leaves the record exactly
as the module left it. If the status is still `running`, it settles
`failed`/`module_failed` itself, through the same `settleTerminated` helper
`MarkInterrupted` uses, which also closes any step or node the module
started but never finished — so a caller never sees a `failed` operation
with a step still reading `running`. A `Runtime` this package did not build
(`managerOf(rt)` fails) falls back to the old unconditional behavior, since
there is no committed state to check.

Tests: `TestRunPanicAfterCommandIssuedLeavesItIntact`,
`TestRunPanicAfterSucceededLeavesItIntact`,
`TestRunPanicAfterAStartedStepClosesItAndSettlesFailed`. The existing
`TestRunPanicSettlesTheOperationFailed` and
`TestRunPanicReleasesTheOperationLock` (ordinary panic, nothing committed
first: still settles `failed`/`module_failed`, still releases the lock,
still never leaks the panic's own text) pass unmodified.

### Minors

1. **Zero-value semantics** — folded into I1 above
   (`checkCommandIssuedUntilTransition`'s `current.IsZero()` branch); no
   separate fix, one shared RED/GREEN.
2. **`TestRecoverPanicDuringResumeLeavesTheRecordIntact`'s busy-poll.** The
   test closed a channel the instant *before* a panic and then polled
   `m.Get` on a 1ms timer for up to a second, hoping to observe the
   deferred `recover()` having already run — correct by construction but
   timing-based. `Deps` gains an unexported `recoveryDone chan string`
   (production never sets it; zero cost when nil) and `Manager.resume`
   sends the operation's id on it after `safeRecover` returns, panic or
   not. The test sets `deps.recoveryDone` before calling
   `NewManagerWithRegistry` and blocks on a channel receive instead of a
   timer — it now fails deterministically (channel timeout) rather than
   flaking if the recovery goroutine is ever slow, and succeeds the instant
   the goroutine actually finishes rather than up to 1ms late. Not
   independently RED/GREEN-able as a logic bug (the old poll already
   reached the correct answer, just non-deterministically); verified by
   running it, along with the whole package, under `-race -count=1`.
3. **`ExecuteNode` nil-registry panic.** RED (see above): an actual
   unrecovered `nil` pointer dereference inside `(*ModuleRegistry).Lookup`,
   thrown *before* `ExecuteNode`'s own `defer`/`recover` was even
   registered, crashing the test binary. Fixed with an explicit `registry
   == nil` check at the top of `ExecuteNode`, returning the same
   `unsupportedNodeOperationError()` an unknown type already gets.
   Test: `TestExecuteNodeRefusesANilRegistry`.
4. **`MarkInterrupted` timestamp consistency.** RED:

   ```
   --- FAIL: TestMarkInterruptedStampsEveryTimestampFromTheSameMoment
       recovery_test.go:365: UpdatedAt = ...20.002 UTC, FinishedAt =
       ...20.001 UTC, want the same settlement moment
   ```

   `MarkInterrupted`/`settleUnknownModule` (via `settleTerminated`) already
   used one `now` consistently for `FinishedAt` and every step/node they
   close — the second, disagreeing clock read came from `applyLocked`,
   which unconditionally called `m.deps.Now()` again for `UpdatedAt` after
   `fn(op)` ran. Fixed by threading the caller's own `now` through:
   `manager.go` adds `updateAt(id, at, fn)` / `applyLockedAt(op, at, fn)`
   (parallel to `update`/`applyLocked`, which now delegate to them with
   `m.deps.Now()`), and the three call sites in `recoverLoadedOperations`
   plus the new `recoverFromRunPanic` compute `now` once and pass it both
   to the settlement helper and to `updateAt`. `update`/`applyLocked`'s own
   single-`Now()`-call behavior for every other caller (`StartStep`,
   `FinishStep`, `Complete`, etc.) is unchanged — this fix is scoped to the
   three "framework forces a terminal settlement with one `now`" call
   sites, not to every checked mutation.
   Test: `TestMarkInterruptedStampsEveryTimestampFromTheSameMoment`
   (asserts `UpdatedAt`, `FinishedAt`, `Steps[0].FinishedAt`, and
   `Nodes[0].FinishedAt` are all `.Equal` to each other).
5. **Recover hang timeout** — explicitly out of scope for this round, per
   instruction; still listed under Concerns below.

## Files

Changed

- `pkg/cluster/clusterop/runtime.go` (`checkCommandIssuedUntilTransition`;
  `Complete` validates through it when `rt.recovery`)
- `pkg/cluster/clusterop/settlement.go` (`Settle` validates through
  `checkCommandIssuedUntilTransition` when `rt.recovery`)
- `pkg/cluster/clusterop/recovery.go` (`recoverFromRunPanic`; `safeRun`
  calls it; `ExecuteNode` nil-registry guard; `recoverLoadedOperations`'s
  three settlement sites now compute `now` once and call `updateAt`)
- `pkg/cluster/clusterop/manager.go` (`Deps.recoveryDone` test seam;
  `resume` signals it; `updateAt`/`applyLockedAt` added, `update`/`applyLocked`
  delegate to them)
- `pkg/cluster/clusterop/runtime_test.go`, `pkg/cluster/clusterop/panic_test.go`,
  `pkg/cluster/clusterop/recovery_test.go` (new tests; the busy-poll fix in
  `panic_test.go`)

## Tests

- `go build ./pkg/cluster/clusterop/...` and
  `go vet ./pkg/cluster/clusterop/...` — clean.
- `gofmt -l ./pkg/cluster/clusterop/` — clean (no output).
- Focused run of the 11 new/rewritten tests (`-v -count=1`) — all PASS,
  confirmed RED beforehand against the reverted production files (see
  transcript above for the exact RED output).
- `go test ./pkg/cluster/clusterop/... -count=1 -v` — 239 `--- PASS`, 0
  `--- FAIL`.
- `go test ./pkg/cluster/clusterop/... -race -count=1` — ok (~16.6s), no
  data races.

## Self-review

- `checkCommandIssuedUntilTransition` is a method on `*operationRuntime`
  (not a free function) so it reads `rt.recovery` and `rt.m.deps.Now()`
  without either caller needing to pass them in; `SetCommandIssuedUntil`,
  `Complete`, and `Settle` each still run their own `rt.settled(op)` check
  first — the new check only ever narrows what recovery may additionally
  do, never widens what an ordinary run may do.
- `recoverFromRunPanic` reads the operation through `m.Get(id)` (a fresh
  lock/copy) rather than through `rt.Operation()`, then commits through
  `m.updateAt` directly rather than `rt.Complete` — this is intentional:
  the settlement it performs (closing steps/nodes) is not expressible
  through `Complete`'s `Outcome`, and it mirrors exactly how
  `MarkInterrupted`/`settleUnknownModule` already commit a forced
  settlement in this package (via `m.update`, now `m.updateAt`), not a new
  pattern.
- Checked whether `recoverFromRunPanic`'s two-step read-then-write
  (`m.Get` then `m.updateAt`) is racy: no other goroutine mutates this
  operation while its own `Run` goroutine is inside `safeRun`'s deferred
  `recover()` — the same single-writer assumption `Manager.settle` (the
  only caller of `safeRun`) already depends on.
- `updateAt`/`applyLockedAt` are additive: `update`/`applyLocked` keep
  their old signatures and behavior for every existing caller
  (`checkedUpdate` still calls `applyLocked`, i.e. `m.deps.Now()` derived
  internally, unchanged for `StartStep`/`FinishStep`/`SetHostBootID`/etc.).
  Only `recoverLoadedOperations` and `recoverFromRunPanic` — both of which
  already had their own external `now` before this round, just not
  threaded all the way through — were switched to the `...At` variants.
- `deps.recoveryDone` is unexported on an exported struct, so it is
  settable only from inside this package (tests), invisible and zero-cost
  (nil channel, `if done != nil` guard) for every production caller and
  every existing test that does not set it.
- Re-ran the full pre-existing suite (not just the new tests) under `-race`
  to make sure `checkCommandIssuedUntilTransition`'s extra `rt.m.deps.Now()`
  call and the `updateAt` refactor did not change behavior for any test
  from the first round — all 239 tests pass.
- Hardcoding scan: no new `reboot`/`shutdown` literal in any of the four
  changed production files.

## Concerns / follow-ups

1. **Recover hang timeout** (noted in the review, explicitly deferred):
   `rebootModule.Recover`'s poll loop is bounded by `Timeouts.Ready`, but
   nothing bounds a third-party `RecoverableModule.Recover` that never
   returns — `recoverLoadedOperations` would block `NewManagerWithRegistry`
   forever on such a module for a non-`command_issued` record (see Concern
   2 from the first round). Not addressed here, per instruction.
2. `recoverFromRunPanic`'s "was anything already committed" check is
   `current.Status != StatusRunning`. This is exactly right for every
   status this package defines (`command_issued` and every terminal status
   are all "not running"), but it means a hypothetical future `Status`
   value that is neither `running` nor one of today's committed/terminal
   meanings would also be treated as "already committed, leave it alone" —
   the same implicit assumption `operationActive`/`Status.Terminal()`
   already make elsewhere in this package, not a new one.
3. The RED demonstrations for this round were done by reverting the four
   touched production files to commit `ab7b4ca7b13ed5f93c5a5a73a4828e7d19d71b3d`
   and adding a *minimal* stand-alone `recoveryDone` stub to that reverted
   `manager.go` so the new tests could compile at all — that stub is not
   itself part of the graded RED/GREEN cycle for I1/I2, only scaffolding so
   `TestRecoverPanicDuringResumeLeavesTheRecordIntact`'s rewrite could run
   in the same batch as the other new tests.

## Commit

`fix(clusterop): close recovery deadline and Run-panic overwrite gaps from
Task 5 review` — local commit only, not pushed, per this task's
authorization.
Hash: `af5d6ab216bf4fa693b6432291b5cd5ffd8fd2b9`.
