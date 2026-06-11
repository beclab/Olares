package authz

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	extAuthzDecisionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_service_ext_authz_decisions_total",
			Help: "ext_authz Check outcomes by decision and code",
		},
		[]string{"decision", "code"},
	)
	extAuthzLatencyMS = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_service_ext_authz_latency_ms",
			Help:    "ext_authz Check latency in milliseconds",
			Buckets: []float64{1, 2, 5, 10, 25, 50, 100, 250, 500, 1000},
		},
		[]string{"decision"},
	)
)

func init() {
	prometheus.MustRegister(extAuthzDecisionsTotal, extAuthzLatencyMS)
}

func recordAuthzMetrics(decision, code string, elapsed time.Duration) {
	if decision == "" {
		decision = "unknown"
	}
	if code == "" {
		code = "-"
	}
	extAuthzDecisionsTotal.WithLabelValues(decision, code).Inc()
	extAuthzLatencyMS.WithLabelValues(decision).Observe(float64(elapsed.Milliseconds()))
}
