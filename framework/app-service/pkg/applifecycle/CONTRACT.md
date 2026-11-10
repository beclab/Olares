# Handler Contract

The following 5 rules are **hard checklist items** for every PR that adds or
modifies a `HandlerFunc` inside `pkg/applifecycle/handlers/`. Violation of any
rule blocks the merge. Reviewers should link this document when rejecting a PR.

## 1. Handlers MUST NOT mutate cluster state

A handler MUST NOT call any of the following, directly or transitively:

- `client.Client.Update` / `client.Client.Patch` / `client.Client.Create`
  / `client.Client.Delete` / `client.Client.Status().Update`
- `versioned.Clientset` mutators (`.Create()` / `.Update()` / etc.)
- Any `helm.Install/Upgrade/Uninstall`
- Any `nats.Conn.Publish` / `EventRecorder.Eventf`

State writes are expressed by returning a `PatchAM` `Effect` on the `Result`.
The `Runner` in `pkg/applifecycle/runner/` is the SOLE writer to `etcd`, and
performs a `DeepEqual` + optimistic-lock `Patch` after the handler returns.

Spec and annotation writes go through `PatchSpec`, which needs reviewer
sign-off per use — `Spec` belongs to the apiserver, `Status` to the state
machine. A `PatchSpec` MUST precede any `PatchAM` in the same `Result`; see
its doc-comment for why.

## 2. Handlers MUST NOT spawn goroutines

A handler MUST NOT contain any `go func`. Long-running side effects
(helm install/upgrade/uninstall, image pull, namespace deletion polling,
etc.) are expressed as a `CallWorkload` `Effect`. The `Runner` is the SOLE
goroutine starter; it manages the workload's lifecycle (context cancellation,
timeout, error propagation).

## 3. Handlers MUST NOT block

A handler MUST return within milliseconds. It MUST NOT call:

- `time.Sleep`
- `k8s.io/apimachinery/pkg/util/wait.Poll` / `.PollImmediate` / `.PollUntil`
- `channel <-` reads
- Any blocking I/O beyond the ambient `context.Context` cancellation semantics

Waiting semantics are expressed as `ScheduleTick{After: duration}`. The
`Runner` translates this into `reconcile.Result{RequeueAfter: duration}` so
`controller-runtime`'s workqueue handles scheduling.

## 4. Handlers' `Result.Next` MUST be declared in the TransitionTable

After the handler returns, `Engine.Step` performs

```go
if res.Next != "" && res.Next != w.Current() && !table.Allows(w.Current(), res.Next) {
    return ErrUndeclaredTransition{From: w.Current(), To: res.Next}
}
```

Any state jump that is not declared in the `TransitionTable` is rejected at
runtime. This eliminates the historical drift between six independent tables
in the legacy `pkg/appstate/state_transition.go`.

## 5. Long-running effects MUST specify explicit callback Actions

Every `CallWorkload` effect MUST carry `OnSuccess Action` and `OnFailure
Action`. When the workload terminates, the `Runner` feeds the corresponding
`Action` back into `Engine.Step(world, action)`, closing the loop through
the state machine. Fire-and-forget or "let some other controller pick this
up implicitly" is forbidden. Every external path back into the state machine
MUST be traceable to a declared transition.

---

## Rationale

These 5 rules together guarantee:

- `handler = pure function of (World, Action) -> Result`
- handlers return in milliseconds
- handlers do not hold any Kubernetes/helm/NATS resource handles
- all state transitions are declared and enforced at runtime
- all external callbacks close the loop through the state machine

They directly prevent the three historical bug patterns that motivated this
refactor:

1. `pending_app.Exec` holding a process-wide mutex during a cross-AM List+Write
   critical section (violates rule 1, 3)
2. `installing_app.Exec` starting nested goroutines that outlive the
   Reconcile call and race with subsequent reconciles (violates rule 2)
3. `types.go::waitForNamespaceDeleted` running `wait.PollImmediate` inside
   the Reconcile worker for up to 30 minutes, blocking every other AM
   (violates rule 3)

## Enforcement

The rules are enforced at three layers:

1. **Code review**: reviewers reject PRs by linking this document
2. **`forbidigo` linter**: `.golangci.yml` disallows the forbidden symbols
   inside `pkg/applifecycle/handlers/**`
3. **Custom analyzer**: a ~50 LoC `golang.org/x/tools/go/analysis` pass
   verifies rule 4 (declared transitions) at build time
