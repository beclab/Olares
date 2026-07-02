package ctrlrequeue

import (
	"context"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// defaultBufferSize matches controller-runtime's source.Channel
// default so we exercise the same backpressure behaviour the
// upstream library ships with. Overridable via WithBufferSize when
// a workload-specific tuning is ever needed.
const defaultBufferSize = 1024

// Requeuer implements runner.Requeuer by writing GenericEvents
// into a buffered channel that a controller-runtime source.Channel
// drains into the reconciler's workqueue.
//
// Zero value is NOT usable — construct via New so the channel is
// sized correctly and the source has something to read from.
type Requeuer struct {
	ch chan event.TypedGenericEvent[*appv1alpha1.ApplicationManager]
}

// Option customises a Requeuer at construction time.
type Option func(*config)

type config struct {
	bufferSize int
}

// WithBufferSize overrides the default channel buffer size.
// Values <= 0 are treated as "use default".
//
// Larger buffers absorb bursts (e.g. reconciling every AM after a
// controller restart) at the cost of memory. In practice the
// default of 1024 comfortably covers even the largest realistic
// per-user Olares deployments.
func WithBufferSize(n int) Option {
	return func(c *config) {
		if n > 0 {
			c.bufferSize = n
		}
	}
}

// New returns a Requeuer ready to be handed to both runner.Config
// and apiserver.NewRunnerDispatcher. The caller MUST also register
// the returned Source with the reconciler builder via
// WatchesRawSource, otherwise Requeue calls sit in the buffer
// forever.
func New(opts ...Option) *Requeuer {
	cfg := config{bufferSize: defaultBufferSize}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Requeuer{
		ch: make(chan event.TypedGenericEvent[*appv1alpha1.ApplicationManager], cfg.bufferSize),
	}
}

// Requeue implements runner.Requeuer.
//
// Non-blocking on purpose: a full buffer means the reconciler is
// deeply backed up and even queueing another Reconcile would land
// on the far side of many others; the next informer resync will
// pick up any state the caller intended to communicate through
// PendingActionQueue. We log a drop at V(1) so an oncall notices
// pathological cases without spamming steady-state logs.
func (r *Requeuer) Requeue(name string) {
	evt := event.TypedGenericEvent[*appv1alpha1.ApplicationManager]{
		Object: &appv1alpha1.ApplicationManager{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		},
	}
	select {
	case r.ch <- evt:
	default:
		klog.V(1).Infof("ctrlrequeue: buffer full, dropping Requeue for %q", name)
	}
}

// Source returns a controller-runtime Source that emits
// reconcile.Request{Name: name} for every Requeue(name) call.
// Wire it into the reconciler builder via WatchesRawSource:
//
//	ctrl.NewControllerManagedBy(mgr).
//	    For(&appv1alpha1.ApplicationManager{}).
//	    WatchesRawSource(requeuer.Source()).
//	    Complete(&YourReconciler{Runner: r})
//
// The map-func synthesises a namespace-less reconcile.Request
// (ApplicationManager is cluster-scoped, matching the Runner's
// runner.Request shape).
func (r *Requeuer) Source() source.Source {
	return source.Channel[*appv1alpha1.ApplicationManager](
		r.ch,
		handler.TypedEnqueueRequestsFromMapFunc[*appv1alpha1.ApplicationManager, reconcile.Request](
			func(_ context.Context, am *appv1alpha1.ApplicationManager) []reconcile.Request {
				if am == nil || am.GetName() == "" {
					return nil
				}
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{Name: am.GetName()}},
				}
			},
		),
	)
}

// Ensure Requeuer satisfies runner.Requeuer at compile-time. This
// is the wire-up invariant callers rely on; a signature drift on
// either side breaks the build here rather than at main.go.
var _ runner.Requeuer = (*Requeuer)(nil)
