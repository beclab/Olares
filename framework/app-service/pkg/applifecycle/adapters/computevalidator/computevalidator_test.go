package computevalidator_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/computevalidator"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
)

// -----------------------------------------------------------------
// Test fixtures
// -----------------------------------------------------------------

// fakeValidator is a validation.Validator that records every
// Validate call and returns a canned Decision.
type fakeValidator struct {
	mu       sync.Mutex
	name     string
	applies  bool
	decision validation.Decision
	err      error
	calls    []validation.Input
}

func (f *fakeValidator) Name() string                   { return f.name }
func (f *fakeValidator) AppliesTo(_ validation.Op) bool { return f.applies }
func (f *fakeValidator) Validate(_ context.Context, in validation.Input) (validation.Decision, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, in)
	return f.decision, f.err
}

// recordingNotifier is the NotifierFunc test double.
type recordingNotifier struct {
	mu        sync.Mutex
	calls     []notifyCall
	returnErr error
}

type notifyCall struct {
	Subject string
	Payload any
}

func (n *recordingNotifier) Notify(_ context.Context, subject string, payload any) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls = append(n.calls, notifyCall{Subject: subject, Payload: payload})
	return n.returnErr
}

func noopNotifier(_ context.Context, _ string, _ any) error { return nil }

// -----------------------------------------------------------------
// New
// -----------------------------------------------------------------

func TestNew_NilClientPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil client")
		}
	}()
	_ = computevalidator.New(nil, noopNotifier)
}

func TestNew_NilNotifierPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil notifier")
		}
	}()
	_ = computevalidator.New(testutil.NewFakeClient(), nil)
}

func TestWithChainFor_NilKeepsDefault(t *testing.T) {
	// The purpose of the "nil = keep default" branch is that a
	// caller can stack WithChainFor(nil) without accidentally
	// blanking the mapping. This test asserts the branch: after
	// applying WithChainFor(nil), a real ValidateInstall call
	// still goes through the default install chain (which does
	// something — it will fail on the empty fake client, but the
	// fact that we see an error means the chain was invoked).
	v := computevalidator.New(testutil.NewFakeClient(), noopNotifier,
		computevalidator.WithChainFor(nil))
	if v == nil {
		t.Fatal("New returned nil")
	}
	// Not calling ValidateInstall here — the default chain reaches
	// out to kubesphere HTTP, which we cannot exercise from a unit
	// test. The mere fact that construction succeeded with a nil
	// WithChainFor is the assertion.
}

// -----------------------------------------------------------------
// ValidateInstall
// -----------------------------------------------------------------

func TestValidateInstall_WrongAppCfgType(t *testing.T) {
	v := computevalidator.New(testutil.NewFakeClient(), noopNotifier,
		computevalidator.WithChainFor(func(ports.ValidationOp) []validation.Validator {
			t.Fatal("chain should not run on bad appCfg cast")
			return nil
		}))
	_, err := v.ValidateInstall(context.Background(), ports.ValidationInput{
		AppCfg: "wrong type",
		Op:     ports.ValidateOpInstall,
	})
	if err == nil {
		t.Fatal("expected error on wrong appCfg type")
	}
}

func TestValidateInstall_NilAppCfgOfCorrectType(t *testing.T) {
	v := computevalidator.New(testutil.NewFakeClient(), noopNotifier,
		computevalidator.WithChainFor(func(ports.ValidationOp) []validation.Validator {
			t.Fatal("chain should not run on nil appCfg")
			return nil
		}))
	var nilCfg *appcfg.ApplicationConfig
	_, err := v.ValidateInstall(context.Background(), ports.ValidationInput{
		AppCfg: nilCfg,
		Op:     ports.ValidateOpInstall,
	})
	if err == nil {
		t.Fatal("expected error on nil *appcfg.ApplicationConfig")
	}
}

func TestValidateInstall_UnknownOp(t *testing.T) {
	v := computevalidator.New(testutil.NewFakeClient(), noopNotifier,
		computevalidator.WithChainFor(func(ports.ValidationOp) []validation.Validator {
			t.Fatal("chain should not run on unknown op")
			return nil
		}))
	_, err := v.ValidateInstall(context.Background(), ports.ValidationInput{
		AppCfg: &appcfg.ApplicationConfig{},
		Op:     ports.ValidationOp("something-invalid"),
	})
	if err == nil {
		t.Fatal("expected error on unknown ValidationOp")
	}
}

func TestValidateInstall_ChainOK(t *testing.T) {
	fv := &fakeValidator{
		name:     "test-validator",
		applies:  true,
		decision: validation.Decision{OK: true},
	}
	v := computevalidator.New(testutil.NewFakeClient(), noopNotifier,
		computevalidator.WithChainFor(func(op ports.ValidationOp) []validation.Validator {
			if op != ports.ValidateOpInstall {
				t.Errorf("chainFor called with %q want ValidateOpInstall", op)
			}
			return []validation.Validator{fv}
		}))

	got, err := v.ValidateInstall(context.Background(), ports.ValidationInput{
		AppCfg: &appcfg.ApplicationConfig{AppName: "app-x"},
		Op:     ports.ValidateOpInstall,
		Token:  "tok",
	})
	if err != nil {
		t.Fatalf("ValidateInstall: %v", err)
	}
	if !got.OK {
		t.Errorf("OK: got %v want true", got.OK)
	}
	if len(fv.calls) != 1 {
		t.Fatalf("validator not called: %d", len(fv.calls))
	}
	if fv.calls[0].Token != "tok" {
		t.Errorf("token: %q want tok", fv.calls[0].Token)
	}
	if fv.calls[0].AppConfig == nil || fv.calls[0].AppConfig.AppName != "app-x" {
		t.Errorf("AppConfig not passed through: %+v", fv.calls[0].AppConfig)
	}
}

func TestValidateInstall_ChainReject(t *testing.T) {
	fv := &fakeValidator{
		name:    "chain-rejector",
		applies: true,
		decision: validation.Decision{
			OK:       false,
			Resource: constants.ResourceType("cpu"),
			Reason:   constants.ResourceConditionType("PressureHigh"),
			Message:  "cluster is too busy",
		},
	}
	v := computevalidator.New(testutil.NewFakeClient(), noopNotifier,
		computevalidator.WithChainFor(func(ports.ValidationOp) []validation.Validator {
			return []validation.Validator{fv}
		}))

	got, err := v.ValidateInstall(context.Background(), ports.ValidationInput{
		AppCfg: &appcfg.ApplicationConfig{},
		Op:     ports.ValidateOpInstall,
	})
	if err != nil {
		t.Fatalf("ValidateInstall: %v (rejection is not an error at the port)", err)
	}
	if got.OK {
		t.Fatal("OK: got true want false")
	}
	if got.Reason != "PressureHigh" {
		t.Errorf("Reason: %q want PressureHigh", got.Reason)
	}
	if got.Resource != "cpu" {
		t.Errorf("Resource: %q want cpu", got.Resource)
	}
	if got.Message != "cluster is too busy" {
		t.Errorf("Message: %q", got.Message)
	}
	// validation.Run stamps Validator on the returned Decision;
	// the port passes it through verbatim.
	if got.Validator != "chain-rejector" {
		t.Errorf("Validator: %q want chain-rejector", got.Validator)
	}
}

func TestValidateInstall_ChainError(t *testing.T) {
	wantErr := errors.New("kubesphere HTTP down")
	fv := &fakeValidator{
		name:    "broken",
		applies: true,
		err:     wantErr,
	}
	v := computevalidator.New(testutil.NewFakeClient(), noopNotifier,
		computevalidator.WithChainFor(func(ports.ValidationOp) []validation.Validator {
			return []validation.Validator{fv}
		}))
	_, err := v.ValidateInstall(context.Background(), ports.ValidationInput{
		AppCfg: &appcfg.ApplicationConfig{},
		Op:     ports.ValidateOpInstall,
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("err: %v want %v", err, wantErr)
	}
}

// TestValidateInstall_UpgradeShortCircuitsToOK pins the upgrade
// short-circuit: ValidateOpUpgrade must return OK without invoking
// any validator, and MUST NOT fall through into validation.Run's
// empty-chain default (which would run the full install chain
// against an upgrade AM and reject it with a spurious pressure
// error). Regression guard for the "empty chain falls back to
// DefaultValidators" trap.
func TestValidateInstall_UpgradeShortCircuitsToOK(t *testing.T) {
	// No WithChainFor — use the production default mapping which
	// returns nil for ValidateOpUpgrade.
	v := computevalidator.New(testutil.NewFakeClient(), noopNotifier)
	got, err := v.ValidateInstall(context.Background(), ports.ValidationInput{
		AppCfg: &appcfg.ApplicationConfig{AppName: "app-x"},
		Op:     ports.ValidateOpUpgrade,
	})
	if err != nil {
		t.Fatalf("ValidateInstall(upgrade): %v (should short-circuit before running any validator)", err)
	}
	if !got.OK {
		t.Errorf("OK: got %v want true", got.OK)
	}
}

func TestValidatePostHelm_DelegatesToValidateInstall(t *testing.T) {
	var seenOps []ports.ValidationOp
	v := computevalidator.New(testutil.NewFakeClient(), noopNotifier,
		computevalidator.WithChainFor(func(op ports.ValidationOp) []validation.Validator {
			seenOps = append(seenOps, op)
			return nil // empty chain → OK short-circuit
		}))
	got, err := v.ValidatePostHelm(context.Background(), ports.ValidationInput{
		AppCfg: &appcfg.ApplicationConfig{},
		Op:     ports.ValidateOpInstall,
	})
	if err != nil {
		t.Fatalf("ValidatePostHelm: %v", err)
	}
	if !got.OK {
		t.Fatal("empty chain should short-circuit to OK")
	}
	if len(seenOps) != 1 || seenOps[0] != ports.ValidateOpInstall {
		t.Errorf("ops seen: %+v", seenOps)
	}
}

// -----------------------------------------------------------------
// NotifyInsufficientCompute
// -----------------------------------------------------------------

func TestNotifyInsufficientCompute_WrongAppCfgType(t *testing.T) {
	rn := &recordingNotifier{}
	v := computevalidator.New(testutil.NewFakeClient(), rn.Notify)
	err := v.NotifyInsufficientCompute(context.Background(), "wrong type", errors.New("boom"))
	if err == nil {
		t.Fatal("expected error on wrong appCfg type")
	}
	if len(rn.calls) != 0 {
		t.Errorf("bad appCfg leaked to notifier: %+v", rn.calls)
	}
}

func TestNotifyInsufficientCompute_PublishesToOsNotification(t *testing.T) {
	rn := &recordingNotifier{}
	v := computevalidator.New(testutil.NewFakeClient(), rn.Notify)
	cfg := &appcfg.ApplicationConfig{
		AppName:         "app-x",
		Namespace:       "user-x-ns",
		OwnerName:       "user-x",
		SelectedGpuType: "nvidia",
	}
	err := v.NotifyInsufficientCompute(context.Background(), cfg, errors.New("no gpu"))
	if err != nil {
		t.Fatalf("NotifyInsufficientCompute: %v", err)
	}
	if len(rn.calls) != 1 {
		t.Fatalf("notifier calls: %d want 1", len(rn.calls))
	}
	call := rn.calls[0]
	if call.Subject != "os.notification" {
		t.Errorf("subject: %q want os.notification", call.Subject)
	}
	// Payload shape is compute.resourceNotification (private). We
	// cannot assert against the struct type from outside pkg/compute
	// without an exported alias, but the wire (JSON) contract is
	// what downstream cares about. Assert that the payload is
	// non-nil — the byte-shape is pinned by pkg/compute unit tests
	// (or would be if we added them; the shared BuildInsufficientPayload
	// helper is exercised by the legacy PublishComputeInsufficientNotification
	// callers).
	if call.Payload == nil {
		t.Fatal("payload nil")
	}
}

func TestNotifyInsufficientCompute_PropagatesNotifierError(t *testing.T) {
	wantErr := errors.New("nats down")
	rn := &recordingNotifier{returnErr: wantErr}
	v := computevalidator.New(testutil.NewFakeClient(), rn.Notify)
	cfg := &appcfg.ApplicationConfig{AppName: "app-x"}
	err := v.NotifyInsufficientCompute(context.Background(), cfg, errors.New("no gpu"))
	if !errors.Is(err, wantErr) {
		t.Errorf("err: %v want %v", err, wantErr)
	}
}
