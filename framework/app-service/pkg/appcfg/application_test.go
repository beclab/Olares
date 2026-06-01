package appcfg

import (
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

// TestResolveRequirementSelectedGpuFallbackSemantics covers the empty-vs-explicit
// distinction in resolveResourceMode for new-format manifests:
//
//   - selectedGpu == ""  →  the caller hasn't picked yet (pre-check / first
//     GetAppConfig before auto-select runs). Prefer cpu, but if the
//     manifest declares no cpu mode (a GPU-only new-format app), fall
//     back to the first declared mode as a placeholder so the chart
//     loader can succeed; the install handler's auto-select step reloads
//     with the real chosen mode immediately afterwards.
//   - selectedGpu != ""  →  the caller explicitly picked a mode. If the
//     manifest doesn't declare it, surface that as an error rather than
//     silently swapping to cpu / first-mode (which used to leave
//     SelectedGpuType and Requirement out of sync and produce a
//     misleading downstream "compute resource is not enough" message at
//     AppInstallable).
func TestResolveRequirementSelectedGpuFallbackSemantics(t *testing.T) {
	nvidiaCpuApp := &ApplicationConfig{
		AppName: "new-format-multi-mode",
		Accelerator: []ResourceMode{
			{
				Mode: utils.NvidiaCardType,
				ResourceRequirement: ResourceRequirement{
					RequiredCPU: "1", RequiredMemory: "11Gi",
					RequiredGPU: "8Gi", LimitedGPU: "16Gi",
				},
			},
			{
				Mode: utils.CPUType,
				ResourceRequirement: ResourceRequirement{
					RequiredCPU: "100m", RequiredMemory: "256Mi",
				},
			},
		},
	}
	nvidiaOnlyApp := &ApplicationConfig{
		AppName: "new-format-nvidia-only",
		Accelerator: []ResourceMode{
			{
				Mode: utils.NvidiaCardType,
				ResourceRequirement: ResourceRequirement{
					RequiredCPU: "1", RequiredMemory: "11Gi",
					RequiredGPU: "8Gi", LimitedGPU: "16Gi",
				},
			},
		},
	}

	tests := []struct {
		name        string
		cfg         *ApplicationConfig
		selectedGpu string
		wantMode    string
		wantGPU     string // expected RequiredGPU on the parsed Requirement, "" means nil
		wantErrFrag string
	}{
		{
			name:        "empty selectedGpu falls back to cpu placeholder",
			cfg:         nvidiaCpuApp,
			selectedGpu: "",
			wantMode:    utils.CPUType,
			wantGPU:     "", // cpu mode has no GPU requirement
		},
		{
			name:        "explicit nvidia matches and keeps GPU values",
			cfg:         nvidiaCpuApp,
			selectedGpu: utils.NvidiaCardType,
			wantMode:    utils.NvidiaCardType,
			wantGPU:     "8Gi",
		},
		{
			name:        "explicit cpu matches",
			cfg:         nvidiaCpuApp,
			selectedGpu: utils.CPUType,
			wantMode:    utils.CPUType,
			wantGPU:     "",
		},
		{
			name:        "explicit unsupported mode returns clear error (no silent cpu fallback)",
			cfg:         nvidiaCpuApp,
			selectedGpu: utils.AppleMChipType,
			wantErrFrag: "is not declared in spec.resources",
		},
		{
			// GPU-only new-format manifest: no cpu mode to fall back to.
			// resolveResourceMode picks the first declared mode (nvidia)
			// as a placeholder so the chart loader succeeds; auto-select
			// in the install handler will subsequently reload with the
			// real chosen mode.
			name:        "empty selectedGpu on nvidia-only manifest falls back to first mode as placeholder",
			cfg:         nvidiaOnlyApp,
			selectedGpu: "",
			wantMode:    utils.NvidiaCardType,
			wantGPU:     "8Gi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.cfg.ResolveRequirement(tt.selectedGpu)
			if tt.wantErrFrag != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got requirement %#v", tt.wantErrFrag, req)
				}
				if !strings.Contains(err.Error(), tt.wantErrFrag) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrFrag, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// We can't directly observe req.Mode here (AppRequirement doesn't
			// carry it), so verify the GPU side-effect: the cpu fallback
			// strips RequiredGPU, the nvidia path keeps it.
			gotGPU := ""
			if req.GPU != nil {
				gotGPU = req.GPU.String()
			}
			if gotGPU != tt.wantGPU {
				t.Fatalf("expected RequiredGPU %q, got %q", tt.wantGPU, gotGPU)
			}
		})
	}
}

// TestHasWorkloadReplicas pins the routing predicate that decides
// whether an app enters the new two-phase install / scale flow. The
// invariants:
//
//   - nil pointer  → false  (legacy v1/v3 manifest, fall back to old
//     single-phase helm + direct-patch suspend)
//   - empty map    → false  (same: the field was rendered but contains
//     no workloads; nothing to drive a Scale on)
//   - non-empty    → true   (opt-in confirmed)
//
// Mistakes here would either silently downgrade modern apps to the
// legacy code path or accidentally drag legacy apps into a flow they
// don't have helm template support for, so this lives in its own test
// with explicit truth-table coverage.
func TestHasWorkloadReplicas(t *testing.T) {
	emptyMap := WorkloadReplicas{}
	declared := WorkloadReplicas{"affine": 1, "worker": 2}

	cases := []struct {
		name string
		cfg  ApplicationConfig
		want bool
	}{
		{name: "nil pointer", cfg: ApplicationConfig{WorkloadReplicas: nil}, want: false},
		{name: "empty map", cfg: ApplicationConfig{WorkloadReplicas: &emptyMap}, want: false},
		{name: "declared map", cfg: ApplicationConfig{WorkloadReplicas: &declared}, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.HasWorkloadReplicas(); got != tc.want {
				t.Fatalf("HasWorkloadReplicas=%v, want %v", got, tc.want)
			}
		})
	}
}

// TestDesiredReplicas exercises the per-workload lookup used by Scale
// and by the helm values renderer. The fallback semantics are
// deliberate:
//
//   - nil map         → 1 (caller is expected to gate on
//     HasWorkloadReplicas first; this is a
//     last-resort safe default rather than a 0)
//   - declared name   → the declared value (including the explicit-zero
//     case, which lets a manifest pin a workload off
//     even when the rest of the app is opted in)
//   - undeclared name → 1 (defensive only: the manifest is required to
//     list every workload, so this branch should be
//     unreachable in practice; the 1 is a safe
//     last-resort default for a malformed manifest)
func TestDesiredReplicas(t *testing.T) {
	zero := int32(0)
	two := int32(2)
	declared := WorkloadReplicas{"affine": two, "worker": zero}

	cases := []struct {
		name     string
		cfg      ApplicationConfig
		workload string
		want     int32
	}{
		{name: "nil map → default 1", cfg: ApplicationConfig{}, workload: "any", want: 1},
		{name: "declared value", cfg: ApplicationConfig{WorkloadReplicas: &declared}, workload: "affine", want: 2},
		{name: "explicit zero is preserved", cfg: ApplicationConfig{WorkloadReplicas: &declared}, workload: "worker", want: 0},
		{name: "undeclared name → default 1", cfg: ApplicationConfig{WorkloadReplicas: &declared}, workload: "missing", want: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.DesiredReplicas(tc.workload); got != tc.want {
				t.Fatalf("DesiredReplicas(%q)=%d, want %d", tc.workload, got, tc.want)
			}
		})
	}
}
