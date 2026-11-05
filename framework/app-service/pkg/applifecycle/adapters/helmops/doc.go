// Package helmops is the ports.WorkloadOps adapter — the last and
// largest of the five real port adapters. It is a thin passthrough
// over the existing pkg/appinstaller/versioned Helm pipeline: the
// adapter does no chart logic itself; it exists only to translate
// between the port-level WorkloadRequest and the legacy
// HelmOpsInterface method surface.
//
// # Why this shape
//
// The legacy Helm pipeline (versioned.NewHelmOps + v1/v2/v3
// HelmOpsInterface implementations) is:
//
//   - large (~2000 LOC across three api-version implementations)
//   - correct today (running in production)
//   - highly interconnected with clientset, kubesphere HTTP, PVC
//     lifecycle, permission unregister, provider registration, and
//     app-cache cleanup
//
// Rewriting it in-place would be a multi-quarter project and would
// destabilise the very pipeline every other port depends on.
// Instead, this adapter treats HelmOpsInterface as an ALREADY-
// FACTORED SUB-DOMAIN and wraps it verbatim:
//
//	Install(req)      → versioned.NewHelmOps(...).Install()
//	Upgrade(req)      → versioned.NewHelmOps(...).Upgrade()
//	ApplyEnv(req)     → versioned.NewHelmOps(...).ApplyEnv()
//	Uninstall(req)    → versioned.NewHelmOps(...).Uninstall()
//	UninstallAll(req) → versioned.NewHelmOps(...).UninstallAll()
//	Stop(req)         → versioned.NewHelmOps(...).Scale(0)
//	Resume(req)       → versioned.NewHelmOps(...).Scale(-1)
//	WaitForStartUp    → versioned.NewHelmOps(...).WaitForStartUp()
//	WaitForLaunch     → versioned.NewHelmOps(...).WaitForLaunch()
//
// The two exceptions:
//
//   - WaitForNamespaceDeleted does NOT go through HelmOpsInterface
//     (there is no such method); the legacy PollImmediate loop in
//     pkg/appstate/types.go:267 is duplicated here on the adapter's
//     own client.Client + Clock seams so it is fully unit-testable
//     without spinning up an apiserver.
//
//   - GetDeployedVersion uses apputils.GetDeployedReleaseVersion
//     directly against an action.Configuration built from
//     helm.InitConfig; HelmOpsInterface exposes no equivalent method.
//
// # Runner-side request today vs future
//
// The Runner currently populates ONLY req.AM and req.Namespace
// (pkg/applifecycle/runner/effects.go:233). The other fields
// (KubeConfig, AppCfg, Token, Opt) are zero-valued. The adapter is
// designed to resolve every one of these from AM state so it is
// drop-in usable TODAY, without any Runner changes:
//
//   - KubeConfig: cast req.KubeConfig if non-nil, else use the
//     adapter's injected kubeConfig.
//   - AppCfg: cast req.AppCfg if non-nil, else json-unmarshal
//     AM.Spec.Config.
//   - Token: use req.Token if non-empty, else read
//     AM.Annotations["bytetrade.io/token"].
//   - Opt.MarketSource: use req.Opt.MarketSource if non-empty, else
//     read AM.Annotations["bytetrade.io/market-source"] via
//     appcfg.GetMarketSource.
//
// Once the Runner learns to populate the full WorkloadRequest (a
// separate PR — see runner.effects.go:233 comment), these fallbacks
// stay in place as belt-and-suspenders — no adapter change required.
//
// # Seams
//
// Four functional-Options seams support end-to-end unit testing
// without a real cluster or Helm daemon:
//
//  1. WithHelmOpsFactory overrides versioned.NewHelmOps. Tests
//     inject a factory returning a stub HelmOpsInterface whose
//     Install/Upgrade/... methods record calls and return canned
//     errors.
//
//  2. WithGetVersion overrides apputils.GetDeployedReleaseVersion.
//     Same pattern as WithHelmOpsFactory but for the one read
//     method.
//
//  3. WithClock overrides time.Now for the WaitForNamespaceDeleted
//     poll loop's deadline enforcement. Production wires it to
//     realclock.New; tests use a fake clock.
//
//  4. WithPollInterval + WithDefaultTimeout tune the
//     WaitForNamespaceDeleted loop. Defaults match legacy
//     (1-second interval, 30-minute deadline).
//
// # Resolved impedance leak: WaitForStartUp side effect
//
// Legacy HelmOps.WaitForStartUp used to write appv1.Initializing to
// the AM via apputils.UpdateAppMgrState when the pod check reported
// Ready. That write hit the live apiserver directly, bypassing this
// adapter's k8sstore. It was load-bearing only during the shadow-mode
// window, when the Runner could not yet persist state.
//
// Once k8sstore went live the write became actively harmful, not
// merely redundant: it landed Initializing BEFORE the WorkloadReady
// produced by this adapter was routed, so Engine.Step dispatched that
// action against Initializing rather than Installing. The payload
// still held installing.phase (never initializing.phase), so
// OnInitializingWorkloadReady hard-errored on the empty-phase guard
// and the reconciler wedged in a backoff loop with the app fully
// installed but permanently stuck in initializing.
//
// The in-line write has been removed. Landing Initializing is owned
// by the handler that dispatched the wait — installing.finishInstall
// on the install path, upgrading.handoffToInitializingAfterUpgrade on
// the upgrade path.
//
// # When to wire this into main.go
//
// Only AFTER Step 3 flips k8sstore to live. Reason: helmops write
// methods are the loudest side effects in the entire system —
// running them from the Runner while the Runner cannot yet persist
// Status means the Runner would helm-install an app but never
// promote its AM from Pending, so the legacy controller (still the
// authoritative state writer during shadow-mode) would see a
// half-installed release and try to reconcile from Pending,
// triggering a second helm-install and permanent divergence.
//
// The safe order for Step 3 is:
//
//  1. Confirm k8sstore live-mode writes are byte-identical to the
//     legacy controller's writes (comparing shadow-mode logs).
//  2. Retire the legacy Reconcile (Step 4 preEnqueueCheck flip).
//  3. Wire this adapter into main.go and start the Runner-driven
//     helm ops.
//  4. Flip router in Step 5 so V2 HTTP handlers deliver
//     TriggerOps against the Runner instead of the legacy path.
package helmops
