package cronjob

import (
	"sort"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
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

// CronJobSpec models batchv1.CronJobSpec. Pointer types distinguish
// "field omitted by the server" from "explicit zero" — for
// startingDeadlineSeconds in particular, 0 means "must start
// immediately or skip", which differs from "unset".
//
// `timeZone` is a v1 addition (alpha in 1.24, beta in 1.25, GA in
// 1.27). We model it so `cronjob get` and `cronjob yaml` surface it
// when the server populates it.
type CronJobSpec struct {
	Schedule                   string          `json:"schedule"`
	TimeZone                   *string         `json:"timeZone,omitempty"`
	StartingDeadlineSeconds    *int64          `json:"startingDeadlineSeconds,omitempty"`
	ConcurrencyPolicy          string          `json:"concurrencyPolicy,omitempty"`
	Suspend                    *bool           `json:"suspend,omitempty"`
	SuccessfulJobsHistoryLimit *int32          `json:"successfulJobsHistoryLimit,omitempty"`
	FailedJobsHistoryLimit     *int32          `json:"failedJobsHistoryLimit,omitempty"`
	JobTemplate                JobTemplateSpec `json:"jobTemplate"`
}

// JobTemplateSpec mirrors batchv1.JobTemplateSpec — only the
// metadata.labels block is exercised today (used by `cronjob jobs` to
// derive a labelSelector that matches the spawned Jobs).
type JobTemplateSpec struct {
	Metadata struct {
		Labels map[string]string `json:"labels,omitempty"`
	} `json:"metadata,omitempty"`
}

// CronJobStatus mirrors batchv1.CronJobStatus. `lastSuccessfulTime`
// is a v1 addition over v1beta1 — kept here so JSON and YAML output
// surface it when the server populates it.
type CronJobStatus struct {
	Active             []ObjectRef `json:"active,omitempty"`
	LastScheduleTime   string      `json:"lastScheduleTime,omitempty"`
	LastSuccessfulTime string      `json:"lastSuccessfulTime,omitempty"`
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
	return clusteropts.Age(c.Status.LastScheduleTime, now) + " ago"
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
	return clusteropts.Age(c.Metadata.CreationTimestamp, now)
}
