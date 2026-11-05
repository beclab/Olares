package ports

import "context"

// ComputeOps is the sole caller of the compute (GPU/CPU allocation)
// subsystem. Elevated to a first-class port because nine legacy files
// call into it — hiding it inside InstallValidator or WorkloadOps
// would obscure the coupling and complicate future GPU-scheduler
// swaps.
//
// Legacy call sites:
//   - installing_app.go:193 / :278      DeleteAllocationsForApp / ManagesSharedServer
//   - installing_canceling_app.go:155   DeleteAllocationsForApp
//   - install_failure_cleanup.go:75     DeleteAllocationsForApp
//   - suspending_app.go:158             DeleteAllocationsForComputeTarget
//   - suspend_failed_app.go:81          DeleteAllocationsForComputeTarget
//   - stopped_app.go:68                 EnsureAllocationsDeletedForComputeTarget
//   - uninstalling_app.go:198           DeleteAllocationsForComputeTarget
//   - uninstalled_app.go:90             EnsureAllocationsDeletedForComputeTarget
//   - types.go:240 (forceDeleteApp)     EnsureAllocationsDeletedForComputeTarget
//
// Adapter reads ConfigMaps via the shared controller client; no HTTP.
//
// AppCfg parameters are typed `any` (concrete:
// *pkg/appcfg.ApplicationConfig) to keep the ports package a leaf
// module. See design note on WorkloadRequest.AppCfg.
type ComputeOps interface {
	// DeleteForApp removes the allocation record keyed by (appName,
	// owner). RUNNER-ONLY WRITE.
	DeleteForApp(ctx context.Context, appName, owner string) error

	// DeleteForTarget removes allocations belonging to the given
	// AppConfig's compute target. If includeSharedServer is true the
	// shared-server allocation is also cleared. RUNNER-ONLY WRITE.
	DeleteForTarget(ctx context.Context, appCfg any, includeSharedServer bool) error

	// EnsureDeletedForTarget is a convergent form: it returns
	// cleaned=true once no allocation remains, or (false, nil) to
	// signal "keep polling". The legacy uninstalled_app.go /
	// stopped_app.go loops on this call.
	// RUNNER-ONLY WRITE (with implicit poll inside adapter, which
	// the Runner wraps in a goroutine bounded by domain.CallWorkload
	// timeout).
	EnsureDeletedForTarget(ctx context.Context, appCfg any, uninstallAll bool) (cleaned bool, err error)

	// ManagesSharedServer reports whether the app manages a shared
	// GPU/CPU server, which decides whether stopping the app has to
	// bring the server down with it.
	//
	// Consulted by the Runner on exactly one path — classifying an
	// install-validation rejection — and forwarded to handlers as
	// WorkloadResult.StopAll rather than as a World field. Keeping
	// it off World matters: this is a cluster read, and a World
	// field would pay for it on every Reconcile of every AM to
	// answer a question only the rejection path asks.
	ManagesSharedServer(ctx context.Context, appCfg any) (bool, error)
}
