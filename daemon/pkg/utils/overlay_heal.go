package utils

import (
	"context"
	"sync"

	nadutils "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

const (
	overlayLANInterface   = "net1"
	overlayMacvlanInitKey = "applications.app.bytetrade.io/macvlan-init"
)

// healedOverlayPods remembers which Pods this process already restarted so a
// Pod that keeps coming up without net1 is not restarted in a loop.
var healedOverlayPods sync.Map

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
	apps, err := GetOverlayGatewaySupportedApps(ctx, "")
	if err != nil {
		return 0, err
	}
	restarted := 0
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
			if !overlayPodMissingNet1(pod) {
				continue
			}
			if _, seen := healedOverlayPods.LoadOrStore(types.UID(pod.UID), struct{}{}); seen {
				continue
			}
			klog.Warningf("overlay-heal: pod %s/%s runs without %s, restarting it once", pod.Namespace, pod.Name, overlayLANInterface)
			if err := client.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
				klog.Errorf("overlay-heal: delete pod %s/%s failed: %v", pod.Namespace, pod.Name, err)
				continue
			}
			restarted++
		}
	}
	return restarted, nil
}
