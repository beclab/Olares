package appcfg

import (
	"encoding/json"
	"testing"

	"github.com/beclab/api/manifest"
)

func consumerTestConfig() *ApplicationConfig {
	two := int32(2)
	return &ApplicationConfig{
		WorkloadOptions: WorkloadOptions{
			"web": {},
			"llm": {
				Replicas: &two,
				Accelerator: &manifest.WorkloadAccelerator{
					PerPod: PerPodSplit,
					Modes: []ResourceMode{
						{Mode: "nvidia", ResourceRequirement: ResourceRequirement{RequiredGPU: "8Gi"}},
						{Mode: "nvidia-gb10"},
					},
				},
			},
			"asr": {
				Accelerator: &manifest.WorkloadAccelerator{
					Containers: []string{"server"},
					Modes:      []ResourceMode{{Mode: "nvidia", ResourceRequirement: ResourceRequirement{RequiredGPU: "4Gi"}}},
				},
			},
		},
	}
}

func TestGPUConsumersSortedAndNormalized(t *testing.T) {
	cfg := consumerTestConfig()
	if !cfg.HasGPUConsumers() {
		t.Fatal("expected consumers")
	}
	consumers := cfg.GPUConsumers()
	if len(consumers) != 2 || consumers[0].Name != "asr" || consumers[1].Name != "llm" {
		t.Fatalf("unexpected consumers %+v", consumers)
	}
	if consumers[0].PerPod != PerPodOne || consumers[0].Replicas != 1 {
		t.Fatalf("asr should default to one/1, got %s/%d", consumers[0].PerPod, consumers[0].Replicas)
	}
	if consumers[1].PerPod != PerPodSplit || consumers[1].Replicas != 2 {
		t.Fatalf("llm should be split/2, got %s/%d", consumers[1].PerPod, consumers[1].Replicas)
	}
}

func TestSelectedConsumerMode(t *testing.T) {
	cfg := consumerTestConfig()
	asr, _ := cfg.GPUConsumer("asr")
	llm, _ := cfg.GPUConsumer("llm")
	if mode, ok := cfg.SelectedConsumerMode(asr); !ok || mode.Mode != "nvidia" {
		t.Fatalf("single-mode consumer should not need a choice, got %v %v", mode, ok)
	}
	if _, ok := cfg.SelectedConsumerMode(llm); ok {
		t.Fatal("multi-mode consumer without a choice must not resolve")
	}
	cfg.SelectedGpuTypes = map[string]string{"llm": "nvidia-gb10"}
	if mode, ok := cfg.SelectedConsumerMode(llm); !ok || mode.Mode != "nvidia-gb10" {
		t.Fatalf("unexpected mode %v %v", mode, ok)
	}
	cfg.SelectedGpuTypes["llm"] = "amd"
	if _, ok := cfg.SelectedConsumerMode(llm); ok {
		t.Fatal("undeclared mode must not resolve")
	}
}

func TestCardsAndReplicasPerPolicy(t *testing.T) {
	cases := []struct {
		consumer     GPUConsumer
		cards, nodes int
		wantK        int
		wantReplicas int32
	}{
		{GPUConsumer{PerPod: PerPodOne, Replicas: 1}, 3, 2, 1, 3},
		{GPUConsumer{PerPod: PerPodAll, Replicas: 1}, 4, 2, 0, 2},
		{GPUConsumer{PerPod: PerPodSplit, Replicas: 2}, 4, 1, 2, 2},
	}
	for _, tc := range cases {
		if got := tc.consumer.CardsPerPod(tc.cards); got != tc.wantK {
			t.Errorf("%s: k=%d want %d", tc.consumer.PerPod, got, tc.wantK)
		}
		if got := tc.consumer.PodReplicas(tc.cards, tc.nodes); got != tc.wantReplicas {
			t.Errorf("%s: replicas=%d want %d", tc.consumer.PerPod, got, tc.wantReplicas)
		}
	}
}

// A Spec.Config written by 1.12.8 has neither SelectedGpuTypes nor workload
// accelerators and must keep resolving through SelectedGpuType.
func TestLegacySpecConfigStillResolves(t *testing.T) {
	raw := `{"AppName":"ollama","SelectedGpuType":"nvidia","PodGPUConsumePolicy":"single",
		"Accelerator":[{"mode":"nvidia","requiredGpu":"8Gi"},{"mode":"cpu"}],
		"WorkloadOptions":{"ollama":{"replicas":1}}}`
	var cfg ApplicationConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.HasGPUConsumers() || cfg.SelectedGpuTypes != nil {
		t.Fatal("legacy config must not turn into workload consumers")
	}
	mode, ok := cfg.SelectedResourceMode()
	if !ok || mode.Mode != "nvidia" {
		t.Fatalf("unexpected mode %v %v", mode, ok)
	}
}
