package ports

import (
	"context"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// MiddlewareOps wraps the two middleware code paths in the legacy
// state machine:
//
//  1. KubeBlocks middleware (Postgres/Redis/MongoDB/Zinc/Nats) —
//     started/stopped alongside app suspend/resume.
//  2. Legacy MongoDB (os-platform) — a pre-KubeBlocks install path
//     that needs to be uninstalled explicitly when the app that
//     depends on it is removed.
//
// Legacy call sites:
//   - suspending_app.go:226-227      NewMiddlewareOp(Stop).Stop
//   - resuming_app.go:230-231        NewMiddlewareOp(Start).Start
//   - suspend_failed_app.go:68-69    Stop retry
//   - types.go:392-398 + uninstalling_app.go:179-181
//     middlewareinstaller.Uninstall
//     (legacy MongoDB path)
type MiddlewareOps interface {
	// Start dispatches the KubeBlocks OpsRequest for middleware
	// resume. RUNNER-ONLY WRITE.
	Start(ctx context.Context, am *appv1.ApplicationManager) error

	// Stop dispatches the KubeBlocks OpsRequest for middleware
	// suspend. RUNNER-ONLY WRITE.
	Stop(ctx context.Context, am *appv1.ApplicationManager) error

	// UninstallLegacyMongo tears down the pre-KubeBlocks MongoDB
	// deployment. RUNNER-ONLY WRITE.
	//
	// `cfg` — concrete type:
	// pkg/middlewareinstaller.MiddlewareConfig. Kept as `any` at the
	// port boundary; adapter casts on entry. See design note on
	// WorkloadRequest.AppCfg for rationale.
	UninstallLegacyMongo(ctx context.Context, cfg any) error
}
