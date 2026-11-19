package ports

import (
	"context"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// ImagePuller drives the ImageManager CR lifecycle used by the
// Downloading and Upgrading states to pre-pull container images
// before helm install/upgrade can succeed.
//
// Legacy call sites:
//   - pkg/appstate/downloading_app.go:133-139           (start pull)
//   - pkg/appstate/downloading_app.go:28                (progress poll)
//   - pkg/appstate/upgrading_app.go:257-269             (upgrade pull)
//   - pkg/appstate/downloading_canceling_app.go:49      (cancel)
//   - pkg/appstate/upgrading_canceling_app.go:46        (cancel)
//   - pkg/appstate/downloading_canceling_failed_app.go:66 (repair)
//   - pkg/images/client.go                              (adapter home)
type ImagePuller interface {
	// ResolveRefs materializes the list of image refs the AM needs.
	// HANDLER-SAFE READ.
	//
	// `appCfg` — concrete type: *pkg/appcfg.ApplicationConfig.
	// Adapters cast on entry; see the design note on
	// WorkloadRequest.AppCfg for rationale.
	ResolveRefs(ctx context.Context, am *appv1.ApplicationManager, appCfg any) ([]appv1.Ref, error)

	// CreateJob starts the ImageManager controller pulling `refs`.
	// RUNNER-ONLY WRITE: dispatched from domain.CallWorkload{Kind:
	// WorkloadPullImage}. Existing IM on the same AM is replaced (see
	// pkg/images/client.go:64-91).
	CreateJob(ctx context.Context, am *appv1.ApplicationManager, refs []appv1.Ref) error

	// CancelJob flips the IM into the DownloadingCanceled state,
	// stopping any in-flight pulls. RUNNER-ONLY WRITE: dispatched
	// from domain.CallWorkload{Kind: WorkloadCancelImagePull}.
	CancelJob(ctx context.Context, amName string) error

	// UpdateJobStatus is the low-level status writer the
	// *_cancel_failed handlers use to repair a stuck IM CR.
	// RUNNER-ONLY WRITE.
	UpdateJobStatus(ctx context.Context, amName, state, message string) error

	// Progress returns the current pull progress for the AM. Zero
	// value + nil error means "no IM CR exists yet".
	// HANDLER-SAFE READ: adapter reads the IM CR from the informer
	// cache; Runner materializes the result into domain.World before
	// Engine.Step so the Downloading handler can decide whether to
	// stay, advance, or fail.
	Progress(ctx context.Context, amName string) (ImageProgress, error)
}

// ImageProgress is the pre-digested pull status the Runner attaches
// to domain.World. Handlers branch on it rather than parsing the raw
// ImageManager CR.
type ImageProgress struct {
	// Percent is 0..100 from images.DownloadPercent. NATS progress
	// events format this to two decimal places, matching legacy
	// ImageManagerClient.updateProgress. 100 with Complete=true
	// means success; 100 with Failed=true means "all refs decided,
	// at least one failed permanently".
	Percent float64

	Complete bool
	Failed   bool

	Reason  string
	Message string
}
