package compute

import (
	"context"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAppSupportedModes(t *testing.T) {
	tests := []struct {
		name string
		cfg  *appcfg.ApplicationConfig
		want []string
	}{
		{
			name: "new format returns spec.resources verbatim in declared order",
			cfg: &appcfg.ApplicationConfig{
				Accelerator: []appcfg.ResourceMode{
					{Mode: utils.NvidiaCardType},
					{Mode: utils.CPUType},
				},
			},
			want: []string{utils.NvidiaCardType, utils.CPUType},
		},
		{
			name: "legacy app with non-zero requiredGpu collapses to nvidia",
			cfg: &appcfg.ApplicationConfig{
				Requirement: appcfg.AppRequirement{GPU: mustParseQty("10Gi")},
			},
			want: []string{utils.NvidiaCardType},
		},
		{
			name: "legacy app with zero requiredGpu is cpu-only",
			cfg: &appcfg.ApplicationConfig{
				Requirement: appcfg.AppRequirement{GPU: mustParseQty("0")},
			},
			want: []string{utils.CPUType},
		},
		{
			name: "legacy app with no GPU requirement is cpu-only",
			cfg:  &appcfg.ApplicationConfig{},
			want: []string{utils.CPUType},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AppSupportedModes(tt.cfg)
			if !equalStringSlices(got, tt.want) {
				t.Fatalf("AppSupportedModes(%#v) = %v, want %v", tt.cfg, got, tt.want)
			}
		})
	}
}

func mustParseQty(s string) *resource.Quantity {
	q := resource.MustParse(s)
	return &q
}

// TestAutoSelectModeFromInputs covers both worked-example clusters from the
// design doc:
//
//	cluster A: a single nvidia node (cluster gpu types = {nvidia})
//	cluster B: nvidia + amd nodes (cluster gpu types = {nvidia, amd})
//
// In both clusters cpu is implicit — every node can run a cpu-mode pod —
// so cpu is never enumerated in the cluster set itself; the auto-selector
// only treats cpu as a valid choice when the app declares it.
func TestAutoSelectModeFromInputs(t *testing.T) {
	clusterNvidiaOnly := stringSet(utils.NvidiaCardType)
	clusterNvidiaPlusAMD := stringSet(utils.NvidiaCardType, utils.AMDType)

	tests := []struct {
		name        string
		appModes    []string
		cluster     map[string]struct{}
		wantMode    string
		wantErrFrag string
	}{
		// case 1: cluster has only nvidia
		{
			name:     "case1/A: app supports nvidia/gb10/amd, only nvidia matches",
			appModes: []string{utils.NvidiaCardType, utils.GB10ChipType, utils.AMDType},
			cluster:  clusterNvidiaOnly,
			wantMode: utils.NvidiaCardType,
		},
		{
			name:        "case1/B: app supports nvidia/gb10/cpu, nvidia + cpu both runnable, ambiguous",
			appModes:    []string{utils.NvidiaCardType, utils.GB10ChipType, utils.CPUType},
			cluster:     clusterNvidiaOnly,
			wantErrFrag: "multiple compute modes",
		},
		{
			name:        "case1/C: app supports amd/apple-m, no overlap with cluster, errors",
			appModes:    []string{utils.AMDType, utils.AppleMChipType},
			cluster:     clusterNvidiaOnly,
			wantErrFrag: "No matching GPU type",
		},
		{
			name:     "case1/D: app supports cpu only, picks cpu",
			appModes: []string{utils.CPUType},
			cluster:  clusterNvidiaOnly,
			wantMode: utils.CPUType,
		},

		// case 2: cluster has both nvidia and amd
		{
			name:     "case2/A: app supports nvidia, picks nvidia",
			appModes: []string{utils.NvidiaCardType},
			cluster:  clusterNvidiaPlusAMD,
			wantMode: utils.NvidiaCardType,
		},
		{
			name:     "case2/B: app supports amd, picks amd",
			appModes: []string{utils.AMDType},
			cluster:  clusterNvidiaPlusAMD,
			wantMode: utils.AMDType,
		},
		{
			name:        "case2/C: app supports both nvidia and amd, ambiguous, errors",
			appModes:    []string{utils.NvidiaCardType, utils.AMDType},
			cluster:     clusterNvidiaPlusAMD,
			wantErrFrag: "multiple compute modes",
		},
		{
			name:     "case2/D: app supports cpu only, picks cpu",
			appModes: []string{utils.CPUType},
			cluster:  clusterNvidiaPlusAMD,
			wantMode: utils.CPUType,
		},
		{
			name:     "case2/E: app supports cpu/apple-m on cluster without apple-m, picks cpu",
			appModes: []string{utils.CPUType, utils.AppleMChipType},
			cluster:  clusterNvidiaPlusAMD,
			wantMode: utils.CPUType,
		},

		// degenerate cluster: no GPU at all
		{
			name:     "no GPU in cluster: cpu-supporting app picks cpu",
			appModes: []string{utils.CPUType, utils.NvidiaCardType},
			cluster:  stringSet(),
			wantMode: utils.CPUType,
		},
		{
			name:        "no GPU in cluster: GPU-only app errors",
			appModes:    []string{utils.NvidiaCardType},
			cluster:     stringSet(),
			wantErrFrag: "No matching GPU type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := autoSelectModeFromInputs(tt.appModes, tt.cluster)
			if tt.wantErrFrag != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got mode %q", tt.wantErrFrag, got)
				}
				if !strings.Contains(err.Error(), tt.wantErrFrag) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrFrag, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMode {
				t.Fatalf("expected mode %q, got %q", tt.wantMode, got)
			}
		})
	}
}

// TestAutoSelectModeShortCircuitsForSharedClientInstall verifies the
// "client-only install" path of v2 shared apps: when the shared subchart
// is already owned by some other user, the current install does not
// allocate compute on its own (the original admin's server install did),
// so AutoSelectMode must return cpu regardless of cluster GPU types or
// manifest-declared GPU modes.
//
// This branch fires for *any subsequent installer* of a v2 shared app
// whose admin-owned server is already in place — both a non-admin user
// installing the client and a second admin (Admin B) installing the same
// shared app that Admin A already owns. Only the very first installer
// of a v2 shared app wires up the server, and that first installer must
// be an admin: non-admin first-installers are rejected upstream by
// CheckDependencies (the manifest declares its own cluster-scoped admin
// install as a mandatory application dependency for non-admin renders).
//
// resolveComputeTarget returning !manage is the precise predicate for
// the "subsequent installer" case, and it doesn't look at the caller's
// admin status at all.
//
// Without this short-circuit the auto-selector would happily pick nvidia
// based on cluster types, then AppInstallable would reject the install
// because the !manage plan only contains a cpu row.
func TestAutoSelectModeShortCircuitsForSharedClientInstall(t *testing.T) {
	const (
		appName     = "comfyuiv2"
		sharedChart = "comfyuiv2server"
		serverOwner = "admin-a"
	)

	// We exercise the same scenario for both a non-admin caller and a
	// second-admin caller: in this layer the only signal that matters is
	// "someone else owns the shared server", so the result must be cpu
	// in both cases. The fake client setup does not differentiate admin
	// status — that's checked elsewhere in the install pipeline.
	cases := []struct {
		name         string
		currentOwner string
	}{
		{name: "non-admin client install", currentOwner: "alice"},
		{name: "second-admin install while admin-a owns shared server", currentOwner: "admin-b"},
	}

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add corev1 to scheme: %v", err)
	}

	// Cluster has a single nvidia node so a naive AutoSelectMode would
	// otherwise pick nvidia on a legacy GPU app.
	nvidiaNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-nvidia",
			Labels: map[string]string{
				utils.NodeGPUTypeLabel: utils.NvidiaCardType,
			},
		},
	}

	// Pre-existing shared namespace owned by serverOwner, signaling the
	// shared server has already been installed by another user.
	sharedNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: appcfg.ChartNamespace(&appcfg.Chart{Name: sharedChart, Shared: true}, serverOwner),
			Labels: map[string]string{
				constants.ApplicationNameLabel:        appName,
				constants.ApplicationInstallUserLabel: serverOwner,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(nvidiaNode, sharedNS).
				Build()

			cfg := &appcfg.ApplicationConfig{
				AppName:    appName,
				OwnerName:  tc.currentOwner,
				APIVersion: appcfg.V2,
				SubCharts: []appcfg.Chart{
					{Name: sharedChart, Shared: true},
				},
				// Legacy GPU app: AppSupportedModes would otherwise
				// return [nvidia] and the cluster has nvidia, so a
				// non-short-circuited path would confidently pick
				// nvidia for what is actually a non-managing install.
				Requirement: appcfg.AppRequirement{GPU: mustParseQty("10Gi")},
			}

			chosen, err := AutoSelectMode(context.Background(), c, cfg)
			if err != nil {
				t.Fatalf("AutoSelectMode unexpected error: %v", err)
			}
			if chosen != utils.CPUType {
				t.Fatalf("expected client-only install to short-circuit to %q, got %q", utils.CPUType, chosen)
			}
		})
	}
}

// TestAutoSelectModeOnNvidiaOnlyNewFormatApp covers the end-to-end install
// pre-check + auto-select path for a hypothetical new-format manifest
// (>= 0.12.0) whose spec.resources matrix declares only an nvidia mode
// (no cpu fallback).
//
// Without the placeholder fallback in resolveResourceMode this case used
// to crash inside ResolveRequirement on the very first GetAppConfig call
// with a confusing "no default cpu mode declared in spec.resources"
// error, which prevented auto-select from ever running. With the
// placeholder in place, the chart loader binds against the first
// declared mode (nvidia) and auto-select then picks the real mode based
// on the cluster.
func TestAutoSelectModeOnNvidiaOnlyNewFormatApp(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add corev1 to scheme: %v", err)
	}

	nvidiaNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-nvidia",
			Labels: map[string]string{
				utils.NodeGPUTypeLabel: utils.NvidiaCardType,
			},
		},
	}

	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(nvidiaNode).
		Build()

	// Simulate the appCfg state right after the *first* GetAppConfig call
	// with empty selectedGpu: SelectedGpuType is empty, but the
	// placeholder fallback in resolveResourceMode has bound Requirement
	// against the first declared mode (nvidia), so AppSupportedModes
	// reports [nvidia].
	cfg := &appcfg.ApplicationConfig{
		AppName:   "nvidia-only-new-format",
		OwnerName: "alice",
		Accelerator: []appcfg.ResourceMode{
			{
				Mode: utils.NvidiaCardType,
				ResourceRequirement: appcfg.ResourceRequirement{
					RequiredCPU: "1", RequiredMemory: "11Gi",
					RequiredGPU: "8Gi", LimitedGPU: "16Gi",
				},
			},
		},
		// Don't set SelectedGpuType — this is the "empty input" state.
	}

	chosen, err := AutoSelectMode(context.Background(), c, cfg)
	if err != nil {
		t.Fatalf("AutoSelectMode unexpected error: %v", err)
	}
	if chosen != utils.NvidiaCardType {
		t.Fatalf("expected nvidia (the only declared mode, also runnable on the cluster), got %q", chosen)
	}
}

func stringSet(values ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, v := range values {
		out[v] = struct{}{}
	}
	return out
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
