package market

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/clusterclient"
)

// app-service's admission webhook stamps these on every Deployment and
// StatefulSet pod template of an app.
const (
	appNameLabel  = "applications.app.bytetrade.io/name"
	appOwnerLabel = "applications.app.bytetrade.io/owner"
)

// podProbeInterval floors how often --watch reads pods: the market row is
// polled every --watch-interval, but a pod read is a cross-namespace list.
const podProbeInterval = 5 * time.Second

// fatalWaitingReasons are the waiting reasons kubelet will not get past
// without a change to the chart or its inputs. Image pulls are left out:
// a slow registry also shows ErrImagePull / ImagePullBackOff for a while.
var fatalWaitingReasons = map[string]bool{
	"CrashLoopBackOff":           true,
	"CreateContainerConfigError": true,
	"CreateContainerError":       true,
	"RunContainerError":          true,
	"InvalidImageName":           true,
}

var notableWaitingReasons = map[string]bool{
	"ErrImagePull":     true,
	"ImagePullBackOff": true,
}

// podSignal is one container that is restarting or stuck.
type podSignal struct {
	Namespace string
	Pod       string
	Container string
	Reason    string
	Restarts  int
}

func (s podSignal) fatal() bool { return fatalWaitingReasons[s.Reason] }

func (s podSignal) String() string {
	reason := s.Reason
	if reason == "" {
		reason = "restarting"
	}
	return fmt.Sprintf("pod %s container %s: %s (restarts=%d)", s.Pod, s.Container, reason, s.Restarts)
}

type podProber func(ctx context.Context, appName string) ([]podSignal, error)

type probedPod struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Status struct {
		InitContainerStatuses []probedContainer `json:"initContainerStatuses"`
		ContainerStatuses     []probedContainer `json:"containerStatuses"`
	} `json:"status"`
}

type probedContainer struct {
	Name         string `json:"name"`
	RestartCount int    `json:"restartCount"`
	State        struct {
		Waiting *struct {
			Reason string `json:"reason"`
		} `json:"waiting"`
	} `json:"state"`
}

func newClusterPodProber(c *clusterclient.Client) podProber {
	owner, _, _ := strings.Cut(c.OlaresID(), "@")
	return func(ctx context.Context, appName string) ([]podSignal, error) {
		selector := appNameLabel + "=" + appName
		if owner != "" {
			selector += "," + appOwnerLabel + "=" + owner
		}
		path := "/kapis/resources.kubesphere.io/v1alpha3/pods?" + url.Values{"labelSelector": {selector}}.Encode()
		list, err := clusterclient.GetKubeSphereList[probedPod](ctx, c, path)
		if err != nil {
			return nil, err
		}
		return podSignals(list.Items), nil
	}
}

func podSignals(pods []probedPod) []podSignal {
	var out []podSignal
	for _, p := range pods {
		statuses := append(append([]probedContainer{}, p.Status.InitContainerStatuses...), p.Status.ContainerStatuses...)
		for _, c := range statuses {
			reason := ""
			if c.State.Waiting != nil {
				reason = c.State.Waiting.Reason
			}
			if c.RestartCount == 0 && !fatalWaitingReasons[reason] && !notableWaitingReasons[reason] {
				continue
			}
			if !fatalWaitingReasons[reason] && !notableWaitingReasons[reason] {
				reason = ""
			}
			out = append(out, podSignal{Namespace: p.Metadata.Namespace, Pod: p.Metadata.Name, Container: c.Name, Reason: reason, Restarts: c.RestartCount})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Pod != out[j].Pod {
			return out[i].Pod < out[j].Pod
		}
		return out[i].Container < out[j].Container
	})
	return out
}

func summarizePodSignals(signals []podSignal) string {
	parts := make([]string, len(signals))
	for i, s := range signals {
		parts[i] = s.String()
	}
	return strings.Join(parts, "; ")
}

// podWatch runs alongside waitForTerminal for install / upgrade. It reports
// each container whose reason or restart count changed, keeps the latest
// signals for the timeout message (stderr lines are suppressed under
// -o json), and stops probing after the first read error.
type podWatch struct {
	probe    podProber
	appName  string
	failFast bool
	next     time.Time
	seen     map[string]podSignal
	latest   []podSignal
	disabled bool
}

func (w *podWatch) poll(ctx context.Context, opts *MarketOptions, now time.Time) *podSignal {
	if w == nil || w.disabled || now.Before(w.next) {
		return nil
	}
	w.next = now.Add(podProbeInterval)
	signals, err := w.probe(ctx, w.appName)
	if err != nil {
		w.disabled = true
		opts.info("watch: cannot read the app's pods (%v); watching the market row only — inspect with `olares-cli cluster pod list -l %s=%s`", err, appNameLabel, w.appName)
		return nil
	}
	if w.seen == nil {
		w.seen = map[string]podSignal{}
	}
	w.latest = signals
	var fatal *podSignal
	for i, s := range signals {
		key := s.Pod + "/" + s.Container
		if prev, ok := w.seen[key]; !ok || prev != s {
			opts.info("[%s] %s", w.appName, s)
		}
		w.seen[key] = s
		if fatal == nil && s.fatal() {
			fatal = &signals[i]
		}
	}
	if w.failFast {
		return fatal
	}
	return nil
}

func (w *podWatch) summary() string {
	if w == nil || len(w.latest) == 0 {
		return ""
	}
	return summarizePodSignals(w.latest)
}

// watchCrashLoopError ends an install / upgrade watch under
// --fail-on-crashloop: the row is still in flight, but a container is in a
// state it will not leave without a chart change.
type watchCrashLoopError struct {
	target watchTarget
	row    statusRow
	signal podSignal
}

func (e *watchCrashLoopError) Error() string {
	return fmt.Sprintf("%s '%s' stopped watching: %s while the row is still %s; the operation is still in flight — read the logs with `olares-cli cluster container logs %s/%s/%s --previous`, fix the chart, then clear the row (`olares-cli market uninstall %s` while installing, `olares-cli market cancel %s` otherwise)",
		e.target.op, e.target.appName, e.signal, valueOrUnknown(e.row.State),
		e.signal.Namespace, e.signal.Pod, e.signal.Container, e.target.appName, e.target.appName)
}
