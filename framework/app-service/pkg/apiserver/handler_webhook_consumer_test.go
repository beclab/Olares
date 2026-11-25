package apiserver

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/api/manifest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func consumerWebhookHandler(t *testing.T, rows []compute.Allocation) *Handler {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: "os-framework", Name: "app-gpu-allocations"},
		Data:       map[string]string{"allocations.json": string(data)},
	}
	return &Handler{ctrlClient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm).Build()}
}

func consumerWebhookConfig() *appcfg.ApplicationConfig {
	return &appcfg.ApplicationConfig{
		AppName:   "llm",
		OwnerName: "alice",
		WorkloadOptions: appcfg.WorkloadOptions{
			"server": {Replicas: ptr.To[int32](2), Accelerator: &manifest.WorkloadAccelerator{
				PerPod:     appcfg.PerPodSplit,
				Containers: []string{"vllm"},
				Modes:      []appcfg.ResourceMode{{Mode: "nvidia", ResourceRequirement: appcfg.ResourceRequirement{RequiredGPU: "32Gi"}}},
			}},
			"asr": {Accelerator: &manifest.WorkloadAccelerator{
				Modes: []appcfg.ResourceMode{{Mode: "nvidia", ResourceRequirement: appcfg.ResourceRequirement{RequiredGPU: "4Gi"}}},
			}},
			"web": {},
		},
	}
}

type jsonPatchOp struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value"`
}

func decodePatch(t *testing.T, raw []byte) map[string]jsonPatchOp {
	t.Helper()
	var ops []jsonPatchOp
	if err := json.Unmarshal(raw, &ops); err != nil {
		t.Fatal(err)
	}
	out := map[string]jsonPatchOp{}
	for _, op := range ops {
		out[op.Path] = op
	}
	return out
}

func TestConsumerPatchRequestsSplitGroupAndReplicas(t *testing.T) {
	rows := []compute.Allocation{
		{AppName: "llm", Owner: "alice", Mode: "nvidia", NodeName: "a", DeviceID: "a0", Consumer: "server"},
		{AppName: "llm", Owner: "alice", Mode: "nvidia", NodeName: "a", DeviceID: "a1", Consumer: "server"},
		{AppName: "llm", Owner: "alice", Mode: "nvidia", NodeName: "b", DeviceID: "b0", Consumer: "server"},
		{AppName: "llm", Owner: "alice", Mode: "nvidia", NodeName: "b", DeviceID: "b1", Consumer: "server"},
		{AppName: "llm", Owner: "alice", Mode: "nvidia", NodeName: "a", DeviceID: "a0", Consumer: "asr"},
	}
	handler := consumerWebhookHandler(t, rows)
	tpl := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "vllm"}, {Name: "proxy"}}}}
	patch, inject, err := handler.consumerGPUPatch(context.Background(), "server", tpl, ptr.To[int32](1), consumerWebhookConfig())
	if err != nil || !inject {
		t.Fatalf("unexpected %v %v", inject, err)
	}
	ops := decodePatch(t, patch)
	limits, ok := ops["/spec/template/spec/containers/0/resources/limits"]
	if !ok {
		t.Fatalf("vllm container got no limits: %s", patch)
	}
	var values map[string]string
	if err := json.Unmarshal(limits.Value, &values); err != nil {
		t.Fatal(err)
	}
	if values[constants.NvidiaGPU] != "2" || values[constants.NvidiaGPUMem] != "8192" {
		t.Fatalf("each pod should ask for 2 cards of 8Gi, got %v", values)
	}
	if _, ok := ops["/spec/template/spec/containers/1/resources/limits"]; ok {
		t.Fatal("only the declared container gets cards")
	}
	if replicas := ops["/spec/replicas"]; string(replicas.Value) != "2" {
		t.Fatalf("split runs its declared replicas, got %s", replicas.Value)
	}
}

func TestConsumerPatchLeavesZeroReplicasAndSkipsPlainWorkloads(t *testing.T) {
	handler := consumerWebhookHandler(t, []compute.Allocation{
		{AppName: "llm", Owner: "alice", Mode: "nvidia", NodeName: "a", DeviceID: "a0", Consumer: "asr"},
		{AppName: "llm", Owner: "alice", Mode: "nvidia", NodeName: "b", DeviceID: "b0", Consumer: "asr"},
	})
	cfg := consumerWebhookConfig()
	tpl := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "asr"}}}}
	patch, inject, err := handler.consumerGPUPatch(context.Background(), "asr", tpl, ptr.To[int32](0), cfg)
	if err != nil || !inject {
		t.Fatalf("unexpected %v %v", inject, err)
	}
	if _, ok := decodePatch(t, patch)["/spec/replicas"]; ok {
		t.Fatal("the install gate's zero replicas must survive")
	}
	patch, _, _ = handler.consumerGPUPatch(context.Background(), "asr", tpl, ptr.To[int32](1), cfg)
	if replicas := decodePatch(t, patch)["/spec/replicas"]; string(replicas.Value) != "2" {
		t.Fatalf("one card per pod runs a pod per card, got %s", replicas.Value)
	}
	if _, inject, _ := handler.consumerGPUPatch(context.Background(), "web", tpl, ptr.To[int32](1), cfg); inject {
		t.Fatal("a workload without an accelerator gets no cards")
	}
}

func TestConsumerPatchPinsNonHAMIModesToTheirRow(t *testing.T) {
	handler := consumerWebhookHandler(t, []compute.Allocation{
		{AppName: "llm", Owner: "alice", Mode: "intel", NodeName: "b", DeviceID: "b-intel-0", Consumer: "igpu"},
	})
	cfg := &appcfg.ApplicationConfig{
		AppName:          "llm",
		OwnerName:        "alice",
		SelectedGpuTypes: map[string]string{"igpu": "intel"},
		WorkloadOptions: appcfg.WorkloadOptions{
			"igpu": {Accelerator: &manifest.WorkloadAccelerator{Modes: []appcfg.ResourceMode{{Mode: "intel"}}}},
		},
	}
	tpl := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "igpu"}}}}
	patch, inject, err := handler.consumerGPUPatch(context.Background(), "igpu", tpl, ptr.To[int32](1), cfg)
	if err != nil || !inject {
		t.Fatalf("unexpected %v %v", inject, err)
	}
	selector, ok := decodePatch(t, patch)["/spec/template/spec/nodeSelector"]
	if !ok || string(selector.Value) != `{"kubernetes.io/hostname":"b"}` {
		t.Fatalf("intel pod must run on the node its row accounts for: %s", patch)
	}
}

func TestConsumerPodLabels(t *testing.T) {
	cfg := consumerWebhookConfig()
	labels := consumerPodLabels(cfg, "server")
	if labels[constants.AppGPUConsumerLabel] != "server" || labels[constants.AppPodGPUConsumePolicy] != "split" {
		t.Fatalf("unexpected labels %v", labels)
	}
	if got := consumerPodLabels(cfg, "asr")[constants.AppPodGPUConsumePolicy]; got != "single" {
		t.Fatalf("one maps onto HAMI's single policy, got %s", got)
	}
	if consumerPodLabels(cfg, "web") != nil {
		t.Fatal("plain workloads carry no GPU labels")
	}
}
