package cronjob

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// CronJob is the minimal typed view used for both the KubeSphere
// list envelope and the K8s native single-resource GET. Only the
// fields actually rendered or summarized by verbs in this package
// are modeled — `cronjob yaml` uses raw bytes so unknown fields
// aren't dropped from that output path.
type CronJob struct {
	APIVersion string          `json:"apiVersion,omitempty"`
	Kind       string          `json:"kind,omitempty"`
	Metadata   CronJobMetadata `json:"metadata"`
	Spec       CronJobSpec     `json:"spec"`
	Status     CronJobStatus   `json:"status"`
}

type CronJobMetadata struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	UID               string            `json:"uid,omitempty"`
	ResourceVersion   string            `json:"resourceVersion,omitempty"`
	CreationTimestamp string            `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
}

type CronJobSpec struct {
	Schedule          string          `json:"schedule"`
	Suspend           *bool           `json:"suspend,omitempty"`
	ConcurrencyPolicy string          `json:"concurrencyPolicy,omitempty"`
	JobTemplate       JobTemplateSpec `json:"jobTemplate"`
}

// JobTemplateSpec mirrors batchv1beta1.JobTemplateSpec — only the
// metadata.labels block is exercised today (used by `cronjob jobs` to
// derive a labelSelector that matches the spawned Jobs).
type JobTemplateSpec struct {
	Metadata struct {
		Labels map[string]string `json:"labels,omitempty"`
	} `json:"metadata,omitempty"`
}

type CronJobStatus struct {
	Active           []ObjectRef `json:"active,omitempty"`
	LastScheduleTime string      `json:"lastScheduleTime,omitempty"`
}

// ObjectRef is the minimal ObjectReference subset — name + uid are
// enough for the active-jobs surfacing in `cronjob get`.
type ObjectRef struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Name       string `json:"name,omitempty"`
	UID        string `json:"uid,omitempty"`
}

// suspendLabel returns the SUSPEND column value: "True" / "False" /
// "-" (for very old objects with the field unset; rare).
func (c CronJob) suspendLabel() string {
	if c.Spec.Suspend == nil {
		return "-"
	}
	if *c.Spec.Suspend {
		return "True"
	}
	return "False"
}

// lastScheduleLabel returns a human-readable "<age> ago" string for
// the LAST-SCHEDULE column. "-" when the controller hasn't set the
// field yet (the cronjob hasn't fired since creation).
func (c CronJob) lastScheduleLabel(now time.Time) string {
	if c.Status.LastScheduleTime == "" {
		return "-"
	}
	return ageOf(c.Status.LastScheduleTime, now) + " ago"
}

// activeJobsLabel returns a comma-joined "<name>, <name>, ..." for
// the active Jobs surfaced in `cronjob get`. Stable order so repeat
// runs diff cleanly.
func (c CronJob) activeJobsLabel() string {
	if len(c.Status.Active) == 0 {
		return "-"
	}
	names := make([]string, 0, len(c.Status.Active))
	for _, a := range c.Status.Active {
		if a.Name != "" {
			names = append(names, a.Name)
		}
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// templateLabelSelector builds the "key=value,key=value" selector the
// `cronjob jobs` verb uses to find the Jobs spawned by this CronJob.
// Matches the SPA's JobsDetails.vue cronjob → jobs lookup which
// rebuilds the selector from spec.jobTemplate.metadata.labels.
//
// Returns "" when the template carries no labels — caller should
// treat that as "cannot find spawned jobs" and surface a clear error
// rather than fanning out to "list every job in the namespace".
func (c CronJob) templateLabelSelector() string {
	labels := c.Spec.JobTemplate.Metadata.Labels
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+labels[k])
	}
	return strings.Join(pairs, ",")
}

func (c CronJob) age(now time.Time) string {
	return ageOf(c.Metadata.CreationTimestamp, now)
}

// ageOf / dashIfEmpty mirror the helpers in
// cmd/ctl/cluster/{pod,workload,application,job}/types.go.
// Re-declared here to keep the cronjob package independent.
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
