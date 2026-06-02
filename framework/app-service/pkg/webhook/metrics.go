package webhook

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

// d2 offloader fail-open skip reason labels.
const (
	d2SkipReasonSnapshotError    = "snapshot_error"
	d2SkipReasonViewerUnderive   = "viewer_underive"
	d2SkipReasonTLSSecretMissing = "tls_secret_missing"
	d2SkipReasonOther            = "other"
)

var d2InjectSkippedTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "app_service_d2_inject_skipped_total",
		Help: "v3 shared-entrance d2 offloader injections skipped (fail-open) by reason",
	},
	[]string{"reason"},
)

func init() {
	prometheus.MustRegister(d2InjectSkippedTotal)
}

// RecordD2InjectSkipped increments the fail-open skip counter for the given
// reason. An empty reason is normalized to "other".
func RecordD2InjectSkipped(reason string) {
	if reason == "" {
		reason = d2SkipReasonOther
	}
	d2InjectSkippedTotal.WithLabelValues(reason).Inc()
}

// ClassifyD2SkipReason maps a d2 offloader prerequisite error to a stable
// reason label by errors.Is priority. Unrecognized errors fall back to "other".
func ClassifyD2SkipReason(err error) string {
	switch {
	case err == nil:
		return d2SkipReasonOther
	case errors.Is(err, ErrD2SnapshotUnavailable):
		return d2SkipReasonSnapshotError
	case errors.Is(err, ErrD2ViewerUnderive):
		return d2SkipReasonViewerUnderive
	case errors.Is(err, ErrD2TLSSecretMissing):
		return d2SkipReasonTLSSecretMissing
	default:
		return d2SkipReasonOther
	}
}
