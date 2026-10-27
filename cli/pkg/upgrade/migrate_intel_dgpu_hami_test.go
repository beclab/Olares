package upgrade

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testCard0UUID = "intel-1a2b3c4d-0000-03-00.0"
	testCard1UUID = "intel-1a2b3c4d-0000-83-00.0"
)

func intelHAMINode(registered bool) *corev1.Node {
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1", Annotations: map[string]string{
		nodeIntelRegisterKey: "igpu,card2,xe,,,,0:dgpu,card0,xe,B70,Xe2,Battlemage,34359738368:dgpu,card1,xe,B70,Xe2,Battlemage,34359738368",
		shareModeAnnotationPrefix + "n1-intel-gpu-card0": "0",
	}}}
	if registered {
		node.Annotations[nodeIntelHAMIRegisterKey] = testCard0UUID + ",50,32768,100,B70,0,true,0,vxpu,0:" + testCard1UUID + ",50,32768,100,B70,1,true,1,vxpu,0:"
		node.Annotations[nodeIntelCardsKey] = `[{"uuid":"` + testCard0UUID + `","card":"card0","bdf":"0000:03:00.0"},{"uuid":"` + testCard1UUID + `","card":"card1","bdf":"0000:83:00.0"}]`
	}
	return node
}

func appWorkload(namespace, name, app, owner string, resourceName corev1.ResourceName) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{appNameLabelKey: app, appOwnerLabelKey: owner}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{
				Name:      "main",
				Resources: corev1.ResourceRequirements{Limits: corev1.ResourceList{resourceName: resource.MustParse("1")}},
			}}},
		}},
	}
}

func allocationConfigMap(t *testing.T, allocations []migratedComputeAllocation) *corev1.ConfigMap {
	t.Helper()
	data, err := json.Marshal(allocations)
	if err != nil {
		t.Fatal(err)
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: gpuAllocationCMNamespace, Name: gpuAllocationCMName},
		Data:       map[string]string{gpuAllocationCMKey: string(data)},
	}
}

func newMigrationClient(t *testing.T, objects ...ctrlclient.Object) ctrlclient.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func TestMigrateIntelDGPUAllocations(t *testing.T) {
	ctx := context.Background()
	c := newMigrationClient(t,
		intelHAMINode(true),
		allocationConfigMap(t, []migratedComputeAllocation{
			{AppName: "vllm", Owner: "alice", Mode: intelGPUMode, NodeName: "n1", DeviceID: "n1-intel-gpu-card0"},
			{AppName: "vllm", Owner: "alice", Mode: intelGPUMode, NodeName: "n1", DeviceID: "n1-intel-gpu-card1"},
			{AppName: "comfy", Owner: "bob", Mode: intelGPUMode, NodeName: "n1", DeviceID: "n1-intel-gpu-card9"},
			{AppName: "jellyfin", Owner: "alice", Mode: "intel", NodeName: "n1", DeviceID: "n1-intel-0"},
		}),
		appWorkload("vllm-alice", "vllm", "vllm", "alice", legacyIntelDGPUResource),
		appWorkload("comfy-bob", "comfy", "comfy", "bob", legacyIntelDGPUResource),
		appWorkload("jellyfin-alice", "jellyfin", "jellyfin", "alice", legacyIntelDGPUResource),
	)
	var nodes corev1.NodeList
	if err := c.List(ctx, &nodes); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 20, 8, 0, 0, 0, time.UTC)
	if err := migrateIntelDGPUAllocations(ctx, c, &nodes, now); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	allocations, err := loadMigratedAllocations(ctx, c)
	if err != nil {
		t.Fatal(err)
	}
	gotIDs := []string{}
	for _, allocation := range allocations {
		gotIDs = append(gotIDs, allocation.DeviceID)
	}
	wantIDs := []string{testCard0UUID, testCard1UUID, "n1-intel-gpu-card9", "n1-intel-0"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("device IDs = %v, want %v", gotIDs, wantIDs)
	}

	binding := &unstructured.Unstructured{}
	binding.SetAPIVersion(legacyGPUBindingAPIVersion)
	binding.SetKind(gpuBindingKind)
	if err := c.Get(ctx, ctrlclient.ObjectKey{Name: "vllm-alice-intel-1a2b3c4d-0000-03-00.0-4363410d"}, binding); err != nil {
		t.Fatalf("binding with app-service's name for card0: %v", err)
	}
	if uuid, _, _ := unstructured.NestedString(binding.Object, "spec", "uuid"); uuid != testCard0UUID {
		t.Fatalf("binding uuid = %q", uuid)
	}
	if _, found, _ := unstructured.NestedFieldNoCopy(binding.Object, "spec", "memory"); found {
		t.Fatal("an Exclusive binding must not carry spec.memory")
	}
	if binding.GetLabels()[allocationModeLabelKey] != intelGPUMode {
		t.Fatalf("binding labels = %v", binding.GetLabels())
	}
	bindings := &unstructured.UnstructuredList{}
	bindings.SetAPIVersion(legacyGPUBindingAPIVersion)
	bindings.SetKind(legacyGPUBindingListKind)
	if err := c.List(ctx, bindings); err != nil {
		t.Fatal(err)
	}
	if len(bindings.Items) != 2 {
		t.Fatalf("expected bindings for vllm's two cards only, got %d", len(bindings.Items))
	}

	var node corev1.Node
	if err := c.Get(ctx, ctrlclient.ObjectKey{Name: "n1"}, &node); err != nil {
		t.Fatal(err)
	}
	if _, ok := node.Annotations[shareModeAnnotationPrefix+"n1-intel-gpu-card0"]; ok {
		t.Fatal("legacy share-mode annotation was kept")
	}
	if node.Annotations[shareModeAnnotationPrefix+testCard0UUID] != "0" {
		t.Fatalf("share mode of card0 = %q, want Exclusive", node.Annotations[shareModeAnnotationPrefix+testCard0UUID])
	}

	for name, want := range map[string]bool{"vllm-alice/vllm": true, "comfy-bob/comfy": false, "jellyfin-alice/jellyfin": false} {
		ns, n, _ := cutNamespacedName(name)
		var d appsv1.Deployment
		if err := c.Get(ctx, ctrlclient.ObjectKey{Namespace: ns, Name: n}, &d); err != nil {
			t.Fatal(err)
		}
		_, restarted := d.Spec.Template.Annotations[restartedAtAnnotation]
		if restarted != want {
			t.Errorf("deployment %s restarted = %v, want %v", name, restarted, want)
		}
	}

	if err := migrateIntelDGPUAllocations(ctx, c, &nodes, now); err != nil {
		t.Fatalf("second run: %v", err)
	}
	again, err := loadMigratedAllocations(ctx, c)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(again, allocations) {
		t.Fatalf("second run changed allocations: %+v", again)
	}
}

func cutNamespacedName(s string) (string, string, bool) {
	for i := range s {
		if s[i] == '/' {
			return s[:i], s[i+1:], true
		}
	}
	return "", s, false
}

func TestMigrateIntelDGPUWaitsForRegister(t *testing.T) {
	ctx := context.Background()
	c := newMigrationClient(t,
		intelHAMINode(false),
		allocationConfigMap(t, []migratedComputeAllocation{
			{AppName: "vllm", Owner: "alice", Mode: intelGPUMode, NodeName: "n1", DeviceID: "n1-intel-gpu-card0"},
		}),
		appWorkload("vllm-alice", "vllm", "vllm", "alice", legacyIntelDGPUResource),
	)
	nodes, err := waitForIntelHAMIRegister(ctx, c, 0, time.Millisecond)
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if err := migrateIntelDGPUAllocations(ctx, c, nodes, time.Now()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	allocations, err := loadMigratedAllocations(ctx, c)
	if err != nil {
		t.Fatal(err)
	}
	if allocations[0].DeviceID != "n1-intel-gpu-card0" {
		t.Fatalf("an unregistered node's allocation was rewritten to %s", allocations[0].DeviceID)
	}
	var d appsv1.Deployment
	if err := c.Get(ctx, ctrlclient.ObjectKey{Namespace: "vllm-alice", Name: "vllm"}, &d); err != nil {
		t.Fatal(err)
	}
	if _, ok := d.Spec.Template.Annotations[restartedAtAnnotation]; ok {
		t.Fatal("a workload whose card is not registered with HAMi was restarted")
	}
}

func TestLegacyIntelCardUUID(t *testing.T) {
	one := map[string]string{"card1": testCard1UUID}
	two := map[string]string{"card0": testCard0UUID, "card1": testCard1UUID}
	tests := []struct {
		deviceID string
		uuids    map[string]string
		want     string
		ok       bool
	}{
		{"n1-intel-gpu-card1", two, testCard1UUID, true},
		{"n1-intel-gpu-0", one, testCard1UUID, true},
		{"n1-intel-gpu-0", two, "", false},
		{"n2-intel-gpu-card1", two, "", false},
		{testCard0UUID, two, "", false},
	}
	for _, tt := range tests {
		got, ok := legacyIntelCardUUID(tt.deviceID, "n1", tt.uuids)
		if got != tt.want || ok != tt.ok {
			t.Errorf("legacyIntelCardUUID(%s) = %q, %v; want %q, %v", tt.deviceID, got, ok, tt.want, tt.ok)
		}
	}
}
