package clusterop

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// fakeModule is an operation this daemon does not have: a test registers it to
// check that the manager creates, runs, reports and recovers whatever its
// registry holds, rather than the two operations that happen to be built in.
type fakeModule struct {
	typ         Type
	phase       nodestatus.Phase
	hasPhase    bool
	validateErr error
	outcome     Outcome

	mu        sync.Mutex
	validated []CreateRequest
	ran       []RunRequest
	recovered []Operation
}

func (f *fakeModule) Type() Type { return f.typ }

func (f *fakeModule) Validate(req CreateRequest) error {
	f.mu.Lock()
	f.validated = append(f.validated, req)
	f.mu.Unlock()
	return f.validateErr
}

func (f *fakeModule) Phase(Operation) (nodestatus.Phase, bool) { return f.phase, f.hasPhase }

func (f *fakeModule) Run(_ context.Context, _ Runtime, req RunRequest) Outcome {
	f.mu.Lock()
	f.ran = append(f.ran, req)
	f.mu.Unlock()
	return f.outcome
}

func (f *fakeModule) calls() (validated []CreateRequest, ran []RunRequest, recovered []Operation) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]CreateRequest(nil), f.validated...),
		append([]RunRequest(nil), f.ran...),
		append([]Operation(nil), f.recovered...)
}

// recoverableFake is fakeModule plus the recovery contract, which is how a
// module says a command that outlived the daemon can still be settled.
type recoverableFake struct {
	*fakeModule
	done chan struct{}
}

func (f *recoverableFake) Recover(_ context.Context, _ Runtime, op Operation) {
	f.mu.Lock()
	f.recovered = append(f.recovered, op)
	f.mu.Unlock()
	close(f.done)
}

func newFake(typ Type) *fakeModule {
	return &fakeModule{
		typ:      typ,
		phase:    nodestatus.Phase("testing"),
		hasPhase: true,
		outcome:  Outcome{Status: StatusSucceeded},
	}
}

func registryWith(t *testing.T, modules ...OperationModule) *ModuleRegistry {
	t.Helper()
	r := NewRegistry()
	for _, module := range modules {
		if err := r.Register(module); err != nil {
			t.Fatalf("Register(%q): %v", module.Type(), err)
		}
	}
	return r
}

func newManagerWith(t *testing.T, c *cluster, registry *ModuleRegistry) (*Manager, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "operations")
	m, err := NewManagerWithRegistry(c.deps(t, dir), registry)
	if err != nil {
		t.Fatalf("NewManagerWithRegistry: %v", err)
	}
	return m, dir
}

func createFake(t *testing.T, m *Manager, typ Type, requestID string) (Operation, error) {
	t.Helper()
	return m.Create(context.Background(), CreateRequest{
		Type:      typ,
		RequestID: requestID,
		Owner:     "alice@olares.com",
	})
}

// A manager without a registry has no way to carry anything out, and building
// one that silently accepts nothing would only move the failure to the first
// operation somebody asks for.
func TestNewManagerWithRegistryRefusesToBuildWithoutOne(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")

	if _, err := NewManagerWithRegistry(c.deps(t, dir), nil); err == nil {
		t.Fatal("NewManagerWithRegistry(deps, nil) = nil error, want a refusal")
	}
}

// The built-in operations reach a plain NewManager because they registered
// themselves, not because the manager names them.
func TestNewManagerCarriesTheBuiltInOperations(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)

	for _, typ := range []Type{TypeReboot, TypeShutdown, TypeSetSSHPassword} {
		if _, ok := m.registry.Lookup(typ); !ok {
			t.Errorf("NewManager cannot carry out %q", typ)
		}
	}
}

// What the manager accepts is exactly what its own registry holds: a type it
// does not carry is refused before anything is recorded, and one it does is
// accepted even though this daemon has never heard of it otherwise.
func TestCreateAcceptsExactlyWhatItsRegistryHolds(t *testing.T) {
	fake := newFake(Type("bake-cake"))
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, fake))

	if _, err := createFake(t, m, TypeReboot, "client-1"); err == nil {
		t.Error("Create(reboot) = nil error, want a refusal from a registry without it")
	}
	op, err := createFake(t, m, fake.typ, "client-2")
	if err != nil {
		t.Fatalf("Create(%q): %v", fake.typ, err)
	}
	if op.Type != fake.typ {
		t.Errorf("Type = %q, want %q", op.Type, fake.typ)
	}
}

// The module is asked before the operation exists. An operation the module
// would refuse must not be recorded, must not take the cluster's
// single-operation lock, and must not leave the next request conflicting with
// something that never ran.
func TestCreateAsksTheModuleBeforeTheOperationExists(t *testing.T) {
	refused := errors.New("this module will not do that")
	fake := newFake(Type("bake-cake"))
	fake.validateErr = refused
	c := newCluster(master("master-1", "10.0.0.1"))
	m, dir := newManagerWith(t, c, registryWith(t, fake))

	if _, err := createFake(t, m, fake.typ, "client-1"); !errors.Is(err, refused) {
		t.Fatalf("Create() = %v, want the module's own refusal", err)
	}

	validated, ran, _ := fake.calls()
	if len(validated) != 1 {
		t.Fatalf("Validate calls = %d, want 1", len(validated))
	}
	if len(ran) != 0 {
		t.Errorf("Run calls = %d, want a refused operation never started", len(ran))
	}
	if _, ok := m.GetByRequest("client-1"); ok {
		t.Error("a refused operation was recorded")
	}
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	stored, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(stored) != 0 {
		t.Errorf("stored operations = %d, want a refused operation never written", len(stored))
	}

	// The lock is free, which is only true if the refusal left nothing active.
	fake.validateErr = nil
	if _, err := createFake(t, m, fake.typ, "client-2"); err != nil {
		t.Errorf("Create() after a refusal = %v, want the cluster still free", err)
	}
}

// A module refusing a request it cannot carry out is the caller's mistake,
// not this daemon's failure. Create says so in a way a route can act on
// without reading the module's own sentence, which is text this package
// cannot vouch for.
func TestCreateReportsAModuleRefusalAsAValidationError(t *testing.T) {
	refused := errors.New("this module will not do that")
	fake := newFake(Type("bake-cake"))
	fake.validateErr = refused
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, fake))

	_, err := createFake(t, m, fake.typ, "client-1")

	var invalid *ModuleValidationError
	if !errors.As(err, &invalid) {
		t.Fatalf("Create() = %v, want a refusal a route can map to a bad request", err)
	}
	if invalid.Type != fake.typ {
		t.Errorf("Type = %q, want the module that refused", invalid.Type)
	}
	if !errors.Is(err, refused) {
		t.Errorf("Create() = %v, want the module's own refusal still reachable", err)
	}
}

// Running an operation is the module's own work, and the manager settles the
// record on whatever the module returns.
func TestRunHandsTheOperationToItsModule(t *testing.T) {
	fake := newFake(Type("bake-cake"))
	fake.outcome = Outcome{Status: StatusFailed, Code: CodePowerUnsupported}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, fake))

	op, err := createFake(t, m, fake.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	settled := awaitTerminal(t, m, op.ID)

	if settled.Status != StatusFailed || settled.Code != CodePowerUnsupported {
		t.Errorf("status = %q code = %q, want the module's own outcome", settled.Status, settled.Code)
	}
	if _, ran, _ := fake.calls(); len(ran) != 1 {
		t.Errorf("Run calls = %d, want 1", len(ran))
	}
}

// The credentials that authorize an operation reach the module that needs
// them and nothing else: no field of the record can hold them.
func TestRunCarriesTheCredentialsToTheModuleAndNowhereElse(t *testing.T) {
	fake := newFake(Type("bake-cake"))
	c := newCluster(master("master-1", "10.0.0.1"))
	m, dir := newManagerWith(t, c, registryWith(t, fake))

	if _, err := m.Create(context.Background(), CreateRequest{
		Type:      fake.typ,
		RequestID: "client-1",
		Owner:     "alice@olares.com",
		Creds:     Credentials{Token: "secret-token", Signature: "secret-jws"},
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	op, _ := m.GetByRequest("client-1")
	awaitTerminal(t, m, op.ID)

	_, ran, _ := fake.calls()
	if len(ran) != 1 {
		t.Fatalf("Run calls = %d, want 1", len(ran))
	}
	if ran[0].Creds.Token != "secret-token" || ran[0].Creds.Signature != "secret-jws" {
		t.Errorf("Run credentials = %+v, want the caller's own", ran[0].Creds)
	}
	requireNoCredentialsOnDisk(t, dir, "secret-token", "secret-jws")
}

// requireNoCredentialsOnDisk reads every record the manager wrote and fails
// if any of them contains a secret the caller handed in.
func requireNoCredentialsOnDisk(t *testing.T, dir string, secrets ...string) {
	t.Helper()
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
		for _, secret := range secrets {
			if strings.Contains(string(raw), secret) {
				t.Errorf("%s holds the caller's credential %q", entry.Name(), secret)
			}
		}
	}
}

// The phase a running operation imposes on the cluster is the module's
// answer. The manager reports it without knowing what the operation is.
func TestActivePhaseIsTheModulesOwn(t *testing.T) {
	fake := newFake(Type("bake-cake"))
	fake.phase = nodestatus.Phase("baking")
	release := make(chan struct{})
	blocking := &blockingModule{fakeModule: fake, release: release}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, blocking))

	op, err := createFake(t, m, fake.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer func() {
		close(release)
		awaitTerminal(t, m, op.ID)
	}()

	phase, ok := awaitPhase(t, m, nodestatus.Phase("baking"))
	if !ok || phase != nodestatus.Phase("baking") {
		t.Errorf("ActivePhase() = %q,%v, want the module's own phase", phase, ok)
	}
}

// A module that imposes no phase leaves the node's own alone, however busy
// its operation is.
func TestActivePhaseIsAbsentWhenTheModuleImposesNone(t *testing.T) {
	fake := newFake(Type("bake-cake"))
	fake.hasPhase = false
	release := make(chan struct{})
	blocking := &blockingModule{fakeModule: fake, release: release}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, blocking))

	op, err := createFake(t, m, fake.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer func() {
		close(release)
		awaitTerminal(t, m, op.ID)
	}()

	if phase, ok := m.ActivePhase(); ok {
		t.Errorf("ActivePhase() = %q,true, want no phase at all", phase)
	}
}

// blockingModule holds its operation open until the test lets it settle, which
// is the only way to observe a manager with something actually in flight.
type blockingModule struct {
	*fakeModule
	release chan struct{}
}

func (b *blockingModule) Run(ctx context.Context, rt Runtime, req RunRequest) Outcome {
	b.fakeModule.Run(ctx, rt, req)
	<-b.release
	return Outcome{Status: StatusSucceeded}
}

func awaitPhase(t *testing.T, m *Manager, want nodestatus.Phase) (nodestatus.Phase, bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if phase, ok := m.ActivePhase(); ok && phase == want {
			return phase, true
		}
		time.Sleep(time.Millisecond)
	}
	return m.ActivePhase()
}

// A module may put its own outcome on the record — a node shutdown does,
// because the record has to say the command went out while the machine is on
// its way down. When it says so, the manager does not write it a second time
// and does not rely on being refused for that to happen.
func TestSettleLeavesAnOutcomeTheModuleAlreadyRecorded(t *testing.T) {
	rt, m, id := newTestRuntime(t, StatusRunning)

	m.settle(rt, id, Outcome{Status: StatusSucceeded}.alreadyRecorded())

	got, _ := m.Get(id)
	if got.Status != StatusRunning {
		t.Fatalf("status = %q, want the record left exactly as the module left it", got.Status)
	}
}

// A module that reports something the record cannot hold has still stopped
// working. Leaving the operation running would hold the cluster's
// single-operation lock until the daemon restarts, so it is settled — as
// failed, which is the only thing known about it.
func TestAnUnusableOutcomeStillSettlesTheOperation(t *testing.T) {
	fake := newFake(Type("bake-cake"))
	// Not a status an operation may end on, so Complete refuses it.
	fake.outcome = Outcome{Status: StatusRunning}
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManagerWith(t, c, registryWith(t, fake))

	op, err := createFake(t, m, fake.typ, "client-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	settled := awaitTerminal(t, m, op.ID)

	if settled.Status != StatusFailed || settled.Code != CodeModuleFailed {
		t.Errorf("status = %q code = %q, want failed/%s", settled.Status, settled.Code, CodeModuleFailed)
	}
	if settled.Error != reasonFor(CodeModuleFailed) {
		t.Errorf("error = %q, want the reviewed sentence %q", settled.Error, reasonFor(CodeModuleFailed))
	}
	// The lock is free again, which is the reason any of this matters. The
	// second operation is waited out rather than left running: its own run
	// goroutine writes to the store, and this test's store is a temporary
	// directory the test framework deletes the moment the test returns.
	next, err := createFake(t, m, fake.typ, "client-2")
	if err != nil {
		t.Fatalf("Create() after an unusable outcome = %v, want the cluster free again", err)
	}
	awaitTerminal(t, m, next.ID)
}

// A module that handed the record a command_issued outcome and then returned
// something the record cannot hold has still committed real state: the
// command went out, and the machine it went to is on its way down. The
// fallback settlement for an unusable outcome exists to release a record
// nothing else will move, so it must leave that one exactly as the module
// left it rather than reporting a failure for a command that was issued.
func TestAnUnusableOutcomeLeavesARecordTheModuleAlreadySettled(t *testing.T) {
	rt, m, id := buildCommandIssuedRuntime(t, time.Minute)

	// Not a status an operation may end on, so Complete refuses it and the
	// fallback runs.
	m.settle(rt, id, Outcome{Status: StatusRunning})

	got, ok := m.Get(id)
	if !ok {
		t.Fatal("the stored operation was lost")
	}
	if got.Status != StatusCommandIssued {
		t.Fatalf("status = %q, want the issued command left on the record", got.Status)
	}
	if got.Code == CodeModuleFailed {
		t.Errorf("code = %q, want the command_issued record not overwritten", got.Code)
	}
}

// A command that outlived the daemon that issued it goes back to the module
// that issued it, because only that module knows what would settle it.
func TestARecoverableModuleIsHandedItsUnfinishedCommand(t *testing.T) {
	fake := newFake(Type("bake-cake"))
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")
	id := storeCommandIssued(t, c, dir, fake.typ)

	recoverable := &recoverableFake{fakeModule: fake, done: make(chan struct{})}
	if _, err := NewManagerWithRegistry(c.deps(t, dir), registryWith(t, recoverable)); err != nil {
		t.Fatalf("NewManagerWithRegistry: %v", err)
	}

	select {
	case <-recoverable.done:
	case <-time.After(time.Second):
		t.Fatal("the module was never handed its unfinished command")
	}
	_, _, recovered := fake.calls()
	if len(recovered) != 1 || recovered[0].ID != id {
		t.Fatalf("Recover(%+v), want the stored command_issued operation %s", recovered, id)
	}
}

// A module that offers no recovery leaves the record exactly as it was found.
// For a machine that was told to switch off, that is the only honest answer.
func TestAModuleWithoutRecoveryLeavesItsCommandAlone(t *testing.T) {
	fake := newFake(Type("bake-cake"))
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")
	id := storeCommandIssued(t, c, dir, fake.typ)

	m, err := NewManagerWithRegistry(c.deps(t, dir), registryWith(t, fake))
	if err != nil {
		t.Fatalf("NewManagerWithRegistry: %v", err)
	}

	got, ok := m.Get(id)
	if !ok {
		t.Fatalf("the stored operation %s was lost", id)
	}
	if got.Status != StatusCommandIssued {
		t.Errorf("status = %q, want it left at command_issued", got.Status)
	}
}

// storeCommandIssued writes the record a daemon that went down mid-operation
// leaves behind: the command was handed out, and nothing here knows what it
// did.
func storeCommandIssued(t *testing.T, c *cluster, dir string, typ Type) string {
	t.Helper()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	at := time.Unix(1700000000, 0).UTC()
	finished := at
	op := Operation{
		ID:                 "op-stored-1",
		Type:               typ,
		RequestID:          "client-stored-1",
		Owner:              "alice@olares.com",
		Status:             StatusCommandIssued,
		CreatedAt:          at,
		UpdatedAt:          at,
		FinishedAt:         &finished,
		CommandIssuedUntil: at.Add(time.Hour),
		Steps:              []Step{},
		Nodes:              []NodeResult{},
	}
	if err := store.Save(op); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return op.ID
}
