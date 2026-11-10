package helmops

import (
	"context"
	"fmt"
	"strings"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	utilwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PodStopGrace is how long a running pod that nobody has asked to
// terminate may hold up a stop. Restored from
// pkg/appstate/suspending_app.go — a var so tests can reach the
// fast-fail branch.
var PodStopGrace = 2 * time.Minute

// podStopBlocker names a pod keeping a stop from completing, plus
// whatever created it.
type podStopBlocker struct {
	name  string
	owner string
}

// ListAppPods returns every pod in the app's namespace. Olares gives
// each app its own namespace. Restored from pkg/appstate/utils.go.
func ListAppPods(am *appv1alpha1.ApplicationManager, cli client.Client) ([]corev1.Pod, error) {
	var pods corev1.PodList
	if err := cli.List(context.TODO(), &pods, client.InNamespace(am.Spec.AppNamespace)); err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// WaitForPodsGone blocks until the app namespace has no non-terminal
// pods, or a non-terminating blocker outlives PodStopGrace.
// Restored from pkg/appstate/suspending_app.go:waitForPodsGone.
func WaitForPodsGone(ctx context.Context, am *appv1alpha1.ApplicationManager, cli client.Client) error {
	start := time.Now()
	return utilwait.PollImmediate(time.Second, 30*time.Minute, func() (bool, error) {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		pods, err := ListAppPods(am, cli)
		if err != nil {
			klog.Errorf("list pods while suspending app %s failed %v", am.Spec.AppName, err)
			return false, err
		}
		terminating, blockers := classifyPodsForStop(pods)
		if terminating == 0 && len(blockers) == 0 {
			return true, nil
		}
		if len(blockers) > 0 && time.Since(start) >= PodStopGrace {
			return false, fmt.Errorf("suspend app %s blocked by pod(s) that are not terminating in namespace %s: %s",
				am.Spec.AppName, am.Spec.AppNamespace, formatPodStopBlockers(blockers))
		}
		klog.Infof("suspend app %s waiting for %d terminating and %d running pod(s) in namespace %s",
			am.Spec.AppName, terminating, len(blockers), am.Spec.AppNamespace)
		return false, nil
	})
}

// classifyPodsForStop splits an app namespace's pods into the ones a
// stop must still wait for and the ones that will not leave on their
// own. Restored from pkg/appstate/suspending_app.go.
func classifyPodsForStop(pods []corev1.Pod) (terminating int, blockers []podStopBlocker) {
	for i := range pods {
		pod := &pods[i]
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		if pod.DeletionTimestamp != nil {
			terminating++
			continue
		}
		blockers = append(blockers, podStopBlocker{name: pod.Name, owner: describePodOwner(pod)})
	}
	return terminating, blockers
}

func describePodOwner(pod *corev1.Pod) string {
	for _, ref := range pod.OwnerReferences {
		if ref.Controller != nil && *ref.Controller {
			return ref.Kind + "/" + ref.Name
		}
	}
	if len(pod.OwnerReferences) > 0 {
		return pod.OwnerReferences[0].Kind + "/" + pod.OwnerReferences[0].Name
	}
	return "no owner"
}

func formatPodStopBlockers(blockers []podStopBlocker) string {
	parts := make([]string, 0, len(blockers))
	for _, b := range blockers {
		parts = append(parts, fmt.Sprintf("%s (owned by %s)", b.name, b.owner))
	}
	return strings.Join(parts, ", ")
}
