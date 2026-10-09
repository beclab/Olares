package utils

import (
	"context"
	"errors"
	"fmt"
	"sync"

	nadutils "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	overlayLANInterface   = "net1"
	overlayMacvlanInitKey = "applications.app.bytetrade.io/macvlan-init"
)

// healedOverlayPods remembers which Pods this process already restarted so a
// Pod that keeps coming up without net1 is not restarted in a loop.
var healedOverlayPods sync.Map

var errOverlayHealingDeferred = errors.New("overlay upgrade recovery is still in progress")

// overlayPodMissingNet1 reports whether a running overlay Pod has no LAN
// interface recorded by Multus. Pods still starting are left alone.
func overlayPodMissingNet1(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning || pod.DeletionTimestamp != nil {
		return false
	}
	if pod.Labels[overlayMacvlanInitKey] != "true" {
		return false
	}
	statuses, err := nadutils.GetNetworkStatus(pod)
	if err != nil {
		return true
	}
	for _, s := range statuses {
		if s.Interface == overlayLANInterface {
			return false
		}
	}
	return true
}

// HealOverlayPodsWithoutNet1 restarts, once per Pod, the running overlay Pods
// of enabled applications that have no net1. It returns how many were deleted.
func HealOverlayPodsWithoutNet1(ctx context.Context) (int, error) {
	client, err := GetKubeClient()
	if err != nil {
		return 0, err
	}
	return healOverlayPods(ctx, client, GetOverlayGatewaySupportedApps)
}

func healOverlayPods(ctx context.Context, client kubernetes.Interface, getApps func(context.Context, string) ([]OverlayGatewaySupportedApp, error)) (int, error) {
	gate, err := client.CoreV1().ConfigMaps("kube-system").Get(ctx, "olares-fixed-mac-upgrade", metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return 0, err
	}
	if err == nil && (gate.Data["phase"] == "paused" || gate.Data["phase"] == "recovering") {
		klog.Infof("overlay-heal: upgrade owns Pod recovery; defer background healing")
		return 0, errOverlayHealingDeferred
	}
	apps, err := getApps(ctx, "")
	if err != nil {
		return 0, err
	}
	restarted := 0
	var failures []error
	for _, app := range apps {
		if !app.Enabled {
			continue
		}
		pods, err := client.CoreV1().Pods(app.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return restarted, err
		}
		for i := range pods.Items {
			pod := &pods.Items[i]
			if !overlayPodMissingNet1(pod) || metav1.GetControllerOf(pod) == nil {
				continue
			}
			if _, seen := healedOverlayPods.LoadOrStore(types.UID(pod.UID), struct{}{}); seen {
				continue
			}
			klog.Warningf("overlay-heal: pod %s/%s runs without %s, restarting it once", pod.Namespace, pod.Name, overlayLANInterface)
			if err := client.PolicyV1().Evictions(pod.Namespace).Evict(ctx, &policyv1.Eviction{ObjectMeta: metav1.ObjectMeta{Name: pod.Name, Namespace: pod.Namespace}, DeleteOptions: &metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &pod.UID}}}); err != nil {
				healedOverlayPods.Delete(types.UID(pod.UID))
				if !apierrors.IsNotFound(err) {
					failures = append(failures, fmt.Errorf("evict %s/%s: %w", pod.Namespace, pod.Name, err))
				}
				klog.Errorf("overlay-heal: evict pod %s/%s failed: %v", pod.Namespace, pod.Name, err)
				continue
			}
			restarted++
		}
	}
	return restarted, errors.Join(failures...)
}
