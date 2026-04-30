package pod

import (
	"fmt"
	"time"
)

// Pod is the minimal typed view we decode from both the KubeSphere
// list envelope (`/kapis/resources.kubesphere.io/v1alpha3/pods`) and
// the K8s-native single-resource GET (`/api/v1/namespaces/<ns>/pods/<name>`).
//
// Only the fields actually rendered or summarized by the cluster-pod
// verbs are modeled — keeping the struct small lets us avoid pulling
// in k8s.io/api/core/v1 just for shape.
//
// Wire compatibility note: SPA's PodItem in
// apps/packages/app/src/apps/controlPanelCommon/network/network.ts:59
// describes the same shape (modulo SPA-only fields). Future fields
// (env, volumeMounts, ...) can be added here when a verb needs them.
type Pod struct {
	APIVersion string      `json:"apiVersion,omitempty"`
	Kind       string      `json:"kind,omitempty"`
	Metadata   PodMetadata `json:"metadata"`
	Spec       PodSpec     `json:"spec"`
	Status     PodStatus   `json:"status"`
}

type PodMetadata struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	UID               string            `json:"uid,omitempty"`
	CreationTimestamp string            `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	OwnerReferences   []PodOwnerRef     `json:"ownerReferences,omitempty"`
}

// PodOwnerRef is the minimal subset of metav1.OwnerReference used to
// surface "controlled by" hints in `cluster pod get`.
type PodOwnerRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Controller bool   `json:"controller,omitempty"`
}

type PodSpec struct {
	NodeName       string         `json:"nodeName,omitempty"`
	HostNetwork    bool           `json:"hostNetwork,omitempty"`
	RestartPolicy  string         `json:"restartPolicy,omitempty"`
	ServiceAccount string         `json:"serviceAccount,omitempty"`
	Containers     []PodContainer `json:"containers,omitempty"`
}

// PodContainer captures the per-container fields used by `cluster pod
// get` and `cluster container ...` to render image / ports / env.
// Not the full corev1.Container — fields are added as verbs need
// them.
type PodContainer struct {
	Name            string             `json:"name"`
	Image           string             `json:"image"`
	ImagePullPolicy string             `json:"imagePullPolicy,omitempty"`
	Ports           []PodContainerPort `json:"ports,omitempty"`
	Env             []PodEnvVar        `json:"env,omitempty"`
	// EnvFrom is intentionally NOT modeled here — `cluster container
	// env` only enumerates explicitly-declared env vars, not the
	// implicit ones imported from configMapRef / secretRef. Add this
	// later when a verb actually needs to render the implicit set.
}

type PodContainerPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
}

// PodEnvVar mirrors corev1.EnvVar — the explicit `name + (value |
// valueFrom)` shape. Either `value` or `valueFrom` is populated, not
// both. `cluster container env` renders Value verbatim or, when
// missing, a `(from <source>)` description of ValueFrom so users can
// see where the value would come from at pod-startup time without
// resolving the reference.
type PodEnvVar struct {
	Name      string         `json:"name"`
	Value     string         `json:"value,omitempty"`
	ValueFrom *PodEnvVarFrom `json:"valueFrom,omitempty"`
}

// PodEnvVarFrom is the corev1.EnvVarSource subset our env renderer
// recognizes. ConfigMapKeyRef / SecretKeyRef / FieldRef /
// ResourceFieldRef are the four upstream variants; we model each as
// an optional pointer so the wire JSON's "exactly one of these"
// invariant carries through into Go without extra plumbing.
type PodEnvVarFrom struct {
	ConfigMapKeyRef  *PodEnvKeyRef           `json:"configMapKeyRef,omitempty"`
	SecretKeyRef     *PodEnvKeyRef           `json:"secretKeyRef,omitempty"`
	FieldRef         *PodEnvFieldRef         `json:"fieldRef,omitempty"`
	ResourceFieldRef *PodEnvResourceFieldRef `json:"resourceFieldRef,omitempty"`
}

type PodEnvKeyRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type PodEnvFieldRef struct {
	APIVersion string `json:"apiVersion,omitempty"`
	FieldPath  string `json:"fieldPath"`
}

type PodEnvResourceFieldRef struct {
	ContainerName string `json:"containerName,omitempty"`
	Resource      string `json:"resource"`
	Divisor       string `json:"divisor,omitempty"`
}

type PodStatus struct {
	Phase             string                  `json:"phase,omitempty"`
	HostIP            string                  `json:"hostIP,omitempty"`
	PodIP             string                  `json:"podIP,omitempty"`
	StartTime         string                  `json:"startTime,omitempty"`
	QOSClass          string                  `json:"qosClass,omitempty"`
	ContainerStatuses []PodContainerStatus    `json:"containerStatuses,omitempty"`
	Conditions        []PodStatusCondition    `json:"conditions,omitempty"`
}

// PodContainerStatus is the minimal containerStatus shape used to
// compute READY (x/y) and RESTARTS columns. State is intentionally
// modeled as raw map so we can read either {running:{}} / {waiting:{
// reason}} / {terminated:{exitCode,reason}} without forcing a decode
// path for each.
type PodContainerStatus struct {
	Name         string                 `json:"name"`
	Ready        bool                   `json:"ready"`
	RestartCount int                    `json:"restartCount"`
	Image        string                 `json:"image,omitempty"`
	State        map[string]interface{} `json:"state,omitempty"`
}

type PodStatusCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// readyCount returns "x/y" — number of ready containers over total.
// Mirrors `kubectl get pods` READY column semantics.
func (p Pod) readyCount() string {
	total := len(p.Status.ContainerStatuses)
	if total == 0 {
		// Spec containers exist but the controller hasn't reported any
		// status yet (very fresh pod); fall back to 0/<spec count> so
		// the column is still meaningful instead of "0/0".
		total = len(p.Spec.Containers)
	}
	ready := 0
	for _, c := range p.Status.ContainerStatuses {
		if c.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d", ready, total)
}

// totalRestarts sums RestartCount across all container statuses.
// Matches `kubectl get pods` RESTARTS column.
func (p Pod) totalRestarts() int {
	n := 0
	for _, c := range p.Status.ContainerStatuses {
		n += c.RestartCount
	}
	return n
}

// statusReason picks the most user-actionable label for the STATUS
// column:
//   - Any container in waiting state with a non-empty reason wins
//     (that's what `kubectl` does — surfaces ImagePullBackOff /
//     CrashLoopBackOff ahead of the otherwise-Pending phase).
//   - Otherwise the pod's top-level phase (Running / Pending /
//     Succeeded / Failed / Unknown).
//   - Otherwise empty so callers print "-".
func (p Pod) statusReason() string {
	for _, c := range p.Status.ContainerStatuses {
		if w, ok := c.State["waiting"].(map[string]interface{}); ok {
			if reason, ok := w["reason"].(string); ok && reason != "" {
				return reason
			}
		}
	}
	for _, c := range p.Status.ContainerStatuses {
		if t, ok := c.State["terminated"].(map[string]interface{}); ok {
			if reason, ok := t["reason"].(string); ok && reason != "" {
				return reason
			}
		}
	}
	return p.Status.Phase
}

// age returns a coarse "X(s|m|h|d)" string mirroring `kubectl get`'s
// AGE column semantics. Uses creationTimestamp; falls back to "-" when
// missing or unparseable.
func (p Pod) age(now time.Time) string {
	return ageOf(p.Metadata.CreationTimestamp, now)
}

// ageOf is the shared "elapsed since RFC3339 ts" formatter used by
// both pod rows and event rows. Defined here (rather than a separate
// helper file) so list.go / events.go can share it without leaking a
// pkg-level dependency.
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

// dashIfEmpty is used by the row formatters to render absent fields
// as "-" instead of empty cells (so columnar output stays alignable).
func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
