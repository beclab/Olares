package ports

import (
	"context"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// WorkloadOps is the sole caller of helm and workload-scale APIs.
// Handlers never invoke helm directly; they return a domain.CallWorkload
// effect whose Kind selects which method the Runner dispatches to.
//
// Method split:
//
//   - Install / Upgrade / Uninstall / UninstallAll / Stop / Resume /
//     ApplyEnv are RUNNER-ONLY long-running writes. Each returns
//     a WorkloadResult which the Runner packs into a WorkloadReady
//     or WorkloadFailed action fed back through Engine.Step.
//
//   - GetDeployedVersion is a HANDLER-SAFE READ materialized into
//     domain.World.Workload before Engine.Step.
type WorkloadOps interface {
	Install(ctx context.Context, req WorkloadRequest) WorkloadResult
	Upgrade(ctx context.Context, req WorkloadRequest) WorkloadResult
	ApplyEnv(ctx context.Context, req WorkloadRequest) WorkloadResult
	Uninstall(ctx context.Context, req WorkloadRequest) WorkloadResult
	UninstallAll(ctx context.Context, req WorkloadRequest) WorkloadResult
	Stop(ctx context.Context, req WorkloadRequest) WorkloadResult
	Resume(ctx context.Context, req WorkloadRequest) WorkloadResult
	WaitForStartUp(ctx context.Context, req WorkloadRequest) WorkloadResult
	WaitForLaunch(ctx context.Context, req WorkloadRequest) WorkloadResult

	// WaitForNamespaceDeleted deletes every non-protected target
	// (NotFound is success) then polls until Get returns NotFound.
	// Pending/Downloading cancel have no helm Uninstall, so the
	// delete lives here; uninstall paths are idempotent if the
	// namespace is already gone. Runner runs it as a goroutine;
	// handler never sees it.
	// RUNNER-ONLY WRITE.
	WaitForNamespaceDeleted(ctx context.Context, req NamespaceWaitRequest) WorkloadResult

	// GetDeployedVersion returns the currently-deployed chart
	// version + revision so the upgrade handler can guard against
	// downgrades.
	// HANDLER-SAFE READ (materialized into domain.World.Workload
	// before Engine.Step).
	GetDeployedVersion(ctx context.Context, namespace, release string) (chartVersion string, revision int, err error)
}

// WorkloadRequest is the common input DTO for every long-running
// WorkloadOps method. Fields are populated by the Runner from
// domain.CallWorkload.Params + AM/appcfg lookups. Unused fields for a
// given Kind must be zero (documented per method).
//
// Design note: AppCfg and KubeConfig are typed as `any` (not the
// concrete legacy types) so this port package does not transitively
// import pkg/appcfg / k8s.io/client-go/rest. The adapter casts to the
// documented concrete type on entry. The trade-off is one lost line
// of compile-time safety at the boundary, in exchange for keeping the
// ports package a leaf module — critical for Hexagonal purity and
// for keeping handler unit tests cheap to build.
type WorkloadRequest struct {
	// AM is the ApplicationManager the operation runs against. Never
	// nil for operations that need Status/Spec inspection (Install,
	// Uninstall, Upgrade); can be nil for pure namespace waits.
	AM *appv1.ApplicationManager

	// KubeConfig — concrete type: *k8s.io/client-go/rest.Config.
	// Adapters cast on entry.
	KubeConfig any

	// AppCfg — concrete type: *pkg/appcfg.ApplicationConfig.
	// Adapters cast on entry. Kept as `any` here to prevent the
	// legacy appcfg package (and its 20+ transitive imports) from
	// leaking into the ports leaf module.
	AppCfg any

	Token     string
	Opt       WorkloadOpt
	Namespace string

	// SharedNS is set for UninstallAll. Empty for Install/Upgrade/Stop.
	SharedNS []string

	// Timeout is the maximum wall time the Runner allows the
	// operation to run before feeding WorkloadFailed. Zero uses the
	// Runner-default (30m for helm ops, 10m for wait-namespace).
	Timeout time.Duration
}

// WorkloadOpt mirrors pkg/appinstaller.Opt field-for-field so the
// adapter can trivially convert it without a struct-tag mismatch
// risk. Inlined (rather than reused) to break the transitive
// dependency on pkg/appinstaller.
type WorkloadOpt struct {
	Source             string
	MarketSource       string
	SkipWaitForStartUp bool
}

// NamespaceWaitRequest is the input DTO for WaitForNamespaceDeleted.
// Isolated from WorkloadRequest because it needs neither AppCfg nor
// helm state — only the namespace name + optional list of "shared"
// namespaces to also wait on.
type NamespaceWaitRequest struct {
	Primary  string
	SharedNS []string
	Timeout  time.Duration
}

// WorkloadResult is what every RUNNER-ONLY WRITE method returns to
// the Runner. The Runner translates it into either WorkloadReady
// (Err == nil) or WorkloadFailed (Err != nil) and re-enqueues the
// action through Engine.Step.
//
// LandedState is populated when the operation ends in a state that
// is NOT the caller's happy-path terminal (e.g. Upgrade of a Stopped
// app lands back on Stopped, not Running). Runner uses it to select
// between multiple OnSuccess targets when the CallWorkload effect
// declares them.
type WorkloadResult struct {
	LandedState  string
	LandedReason string
	Err          error

	// StopAll is set alongside LandedState == Stopping when the
	// workload's pods live on a shared server, so the AM must carry
	// the stop-all flag into Stopping for the stop path to tear the
	// server down rather than just this app's workloads.
	//
	// Only Install populates it today, from the
	// errcode.ErrServerSidePodPending branch of legacy
	// installing_app.go.
	StopAll bool
}
