package validation

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// withMetricsProvider swaps the package-level cluster metrics provider
// for the duration of a test. Test isolation matters because all
// validator tests share the same process-level variable: any test
// leaving the override in place would taint everything that follows.
func withMetricsProvider(t *testing.T, stub func(token string) (*prometheus.ClusterMetrics, []string, error)) {
	t.Helper()
	orig := clusterMetricsProvider
	clusterMetricsProvider = stub
	t.Cleanup(func() { clusterMetricsProvider = orig })
}

// constantMetrics builds a stub provider that returns the given
// totals on every call, regardless of token. The capture pointer (if
// supplied) is updated with the token the validator passed so tests
// can confirm forwarding.
func constantMetrics(cpuCores, memBytes, diskBytes float64, capture *string) func(token string) (*prometheus.ClusterMetrics, []string, error) {
	return func(token string) (*prometheus.ClusterMetrics, []string, error) {
		if capture != nil {
			*capture = token
		}
		return &prometheus.ClusterMetrics{
			CPU:    prometheus.Value{Total: cpuCores},
			Memory: prometheus.Value{Total: memBytes},
			Disk:   prometheus.Value{Total: diskBytes},
		}, nil, nil
	}
}

func appWithLegacyReq(cpuMilli, memBytes, diskBytes int64) *appcfg.ApplicationConfig {
	cfg := &appcfg.ApplicationConfig{AppName: "test-app"}
	if cpuMilli > 0 {
		q := resource.NewMilliQuantity(cpuMilli, resource.DecimalSI)
		cfg.Requirement.CPU = q
	}
	if memBytes > 0 {
		q := resource.NewQuantity(memBytes, resource.BinarySI)
		cfg.Requirement.Memory = q
	}
	if diskBytes > 0 {
		q := resource.NewQuantity(diskBytes, resource.BinarySI)
		cfg.Requirement.Disk = q
	}
	return cfg
}

// TestClusterCapacityValidator_Pass covers the happy path: cluster
// totals strictly exceed the app's requirement across all three
// resources, so the validator returns OK.
func TestClusterCapacityValidator_Pass(t *testing.T) {
	withMetricsProvider(t, constantMetrics(
		8,       // 8 cores
		16<<30,  // 16 GiB
		200<<30, // 200 GiB
		nil,
	))

	app := appWithLegacyReq(3000 /*3 CPUs*/, 6<<30 /*6Gi*/, 10<<30 /*10Gi*/)
	d, err := clusterCapacityValidator{}.Validate(context.Background(), Input{
		AppConfig: app,
		Op:        v1alpha1.InstallOp,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.OK {
		t.Fatalf("expected OK, got %+v", d)
	}
}

// TestClusterCapacityValidator_NoRequirementSkipsProvider confirms
// that an app with no declared cpu / mem / disk requirement passes
// without calling the kubesphere metrics provider. The provider stub
// is wired up to fail loudly if it's invoked, which would otherwise
// be an easy regression to miss.
func TestClusterCapacityValidator_NoRequirementSkipsProvider(t *testing.T) {
	var called bool
	withMetricsProvider(t, func(token string) (*prometheus.ClusterMetrics, []string, error) {
		called = true
		return &prometheus.ClusterMetrics{}, nil, nil
	})

	d, err := clusterCapacityValidator{}.Validate(context.Background(), Input{
		AppConfig: &appcfg.ApplicationConfig{AppName: "no-req"},
		Op:        v1alpha1.InstallOp,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.OK {
		t.Fatalf("expected OK for zero-requirement app, got %+v", d)
	}
	if called {
		t.Fatalf("metrics provider should not have been called for zero-requirement app")
	}
}

// TestClusterCapacityValidator_PerResourceFailure pins the exact
// Decision shape for each shortfall. The Resource / Reason / Message
// fields are consumed by the HTTP handler to produce
// api.RequirementResp, so silently changing them would break clients.
func TestClusterCapacityValidator_PerResourceFailure(t *testing.T) {
	// A modest cluster: 4 CPUs, 8 GiB memory, 80 GiB disk.
	withMetricsProvider(t, constantMetrics(4, 8<<30, 80<<30, nil))

	cases := []struct {
		name         string
		app          *appcfg.ApplicationConfig
		wantResource constants.ResourceType
		wantReason   constants.ResourceConditionType
	}{
		{
			name:         "cpu shortfall",
			app:          appWithLegacyReq(100*1000 /*100 CPUs*/, 1<<30, 10<<30),
			wantResource: constants.CPU,
			wantReason:   constants.ClusterCPUInsufficient,
		},
		{
			name:         "memory shortfall",
			app:          appWithLegacyReq(1000, 1<<40 /*1 TiB*/, 10<<30),
			wantResource: constants.Memory,
			wantReason:   constants.ClusterMemoryInsufficient,
		},
		{
			name:         "disk shortfall",
			app:          appWithLegacyReq(1000, 1<<30, 1<<50 /*1 PiB*/),
			wantResource: constants.Disk,
			wantReason:   constants.ClusterDiskInsufficient,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := clusterCapacityValidator{}.Validate(context.Background(), Input{
				AppConfig: tc.app,
				Op:        v1alpha1.InstallOp,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.OK {
				t.Fatalf("expected non-OK, got %+v", d)
			}
			if d.Resource != tc.wantResource {
				t.Errorf("Decision.Resource=%q, want %q", d.Resource, tc.wantResource)
			}
			if d.Reason != tc.wantReason {
				t.Errorf("Decision.Reason=%q, want %q", d.Reason, tc.wantReason)
			}
			if d.Message == "" || !strings.Contains(d.Message, string(v1alpha1.InstallOp)) {
				t.Errorf("Decision.Message=%q, expected to include op verb %q", d.Message, v1alpha1.InstallOp)
			}
		})
	}
}

// TestClusterCapacityValidator_CPUUnitConversion is the highest-risk
// edge case: GetClusterResource returns CPU in WHOLE CORES (float64),
// while AddedResources.CPU is in MILLI-cores (int64). A unit slip
// here silently lets through apps that the cluster can't host or
// rejects apps it could host. The test wires up a cluster with
// exactly 2 cores total and probes both sides of 2000m.
func TestClusterCapacityValidator_CPUUnitConversion(t *testing.T) {
	withMetricsProvider(t, constantMetrics(2.0 /*2 cores*/, 1<<40, 1<<40, nil))

	// 1999m should pass: just under 2 cores.
	d, err := clusterCapacityValidator{}.Validate(context.Background(), Input{
		AppConfig: appWithLegacyReq(1999, 0, 0),
		Op:        v1alpha1.InstallOp,
	})
	if err != nil || !d.OK {
		t.Fatalf("expected 1999m to pass under 2-core cluster, got OK=%v err=%v decision=%+v", d.OK, err, d)
	}

	// 2001m must fail: just over 2 cores.
	d, err = clusterCapacityValidator{}.Validate(context.Background(), Input{
		AppConfig: appWithLegacyReq(2001, 0, 0),
		Op:        v1alpha1.InstallOp,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.OK {
		t.Fatalf("expected 2001m to fail under 2-core cluster, got OK")
	}
	if d.Reason != constants.ClusterCPUInsufficient {
		t.Fatalf("Decision.Reason=%q, want %q", d.Reason, constants.ClusterCPUInsufficient)
	}
}

// TestClusterCapacityValidator_PropagatesProviderError ensures that a
// kubesphere outage is surfaced as a validator error (which Run will
// propagate up to the caller) rather than silently passing or
// failing. This matches the clusterPressureValidator behavior — both
// depend on the same upstream.
func TestClusterCapacityValidator_PropagatesProviderError(t *testing.T) {
	boom := errors.New("kubesphere unreachable")
	withMetricsProvider(t, func(token string) (*prometheus.ClusterMetrics, []string, error) {
		return nil, nil, boom
	})

	_, err := clusterCapacityValidator{}.Validate(context.Background(), Input{
		AppConfig: appWithLegacyReq(1000, 1<<30, 1<<30),
		Op:        v1alpha1.InstallOp,
	})
	if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped provider error, got %v", err)
	}
}

// TestClusterCapacityValidator_ForwardsToken verifies that the
// Input.Token reaches the metrics provider unchanged. The kubesphere
// monitoring endpoint authenticates with the caller's service account
// token; dropping or stubbing the token would cause production
// failures that don't show up in any other test.
func TestClusterCapacityValidator_ForwardsToken(t *testing.T) {
	var seen string
	withMetricsProvider(t, constantMetrics(100, 1<<40, 1<<40, &seen))

	_, err := clusterCapacityValidator{}.Validate(context.Background(), Input{
		AppConfig: appWithLegacyReq(1000, 1<<30, 1<<30),
		Op:        v1alpha1.InstallOp,
		Token:     "sa-token-abcdef",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seen != "sa-token-abcdef" {
		t.Fatalf("provider received token %q, want %q", seen, "sa-token-abcdef")
	}
}
