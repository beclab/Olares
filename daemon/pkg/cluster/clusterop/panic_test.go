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

// --- Recover: a module that panics while recovering must not corrupt the
// record it was trying to explain, and must not take the daemon down. ---

// panickingRecoverModule is a RecoverableModule whose Recover always
// panics. done, if set, is closed the instant before the panic happens, so
// a test driving it through an asynchronous path (Manager.resume) can wait
// for it to have run without racing the goroutine that runs it.
type panickingRecoverModule struct {
	fakeModule
	typ      Type
	panicMsg string
	done     chan struct{}
}

func (p *panickingRecoverModule) Type() Type { return p.typ }

func (p *panickingRecoverModule) Recover(context.Context, Runtime, Operation) {
	if p.done != nil {
		close(p.done)
	}
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
func TestRecoverPanicDuringResumeLeavesTheRecordIntact(t *testing.T) {
	module := &panickingRecoverModule{
		typ: Type("kaboom-recover"), panicMsg: "boom: recover panic detail",
		done: make(chan struct{}),
	}
	reg := NewRegistry()
	if err := reg.Register(module); err != nil {
		t.Fatal(err)
	}
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")
	id := storeCommandIssued(t, c, dir, module.typ)

	m, err := NewManagerWithRegistry(c.deps(t, dir), reg)
	if err != nil {
		t.Fatalf("NewManagerWithRegistry: %v", err)
	}

	select {
	case <-module.done:
	case <-time.After(time.Second):
		t.Fatal("the module was never handed its unfinished command")
	}
	// The panic already happened by the time done closed; give the deferred
	// recover a moment to run before asserting on the record.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if got, ok := m.Get(id); ok && got.Status == StatusCommandIssued {
			return
		}
		time.Sleep(time.Millisecond)
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
