package imgpull

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/images"
)

// RefResolver is the seam that resolves the image refs an AM needs.
// The default production wiring runs a helm dry-run through
// appinstaller.BuildBaseHelmValues + appstate.GetRefsForImageManager;
// tests inject a stub returning canned refs. See doc.go for why the
// seam lives at the adapter boundary rather than inside the adapter.
//
// The seam signature intentionally uses the concrete
// *appcfg.ApplicationConfig rather than the port-level `any` so the
// production wiring gets compile-time safety and tests get IDE
// autocompletion on the config.
type RefResolver func(ctx context.Context, am *appv1.ApplicationManager, cfg *appcfg.ApplicationConfig) ([]appv1.Ref, error)

// Puller is the production ports.ImagePuller adapter. See doc.go
// for the composition rationale (wraps legacy images.ImageManager
// for write paths; re-implements Progress).
type Puller struct {
	c        client.Client
	im       images.ImageManager
	resolver RefResolver
}

// Option customises a Puller at construction time.
type Option func(*Puller)

// WithImageManager injects an alternate images.ImageManager
// implementation. Production wiring never sets this (the default
// images.NewImageManager(client) is correct); tests use it to
// inject a mock that records calls without needing an
// ImageManager Status subresource on the fake controller-runtime
// client.
func WithImageManager(im images.ImageManager) Option {
	return func(p *Puller) {
		if im != nil {
			p.im = im
		}
	}
}

// New constructs a Puller. Panics on nil client or nil resolver —
// both would silently misbehave (nil client = every method NPEs
// deep inside images.ImageManagerClient; nil resolver = ResolveRefs
// NPE on invocation) and both are unrecoverable at runtime.
func New(c client.Client, resolver RefResolver, opts ...Option) *Puller {
	if c == nil {
		panic("imgpull.New: client is nil")
	}
	if resolver == nil {
		panic("imgpull.New: resolver is nil")
	}
	p := &Puller{
		c:        c,
		im:       images.NewImageManager(c),
		resolver: resolver,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// ResolveRefs implements ports.ImagePuller.
//
// The port declares `appCfg any` and requires adapters to cast on
// entry. This cast is strict: a nil *appcfg.ApplicationConfig OR a
// value of any other type is rejected as a programming error
// (rather than silently returning an empty ref list which would
// let the state machine "successfully" create an empty ImageManager).
func (p *Puller) ResolveRefs(ctx context.Context, am *appv1.ApplicationManager, appCfg any) ([]appv1.Ref, error) {
	if am == nil {
		return nil, fmt.Errorf("imgpull: ResolveRefs requires non-nil AM")
	}
	cfg, ok := appCfg.(*appcfg.ApplicationConfig)
	if !ok || cfg == nil {
		return nil, fmt.Errorf("imgpull: ResolveRefs expected non-nil *appcfg.ApplicationConfig, got %T", appCfg)
	}
	return p.resolver(ctx, am, cfg)
}

// CreateJob implements ports.ImagePuller. Delegates verbatim to
// images.ImageManagerClient.Create: lists Ready+Schedulable nodes,
// removes any stale IM CR for this AM, applies the dev-owner label
// (for -dev / devbox apps), and creates a fresh IM CR with the
// resolved refs pinned to the schedulable node set.
//
// The idempotency guarantee comes from images.ImageManagerClient
// — the Get+Delete+Create sequence tolerates NotFound and never
// double-creates. That's what makes it safe to wire live even
// while the legacy DownloadingApp handler may also be trying to
// create the same IM CR during the shadow-mode migration window.
func (p *Puller) CreateJob(ctx context.Context, am *appv1.ApplicationManager, refs []appv1.Ref) error {
	if am == nil {
		return fmt.Errorf("imgpull: CreateJob requires non-nil AM")
	}
	return p.im.Create(ctx, am, refs)
}

// CancelJob implements ports.ImagePuller. Flips the IM CR's
// Status.State to appv1.DownloadingCanceled — the value the
// ImageManagerController watches for in preEnqueueCheckForUpdate,
// which triggers the in-process containerd cancel of any active
// pull for that IM.
//
// State + message BOTH use DownloadingCanceled.String() to match
// legacy call sites verbatim (download_canceling_app.go:49,
// upgrading_canceling_app.go:46, downloading_canceling_failed_app.go:66).
// The message-equals-state pattern is odd but every legacy caller
// does it, so preserving it keeps observability (Status.Message
// grep-able) identical.
//
// Idempotency: images.ImageManagerClient.UpdateStatus + the
// ImageManagerController's isTerminalState guard (see
// controllers/image_controller.go:270-275) together ensure that
// calling CancelJob on an already-canceled IM is a no-op Patch —
// the state does not regress and the controller does not re-enter
// downloading.
func (p *Puller) CancelJob(ctx context.Context, amName string) error {
	if amName == "" {
		return fmt.Errorf("imgpull: CancelJob requires non-empty AM name")
	}
	canceled := appv1.DownloadingCanceled.String()
	return p.im.UpdateStatus(ctx, amName, canceled, canceled)
}

// UpdateJobStatus implements ports.ImagePuller. The generic escape
// hatch used by *_cancel_failed handlers to nudge a stuck IM CR
// back into a known state. State and message pass through unchanged.
func (p *Puller) UpdateJobStatus(ctx context.Context, amName, state, message string) error {
	if amName == "" {
		return fmt.Errorf("imgpull: UpdateJobStatus requires non-empty AM name")
	}
	return p.im.UpdateStatus(ctx, amName, state, message)
}

// Interface conformance guard. Only reads live in this file; the
// Progress method lives in progress.go alongside the pure
// computation. The guard is registered here (the primary package
// file) so any future signature drift breaks the build in the
// obvious place.
var _ ports.ImagePuller = (*Puller)(nil)
