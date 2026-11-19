package webhook

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sync"
	"testing"
)

func TestLegacyMACAdoptionRetainsAddress(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := macvlanPod()
	pod.Labels["applications.app.bytetrade.io/macvlan-init"] = "true"
	app, err := wh.dynamicClient.AppV1alpha1().Applications().Get(t.Context(), "app-space-jellyfin", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	deployment, err := wh.kubeClient.AppsV1().Deployments(pod.Namespace).Get(t.Context(), "jellyfin", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	deployment.Spec.Template.Labels["applications.app.bytetrade.io/macvlan-init"] = "true"
	if _, err = wh.kubeClient.AppsV1().Deployments(pod.Namespace).Update(t.Context(), deployment, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	// An auxiliary workload with the same app label must not make MAC ownership ambiguous.
	auxiliary := deployment.DeepCopy()
	auxiliary.Name, auxiliary.UID = "helper", "helper"
	delete(auxiliary.Spec.Template.Labels, "applications.app.bytetrade.io/macvlan-init")
	if _, err = wh.kubeClient.AppsV1().Deployments(pod.Namespace).Create(t.Context(), auxiliary, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	mac := "02:00:00:00:00:01"
	app.Spec.Settings[overlayMACSetting] = mac
	if _, err = wh.dynamicClient.AppV1alpha1().Applications().Update(t.Context(), app, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err = wh.createOverlayMACAllocation(t.Context(), app, pod.Namespace+"/jellyfin", mac); err != nil {
		t.Fatal(err)
	}
	got, err := wh.ensureOverlayMAC(t.Context(), pod, false)
	if err != nil || got != mac {
		t.Fatalf("MAC = %s, error = %v", got, err)
	}
	claim, err := wh.allocationClient.Resource(overlayMACAllocationGVR).Get(t.Context(), overlayMACKey(mac), metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	key, _, _ := unstructured.NestedString(claim.Object, "spec", "instanceKey")
	if key != pod.Namespace+"/jellyfin/Deployment/jellyfin/singleton" {
		t.Fatalf("instance = %s", key)
	}
}
func TestFixedMACConcurrentFinalValidationReservesOnePod(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := macvlanPod()
	pod.Labels["applications.app.bytetrade.io/macvlan-init"] = "true"
	if _, err := wh.CreateMacvlanInitPatch(macvlanBypassAdmissionRequest(t, pod), pod); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, name := range []string{"first", "second"} {
		copy := pod.DeepCopy()
		copy.Name = name
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- wh.ValidateMacvlanAnnotation(t.Context(), copy, copy.Namespace, false)
		}()
	}
	wg.Wait()
	close(results)
	allowed := 0
	for err := range results {
		if err == nil {
			allowed++
		} else {
			t.Log(err)
		}
	}
	if allowed != 1 {
		t.Fatalf("allowed %d concurrent Pods", allowed)
	}
}
