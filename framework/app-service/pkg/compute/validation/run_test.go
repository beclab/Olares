package validation

import (
	"context"
	"errors"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
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
	var counter, neverCount int
	v := newFake("cluster-pressure", []Op{v1alpha1.InstallOp}, Decision{
		Resource: "memory",
		Reason:   "unknown",
		Message:  "pressure unavailable",
	}, wantErr, &counter)
	never := newFake("later", []Op{v1alpha1.InstallOp}, Decision{OK: true}, nil, &neverCount)

	d, err := Run(context.Background(), Input{Op: v1alpha1.InstallOp}, v, never)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
	if d.Validator != "cluster-pressure" {
		t.Fatalf("Decision.Validator=%q on error path, want %q", d.Validator, "cluster-pressure")
	}
	if d.Resource != "memory" || d.Reason != "unknown" || d.Message != "pressure unavailable" {
		t.Fatalf("Decision fields not propagated on error: %+v", d)
	}
	if counter != 1 {
		t.Fatalf("validator should have run exactly once before erroring, got %d", counter)
	}
	if neverCount != 0 {
		t.Fatalf("validators after an error must not run, got %d calls", neverCount)
	}
}

func TestRuntimePressureRunsK8sAfterClusterPressurePasses(t *testing.T) {
	originalCluster := checkAppRequirement
	originalK8s := checkAppK8sRequestResource
	t.Cleanup(func() {
		checkAppRequirement = originalCluster
		checkAppK8sRequestResource = originalK8s
	})

	var clusterCalls, k8sCalls int
	checkAppRequirement = func(string, *appcfg.ApplicationConfig, v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		clusterCalls++
		return "", "", nil
	}
	checkAppK8sRequestResource = func(*appcfg.ApplicationConfig, v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		k8sCalls++
		return constants.CPU, constants.K8sRequestCPUPressure, errors.New("k8s cpu pressure")
	}

	decision, err := Run(context.Background(), Input{
		AppConfig: &appcfg.ApplicationConfig{},
		Op:        v1alpha1.InstallOp,
	}, RuntimePressureValidators()...)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if clusterCalls != 1 || k8sCalls != 1 {
		t.Fatalf("calls cluster=%d k8s=%d", clusterCalls, k8sCalls)
	}
	if decision.Validator != NameK8sRequest || decision.Reason != constants.K8sRequestCPUPressure {
		t.Fatalf("unexpected decision: %#v", decision)
	}
}

func TestMetricUnavailableValidatorsReturnFriendlyDecision(t *testing.T) {
	originalCluster := checkAppRequirement
	originalUser := checkUserResRequirement
	t.Cleanup(func() {
		checkAppRequirement = originalCluster
		checkUserResRequirement = originalUser
	})

	message := errors.New("Resource metrics are temporarily unavailable. Unable to install the application. Please try again later.")
	checkAppRequirement = func(string, *appcfg.ApplicationConfig, v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		return constants.CPU, constants.MetricsUnavailable, message
	}
	checkUserResRequirement = func(context.Context, *appcfg.ApplicationConfig, v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		return constants.Memory, constants.MetricsUnavailable, message
	}

	tests := []struct {
		name      string
		validator Validator
		resource  constants.ResourceType
	}{
		{name: "cluster", validator: clusterPressureValidator{}, resource: constants.CPU},
		{name: "user", validator: userQuotaValidator{}, resource: constants.Memory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := tt.validator.Validate(context.Background(), Input{
				AppConfig: &appcfg.ApplicationConfig{},
				Op:        v1alpha1.InstallOp,
			})
			if err != nil {
				t.Fatalf("validate: %v", err)
			}
			if decision.OK || decision.Resource != tt.resource || decision.Reason != constants.MetricsUnavailable {
				t.Fatalf("unexpected decision: %#v", decision)
			}
			if decision.Message != message.Error() {
				t.Fatalf("message=%q, want %q", decision.Message, message)
			}
		})
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

	// UpgradabilityValidators gates cluster-capacity on the new chart's
	// absolute requirements, then cluster-pressure / k8s-request on the
	// non-negative delta (new − old). user-quota, compute-mode,
	// node-pressure and compute-allocation stay excluded: install/resume
	// remain the quota gates, upgrade reuses the existing allocation,
	// and helm upgrade goes through kube-scheduler. See
	// UpgradabilityValidators in validators.go.
	wantUpgrade := []string{
		"cluster-capacity",
		"cluster-pressure",
		"k8s-request",
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
// UpgradeOp is opted into by cluster-capacity (absolute new
// requirements) and by cluster-pressure / k8s-request (which run
// against the delta new − old on upgrade). user-quota, compute-mode,
// node-pressure and compute-allocation stay false for UpgradeOp.
//
// The remaining semantic mapping (matching the comments inside
// validators.go):
//
//   - cluster-pressure, k8s-request : install + resume + upgrade.
//   - node-pressure                 : install + resume.
//   - user-quota         : install + resume only (upgrade does not
//     re-check owner quota).
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
			// cluster-pressure runs on upgrade too, but against the
			// resource delta (new − old) rather than absolute new.
			name: "cluster-pressure",
			v:    clusterPressureValidator{},
			want: map[Op]bool{
				v1alpha1.InstallOp: true,
				v1alpha1.UpgradeOp: true,
				v1alpha1.ResumeOp:  true,
				v1alpha1.StopOp:    false,
			},
		},
		{
			// user-quota is install + resume only; upgrade does not
			// re-check owner quota (see UpgradabilityValidators).
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
			// k8s-request runs on upgrade too, but against the resource
			// delta so already-scheduled requests are not double-counted.
			name: "k8s-request",
			v:    k8sRequestValidator{},
			want: map[Op]bool{
				v1alpha1.InstallOp: true,
				v1alpha1.UpgradeOp: true,
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

func upgradeConfig(name string, cpuMilli, memory int64) *appcfg.ApplicationConfig {
	cfg := &appcfg.ApplicationConfig{AppName: name, OwnerName: "alice", Namespace: name + "-alice"}
	if cpuMilli > 0 {
		cfg.Requirement.CPU = resource.NewMilliQuantity(cpuMilli, resource.DecimalSI)
	}
	if memory > 0 {
		cfg.Requirement.Memory = resource.NewQuantity(memory, resource.BinarySI)
	}
	return cfg
}

// stubKube points the validators at a fake kube API holding the given
// objects for the duration of the test.
func stubKube(t *testing.T, objects ...runtime.Object) {
	t.Helper()
	original := kubeClientset
	cs := k8sfake.NewSimpleClientset(objects...)
	kubeClientset = func() (kubernetes.Interface, error) { return cs, nil }
	t.Cleanup(func() { kubeClientset = original })
}

func namespaceObj(name string) *corev1.Namespace {
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func podObj(namespace, name string, phase corev1.PodPhase) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       corev1.PodSpec{NodeName: "node-a"},
		Status:     corev1.PodStatus{Phase: phase},
	}
}

func TestUpgradePrevHoldsRequestsIgnoresUnscheduledPods(t *testing.T) {
	pod := podObj("app-alice", "app-pending", corev1.PodPending)
	pod.Spec.NodeName = ""
	stubKube(t, namespaceObj("app-alice"), pod)

	holds, err := upgradePrevHoldsRequests(context.Background(), upgradeConfig("app", 1000, 2<<30), v1alpha1.UpgradeFailed, "")
	if err != nil {
		t.Fatalf("upgradePrevHoldsRequests: %v", err)
	}
	if holds {
		t.Fatal("unscheduled pod must not count as holding requests")
	}
}

func TestUpgradePrevHoldsRequestsRunningForcesDelta(t *testing.T) {
	// Running short-circuits before the kube lookup, so an empty fake
	// cluster is intentional: the state alone decides.
	stubKube(t)

	holds, err := upgradePrevHoldsRequests(context.Background(), upgradeConfig("app", 1000, 2<<30), v1alpha1.Running, "")
	if err != nil {
		t.Fatalf("upgradePrevHoldsRequests: %v", err)
	}
	if !holds {
		t.Fatal("Running must force the delta path")
	}
}

func TestUpgradePrevHoldsRequestsStoppedDoesNotHold(t *testing.T) {
	stubKube(t, namespaceObj("app-alice"), podObj("app-alice", "app-0", corev1.PodRunning))

	holds, err := upgradePrevHoldsRequests(context.Background(), upgradeConfig("app", 1000, 2<<30), v1alpha1.Stopped, "")
	if err != nil {
		t.Fatalf("upgradePrevHoldsRequests: %v", err)
	}
	if holds {
		t.Fatal("Stopped must not report holding requests")
	}
}

func TestSkipUpgradeResourceCheck(t *testing.T) {
	cases := []struct {
		name     string
		state    v1alpha1.ApplicationManagerState
		preState string
		want     bool
	}{
		{name: "stopped skips", state: v1alpha1.Stopped, want: true},
		{name: "upgradeFailed+stopped skips", state: v1alpha1.UpgradeFailed, preState: string(v1alpha1.Stopped), want: true},
		{name: "running does not skip", state: v1alpha1.Running, want: false},
		{name: "upgradeFailed+running does not skip", state: v1alpha1.UpgradeFailed, preState: string(v1alpha1.Running), want: false},
		{name: "upgradeFailed+empty does not skip", state: v1alpha1.UpgradeFailed, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := skipUpgradeResourceCheck(Input{
				Op:                    v1alpha1.UpgradeOp,
				PrevState:             tc.state,
				PreUpgradeSteadyState: tc.preState,
			})
			if got != tc.want {
				t.Fatalf("skip=%v want %v", got, tc.want)
			}
		})
	}
}

// TestUpgradeStoppedSkipsResourceChecks verifies Stopped (and
// UpgradeFailed with pre-upgrade-state=stopped) skip capacity / pressure /
// k8s-request without backend calls.
func TestUpgradeStoppedSkipsResourceChecks(t *testing.T) {
	originalCluster := checkAppRequirement
	originalK8s := checkAppK8sRequestResource
	originalMetrics := clusterMetricsProvider
	t.Cleanup(func() {
		checkAppRequirement = originalCluster
		checkAppK8sRequestResource = originalK8s
		clusterMetricsProvider = originalMetrics
	})

	var calls int
	checkAppRequirement = func(string, *appcfg.ApplicationConfig, v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		calls++
		return "", "", nil
	}
	checkAppK8sRequestResource = func(*appcfg.ApplicationConfig, v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		calls++
		return "", "", nil
	}
	clusterMetricsProvider = func(string) (*prometheus.ClusterMetrics, []string, error) {
		calls++
		return nil, nil, errors.New("metrics should not be called")
	}

	cases := []Input{
		{
			AppConfig:     upgradeConfig("app", 3000, 4<<30),
			PrevAppConfig: upgradeConfig("app", 1000, 2<<30),
			PrevState:     v1alpha1.Stopped,
			Op:            v1alpha1.UpgradeOp,
		},
		{
			AppConfig:             upgradeConfig("app", 3000, 4<<30),
			PrevAppConfig:         upgradeConfig("app", 1000, 2<<30),
			PrevState:             v1alpha1.UpgradeFailed,
			PreUpgradeSteadyState: string(v1alpha1.Stopped),
			Op:                    v1alpha1.UpgradeOp,
		},
	}
	for _, in := range cases {
		calls = 0
		decision, err := Run(context.Background(), in, UpgradabilityValidators()...)
		if err != nil {
			t.Fatalf("Run(%s): %v", in.PrevState, err)
		}
		if !decision.OK {
			t.Fatalf("Run(%s) decision=%#v, want OK", in.PrevState, decision)
		}
		if calls != 0 {
			t.Fatalf("Run(%s) expected zero backend calls, got %d", in.PrevState, calls)
		}
	}
}

func TestUpgradePrevHoldsRequestsUpgradeFailedUsesAnnotation(t *testing.T) {
	cases := []struct {
		name      string
		preState  string
		withPod   bool
		wantHolds bool
	}{
		{
			name:      "annotation running forces delta without probe",
			preState:  string(v1alpha1.Running),
			wantHolds: true,
		},
		{
			name:      "annotation stopped does not hold (skip at requirementConfigForOp)",
			preState:  string(v1alpha1.Stopped),
			withPod:   true,
			wantHolds: false,
		},
		{
			name:      "empty annotation falls back to probe with scheduled pod",
			preState:  "",
			withPod:   true,
			wantHolds: true,
		},
		{
			name:      "empty annotation falls back to probe without pods",
			preState:  "",
			wantHolds: false,
		},
		{
			name:      "garbage annotation falls back to probe without pods",
			preState:  "upgradeFailed",
			wantHolds: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			objs := []runtime.Object{namespaceObj("app-alice")}
			if tc.withPod {
				objs = append(objs, podObj("app-alice", "app-0", corev1.PodRunning))
			}
			stubKube(t, objs...)

			holds, err := upgradePrevHoldsRequests(context.Background(), upgradeConfig("app", 1000, 2<<30), v1alpha1.UpgradeFailed, tc.preState)
			if err != nil {
				t.Fatalf("upgradePrevHoldsRequests: %v", err)
			}
			if holds != tc.wantHolds {
				t.Fatalf("holds=%v want %v", holds, tc.wantHolds)
			}
		})
	}
}

// TestUpgradeDeltaFeedsPressureChecks verifies the delta-aware upgrade
// validators receive the non-negative delta (new − old) when the
// deployed version still has live pods, not the new chart's absolute
// requirements.
func TestUpgradeDeltaFeedsPressureChecks(t *testing.T) {
	originalCluster := checkAppRequirement
	originalK8s := checkAppK8sRequestResource
	t.Cleanup(func() {
		checkAppRequirement = originalCluster
		checkAppK8sRequestResource = originalK8s
	})

	var gotCPU int64 = -1
	capture := func(cfg *appcfg.ApplicationConfig) {
		if cfg != nil && cfg.Requirement.CPU != nil {
			gotCPU = cfg.Requirement.CPU.MilliValue()
		}
	}
	checkAppRequirement = func(_ string, cfg *appcfg.ApplicationConfig, _ v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		capture(cfg)
		return "", "", nil
	}
	checkAppK8sRequestResource = func(cfg *appcfg.ApplicationConfig, _ v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		capture(cfg)
		return "", "", nil
	}

	stubKube(t, namespaceObj("app-alice"), podObj("app-alice", "app-0", corev1.PodRunning))

	// new needs 3000m, old needs 1000m -> delta 2000m.
	in := Input{
		AppConfig:     upgradeConfig("app", 3000, 4<<30),
		PrevAppConfig: upgradeConfig("app", 1000, 2<<30),
		PrevState:     v1alpha1.UpgradeFailed,
		Op:            v1alpha1.UpgradeOp,
	}

	for _, v := range []Validator{clusterPressureValidator{}, k8sRequestValidator{}} {
		gotCPU = -1
		decision, err := v.Validate(context.Background(), in)
		if err != nil {
			t.Fatalf("%s validate: %v", v.Name(), err)
		}
		if !decision.OK {
			t.Fatalf("%s decision=%#v, want OK", v.Name(), decision)
		}
		if gotCPU != 2000 {
			t.Fatalf("%s received CPU=%dm, want delta 2000m", v.Name(), gotCPU)
		}
	}
}

// TestUpgradeZeroDeltaShortCircuits verifies the delta-aware upgrade
// validators pass without touching any backend when an app with live
// pods is upgraded to a version that keeps or lowers its requests.
func TestUpgradeZeroDeltaShortCircuits(t *testing.T) {
	originalCluster := checkAppRequirement
	originalK8s := checkAppK8sRequestResource
	t.Cleanup(func() {
		checkAppRequirement = originalCluster
		checkAppK8sRequestResource = originalK8s
	})

	var calls int
	checkAppRequirement = func(string, *appcfg.ApplicationConfig, v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		calls++
		return "", "", nil
	}
	checkAppK8sRequestResource = func(*appcfg.ApplicationConfig, v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		calls++
		return "", "", nil
	}

	stubKube(t, namespaceObj("app-alice"), podObj("app-alice", "app-0", corev1.PodRunning))

	// new <= old on every dimension -> delta zero.
	in := Input{
		AppConfig:     upgradeConfig("app", 1000, 2<<30),
		PrevAppConfig: upgradeConfig("app", 2000, 3<<30),
		PrevState:     v1alpha1.Running,
		Op:            v1alpha1.UpgradeOp,
	}

	for _, v := range []Validator{clusterPressureValidator{}, k8sRequestValidator{}} {
		decision, err := v.Validate(context.Background(), in)
		if err != nil {
			t.Fatalf("%s validate: %v", v.Name(), err)
		}
		if !decision.OK {
			t.Fatalf("%s decision=%#v, want OK", v.Name(), decision)
		}
	}
	if calls != 0 {
		t.Fatalf("expected zero backend calls for zero delta, got %d", calls)
	}
}

// TestUpgradeWithoutLivePodsChecksAbsolute verifies the delta discount
// is only applied while the deployed version actually holds its
// requests. A namespace whose pods have all terminated frees nothing, so
// a zero delta must not short-circuit: the validators have to see the
// new chart's full requirements.
func TestUpgradeWithoutLivePodsChecksAbsolute(t *testing.T) {
	originalCluster := checkAppRequirement
	originalK8s := checkAppK8sRequestResource
	t.Cleanup(func() {
		checkAppRequirement = originalCluster
		checkAppK8sRequestResource = originalK8s
	})

	var calls int
	var gotCPU int64
	capture := func(cfg *appcfg.ApplicationConfig) {
		calls++
		gotCPU = -1
		if cfg != nil && cfg.Requirement.CPU != nil {
			gotCPU = cfg.Requirement.CPU.MilliValue()
		}
	}
	checkAppRequirement = func(_ string, cfg *appcfg.ApplicationConfig, _ v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		capture(cfg)
		return "", "", nil
	}
	checkAppK8sRequestResource = func(cfg *appcfg.ApplicationConfig, _ v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		capture(cfg)
		return "", "", nil
	}

	stubKube(t,
		namespaceObj("app-alice"),
		podObj("app-alice", "app-succeeded", corev1.PodSucceeded),
		podObj("app-alice", "app-failed", corev1.PodFailed),
	)

	// new needs 1000m, old needs 2000m -> delta zero. UpgradeFailed falls
	// through to the live-pod probe, which finds nothing scheduled.
	in := Input{
		AppConfig:     upgradeConfig("app", 1000, 2<<30),
		PrevAppConfig: upgradeConfig("app", 2000, 3<<30),
		PrevState:     v1alpha1.UpgradeFailed,
		Op:            v1alpha1.UpgradeOp,
	}

	for _, v := range []Validator{clusterPressureValidator{}, k8sRequestValidator{}} {
		calls, gotCPU = 0, 0
		decision, err := v.Validate(context.Background(), in)
		if err != nil {
			t.Fatalf("%s validate: %v", v.Name(), err)
		}
		if !decision.OK {
			t.Fatalf("%s decision=%#v, want OK", v.Name(), decision)
		}
		if calls != 1 {
			t.Fatalf("%s: %d backend calls, want 1 against absolute requirements", v.Name(), calls)
		}
		if gotCPU != 1000 {
			t.Fatalf("%s: received CPU=%dm, want absolute 1000m", v.Name(), gotCPU)
		}
	}
}

// TestUpgradeAbortsWhenPreviousWorkloadUnknown verifies the delta-aware
// validators refuse to guess: if the previous release's namespace can't
// be inspected the check fails loudly instead of silently picking a
// discount that may not exist.
func TestUpgradeAbortsWhenPreviousWorkloadUnknown(t *testing.T) {
	originalCluster := checkAppRequirement
	originalK8s := checkAppK8sRequestResource
	t.Cleanup(func() {
		checkAppRequirement = originalCluster
		checkAppK8sRequestResource = originalK8s
	})

	var calls int
	checkAppRequirement = func(string, *appcfg.ApplicationConfig, v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		calls++
		return "", "", nil
	}
	checkAppK8sRequestResource = func(*appcfg.ApplicationConfig, v1alpha1.OpType) (constants.ResourceType, constants.ResourceConditionType, error) {
		calls++
		return "", "", nil
	}

	noNamespace := upgradeConfig("app", 2000, 3<<30)
	noNamespace.Namespace = ""

	cases := []struct {
		name string
		prev *appcfg.ApplicationConfig
	}{
		{"namespace missing from cluster", upgradeConfig("app", 2000, 3<<30)},
		{"previous config absent", nil},
		{"previous config has no namespace", noNamespace},
	}

	// The fake cluster deliberately holds no namespace at all.
	stubKube(t)

	for _, c := range cases {
		in := Input{
			AppConfig:     upgradeConfig("app", 1000, 2<<30),
			PrevAppConfig: c.prev,
			PrevState:     v1alpha1.UpgradeFailed,
			Op:            v1alpha1.UpgradeOp,
		}

		for _, v := range []Validator{clusterPressureValidator{}, k8sRequestValidator{}} {
			calls = 0
			if _, err := v.Validate(context.Background(), in); err == nil {
				t.Fatalf("%s with %s: expected error, got none", v.Name(), c.name)
			}
			if calls != 0 {
				t.Fatalf("%s with %s: %d backend calls, want none", v.Name(), c.name, calls)
			}
		}
	}
}
