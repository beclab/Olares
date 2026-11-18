# Real Adapters — Migration Cheatsheet

The `adapters/fake/` package now satisfies every Port with an
in-memory implementation, unblocking handler and Runner unit tests.
This file is the map for the *next* step: real adapters that talk to
the actual K8s API, Helm, NATS, HTTP registries, and the local
filesystem.

Each section names the port, the target adapter directory, and the
exact legacy source lines the adapter should absorb. Preserve
behaviour verbatim during the port; only refactor after
`Runner` + `Engine.Step` are wired end-to-end and the legacy files
have been deleted.

## Directory layout

```
adapters/
├── fake/              (done — see doc.go for design principles)
├── k8sstore/          real AMStore backed by controller-runtime client
├── helmops/           real WorkloadOps backed by pkg/appinstaller helm ops
├── imagemgr/          real ImagePuller backed by pkg/images
├── validator/         real InstallValidator backed by pkg/compute/validation
├── natsrecorder/      real EventRecorder backed by NATS JetStream
├── realclock/         real Clock backed by time.{Now,After,NewTicker}
├── mwops/             real MiddlewareOps (KubeBlocks + legacy MongoDB)
├── identity/          real Identity backed by kubesphere role bindings
├── computeops/        real ComputeOps backed by ConfigMap storage
└── marketrepo/        real MarketChartRepo backed by resty + go-getter
```

## Port → legacy source mapping

### `AMStore` → `adapters/k8sstore/`

Absorb the three drifting write paths into one adapter that uses
controller-runtime's `client.Patch` with `client.MergeFrom` for
`PatchStatus` / `PatchSpecAndAnnotations` and `client.Update` for
`UpdateFull`.

Legacy sources:

- `pkg/appstate/types.go:71-120` — `updateStatus` (Patch)
- `pkg/utils/app/app.go:167-229` — `UpdateAppMgrStatus` (Update)
- `pkg/appstate/pending_app.go` — raw `client.Update` for the
  download-concurrency gate. Move as-is into `UpdateFull`; retire
  once World.ClusterStats replaces the mutex (plan §1.5).

The DeepEqual short-circuit lives on the adapter, not the caller —
mirroring the current behaviour of `updateStatus`.

### `WorkloadOps` → `adapters/helmops/`

Thin wrapper over `pkg/appinstaller/helm_ops*.go`. Cast the
`WorkloadRequest.AppCfg` from `any` to `*appcfg.ApplicationConfig`
and `KubeConfig` to `*rest.Config` on entry. Each method maps 1:1 to
an existing helm-ops function.

Legacy sources (one per method):

| Port method                | Legacy call site                                              |
|----------------------------|---------------------------------------------------------------|
| `Install`                  | `pkg/appinstaller/helm_ops_install.go`                        |
| `Upgrade`                  | `pkg/appinstaller/helm_ops_install.go`                        |
| `ApplyEnv`                 | `pkg/appinstaller/helm_utils.go`                              |
| `Uninstall`                | `pkg/appinstaller/helm_ops_uninstall.go:289`                  |
| `UninstallAll`             | `pkg/appinstaller/helm_ops_uninstall.go:32`                   |
| `Stop`                     | `pkg/appinstaller/helm_ops.go` Scale(0)                       |
| `Resume`                   | `pkg/appinstaller/helm_ops.go` Scale(-1)                      |
| `WaitForStartUp`           | `pkg/appinstaller/helm_ops_install.go:626`                    |
| `WaitForLaunch`            | `pkg/appstate/initializing_app.go:82` inline poll             |
| `WaitForNamespaceDeleted`  | `pkg/appstate/types.go:267` (`wait.PollImmediate`)            |
| `GetDeployedVersion`       | `pkg/appstate/upgrading_app.go:163-169` (helm history query)  |

`WaitForNamespaceDeleted` is the one the whisper-service-v2 bug
report was about — it must observe `Namespace.DeletionTimestamp` on
an informer rather than polling by list, so the adapter should
subscribe to the namespace informer once and multiplex.

### `ImagePuller` → `adapters/imagemgr/`

Wrap `pkg/images/client.go` verbatim. Convert the raw IM CR into
`ports.ImageProgress` inside the adapter — handlers must not see
CR-level JSON.

Legacy sources:

- `pkg/images/client.go:64-91` — CreateJob (replace-in-place)
- `pkg/images/client.go:121` — 1s poll ticker (delete; Runner drives
  the poll via `Clock.NewTicker`)
- `pkg/appstate/downloading_app.go:133-139` — start pull
- `pkg/appstate/downloading_app.go:28` — progress read
- `pkg/appstate/upgrading_app.go:257-269` — upgrade pull
- `pkg/appstate/downloading_canceling_app.go:49` — CancelJob
- `pkg/appstate/upgrading_canceling_app.go:46` — CancelJob
- `pkg/appstate/downloading_canceling_failed_app.go:66` — status repair

### `InstallValidator` → `adapters/validator/`

The chain (cluster-capacity via kubesphere HTTP → user-quota →
k8s-request → compute-mode → node-pressure → compute-allocation) is
stitched inside `pkg/compute/validation`. The adapter selects the
chain by `ValidationOp` and returns a `ValidationDecision` — the
existing `Decision` struct is a byte-compatible copy.

Legacy sources:

- `pkg/appstate/installing_app.go:157` — `deps.RunInstallValidation`
- `pkg/appstate/installing_app.go:148` — `PublishComputeInsufficientNotification`
- `pkg/compute/validation/*` — the chain
- `pkg/compute/notification.go` — the NATS payload

### `EventRecorder` → `adapters/natsrecorder/`

Replace the process-global `pkg/event.AppEventQueue` singleton with
a per-Runner adapter that owns its own NATS connection.

Legacy sources:

- `pkg/event/event.go` — queue, retry ticker (delete both)
- `pkg/utils/event.go` — `Event` shape (copy field names into
  `AppLifecycleEvent`, already done in `ports/recorder.go`)
- `pkg/appstate/*_app.go` — every `deps.EventPublisher.Publish` call
  becomes a `domain.EmitEvent` returned by the handler

Preserve NATS subject naming (`os.application.{owner}` /
`os.notification`) — bfl consumers are pinned to it.

### `Clock` → `adapters/realclock/`

One-liner adapter: `time.Now()`, `time.After(d)`, `time.NewTicker(d)`
returning a struct that wraps `*time.Ticker`. No legacy migration.

### `MiddlewareOps` → `adapters/mwops/`

Two independent code paths, single adapter:

- KubeBlocks OpsRequest (`suspending_app.go:226-227` /
  `resuming_app.go:230-231`) — wraps `pkg/middlewareops`
- Legacy MongoDB (`types.go:392-398` +
  `uninstalling_app.go:179-181`) — wraps `pkg/middlewareinstaller`

Cast `cfg` to `*middlewareinstaller.MiddlewareConfig` on entry to
`UninstallLegacyMongo`.

### `Identity` → `adapters/identity/`

Wrap `kubesphere.GetAdminUsername` + `deps.IsAdmin` +
`apputils.GetClusterOwner`. All three are HTTP calls to kubesphere;
adapter should share a resty client with the validator adapter.

Legacy sources:

- `pkg/appstate/upgrading_app.go:186-191` — GetAdminUsername + IsAdmin
- `pkg/appstate/installing_canceling_app.go:108-112` — IsAdmin
- `pkg/utils/app/app.go` (search `GetClusterOwner`) — ClusterOwner

### `ComputeOps` → `adapters/computeops/` (done)

ConfigMap-backed allocation store. The `EnsureDeletedForTarget`
poll-inside-adapter contract is retained; Runner wraps the call in a
goroutine bounded by the CallWorkload timeout.

Wired into the Runner as an OPTIONAL port: only
`ManagesSharedServer` has a caller today (classifying an
install-validation rejection so the resulting stop gets the right
blast radius). The three delete/ensure methods exist and are
pass-throughs, but the handlers that need them — the *Failed and
*Canceling cleanup states — still leave allocation release to the
legacy path.

Legacy sources (nine files listed in `ports/compute.go`
doc-comment):

- `installing_app.go:193/278`, `installing_canceling_app.go:155`,
  `install_failure_cleanup.go:75`, `suspending_app.go:158`,
  `suspend_failed_app.go:81`, `stopped_app.go:68`,
  `uninstalling_app.go:198`, `uninstalled_app.go:90`,
  `types.go:240` (`forceDeleteApp`)

### `MarketChartRepo` → `adapters/marketrepo/`

External HTTP + filesystem side effects. Cast the return from
`GetAppConfig` to `*appcfg.ApplicationConfig` on entry into the
handler-side World materialiser.

Legacy sources:

- `pkg/appstate/upgrading_app.go:214-244` — dispatch by app type
- `pkg/utils/app.go:1173-1237` — resty + go-getter tar extract

## Non-goals for this cheatsheet

- No new tests in real adapters at first — they are covered by the
  existing envtest / integration suites (`pkg/appstate/*_test.go`
  and `controllers/*_test.go`). Once Runner + handler tests land, the
  real adapters can grow their own envtest cases as needed.
- No behaviour changes. Every real adapter is a mechanical port. If
  a legacy caller has a bug, keep it — file a follow-up issue rather
  than fixing it mid-migration; otherwise the "big-bang cutover" is
  no longer a bit-for-bit swap.

## Cutover gate

Real adapters go in only after:

1. `Runner` skeleton is in place and dispatches every Effect to the
   corresponding fake port (this makes the ports contract executable).
2. At least three handlers (Install, Uninstall, Upgrade) are ported
   to the new signature and pass under fake ports.

Otherwise the risk profile of touching helm / NATS / kubesphere
early is not worth the speed-up.
