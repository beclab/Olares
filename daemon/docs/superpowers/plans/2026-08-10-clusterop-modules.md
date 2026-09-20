# Extensible Cluster Operation Modules Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make cluster operation types compile-time modules so adding one same-package Go file adds a supported operation without modifying existing core code.

**Architecture:** A frozen registry maps operation type strings to modules. `Manager` owns idempotency, serialization, persistence, and legal state transitions, while each module validates, executes, reports phase, optionally recovers, and optionally executes on a worker through a restricted runtime. Reboot and shutdown become built-in modules while preserving their existing power behavior and rolling-upgrade endpoint.

**Tech Stack:** Go 1.25, Fiber v2, existing JSON file store, existing JWS/replay protection, Go `testing`

## Global Constraints

- Modules are compiled Go files in `pkg/cluster/clusterop`; there is no runtime plugin loader.
- Production modules self-register from `init()` and the registry freezes before serving requests.
- All operation types share the existing global single-operation lock.
- `params` is `json.RawMessage`, is deliberately not JWS-bound, and must never be persisted.
- Persist only a canonical SHA-256 parameter digest for request ID idempotency.
- Persist module recovery evidence only in optional `moduleState`; modules must
  keep secrets, raw params, credentials, passwords, and tokens out of it.
- Preserve existing operation JSON compatibility, stable error codes, credential non-persistence, power ordering, reboot proof, and shutdown `command_issued` semantics.
- Keep `/command/power-node` for reboot/shutdown rolling compatibility; new node-capable modules use `/command/cluster-operation`.
- Use test-first changes and run the focused test before each broader regression command.
- Do not add third-party dependencies.

---

### Task 1: Define module contracts and an isolated, freezable registry

**Files:**
- Create: `pkg/cluster/clusterop/module.go`
- Create: `pkg/cluster/clusterop/registry.go`
- Create: `pkg/cluster/clusterop/registry_test.go`
- Modify later, not in this task: `pkg/cluster/clusterop/types.go`

**Interfaces:**
- Produces: `OperationModule`, `RecoverableModule`, `NodeOperationModule`, `Outcome`, `RunRequest`, `NodeRequest`
- Produces: `NewRegistry() *ModuleRegistry`, `DefaultRegistry() *ModuleRegistry`, `MustRegisterModule(OperationModule)`, `(*ModuleRegistry).Register`, `Lookup`, `Parse`, and `Freeze`
- Keeps: the existing `ParseType` implementation until built-in modules register in Task 4

- [ ] **Step 1: Write failing registry tests**

Create `registry_test.go` with table-driven fakes and these behaviors:

```go
type registryTestModule struct{ typ Type }

func (m registryTestModule) Type() Type { return m.typ }
func (m registryTestModule) Validate(CreateRequest) error { return nil }
func (m registryTestModule) Phase(Operation) (nodestatus.Phase, bool) {
	return "", false
}
func (m registryTestModule) Run(context.Context, Runtime, RunRequest) Outcome {
	return Outcome{Status: StatusSucceeded}
}

func TestRegistryParseRegisteredType(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(registryTestModule{typ: Type("example")}); err != nil {
		t.Fatal(err)
	}
	got, err := reg.Parse("example")
	if err != nil || got != Type("example") {
		t.Fatalf("Parse() = %q, %v", got, err)
	}
}

func TestRegistryRejectsDuplicateType(t *testing.T) {
	reg := NewRegistry()
	module := registryTestModule{typ: Type("example")}
	if err := reg.Register(module); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(module); err == nil {
		t.Fatal("duplicate registration succeeded")
	}
}

func TestRegistryRejectsEmptyType(t *testing.T) {
	if err := NewRegistry().Register(registryTestModule{}); err == nil {
		t.Fatal("empty type registration succeeded")
	}
}

func TestRegistryRejectsRegistrationAfterFreeze(t *testing.T) {
	reg := NewRegistry()
	reg.Freeze()
	if err := reg.Register(registryTestModule{typ: Type("late")}); err == nil {
		t.Fatal("registration after Freeze succeeded")
	}
}
```

- [ ] **Step 2: Run tests and verify the missing API failure**

Run:

```bash
go test ./pkg/cluster/clusterop -run 'TestRegistry' -count=1
```

Expected: FAIL to compile because `OperationModule`, `Runtime`, `Outcome`, and `NewRegistry` do not exist.

- [ ] **Step 3: Add the module contracts**

Create `module.go`:

```go
package clusterop

import (
	"context"
	"encoding/json"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

type Outcome struct {
	Status             Status
	Code               string
	Error              string
	CommandIssuedUntil time.Time
}

func (o Outcome) valid() bool {
	switch o.Status {
	case StatusSucceeded, StatusPartiallyFailed, StatusFailed, StatusCommandIssued:
		return true
	default:
		return false
	}
}

type RunRequest struct {
	Creds  Credentials
	Params json.RawMessage
}

type NodeRequest struct {
	PeerRequest
	Params json.RawMessage `json:"params,omitempty"`
}

type Runtime interface {
	Operation() (Operation, bool)
	CanContinue() bool
	StartStep(string) error
	FinishStep(string, StepStatus, string, string) error
	InitNodes([]NodeResult) error
	UpdateNode(string, func(*NodeResult)) error
	SetHostBootID(string) error
	SetModuleState(json.RawMessage) error
	SetCommandIssuedUntil(time.Time) error
	Complete(Outcome) error
	Now() time.Time
	Context() context.Context
}

type OperationModule interface {
	Type() Type
	Validate(CreateRequest) error
	Phase(Operation) (nodestatus.Phase, bool)
	Run(context.Context, Runtime, RunRequest) Outcome
}

type RecoverableModule interface {
	Recover(context.Context, Runtime, Operation)
}

type NodeOperationModule interface {
	ExecuteNode(context.Context, NodeRequest) error
}
```

Optional interfaces avoid forcing modules with no worker or special recovery
behavior to implement empty methods.

- [ ] **Step 4: Implement the registry**

Create `registry.go` with a mutex-protected map. `Register` must trim and reject
empty type names, reject nil modules and duplicates, and reject writes after
freeze. `Lookup` and `Parse` must work after freeze. Use this error form to
preserve existing client behavior:

```go
fmt.Errorf("unsupported cluster operation type %q", value)
```

`MustRegisterModule` registers against a package-level default and panics on
error:

```go
var defaultRegistry = NewRegistry()

func DefaultRegistry() *ModuleRegistry { return defaultRegistry }

func MustRegisterModule(module OperationModule) {
	if err := defaultRegistry.Register(module); err != nil {
		panic(err)
	}
}
```

- [ ] **Step 5: Run focused and package tests**

Run:

```bash
go test ./pkg/cluster/clusterop -run 'TestRegistry' -count=1
go test ./pkg/cluster/clusterop -count=1
```

Expected: PASS. Existing type parsing still uses its old switch at this
intermediate point.

- [ ] **Step 6: Commit this independently reviewable foundation**

```bash
git add pkg/cluster/clusterop/module.go pkg/cluster/clusterop/registry.go pkg/cluster/clusterop/registry_test.go
git commit -m "refactor(clusterop): add operation module registry"
```

---

### Task 2: Add canonical, non-secret parameter idempotency

**Files:**
- Create: `pkg/cluster/clusterop/params.go`
- Create: `pkg/cluster/clusterop/params_test.go`
- Modify: `pkg/cluster/clusterop/manager.go`
- Modify: `pkg/cluster/clusterop/types.go`
- Test: `pkg/cluster/clusterop/manager_test.go`

**Interfaces:**
- Produces: `CanonicalParams(json.RawMessage) (json.RawMessage, error)`
- Produces: `DigestParams(json.RawMessage) (string, error)`
- Adds: `CreateRequest.Params json.RawMessage`
- Adds: `Operation.ParamsDigest string`
- Adds: `Operation.ModuleState json.RawMessage`

- [ ] **Step 1: Write failing canonicalization tests**

Cover omitted/empty/`{}` equivalence, object key ordering, changed values, arrays,
and invalid JSON:

```go
func TestDigestParamsEmptyAndObjectAreEqual(t *testing.T) {
	omitted, err := DigestParams(nil)
	if err != nil {
		t.Fatal(err)
	}
	emptyObject, err := DigestParams(json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if omitted != emptyObject {
		t.Fatalf("digests differ: %q != %q", omitted, emptyObject)
	}
}

func TestDigestParamsCanonicalizesObjectOrder(t *testing.T) {
	left, _ := DigestParams(json.RawMessage(`{"a":1,"b":2}`))
	right, _ := DigestParams(json.RawMessage(`{"b":2,"a":1}`))
	if left != right {
		t.Fatalf("digests differ: %q != %q", left, right)
	}
}

func TestDigestParamsRejectsInvalidJSON(t *testing.T) {
	if _, err := DigestParams(json.RawMessage(`{`)); err == nil {
		t.Fatal("invalid JSON accepted")
	}
}
```

- [ ] **Step 2: Verify tests fail**

Run:

```bash
go test ./pkg/cluster/clusterop -run 'TestDigestParams' -count=1
```

Expected: FAIL to compile because `DigestParams` does not exist.

- [ ] **Step 3: Implement canonicalization and digest**

In `params.go`, treat omitted and whitespace-only input as `{}`. Decode with
`json.Decoder`, require exactly one JSON value, normalize through
`json.Marshal`, and compute a lowercase hex SHA-256 digest:

```go
func DigestParams(raw json.RawMessage) (string, error) {
	canonical, err := CanonicalParams(raw)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}
```

- [ ] **Step 4: Write failing manager idempotency tests**

Extend `manager_test.go`:

```go
func TestCreateRejectsSameRequestIDWithDifferentParams(t *testing.T) {
	m := newTestManager(t)
	first := validCreateRequest()
	first.Params = json.RawMessage(`{"value":1}`)
	if _, err := m.Create(context.Background(), first); err != nil {
		t.Fatal(err)
	}

	second := first
	second.Params = json.RawMessage(`{"value":2}`)
	_, err := m.Create(context.Background(), second)
	var conflict *RequestConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("Create() error = %v, want RequestConflictError", err)
	}
}
```

Also assert the persisted `Operation` has a non-empty `ParamsDigest`, marshaled
JSON contains `paramsDigest`, and neither JSON nor `Operation` contains raw
parameters.

- [ ] **Step 5: Add model fields and manager comparison**

Add:

```go
ParamsDigest string `json:"paramsDigest,omitempty"`
ModuleState  json.RawMessage `json:"moduleState,omitempty"`
```

to `Operation`, and:

```go
Params json.RawMessage
```

to `CreateRequest`. Compute the digest before locking in `Manager.Create`.
Include it in `sameIntent` and in the new operation record.

For records created by an older daemon, an empty stored digest matches only
empty/`{}` input. Do not backfill old files during load.

`ModuleState` is the only generic persistence field a module may use for
restart evidence. It must contain no raw params, credentials, passwords, or
tokens; add a JSON round-trip test proving it is optional and backward
compatible.

- [ ] **Step 6: Run tests**

Run:

```bash
go test ./pkg/cluster/clusterop -run 'TestDigestParams|TestCreateRejectsSameRequestIDWithDifferentParams|TestCreate' -count=1
go test ./pkg/cluster/clusterop -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit parameter idempotency**

```bash
git add pkg/cluster/clusterop/params.go pkg/cluster/clusterop/params_test.go pkg/cluster/clusterop/manager.go pkg/cluster/clusterop/types.go pkg/cluster/clusterop/manager_test.go
git commit -m "feat(clusterop): track module parameter intent"
```

---

### Task 3: Add a persistence-safe module runtime

**Files:**
- Create: `pkg/cluster/clusterop/runtime.go`
- Create: `pkg/cluster/clusterop/runtime_test.go`
- Modify: `pkg/cluster/clusterop/manager.go`

**Interfaces:**
- Implements the `Runtime` interface defined in Task 1
- Produces: `newRuntime(*Manager, string, context.Context) Runtime`
- Produces: `(*Manager).complete(string, Outcome) error`

- [ ] **Step 1: Write failing runtime state tests**

Use a stored pending operation and test:

```go
func TestRuntimeRejectsMissingStep(t *testing.T) {
	rt := newRuntimeWithOperation(t, StatusRunning)
	err := rt.FinishStep("missing", StepSucceeded, "", "")
	if !errors.Is(err, ErrStepNotFound) {
		t.Fatalf("FinishStep() = %v, want ErrStepNotFound", err)
	}
}

func TestRuntimeRejectsTerminalRollback(t *testing.T) {
	rt := newRuntimeWithOperation(t, StatusSucceeded)
	if err := rt.StartStep("late"); !errors.Is(err, ErrOperationTerminal) {
		t.Fatalf("StartStep() = %v, want ErrOperationTerminal", err)
	}
}

func TestRuntimeRejectsDuplicateStepCompletion(t *testing.T) {
	rt := newRuntimeWithOperation(t, StatusRunning)
	if err := rt.StartStep("work"); err != nil {
		t.Fatal(err)
	}
	if err := rt.FinishStep("work", StepSucceeded, "", ""); err != nil {
		t.Fatal(err)
	}
	if err := rt.FinishStep("work", StepSucceeded, "", ""); !errors.Is(err, ErrStepSettled) {
		t.Fatalf("second FinishStep() = %v, want ErrStepSettled", err)
	}
}
```

Add node-not-found, duplicate completion, invalid `Outcome.Status`, and
persistence-failed coverage.

- [ ] **Step 2: Verify runtime tests fail**

Run:

```bash
go test ./pkg/cluster/clusterop -run 'TestRuntime' -count=1
```

Expected: FAIL because `Runtime` methods and stable internal sentinel errors do
not exist.

- [ ] **Step 3: Add runtime errors and the concrete implementation**

Implement the Task 1 interface with error-returning methods so rejected
mutations are observable. Define package-private sentinel errors `ErrOperationTerminal`,
`ErrStepNotFound`, `ErrStepSettled`, `ErrNodeNotFound`, and
`ErrInvalidOutcome`.

- [ ] **Step 4: Implement checked mutations**

`operationRuntime` holds only manager, operation ID, and context. Every method
calls a new checked manager mutation that:

1. locks `Manager.mu`;
2. rejects missing operations and `persistFailed`;
3. validates against current state;
4. mutates;
5. updates timestamps;
6. saves under the same lock;
7. applies existing persistence-failure settlement.

Do not expose `Store`, `Manager.ops`, or `Manager.mu` through the interface.
`Complete` accepts only `Outcome.valid()` and sets `FinishedAt`,
`CommandIssuedUntil`, active-lock state, code, and safe error atomically.
`SetModuleState` copies the raw bytes before persistence so callers cannot
mutate stored state through a shared slice.

- [ ] **Step 5: Run focused and regression tests**

Run:

```bash
go test ./pkg/cluster/clusterop -run 'TestRuntime' -count=1
go test ./pkg/cluster/clusterop -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit the runtime boundary**

```bash
git add pkg/cluster/clusterop/runtime.go pkg/cluster/clusterop/runtime_test.go pkg/cluster/clusterop/manager.go
git commit -m "refactor(clusterop): constrain module state updates"
```

---

### Task 4: Convert reboot and shutdown into self-registering modules

**Files:**
- Create: `pkg/cluster/clusterop/module_reboot.go`
- Create: `pkg/cluster/clusterop/module_shutdown.go`
- Create: `pkg/cluster/clusterop/module_reboot_test.go`
- Create: `pkg/cluster/clusterop/module_shutdown_test.go`
- Create: `pkg/cluster/clusterop/power_sequence.go`
- Modify: `pkg/cluster/clusterop/orchestrate.go`
- Modify: `pkg/cluster/clusterop/power.go`
- Modify: `pkg/cluster/clusterop/types.go`
- Modify: `pkg/cluster/clusterop/manager.go`
- Test: all existing files in `pkg/cluster/clusterop/*_test.go`

**Interfaces:**
- Registers: `rebootModule` and `shutdownModule`
- Changes: `ParseType` to `DefaultRegistry().Parse`
- Changes: `PhaseFor` and `Manager.ActivePhase` to module lookup
- Changes: `Manager.run` to invoke `module.Run`
- Produces: `NewManagerWithRegistry(Deps, *ModuleRegistry) (*Manager, error)`
- Preserves: `PowerHost`, `LocalPowerSupport`, all existing power wire behavior

- [ ] **Step 1: Write failing built-in module contract tests**

```go
func TestRebootModuleContract(t *testing.T) {
	module, ok := DefaultRegistry().Lookup(TypeReboot)
	if !ok {
		t.Fatal("reboot module is not registered")
	}
	phase, ok := module.Phase(Operation{Type: TypeReboot, Status: StatusRunning})
	if !ok || phase != nodestatus.PhaseRestarting {
		t.Fatalf("Phase() = %q, %v", phase, ok)
	}
}

func TestShutdownModuleContract(t *testing.T) {
	module, ok := DefaultRegistry().Lookup(TypeShutdown)
	if !ok {
		t.Fatal("shutdown module is not registered")
	}
	phase, ok := module.Phase(Operation{Type: TypeShutdown, Status: StatusRunning})
	if !ok || phase != nodestatus.PhaseShuttingDown {
		t.Fatalf("Phase() = %q, %v", phase, ok)
	}
}
```

Add validation assertions for the existing cluster/node scope behavior,
including control-node node shutdown rejection.

- [ ] **Step 2: Verify module tests fail**

Run:

```bash
go test ./pkg/cluster/clusterop -run 'TestRebootModule|TestShutdownModule' -count=1
```

Expected: FAIL because no built-in module registers yet.

- [ ] **Step 3: Extract shared power sequencing without behavior changes**

Move these helpers from `orchestrate.go` to `power_sequence.go`, retaining their
current bodies and tests:

- `plan`, `errPrecheckFailed`
- `runNode`, `precheck`, `checkControlNode`, `checkWorker`, `baseline`
- `splitCluster`, `commandWorkers`, `rebootProgress`, `nodeUnavailable`
- `awaitWorkerShutdown`, `awaitRestarts`, `powerMaster`
- step/node update helpers until callers use Runtime

Remove `requiredCapability`, `runReboot`, `runShutdown`, and the type switch from
core orchestration. Pass capability, baseline requirement, wait strategy, and
master grace duration from the selected module instead.

First run the unchanged behavior anchors:

```bash
go test ./pkg/cluster/clusterop -run 'TestRebootPowersTheControlNodeLast|TestShutdownIssuesCommandsAndClaimsNothingMore|TestNodeScope' -count=1
```

Expected: PASS before changing manager dispatch.

- [ ] **Step 4: Implement the reboot module**

`module_reboot.go` must:

- register from `init`;
- validate existing reboot scopes;
- return `PhaseRestarting`;
- run precheck with `nodestatus.CapPowerReboot`;
- require BootID baselines;
- dispatch workers through the existing power peer path;
- await workers on a new Ready boot;
- issue the control-node reboot last;
- return `StatusCommandIssued` with the Ready timeout;
- implement `Recover` with the existing host BootID and control-node Ready
  confirmation;
- implement `ExecuteNode` by calling `PowerHost(ctx, TypeReboot)`.

Keep all existing stable error codes and safe messages byte-for-byte unless a
test demonstrates an existing inconsistency.

- [ ] **Step 5: Implement the shutdown module**

`module_shutdown.go` must:

- register from `init`;
- validate existing shutdown scopes;
- return `PhaseShuttingDown`;
- run precheck with `nodestatus.CapPowerShutdown`;
- skip BootID baselines and restart waiting;
- dispatch workers before the control node;
- return `StatusCommandIssued` with the Down timeout;
- keep command-issued records unchanged during recovery;
- implement `ExecuteNode` by calling `PowerHost(ctx, TypeShutdown)`.

- [ ] **Step 6: Switch core lookup and dispatch to the registry**

Make `ParseType`:

```go
func ParseType(value string) (Type, error) {
	return DefaultRegistry().Parse(value)
}
```

Keep `NewManager(deps)` as a compatibility wrapper and add:

```go
func NewManagerWithRegistry(deps Deps, registry *ModuleRegistry) (*Manager, error)
```

The manager stores that registry and uses it for create, run, phase, and
recovery. Tests use `NewManagerWithRegistry`; production `NewManager` injects
`DefaultRegistry`.

Make `Manager.run` lookup the module, set running, create a runtime, invoke
`Run`, and call `Runtime.Complete`. If lookup fails, complete with
`StatusFailed` and `CodeUnsupportedOperation`.

Make `PhaseFor`/`ActivePhase` lookup `op.Type` and call `module.Phase`. There
must be no reboot/shutdown switch in `types.go` or `manager.go`.

Move `newPowerCommand` and `powerCommandName` selection into their respective
power modules. `PowerHost` should execute behavior supplied by the selected
node-capable power module rather than switching on type.

- [ ] **Step 7: Run all clusterop tests and inspect hardcoding**

Run:

```bash
go test ./pkg/cluster/clusterop -count=1
rg 'switch (opType|t|op)|case Type(Reboot|Shutdown)' pkg/cluster/clusterop
```

Expected: tests PASS. Matches may remain inside `module_reboot.go`,
`module_shutdown.go`, compatibility constants, and tests; no type dispatch
switch remains in `types.go`, `manager.go`, `power.go`, or shared sequencing.

- [ ] **Step 8: Commit the built-in module migration**

```bash
git add pkg/cluster/clusterop
git commit -m "refactor(clusterop): move power operations into modules"
```

---

### Task 5: Make restart recovery and panic containment module-driven

**Files:**
- Create: `pkg/cluster/clusterop/recovery.go`
- Create: `pkg/cluster/clusterop/recovery_test.go`
- Create: `pkg/cluster/clusterop/panic_test.go`
- Modify: `pkg/cluster/clusterop/manager.go`
- Modify: `pkg/cluster/clusterop/module_reboot.go`
- Modify: `pkg/cluster/clusterop/module_shutdown.go`

**Interfaces:**
- Produces: `MarkInterrupted(*Operation, time.Time)`
- Uses: optional `RecoverableModule`
- Adds safe wrappers around module `Run` and `Recover`
- Produces: `ExecuteNode(context.Context, *ModuleRegistry, NodeRequest) error`

- [ ] **Step 1: Write failing recovery tests**

Cover unknown historical types and fake recovery dispatch:

```go
func TestUnknownHistoricalModuleSettlesFailed(t *testing.T) {
	op := Operation{ID: "op-1", Type: Type("removed"), Status: StatusRunning}
	storeOperation(t, op)
	m := newManagerWithRegistry(t, NewRegistry())
	got, _ := m.Get(op.ID)
	if got.Status != StatusFailed || got.Code != CodeUnsupportedOperation {
		t.Fatalf("loaded operation = %#v", got)
	}
}

func TestManagerCallsModuleRecovery(t *testing.T) {
	module := &recoveringTestModule{typ: Type("recoverable")}
	reg := NewRegistry()
	if err := reg.Register(module); err != nil {
		t.Fatal(err)
	}
	storeOperation(t, Operation{ID: "op-1", Type: module.typ, Status: StatusRunning})
	_ = newManagerWithRegistry(t, reg)
	if module.recoverCalls != 1 {
		t.Fatalf("Recover calls = %d", module.recoverCalls)
	}
}
```

- [ ] **Step 2: Write failing panic tests**

Add fake modules that panic from `Run` and `Recover`. Assert:

- the process does not panic;
- the operation becomes failed;
- `Code` is a stable safe framework code;
- panic text is absent from persisted `Error`;
- the operation lock is released unless persistence itself failed.

- [ ] **Step 3: Verify focused tests fail**

Run:

```bash
go test ./pkg/cluster/clusterop -run 'TestUnknownHistoricalModule|TestManagerCallsModuleRecovery|TestRunPanic|TestRecoverPanic' -count=1
```

Expected: FAIL because generic recovery and panic boundaries are missing.

- [ ] **Step 4: Implement generic recovery**

Extract current `markInterrupted` to:

```go
func MarkInterrupted(op *Operation, now time.Time) {
	// Preserve the current daemon_restarted status, step, node, and timestamp
	// transitions exactly.
}
```

At load, lookup every operation's module:

- unknown type: settle with `CodeUnsupportedOperation`;
- module implements `RecoverableModule`: invoke it through a safe wrapper;
- no recovery interface and non-terminal: call `MarkInterrupted`;
- terminal record: retain it unless the module's recovery logic explicitly
  confirms a command-issued result.

- [ ] **Step 5: Add panic boundaries**

Use narrow `defer`/`recover` wrappers around `Run` and `Recover`. Log the panic
and stack with `klog`, but persist only a fixed safe message. Do not catch
panics around unrelated manager code.

Add a package-level `ExecuteNode` helper that looks up
`NodeOperationModule`, recovers only module panics, and maps unsupported types
or missing node capability to `CodeUnsupportedOperation`. Both HTTP node
handlers in Task 6 call this helper, so panic policy is not duplicated.

- [ ] **Step 6: Run recovery and full package tests**

Run:

```bash
go test ./pkg/cluster/clusterop -run 'TestUnknownHistoricalModule|TestManagerCallsModuleRecovery|TestRunPanic|TestRecoverPanic|TestControlNodeReboot' -count=1
go test ./pkg/cluster/clusterop -count=1
```

Expected: PASS, including existing reboot restart-confirmation tests.

- [ ] **Step 7: Commit recovery boundaries**

```bash
git add pkg/cluster/clusterop/recovery.go pkg/cluster/clusterop/recovery_test.go pkg/cluster/clusterop/panic_test.go pkg/cluster/clusterop/manager.go pkg/cluster/clusterop/module_reboot.go pkg/cluster/clusterop/module_shutdown.go
git commit -m "refactor(clusterop): delegate recovery to operation modules"
```

---

### Task 6: Expose module parameters and generic worker execution over HTTP

**Files:**
- Modify: `pkg/cluster/clusterop/peer.go`
- Modify: `pkg/cluster/clusterop/peer_test.go`
- Modify: `internel/apiserver/handlers/handler_cluster_operations.go`
- Modify: `internel/apiserver/handlers/handler_cluster_operations_test.go`
- Create: `internel/apiserver/handlers/handler_cluster_operation_node.go`
- Create: `internel/apiserver/handlers/handler_cluster_operation_node_test.go`
- Modify: `internel/apiserver/handlers/handler_power_node.go`
- Modify: `internel/apiserver/handlers/handler_power_node_test.go`
- Modify: `internel/apiserver/handlers/command_group.go`

**Interfaces:**
- Adds request field: `Params json.RawMessage`
- Adds path: `ClusterOperationPath = "/command/cluster-operation"`
- Adds handler: `PostClusterOperationNode`
- Keeps path and wire behavior: `/command/power-node`

- [ ] **Step 1: Write failing create-request HTTP tests**

Add a handler test that posts:

```json
{"type":"example","requestId":"request-1","scope":"cluster","clusterId":"cluster-1","params":{"value":1}}
```

Use a fake `clusterOperationManager` to assert `CreateRequest.Params` contains
the exact raw JSON value. Add invalid JSON and module-validation error cases.

- [ ] **Step 2: Write failing generic node endpoint tests**

Register a test module in an isolated handler registry and assert:

- a matching signed request calls `ExecuteNode` once;
- unknown type returns bad request with `unsupported_operation`;
- replayed signature is rejected;
- module validation failure does not execute;
- a panic returns a safe server error without exposing panic text.

- [ ] **Step 3: Verify handler tests fail**

Run:

```bash
go test ./internel/apiserver/handlers -run 'TestCreateClusterOperationAcceptsParams|TestClusterOperationNode' -count=1
```

Expected: FAIL because the request field, route, and handler do not exist.

- [ ] **Step 4: Add params to the public create path**

Add:

```go
Params json.RawMessage `json:"params,omitempty"`
```

to `createClusterOperationRequest` and pass it unchanged into
`clusterop.CreateRequest`. Keep common request ID, cluster ID, owner, and JWS
binding checks in the handler. Move scope/target semantic support checks into
`module.Validate`; the binding must still compare the signed scope and target
exactly.

- [ ] **Step 5: Add generic peer dispatch**

In `peer.go`, keep:

```go
const PeerPath = "/command/power-node"
```

and add:

```go
const ClusterOperationPath = "/command/cluster-operation"
```

Implement generic dispatch using the existing fan-out transport and
`NodeRequest`. Do not send the access token on the peer hop; preserve the
signature-only behavior.

- [ ] **Step 6: Add and register the generic node handler**

`PostClusterOperationNode` must parse `clusterop.NodeRequest`, perform the same
common binding, target, cluster ID, and replay checks as `PostPowerNode`, lookup
the module, run node-side validation, assert `NodeOperationModule`, and invoke
`ExecuteNode` through a panic-safe package function.

Register:

```go
commandGroup.Post("/cluster-operation", handler.PostClusterOperationNode)
```

in `command_group.go`.

- [ ] **Step 7: Preserve the rolling power endpoint**

Keep `/command/power-node` and its request JSON compatible. Replace only its
execution switch: lookup the registered reboot/shutdown module and call its
`ExecuteNode`. Reboot and shutdown master fan-out must continue to use
`PeerPath`, so new master to old worker remains supported.

Add explicit peer-path assertions:

```go
func TestBuiltInPowerDispatchUsesLegacyPath(t *testing.T) {
	// The fake HTTP peer records RequestURI.
	// Run reboot and shutdown dispatch and assert "/command/power-node".
}
```

- [ ] **Step 8: Run handler, peer, and signature regressions**

Run:

```bash
go test ./pkg/cluster/clusterop -run 'TestPeer|TestBuiltInPowerDispatch' -count=1
go test ./internel/apiserver/handlers -run 'ClusterOperation|PowerNode|Signature|Owner|InitCluster' -count=1
```

Expected: PASS. Existing power-node and signature-binding tests must remain
unchanged where possible.

- [ ] **Step 9: Commit the generic HTTP execution path**

```bash
git add pkg/cluster/clusterop/peer.go pkg/cluster/clusterop/peer_test.go internel/apiserver/handlers
git commit -m "feat(clusterop): add generic module execution endpoint"
```

---

### Task 7: Prove add-a-file extensibility and run final verification

**Files:**
- Create: `pkg/cluster/clusterop/module_extensibility_test.go`
- Modify: `internel/apiserver/handlers/handler_cluster_operations.go`
- Modify: `docs/superpowers/specs/2026-08-10-clusterop-modules-design.md`

**Interfaces:**
- Demonstrates that an operation type is accepted, dispatched, phased, and
  completed solely through one test module file
- Freezes the default registry from `InitClusterOperations` before manager
  construction

- [ ] **Step 1: Add the single-file extensibility proof**

In `module_extensibility_test.go`, define one module, register it in an isolated
registry, build a manager with that registry, and prove:

```go
func TestAddingModuleRequiresNoCoreTypeChanges(t *testing.T) {
	reg := NewRegistry()
	module := &exampleOperationModule{typ: Type("example")}
	if err := reg.Register(module); err != nil {
		t.Fatal(err)
	}
	m := newManagerWithRegistry(t, reg)
	req := validCreateRequest()
	req.Type = module.typ
	req.Params = json.RawMessage(`{"value":1}`)

	op, err := m.Create(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	waitForStatus(t, m, op.ID, StatusSucceeded)
	if module.runCalls != 1 {
		t.Fatalf("Run calls = %d", module.runCalls)
	}
}
```

The test module returns its own phase and validates its own params. The test
must not edit a switch or production registry list.

- [ ] **Step 2: Freeze production registration at initialization**

In `InitClusterOperations`, freeze `DefaultRegistry` before constructing the
manager. Inject the same registry into the manager and generic node handler
lookup. Add an initialization test proving late registration fails.

- [ ] **Step 3: Run formatting and focused static checks**

Run:

```bash
gofmt -w pkg/cluster/clusterop internel/apiserver/handlers
go vet ./pkg/cluster/clusterop/... ./internel/apiserver/handlers/...
```

Expected: no output and exit code 0.

- [ ] **Step 4: Run complete related regression suites**

Run:

```bash
go test ./pkg/cluster/clusterop/... -count=1
go test ./pkg/cluster/fanout/... -count=1
go test ./internel/apiserver/handlers/... -count=1
```

Expected: PASS.

- [ ] **Step 5: Run daemon-wide tests**

Run:

```bash
go test ./... -count=1
```

Expected: PASS. If an unrelated pre-existing failure occurs, record its exact
package and output instead of weakening the related test gates.

- [ ] **Step 6: Verify no core operation-type hardcoding remains**

Run:

```bash
rg -n 'case TypeReboot|case TypeShutdown|TypeReboot.*TypeShutdown|TypeShutdown.*TypeReboot' \
  pkg/cluster/clusterop internel/apiserver/handlers
```

Expected: matches only in built-in module files, compatibility constants, and
tests. No match may implement type acceptance, manager dispatch, phase mapping,
capability mapping, command creation, or recovery in shared core files.

- [ ] **Step 7: Update design status and commit final proof**

Change the design status to `Implemented` only after all preceding commands
pass, then run:

```bash
git add pkg/cluster/clusterop/module_extensibility_test.go internel/apiserver/handlers/handler_cluster_operations.go docs/superpowers/specs/2026-08-10-clusterop-modules-design.md
git commit -m "test(clusterop): prove module-only extensibility"
```

## Final acceptance checklist

- [ ] Adding one same-package module file makes a type parseable and executable.
- [ ] Shared core files contain no reboot/shutdown dispatch switch.
- [ ] Existing reboot/shutdown HTTP and persisted operation formats remain
  compatible.
- [ ] Reboot still proves a changed BootID and Ready state.
- [ ] Shutdown still ends at `command_issued`.
- [ ] Raw params and credentials never appear in operation JSON.
- [ ] `moduleState` round-trips without changing old operation files and
  contains only module-approved non-secret recovery evidence.
- [ ] Different params under one request ID conflict.
- [ ] Unknown, duplicate, late, invalid, and panicking modules fail safely.
- [ ] Reboot/shutdown still use `/command/power-node` for old workers.
- [ ] Focused, handler, fan-out, vet, and daemon-wide verification all pass.
