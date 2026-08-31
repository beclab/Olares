package webhook

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func macvlanPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jellyfin-pod",
			Namespace: "app-space",
			Labels: map[string]string{
				constants.ApplicationNameLabel: "jellyfin",
			},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "jellyfin"}}},
	}
}

func TestGenerateOverlayMACIsLocalUnicast(t *testing.T) {
	mac, err := generateOverlayMAC()
	if err != nil {
		t.Fatalf("generateOverlayMAC: %v", err)
	}
	if err := validateOverlayMAC(mac); err != nil {
		t.Fatalf("validateOverlayMAC(%q): %v", mac, err)
	}
}

func TestCreateMacvlanInitPatchPersistsAndReusesMAC(t *testing.T) {
	wh := testMacvlanWebhook()
	first := macvlanPod()
	req := macvlanBypassAdmissionRequest(t, first)
	if _, err := wh.CreateMacvlanInitPatch(req, first); err != nil {
		t.Fatalf("first patch: %v", err)
	}

	var firstNetworks []map[string]interface{}
	if err := json.Unmarshal([]byte(first.Annotations["k8s.v1.cni.cncf.io/networks"]), &firstNetworks); err != nil {
		t.Fatalf("decode first networks annotation: %v", err)
	}
	if len(firstNetworks) != 1 {
		t.Fatalf("unexpected first network selection: %#v", firstNetworks)
	}
	firstMAC, _ := firstNetworks[0]["mac"].(string)
	if !strings.HasPrefix(firstMAC, "02:") {
		t.Fatalf("unexpected first network MAC: %#v", firstNetworks[0])
	}

	second := macvlanPod()
	second.Name = "jellyfin-recreated"
	if _, err := wh.CreateMacvlanInitPatch(macvlanBypassAdmissionRequest(t, second), second); err != nil {
		t.Fatalf("second patch: %v", err)
	}
	var secondNetworks []map[string]interface{}
	if err := json.Unmarshal([]byte(second.Annotations["k8s.v1.cni.cncf.io/networks"]), &secondNetworks); err != nil {
		t.Fatalf("decode second networks annotation: %v", err)
	}
	if len(secondNetworks) != 1 {
		t.Fatalf("unexpected second network selection: %#v", secondNetworks)
	}
	secondMAC, _ := secondNetworks[0]["mac"].(string)
	if got, want := secondMAC, firstMAC; got != want {
		t.Fatalf("recreated pod MAC = %q, want %q", got, want)
	}
}

func TestCreateMacvlanInitPatchDryRunHasNoAllocationSideEffect(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := macvlanPod()
	dryRun := true
	req := macvlanBypassAdmissionRequest(t, pod)
	req.DryRun = &dryRun
	if _, err := wh.CreateMacvlanInitPatch(req, pod); err != nil {
		t.Fatalf("dry-run patch: %v", err)
	}

	app, err := wh.dynamicClient.AppV1alpha1().Applications().Get(t.Context(), "app-space-jellyfin", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get application: %v", err)
	}
	if app.Spec.Settings[overlayMACSetting] != "" {
		t.Fatalf("dry-run persisted MAC %q", app.Spec.Settings[overlayMACSetting])
	}
	if _, err := wh.allocationClient.Resource(overlayMACAllocationGVR).Get(t.Context(), "does-not-exist", metav1.GetOptions{}); err == nil {
		t.Fatal("dry-run unexpectedly created an allocation")
	}
}

func TestMacvlanNetworkAnnotationPreservesExistingSelections(t *testing.T) {
	existing := `[{"name":"other-net","interface":"net2","cni-args":{"foo":"bar"}},{"name":"underlay-macvlan","interface":"net1"}]`
	raw, err := macvlanNetworkAnnotation(existing, "02:00:00:00:00:01")
	if err != nil {
		t.Fatalf("macvlanNetworkAnnotation: %v", err)
	}
	var selections []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &selections); err != nil {
		t.Fatalf("decode network selections: %v", err)
	}
	if got := selections[0]["interface"]; got != "net2" {
		t.Fatalf("existing interface = %v, want net2", got)
	}
	if got := selections[0]["cni-args"].(map[string]interface{})["foo"]; got != "bar" {
		t.Fatalf("existing cni-args = %v, want bar", got)
	}
	if got := selections[1]["mac"]; got != "02:00:00:00:00:01" {
		t.Fatalf("underlay MAC = %v", got)
	}
}

func TestCreateMacvlanInitPatchAddsMasterPlacement(t *testing.T) {
	wh := testMacvlanWebhook()
	pod := macvlanPod()
	if _, err := wh.CreateMacvlanInitPatch(macvlanBypassAdmissionRequest(t, pod), pod); err != nil {
		t.Fatalf("patch: %v", err)
	}
	required := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if required == nil {
		t.Fatal("expected required node affinity")
	}
	found := false
	for _, requirement := range required.NodeSelectorTerms[0].MatchExpressions {
		if requirement.Key == overlayMACMasterNodeLabel && requirement.Operator == corev1.NodeSelectorOpExists {
			found = true
		}
	}
	if !found {
		t.Fatalf("master placement requirement missing: %#v", required.NodeSelectorTerms)
	}
}

func TestOverlayMACRejectsMultiReplicaDeployment(t *testing.T) {
	wh := testMacvlanWebhook()
	controller := true
	replicas := int32(2)
	wh.kubeClient = fakeKubeClient(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "jellyfin", Namespace: "app-space"},
			Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "jellyfin-rs",
				Namespace: "app-space",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "jellyfin",
					Controller: &controller,
				}},
			},
		},
	)
	pod := macvlanPod()
	pod.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: "apps/v1",
		Kind:       "ReplicaSet",
		Name:       "jellyfin-rs",
		Controller: &controller,
	}}
	_, err := wh.CreateMacvlanInitPatch(macvlanBypassAdmissionRequest(t, pod), pod)
	if err == nil || !strings.Contains(err.Error(), "stable per-replica identity") {
		t.Fatalf("expected multi-replica rejection, got %v", err)
	}
}

func TestOverlayMACRejectsConflictingAllocationOwner(t *testing.T) {
	wh := testMacvlanWebhook()
	app, err := wh.dynamicClient.AppV1alpha1().Applications().Get(t.Context(), "app-space-jellyfin", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get application: %v", err)
	}
	mac := "02:00:00:00:00:01"
	app.Spec.Settings[overlayMACSetting] = mac
	if _, err := wh.dynamicClient.AppV1alpha1().Applications().Update(t.Context(), app, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("update application: %v", err)
	}
	_, err = wh.allocationClient.Resource(overlayMACAllocationGVR).Create(t.Context(), &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "app.bytetrade.io/v1alpha1",
			"kind":       "OverlayMACAllocation",
			"metadata":   map[string]interface{}{"name": overlayMACKey(mac)},
			"spec": map[string]interface{}{
				"mac":            mac,
				"instanceKey":    "other/instance",
				"applicationUID": "other-uid",
				"applicationRef": "other-app",
				"phase":          overlayMACAllocationPhase,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create conflicting allocation: %v", err)
	}
	_, err = wh.CreateMacvlanInitPatch(macvlanBypassAdmissionRequest(t, macvlanPod()), macvlanPod())
	if err == nil || !strings.Contains(err.Error(), "owned by another instance") {
		t.Fatalf("expected allocation owner rejection, got %v", err)
	}
}

func TestOverlayMACRejectsInvalidPersistedValue(t *testing.T) {
	wh := testMacvlanWebhook()
	app, err := wh.dynamicClient.AppV1alpha1().Applications().Get(t.Context(), "app-space-jellyfin", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get application: %v", err)
	}
	app.Spec.Settings[overlayMACSetting] = "02:not-a-mac"
	if _, err := wh.dynamicClient.AppV1alpha1().Applications().Update(t.Context(), app, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("update application: %v", err)
	}
	_, err = wh.CreateMacvlanInitPatch(macvlanBypassAdmissionRequest(t, macvlanPod()), macvlanPod())
	if err == nil || !strings.Contains(err.Error(), "invalid overlay MAC") {
		t.Fatalf("expected invalid persisted MAC rejection, got %v", err)
	}
}

func fakeKubeClient(objects ...k8sruntime.Object) kubernetes.Interface {
	return k8sfake.NewSimpleClientset(objects...)
}
