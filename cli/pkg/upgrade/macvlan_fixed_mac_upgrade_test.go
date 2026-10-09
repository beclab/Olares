package upgrade

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func fixedMACFixturePod(name, uid string) *corev1.Pod {
	yes := true
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "apps", UID: types.UID(uid), Labels: map[string]string{fixedMACLabel: "true", "applications.app.bytetrade.io/name": "media"}, Annotations: map[string]string{fixedMACNetworks: `[{"name":"underlay-macvlan","namespace":"kube-system","mac":"02:00:00:00:00:01","interface":"net1"}]`, "k8s.v1.cni.cncf.io/network-status": `[{"interface":"net1","mac":"02:00:00:00:00:01","ips":["192.168.1.20"]}]`}, OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "StatefulSet", Name: "media", UID: "sts", Controller: &yes}}}, Spec: corev1.PodSpec{NodeName: "master"}, Status: corev1.PodStatus{Phase: corev1.PodRunning, Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}}}
}
func fixedMACDynamic() *dynamicfake.FakeDynamicClient {
	claim := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "app.bytetrade.io/v1alpha1", "kind": "OverlayMACAllocation", "metadata": map[string]interface{}{"name": "020000000001"}, "spec": map[string]interface{}{"phase": "Bound", "applicationRef": "apps-media", "applicationUID": "app"}}}
	app := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "app.bytetrade.io/v1alpha1", "kind": "Application", "metadata": map[string]interface{}{"name": "apps-media", "uid": "app"}, "spec": map[string]interface{}{"namespace": "apps", "name": "media"}}}
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{fixedMACAllocations: "OverlayMACAllocationList", fixedMACNAD: "NetworkAttachmentDefinitionList", fixedMACApplications: "ApplicationList"}, claim, app)
}
func TestFixedMACWaitDoesNotTreatEmptyListAsSuccess(t *testing.T) {
	kube := kubefake.NewSimpleClientset()
	item := fixedMACInstance{Namespace: "apps", Name: "media-0", OldUID: "old", Kind: "StatefulSet", Workload: "media", WorkloadUID: "sts"}
	if pod, err := replacementFixedMAC(t.Context(), kube, fixedMACDynamic(), item); err != nil || pod != nil {
		t.Fatalf("empty list: %v %v", pod, err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := recoverFixedMAC(ctx, kube, fixedMACDynamic(), &fixedMACProgress{Instances: []fixedMACInstance{item}}, func(*fixedMACProgress) error { return nil }); err == nil {
		t.Fatal("missing replacement reported success")
	}
}
func TestFixedMACWaitRequiresNewReadyPodAndLedger(t *testing.T) {
	old := fixedMACFixturePod("media-0", "old")
	kube := kubefake.NewSimpleClientset(old)
	dc := fixedMACDynamic()
	item := fixedMACInstance{Namespace: "apps", Name: old.Name, OldUID: old.UID, Kind: "StatefulSet", WorkloadUID: "sts"}
	if got, _ := replacementFixedMAC(t.Context(), kube, dc, item); got != nil {
		t.Fatal("old pod accepted")
	}
	kube.CoreV1().Pods("apps").Delete(t.Context(), old.Name, metav1.DeleteOptions{})
	fresh := fixedMACFixturePod("media-0", "new")
	fresh.Status.Conditions = nil
	kube.CoreV1().Pods("apps").Create(t.Context(), fresh, metav1.CreateOptions{})
	if got, _ := replacementFixedMAC(t.Context(), kube, dc, item); got != nil {
		t.Fatal("unready pod accepted")
	}
	fresh.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
	kube.CoreV1().Pods("apps").Update(t.Context(), fresh, metav1.UpdateOptions{})
	if got, err := replacementFixedMAC(t.Context(), kube, dc, item); err != nil || got == nil || got.UID != "new" {
		t.Fatalf("new ready pod: %v %v", got, err)
	}
	fresh.Annotations["k8s.v1.cni.cncf.io/network-status"] = `[{"interface":"net1","mac":"02:00:00:00:00:02","ips":["192.168.1.20"]}]`
	if ready, _ := fixedMACPodReady(t.Context(), dc, fresh); ready {
		t.Fatal("wrong runtime MAC accepted")
	}
}
func TestFixedMACRecoveryEvictsWithUIDAndPersistsEachReplacement(t *testing.T) {
	old := fixedMACFixturePod("media-0", "old")
	kube := kubefake.NewSimpleClientset(old)
	dc := fixedMACDynamic()
	kube.PrependReactor("create", "pods", func(a ktesting.Action) (bool, runtime.Object, error) {
		if a.GetSubresource() != "eviction" {
			return false, nil, nil
		}
		raw, _ := json.Marshal(a.(ktesting.CreateAction).GetObject())
		var body struct {
			DeleteOptions metav1.DeleteOptions `json:"deleteOptions"`
		}
		json.Unmarshal(raw, &body)
		if body.DeleteOptions.Preconditions == nil || *body.DeleteOptions.Preconditions.UID != "old" {
			t.Fatal("eviction has no old UID precondition")
		}
		kube.Tracker().Delete(corev1.SchemeGroupVersion.WithResource("pods"), "apps", old.Name)
		kube.Tracker().Create(corev1.SchemeGroupVersion.WithResource("pods"), fixedMACFixturePod(old.Name, "new"), "apps")
		return true, nil, nil
	})
	p := &fixedMACProgress{Instances: []fixedMACInstance{{Namespace: "apps", Name: old.Name, OldUID: "old", Kind: "StatefulSet", WorkloadUID: "sts"}}}
	saved := 0
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := recoverFixedMAC(ctx, kube, dc, p, func(*fixedMACProgress) error { saved++; return nil }); err != nil {
		t.Fatal(err)
	}
	if saved != 1 || p.Instances[0].ReplacementUID != "new" {
		t.Fatalf("progress: %+v saves=%d", p, saved)
	}
}
func TestFixedMACSnapshotRejectsUnmanagedBeforeDeletion(t *testing.T) {
	pod := fixedMACFixturePod("media-0", "old")
	pod.OwnerReferences = nil
	if _, err := snapshotFixedMAC(t.Context(), kubefake.NewSimpleClientset(pod), fixedMACDynamic(), "master"); err == nil {
		t.Fatal("unmanaged pod accepted")
	}
}
func TestFixedMACSnapshotRecordsControllerIdentity(t *testing.T) {
	pod := fixedMACFixturePod("media-0", "old")
	sts := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "media", Namespace: "apps", UID: "sts"}, Spec: appsv1.StatefulSetSpec{Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{fixedMACLabel: "true"}}}}}
	p, err := snapshotFixedMAC(t.Context(), kubefake.NewSimpleClientset(pod, sts), fixedMACDynamic(), "master")
	if err != nil || len(p.Instances) != 1 || p.Instances[0].WorkloadUID != "sts" {
		t.Fatalf("snapshot: %v %v", p, err)
	}
}
func TestFixedMACNADMigratesToWiredParent(t *testing.T) {
	dc := fixedMACDynamic()
	nad := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "k8s.cni.cncf.io/v1", "kind": "NetworkAttachmentDefinition", "metadata": map[string]interface{}{"name": "underlay-macvlan", "namespace": "kube-system"}, "spec": map[string]interface{}{"config": `{"type":"macvlan","master":"br-olares","ipam":{"type":"dhcp","omitDefaultGateway":true}}`}}}
	dc.Resource(fixedMACNAD).Namespace("kube-system").Create(t.Context(), nad, metav1.CreateOptions{})
	if err := upgradeFixedMACNAD(t.Context(), dc); err != nil {
		t.Fatal(err)
	}
	got, _ := dc.Resource(fixedMACNAD).Namespace("kube-system").Get(t.Context(), "underlay-macvlan", metav1.GetOptions{})
	raw, _, _ := unstructured.NestedString(got.Object, "spec", "config")
	var conf map[string]interface{}
	json.Unmarshal([]byte(raw), &conf)
	if conf["master"] != "olares-lan" || conf["ipam"].(map[string]interface{})["sendRelease"] != false {
		t.Fatalf("NAD: %s", raw)
	}
}
func TestFixedMACPlansDesiredInstanceWithoutAnyPod(t *testing.T) {
	dc := fixedMACDynamic()
	app, err := dc.Resource(fixedMACApplications).Get(t.Context(), "apps-media", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	unstructured.SetNestedField(app.Object, "true", "spec", "settings", "enableOverlayGateway")
	dc.Resource(fixedMACApplications).Update(t.Context(), app, metav1.UpdateOptions{})
	one := int32(1)
	sts := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "media", Namespace: "apps", UID: "sts"}, Spec: appsv1.StatefulSetSpec{Replicas: &one, Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{fixedMACLabel: "true", "applications.app.bytetrade.io/name": "media"}}}}}
	kube := kubefake.NewSimpleClientset(sts)
	p, err := snapshotFixedMAC(t.Context(), kube, dc, "master")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Instances) != 1 || p.Instances[0].Name != "media-0" || p.Instances[0].OldUID != "" {
		t.Fatalf("missing expected instance: %+v", p)
	}
	if pod, err := replacementFixedMAC(t.Context(), kube, dc, p.Instances[0]); pod != nil || err != nil {
		t.Fatalf("expected pending instance: %v %v", pod, err)
	}
}
