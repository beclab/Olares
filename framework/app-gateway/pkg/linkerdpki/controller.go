package linkerdpki

import (
	"context"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/framework/app-gateway/pkg/entrypki"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// heartbeatInterval keeps /healthz fresh while idling between reconciles;
	// it must be well below livenessHeartbeatThreshold.
	heartbeatInterval      = 30 * time.Second
	entryReconcileInterval = 6 * time.Hour

	// backoffBaseDelay / backoffMaxDelay bound transient-failure retries
	// (detailed design §4: 5s initial, 5min cap, plus ±20% jitter).
	backoffBaseDelay = 5 * time.Second
	backoffMaxDelay  = 5 * time.Minute

	backoffJitterFraction = 0.2
	reconcileTimeout      = time.Minute

	reconcileKey = "reconcile"
)

// Controller runs a level-triggered reconcile loop: an immediate pass at start,
// then one pass per interval, with exponential backoff on transient failures.
// It never calls os.Exit; the process exits only when ctx is cancelled.
type Controller struct {
	client           client.Client
	ns               string
	interval         time.Duration
	probes           *ProbeState
	metrics          *Metrics
	entryMTLSEnabled bool
	entryNamespace   string
}

// NewController wires the reconcile loop to its client, probes and metrics.
func NewController(c client.Client, ns string, interval time.Duration, probes *ProbeState, metrics *Metrics) *Controller {
	return &Controller{
		client:   c,
		ns:       ns,
		interval: interval,
		probes:   probes,
		metrics:  metrics,
	}
}

// WithEntryMTLS enables the optional L4→AGW Entry PKI maintain loop (separate CA).
func (c *Controller) WithEntryMTLS(enabled bool, entryNamespace string) *Controller {
	c.entryMTLSEnabled = enabled
	c.entryNamespace = entryNamespace
	return c
}

// Run drives reconciliation until ctx is cancelled, returning on graceful
// shutdown. Transient failures are retried with backoff and never terminate.
func (c *Controller) Run(ctx context.Context) {
	heartbeatTicker := time.NewTicker(heartbeatInterval)
	defer heartbeatTicker.Stop()
	c.probes.Heartbeat()

	var workers sync.WaitGroup
	start := func(domain string, interval time.Duration, reconcile func(context.Context) error) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			c.runReconciler(ctx, domain, interval, reconcile)
		}()
	}
	start("linkerd", c.interval, c.reconcileOnce)
	if c.entryMTLSEnabled {
		start("entry", entryReconcileInterval, c.reconcileEntry)
	}
	defer workers.Wait()

	for {
		select {
		case <-ctx.Done():
			slog.Info("guardian shutting down")
			return
		case <-heartbeatTicker.C:
			c.probes.Heartbeat()
		}
	}
}

func (c *Controller) runReconciler(ctx context.Context, domain string, interval time.Duration, reconcile func(context.Context) error) {
	limiter := workqueue.NewTypedItemExponentialFailureRateLimiter[string](backoffBaseDelay, backoffMaxDelay)
	for ctx.Err() == nil {
		attempt, cancel := context.WithTimeout(ctx, reconcileTimeout)
		err := reconcile(attempt)
		cancel()
		if ctx.Err() != nil {
			return
		}
		delay := interval
		if err != nil {
			delay = min(withJitter(limiter.When(reconcileKey)), backoffMaxDelay)
			slog.Warn("PKI reconcile failed; backing off",
				"domain", domain, "reason", classifyReason(err), "error", err, "retry_in", delay.String())
		} else {
			limiter.Forget(reconcileKey)
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (c *Controller) reconcileEntry(ctx context.Context) error {
	err := entrypki.MaintainEntryMTLS(ctx, c.client, c.entryNamespace, true, time.Now().UTC())
	if ctx.Err() == context.Canceled {
		return err
	}
	c.metrics.recordEntryResult(time.Now(), err)
	return err
}

func (c *Controller) reconcileOnce(ctx context.Context) error {
	start := time.Now()
	err := MaintainLinkerdPKI(ctx, c.client, c.ns)
	c.metrics.ObserveReconcile(time.Since(start))
	c.probes.MarkAttempted()
	c.observeIssuer(ctx)
	if err != nil {
		c.metrics.IncFailure(classifyReason(err))
		return err
	}
	c.metrics.MarkSuccess(time.Now())
	c.probes.MarkSuccess()
	return nil
}

// observeIssuer best-effort updates linkerd_issuer_not_after_seconds without
// failing the reconcile; -1 marks an unreadable issuer.
func (c *Controller) observeIssuer(ctx context.Context) {
	mat, ok, err := loadPKISecret(ctx, c.client, c.ns)
	if err != nil || !ok {
		c.metrics.SetIssuerNotAfterSeconds(-1)
		return
	}
	notAfter, err := certificateNotAfter(mat.IssuerCrt)
	if err != nil {
		c.metrics.SetIssuerNotAfterSeconds(-1)
		return
	}
	c.metrics.SetIssuerNotAfterSeconds(time.Until(notAfter).Seconds())
}

// withJitter applies ±backoffJitterFraction to a backoff delay to avoid
// synchronized retries.
func withJitter(d time.Duration) time.Duration {
	factor := 1 + (rand.Float64()*2-1)*backoffJitterFraction
	return time.Duration(float64(d) * factor)
}

// classifyReason maps a reconcile error to a stable metric reason label.
func classifyReason(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		return "secret_not_found"
	case strings.Contains(msg, "rotate"):
		return "rotate"
	case strings.Contains(msg, "sync identity"):
		return "sync"
	case strings.Contains(msg, "rollout"), strings.Contains(msg, "identity"):
		return "identity_restart"
	case strings.Contains(msg, "patch"), strings.Contains(msg, "update"):
		return "patch"
	default:
		return "apiserver"
	}
}
