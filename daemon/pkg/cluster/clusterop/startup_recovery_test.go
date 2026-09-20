package clusterop

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// A module's Recover is other people's code, and this daemon runs it while it
// is still being built: olaresd does not serve a single request until
// NewManager returns. A Recover that never returns — because it waits on
// something that is gone, or ignores the context it was handed — must
// therefore not be able to stop the daemon from starting.
//
// The tests below drive that with a module that blocks on a channel and never
// looks at its context, against a manager whose own startup limit is a few
// milliseconds rather than the framework's ten seconds, so nothing here waits
// for a real timeout.

// blockingRecoverModule is a RecoverableModule that ignores its context
// entirely: Recover waits on release, which only the test can close, and then
// reports what the record made of the mutation it tried afterwards.
type blockingRecoverModule struct {
	fakeModule
	typ Type

	entered chan struct{}
	release chan struct{}
	lateErr chan error
}

func newBlockingRecoverModule(typ Type) *blockingRecoverModule {
	return &blockingRecoverModule{
		typ:     typ,
		entered: make(chan struct{}),
		release: make(chan struct{}),
		lateErr: make(chan error, 1),
	}
}

func (b *blockingRecoverModule) Type() Type { return b.typ }

func (b *blockingRecoverModule) Recover(_ context.Context, rt Runtime, _ Operation) {
	close(b.entered)
	<-b.release
	b.lateErr <- rt.Complete(Outcome{Status: StatusSucceeded})
}

// startupWith builds a manager over reg against what is already stored at
// dir, with the startup recovery limit shortened for the test, and reports
// how long the constructor was willing to wait. It fails the test rather than
// hanging when the constructor does not come back at all.
func startupWith(t *testing.T, dir string, reg *ModuleRegistry, limit time.Duration) *Manager {
	t.Helper()
	c := newCluster(master("master-1", "10.0.0.1"))
	deps := c.deps(t, dir)
	deps.recoveryTimeout = limit

	type built struct {
		m   *Manager
		err error
	}
	done := make(chan built, 1)
	go func() {
		m, err := NewManagerWithRegistry(deps, reg)
		done <- built{m, err}
	}()
	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("NewManagerWithRegistry: %v", got.err)
		}
		return got.m
	case <-time.After(10 * time.Second):
		t.Fatal("NewManagerWithRegistry never returned: a module's Recover is holding up the daemon's start")
		return nil
	}
}

// A record still moving when olaresd stopped is handed to its module before
// this daemon assumes the worst about it. A module that never hands it back
// does not get to hold the daemon there: the record is settled the same way
// one whose module offers no recovery at all is, and startup carries on.
func TestStartupDoesNotWaitForeverForAModuleThatNeverReturns(t *testing.T) {
	for _, status := range []Status{StatusPending, StatusRunning} {
		t.Run(string(status), func(t *testing.T) {
			module := newBlockingRecoverModule(Type("never-returns"))
			dir := filepath.Join(t.TempDir(), "operations")
			storeOperation(t, dir, historicalOperation("op-1", module.typ, status))

			m := startupWith(t, dir, registryWith(t, module), 20*time.Millisecond)
			t.Cleanup(func() { close(module.release) })

			select {
			case <-module.entered:
			case <-time.After(time.Second):
				t.Fatal("the module was never asked to recover the record")
			}
			got, ok := m.Get("op-1")
			if !ok {
				t.Fatal("operation op-1 is gone")
			}
			if got.Status != StatusFailed || got.Code != CodeDaemonRestarted {
				t.Fatalf("status = %q code = %q, want failed/%s once the module ran out of time",
					got.Status, got.Code, CodeDaemonRestarted)
			}
			if got.FinishedAt == nil || !got.UpdatedAt.Equal(*got.FinishedAt) {
				t.Errorf("UpdatedAt = %v FinishedAt = %v, want one settlement moment",
					got.UpdatedAt, got.FinishedAt)
			}
		})
	}
}

// The goroutine a timed-out Recover was left running still holds a Runtime.
// What it commits through it afterwards is a module reporting on an operation
// this daemon has already settled and told the rest of the cluster about, so
// it is refused rather than allowed to reopen the record.
func TestARecoverThatComesBackTooLateCannotOverwriteTheSettledRecord(t *testing.T) {
	module := newBlockingRecoverModule(Type("comes-back-late"))
	dir := filepath.Join(t.TempDir(), "operations")
	storeOperation(t, dir, historicalOperation("op-1", module.typ, StatusRunning))

	m := startupWith(t, dir, registryWith(t, module), 20*time.Millisecond)
	settled, ok := m.Get("op-1")
	if !ok {
		t.Fatal("operation op-1 is gone")
	}

	close(module.release)
	select {
	case err := <-module.lateErr:
		if !errors.Is(err, ErrOperationTerminal) {
			t.Fatalf("the late mutation returned %v, want it refused with ErrOperationTerminal", err)
		}
	case <-time.After(time.Second):
		t.Fatal("the module never came back")
	}

	got, _ := m.Get("op-1")
	if got.Status != settled.Status || got.Code != settled.Code {
		t.Fatalf("status = %q code = %q, want the settlement left as it was (%q/%q)",
			got.Status, got.Code, settled.Status, settled.Code)
	}
}

// slowConfirmingModule confirms an outstanding command, but only once the
// test lets it — later than any startup limit would have allowed.
type slowConfirmingModule struct {
	fakeModule
	typ     Type
	release chan struct{}
}

func (s *slowConfirmingModule) Type() Type { return s.typ }

func (s *slowConfirmingModule) Recover(_ context.Context, rt Runtime, _ Operation) {
	<-s.release
	if err := rt.Complete(Outcome{Status: StatusSucceeded}); err != nil {
		panic("confirming an outstanding command failed: " + err.Error())
	}
}

// The startup limit bounds what the daemon waits for before it starts
// serving, and nothing else. Confirming a command that outlived this daemon
// is the case that legitimately takes minutes — a rebooting control node is
// watched for up to Timeouts.Ready — so it runs in the background and is
// never force-failed by the limit that keeps startup moving.
func TestConfirmingAnOutstandingCommandIsNotBoundedByTheStartupLimit(t *testing.T) {
	module := &slowConfirmingModule{typ: Type("slow-confirm"), release: make(chan struct{})}
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")
	id := storeCommandIssued(t, c, dir, module.typ)

	deps := c.deps(t, dir)
	deps.recoveryTimeout = time.Millisecond
	recoveryDone := make(chan string, 1)
	deps.recoveryDone = recoveryDone

	m, err := NewManagerWithRegistry(deps, registryWith(t, module))
	if err != nil {
		t.Fatalf("NewManagerWithRegistry: %v", err)
	}

	// Held well past the startup limit, so a confirmation subjected to it
	// would already have been settled as interrupted by the time the module
	// reports what it found.
	time.Sleep(20 * time.Millisecond)
	close(module.release)

	select {
	case <-recoveryDone:
	case <-time.After(5 * time.Second):
		t.Fatal("the confirmation never finished")
	}
	got, ok := m.Get(id)
	if !ok {
		t.Fatal("the stored operation was lost")
	}
	if got.Status != StatusSucceeded {
		t.Fatalf("status = %q, want the module's own confirmation", got.Status)
	}
}
