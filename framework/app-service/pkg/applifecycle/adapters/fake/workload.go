package fake

import (
	"context"
	"sync"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// WorkloadOps is an in-memory ports.WorkloadOps. Every method
// defaults to a successful WorkloadResult (Err=nil), which is what
// most handler tests want. Individual methods can be overridden via
// the ...Func hook fields so error paths and landing-state variants
// stay testable without forking the fake.
//
// All calls are recorded in Calls() in the order received, so tests
// can assert both the shape (which method) and the payload (which
// AM, which timeout).
type WorkloadOps struct {
	mu    sync.Mutex
	calls []WorkloadCall

	InstallFunc                 func(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult
	UpgradeFunc                 func(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult
	ApplyEnvFunc                func(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult
	UninstallFunc               func(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult
	UninstallAllFunc            func(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult
	StopFunc                    func(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult
	ResumeFunc                  func(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult
	WaitForStartUpFunc          func(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult
	WaitForLaunchFunc           func(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult
	WaitForNamespaceDeletedFunc func(ctx context.Context, req ports.NamespaceWaitRequest) ports.WorkloadResult
	GetDeployedVersionFunc      func(ctx context.Context, ns, release string) (string, int, error)
}

// WorkloadCall is one recorded invocation. Method is the port
// method name; exactly one of Req / NSReq is populated.
type WorkloadCall struct {
	Method string
	Req    *ports.WorkloadRequest
	NSReq  *ports.NamespaceWaitRequest
}

// Calls returns a copy of the recorded call log.
func (w *WorkloadOps) Calls() []WorkloadCall {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]WorkloadCall, len(w.calls))
	copy(out, w.calls)
	return out
}

// Reset clears the call log. Convenient between sub-tests.
func (w *WorkloadOps) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.calls = nil
}

func (w *WorkloadOps) recordWorkloadCall(method string, req ports.WorkloadRequest) {
	w.mu.Lock()
	defer w.mu.Unlock()
	r := req
	w.calls = append(w.calls, WorkloadCall{Method: method, Req: &r})
}

func (w *WorkloadOps) recordNamespaceCall(method string, req ports.NamespaceWaitRequest) {
	w.mu.Lock()
	defer w.mu.Unlock()
	r := req
	w.calls = append(w.calls, WorkloadCall{Method: method, NSReq: &r})
}

func (w *WorkloadOps) Install(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	w.recordWorkloadCall("Install", req)
	if w.InstallFunc != nil {
		return w.InstallFunc(ctx, req)
	}
	return ports.WorkloadResult{}
}

func (w *WorkloadOps) Upgrade(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	w.recordWorkloadCall("Upgrade", req)
	if w.UpgradeFunc != nil {
		return w.UpgradeFunc(ctx, req)
	}
	return ports.WorkloadResult{}
}

func (w *WorkloadOps) ApplyEnv(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	w.recordWorkloadCall("ApplyEnv", req)
	if w.ApplyEnvFunc != nil {
		return w.ApplyEnvFunc(ctx, req)
	}
	return ports.WorkloadResult{}
}

func (w *WorkloadOps) Uninstall(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	w.recordWorkloadCall("Uninstall", req)
	if w.UninstallFunc != nil {
		return w.UninstallFunc(ctx, req)
	}
	return ports.WorkloadResult{}
}

func (w *WorkloadOps) UninstallAll(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	w.recordWorkloadCall("UninstallAll", req)
	if w.UninstallAllFunc != nil {
		return w.UninstallAllFunc(ctx, req)
	}
	return ports.WorkloadResult{}
}

func (w *WorkloadOps) Stop(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	w.recordWorkloadCall("Stop", req)
	if w.StopFunc != nil {
		return w.StopFunc(ctx, req)
	}
	return ports.WorkloadResult{}
}

func (w *WorkloadOps) Resume(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	w.recordWorkloadCall("Resume", req)
	if w.ResumeFunc != nil {
		return w.ResumeFunc(ctx, req)
	}
	return ports.WorkloadResult{}
}

func (w *WorkloadOps) WaitForStartUp(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	w.recordWorkloadCall("WaitForStartUp", req)
	if w.WaitForStartUpFunc != nil {
		return w.WaitForStartUpFunc(ctx, req)
	}
	return ports.WorkloadResult{}
}

func (w *WorkloadOps) WaitForLaunch(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	w.recordWorkloadCall("WaitForLaunch", req)
	if w.WaitForLaunchFunc != nil {
		return w.WaitForLaunchFunc(ctx, req)
	}
	return ports.WorkloadResult{}
}

func (w *WorkloadOps) WaitForNamespaceDeleted(ctx context.Context, req ports.NamespaceWaitRequest) ports.WorkloadResult {
	w.recordNamespaceCall("WaitForNamespaceDeleted", req)
	if w.WaitForNamespaceDeletedFunc != nil {
		return w.WaitForNamespaceDeletedFunc(ctx, req)
	}
	return ports.WorkloadResult{}
}

func (w *WorkloadOps) GetDeployedVersion(ctx context.Context, ns, release string) (string, int, error) {
	if w.GetDeployedVersionFunc != nil {
		return w.GetDeployedVersionFunc(ctx, ns, release)
	}
	return "", 0, nil
}

// Interface conformance guard.
var _ ports.WorkloadOps = (*WorkloadOps)(nil)
