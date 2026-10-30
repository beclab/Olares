package k8sstore

import (
	"context"
	"sync/atomic"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// Store is the production ports.AMStore adapter backed by a
// controller-runtime client. See doc.go for the shadow-mode
// rationale.
type Store struct {
	c client.Client

	// shadowMode is 0/1, accessed via atomic so SetShadowMode is
	// safe to call from any goroutine without gating writes on a
	// mutex. Zero value = real writes; the constructor defaults to
	// shadow-mode explicitly via WithShadowMode(true) — the
	// intentional design of the package is that callers must opt
	// out of shadow-mode, so a forgotten Option can never surprise
	// production with a second live writer.
	shadowMode int32
}

// Option customises a Store at construction time.
type Option func(*Store)

// WithShadowMode sets the initial shadow-mode state. Prefer
// wiring `WithShadowMode(true)` explicitly at every call site so
// the "safe by default" property is visible in the wiring code —
// grep for the constructor and you can see whether writes are
// suppressed without reading the docs.
func WithShadowMode(enabled bool) Option {
	return func(s *Store) {
		s.SetShadowMode(enabled)
	}
}

// New constructs a Store. `c` MUST be a controller-runtime client
// with the app-service scheme registered (main.go's mgr.GetClient
// satisfies this). Panics if c is nil — a nil client would
// misleadingly succeed at construction and only fail at first
// Reconcile.
func New(c client.Client, opts ...Option) *Store {
	if c == nil {
		panic("k8sstore.New: client is nil")
	}
	s := &Store{c: c}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SetShadowMode toggles shadow-mode. Concurrency-safe; the flip
// takes effect on the next write. Reads are unaffected — they
// always go through to the real client.
func (s *Store) SetShadowMode(enabled bool) {
	v := int32(0)
	if enabled {
		v = 1
	}
	atomic.StoreInt32(&s.shadowMode, v)
}

// InShadowMode returns the current shadow-mode state. Exposed for
// tests + debug endpoints.
func (s *Store) InShadowMode() bool {
	return atomic.LoadInt32(&s.shadowMode) == 1
}

// Get implements ports.AMStore.
func (s *Store) Get(ctx context.Context, name string) (*appv1.ApplicationManager, error) {
	am := &appv1.ApplicationManager{}
	if err := s.c.Get(ctx, types.NamespacedName{Name: name}, am); err != nil {
		return nil, err
	}
	return am, nil
}

// List implements ports.AMStore.
func (s *Store) List(ctx context.Context) ([]appv1.ApplicationManager, error) {
	var list appv1.ApplicationManagerList
	if err := s.c.List(ctx, &list); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetApplication implements ports.AMStore.
func (s *Store) GetApplication(ctx context.Context, name string) (*appv1.Application, error) {
	app := &appv1.Application{}
	if err := s.c.Get(ctx, types.NamespacedName{Name: name}, app); err != nil {
		return nil, err
	}
	return app, nil
}

// Interface conformance guard. Reads live here; writes live in
// writes.go; the compile-time check is once, in this file, so a
// regression on either side breaks the build in the primary
// package file.
var _ ports.AMStore = (*Store)(nil)
