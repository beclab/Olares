package webhook

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	multusNetworksAnnotation       = "k8s.v1.cni.cncf.io/networks"
	multusDefaultNetworkAnnotation = "v1.multus-cni.io/default-network"

	// Event reasons surfaced to the user when the platform changes a pod's
	// network selection during admission.
	EventReasonOverlayGatewayDisabled   = "OverlayGatewayDisabled"
	EventReasonDefaultNetworkNotAllowed = "DefaultNetworkNotAllowed"
	EventReasonOverlayNetworkRebuilt    = "OverlayNetworkRebuilt"
)

// macvlanSelectionAnnotations lists every annotation Multus reads to pick a
// pod's networks. Both keys are platform-owned in user namespaces.
var macvlanSelectionAnnotations = []string{multusNetworksAnnotation, multusDefaultNetworkAnnotation}

// HasMacvlanSelectionAnnotations reports whether the pod carries any Multus
// network selection, regardless of how it is written.
func HasMacvlanSelectionAnnotations(pod *corev1.Pod) bool {
	if pod == nil || pod.Annotations == nil {
		return false
	}
	for _, key := range macvlanSelectionAnnotations {
		if _, ok := pod.Annotations[key]; ok {
			return true
		}
	}
	return false
}

// MacvlanSelectionAnnotationsChanged reports whether either selection
// annotation differs between two pod revisions. Unrelated updates (finalizers,
// labels, ownership fixes) must not re-run the selection checks, otherwise
// pods created before this check existed could never be updated again.
func MacvlanSelectionAnnotationsChanged(oldPod, newPod *corev1.Pod) bool {
	for _, key := range macvlanSelectionAnnotations {
		oldValue, oldOK := annotationValue(oldPod, key)
		newValue, newOK := annotationValue(newPod, key)
		if oldOK != newOK || oldValue != newValue {
			return true
		}
	}
	return false
}

func annotationValue(pod *corev1.Pod, key string) (string, bool) {
	if pod == nil || pod.Annotations == nil {
		return "", false
	}
	value, ok := pod.Annotations[key]
	return value, ok
}

// stripMacvlanSelectionAnnotations deletes every network selection from the pod
// and returns the keys that were present.
func stripMacvlanSelectionAnnotations(pod *corev1.Pod) []string {
	if pod == nil || pod.Annotations == nil {
		return nil
	}
	removed := make([]string, 0, len(macvlanSelectionAnnotations))
	for _, key := range macvlanSelectionAnnotations {
		if _, ok := pod.Annotations[key]; ok {
			delete(pod.Annotations, key)
			removed = append(removed, key)
		}
	}
	return removed
}

// ValidateMacvlanAnnotation compares the pod's final network selection with
// what the platform itself would have generated. It deliberately never parses
// the user-supplied value: the only two acceptable states are "exactly the
// platform selection" (Overlay Gateway enabled) and "no selection at all".
// Anything else means the mutating step was bypassed or overwritten.
func (wh *Webhook) ValidateMacvlanAnnotation(ctx context.Context, pod *corev1.Pod, ns string, dryRun bool) error {
	if pod == nil {
		return nil
	}
	if _, ok := annotationValue(pod, multusDefaultNetworkAnnotation); ok {
		return errors.New("replacing the pod default network is not allowed")
	}
	shouldInject, err := wh.ShouldInjectMacvlanInit(ctx, pod, ns)
	if err != nil {
		klog.Errorf("macvlan-guard: failed to evaluate Overlay Gateway for pod=%s/%s err=%v", ns, pod.Name, err)
		return err
	}
	actual, present := annotationValue(pod, multusNetworksAnnotation)
	if !shouldInject {
		if present {
			return errors.New("underlay macvlan selection requires Overlay Gateway to be enabled for the application")
		}
		return nil
	}
	if dryRun {
		// A dry-run admission never persists a MAC, so there is no ledger
		// entry to compare against; the mutating step already rebuilt the
		// selection for the caller to inspect.
		return nil
	}
	mac, err := wh.boundOverlayMAC(ctx, pod)
	if err != nil {
		klog.Errorf("macvlan-guard: failed to resolve bound overlay MAC for pod=%s/%s err=%v", ns, pod.Name, err)
		return fmt.Errorf("resolve bound overlay MAC: %w", err)
	}
	if expected := platformMacvlanSelection(mac); !present || actual != expected {
		return errors.New("underlay macvlan selection does not match the platform-generated selection")
	}
	return nil
}
