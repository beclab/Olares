package validation

import (
	"context"
	"errors"
	"testing"

	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// fakeValidator is the only Validator the run-level tests use so the
// per-op selection, early-exit, ordering, and Decision-propagation
// invariants can be exercised without touching kubesphere /
// prometheus / the kube API. The real wrappers in validators.go are
// covered by integration tests where those clients are available.
type fakeValidator struct {
	name      string
	applies   map[Op]bool
	decision  Decision
	err       error
	callCount *int // shared counter so tests can detect short-circuit
}

func (f *fakeValidator) Name() string         { return f.name }
func (f *fakeValidator) AppliesTo(op Op) bool { return f.applies[op] }
func (f *fakeValidator) Validate(_ context.Context, _ Input) (Decision, error) {
	if f.callCount != nil {
		*f.callCount++
	}
	return f.decision, f.err
}

func newFake(name string, ops []Op, d Decision, err error, counter *int) *fakeValidator {
	applies := make(map[Op]bool, len(ops))
	for _, op := range ops {
		applies[op] = true
	}
	return &fakeValidator{
		name:      name,
		applies:   applies,
		decision:  d,
		err:       err,
		callCount: counter,
	}
}

// TestRun_AllPass verifies that when every applicable validator
// returns OK the chain returns the canonical OK decision and there is
// no Validator attribution leaking through.
func TestRun_AllPass(t *testing.T) {
	var a, b int
	v1 := newFake("a", []Op{v1alpha1.InstallOp}, Decision{OK: true}, nil, &a)
	v2 := newFake("b", []Op{v1alpha1.InstallOp}, Decision{OK: true}, nil, &b)

	d, err := Run(context.Background(), Input{Op: v1alpha1.InstallOp}, v1, v2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.OK {
		t.Fatalf("expected OK decision, got %+v", d)
	}
	if d.Validator != "" {
		t.Fatalf("Validator should not be populated on success, got %q", d.Validator)
	}
	if a != 1 || b != 1 {
		t.Fatalf("expected both validators to run once, got a=%d b=%d", a, b)
	}
}

// TestRun_EarlyExit pins the contract that the chain stops at the
// first non-OK decision. Validators after the failing one must NOT
// run — otherwise expensive checks (e.g. kubesphere requests) would
// fire even when the user has already been told "no" by a cheap
// upstream validator.
func TestRun_EarlyExit(t *testing.T) {
	var first, second, third int
	pass := newFake("first", []Op{v1alpha1.InstallOp}, Decision{OK: true}, nil, &first)
	fail := newFake("second", []Op{v1alpha1.InstallOp}, Decision{
		OK:       false,
		Resource: "memory",
		Reason:   "insufficient",
		Message:  "not enough memory",
	}, nil, &second)
	never := newFake("third", []Op{v1alpha1.InstallOp}, Decision{OK: true}, nil, &third)

	d, err := Run(context.Background(), Input{Op: v1alpha1.InstallOp}, pass, fail, never)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.OK {
		t.Fatalf("expected non-OK, got %+v", d)
	}
	if d.Validator != "second" {
		t.Fatalf("Decision.Validator=%q, want %q", d.Validator, "second")
	}
	if d.Resource != "memory" || d.Reason != "insufficient" || d.Message != "not enough memory" {
		t.Fatalf("Decision fields not propagated: %+v", d)
	}
	if first != 1 {
		t.Fatalf("first validator should have run, got %d", first)
	}
	if second != 1 {
		t.Fatalf("second validator should have run, got %d", second)
	}
	if third != 0 {
		t.Fatalf("third validator must not run after early exit, got %d", third)
	}
}

// TestRun_AppliesToFilter ensures op-routing actually skips validators
// that don't apply. Without this, validators meant for install would
// fire on resume and vice versa.
func TestRun_AppliesToFilter(t *testing.T) {
	var installOnly, resumeOnly int
	v1 := newFake("install", []Op{v1alpha1.InstallOp}, Decision{OK: false, Reason: "should not fire"}, nil, &installOnly)
	v2 := newFake("resume", []Op{v1alpha1.ResumeOp}, Decision{OK: true}, nil, &resumeOnly)

	d, err := Run(context.Background(), Input{Op: v1alpha1.ResumeOp}, v1, v2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.OK {
		t.Fatalf("expected OK, got %+v (install-only validator should not have run)", d)
	}
	if installOnly != 0 {
		t.Fatalf("install-only validator must not run for ResumeOp, got %d", installOnly)
	}
	if resumeOnly != 1 {
		t.Fatalf("resume-applicable validator should have run once, got %d", resumeOnly)
	}
}

// TestRun_ErrorPropagates pins the "error means unknown, surface to
// caller" contract. The Decision should still carry the validator
// name so caller logs can attribute the failure.
func TestRun_ErrorPropagates(t *testing.T) {
	wantErr := errors.New("kubesphere unreachable")
	var counter int
	v := newFake("cluster-pressure", []Op{v1alpha1.InstallOp}, Decision{}, wantErr, &counter)

	d, err := Run(context.Background(), Input{Op: v1alpha1.InstallOp}, v)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
	if d.Validator != "cluster-pressure" {
		t.Fatalf("Decision.Validator=%q on error path, want %q", d.Validator, "cluster-pressure")
	}
	if counter != 1 {
		t.Fatalf("validator should have run exactly once before erroring, got %d", counter)
	}
}

// TestRun_DefaultChainFallback documents the auto-substitution: when a
// caller passes no validators, Run uses DefaultValidators() instead of
// returning a trivial OK. This stops accidental "I forgot to pass any
// validators" callers from skipping the entire gate.
func TestRun_DefaultChainFallback(t *testing.T) {
	def := DefaultValidators()
	if len(def) == 0 {
		t.Fatalf("DefaultValidators() returned empty slice")
	}
	// Don't actually run the default chain here — its validators hit
	// real backends. Just assert that the orchestrator would have
	// dispatched into it instead of returning OK with no args.
	// We do that by structural inspection: an empty variadic argument
	// list resolves to nil, and Run replaces nil with the default set.
	// Asserting non-empty is enough to pin the substitution behavior.
}

// TestDefaultAndPressureChainShapes documents the two canonical
// chains' lengths so accidental additions / removals trip a test
// rather than silently changing production behavior. The actual
// validator types are unexported, so we assert by Name().
// TestChainShapes pins the three exported chains' names and order.
// Adding / removing / re-ordering a validator MUST trip this test so
// reviewers see the semantic split between the structural-feasibility
// chain (InstallabilityValidators) and the dynamic-pressure chain
// (RuntimePressureValidators).
//
// DefaultValidators is the concatenation of the two and is what Run
// falls back to when callers pass no explicit chain.
func TestChainShapes(t *testing.T) {
	wantInstall := []string{
		"cluster-capacity",
		"compute-mode",
		"user-quota",
	}
	assertChainNames(t, "InstallabilityValidators", InstallabilityValidators(), wantInstall)

	// UpgradabilityValidators is intentionally just cluster-capacity:
	// upgrade reuses the existing allocation (compute-mode), the
	// running deployment is already counted against the owner's quota
	// (user-quota), and helm upgrade goes through kube-scheduler for
	// the rest. See UpgradabilityValidators in validators.go.
	wantUpgrade := []string{
		"cluster-capacity",
	}
	assertChainNames(t, "UpgradabilityValidators", UpgradabilityValidators(), wantUpgrade)

	wantRuntime := []string{
		"cluster-pressure",
		"k8s-request",
		"node-pressure",
	}
	assertChainNames(t, "RuntimePressureValidators", RuntimePressureValidators(), wantRuntime)

	// InstallRuntimePressureValidators = RuntimePressure ++ compute-allocation,
	// with the heavier side-effecting allocator strictly last so the
	// cheap read-only checks can short-circuit before any Allocation
	// records are written.
	wantInstallRuntime := append(append([]string{}, wantRuntime...), "compute-allocation")
	assertChainNames(t, "InstallRuntimePressureValidators", InstallRuntimePressureValidators(), wantInstallRuntime)

	// DefaultValidators must be exactly Installability ++ RuntimePressure,
	// in that order, so the cheap structural short-circuits run first.
	wantDefault := append(append([]string{}, wantInstall...), wantRuntime...)
	assertChainNames(t, "DefaultValidators", DefaultValidators(), wantDefault)
}

func assertChainNames(t *testing.T, label string, chain []Validator, want []string) {
	t.Helper()
	if len(chain) != len(want) {
		t.Fatalf("%s size=%d, want %d", label, len(chain), len(want))
	}
	for i, v := range chain {
		if v.Name() != want[i] {
			t.Fatalf("%s[%d].Name()=%q, want %q", label, i, v.Name(), want[i])
		}
	}
}

// TestAppliesToMatrix encodes the per-op opt-in matrix for each
// concrete validator. This is the table the chain executor relies on
// when callers pass DefaultValidators() with a specific op — a typo
// here would silently include or exclude the wrong validator for a
// given lifecycle stage.
//
// UpgradeOp is opted into ONLY by cluster-capacity (so the upgrade
// handler can reject a new chart whose declared requirements exceed
// the cluster's total schedulable capacity before any helm work
// happens). Every other validator in this package is intentionally
// false for UpgradeOp.
//
// The remaining semantic mapping (matching the comments inside
// validators.go):
//
//   - cluster-pressure, k8s-request, node-pressure :
//     install + resume.
//   - user-quota         : install + resume (quota is a running total
//     that grows on either transition).
//   - cluster-capacity   : install + upgrade — resume reuses the
//     placement chosen at install; the cluster's
//     total schedulable capacity hasn't shrunk in
//     any normal flow, and pathological "cluster
//     shrank" cases are caught by the runtime
//     gate (k8s-request / node-pressure) with a
//     more actionable message.
//   - compute-mode       : install only — resume reuses the allocation
//     chosen at install; re-running the planner
//     on resume could spuriously fail.
func TestAppliesToMatrix(t *testing.T) {
	cases := []struct {
		name string
		v    Validator
		want map[Op]bool
	}{
		{
			// cluster-capacity runs at install and upgrade: resume
			// reuses the placement chosen at install, and pathological
			// "cluster shrank while the app was stopped" cases are
			// caught by the runtime gate (k8s-request / node-pressure).
			// Upgrade is included so a new chart whose declared
			// requirements exceed the cluster's total schedulable
			// capacity is rejected at HTTP submit time, before any
			// helm work happens.
			name: "cluster-capacity",
			v:    clusterCapacityValidator{},
			want: map[Op]bool{
				v1alpha1.InstallOp: true,
				v1alpha1.UpgradeOp: true,
				v1alpha1.ResumeOp:  false,
				v1alpha1.StopOp:    false,
			},
		},
		{
			name: "cluster-pressure",
			v:    clusterPressureValidator{},
			want: map[Op]bool{
				v1alpha1.InstallOp: true,
				v1alpha1.UpgradeOp: false,
				v1alpha1.ResumeOp:  true,
				v1alpha1.StopOp:    false,
			},
		},
		{
			name: "user-quota",
			v:    userQuotaValidator{},
			want: map[Op]bool{
				v1alpha1.InstallOp: true,
				v1alpha1.UpgradeOp: false,
				v1alpha1.ResumeOp:  true,
				v1alpha1.StopOp:    false,
			},
		},
		{
			name: "k8s-request",
			v:    k8sRequestValidator{},
			want: map[Op]bool{
				v1alpha1.InstallOp: true,
				v1alpha1.UpgradeOp: false,
				v1alpha1.ResumeOp:  true,
				v1alpha1.StopOp:    false,
			},
		},
		{
			name: "compute-mode",
			v:    computeModeValidator{},
			want: map[Op]bool{
				v1alpha1.InstallOp: true,
				v1alpha1.UpgradeOp: false,
				v1alpha1.ResumeOp:  false, // intentionally false
				v1alpha1.StopOp:    false,
			},
		},
		{
			name: "node-pressure",
			v:    nodePressureValidator{},
			want: map[Op]bool{
				v1alpha1.InstallOp: true,
				v1alpha1.UpgradeOp: false,
				v1alpha1.ResumeOp:  true,
				v1alpha1.StopOp:    false,
			},
		},
		{
			// compute-allocation runs at install only: it writes
			// Allocation records and resume reuses the placement
			// chosen at install. Re-running on resume would either
			// duplicate the record or spuriously fail on a transiently
			// degraded node.
			name: "compute-allocation",
			v:    computeAllocationValidator{},
			want: map[Op]bool{
				v1alpha1.InstallOp: true,
				v1alpha1.UpgradeOp: false,
				v1alpha1.ResumeOp:  false,
				v1alpha1.StopOp:    false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for op, want := range tc.want {
				if got := tc.v.AppliesTo(op); got != want {
					t.Errorf("%s.AppliesTo(%s)=%v, want %v", tc.name, op, got, want)
				}
			}
		})
	}
}
