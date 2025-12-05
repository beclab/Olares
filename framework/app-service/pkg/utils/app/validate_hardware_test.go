package app

import (
	"context"
	"testing"

	"bytetrade.io/web3os/app-service/pkg/apiserver/api"
	"bytetrade.io/web3os/app-service/pkg/appcfg"
)

func TestCheckHardwareRequirement(t *testing.T) {
	origGet := getNodeInfo
	defer func() { getNodeInfo = origGet }()

	// fake cluster nodes
	nodes := []api.NodeInfo{
		{
			CudaVersion: "12.0",
			CPU: []api.CPUInfo{{
				CoreNumber: 16,
				Arch:       "amd64",
				Frequency:  0,
				Model:      "151",
				ModelName:  "12th Gen Intel(R) Core(TM) i5-12600KF",
				Vendor:     "GenuineIntel",
			},
			},
			Memory: api.MemInfo{Total: 32 << 30}, // 32Gi
			GPUS: []api.GPUInfo{
				{Vendor: "NVIDIA", Architecture: "Ampere", Model: "RTX 3080", Memory: 12 << 30},
				{Vendor: "NVIDIA", Architecture: "Ampere", Model: "RTX 3080", Memory: 8 << 30},
			},
		},
		{
			CudaVersion: "",
			CPU: []api.CPUInfo{{
				CoreNumber: 8,
				Arch:       "amd64",
				Frequency:  0,
				Model:      "23",
				ModelName:  "AMD Ryzen",
				Vendor:     "AuthenticAMD",
			},
			},
			Memory: api.MemInfo{Total: 16 << 30}, // 16Gi
			GPUS:   []api.GPUInfo{},
		},
	}
	getNodeInfo = func(ctx context.Context) ([]api.NodeInfo, error) { return nodes, nil }

	tests := []struct {
		name   string
		hw     appcfg.Hardware
		wantOK bool
	}{
		{
			name:   "no constraints passes",
			hw:     appcfg.Hardware{},
			wantOK: true,
		},
		{
			name: "cpu vendor matches intel",
			hw: appcfg.Hardware{
				Cpu: appcfg.CpuConfig{Vendor: "GenuineIntel"},
			},
			wantOK: true,
		},
		{
			name: "cpu vendor mismatch",
			hw: appcfg.Hardware{
				Cpu: appcfg.CpuConfig{Vendor: "NonExistVendor"},
			},
			wantOK: false,
		},
		{
			name: "gpu vendor matches NVIDIA",
			hw: appcfg.Hardware{
				Gpu: appcfg.GpuConfig{Vendor: "nvidia"},
			},
			wantOK: true,
		},
		{
			name: "gpu arch matches Ampere",
			hw: appcfg.Hardware{
				Gpu: appcfg.GpuConfig{Arch: []string{"Ampere"}},
			},
			wantOK: true,
		},
		{
			name: "gpu single memory 10Gi satisfied",
			hw: appcfg.Hardware{
				Gpu: appcfg.GpuConfig{SingleMemory: "10Gi"},
			},
			wantOK: true,
		},
		{
			name: "gpu single memory 16Gi not satisfied",
			hw: appcfg.Hardware{
				Gpu: appcfg.GpuConfig{SingleMemory: "16Gi"},
			},
			wantOK: false,
		},
		{
			name: "gpu total memory 20Gi satisfied",
			hw: appcfg.Hardware{
				Gpu: appcfg.GpuConfig{TotalMemory: "20Gi"},
			},
			wantOK: true,
		},
		{
			name: "gpu total memory 24Gi satisfied",
			hw: appcfg.Hardware{
				Gpu: appcfg.GpuConfig{TotalMemory: "24Gi"},
			},
			wantOK: false,
		},
		{
			name: "gpu total memory 10Gi satisfied",
			hw: appcfg.Hardware{
				Gpu: appcfg.GpuConfig{TotalMemory: "10Gi"},
			},
			wantOK: true,
		},
		{
			name: "gpu total memory 25Gi not satisfied",
			hw: appcfg.Hardware{
				Gpu: appcfg.GpuConfig{TotalMemory: "25Gi"},
			},
			wantOK: false,
		},
		{
			name: "intersection cpu vendor intel and gpu vendor nvidia satisfied on same node",
			hw: appcfg.Hardware{
				Cpu: appcfg.CpuConfig{Vendor: "GenuineIntel"},
				Gpu: appcfg.GpuConfig{Vendor: "NVIDIA"},
			},
			wantOK: true,
		},
		{
			name: "intersection cpu vendor amd and gpu vendor nvidia not satisfied",
			hw: appcfg.Hardware{
				Cpu: appcfg.CpuConfig{Vendor: "AuthenticAMD"},
				Gpu: appcfg.GpuConfig{Vendor: "NVIDIA"},
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &appcfg.ApplicationConfig{
				HardwareRequirement: tt.hw,
			}
			r, err := CheckHardwareRequirement(context.TODO(), cfg)
			gotOK := len(r) == 0
			if gotOK != tt.wantOK {
				t.Fatalf("expected ok=%v, got err=%v", tt.wantOK, err)
			}
		})
	}
}
