package fake

import (
	"context"
	"sync"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// InstallValidator is an in-memory ports.InstallValidator. Default
// behaviour returns a passing ValidationDecision from both validate
// methods, which is what most handler tests want. Override via the
// ...Func hooks to reject.
type InstallValidator struct {
	mu    sync.Mutex
	calls []ValidatorCall

	ValidateInstallFunc           func(ctx context.Context, in ports.ValidationInput) (ports.ValidationDecision, error)
	ValidatePostHelmFunc          func(ctx context.Context, in ports.ValidationInput) (ports.ValidationDecision, error)
	NotifyInsufficientComputeFunc func(ctx context.Context, appCfg any, reason error) error
}

// ValidatorCall is one recorded invocation. Method is one of
// "ValidateInstall", "ValidatePostHelm", "NotifyInsufficientCompute".
type ValidatorCall struct {
	Method string
	Input  ports.ValidationInput // populated for Validate*
	Reason error                 // populated for NotifyInsufficientCompute
}

// Calls returns a copy of the recorded call log.
func (v *InstallValidator) Calls() []ValidatorCall {
	v.mu.Lock()
	defer v.mu.Unlock()
	out := make([]ValidatorCall, len(v.calls))
	copy(out, v.calls)
	return out
}

func (v *InstallValidator) recordCall(c ValidatorCall) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.calls = append(v.calls, c)
}

func (v *InstallValidator) ValidateInstall(ctx context.Context, in ports.ValidationInput) (ports.ValidationDecision, error) {
	v.recordCall(ValidatorCall{Method: "ValidateInstall", Input: in})
	if v.ValidateInstallFunc != nil {
		return v.ValidateInstallFunc(ctx, in)
	}
	return ports.ValidationDecision{OK: true}, nil
}

func (v *InstallValidator) ValidatePostHelm(ctx context.Context, in ports.ValidationInput) (ports.ValidationDecision, error) {
	v.recordCall(ValidatorCall{Method: "ValidatePostHelm", Input: in})
	if v.ValidatePostHelmFunc != nil {
		return v.ValidatePostHelmFunc(ctx, in)
	}
	return ports.ValidationDecision{OK: true}, nil
}

func (v *InstallValidator) NotifyInsufficientCompute(ctx context.Context, appCfg any, reason error) error {
	v.recordCall(ValidatorCall{Method: "NotifyInsufficientCompute", Reason: reason})
	if v.NotifyInsufficientComputeFunc != nil {
		return v.NotifyInsufficientComputeFunc(ctx, appCfg, reason)
	}
	return nil
}

// Interface conformance guard.
var _ ports.InstallValidator = (*InstallValidator)(nil)
