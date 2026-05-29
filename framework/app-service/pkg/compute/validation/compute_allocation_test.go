package validation

import (
	"context"
	"errors"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// withComputeAllocationProvider swaps the package-level allocation
// provider for the duration of a test. Mirrors withMetricsProvider —
// these are all process-level singletons and must be restored
// explicitly to keep the test suite hermetic.
func withComputeAllocationProvider(t *testing.T, stub func(context.Context, client.Client, *appcfg.ApplicationConfig) (*compute.Allocation, error)) {
	t.Helper()
	orig := computeAllocationProvider
	computeAllocationProvider = stub
	t.Cleanup(func() { computeAllocationProvider = orig })
}

func TestComputeAllocationValidatorPasses(t *testing.T) {
	// Successful allocation returns (*Allocation, nil); the validator
	// should ignore the allocation value and report OK.
	called := false
	withComputeAllocationProvider(t, func(_ context.Context, _ client.Client, _ *appcfg.ApplicationConfig) (*compute.Allocation, error) {
		called = true
		return &compute.Allocation{}, nil
	})

	d, err := computeAllocationValidator{}.Validate(context.Background(), Input{
		AppConfig: &appcfg.ApplicationConfig{AppName: "ok-app"},
		Op:        v1alpha1.InstallOp,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.OK {
		t.Fatalf("expected OK decision, got %#v", d)
	}
	if !called {
		t.Fatal("provider was not invoked")
	}
}

func TestComputeAllocationValidatorFailsOnSchedulerError(t *testing.T) {
	// The scheduler's error must round-trip to Decision.Message
	// verbatim so the caller's legacy
	// "Insufficient compute resource for selected mode %s: %v"
	// log line keeps reading the same shape it always has.
	withComputeAllocationProvider(t, func(_ context.Context, _ client.Client, _ *appcfg.ApplicationConfig) (*compute.Allocation, error) {
		return nil, errors.New("no node satisfies gpu mode shared")
	})

	d, err := computeAllocationValidator{}.Validate(context.Background(), Input{
		AppConfig: &appcfg.ApplicationConfig{AppName: "needy", SelectedGpuType: "shared"},
		Op:        v1alpha1.InstallOp,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.OK {
		t.Fatalf("expected !OK decision, got %#v", d)
	}
	if d.Resource != "compute" {
		t.Errorf("Resource=%q want %q", d.Resource, "compute")
	}
	if d.Reason != constants.ComputeAllocationFailed {
		t.Errorf("Reason=%q want %q", d.Reason, constants.ComputeAllocationFailed)
	}
	if d.Message != "no node satisfies gpu mode shared" {
		t.Errorf("Message=%q should be the raw scheduler error so the caller can re-format it", d.Message)
	}
}

func TestComputeAllocationValidatorAppliesTo(t *testing.T) {
	// Pin the install-only contract. Upgrade does not use validation.
	// whole package; resume reuses the placement chosen at install so
	// re-running the scheduler would either duplicate the Allocation
	// record or spuriously fail.
	v := computeAllocationValidator{}
	if !v.AppliesTo(v1alpha1.InstallOp) {
		t.Error("AppliesTo(InstallOp) = false, want true")
	}
	if v.AppliesTo(v1alpha1.UpgradeOp) {
		t.Error("AppliesTo(UpgradeOp) = true, want false")
	}
	if v.AppliesTo(v1alpha1.ResumeOp) {
		t.Error("AppliesTo(ResumeOp) = true, want false")
	}
	if v.AppliesTo(v1alpha1.StopOp) {
		t.Error("AppliesTo(StopOp) = true, want false")
	}
}
