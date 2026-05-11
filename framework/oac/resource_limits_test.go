package oac

import (
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/manifest"

	olm "github.com/beclab/Olares/framework/oac/internal/manifest"
)

func TestResourceLimitsForResourceMode(t *testing.T) {
	cfg := &AppConfiguration{
		ConfigVersion: "0.13.0",
		ConfigType:    "app",
		APIVersion:    olm.APIVersionV2,
		Metadata:      manifest.AppMetaData{Name: "demo", Title: "D", Version: "1"},
		Entrances:     []appv1.Entrance{{Name: "w", Host: "d", Port: 80}},
		Spec: manifest.AppSpec{
			SupportArch: []string{"amd64"},
			Resources: []manifest.ResourceMode{{
				Mode: olm.ResourceModeNvidia,
				ResourceRequirement: manifest.ResourceRequirement{
					RequiredCPU: "150m", LimitedCPU: "300m",
					RequiredMemory: "192Mi", LimitedMemory: "384Mi",
					RequiredDisk: "3Gi", LimitedDisk: "6Gi",
					RequiredGPU: "6Gi", LimitedGPU: "12Gi",
				},
			}},
		},
	}
	lim, err := ResourceLimitsForResourceMode(cfg, "nvidia")
	if err != nil {
		t.Fatal(err)
	}
	if lim.RequiredCPU != "150m" || lim.LimitedCPU != "300m" {
		t.Fatalf("cpu mismatch: %+v", lim)
	}
	if lim.RequiredMemory != "192Mi" || lim.LimitedMemory != "384Mi" {
		t.Fatalf("memory mismatch: %+v", lim)
	}
	if lim.RequiredDisk != "3Gi" || lim.LimitedDisk != "6Gi" {
		t.Fatalf("disk mismatch: %+v", lim)
	}
	if lim.RequiredGPU != "6Gi" || lim.LimitedGPU != "12Gi" {
		t.Fatalf("gpu mismatch: %+v", lim)
	}

	// Mode lookup is case-insensitive.
	lim2, err := ResourceLimitsForResourceMode(cfg, "NVIDIA")
	if err != nil {
		t.Fatal(err)
	}
	if lim2 != lim {
		t.Fatalf("case-insensitive lookup mismatch: got %+v, want %+v", lim2, lim)
	}

	if _, err := ResourceLimitsForResourceMode(cfg, "cpu"); err == nil {
		t.Fatal("expected missing mode error")
	}
}

func TestResourceLimitsForResourceMode_NilCfg(t *testing.T) {
	if _, err := ResourceLimitsForResourceMode(nil, "cpu"); err == nil {
		t.Fatal("expected nil cfg error")
	}
}
