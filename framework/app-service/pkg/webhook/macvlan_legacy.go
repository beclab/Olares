package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/retry"
	"strings"
)

// Adopt an old application/ordinal claim only when its workload is unambiguous.
// The claim's UID ownership and resourceVersion prevent cross-application reuse.
func (wh *Webhook) adoptLegacyOverlayMAC(ctx context.Context, pod *corev1.Pod, app *appv1alpha1.Application, instance string) (string, error) {
	base := pod.Namespace + "/" + app.Spec.Name
	legacyKey, mac := base, app.Spec.Settings[overlayMACSetting]
	parts := strings.Split(instance, "/")
	if len(parts) != 5 {
		return "", fmt.Errorf("invalid workload instance")
	}
	if parts[2] == "StatefulSet" {
		var values map[string]string
		raw := app.Spec.Settings["overlayMacvlanMacByOrdinal"]
		if raw == "" {
			return "", nil
		}
		if err := json.Unmarshal([]byte(raw), &values); err != nil {
			return "", err
		}
		mac, legacyKey = values[parts[4]], base+"/"+parts[4]
	}
	if mac == "" {
		return "", nil
	}
	if err := validateOverlayMAC(mac); err != nil {
		return "", err
	}
	// Old keys did not distinguish two workloads of the same kind. Never guess.
	count := 0
	if parts[2] == "Deployment" {
		list, err := wh.kubeClient.AppsV1().Deployments(pod.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return "", err
		}
		for _, d := range list.Items {
			if d.Spec.Template.Labels[constants.ApplicationNameLabel] == app.Spec.Name && d.Spec.Template.Labels["applications.app.bytetrade.io/macvlan-init"] == "true" {
				count++
			}
		}
	} else {
		list, err := wh.kubeClient.AppsV1().StatefulSets(pod.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return "", err
		}
		for _, d := range list.Items {
			if d.Spec.Template.Labels[constants.ApplicationNameLabel] == app.Spec.Name && d.Spec.Template.Labels["applications.app.bytetrade.io/macvlan-init"] == "true" {
				count++
			}
		}
	}
	if count != 1 {
		return "", fmt.Errorf("legacy MAC belongs to %d possible workloads; cannot safely reuse", count)
	}
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		claim, err := wh.allocationClient.Resource(overlayMACAllocationGVR).Get(ctx, overlayMACKey(mac), metav1.GetOptions{})
		if err != nil {
			return err
		}
		key, _, _ := unstructured.NestedString(claim.Object, "spec", "instanceKey")
		if key == instance {
			return validateOverlayMACAllocation(claim, app, instance, mac)
		}
		if err := validateOverlayMACAllocation(claim, app, legacyKey, mac); err != nil {
			return err
		}
		if err := unstructured.SetNestedField(claim.Object, instance, "spec", "instanceKey"); err != nil {
			return err
		}
		_, err = wh.allocationClient.Resource(overlayMACAllocationGVR).Update(ctx, claim, metav1.UpdateOptions{})
		return err
	})
	if err != nil {
		return "", err
	}
	if err := wh.persistOverlayMAC(ctx, app.Name, app.UID, instance, true, mac); err != nil {
		return "", err
	}
	if err := wh.ensureOverlayMACFinalizer(ctx, app.Name, app.UID); err != nil {
		return "", err
	}
	if err := wh.markOverlayMACAllocationPhase(ctx, mac, overlayMACAllocationPhase); err != nil {
		return "", err
	}
	return mac, nil
}
