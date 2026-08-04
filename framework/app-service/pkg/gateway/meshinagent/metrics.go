package meshinagent

import "github.com/prometheus/client_golang/prometheus"

var meshInForwardTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "olares_mesh_in_forward_total",
		Help: "mesh-in agent forward decisions by path (http|https|passthrough)",
	},
	[]string{"path"},
)

func init() {
	prometheus.MustRegister(meshInForwardTotal)
}

// RecordForward increments the forward counter (optional; primarily for tests/future hooks).
func RecordForward(path string) {
	switch path {
	case "http", "https", "passthrough":
	default:
		path = "http"
	}
	meshInForwardTotal.WithLabelValues(path).Inc()
}
