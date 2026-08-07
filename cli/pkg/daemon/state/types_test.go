package state

import (
	"encoding/json"
	"testing"
)

func TestStateGPUFieldsKeepLegacyInfoAndExposeFullList(t *testing.T) {
	primary := "NVIDIA Corporation GeForce RTX 4070"
	s := State{
		GpuInfo: &primary,
		GPUList: []string{
			primary,
			"Intel Corporation Arc Graphics",
		},
	}

	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got["gpu_info"] != primary {
		t.Fatalf("gpu_info = %#v, want %q", got["gpu_info"], primary)
	}
	list, ok := got["gpu_list"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("gpu_list = %#v, want two-element array", got["gpu_list"])
	}
}

func TestStateOsKernelIsExcludedFromSystemStatusWire(t *testing.T) {
	raw, err := json.Marshal(State{
		OsArch:   "amd64",
		OsKernel: "6.8.0-60-generic",
		HostIP:   "10.0.0.2",
	})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if _, exists := got["os_kernel"]; exists {
		t.Fatalf("os_kernel must not appear on /system/status wire, got %#v", got)
	}
	if _, exists := got["OsKernel"]; exists {
		t.Fatalf("OsKernel must not appear on /system/status wire, got %#v", got)
	}
	if got["os_arch"] != "amd64" {
		t.Fatalf("os_arch = %#v, want amd64", got["os_arch"])
	}
	if got["hostIp"] != "10.0.0.2" {
		t.Fatalf("hostIp = %#v, want 10.0.0.2", got["hostIp"])
	}
}

func TestStateGPUListIsEmptyArrayWhenNoGPUDetected(t *testing.T) {
	raw, err := json.Marshal(State{GPUList: []string{}})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	list, ok := got["gpu_list"].([]any)
	if !ok || len(list) != 0 {
		t.Fatalf("gpu_list = %#v, want empty array", got["gpu_list"])
	}
	if _, exists := got["gpu_info"]; exists {
		t.Fatalf("gpu_info should be omitted when nil, got %#v", got["gpu_info"])
	}
}
