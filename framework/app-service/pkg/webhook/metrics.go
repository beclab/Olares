package webhook

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

// d2 offloader fail-open skip reason labels.
//
// requirement: exhaustive reason set (8 reasons). Includes caller_viewer_unresolved
// and multi_ref_unsupported, plus the caller-mode reasons clusterappref_empty
// and image_unconfigured.
const (
	d2SkipReasonSnapshotError          = "snapshot_error"
	d2SkipReasonViewerUnderive         = "viewer_underive"
	d2SkipReasonTLSSecretMissing       = "tls_secret_missing"
	d2SkipReasonCallerViewerUnresolved = "caller_viewer_unresolved"
	d2SkipReasonMultiRefUnsupported    = "multi_ref_unsupported"
	d2SkipReasonClusterAppRefEmpty     = "clusterappref_empty"
	d2SkipReasonImageUnconfigured      = "image_unconfigured"
	d2SkipReasonOther                  = "other"
)

// d2 offloader inject mode/scenario labels for the succeeded counter:
// mode distinguishes server vs caller injection. Caller-mode
// scenarios are bypass-specific: scenario A = Type-1 (v1/v2, bypass a),
// scenario B = Type-2b (v3 non-shared, bypass b). Server-mode has a single
// scenario, the v3 shared-entrance main path (ServerMain), which is not a
// bypass and must not reuse the caller A/B labels.
const (
	D2InjectModeServer         = "server"
	D2InjectModeCaller         = "caller"
	D2InjectScenarioA          = "A"
	D2InjectScenarioB          = "B"
	D2InjectScenarioServerMain = "server_main"
)

var d2InjectSkippedTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "app_service_d2_inject_skipped_total",
		Help: "v3 shared-entrance d2 offloader injections skipped (fail-open) by reason",
	},
	[]string{"reason"},
)

var d2InjectSucceededTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "app_service_d2_inject_succeeded_total",
		Help: "d2 offloader injections succeeded by mode (server/caller) and scenario (A/B)",
	},
	[]string{"mode", "scenario"},
)

func init() {
	prometheus.MustRegister(d2InjectSkippedTotal)
	prometheus.MustRegister(d2InjectSucceededTotal)
}

// RecordD2InjectSucceeded increments the success counter for the given inject
// mode and scenario.
func RecordD2InjectSucceeded(mode, scenario string) {
	d2InjectSucceededTotal.WithLabelValues(mode, scenario).Inc()
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
	case errors.Is(err, ErrD2CallerViewerUnresolved):
		return d2SkipReasonCallerViewerUnresolved
	case errors.Is(err, ErrD2ClusterAppRefEmpty):
		return d2SkipReasonClusterAppRefEmpty
	case errors.Is(err, ErrD2ImageUnconfigured):
		return d2SkipReasonImageUnconfigured
	default:
		return d2SkipReasonOther
	}
}
