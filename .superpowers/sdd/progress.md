# Clusterop Modules SDD Progress

Plan: `daemon/docs/superpowers/plans/2026-08-10-clusterop-modules.md`
Baseline: related clusterop, handlers, and fanout packages pass; daemon-wide baseline has approved unrelated macOS/BLE/Kubernetes failures.
# Offline preinstall convergence

- Task 5: Olares production `+369/-375` (net `-6`); tests/fixtures `+241/-278` (net `-37`); repository total `+640/-653` (net `-13`).
- Shared secure filesystem primitives now serve bundle and HF materialization while their publish policies remain separate.
- Minimal cross-repository contract constants remain because Market and Olares CI cannot share one generated artifact.
- Task 6: complete; contract decode coverage was consolidated without removing independent filesystem, ownership, crash, rollback, no-replace, or TOCTOU scenarios.

## Clusterop modules

- Task 1: complete (commits d1117ab9f..39a0be8df, review clean).
- Task 2: complete (commits 39a0be8df..efd5db616, review clean).
- Task 3: complete (commit efd5db616..8022cdf92). Runtime and checked mutations only; reboot/shutdown/recovery/HTTP untouched.
- Task 3 fix round 1: complete. Addressed review I1 (UpdateNode callback no longer runs under Manager.mu; optimistic compare-and-replace with ErrConcurrentUpdate), I2 (command_issued stays mutable through its operationActive grace window via new rejectSettled), I3 (Complete/FinishStep persist only safeReason(code), never a module's raw error text), and related Minor fixes (InitNodes empty-slice/deep-copy, SetModuleState JSON pre-validation, Outcome cross-field validation, sentinel comment corrections, real concurrent -race coverage). See `.superpowers/sdd/task-3-report.md` "Fix Round 1" section.
- Task 4: complete (baseline 3e8a72740). reboot/shutdown are self-registering modules; Manager/ParseType/Phase/PowerHost dispatch through the registry; shared power sequencing extracted to power_sequence.go. See `.superpowers/sdd/task-4-report.md`.
- Task 3: review clean at 3e8a72740. Final-review notes: document or remove UpdateNode CAS time-pointer identity dependency; validate custom code shape if needed; register safe reasons for module codes; bound command-issued grace; preserve confirmReboot FinishedAt semantics.
- Task 4: review clean at 659fe7911. Final-review notes: suppress expected persistence-failure settle warning; reject typed-nil OperationStore; bound recovery deadline mutation; clarify recoveryRuntime assertion; log detail when trusted reason is used; make recovery write-failure test setup deterministic.
- Task 5: complete (baseline 659fe7911). Generic recovery (`MarkInterrupted` + `Manager.recoverLoadedOperations`) and panic containment (`safeRun`, `safeRecover`, package-level `ExecuteNode`) are module-driven, not reboot-specific; all six Task 4 final-review notes above are fixed. See `.superpowers/sdd/task-5-report.md`.
- Task 5 fix round: review clean at ab7b4ca7b. Fixed I1 (recovery's `ErrRecoveryCannotExtendDeadline` guard now shared by `SetCommandIssuedUntil`/`Complete`/`Settle` via `checkCommandIssuedUntilTransition`, including the zero-value-deadline minor) and I2 (`safeRun`'s `recoverFromRunPanic` no longer overwrites a `command_issued`/terminal outcome a module already committed before panicking; a still-running panic now closes any open step/node alongside the `failed/module_failed` settlement). Also fixed: `ExecuteNode` nil-registry panic, the busy-poll in `TestRecoverPanicDuringResumeLeavesTheRecordIntact` (replaced with a `Deps.recoveryDone` test seam), and a double-`Now()`-call timestamp inconsistency in `MarkInterrupted`/`recoverFromRunPanic`'s shared settlement path (`Manager.updateAt`/`applyLockedAt`). Recover hang timeout deferred, as instructed. See `.superpowers/sdd/task-5-report.md` "Review fix round" section.
