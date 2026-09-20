package clusterop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// This file is the claim the whole refactor was for: a cluster operation is
// one module file, and adding one is all adding an operation takes.
//
// Everything the operation below needs is in this file — its name, what it
// accepts, what it makes the cluster look like, what it does on the master
// and what it does on a node. Nothing else in the package was edited to make
// it work: no switch gained a case, no list gained an entry, and no type
// constant was declared next to reboot and shutdown. The tests prove that
// twice over: the operation runs end to end through a registry of its own,
// and the daemon's own ParseType still refuses its name, which it could not
// do if any production list had been told about it.

const exampleType = Type("example")

// examplePhase is what this operation makes the cluster look like while it
// is happening. It is declared here, by the module, and no mapping anywhere
// else knows it.
const examplePhase = nodestatus.Phase("exampling")

const exampleStep = "example-work"

// exampleParams is this operation's own input. Nothing signs it at any hop,
// so the module is the only thing that says what it will accept — see
// OperationModule.Validate.
type exampleParams struct {
	Value *int `json:"value"`
}

// exampleOperationModule is an operation this daemon does not have and never
// will. It exists to be added the way a real one would be: written once,
// registered once, and reachable everywhere without anything else changing.
type exampleOperationModule struct {
	typ     Type
	outcome Outcome

	// release, when non-nil, holds Run open until the test closes it, which
	// is the only way to observe a cluster with something actually in
	// flight.
	release chan struct{}

	mu        sync.Mutex
	validated []CreateRequest
	runCalls  int
	ranParams []string
	nodeCalls []NodeRequest
}

func newExampleModule() *exampleOperationModule {
	return &exampleOperationModule{typ: exampleType, outcome: Outcome{Status: StatusSucceeded}}
}

func (e *exampleOperationModule) Type() Type { return e.typ }

func (e *exampleOperationModule) Validate(req CreateRequest) error {
	e.mu.Lock()
	e.validated = append(e.validated, req)
	e.mu.Unlock()

	var params exampleParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return fmt.Errorf("example params are not readable: %w", err)
	}
	if params.Value == nil || *params.Value < 1 {
		return errors.New("example params need a value of at least 1")
	}
	return nil
}

func (e *exampleOperationModule) Phase(op Operation) (nodestatus.Phase, bool) {
	if op.Status.Terminal() {
		return "", false
	}
	return examplePhase, true
}

func (e *exampleOperationModule) Run(_ context.Context, rt Runtime, req RunRequest) Outcome {
	e.mu.Lock()
	e.runCalls++
	e.ranParams = append(e.ranParams, string(req.Params))
	release, outcome := e.release, e.outcome
	e.mu.Unlock()

	if err := rt.StartStep(exampleStep); err != nil {
		return Outcome{Status: StatusFailed, Code: CodeModuleFailed, Error: err.Error()}
	}
	if release != nil {
		<-release
	}
	if err := rt.FinishStep(exampleStep, StepSucceeded, "", ""); err != nil {
		return Outcome{Status: StatusFailed, Code: CodeModuleFailed, Error: err.Error()}
	}
	return outcome
}

func (e *exampleOperationModule) ExecuteNode(_ context.Context, req NodeRequest) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nodeCalls = append(e.nodeCalls, req)
	return nil
}

func (e *exampleOperationModule) seen() (validated []CreateRequest, runCalls int,
	ranParams []string, nodeCalls []NodeRequest) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]CreateRequest(nil), e.validated...), e.runCalls,
		append([]string(nil), e.ranParams...), append([]NodeRequest(nil), e.nodeCalls...)
}

var (
	_ OperationModule     = (*exampleOperationModule)(nil)
	_ NodeOperationModule = (*exampleOperationModule)(nil)
)

// exampleRequest is a caller asking for the operation this file adds.
func exampleRequest(requestID string, params string) CreateRequest {
	return CreateRequest{
		Type:      exampleType,
		RequestID: requestID,
		Scope:     ScopeCluster,
		ClusterID: "cluster-test",
		Owner:     "alice@olares.com",
		Params:    json.RawMessage(params),
	}
}

// A type is whatever a module registered itself for. The registry the
// operation was added to accepts its name; the daemon's own module set, which
// this file did not touch, still refuses it.
func TestAddingOneModuleFileMakesItsTypeParseable(t *testing.T) {
	reg := registryWith(t, newExampleModule())

	got, err := reg.Parse(string(exampleType))
	if err != nil {
		t.Fatalf("Parse(%q) = %v, want a type the module registered itself for", exampleType, err)
	}
	if got != exampleType {
		t.Errorf("Parse(%q) = %q", exampleType, got)
	}

	if _, err := ParseType(string(exampleType)); err == nil {
		t.Errorf("ParseType(%q) accepted a type no production list was told about", exampleType)
	}
}

// The whole chain, driven by nothing but the module: the type is accepted,
// the module judges the params, the manager runs it, and the record settles
// on the outcome the module returned.
func TestAddingOneModuleFileCarriesAnOperationEndToEnd(t *testing.T) {
	module := newExampleModule()
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	op, err := m.Create(context.Background(), exampleRequest("client-1", `{"value":1}`))
	if err != nil {
		t.Fatalf("Create(%q): %v", exampleType, err)
	}
	settled := awaitTerminal(t, m, op.ID)

	if settled.Status != StatusSucceeded {
		t.Errorf("status = %q, want the outcome the module returned", settled.Status)
	}
	validated, runCalls, ranParams, _ := module.seen()
	if len(validated) != 1 {
		t.Errorf("Validate calls = %d, want 1", len(validated))
	}
	if runCalls != 1 {
		t.Fatalf("Run calls = %d, want 1", runCalls)
	}
	if ranParams[0] != `{"value":1}` {
		t.Errorf("Run params = %s, want the caller's own bytes", ranParams[0])
	}
	if len(settled.Steps) != 1 || settled.Steps[0].Name != exampleStep ||
		settled.Steps[0].Status != StepSucceeded {
		t.Errorf("steps = %+v, want the stage the module recorded", settled.Steps)
	}
}

// The module is the only thing that judges its params, and a refusal happens
// before the operation exists — so nothing is recorded and the cluster's
// single-operation lock is never taken.
func TestTheAddedModuleJudgesItsOwnParams(t *testing.T) {
	module := newExampleModule()
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	for _, params := range []string{``, `{}`, `{"value":0}`, `{"value":"lots"}`} {
		_, err := m.Create(context.Background(), exampleRequest("client-1", params))

		var refused *ModuleValidationError
		if !errors.As(err, &refused) {
			t.Fatalf("Create(params=%q) = %v, want the module's own refusal", params, err)
		}
		if refused.Type != exampleType {
			t.Errorf("refusal names %q, want %q", refused.Type, exampleType)
		}
	}
	if _, ok := m.GetByRequest("client-1"); ok {
		t.Error("a refused request was recorded anyway")
	}
	if _, runCalls, _, _ := module.seen(); runCalls != 0 {
		t.Errorf("Run calls = %d, want a refused request never started", runCalls)
	}
}

// What the cluster looks like while the operation happens is the module's
// answer, and nothing outside this file maps the type to a phase.
func TestTheAddedModuleDecidesTheClusterPhaseItImposes(t *testing.T) {
	module := newExampleModule()
	module.release = make(chan struct{})
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	op, err := m.Create(context.Background(), exampleRequest("client-1", `{"value":1}`))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	phase, ok := awaitPhase(t, m, examplePhase)
	if !ok || phase != examplePhase {
		t.Errorf("ActivePhase() = %q,%v, want the module's own phase", phase, ok)
	}

	close(module.release)
	awaitTerminal(t, m, op.ID)
	if phase, ok := m.ActivePhase(); ok {
		t.Errorf("ActivePhase() = %q, want nothing once the operation is over", phase)
	}
}

// The record settles on whatever the module reports, including a failure
// with a code the manager itself knows nothing about the meaning of.
func TestTheAddedModuleSettlesTheRecordOnItsOwnOutcome(t *testing.T) {
	module := newExampleModule()
	module.outcome = Outcome{Status: StatusPartiallyFailed, Code: CodeNodeUnreachable}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, module))

	op, err := m.Create(context.Background(), exampleRequest("client-1", `{"value":1}`))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	settled := awaitTerminal(t, m, op.ID)

	if settled.Status != StatusPartiallyFailed || settled.Code != CodeNodeUnreachable {
		t.Errorf("status = %q code = %q, want the module's own outcome", settled.Status, settled.Code)
	}
}

// The node-local half is the module's too, and it is reached through the
// same helper the generic node endpoint calls — with the registry the node
// holds, not a list of operations written into that endpoint.
func TestTheAddedModuleCarriesOutItsOwnNodeHalf(t *testing.T) {
	module := newExampleModule()
	reg := registryWith(t, module)
	req := NodeRequest{
		PeerRequest: PeerRequest{Type: exampleType, OperationID: "op-1", RequestID: "client-1"},
		Params:      json.RawMessage(`{"value":1}`),
	}

	if err := ExecuteNode(context.Background(), reg, req); err != nil {
		t.Fatalf("ExecuteNode(%q): %v", exampleType, err)
	}

	_, _, _, nodeCalls := module.seen()
	if len(nodeCalls) != 1 {
		t.Fatalf("ExecuteNode calls = %d, want 1", len(nodeCalls))
	}
	if string(nodeCalls[0].Params) != `{"value":1}` {
		t.Errorf("node params = %s, want the caller's own bytes", nodeCalls[0].Params)
	}
}

// A new operation gets the same treatment of its input the built-in ones
// get: the record keeps a digest so a retried request id can be told apart
// from a changed one, and never the params themselves.
func TestTheAddedModulesParamsAreNeverWrittenDown(t *testing.T) {
	const secretish = `{"value":424242}`
	module := newExampleModule()
	c := newCluster(master("master-1", "10.0.0.1"))
	m, dir := newManagerWith(t, c, registryWith(t, module))

	op, err := m.Create(context.Background(), exampleRequest("client-1", secretish))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	settled := awaitTerminal(t, m, op.ID)

	if settled.ParamsDigest == "" {
		t.Error("no digest was recorded, so a changed retry could not be told apart")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s): %v", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", entry.Name(), err)
		}
		if strings.Contains(string(raw), "424242") {
			t.Errorf("%s holds the module's raw params", entry.Name())
		}
	}
}
