package job

import (
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
)

// Job is the minimal typed view we decode from both the KubeSphere
// list envelope (`/kapis/resources.kubesphere.io/v1alpha3/jobs`) and
// the K8s native single-resource GET (`/apis/batch/v1/namespaces/
// <ns>/jobs/<name>`).
//
// As with cmd/ctl/cluster/{pod,workload}/types.go, we model only the
// fields actually rendered or summarized by verbs in this package.
// `cluster job yaml` uses raw bytes, not this struct, so unknown
// fields aren't dropped from that output path.
type Job struct {
	APIVersion string      `json:"apiVersion,omitempty"`
	Kind       string      `json:"kind,omitempty"`
	Metadata   JobMetadata `json:"metadata"`
	Spec       JobSpec     `json:"spec"`
	Status     JobStatus   `json:"status"`
}

type JobMetadata struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace,omitempty"`
	UID               string            `json:"uid,omitempty"`
	ResourceVersion   string            `json:"resourceVersion,omitempty"`
	CreationTimestamp string            `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	OwnerReferences   []JobOwnerRef     `json:"ownerReferences,omitempty"`
}

// JobOwnerRef is the minimal subset of metav1.OwnerReference used to
// surface the parent CronJob in `cluster job get`. We only care about
// Kind + Name (controller flag is also included so a non-controller
// owner ref doesn't masquerade as the parent).
type JobOwnerRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Controller bool   `json:"controller,omitempty"`
}

type JobSpec struct {
	// Pointers so we can distinguish "missing" from "explicit zero" —
	// completions=0 (rare, but valid) shouldn't render as "-/-".
	Completions  *int  `json:"completions,omitempty"`
	Parallelism  *int  `json:"parallelism,omitempty"`
	BackoffLimit *int  `json:"backoffLimit,omitempty"`
	Suspend      *bool `json:"suspend,omitempty"`
}

type JobStatus struct {
	Active         int             `json:"active,omitempty"`
	Succeeded      int             `json:"succeeded,omitempty"`
	Failed         int             `json:"failed,omitempty"`
	StartTime      string          `json:"startTime,omitempty"`
	CompletionTime string          `json:"completionTime,omitempty"`
	Conditions     []JobCondition  `json:"conditions,omitempty"`
}

// JobCondition mirrors batchv1.JobCondition. The Complete / Failed
// conditions (Type=Complete, Type=Failed; Status=True) are how the
// list table picks STATUS without re-implementing kubectl's heuristic.
type JobCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// completionsLabel returns the kubectl-style "succeeded/completions"
// string for the COMPLETIONS column. When spec.completions is unset
// (parallel-without-completions style), falls back to just
// "<succeeded>".
func (j Job) completionsLabel() string {
	if j.Spec.Completions != nil {
		return fmt.Sprintf("%d/%d", j.Status.Succeeded, *j.Spec.Completions)
	}
	return fmt.Sprintf("%d", j.Status.Succeeded)
}

// duration returns the wall-clock time the Job ran (or has been
// running). Falls back to "-" when start/completion timestamps are
// missing or unparseable.
//
// While the Job is still running we use now()-start so the column
// keeps moving under --watch. After completion (CompletionTime set)
// we use completion-start so the column freezes at the final value.
func (j Job) duration(now time.Time) string {
	start, err := parseRFC3339(j.Status.StartTime)
	if err != nil {
		return "-"
	}
	end := now
	if j.Status.CompletionTime != "" {
		if c, err := parseRFC3339(j.Status.CompletionTime); err == nil {
			end = c
		}
	}
	d := end.Sub(start)
	if d < 0 {
		d = 0
	}
	return formatDuration(d)
}

// status returns the kubectl-style STATUS label: "Complete" /
// "Failed" / "Suspended" (when spec.suspend=true) / "Running" /
// "Pending". Mirrors how the SPA's `getWorkloadStatus` derives a
// job's state for the tree view.
func (j Job) status() string {
	if j.Spec.Suspend != nil && *j.Spec.Suspend {
		return "Suspended"
	}
	for _, c := range j.Status.Conditions {
		if c.Status != "True" {
			continue
		}
		switch strings.ToLower(c.Type) {
		case "complete":
			return "Complete"
		case "failed":
			return "Failed"
		}
	}
	switch {
	case j.Status.Active > 0:
		return "Running"
	case j.Status.Failed > 0:
		return "Failing"
	default:
		return "Pending"
	}
}

// age returns the AGE column string.
func (j Job) age(now time.Time) string {
	return clusteropts.Age(j.Metadata.CreationTimestamp, now)
}

// parentCronJob returns "<cronjob name>" when this Job is owned by a
// CronJob (Kind=CronJob, Controller=true), or "" otherwise. Surfaced
// in `job get` so users can pivot to `cluster cronjob get` from a
// child Job.
func (j Job) parentCronJob() string {
	for _, o := range j.Metadata.OwnerReferences {
		if o.Controller && o.Kind == "CronJob" {
			return o.Name
		}
	}
	return ""
}

// parseRFC3339 is a thin wrapper that swallows the typed time.Time on
// error. Local helper to keep call sites compact.
func parseRFC3339(ts string) (time.Time, error) {
	if ts == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	return time.Parse(time.RFC3339, ts)
}

// formatDuration prints a duration in kubectl-list shorthand
// (s / m / h / d) — coarse enough to fit one column, precise enough
// to distinguish "started seconds ago" from "ran for hours".
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	default:
		return fmt.Sprintf("%dd%dh", int(d.Hours()/24), int(d.Hours())%24)
	}
}

