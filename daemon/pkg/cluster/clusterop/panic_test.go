package clusterop

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- Run: a module that panics must not take the daemon down with it. ---

// panickingModule is an OperationModule whose Run always panics with text
// that must never reach a caller: a real panic can carry a stack trace, a
// raw error, or anything else the module happened to be holding when it
// went wrong.
type panickingModule struct {
	fakeModule
	typ      Type
	panicMsg string
}

func (p *panickingModule) Type() Type { return p.typ }

func (p *panickingModule) Run(context.Context, Runtime, RunRequest) Outcome {
	panic(p.panicMsg)
}

// A module that panics mid-run has stopped the same way one that returns no
// usable outcome has: nothing is known about what it did, and the record
// must not be left running, holding the cluster's single-operation lock
// until the daemon restarts.
func TestRunPanicSettlesTheOperationFailed(t *testing.T) {
	module := &panickingModule{typ: Type("kaboom"), panicMsg: "boom: raw panic detail nobody should see"}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	op, err := createFake(t, m, module.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	settled := awaitTerminal(t, m, op.ID)

	if settled.Status != StatusFailed || settled.Code != CodeModuleFailed {
		t.Fatalf("status = %q code = %q, want failed/%s", settled.Status, settled.Code, CodeModuleFailed)
	}
	if strings.Contains(settled.Error, "boom") {
		t.Fatalf("Error = %q, leaked the panic's own text", settled.Error)
	}
	if settled.Error != reasonFor(CodeModuleFailed) {
		t.Fatalf("Error = %q, want the reviewed sentence %q", settled.Error, reasonFor(CodeModuleFailed))
	}
}

// The lock a panicking run held is released like any other settlement, so
// the cluster is not left waiting on a daemon restart to try again.
func TestRunPanicReleasesTheOperationLock(t *testing.T) {
	module := &panickingModule{typ: Type("kaboom"), panicMsg: "boom"}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	op, err := createFake(t, m, module.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	awaitTerminal(t, m, op.ID)

	next, err := createFake(t, m, module.typ, "client-2")
	if err != nil {
		t.Fatalf("Create() after a panicking Run = %v, want the cluster free again", err)
	}
	// Awaited so this second run's own goroutine is done before the test's
	// temp directory is cleaned up.
	awaitTerminal(t, m, next.ID)
}

// panicAfterCommitModule is an OperationModule whose Run first commits
// something real through its Runtime and only then panics — the shape every
// "the module already said something before it blew up" scenario shares.
type panicAfterCommitModule struct {
	fakeModule
	typ      Type
	panicMsg string
	commit   func(Runtime) error
}

func (p *panicAfterCommitModule) Type() Type { return p.typ }

func (p *panicAfterCommitModule) Run(_ context.Context, rt Runtime, _ RunRequest) Outcome {
	if err := p.commit(rt); err != nil {
		panic("commit before the real panic failed: " + err.Error())
	}
	panic(p.panicMsg)
}

// A module that hands out a command and then panics has already told the
// record the truth: the command_issued handoff is real, committed state,
// and a panic afterward must not be allowed to overwrite it with a generic
// module_failed the module never actually reported.
func TestRunPanicAfterCommandIssuedLeavesItIntact(t *testing.T) {
	module := &panicAfterCommitModule{
		typ: Type("kaboom-issued"), panicMsg: "boom: after command_issued",
		commit: func(rt Runtime) error {
			return rt.Complete(Outcome{Status: StatusCommandIssued, CommandIssuedUntil: rt.Now().Add(time.Hour)})
		},
	}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	op, err := createFake(t, m, module.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	settled := awaitTerminal(t, m, op.ID)
	// Give a wrongly-still-running settlement a moment to land, if the fix
	// is not in place, rather than trusting the first terminal read.
	time.Sleep(20 * time.Millisecond)
	settled, _ = m.Get(op.ID)

	if settled.Status != StatusCommandIssued {
		t.Fatalf("status = %q, want command_issued left exactly as the module committed it", settled.Status)
	}
	if settled.Code != "" || settled.Error != "" {
		t.Errorf("code = %q error = %q, want the committed outcome untouched by the later panic",
			settled.Code, settled.Error)
	}
}

// The same rule applies to any terminal status, not just command_issued: a
// module that already reported succeeded before it panicked must not have
// that overwritten with failed/module_failed.
func TestRunPanicAfterSucceededLeavesItIntact(t *testing.T) {
	module := &panicAfterCommitModule{
		typ: Type("kaboom-succeeded"), panicMsg: "boom: after succeeded",
		commit: func(rt Runtime) error {
			return rt.Complete(Outcome{Status: StatusSucceeded})
		},
	}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	op, err := createFake(t, m, module.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	awaitTerminal(t, m, op.ID)
	time.Sleep(20 * time.Millisecond)
	settled, _ := m.Get(op.ID)

	if settled.Status != StatusSucceeded {
		t.Fatalf("status = %q, want succeeded left exactly as the module committed it", settled.Status)
	}
	if settled.Code != "" || settled.Error != "" {
		t.Errorf("code = %q error = %q, want the committed outcome untouched by the later panic",
			settled.Code, settled.Error)
	}
}

// A module that panics without ever committing anything — the ordinary
// case, still exactly at running — is settled failed/module_failed as
// before, and any step it started but never finished is closed alongside
// it rather than left reading running under a failed operation.
func TestRunPanicAfterAStartedStepClosesItAndSettlesFailed(t *testing.T) {
	module := &panicAfterCommitModule{
		typ: Type("kaboom-step"), panicMsg: "boom: after starting a step",
		commit: func(rt Runtime) error {
			return rt.StartStep("work")
		},
	}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	op, err := createFake(t, m, module.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	settled := awaitTerminal(t, m, op.ID)

	if settled.Status != StatusFailed || settled.Code != CodeModuleFailed {
		t.Fatalf("status = %q code = %q, want failed/%s", settled.Status, settled.Code, CodeModuleFailed)
	}
	if len(settled.Steps) == 0 || settled.Steps[0].Status != StepFailed {
		t.Fatalf("Steps[0] = %+v, want the started step closed failed rather than left running", settled.Steps)
	}
	if settled.Steps[0].Code != CodeModuleFailed {
		t.Errorf("Steps[0].Code = %q, want %s", settled.Steps[0].Code, CodeModuleFailed)
	}
}

// --- Recover: a module that panics while recovering must not corrupt the
// record it was trying to explain, and must not take the daemon down. ---

// panickingRecoverModule is a RecoverableModule whose Recover always
// panics.
type panickingRecoverModule struct {
	fakeModule
	typ      Type
	panicMsg string
}

func (p *panickingRecoverModule) Type() Type { return p.typ }

func (p *panickingRecoverModule) Recover(context.Context, Runtime, Operation) {
	panic(p.panicMsg)
}

// A Recover that panics while this daemon is still loading leaves the
// operation exactly as if the module had looked and had nothing to add: the
// framework's own MarkInterrupted fallback still applies, because the
// operation is still non-terminal once the panic has been recovered from.
func TestRecoverPanicDuringLoadFallsBackToMarkInterrupted(t *testing.T) {
	module := &panickingRecoverModule{typ: Type("kaboom-recover"), panicMsg: "boom: recover panic detail"}
	reg := NewRegistry()
	if err := reg.Register(module); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "operations")
	storeOperation(t, dir, historicalOperation("op-1", module.typ, StatusRunning))

	m := newManagerWithRegistry(t, dir, reg)

	got, ok := m.Get("op-1")
	if !ok {
		t.Fatal("operation op-1 is gone")
	}
	if got.Status != StatusFailed || got.Code != CodeDaemonRestarted {
		t.Fatalf("status = %q code = %q, want failed/%s (the process did not crash, "+
			"which is the main thing this test checks)", got.Status, got.Code, CodeDaemonRestarted)
	}
	if strings.Contains(got.Error, "boom") {
		t.Fatalf("Error = %q, leaked the panic's own text", got.Error)
	}
}

// A Recover that panics while confirming an outstanding command (the
// asynchronous path, Manager.resume) must not corrupt the record: nothing
// committed through its Runtime before the panic is rolled back, and
// nothing after it ever ran, so the record is left exactly at command_issued
// — the same answer offering no recovery at all would have given for a
// command nothing could confirm.
//
// deps.recoveryDone (a test-only seam on Deps, see manager.go) is what
// makes this deterministic: it is signalled exactly when the resume
// goroutine that ran this panicking Recover returns, so the assertion below
// never races that goroutine's own deferred recover.
func TestRecoverPanicDuringResumeLeavesTheRecordIntact(t *testing.T) {
	module := &panickingRecoverModule{typ: Type("kaboom-recover"), panicMsg: "boom: recover panic detail"}
	reg := NewRegistry()
	if err := reg.Register(module); err != nil {
		t.Fatal(err)
	}
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")
	id := storeCommandIssued(t, c, dir, module.typ)

	deps := c.deps(t, dir)
	recoveryDone := make(chan string, 1)
	deps.recoveryDone = recoveryDone

	m, err := NewManagerWithRegistry(deps, reg)
	if err != nil {
		t.Fatalf("NewManagerWithRegistry: %v", err)
	}

	select {
	case got := <-recoveryDone:
		if got != id {
			t.Fatalf("recovery finished for %q, want %q", got, id)
		}
	case <-time.After(time.Second):
		t.Fatal("the panicking recovery never finished")
	}

	got, ok := m.Get(id)
	if !ok {
		t.Fatal("the stored operation was lost")
	}
	if got.Status != StatusCommandIssued {
		t.Errorf("status = %q, want the record left at command_issued", got.Status)
	}
}

// --- ExecuteNode: the package-level helper both task-6 handlers share. ---

// nodeCapableModule is an OperationModule that also implements
// NodeOperationModule, optionally panicking instead of returning.
type nodeCapableModule struct {
	fakeModule
	typ Type

	mu     sync.Mutex
	calls  []NodeRequest
	err    error
	panics bool
}

func (n *nodeCapableModule) Type() Type { return n.typ }

func (n *nodeCapableModule) ExecuteNode(_ context.Context, req NodeRequest) error {
	n.mu.Lock()
	n.calls = append(n.calls, req)
	n.mu.Unlock()
	if n.panics {
		panic("boom: node execution panic detail")
	}
	return n.err
}

func (n *nodeCapableModule) callCount() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.calls)
}

// ExecuteNode dispatches to whatever module the registry holds for the
// request's type, the same lookup Create uses.
func TestExecuteNodeDispatchesToTheRegisteredModule(t *testing.T) {
	module := &nodeCapableModule{typ: Type("node-op")}
	reg := registryWith(t, module)
	req := NodeRequest{PeerRequest: PeerRequest{Type: module.typ, OperationID: "op-1", RequestID: "client-1"}}

	if err := ExecuteNode(context.Background(), reg, req); err != nil {
		t.Fatalf("ExecuteNode: %v", err)
	}
	if module.callCount() != 1 {
		t.Fatalf("ExecuteNode calls = %d, want 1", module.callCount())
	}
}

// A nil registry is refused the same safe way an unknown type is, rather
// than panicking on the nil pointer before the function's own recover is
// even in place to catch it.
func TestExecuteNodeRefusesANilRegistry(t *testing.T) {
	req := NodeRequest{PeerRequest: PeerRequest{Type: Type("node-op"), OperationID: "op-1", RequestID: "client-1"}}

	err := ExecuteNode(context.Background(), nil, req)
	if powerCode(t, err) != CodeUnsupportedOperation {
		t.Fatalf("ExecuteNode() code = %q, want %s", powerCode(t, err), CodeUnsupportedOperation)
	}
}

// A type the registry does not hold is refused with the same stable code
// and sentence PowerHost already refuses an unknown operation with.
func TestExecuteNodeRefusesAnUnknownType(t *testing.T) {
	reg := NewRegistry()
	req := NodeRequest{PeerRequest: PeerRequest{Type: Type("no-such-type"), OperationID: "op-1", RequestID: "client-1"}}

	err := ExecuteNode(context.Background(), reg, req)
	if powerCode(t, err) != CodeUnsupportedOperation {
		t.Fatalf("ExecuteNode() code = %q, want %s", powerCode(t, err), CodeUnsupportedOperation)
	}
}

// A module the registry does know about, but which never declared a node
// capability, is refused the same way an unknown type is: there is nothing
// here that can act on a single node for it.
func TestExecuteNodeRefusesAModuleWithNoNodeCapability(t *testing.T) {
	fake := newFake(Type("no-node-capability"))
	reg := registryWith(t, fake)
	req := NodeRequest{PeerRequest: PeerRequest{Type: fake.typ, OperationID: "op-1", RequestID: "client-1"}}

	err := ExecuteNode(context.Background(), reg, req)
	if powerCode(t, err) != CodeUnsupportedOperation {
		t.Fatalf("ExecuteNode() code = %q, want %s", powerCode(t, err), CodeUnsupportedOperation)
	}
}

// A module's own ExecuteNode error is returned as-is: task 6's handlers, not
// this helper, decide how a module's own refusal is reported.
func TestExecuteNodeReturnsTheModulesOwnError(t *testing.T) {
	refused := errors.New("this node cannot do that right now")
	module := &nodeCapableModule{typ: Type("node-op"), err: refused}
	reg := registryWith(t, module)
	req := NodeRequest{PeerRequest: PeerRequest{Type: module.typ, OperationID: "op-1", RequestID: "client-1"}}

	if err := ExecuteNode(context.Background(), reg, req); !errors.Is(err, refused) {
		t.Fatalf("ExecuteNode() = %v, want the module's own refusal", err)
	}
}

// A module that panics while executing a node command must not crash the
// daemon or leak the panic's own text to whatever is waiting on the HTTP
// response.
func TestExecuteNodePanicSettlesModuleFailed(t *testing.T) {
	module := &nodeCapableModule{typ: Type("node-op"), panics: true}
	reg := registryWith(t, module)
	req := NodeRequest{PeerRequest: PeerRequest{Type: module.typ, OperationID: "op-1", RequestID: "client-1"}}

	err := ExecuteNode(context.Background(), reg, req)
	if err == nil {
		t.Fatal("ExecuteNode() = nil, want the panic reported as a module failure")
	}
	if powerCode(t, err) != CodeModuleFailed {
		t.Fatalf("ExecuteNode() code = %q, want %s", powerCode(t, err), CodeModuleFailed)
	}
	if strings.Contains(err.Error(), "boom") {
		t.Fatalf("ExecuteNode() = %v, leaked the panic's own text", err)
	}
}
