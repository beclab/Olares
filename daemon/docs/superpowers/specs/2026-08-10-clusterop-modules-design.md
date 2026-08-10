# Design: Extensible cluster operation modules

**Date:** 2026-08-10
**Status:** Approved for implementation planning

## Goal

Refactor `pkg/cluster/clusterop` so a new operation type can be added by creating
one Go source file in the `clusterop` package. Adding an operation must not
require editing the manager, HTTP handlers, type parser, phase mapper, or
existing modules.

The modules are compiled into `olaresd`; this design does not load `.so`,
scripts, or configuration at runtime.

## Confirmed decisions

- Modules are Go files in the existing `clusterop` package.
- A module owns an independent validation, execution, and restart-recovery
  lifecycle; it does not have to follow the power workflow.
- Modules register themselves during package initialization.
- All cluster operations remain globally serialized.
- Operations may carry module-specific JSON parameters.
- Module parameters are deliberately not covered by the JWS binding. This means
  they can be modified independently of the signed type, scope, target, and
  request ID. The caller accepts this risk.
- Raw parameters, which may contain secrets, are never persisted. A digest is
  persisted for idempotency checks.
- Reboot and shutdown become built-in modules without changing their public
  behavior.

## Architecture

### Registry

Introduce a registry keyed by `Type`. It validates registrations and becomes
read-only before requests are served.

```go
type Registry interface {
	Register(OperationModule) error
	Lookup(Type) (OperationModule, bool)
	Parse(string) (Type, error)
}

func MustRegisterModule(module OperationModule)
```

Each production module is declared in `module_<type>.go` and calls
`MustRegisterModule` from `init`. Registration rejects an empty type, a
duplicate type, and an incomplete module. Invalid production registration is a
startup error rather than a partially available operation set.

`ParseType` becomes a compatibility wrapper over the default registry. It no
longer contains a switch. Tests can construct an isolated registry so test-only
modules do not contaminate global state.

### Module contract

The contract separates operation-specific policy from framework-owned state:

```go
type OperationModule interface {
	Type() Type
	Validate(CreateRequest) error
	Phase(Operation) (nodestatus.Phase, bool)
	Run(context.Context, Runtime, RunRequest) Outcome
	Recover(context.Context, Runtime, Operation)
	ExecuteNode(context.Context, NodeRequest) error
}
```

- `Validate` checks supported scopes, target rules, and module parameters.
- `Phase` describes how an active operation appears in cluster status.
- `Run` owns the complete operation lifecycle.
- `Recover` handles records found after daemon restart. A common default helper
  marks interrupted work as `daemon_restarted`.
- `ExecuteNode` supports modules that fan out work to nodes. Modules that do not
  support node execution return `unsupported_operation`.

The exact interface may be split into optional capability interfaces during
implementation if that avoids empty methods, but lookup and dispatch remain
type-driven and require no core switch.

### Framework responsibilities

`Manager` remains responsible for:

- request ID idempotency
- the global single-operation lock
- operation creation and retention
- atomic persistence
- invoking the selected module
- enforcing legal state transitions
- catching module panics
- exposing operation queries and the active phase

The manager does not know the names or semantics of reboot, shutdown, or future
operation types.

### Runtime

Modules receive a restricted `Runtime`, not the manager's maps, mutex, or
store. Runtime supports:

- reading a cloned current operation
- starting and finishing steps
- initializing and updating node results
- persisting module-owned, non-secret recovery state
- checking whether persistence still allows execution
- obtaining the framework clock and cancellation context

Every mutation is serialized and persisted by `Manager.update`. Runtime rejects
updates to missing steps or nodes, duplicate completion, and attempts to move a
terminal operation back to a non-terminal state.

`Run` returns an `Outcome` with one of `succeeded`, `failed`,
`partially_failed`, or `command_issued`. The manager applies and persists that
outcome. Internal causes are logged; only stable codes and safe messages enter
the operation record.

## Request and persistence model

`POST /cluster/operations` keeps its existing fields and adds:

```json
{
  "type": "example",
  "requestId": "request-1",
  "scope": "cluster",
  "target": "",
  "clusterId": "cluster-1",
  "params": {}
}
```

`params` is decoded as `json.RawMessage` and passed to the selected module.
Handlers perform only common parsing and authentication. The module validates
scope, target, and parameter semantics.

Before creation, the framework canonicalizes the JSON parameters and computes a
digest. `Operation` persists `paramsDigest`, but never the raw parameters.
Retries with the same request ID and a different digest return
`RequestConflictError`. Empty and omitted parameters have one canonical digest.

`Operation` also has an optional `moduleState` JSON field. A module may use it
for non-secret evidence required by restart recovery. The module owns its schema
and must not copy raw parameters, credentials, passwords, or tokens into it.
Runtime persists this field through the same checked state-update path.

Because the raw value is memory-only:

- it disappears when the run ends or the daemon exits;
- it cannot be used by recovery code after a restart;
- modules must persist only non-secret recovery evidence in their normal
  operation fields or `moduleState`;
- recovery must settle an interrupted operation if it cannot continue without
  the original parameters.

The JWS binding remains limited to `clusterId`, `type`, `requestId`, `scope`,
`target`, and expiry. The parameter digest is not signed. TLS and request
authentication do not provide cryptographic binding between those parameters
and the signed intent.

## Node execution

Add a generic internal endpoint:

`POST /command/cluster-operation`

The endpoint:

1. parses the common node request;
2. validates the existing operation binding and replay claim;
3. looks up the module by type;
4. lets the module validate its parameters;
5. calls `ExecuteNode`.

The generic peer request carries operation ID, request ID, type, scope, target,
cluster ID, and raw parameters. A shared fan-out helper sends it to selected
nodes. A new node-capable module therefore needs no new route.

Keep `/command/power-node` as a compatibility endpoint during rolling upgrades.
It delegates to the reboot or shutdown module. Built-in power modules may keep
using that endpoint until the compatibility window ends, so a newly upgraded
master can still communicate with older workers.

## Built-in power module migration

Move operation-specific behavior out of core switches:

- `module_reboot.go` declares reboot capability, restarting phase, BootID
  baseline and confirmation, worker restart waiting, and host command creation.
- `module_shutdown.go` declares shutdown capability, shutting-down phase,
  worker-first/master-last sequencing, and command-issued recovery semantics.
- shared power sequencing and command helpers remain in focused private files.

Existing constants and exported helpers may remain as compatibility aliases,
but type acceptance, execution dispatch, required capability, phase selection,
command creation, and recovery must resolve through the registry/module.

The migration must preserve:

- current HTTP request and response fields
- stable error codes and safe messages
- worker-first/master-last ordering
- reboot BootID and Ready proof
- shutdown `command_issued` behavior
- credential non-persistence
- existing operation files

## Restart recovery

When `NewManager` loads records:

- non-terminal records are passed to their module's `Recover`;
- reboot retains its BootID/Ready promotion logic;
- shutdown retains its command-issued behavior;
- modules that cannot resume use the common interrupted-operation helper;
- an unknown historical module type remains queryable but is settled as failed
  with `unsupported_operation`;
- command-issued grace periods continue to participate in the global lock.

Recovery never receives raw parameters.

## Error handling

- Registry errors fail startup deterministically.
- Unsupported request types return the existing
  `unsupported_operation` behavior.
- Module validation errors are safe client errors with stable codes.
- Module execution errors carry stable code, safe message, and an internal
  cause that is logged only.
- A panic in `Run`, `Recover`, or `ExecuteNode` is recovered at the framework
  boundary. The affected operation is settled as failed when possible, and the
  global lock is released according to normal persistence rules.
- Store failure stops further module execution and preserves the existing
  `state_persistence_failed` behavior.

## Testing

1. **Registry** — successful registration, duplicate and empty type rejection,
   incomplete module rejection, dynamic parsing, and isolated test registries.
2. **Manager dispatch** — fake modules verify validation, run dispatch, phase,
   global serialization, idempotency, parameter digest conflicts, and outcome
   persistence.
3. **Runtime state machine** — legal mutations plus rejection of missing
   steps/nodes, terminal rollback, duplicate completion, and updates after a
   persistence failure.
4. **Failure containment** — panics in all module entry points settle or reject
   safely without leaking internal errors.
5. **Recovery** — module recovery dispatch, unknown historical type, missing
   raw parameters, reboot promotion, shutdown command-issued behavior.
6. **Power regression** — run the existing clusterop and handler test suites
   unchanged where possible; wire format and ordering assertions must pass.
7. **HTTP** — dynamically registered type, optional parameters, digest
   idempotency, generic node endpoint, replay protection, and old power endpoint
   compatibility.
8. **Per-module contract** — every new module tests validation, run, recovery,
   phase, and node execution when supported.

## Out of scope

- Runtime-loaded Go plugins, scripts, or configuration modules
- Running different operation types concurrently
- Signing module parameters or their digest
- Persisting raw module parameters
- Dynamically adding HTTP routes
- Changing existing stable error codes

## Success criteria

- Core production code contains no reboot/shutdown switch for type acceptance,
  dispatch, phase, capability, command creation, or recovery.
- A test operation can be added as one new same-package module file and becomes
  accepted and executable without editing existing production files.
- Reboot and shutdown retain their current observable behavior.
- Unknown, duplicate, invalid, and panicking modules fail safely.
- All existing and new tests pass.
