package linkerdpki

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds the guardian's Prometheus collectors on a private registry so
// the controller can be instantiated more than once (e.g. in tests) without a
// duplicate-registration panic. Metrics never carry PEM/private-key material.
type Metrics struct {
	registry       *prometheus.Registry
	issuerNotAfter prometheus.Gauge
	failures       *prometheus.CounterVec
	lastSuccess    prometheus.Gauge
	reconcileDur   prometheus.Histogram
}

// NewMetrics builds and registers the guardian metrics (detailed design §6.1).
func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		registry: reg,
		issuerNotAfter: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "linkerd_issuer_not_after_seconds",
			Help: "Seconds from now until the Linkerd identity issuer NotAfter; -1 if unreadable.",
		}),
		failures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "reconcile_failures_total",
			Help: "Total reconcile failures labelled by reason.",
		}, []string{"reason"}),
		lastSuccess: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "last_success_timestamp_seconds",
			Help: "Unix timestamp of the last successful reconcile.",
		}),
		reconcileDur: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "reconcile_duration_seconds",
			Help:    "Duration of a single ReconcileOnce.",
			Buckets: prometheus.DefBuckets,
		}),
	}
	reg.MustRegister(m.issuerNotAfter, m.failures, m.lastSuccess, m.reconcileDur)
	return m
}

// Handler returns the /metrics HTTP handler bound to the private registry.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// SetIssuerNotAfterSeconds records the issuer remaining validity in seconds.
func (m *Metrics) SetIssuerNotAfterSeconds(v float64) { m.issuerNotAfter.Set(v) }

// IncFailure increments the failure counter for the given reason.
func (m *Metrics) IncFailure(reason string) { m.failures.WithLabelValues(reason).Inc() }

// MarkSuccess records the timestamp of a successful reconcile.
func (m *Metrics) MarkSuccess(t time.Time) { m.lastSuccess.Set(float64(t.Unix())) }

// ObserveReconcile records the duration of one ReconcileOnce.
func (m *Metrics) ObserveReconcile(d time.Duration) { m.reconcileDur.Observe(d.Seconds()) }
