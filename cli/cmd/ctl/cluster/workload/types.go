package workload

import (
	"fmt"
	"strings"
	"time"
)

// SupportedKinds lists the workload kinds the cluster tree models
// today. Kept in canonical singular form — list/get verbs accept
// either singular (deployment) or plural (deployments); the
// normalize helper maps both to the path-segment shape the
// KubeSphere / K8s native APIs expect (always plural).
var SupportedKinds = []string{"deployment", "statefulset", "daemonset"}

// KindAll is the magic value the `list` verb uses to mean "fan out
// to every kind in SupportedKinds". Not part of SupportedKinds
// itself so the get/yaml verbs can reject it cleanly.
const KindAll = "all"

// NormalizeKind converts user-facing kind input to the canonical
// plural lowercase form expected by both KubeSphere and K8s APIs.
//
// Accepts: deployment / deployments / Deployment / DEPLOYMENT (and
// the same for statefulset / daemonset). Returns ("deployments",
// nil) etc. on success, ("", error) on unknown.
//
// Special case: KindAll is preserved verbatim so list.go can branch
// on it without re-checking case.
func NormalizeKind(s string) (string, error) {
	t := strings.ToLower(strings.TrimSpace(s))
	switch t {
	case "":
		return "", fmt.Errorf("kind is required")
	case KindAll:
		return KindAll, nil
	case "deployment", "deployments", "deploy":
		return "deployments", nil
	case "statefulset", "statefulsets", "sts":
		return "statefulsets", nil
	case "daemonset", "daemonsets", "ds":
		return "daemonsets", nil
	default:
		return "", fmt.Errorf("unsupported kind %q (want one of: deployment, statefulset, daemonset, all)", s)
	}
}

// SingularKind returns the user-facing singular form of a normalized
// plural kind (used to populate the KIND column when the wire item
// didn't carry kind). plural must be a value from NormalizeKind's
// success branch.
func SingularKind(plural string) string {
	switch plural {
	case "deployments":
		return "Deployment"
	case "statefulsets":
		return "StatefulSet"
	case "daemonsets":
		return "DaemonSet"
	default:
		return plural
	}
}

// Workload is the minimal typed view we decode from both the
// KubeSphere paginated list envelope and the K8s native single-
// resource GET. Spans Deployment / StatefulSet / DaemonSet — fields
// not relevant to a particular kind stay zero-valued.
//
// As with cmd/ctl/cluster/pod/types.go::Pod, the goal is to model
// only the fields actually rendered or summarized by verbs in this
// package. Verbs that need richer detail can decode the same body
// into their own typed view; for `yaml` we use the raw response
// bytes to avoid dropping unknown fields entirely.
type Workload struct {
	APIVersion string           `json:"apiVersion,omitempty"`
	Kind       string           `json:"kind,omitempty"`
	Metadata   WorkloadMetadata `json:"metadata"`
	Spec       WorkloadSpec     `json:"spec"`
	Status     WorkloadStatus   `json:"status"`
}

type WorkloadMetadata struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	UID               string            `json:"uid,omitempty"`
	CreationTimestamp string            `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Generation        int64             `json:"generation,omitempty"`
}

type WorkloadSpec struct {
	// Replicas is the desired count for Deployment / StatefulSet.
	// Pointer so we can distinguish "missing" (DaemonSet) from
	// "explicit zero" (scaled-to-zero Deployment).
	Replicas *int             `json:"replicas,omitempty"`
	Selector WorkloadSelector `json:"selector,omitempty"`
	// Strategy carries Deployment.spec.strategy.type
	// (RollingUpdate / Recreate); we surface it in get -o table.
	Strategy struct {
		Type string `json:"type,omitempty"`
	} `json:"strategy,omitempty"`
	// UpdateStrategy carries StatefulSet/DaemonSet update strategy
	// (OnDelete / RollingUpdate). One of Strategy or UpdateStrategy
	// is populated depending on the kind.
	UpdateStrategy struct {
		Type string `json:"type,omitempty"`
	} `json:"updateStrategy,omitempty"`
	ServiceName string `json:"serviceName,omitempty"` // StatefulSet only
}

type WorkloadSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

type WorkloadStatus struct {
	// Deployment + StatefulSet fields.
	Replicas          int `json:"replicas,omitempty"`
	ReadyReplicas     int `json:"readyReplicas,omitempty"`
	AvailableReplicas int `json:"availableReplicas,omitempty"`
	UpdatedReplicas   int `json:"updatedReplicas,omitempty"`
	// DaemonSet fields.
	CurrentNumberScheduled int `json:"currentNumberScheduled,omitempty"`
	NumberReady            int `json:"numberReady,omitempty"`
	DesiredNumberScheduled int `json:"desiredNumberScheduled,omitempty"`
	NumberMisscheduled     int `json:"numberMisscheduled,omitempty"`
	NumberAvailable        int `json:"numberAvailable,omitempty"`
	// UpdatedNumberScheduled is the DaemonSet equivalent of
	// updatedReplicas — number of nodes whose pod is on the latest
	// template version. Used by `cluster workload rollout-status` to
	// decide whether a DaemonSet rollout has converged.
	UpdatedNumberScheduled int `json:"updatedNumberScheduled,omitempty"`
	// Generation observed by the controller; if it lags
	// metadata.generation the rollout is still in progress.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// Ready returns the "x/y" READY column for a workload, picking the
// right counts based on kind. Falls back to "-/-" when neither shape
// applies (shouldn't happen in practice).
func (w Workload) Ready(plural string) string {
	switch plural {
	case "deployments", "statefulsets":
		desired := 0
		if w.Spec.Replicas != nil {
			desired = *w.Spec.Replicas
		} else {
			// status.replicas as a fallback (controller-set when
			// the spec lacks an explicit value, which only happens
			// for very old objects).
			desired = w.Status.Replicas
		}
		return fmt.Sprintf("%d/%d", w.Status.ReadyReplicas, desired)
	case "daemonsets":
		return fmt.Sprintf("%d/%d", w.Status.NumberReady, w.Status.DesiredNumberScheduled)
	}
	// Unknown kind — surface the generic shape so the user can at
	// least see "what does the server think the count is".
	return "-/-"
}

// Available returns a coarse boolean sentence for the get summary:
// "all replicas available" / "X of Y available" / "rolling out". Used
// in the vertical key/value layout, NOT in the list table (the
// READY column there already conveys availability).
func (w Workload) Available(plural string) string {
	switch plural {
	case "deployments", "statefulsets":
		desired := 0
		if w.Spec.Replicas != nil {
			desired = *w.Spec.Replicas
		}
		switch {
		case desired == 0:
			return "scaled to zero"
		case w.Status.AvailableReplicas == desired:
			return "all replicas available"
		default:
			return fmt.Sprintf("%d of %d replicas available", w.Status.AvailableReplicas, desired)
		}
	case "daemonsets":
		switch {
		case w.Status.DesiredNumberScheduled == 0:
			return "no nodes selected"
		case w.Status.NumberReady == w.Status.DesiredNumberScheduled:
			return "all scheduled pods ready"
		default:
			return fmt.Sprintf("%d of %d scheduled pods ready",
				w.Status.NumberReady, w.Status.DesiredNumberScheduled)
		}
	}
	return "-"
}

// RolloutInProgress reports whether the controller hasn't caught up
// with the spec yet. Used by the get summary to flag in-flight
// rollouts ("rollout in progress" suffix). Compared against
// metadata.generation rather than status.replicas so the verdict
// matches kubectl rollout status semantics.
func (w Workload) RolloutInProgress() bool {
	return w.Metadata.Generation > 0 && w.Status.ObservedGeneration < w.Metadata.Generation
}

// updateStrategyLabel returns the kind-appropriate update-strategy
// string for the get summary. Exposed so get.go can render it
// without re-implementing the per-kind branching.
func (w Workload) UpdateStrategyLabel(plural string) string {
	switch plural {
	case "deployments":
		return dashIfEmpty(w.Spec.Strategy.Type)
	case "statefulsets", "daemonsets":
		return dashIfEmpty(w.Spec.UpdateStrategy.Type)
	}
	return "-"
}

// Age returns the AGE column string. Mirrors pod.ageOf semantics so
// the two list tables read the same way.
func (w Workload) Age(now time.Time) string {
	return ageOf(w.Metadata.CreationTimestamp, now)
}

// ageOf / dashIfEmpty mirror the helpers in cmd/ctl/cluster/pod/types.go.
// Re-declared here to keep the workload package independent of the
// pod package — both are leaf nouns, neither should depend on the
// other.
func ageOf(ts string, now time.Time) string {
	if ts == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return "-"
	}
	d := now.Sub(t)
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
