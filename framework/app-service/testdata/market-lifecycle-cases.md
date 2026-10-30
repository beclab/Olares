# Market lifecycle test cases

Source: the market lifecycle plan, aligned with `controllers/applifecycle_runner_integration_test.go` (`It`, `DescribeTable`, `remainingTransitionCases`), `pkg/applifecycle/domain/table_default.go` `setTimeouts`, and the timeout reasons in `pkg/applifecycle/handlers/timeouts.go`.

Scope: 42 states and edges T01–T92, plus extra edges S1, S3, and S4. `install` on the empty state (T01) is not a 43rd state. Five names have no edge in this file and are not tested, and they are not substitutes for any row below:

- `initializingCanceled`
- `upgradingCanceled`
- `applyingEnvCanceled`
- `resumingCanceled`
- `failed`

These edges are also absent. Observe them if they happen, but do not count them in the 92:

- `cancel` on `running`, `stopped`, `stopping`, or `uninstalling`
- `install` on `installFailed`
- `stop` on `upgradeFailed`
- `uninstall` on `resumingCancelFailed`

Only one app may be in `downloading` at a time. Any other app stays in `pending` and waits for that slot. ApplicationManager has no status subresource. `status.state`, `status.statusTime`, and `status.payload` are written on the main object. `--subresource=status` returns 404. App names match `^[a-z0-9]{1,30}$`. The chart `name`, the OlaresManifest `metadata.name`, and the Deployment name are the same. `Chart.yaml` `version` and OlaresManifest `metadata.version` change together.

## Methods

| Method | Meaning |
|---|---|
| natural | Drive the app with a real install, stop, cancel, resume, upgrade, or env change. |
| backdate | Move `statusTime` into the past on the main object. app-service keeps running, and the next reconcile treats the state as timed out. |
| fixture | Scale `app-service` to 0, write the state and an already-expired `statusTime`, then scale it back. The first reconcile takes that edge. |
| chart | Use one of the charts below. |

A state that leaves on the next tick cannot be backdated in place. `initializingCanceling`, `resumingCanceling`, `applyingEnvCanceling`, `upgradingCanceling`, and `upgradePendingCanceling` hand off on an empty tick. Their timeout edges (T52–T56) use a fixture. Edges that can sit still (image pull, helm, a namespace finalizer, a Pending pod blocking scale-down) use a chart or `test.olares/hold`, then backdate or wait for the shortened test-image timeout. Remove the finalizer after the assertion. If `pending` has not created a namespace yet, create `spec.appNamespace` with the finalizer first, or the wait-ns step completes immediately.

Test-image timeouts, when a shortened image is in use. Repository backdates (1 minute, 11 minutes, 31 minutes, 61 minutes, 2 hours, 25 hours) stay longer than these values.

| State | Repository | Test image |
|---|---|---|
| `downloading`, `upgradeDownloading` | 24h | 24h, unchanged. Backdate those timeout edges by 25h. |
| `installing` | 30m | 5m |
| `initializing` | 1h | 8m |
| `upgrading` | 1h | 5m |
| `applyingEnv` | 30m | 3m |
| `resuming` | 1h | 5m |
| `stopping` | 30m | 3m |
| `uninstalling` | 30m | 3m |
| nine `*Canceling` states | 10m | 90s |

`initializing` is 8 minutes because `HardErrorGrace` in `pkg/podhealth/podhealth.go` is 5 minutes. A shorter window reaches `initializingCanceling` before `InitFailed`. A normal `fault-good` install must finish inside the new `installing` and `initializing` windows. If it does not, raise only those two numbers. Leave the download timeout at 24h.

## Charts

Charts live in `testdata/fault-charts/`. The directory name can differ from the install name. `Chart.yaml` `name` is the install name. Copy the shape of `fault-good`: `Chart.yaml`, `OlaresManifest.yaml`, `values.yaml`, `templates/app.yaml`. `workloadReplicas` keys, the Deployment name, and the Service name follow the app name. Package with `olares-cli chart package`, upload with `olares-cli market upload`, and install with `-s upload`. Do not commit the tgz.

| Directory | Install name | Used for |
|---|---|---|
| `fault-good` | `faultgood` 0.1.0 | Reach `running`. `BREAK=1` makes the template `fail` and hits `applyEnvFailed`. |
| `fault-good-upgrade` | `faultgood` 0.2.0 | Unknown kind `NoSuch`. Upload render succeeds; the upgrade apply fails. `upgrading` to `upgradeFailed`. |
| `fault-pull` | `faultpull` | Missing image. `downloadFailed`. |
| `fault-install` | `faultinstall` | `NoSuch` CR. Helm fails, then `installingCanceling` to `installFailed`. |
| `holdinst` | `holdinst` | `nodeSelector` matches nothing. Park `installing` on scheduling. |
| `stopinit` | `stopinit` | Container exits. Stop reason `InitFailed`. |
| `stopdisk` / `stopmem` / `stopk8scpu` / `stopcpu` | same name | Stop reasons after resource validation. |
| `upgradetest` | `upgradetest` | Version comparison. |
| `ginit` / `gpull` / `gcont` / `applyenv` | same name | Hold an in-progress state. |
| `fault-unready` | `faultunready` | Pod mounts a Secret that does not exist and stays `Pending`. After 1 minute, T35: `installing` wait-startup returns `ErrPodPending` and enters `stopping` with reason `Unschedulable`. |
| `fault-hanginit` | `faulthanginit` | Container `sleep 3600`, no listen port, no failing probe. T46: sit in `initializing` until timeout. If it becomes `InitFailed` first, record the actual edge. |
| `fault-slowhelm` | `faultslowhelm` | pre-install / pre-upgrade hook `sleep 2100` (35 minutes, longer than the repository 30-minute installing timeout). T45. |
| `fault-good-missimg` | `faultgood` 0.2.0 | Missing image, template still renders. Same 0.2.0 as `fault-good-upgrade`; upload only one of them at a time. T38. |
| `fault-delete-hook` | `faultdelhook` | pre-delete hook `exit 1`. T42, and the cleanup uninstall failure in T36. |
| `fault-mw` | `faultmw` | `olaresManifest.type: middleware`. T03. If upload rejects middleware, set `spec.type` to `middleware` once `installing` phase is `wait-launch`. |
| `fault-gpu` | `faultgpu` | Manifest `requiredGpu: 1Gi`. The pod requests `nvidia.com/gpu: 64` and sets `gpu.bytetrade.io/app-pod-consume-policy: single`, so HAMi records `InsufficientGPU` when fewer GPUs are bound than requested. Deployment annotation `applications.app.bytetrade.io/gpu-inject: "false"` so the webhook does not rewrite that request down to one card. Stop reason is `HamiUnschedulable` in the stop-reason table. |

## States (42)

Every row must be entered, and must leave through the exits listed here. The entry column is the first T id that reaches the state. The exit column is every exit in this file.

| State | Entry | Exits |
|---|---|---|
| `pending` | T01 | T02 `downloading`; T12 `pendingCanceling` |
| `downloading` | T02 | T02 `installing`; T13, T57 `downloadingCanceling`; T33 `downloadFailed` |
| `installing` | T02 | T02 `initializing`; T03 `running`; T14, T34, T45 `installingCanceling`; T35 `stopping`; T36 `installFailed`; T44 backdate 1 minute and stay in `installing` |
| `initializing` | T02 | T02, T06 `running`; T15, T46 `initializingCanceling` |
| `running` | T02 | T04 `stopping`; T05 `uninstalling`; T18 `upgradePending`; T10 `applyingEnv`; a healthy tick stays `running` |
| `stopping` | T04 | T04 `stopped`; T41, T47 `stopFailed` |
| `stopped` | T04 | T06 `resuming`; T07 `uninstalling`; T08 `upgradePending`; T09 `applyingEnv` |
| `resuming` | T06 | T06 `initializing`; T16, T60 `resumingCanceling`; T32 `stopping`; T40 `resumeFailed` |
| `uninstalling` | T05 | T05 `uninstalled`; T42, T48 `uninstallFailed` |
| `uninstalled` | T05 | T11 `pending` |
| `upgradePending` | T18 | T18 `upgradeDownloading`; T22 `upgradePendingCanceling` |
| `upgradeDownloading` | T18 | T19 `upgrading`; T25, T58 `upgradeDownloadingCanceling`; T38 `upgradeFailed` |
| `upgrading` | T19 | T20 `initializing`; T21 `stopped`; T27, T61 `upgradingCanceling`; T37 `upgradeFailed` |
| `applyingEnv` | T10 | T17, T59 `applyingEnvCanceling`; T29 `initializing`; T30 `stopped`; T31 `stopping`; T39 `applyEnvFailed` |
| `pendingCanceling` | T12 | T12 `pendingCanceled`; T50 `pendingCancelFailed` |
| `downloadingCanceling` | T13 | T13 `downloadingCanceled`; T51 `downloadingCancelFailed` |
| `installingCanceling` | T14 | T14 `installingCanceled`; T34, T45 `installFailed`; T49 `installingCancelFailed` |
| `initializingCanceling` | T15 | T15, T46 `stopping`; T52 `initializingCancelFailed` |
| `resumingCanceling` | T16 | T16 `stopping`; T53 `resumingCancelFailed` |
| `applyingEnvCanceling` | T17 | T17 `stopping`; T54 `applyingEnvCancelFailed` |
| `upgradePendingCanceling` | T22 | T23 `running`; T24 `stopped`; T55 `upgradePendingCancelFailed` |
| `upgradeDownloadingCanceling` | T25 | T25 `running`; T26 `stopped`; T43 `upgradeDownloadingCancelFailed`; S1 timeout also to `upgradeDownloadingCancelFailed` |
| `upgradingCanceling` | T27 | T28 `upgradeFailed`; T56 `upgradingCancelFailed` |
| `pendingCanceled` | T12 | T77 `pending` |
| `downloadingCanceled` | T13 | T78 `pending` |
| `installingCanceled` | T14 | T79 `pending` |
| `downloadFailed` | T33 | T65 `pending` |
| `installFailed` | T34 | T62 `uninstalled`; T64 `uninstalling` |
| `stopFailed` | T47 | T67 `stopping`; T68 `uninstalling` |
| `uninstallFailed` | T48 | T66 `uninstalling` |
| `resumeFailed` | T40 | T69 `resuming`; T70 `applyingEnv`; T71 `uninstalling` |
| `upgradeFailed` | T37 | T72 `upgradePending`; T73 `uninstalling` |
| `applyEnvFailed` | T39 | T74 `applyingEnv`; T75 `stopping`; T76 `uninstalling` |
| `pendingCancelFailed` | T50 | T80 `pendingCanceling` |
| `downloadingCancelFailed` | T51 | T81 `downloadingCanceling` |
| `installingCancelFailed` | T49 | T82 `installingCanceling`; T83 `uninstalling` |
| `initializingCancelFailed` | T52 | T84 `initializingCanceling`; T85 `uninstalling` |
| `resumingCancelFailed` | T53 | T63 empty tick stays; T90 `resumingCanceling` |
| `applyingEnvCancelFailed` | T54 | T88 `applyingEnvCanceling`; T89 `uninstalling` |
| `upgradePendingCancelFailed` | T55 | T91 `upgradePendingCanceling` |
| `upgradeDownloadingCancelFailed` | T43 | T92 `upgradeDownloadingCanceling` |
| `upgradingCancelFailed` | T56 | T86 `upgradingCanceling`; T87 `uninstalling` |

## Edges

`Test name` is the `It` description or the `name` in `remainingTransitionCases`. A named test that spans several hops is split into the exits in this table. Intermediate states must appear in the log. `natural` means a CLI or spec operation. A chart name means that package. `backdate` changes `statusTime` on the main object. `fixture` scales app-service to 0, writes the state, and scales it back.

### Install path

| ID | Test name | From | Trigger | To | How |
|---|---|---|---|---|---|
| T01 | `empty-install` | no AM | `install` | `pending` | natural. `market install` creates the AM. |
| T02 | `installs Pending through Running` | `pending` | tick through pull, expose-ports, validate, helm, scale-up, wait-startup, launch | `downloading` to `installing` to `initializing` to `running`; a further tick stays `running` | `fault-good`. Middleware may go `installing` straight to `running`; that edge is T03, not a substitute for T02. |
| T03 | `mw-running` | `installing`, phase `wait-launch`, type `middleware` | WorkloadReady | `running` | `fault-mw` |
| T44 | `keeps a fresh Installing state inside its deadline` | `installing` | tick after backdating `statusTime` by 1 minute | still `installing` | Park `holdinst` in `installing`, then backdate. With a 5-minute installing timeout, 1 minute is still inside the deadline. |

### Steady operations

| ID | Test name | From | Trigger | To | How |
|---|---|---|---|---|---|
| T04 | `stops a Running app` | `running` | `stop` | `stopping` to `stopped` | natural |
| T05 | `uninstalls a Running app` | `running` | `uninstall` | `uninstalling` to `uninstalled`, namespace gone | natural |
| T06 | `resumes Stopped through Initializing to Running` | `stopped` | `resume` | `resuming` to `initializing` to `running` | natural. At least one `WorkloadReplicas` entry. |
| T07 | `stopped-uninst` | `stopped` | `uninstall` | `uninstalling` | natural |
| T08 | `stopped-upg` | `stopped` | `upgrade` | `upgradePending` | Upload a higher version, then `upgrade -s upload`. |
| T09 | `stopped-aenv` | `stopped` | `applyEnv` | `applyingEnv` | `settings apps env set` |
| T10 | `run-applyenv` | `running` | `applyEnv` | `applyingEnv` | same as T09 |
| T11 | `uninstalled install` | `uninstalled` | `install` | `pending` | natural |

### Cancel during install

After the terminal state, `install` must be able to enter `pending` again (T77–T79). Another `cancel` during `*Canceling` continues the same cancel.

| ID | Test name | From | Trigger | To | How |
|---|---|---|---|---|---|
| T12 | `cancels Pending` | `pending` | `cancel`, then wait-ns completes | `pendingCanceling` to `pendingCanceled`, namespace deleted | Another app holds `downloading`, so this app stays in `pending`. |
| T13 | `cancels an in-progress download` | `downloading` | `cancel`, cancel the pull, then wait-ns | `downloadingCanceling` to `downloadingCanceled`, namespace deleted. A late pull callback after the terminal state must not move the state. | `fault-pull` or a large image. Cancel while the pull is in progress. |
| T14 | `cancels an in-progress helm install` | `installing` | `cancel`, uninstall the release, then wait-ns | `installingCanceling` to `installingCanceled`, namespace deleted | `fault-slowhelm` or `holdinst`. Cancel after helm has started. |

### Cancel that hands off to stop

The terminal state is `stopped`. The namespace remains, and a later `resume` can return to `running`. There is no `initializingCanceled`, `resumingCanceled`, or `applyingEnvCanceled`.

| ID | Test name | From | Trigger | To | How |
|---|---|---|---|---|---|
| T15 | `cancels Initializing mid-launch` | `initializing` | `cancel` | `initializingCanceling` to `stopping` to `stopped` | A fresh install's `initializing`, not only the one entered by resume. |
| T16 | `cancels Resuming mid-scale` | `resuming` | `cancel` | `resumingCanceling` to `stopping` to `stopped` | Stop `fault-good`, resume, and cancel before scale finishes. |
| T17 | `cancels ApplyingEnv mid-helm` | `applyingEnv` | `cancel` | `applyingEnvCanceling` to `stopping` to `stopped` | Change `BREAK` on `fault-good` and cancel before helm returns. |

### Upgrade

Prepare the upgrade package: upload and install the lower version to `running` or `stopped`, raise both `Chart.yaml` `version` and OlaresManifest `metadata.version`, upload, then `upgrade -s upload --version <higher>`. If the catalog has only one version, `upgrade` of that same version is rejected by the CLI. An app installed from `market.olares` returns HTTP 404 on `upgrade -s upload`. Both install and upgrade use `-s upload`.

Upgrade from `running` and upgrade from `stopped` are different edges. When `bytetrade.io/pre-upgrade-state=stopped`, canceling the queue or the image pull returns to `stopped`, and a successful helm phase also returns to `stopped`.

| ID | Test name | From | Trigger | To | How |
|---|---|---|---|---|---|
| T18 | first half of `cancels an upgrade download` | `running` | `upgrade` | `upgradePending` to `upgradeDownloading` | Stays in `upgradePending` while another app holds the download slot. |
| T19 | `upgdl-upg` | `upgradeDownloading`, phase `pull` | WorkloadReady | `upgrading` | natural, after the image pull finishes |
| T20 | `upg-init` | `upgrading`, phase `wait-startup`, target and chart annotations set | WorkloadReady | `initializing` | Successful upgrade from `running`. It should then reach `running`, and the chart annotation becomes the target version. |
| T21 | `upg-stopped` | `upgrading`, phase `helm`, `pre-upgrade-state=stopped`, target and chart annotations set | WorkloadReady | `stopped` | Successful upgrade from `stopped`. Chart is the target version, temporary annotations are cleared, and the operation record carries the new version. S4 is this same landing after a failed attempt: the retry must not replace `pre-upgrade-state=stopped` with `upgradeFailed`. |
| T22 | `upgpend-cancel` | `upgradePending` | `cancel` | `upgradePendingCanceling` | Cancel while the download slot is held. |
| T23 | `upgpendc-run` | `upgradePendingCanceling` | tick | `running` | Natural next tick of T22. Reason `UpgradePendingCanceled`. Chart unchanged, target and pre cleared, config restored to the pre-upgrade value. |
| T24 | `upgpendc-stop` | `upgradePendingCanceling`, annotation `pre-upgrade-state=stopped` | tick | `stopped` | Upgrade started from `stopped`, canceled in `upgradePending`. |
| T25 | `cancels an upgrade download mid-pull` | `upgradeDownloading` | `cancel`, pull cancel succeeds | `upgradeDownloadingCanceling` to `running` | Cancel during the image pull. Version fields match T23. |
| T26 | `upgdlc-stop` | `upgradeDownloadingCanceling`, phase `cancel-pull`, annotation `stopped` | WorkloadReady | `stopped` | Upgrade from `stopped`, canceled in `upgradeDownloading`. |
| T27 | `upg-cancel` | `upgrading` | `cancel` | `upgradingCanceling` | Cancel after helm has started (log contains starting helm upgrade). |
| T28 | `upgc-failed` | `upgradingCanceling` | tick | `upgradeFailed` | Natural next tick of T27. Does not enter `stopping`. Chart stays the deployed version, target and pre remain, config stays the target version. See the version table. |

### Apply-env success and stop

| ID | Test name | From | Trigger | To | How |
|---|---|---|---|---|---|
| T29 | `aenv-init` | `applyingEnv`, phase `helm` | WorkloadReady | `initializing` | `fault-good` on `running` with `BREAK=0` or another harmless change. Success should then reach `running`. |
| T30 | `aenv-stopped` | `applyingEnv`, phase `helm`, annotation `pre-upgrade-state=stopped` | WorkloadReady | `stopped` | On `stopped`, change an env var that renders successfully. S3 also checks the annotation is `stopped` and the Deployment stays at 0 replicas, then a later apply from `running` overwrites that value and stays `running`. |
| T31 | `aenv-stop` | `applyingEnv` | `stop` | `stopping` | `stop` during apply-env, not cancel. |
| T32 | `res-stop` | `resuming` | `stop` | `stopping` | `stop` during resume. |

### Workload failures

| ID | Test name | From | Trigger | To | How |
|---|---|---|---|---|---|
| T33 | `fails an in-progress download` | `downloading` | image pull fails | `downloadFailed` | `fault-pull`. Reason is `ImagePullFailed` when a workload reason is present. |
| T34 | `fails helm install` | `installing` | helm error | `installingCanceling` to `installFailed`, namespace deleted | `fault-install`. The next tick after cleanup completes is T62. |
| T35 | `inst-stopping` | `installing`, phase `wait-startup` | pod stays `Pending` for more than 1 minute | `stopping` | `fault-unready`. `WaitForStartUp` returns `ErrPodPending` and the reason is `Unschedulable`. A startup error that is not pending still enters `installingCanceling` with reason `StartUpFailed`. |
| T36 | `inst-failed` | `installing`, phase `cleanup` | WorkloadFailed | `installFailed` | `fault-delete-hook`. Write phase `cleanup` so the cleanup uninstall fails. The current install-failure path is T34; this edge is still required because the integration test has it. |
| T37 | `upg-failed` | `upgrading` | WorkloadFailed | `upgradeFailed` | `fault-good-upgrade`. The template renders, then Helm rejects kind `NoSuch`. A `fail` action is rejected at upload and never reaches `upgrading`. |
| T38 | `upgdl-failed` | `upgradeDownloading` | WorkloadFailed | `upgradeFailed` | `fault-good-missimg`. The failure is the image pull, not helm render. |
| T39 | `aenv-failed` | `applyingEnv` | WorkloadFailed | `applyEnvFailed` | `fault-good`, `BREAK=1`. |
| T40 | `res-failed` | `resuming` | WorkloadFailed | `resumeFailed` | Stop to `stopped`, delete the Helm release secret in that namespace, then `resume`. Reason `HelmResumeFailed`. |
| T41 | `stop-failed-wl` | `stopping` | WorkloadFailed | `stopFailed` | Separate from the T47 timeout. During the `stopping` scale phase, delete the Helm release secret or make scale return an error. Reason is a workload reason such as `HelmStopFailed`, not `StopTimedOut`. |
| T42 | `uninst-failed-wl` | `uninstalling` | WorkloadFailed | `uninstallFailed` | `fault-delete-hook`. Reason is a workload reason such as `HelmUninstallFailed`, not `UninstallTimedOut`. |
| T43 | `upgdlc-fail` | `upgradeDownloadingCanceling` | WorkloadFailed | `upgradeDownloadingCancelFailed` | While canceling an upgrade pull, write the ImageManager status as `failed`. |

### Timeouts

Backdate by the repository timeout. After the test image shortens a timeout, backdating "new timeout plus 1 minute" is also valid. The repository backdate is listed so it can be compared with the integration test.

| ID | Test name | From | Repository backdate | To | How |
|---|---|---|---|---|---|
| T45 | `times out a parked helm install` | `installing`, helm still held | 31m | `installingCanceling`, payload terminal `installFailed`, then `installFailed`, namespace deleted | `fault-slowhelm`, or backdate while helm is in progress. |
| T46 | `times out a parked launch` | `initializing`, launch still held | 2h | `initializingCanceling` to `stopping` to `stopped` | `fault-hanginit`. Reason is `initializeCancelBySystem`, not `InitFailed`. |
| T47 | `times out Stopping` | `stopping` | 31m | `stopFailed`, reason `StopTimedOut` | Hold scale-down with a namespace finalizer, then backdate. |
| T48 | `times out Uninstalling` | `uninstalling` | 31m | `uninstallFailed`, reason `UninstallTimedOut` | Same, blocked in `wait-ns`. |
| T49 | `times out InstallingCanceling` | `installingCanceling`, phase `wait-ns`, terminal `installFailed` | 11m | `installingCancelFailed`. The payload terminal stays `installFailed`. | Finalizer blocks wait-ns. Reason `InstallingCancelTimedOut`. |
| T50 | `pendc-fail` | `pendingCanceling` | 11m | `pendingCancelFailed` | Create the namespace with a finalizer first, or it goes straight to `pendingCanceled`. Reason `PendingCancelTimedOut`. |
| T51 | `dlc-fail` | `downloadingCanceling` | 11m | `downloadingCancelFailed` | Finalizer plus backdate. Reason `DownloadingCancelTimedOut`. |
| T52 | `initc-fail` | `initializingCanceling` | 11m | `initializingCancelFailed` | Fixture. An empty tick goes to `stopping` before a backdate can win. Reason `InitializingCancelTimedOut`. |
| T53 | `resc-fail` | `resumingCanceling` | 11m | `resumingCancelFailed` | Fixture. Reason `ResumingCancelTimedOut`. |
| T54 | `aenvc-fail` | `applyingEnvCanceling` | 11m | `applyingEnvCancelFailed` | Fixture. Reason `ApplyingEnvCancelTimedOut`. |
| T55 | `upgpendc-fail` | `upgradePendingCanceling` | 11m | `upgradePendingCancelFailed` | Fixture. The natural next tick is T23 or T24, which beats a backdate. Reason `UpgradePendingCancelTimedOut`. |
| T56 | `upgc-fail` | `upgradingCanceling` | 11m | `upgradingCancelFailed` | Fixture. The empty tick is T28 `upgradeFailed`. Reason `UpgradingCancelTimedOut`. |
| T57 | `dl-timeout` | `downloading` | 25h | `downloadingCanceling` | Backdate 25h during a real pull. This is not `downloadFailed`. Reason `downloadCancelBySystem`. |
| T58 | `upgdl-timeout` | `upgradeDownloading` | 25h | `upgradeDownloadingCanceling` | Same. Reason `upgradeCancelBySystem`. |
| T59 | `aenv-timeout` | `applyingEnv` | 31m | `applyingEnvCanceling` | Backdate during helm, or wait for the test image's 3 minutes. Reason `applyEnvCancelBySystem`. |
| T60 | `res-timeout` | `resuming` | 61m | `resumingCanceling` | Backdate during scale, or wait for the test image's 5 minutes. Reason `resumeCancelBySystem`. |
| T61 | `upg-timeout` | `upgrading` | 61m | `upgradingCanceling` | Backdate during helm, or wait for the test image's 5 minutes. Reason `upgradeCancelBySystem`. |

### Install-failure cleanup and a cancel-failed state that stays

| ID | Test name | From | Trigger | To | How |
|---|---|---|---|---|---|
| T62 | `instfail-done` | `installFailed`, payload `install-failed.cleanup=done` | tick | `uninstalled` | Natural next tick after T34 cleanup, reason `InstallFailureCleanedUp`. Also: if cleanup is not done yet, backdate `statusTime` by 11 minutes, reason `InstallCleanupTimedOut`, and the state is still `uninstalled`. That second path is handler behavior. The integration-test edge only asserts cleanup done. |
| T63 | `keeps ResumingCancelFailed terminal on Tick` | `resumingCancelFailed` | empty tick | still `resumingCancelFailed` | After T53, send nothing and reconcile once more. |

### Next operation from a terminal state

The integration test sends the operation directly from that state. On the cluster, enter the state with an earlier edge, then send the operation. If the state cannot be entered that way, write it with a fixture and mark the record as a fixture entry.

| ID | Test name | From | Operation | To |
|---|---|---|---|---|
| T64 | `installfailed uninstall` | `installFailed` | `uninstall` | `uninstalling` |
| T65 | `downloadfailed install` | `downloadFailed` | `install` | `pending` |
| T66 | `uninstallfailed uninstall` | `uninstallFailed` | `uninstall` | `uninstalling` |
| T67 | `stopfailed stop` | `stopFailed` | `stop` | `stopping` |
| T68 | `stopfailed uninstall` | `stopFailed` | `uninstall` | `uninstalling` |
| T69 | `resumefailed resume` | `resumeFailed` | `resume` | `resuming` |
| T70 | `resumefailed applyenv` | `resumeFailed` | `applyEnv` | `applyingEnv` |
| T71 | `resumefailed uninstall` | `resumeFailed` | `uninstall` | `uninstalling` |
| T72 | `upgradefailed upgrade` | `upgradeFailed` | `upgrade` | `upgradePending` |
| T73 | `upgradefailed uninstall` | `upgradeFailed` | `uninstall` | `uninstalling` |
| T74 | `applyenvfailed applyenv` | `applyEnvFailed` | `applyEnv` | `applyingEnv` |
| T75 | `applyenvfailed stop` | `applyEnvFailed` | `stop` | `stopping` |
| T76 | `applyenvfailed uninstall` | `applyEnvFailed` | `uninstall` | `uninstalling` |
| T77 | `pendingcanceled install` | `pendingCanceled` | `install` | `pending` |
| T78 | `downloadingcanceled install` | `downloadingCanceled` | `install` | `pending` |
| T79 | `installingcanceled install` | `installingCanceled` | `install` | `pending` |
| T80 | `pendingcancelfailed cancel` | `pendingCancelFailed` | `cancel` | `pendingCanceling` |
| T81 | `downloadingcancelfailed cancel` | `downloadingCancelFailed` | `cancel` | `downloadingCanceling` |
| T82 | `installingcancelfailed cancel` | `installingCancelFailed` | `cancel` | `installingCanceling` |
| T83 | `installingcancelfailed uninstall` | `installingCancelFailed` | `uninstall` | `uninstalling` |
| T84 | `initializingcancelfailed cancel` | `initializingCancelFailed` | `cancel` | `initializingCanceling` |
| T85 | `initializingcancelfailed uninstall` | `initializingCancelFailed` | `uninstall` | `uninstalling` |
| T86 | `upgradingcancelfailed cancel` | `upgradingCancelFailed` | `cancel` | `upgradingCanceling` |
| T87 | `upgradingcancelfailed uninstall` | `upgradingCancelFailed` | `uninstall` | `uninstalling` |
| T88 | `applyingenvcancelfailed cancel` | `applyingEnvCancelFailed` | `cancel` | `applyingEnvCanceling` |
| T89 | `applyingenvcancelfailed uninstall` | `applyingEnvCancelFailed` | `uninstall` | `uninstalling` |
| T90 | `resumingcancelfailed cancel` | `resumingCancelFailed` | `cancel` | `resumingCanceling` |
| T91 | `upgpendcf-cancel` | `upgradePendingCancelFailed` | `cancel` | `upgradePendingCanceling` |
| T92 | `upgdlcf-cancel` | `upgradeDownloadingCancelFailed` | `cancel` | `upgradeDownloadingCanceling` |

The reason for another `cancel` on T80–T92 is `CancelRetryRequested`. `uninstall` on a cancel-failed state enters `uninstalling`. `resumingCancelFailed` has cancel in this file and does not have uninstall.

### Extra edges

| ID | From | Trigger | To | Notes |
|---|---|---|---|---|
| S1 | `upgradeDownloadingCanceling` | timeout (repository 10m, test image 90s) | `upgradeDownloadingCancelFailed` | Present in `timeouts.go`. The integration test reaches the same state with T43 WorkloadFailed and has no age case. Backdate `statusTime` while the pull cancel is still in progress. Reason `UpgradeDownloadingCancelTimedOut`. |
| S3 | `stopped`, then `running` | `applyEnv` | `stopped`, then `running` | From `stopped`, set `BREAK=2` (the chart default `0` does not mark the env dirty, and `BREAK=1` fails render). The state must leave `stopped` and land back in `stopped`, with `bytetrade.io/pre-upgrade-state=stopped` and replicas 0. Resume to `running`, set `BREAK=0`, and that apply must leave `running` and land back in `running` with the annotation overwritten to `running`. |
| S4 | `stopped`, then `upgradeFailed` | `upgrade`, then `upgrade` again | `upgradeFailed`, then `stopped` | Upgrade `upgradetest` from `stopped` with the same `NoSuch` kind as `fault-good-upgrade`. Upload render succeeds; the upgrade apply fails. The failed attempt keeps `pre-upgrade-state=stopped`. Replace it with a working chart and upgrade again. The retry still reads `stopped`, lands in `stopped`, and leaves replicas at 0. |

### Installing phase failures

A change of phase inside `installing` does not change the state name and is not an extra state. The integration test already lists helm failure (T34), wait-startup failure (T35), cleanup failure (T36), and helm timeout (T45). The phases below land in states that are already in the 42. Each one still has to be produced once. Record the handler's terminal state. If it cannot be produced, record it as not tested. Do not substitute another phase.

| Phase | payload `installing.phase` | State after failure | How |
|---|---|---|---|
| expose-ports | `expose-ports` | `installingCanceling`, terminal `installFailed` | Record not tested when there is no natural fault. |
| validate | `validate` | `installingCanceling`, terminal `installFailed` | A requirement rejected at submit time for exceeding cluster capacity never enters the lifecycle, so it does not count. |
| helm-install | `helm-install` | usually T34; `stopping` when the adapter treats it as unschedulable | `fault-install`; unschedulable uses `holdinst` |
| post-validate | `post-validate` | `stopping` | `stopdisk`, `stopmem`, `stopk8scpu`, and the same family. Reason matches the stop-reason table. |
| scale-up | `scale-up` | `installingCanceling`, terminal `installFailed` | Record not tested when there is no natural fault. |
| wait-startup | `wait-startup` | `stopping` when the pod is still `Pending` after 1 minute, or immediately when it is `Unschedulable` (`ErrPodPending` or `ErrServerSidePodPending`; `HamiUnschedulable` only changes the stop reason). `installingCanceling`, reason `StartUpFailed`, only when the error is not a pending error. | T35, `fault-unready`. Also `gcont` and `faultgpu`. |
| wait-launch | `wait-launch` | success is T03 `running`. Failure enters `installingCanceling` without an installFailed terminal, and cleanup lands on the user-cancel terminal `installingCanceled`. | Failure uses a middleware chart that cannot start. If middleware cannot be uploaded, record not tested. |
| cleanup | `cleanup` | `installFailed` | T36 |

Upgrade phases are the same idea: `prepare`, `pull`, `helm`, `wait-startup`. `pull` failure is T38, helm render failure is T37, and `wait-startup` failure also enters `upgradeFailed` (`upg-failed` does not look at phase). All three must show `upgradeFailed`, and the record must include the phase at failure.

## Upgrade-cancel version comparison

Capture versions for T20, T21, T23, T24, T25, T26, T28, T37, T38, and T72. The example upgrades `0.1.0` to `0.2.0`.

| Case | Terminal state | chart | target | pre | config | Helm |
|---|---|---|---|---|---|---|
| before upgrade | `running` | 0.1.0 | none | none | 0.1.0 | 0.1.0 `deployed` |
| T23 / T25 cancel queue or pull from running | `running` | 0.1.0 | none | none | 0.1.0 | no new revision |
| T24 / T26 cancel from stopped | `stopped` | 0.1.0 | none | none | 0.1.0 | no new revision |
| T28 cancel during helm | `upgradeFailed` | 0.1.0 | 0.2.0 | 0.1.0 | 0.2.0 | latest 0.2.0 `failed`, 0.1.0 still `deployed` |
| T37 / T38 upgrade failed | `upgradeFailed` | 0.1.0 | 0.2.0 | 0.1.0 | 0.2.0 | latest 0.2.0 `failed`; a template render failure or an image that never pulled may have no new revision |
| T20 upgrade from running succeeds | `initializing` to `running` | 0.2.0 | none | none | 0.2.0 | 0.2.0 `deployed` |
| T21 upgrade from stopped succeeds | `stopped` | 0.2.0 | none | none | 0.2.0 | 0.2.0 `deployed`, latest record is 0.2.0 |
| T56 cancel timed out | `upgradingCancelFailed` | 0.1.0 | 0.2.0 | 0.1.0 | 0.2.0 | record whatever helm revision exists at cancel time; it does not have to match T28 |

Rules:

- `chart` becomes 0.2.0 only on T20 and T21. After T23–T26, T28, T37, T38, and T56 it is still 0.1.0.
- T23–T26 restore config from the snapshot. `config` returns to 0.1.0, and `target` and `pre` are cleared.
- T28, T37, and T38 do not restore. `config` stays 0.2.0, and both `target` and `pre` remain.
- After T28, a latest Helm revision stuck in `pending-upgrade` is a failure. The next upgrade then reports `another operation (install/upgrade/rollback) is in progress`.
- T72: upgrading again from `upgradeFailed` to a working 0.2.0 package (replace the T37 package with one that works) must enter `upgradePending`, Helm must not report operation in progress, and success must match the T20 row when the failed upgrade started from `running`.
- S4: the same retry after a failed upgrade that started from `stopped` must match the T21 row. `bytetrade.io/pre-upgrade-state` stays `stopped` when the retry is accepted. Success stays `stopped` with replicas 0. Writing `upgradeFailed` into that annotation and then scaling the workload back up is a failure.
- If the market `version` and `chart` disagree, record it. That disagreement alone is not a failure.

`uninstall` on `upgrading` first cancels to `upgradeFailed` (T27, T28), then uninstalls from `upgradeFailed` (T73) to `uninstalled`. `stop` on `upgradeFailed` is rejected. That is a side assertion, not an edge in this file.

Version fields:

- An upgrade request writes only `bytetrade.io/target-chart-version` and snapshots the pre-upgrade config in `bytetrade.io/pre-upgrade-config`. `bytetrade.io/chart-version` stays the deployed version.
- The target version is committed, and those two temporary annotations are cleared, only when success enters `initializing` (T20) or `stopped` (T21). The path that started from `stopped` records an operation with the new version.
- Canceling `upgradePending` or `upgradeDownloading` returns to `running` or `stopped` (T23–T26), restores `Spec.Config` from the snapshot, and clears the target version and the snapshot annotation.
- `Scale` and `ApplyEnv` use the chart on the current Helm release. They do not read the chart from disk.

## State-change reasons

`applyEmitEvent` emits `state change` only when `status.state` changes. A phase update inside the same state does not. The log must not contain `DownloadStarted`, `HelmInstallStarted`, `InstallWaitStartUpStarted`, `InitializingStarted`, `UpgradeStarted`, `CancelPullStarted`, `CancelWaitNSStarted`, or `StopStarted`.

Download is the exception: a progress percentage change inside `downloading` or `upgradeDownloading` still publishes NATS with an empty `Reason` and does not emit `state change`. The progress bar still moving is enough.

The line format is `state change <AM name> <from> -> <to> reason=<Reason>`. A jump written straight to status by the controller (for example eviction) may have no such line. The next jump the runner takes itself must have one.

| Change | Reason |
|---|---|
| install accepted, enter `pending` | `InstallRequested` |
| `pending` to `downloading` | `Downloading` |
| `downloading` to `installing` | `DownloadCompleted` |
| `downloading` to `downloadFailed` | `ImagePullFailed` (or `downloadFailed` when there is no workload reason) |
| `installing` to `initializing` | `InstallSucceeded` |
| `installing` to `installFailed` | `HelmInstallFailed` |
| `initializing` to `running` | `Running` |
| stop accepted, enter `stopping` | `StopRequested` |
| user stop lands on `stopped` | `StopByUser` |
| resume accepted, enter `resuming` | `ResumeRequested` |
| `resuming` to `initializing` | `ResumeHelmDone` |
| `resuming` to `resumeFailed` | `HelmResumeFailed` |
| upgrade accepted, enter `upgradePending` | `UpgradeRequested` |
| `upgradePending` to `upgradeDownloading` | `UpgradeDownloading` |
| `upgradeDownloading` to `upgrading` | `UpgradeDownloadCompleted` |
| `upgrading` to `initializing` | `UpgradeHelmDone` |
| `upgrading` to `upgradeFailed` | a workload reason such as `HelmUpgradeFailed`; T28 cancel landing here is `upgradeFailed` |
| successful upgrade from `stopped` back to `stopped` | `StopByUser` |
| apply-env accepted, enter `applyingEnv` | `ApplyEnvRequested` |
| `applyingEnv` to `initializing` | `ApplyEnvHelmDone` |
| `applyingEnv` to `applyEnvFailed` | `HelmApplyEnvFailed` |
| successful apply-env from `stopped` back to `stopped` | `StopByUser` |
| uninstall accepted, enter `uninstalling` | `UninstallRequested` |
| `uninstalling` to `uninstalled` | `UninstallCompleted` |
| `uninstalling` to `uninstallFailed` | `UninstallTimedOut` or `HelmUninstallFailed` |
| user cancel, enter `*Canceling` | `CancelRequested` |
| `*Canceling` to `stopping` | `CancelHandoffToStopping`. If the current reason is already a stop reason from the table below, keep that reason. |
| `pendingCanceling` to `pendingCanceled` | `PendingCancelCompleted` |
| `downloadingCanceling` to `downloadingCanceled` | `DownloadingCancelCompleted` |
| `installingCanceling` to `installingCanceled` | `InstallingCancelCompleted` |
| cancel of `upgradePending` / `upgradeDownloading` back to a steady state | `UpgradePendingCanceled` / `UpgradeDownloadCanceled` |
| cancel again from `*CancelFailed` | `CancelRetryRequested` |
| system timeout into `*Canceling` | `downloadCancelBySystem`, `installCancelBySystem`, `initializeCancelBySystem`, `upgradeCancelBySystem`, `applyEnvCancelBySystem`, `resumeCancelBySystem` |
| `*Canceling` timeout into `*CancelFailed` | `PendingCancelTimedOut`, `DownloadingCancelTimedOut`, `InstallingCancelTimedOut`, `InitializingCancelTimedOut`, `ResumingCancelTimedOut`, `ApplyingEnvCancelTimedOut`, `UpgradePendingCancelTimedOut`, `UpgradeDownloadingCancelTimedOut`, `UpgradingCancelTimedOut` |
| `stopping` timeout | `StopTimedOut` |
| `installFailed` cleanup finished | `InstallFailureCleanedUp`; cleanup timeout is `InstallCleanupTimedOut` |

When the state becomes `stopped`, the reason must be one of the stop reasons below. A user stop is recognized by `status.reason == StopRequested` on entry to `stopping`, and the landing on `stopped` rewrites it to `StopByUser`. The other reasons are already set on entry to `stopping` and are not overwritten during scale-down.

| Reason | How | Expected |
|---|---|---|
| `StopByUser` | T04: `market stop` | `running -> stopping reason=StopRequested`, then `stopping -> stopped reason=StopByUser` |
| `Evicted` | On a `running` pod, set `status.reason` to `Evicted` | The controller writes `stopping` / `Evicted`. Then `stopping -> stopped reason=Evicted`. |
| `InitFailed` | `stopinit`: helm succeeds, the container exits immediately, and the failure signal lasts more than 5 minutes | `initializing` to `initializingCanceling`, then `stopping`, then `stopped`, and all three hops use reason `InitFailed`. This is a different edge from T46 `initializeCancelBySystem`. |
| `Unschedulable` | `fault-unready`: the pod stays `Pending` because its Secret does not exist, and `WaitForStartUp` returns `ErrPodPending` after 1 minute. `gcont`: `nodeSelector` names a node that does not exist, so the pod is `Pending` with `PodScheduled=False` / `Unschedulable` and the same error returns on that observation. The same reason also comes from a permanently Pending pod after the app is up. | During install the next state is `stopping`. Terminal state `stopped`, reason `Unschedulable`. A leftover Pending pod blocks scale-down; delete it after asserting `stopping`, or the app becomes `stopFailed` after about 2 minutes. |
| `HamiUnschedulable` | `faultgpu`. The pod requests 64 cards and sets consume policy `single`. HAMi records `InsufficientGPU` when fewer GPUs are bound than requested, and `WaitForStartUp` returns `ErrHamiUnschedulable`. If the event reason is `FailedScheduling`, record `Unschedulable`. That does not pass this row. | During install the next state is `stopping`, not `installingCanceling`. Terminal state `stopped`, reason `HamiUnschedulable`. |
| `DiskPressure` / `SystemCPUPressure` / `SystemMemoryPressure` | Raise the requirement above the cluster's current headroom, but not so high that submit-time total capacity rejects it | `installing` to `stopping`, then `stopped`. The reason matches the resource that is actually short. |
| `K8sRequestCPUPressure` / `K8sRequestMemoryPressure` | The pod request exceeds node allocatable, and requirement validation did not reject it first | Terminal state `stopped`, reason is one of these two. The emitted spelling is `K8sRequestCPUPressure`. |

A stop reason that was not produced is not tested. A produced stop reason that does not match is a failure.

## Restart and continue

Do this once for all 42 states. Workload claims in the process are gone after restart. Whether the operation continues depends only on `status.state`, `status.opID`, and the phase in the payload already written on the ApplicationManager. The restart itself must not move the object into a failure state and must not replace `opID`.

Delete the pod. Do not scale the replica count to 0. The image after startup must still be the current tag. Being rolled back to `beclab/app-service` is a failure. If the state flashed past before the process stopped, record it as missed and retry from an entry that can hold the state. Do not substitute a different state.

| State | Before stop | After startup |
|---|---|---|
| In progress: `pending`, `downloading`, `installing`, `initializing`, `stopping`, `resuming`, `upgradePending`, `upgradeDownloading`, `upgrading`, `applyingEnv`, `uninstalling`, and the nine `*Canceling` states | record `state`, `opID`, and payload phase | Still the same state and the same `opID`. If the phase is already set, the tick only re-dispatches that step's workload. If the phase is still empty, the tick restarts from the first step of that state and it is still the same operation. It then reaches that operation's original terminal state. |
| Steady and failed: `running`, `stopped`, `uninstalled`, the seven operation `*Failed` states, the three `*Canceled` states, and the nine `*CancelFailed` states | record `state`, `reason`, and `opID` | The restart itself emits no `state change`, and those three fields do not change. After the pod is Ready, the next operation for that state must be accepted: `stop` on `running`, `resume` on `stopped`, `install` on `uninstalled`, and the matching T64–T92 row for the rest. |

The empty tick of `initializingCanceling`, `resumingCanceling`, `applyingEnvCanceling`, `upgradingCanceling`, and `upgradePendingCanceling` leaves immediately. For restart coverage, hold that state and then delete the pod. If it cannot be held, record it as missed. The matching `*CancelFailed` states do hold (`resumingCancelFailed` does not leave on an empty tick, see T63). Reach the failed state, then run the restart check.

## Pass

An edge passes only when all of these hold: `state` is the terminal state of that T row; the `state change` reason matches the reason table; upgrade edges match the version table for chart, target, pre, config, and Helm; T12–T14 delete the namespace and T77–T79 can install again; T15–T17 keep the namespace and `resume` can return to `running`; T28 keeps the namespace and then T72 can enter `upgradePending` or T73 reaches `uninstalled`; a failure edge enters the matching failed state and then recovers through T64–T92; T44 and T63 do not change state on restart or an empty tick; an operation that must be rejected is rejected and the state is not rewritten. Restart coverage stops each of the 42 states once: an in-progress operation reaches its original terminal state with the original `opID`, and a steady or failed state still accepts the next operation that state allows. After each fault edge, remove `test.olares/hold` and uninstall the fault app.

All 42 rows of the state table need one entry record. A missing row means the run is not finished. Record S1, S3, and S4 on their own. Do not substitute T43 for S1. S3 fails when an apply from `stopped` starts pods, or a later apply from `running` stops them. S4 fails when a retry of a stopped-app upgrade leaves `stopped`.
