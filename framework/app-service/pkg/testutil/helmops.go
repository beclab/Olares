package testutil

import (
	"sync"

	"github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
)

var _ appinstaller.HelmOpsInterface = (*FakeHelmOps)(nil)

// FakeHelmOps is a configurable, thread-safe implementation of
// appinstaller.HelmOpsInterface for tests. It records the order of method
// calls and lets each method's result be configured.
type FakeHelmOps struct {
	mu sync.Mutex

	InstallErr      error
	UninstallErr    error
	UpgradeErr      error
	ApplyEnvErr     error
	RollBackErr     error
	UninstallAllErr error
	ScaleErr        error

	WaitForLaunchResult  bool
	WaitForLaunchErr     error
	WaitForStartUpResult bool
	WaitForStartUpErr    error

	calls         []string
	scaleReplicas []int32
}

// NewFakeHelmOps returns a FakeHelmOps whose wait methods report success by
// default, so the happy-path install/resume flows reach their running state.
func NewFakeHelmOps() *FakeHelmOps {
	return &FakeHelmOps{
		WaitForLaunchResult:  true,
		WaitForStartUpResult: true,
	}
}

func (f *FakeHelmOps) record(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, name)
}

// Calls returns a copy of the recorded method-call order.
func (f *FakeHelmOps) Calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	copy(out, f.calls)
	return out
}

// CallCount returns how many times the named method was invoked.
func (f *FakeHelmOps) CallCount(name string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if c == name {
			n++
		}
	}
	return n
}

// ScaleReplicas returns a copy of the replica arguments passed to Scale.
func (f *FakeHelmOps) ScaleReplicas() []int32 {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]int32, len(f.scaleReplicas))
	copy(out, f.scaleReplicas)
	return out
}

func (f *FakeHelmOps) Install() error { f.record("Install"); return f.InstallErr }

func (f *FakeHelmOps) Uninstall() error { f.record("Uninstall"); return f.UninstallErr }

func (f *FakeHelmOps) Upgrade() error { f.record("Upgrade"); return f.UpgradeErr }

func (f *FakeHelmOps) ApplyEnv() error { f.record("ApplyEnv"); return f.ApplyEnvErr }

func (f *FakeHelmOps) RollBack() error { f.record("RollBack"); return f.RollBackErr }

func (f *FakeHelmOps) UninstallAll() error { f.record("UninstallAll"); return f.UninstallAllErr }

func (f *FakeHelmOps) WaitForLaunch() (bool, error) {
	f.record("WaitForLaunch")
	return f.WaitForLaunchResult, f.WaitForLaunchErr
}

func (f *FakeHelmOps) WaitForStartUp() (bool, error) {
	f.record("WaitForStartUp")
	return f.WaitForStartUpResult, f.WaitForStartUpErr
}

func (f *FakeHelmOps) Scale(replicas int32) error {
	f.mu.Lock()
	f.calls = append(f.calls, "Scale")
	f.scaleReplicas = append(f.scaleReplicas, replicas)
	f.mu.Unlock()
	return f.ScaleErr
}
