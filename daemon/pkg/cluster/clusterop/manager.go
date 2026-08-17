package clusterop

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"k8s.io/klog/v2"
)

// defaultRetention bounds how many operation records the state directory
// keeps. Records are tiny and a cluster performs very few power operations,
// but a daemon that runs for years must not accumulate them without limit.
const defaultRetention = 20

// Credentials authorize one run. They are passed to the seams that talk to
// other nodes and are never written anywhere: no field of Operation can hold
// them, and the run goroutine drops them when it returns.
type Credentials struct {
	// Token is the caller's verified access token, forwarded the way the
	// existing fan-out forwards it.
	Token string

	// Signature is the caller's JWS. The node-local power endpoint applies
	// the same owner check the single-node power commands apply, so the
	// second hop is not weaker than the first.
	Signature string
}

// PeerRequest is the body the master sends to a node's own power endpoint. It
// names no target: the node it reaches is the node it acts on.
type PeerRequest struct {
	Type        Type   `json:"type"`
	OperationID string `json:"operationId"`

	// RequestID is the caller's own id for this operation, and it is what the
	// owner's signature is bound to. The receiving node checks the two
	// against each other, so a signature captured from one operation cannot
	// be replayed to power a node during another.
	RequestID string `json:"requestId"`
	Scope     string `json:"scope"`
	Target    string `json:"target"`
	ClusterID string `json:"clusterId"`
}

// DispatchOutcome is the per-node result of handing out one power command.
// An empty Code means the node accepted it.
type DispatchOutcome struct {
	NodeName string
	Code     string
	Err      string
}

// Timeouts bound the waits an operation performs.
type Timeouts struct {
	// Poll is how often a restarting node is probed.
	Poll time.Duration
	// Down is how long a node may take to stop answering after being told to
	// reboot. A node that never stops answering never rebooted.
	Down time.Duration
	// Ready is how long a node may take to come back and be Ready again.
	Ready time.Duration
	// Stage is how long one node may take over one upgrade stage. It is an
	// order of magnitude larger than the power timeouts because a stage
	// downloads a release, imports images and upgrades a container runtime.
	Stage time.Duration
}

func (t Timeouts) withDefaults() Timeouts {
	if t.Poll <= 0 {
		t.Poll = 5 * time.Second
	}
	if t.Down <= 0 {
		t.Down = 3 * time.Minute
	}
	if t.Ready <= 0 {
		t.Ready = 15 * time.Minute
	}
	if t.Stage <= 0 {
		t.Stage = 60 * time.Minute
	}
	return t
}

// Deps are every side effect the orchestrator can have. All of them are
// injected: a unit test drives a whole cluster reboot without a cluster, a
// network, a clock or a machine that can be powered off.
type Deps struct {
	Store OperationStore

	// Inventory is the node directory, including nodes that are NotReady or
	// have no address — a precheck cannot refuse what it cannot see.
	Inventory func(ctx context.Context) ([]inventory.Node, error)

	// Inspect reads one node's own status, which is where its power
	// capabilities are declared.
	Inspect func(ctx context.Context, node inventory.Node, creds Credentials) (nodestatus.Status, error)

	// Dispatch hands the power command to each node's own endpoint.
	Dispatch func(ctx context.Context, nodes []inventory.Node, req PeerRequest, creds Credentials) []DispatchOutcome

	// Observe is the cluster's own view of every node: Ready, and the boot it
	// is on. Watching a restart this way needs no credential for the node
	// being watched, which matters because the node-local status endpoint is
	// behind a user access token and a cluster mid-reboot is exactly when the
	// service issuing those is unavailable.
	Observe func(ctx context.Context) (map[string]inventory.Observation, error)

	// LocalPowerSupport answers whether this machine's own execution point
	// would carry out the operation: not what the control node declares to
	// the cluster, but what PowerHost checks a moment before it runs the
	// command. The precheck asks it first, because the control node goes
	// last and a refusal discovered then has already cost every worker.
	LocalPowerSupport func(t Type) error

	// HostBootID reports the boot the machine running this daemon is on.
	HostBootID func() (string, error)

	// PowerSelf powers the machine this daemon runs on. It is the last thing
	// any cluster power operation does.
	PowerSelf func(ctx context.Context, t Type) error

	// Upgrade is everything an upgrade needs and a power operation does not.
	//
	// It is one pointer rather than six fields because it is one decision: a
	// daemon either orchestrates upgrades or it does not, and a half-wired
	// set is not a state anything should have to cope with at each step. nil
	// is the normal answer on a node that was built without an upgrade
	// executor; the upgrade module refuses once, up front, rather than every
	// function on the path checking for itself and quietly carrying on.
	Upgrade *UpgradeDeps

	// Base is what a run executes on. It is deliberately not derived from the
	// request that asked for the operation: that context is recycled as soon
	// as the response is written, and a cluster reboot outlives it by minutes.
	// It is also not the daemon's shutdown context, because on the control
	// node olaresd is stopped by the very command the run just issued.
	Base context.Context

	Now   func() time.Time
	Sleep func(ctx context.Context, d time.Duration) error
	NewID func() string

	Retention int
	Timeouts  Timeouts

	// recoveryTimeout, if set, replaces startupRecoveryTimeout for the one
	// recovery this daemon waits for while it is still starting. It is
	// unexported for the same reason recoveryDone below is: how long
	// olaresd is willing to be held up by a module is a framework decision,
	// not something a caller configures. A test in this package sets it so
	// that proving the limit works does not take the limit's own length.
	recoveryTimeout time.Duration

	// recoveryDone, if set, receives the id of every operation whose
	// asynchronous recovery (Manager.resume) has just returned, whether it
	// panicked or not. It is unexported: nothing outside this package may
	// depend on when a background recovery goroutine finishes, and
	// production never sets it. A test in this package uses it instead of
	// polling the record on a timer to know a panicking Recover has
	// already run its course.
	recoveryDone chan string
}

// storeIsNil reports whether s is unusable as an OperationStore: either the
// interface itself is nil, or it holds a nil pointer of a concrete type — a
// typed nil, which compares unequal to nil as an interface but panics the
// moment anything calls through it.
func storeIsNil(s OperationStore) bool {
	if s == nil {
		return true
	}
	v := reflect.ValueOf(s)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

func (d Deps) validate() error {
	var missing []string
	if storeIsNil(d.Store) {
		missing = append(missing, "Store")
	}
	if d.Inventory == nil {
		missing = append(missing, "Inventory")
	}
	if d.Inspect == nil {
		missing = append(missing, "Inspect")
	}
	if d.Dispatch == nil {
		missing = append(missing, "Dispatch")
	}
	if d.Observe == nil {
		missing = append(missing, "Observe")
	}
	if d.LocalPowerSupport == nil {
		missing = append(missing, "LocalPowerSupport")
	}
	if d.HostBootID == nil {
		missing = append(missing, "HostBootID")
	}
	if d.PowerSelf == nil {
		missing = append(missing, "PowerSelf")
	}
	if len(missing) > 0 {
		return fmt.Errorf("cluster operation manager is missing: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (d Deps) withDefaults() Deps {
	if d.Base == nil {
		d.Base = context.Background()
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	if d.Sleep == nil {
		d.Sleep = sleepCtx
	}
	if d.NewID == nil {
		d.NewID = newOperationID
	}
	if d.Retention <= 0 {
		d.Retention = defaultRetention
	}
	if d.recoveryTimeout <= 0 {
		d.recoveryTimeout = startupRecoveryTimeout
	}
	d.Timeouts = d.Timeouts.withDefaults()
	return d
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func newOperationID() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "op-" + time.Now().UTC().Format("20060102150405.000000000")
	}
	return "op-" + time.Now().UTC().Format("20060102-150405") + "-" + hex.EncodeToString(buf)
}

// CreateRequest is a caller asking for one cluster power operation.
type CreateRequest struct {
	Type      Type
	RequestID string
	Scope     string
	Target    string
	ClusterID string
	Owner     string
	// Params carries optional module input for idempotency only. It is not
	// bound into the caller's JWS, stays in memory, and is never persisted.
	Params json.RawMessage
	Creds  Credentials
}

// Errors a caller can act on.
var (
	ErrRequestIDRequired = errors.New("requestId is required")
	ErrOwnerRequired     = errors.New("owner is required")
)

// ModuleValidationError is a module refusing a request before anything is
// recorded: the request is one this daemon knows the type of and cannot
// carry out as asked. It is distinct from the errors above so a route can
// answer "bad request" without reading the module's own sentence, which is
// text written outside this package and never shown to a caller.
type ModuleValidationError struct {
	Type Type
	Err  error
}

func (e *ModuleValidationError) Error() string {
	return fmt.Sprintf("cluster operation %s refused the request: %v", e.Type, e.Err)
}

func (e *ModuleValidationError) Unwrap() error { return e.Err }

// ConflictError refuses a second cluster power operation while one is running.
// Two of them at once would race over the same machines, and the second one's
// precheck would be answered by nodes the first one is already powering down.
type ConflictError struct {
	ActiveID   string
	ActiveType Type
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("cluster operation %s (%s) is already in progress", e.ActiveID, e.ActiveType)
}

// RequestConflictError refuses a requestId already bound to another intent.
type RequestConflictError struct {
	RequestID  string
	ExistingID string
}

func (e *RequestConflictError) Error() string {
	return fmt.Sprintf("requestId %q is already bound to cluster operation %s", e.RequestID, e.ExistingID)
}

// Manager owns every cluster power operation this master has performed.
type Manager struct {
	deps Deps

	// registry is the one module set this manager answers from. Creating,
	// running, reporting a phase for and recovering an operation all look
	// the type up here, so a manager can never carry an operation out
	// through a module that a different part of it would not recognize.
	registry *ModuleRegistry

	mu            sync.Mutex
	ops           map[string]*Operation
	order         []string
	byRequest     map[string]string
	activeID      string
	persistFailed map[string]bool
}

// NewManager loads the recorded operations and returns a manager over the
// modules built into this daemon.
func NewManager(deps Deps) (*Manager, error) {
	return NewManagerWithRegistry(deps, DefaultRegistry())
}

// NewManagerWithRegistry is NewManager over an explicit set of modules. It is
// how a test drives the manager against a module of its own; the daemon
// passes the default registry, which is where the built-in power operations
// register themselves.
func NewManagerWithRegistry(deps Deps, registry *ModuleRegistry) (*Manager, error) {
	if registry == nil {
		return nil, errors.New("cluster operation manager is missing: ModuleRegistry")
	}
	if err := deps.validate(); err != nil {
		return nil, err
	}
	deps = deps.withDefaults()

	m := &Manager{
		deps:          deps,
		registry:      registry,
		ops:           map[string]*Operation{},
		byRequest:     map[string]string{},
		persistFailed: map[string]bool{},
	}

	stored, err := deps.Store.List()
	if err != nil {
		return nil, fmt.Errorf("load cluster operations: %w", err)
	}
	for i := range stored {
		op := stored[i]
		m.ops[op.ID] = &op
		m.order = append(m.order, op.ID)

		// A retired record is kept and readable by id, but it no longer
		// answers for its request: the operation that replaced it does. See
		// Operation.SupersededBy.
		if op.SupersededBy != "" {
			continue
		}
		// Two unretired records claiming one request id is a contradiction,
		// and it is resolved rather than refused. Refusing means returning an
		// error from here, which means olaresd does not start — and it does
		// not start next time either, because nothing about the records will
		// have changed. Trading every function this daemon has for one
		// ambiguous pair of records is not a trade worth making, and picking
		// the later one is what the caller asking by that id wants in every
		// case anybody has been able to describe.
		if id, ok := m.byRequest[op.RequestID]; ok && id != op.ID {
			klog.Errorf("clusterop: operations %s and %s both claim request %q; answering with the later one",
				id, op.ID, op.RequestID)
			if kept, ok := m.ops[id]; ok && kept.CreatedAt.After(op.CreatedAt) {
				continue
			}
		}
		m.byRequest[op.RequestID] = op.ID
	}

	// Every loaded record now has an entry in m.ops/m.order/m.byRequest, so
	// a module's Recover — which mutates through the same checked Runtime
	// any other caller uses — can look any of them up. See
	// recoverLoadedOperations for what happens to each: an unknown type is
	// settled, a non-terminal record without a RecoverableModule is marked
	// interrupted, and one with a RecoverableModule is handed to Recover
	// first and only marked interrupted if that leaves it still moving.
	unfinished := m.recoverLoadedOperations()

	// Under the lock, because a module's Recover that ran out of time up
	// there is still running and still commits through the same map.
	m.mu.Lock()
	for _, id := range m.order {
		if op := m.ops[id]; op != nil && m.operationActive(op, deps.Now()) {
			m.activeID = id
		}
	}
	m.mu.Unlock()
	for _, id := range unfinished {
		m.resume(id)
	}
	return m, nil
}

// resume hands an operation whose command outlived this daemon back to the
// module that issued it. Only that module knows what evidence would settle
// it, and a module that offers no recovery leaves the record exactly as it
// was found — which is the only honest answer for a machine that was told to
// switch off.
//
// It is called from NewManagerWithRegistry while the manager is still being
// built. The goroutine it starts runs concurrently with the rest of the
// constructor and with every later caller, and reaches the record only
// through a Runtime, which takes the lock like everything else.
func (m *Manager) resume(id string) {
	op, ok := m.Get(id)
	if !ok {
		return
	}
	module, ok := m.registry.Lookup(op.Type)
	if !ok {
		return
	}
	recoverable, ok := module.(RecoverableModule)
	if !ok {
		return
	}
	done := m.deps.recoveryDone
	go func() {
		m.safeRecover(m.deps.Base, recoverable, id)
		if done != nil {
			done <- id
		}
	}()
}

func sameIntent(op *Operation, req CreateRequest, opType Type, paramsDigest, emptyParamsDigest string) bool {
	storedDigest := op.ParamsDigest
	if storedDigest == "" {
		storedDigest = emptyParamsDigest
	}
	return op.Owner == req.Owner &&
		op.Type == opType &&
		op.Scope == req.Scope &&
		op.Target == req.Target &&
		op.ClusterID == req.ClusterID &&
		storedDigest == paramsDigest
}

// supersedable reports whether an identical request should start a new
// operation instead of being answered with this one.
//
// Only a failure is ever replaced. A run that is still going is what the
// caller is asking for, and one that succeeded is the answer to the question —
// repeating the work because someone asked twice is how an upgrade gets run
// twice. See RetryableModule for why this is the module's decision.
func supersedable(module OperationModule, op Operation) bool {
	if !op.Status.Terminal() || op.Status == StatusSucceeded {
		return false
	}
	retryable, ok := module.(RetryableModule)
	return ok && retryable.RetryAfterFailure(op)
}

// Create starts a cluster power operation, or returns the one this request
// already started. It returns as soon as the operation is recorded: the power
// commands themselves are issued by the run it launches.
// The caller's context is deliberately unused: see Deps.Base.
func (m *Manager) Create(_ context.Context, req CreateRequest) (Operation, error) {
	opType, err := m.registry.Parse(string(req.Type))
	if err != nil {
		return Operation{}, err
	}
	module, ok := m.registry.Lookup(opType)
	if !ok {
		// Parse only accepts a type this registry holds, so reaching here
		// means it lost the module in between.
		return Operation{}, fmt.Errorf("unsupported cluster operation type %q", req.Type)
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return Operation{}, ErrRequestIDRequired
	}
	if strings.TrimSpace(req.Owner) == "" {
		return Operation{}, ErrOwnerRequired
	}
	paramsDigest, err := DigestParams(req.Params)
	if err != nil {
		return Operation{}, err
	}
	// Asked before anything is recorded or started: what the module cannot
	// carry out must not become an operation that exists, holds the
	// cluster's single-operation lock, and then fails. A module that cannot
	// answer at all is the same for those purposes and different for the
	// caller's: nothing is recorded either way, but a request nothing judged
	// is not a request that was refused. See SafeValidate.
	refusal, answered := SafeValidate(module, req)
	if !answered {
		return Operation{}, ErrModuleFailed
	}
	if refusal != nil {
		return Operation{}, &ModuleValidationError{Type: opType, Err: refusal}
	}

	m.mu.Lock()
	supersededID := ""
	if id, ok := m.byRequest[req.RequestID]; ok {
		existing := m.ops[id]
		if !sameIntent(existing, req, opType, paramsDigest, emptyParamsDigest) {
			m.mu.Unlock()
			return Operation{}, &RequestConflictError{RequestID: req.RequestID, ExistingID: id}
		}
		if !supersedable(module, *existing) {
			cloned := existing.Clone()
			m.mu.Unlock()
			return cloned, nil
		}
		// The module wants another attempt. The old record stays on disk as
		// history, but it has to stop claiming the request id before a second
		// record starts claiming it — see Operation.SupersededBy.
		supersededID = id
	}
	if active := m.activeOperationLocked(); active != nil {
		err := &ConflictError{ActiveID: active.ID, ActiveType: active.Type}
		m.mu.Unlock()
		return Operation{}, err
	}

	at := m.deps.Now()
	newID := m.deps.NewID()

	// Written before the record that replaces it, and the order is what keeps
	// the invariant true on disk as well as in memory. Marked first, then
	// interrupted: the request id belongs to nobody, and the next request
	// starts a fresh operation. Written second and interrupted: two records
	// claim one request id, which is the state the loader refuses to start
	// from — so the daemon would not come back at all.
	if supersededID != "" {
		superseded := m.ops[supersededID]
		superseded.SupersededBy = newID
		if err := m.deps.Store.Save(*superseded); err != nil {
			superseded.SupersededBy = ""
			m.mu.Unlock()
			return Operation{}, fmt.Errorf("retire the previous attempt at %s: %w", req.RequestID, err)
		}
	}

	op := &Operation{
		ID:           newID,
		Type:         opType,
		RequestID:    req.RequestID,
		Scope:        req.Scope,
		Target:       req.Target,
		ClusterID:    req.ClusterID,
		Owner:        req.Owner,
		ParamsDigest: paramsDigest,
		Status:       StatusPending,
		CreatedAt:    at,
		UpdatedAt:    at,
		Steps:        []Step{},
		Nodes:        []NodeResult{},
	}
	m.ops[op.ID] = op
	m.order = append(m.order, op.ID)
	m.byRequest[req.RequestID] = op.ID
	m.activeID = op.ID
	if err := m.deps.Store.Save(*op); err != nil {
		delete(m.ops, op.ID)
		m.order = m.order[:len(m.order)-1]
		// The request id is left unclaimed rather than handed back to the
		// superseded record: that one has been written to disk saying it no
		// longer answers for it, and an in-memory map that disagreed with the
		// disk would be undone by the next restart anyway.
		delete(m.byRequest, req.RequestID)
		m.activeID = ""
		m.mu.Unlock()
		return Operation{}, fmt.Errorf("record cluster operation: %w", err)
	}
	m.pruneLocked()
	created := op.Clone()
	m.mu.Unlock()

	// Detached from the request: the caller polls by id, and a browser that
	// navigates away must not cancel a reboot half way through. The
	// credentials and the params go with it and no further: neither is
	// reachable from the record this returns.
	go m.run(m.deps.Base, created.ID, created.Type, RunRequest{Creds: req.Creds, Params: req.Params})

	return created, nil
}

// Get returns a copy of one operation.
func (m *Manager) Get(id string) (Operation, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	op, ok := m.ops[id]
	if !ok {
		return Operation{}, false
	}
	return op.Clone(), true
}

// GetByRequest returns a copy of the operation bound to requestID.
func (m *Manager) GetByRequest(requestID string) (Operation, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.byRequest[requestID]
	if !ok {
		return Operation{}, false
	}
	op, ok := m.ops[id]
	if !ok {
		return Operation{}, false
	}
	return op.Clone(), true
}

// ActiveOfType is the operation of this type currently in flight, if there is
// one. It exists so a cancel that has lost the request id — the upgrade
// target has been deleted, and the version that names the run went with it —
// can still find the run to stop.
func (m *Manager) ActiveOfType(t Type) (Operation, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	op := m.activeOperationLocked()
	if op == nil || op.Type != t {
		return Operation{}, false
	}
	return op.Clone(), true
}

// ActivePhase is the cluster phase implied by the operation in flight, or by
// one whose command has been issued and whose machine has not gone down yet. ok
// is false when there is neither, which leaves the caller's own phase alone.
// The module is asked outside the manager's lock, and about a copy: a module
// is other people's code, and neither reaching back into the manager nor
// writing to the stored record is something it should be able to do from
// here.
func (m *Manager) ActivePhase() (nodestatus.Phase, bool) {
	m.mu.Lock()
	active := m.activeOperationLocked()
	if active == nil {
		m.mu.Unlock()
		return "", false
	}
	op := active.Clone()
	m.mu.Unlock()

	// A command_issued operation is terminal, and PhaseFor answers nothing
	// for it. Here it is exactly the case that matters: the command has gone
	// out and the machine has not gone down yet.
	if op.Status != StatusCommandIssued && op.Status.Terminal() {
		return "", false
	}
	return phaseOf(m.registry, &op)
}

func (m *Manager) activeOperationLocked() *Operation {
	now := m.deps.Now()
	if m.activeID != "" {
		if active := m.ops[m.activeID]; active != nil && m.operationActive(active, now) {
			return active
		}
		m.activeID = ""
	}
	for _, id := range m.order {
		if active := m.ops[id]; active != nil && m.operationActive(active, now) {
			m.activeID = id
			return active
		}
	}
	return nil
}

func (m *Manager) operationActive(op *Operation, now time.Time) bool {
	if !op.Status.Terminal() {
		return true
	}
	return op.Status == StatusCommandIssued && now.Before(op.CommandIssuedUntil)
}

// settledCheck decides whether an operation is still allowed to move. There
// are two of them — one for a run, one for a recovery — and a runtime is
// built with the one that applies to it, so no caller chooses at the point of
// mutation which rules it would rather be judged by.
type settledCheck func(*Operation) error

// rejectSettled is the validate check shared by every checked mutation that
// has no state to inspect beyond "is this operation still allowed to move".
// An operation is settled — and every further mutation refused — once it is
// no longer operationActive: a normal terminal status (succeeded, failed,
// partially_failed), or a command_issued operation whose grace deadline has
// already passed. A command_issued operation still inside that deadline
// stays mutable on purpose: it is exactly the "confirm what a command
// already issued actually did" window a RecoverableModule needs in order to
// still clear the deadline, finish a step, update a node, or Complete to a
// final status once the outcome is known.
func (m *Manager) rejectSettled(op *Operation) error {
	if !m.operationActive(op, m.deps.Now()) {
		return ErrOperationTerminal
	}
	return nil
}

// rejectSettledDuringRecovery is rejectSettled for the one caller that
// exists to settle a command nobody was left to watch: a module handed an
// operation back after the daemon that issued its command was replaced.
//
// A record still at command_issued stays mutable however long ago its grace
// deadline passed. That deadline bounds how long the cluster is held for an
// operation, not how long the evidence stays true: a machine that came back
// on a boot other than the one it was told to leave rebooted whether that
// took two minutes or two days, and refusing to record it would leave a
// record saying a command is outstanding about a reboot this daemon can
// prove finished. Every other terminal status is refused exactly as it is
// for a run, so recovery can settle an outstanding command and nothing else.
func (m *Manager) rejectSettledDuringRecovery(op *Operation) error {
	if op.Status == StatusCommandIssued {
		return nil
	}
	return m.rejectSettled(op)
}

// update applies fn to the stored operation and writes the result out. The
// record is saved under the same lock that changed it, so the file never goes
// backwards relative to what a reader is told in memory.
func (m *Manager) update(id string, fn func(*Operation)) bool {
	return m.updateAt(id, m.deps.Now(), fn)
}

// updateAt is update with an explicit settlement moment rather than one
// applyLocked derives itself. It exists for a caller that has already used
// one "now" to stamp several fields of its own — MarkInterrupted's
// FinishedAt and every step and node it closes with it, for instance — and
// must not let applyLocked's own, separate clock read disagree with what it
// already wrote: against a real clock the skew is a few nanoseconds, but it
// is still the same settlement pretending to be two different moments.
func (m *Manager) updateAt(id string, at time.Time, fn func(*Operation)) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	op, ok := m.ops[id]
	if !ok {
		return false
	}
	if m.persistFailed[id] {
		return false
	}
	return m.applyLockedAt(op, at, fn) == nil
}

// checkedUpdate is the persistence-safe primitive every Runtime mutation
// uses instead of update. Unlike update, validate inspects the operation's
// current state and can reject the mutation before fn ever runs — so a
// rejected mutation leaves no trace, not even the in-memory copy update
// would otherwise have produced. It shares applyLocked with update so a save
// failure settles exactly one way no matter which caller triggered it.
func (m *Manager) checkedUpdate(id string, validate func(*Operation) error, fn func(*Operation)) error {
	return m.checkedUpdateAt(id, m.deps.Now(), validate, fn)
}

// checkedUpdateAt is checkedUpdate with an explicit settlement moment, for
// the same reason updateAt is update with one: a caller whose fn already
// stamped several fields from a single "now" must not have applyLocked read
// the clock again and disagree with it. The clock is read before the lock is
// taken, so a seam behind Deps.Now never runs while the manager is held.
func (m *Manager) checkedUpdateAt(id string, at time.Time,
	validate func(*Operation) error, fn func(*Operation)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	op, ok := m.ops[id]
	if !ok {
		return errOperationNotFound
	}
	// A previous save failure already forced this operation terminal (see
	// applyLocked); reporting the same rejection here keeps callers from
	// having to distinguish "terminal from a normal outcome" from "terminal
	// because its state could no longer be recorded".
	if m.persistFailed[id] {
		return ErrOperationTerminal
	}
	if validate != nil {
		if err := validate(op); err != nil {
			return err
		}
	}
	return m.applyLockedAt(op, at, fn)
}

// complete is the checked mutation behind Runtime.Complete. Outcome.valid()
// has already been checked by the caller, so this only has to refuse an
// operation that has already settled. The text it persists comes from
// Outcome.persistedReason: a reviewed sentence written in this package, or
// the reviewed sentence for the code. A module's own outcome.Error is never
// written to the record or the file it is saved to, only to the log.
func (m *Manager) complete(id string, outcome Outcome, settled settledCheck) error {
	return m.checkedUpdate(id, settled, func(op *Operation) {
		op.Status = outcome.Status
		op.Code = outcome.persistedCode()
		op.Error = outcome.persistedReason()
		op.CommandIssuedUntil = outcome.CommandIssuedUntil
		// Cleared so applyLocked stamps it with this settlement's own time.
		// A command_issued record already carries the moment its command
		// went out, and an operation promoted to a final status once its
		// outcome is known finished when that outcome was established.
		op.FinishedAt = nil
	})
}

// snapshotNode locks just long enough to read one node's current value. It
// applies the same activity check every other checked mutation applies, so
// UpdateNode fails the same way the others do for a settled operation; the
// check that actually matters for correctness is the one replaceNode repeats
// once the caller's mutate callback has run, because the operation or the
// node can change while that callback is running without any manager lock
// held.
//
// What it returns shares nothing with the stored node, times included: the
// caller mutates it with no lock held, and a *time.Time it could write
// through would be a way into the record that no validation and no save ever
// sees.
func (m *Manager) snapshotNode(id, name string, settled settledCheck) (NodeResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	op, ok := m.ops[id]
	if !ok {
		return NodeResult{}, errOperationNotFound
	}
	if m.persistFailed[id] {
		return NodeResult{}, ErrOperationTerminal
	}
	if err := settled(op); err != nil {
		return NodeResult{}, err
	}
	node := findNode(op, name)
	if node == nil {
		return NodeResult{}, ErrNodeNotFound
	}
	return cloneNode(*node), nil
}

// replaceNode is UpdateNode's compare-and-replace commit. before is the
// value snapshotNode returned before the caller's mutate callback ran; if
// the stored node no longer equals it, some other checked mutation committed
// a change in between, and this call refuses to overwrite it — it returns
// ErrConcurrentUpdate instead of silently discarding whatever that other
// writer recorded.
//
// "Equals" is by value, down to the moments its timestamps name, because
// that is what a caller can observe and therefore lose. Comparing the
// pointers instead would make a rewritten but identical timestamp look like
// somebody else's change, and it would make a timestamp mutated in place
// look like no change at all.
func (m *Manager) replaceNode(id, name string, before, after NodeResult, settled settledCheck) error {
	return m.checkedUpdate(id, func(op *Operation) error {
		if err := settled(op); err != nil {
			return err
		}
		current := findNode(op, name)
		if current == nil {
			return ErrNodeNotFound
		}
		if !sameNode(*current, before) {
			return ErrConcurrentUpdate
		}
		return nil
	}, func(op *Operation) {
		if current := findNode(op, name); current != nil {
			// Stored by value for the same reason it was handed out by
			// value: the caller keeps its copy and may write to it again.
			*current = cloneNode(after)
		}
	})
}

// cloneNode copies a node result and the moments behind its timestamps, so
// the copy and the original can be mutated without reaching each other.
func cloneNode(n NodeResult) NodeResult {
	n.StartedAt = cloneTime(n.StartedAt)
	n.FinishedAt = cloneTime(n.FinishedAt)
	return n
}

// sameNode compares two node results the way a reader of the record would:
// same fields, and timestamps naming the same moment.
func sameNode(a, b NodeResult) bool {
	return a.NodeName == b.NodeName &&
		a.Role == b.Role &&
		a.Status == b.Status &&
		a.Code == b.Code &&
		a.Error == b.Error &&
		sameMoment(a.StartedAt, b.StartedAt) &&
		sameMoment(a.FinishedAt, b.FinishedAt)
}

func sameMoment(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

// applyLockedAt mutates op through fn and persists the result, then applies
// the same state_persistence_failed settlement every failed write applies.
// "at" is the moment the change is stamped with — UpdatedAt, and FinishedAt
// too when fn left it nil — and it is the caller's rather than one derived
// here, so a settlement that already stamped a step or a node from its own
// clock reading cannot end up disagreeing with itself. The caller must
// already hold m.mu and must already have confirmed the operation exists and
// has not previously failed to persist.
func (m *Manager) applyLockedAt(op *Operation, at time.Time, fn func(*Operation)) error {
	fn(op)
	op.UpdatedAt = at
	if op.Status.Terminal() {
		if op.FinishedAt == nil {
			finishedAt := at
			op.FinishedAt = &finishedAt
		}
		if m.activeID == op.ID && !m.operationActive(op, op.UpdatedAt) {
			m.activeID = ""
		}
	}
	if err := m.deps.Store.Save(*op); err != nil {
		klog.Errorf("clusterop: persist operation %s: %v", op.ID, err)
		m.persistFailed[op.ID] = true
		m.forceStatePersistenceFailedLocked(op, m.deps.Now())
		return errStatePersistenceFailed
	}
	return nil
}

func (m *Manager) canContinue(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.persistFailed[id]
}

// RequestStop asks a running operation to stop at its next safe point.
//
// It is not an abort. Nothing already dispatched is recalled, because there is
// no safe way to recall it: a node part way through a helm upgrade or a
// container runtime replacement has to be allowed to finish, and an upgrade
// stage cannot be un-run. What stops is the scheduling of further work, at the
// boundary the design already treats as safe — the barrier between two stages,
// where every node is between pieces of work rather than inside one.
//
// It reports whether there was anything to stop. An operation that has already
// settled is left exactly as it is.
func (m *Manager) RequestStop(id string) bool {
	m.mu.Lock()
	op, ok := m.ops[id]
	stoppable := ok && !op.Status.Terminal()
	m.mu.Unlock()
	if !stoppable {
		return false
	}
	// Through the recording path, so it reaches disk: a cancellation the
	// replacement daemon cannot see is a cancellation an upgrade ignores.
	return m.update(id, func(op *Operation) { op.StopRequested = true })
}

// StopRequested reports whether someone has asked this operation to stop.
func (m *Manager) StopRequested(id string) bool {
	op, ok := m.Get(id)
	return ok && op.StopRequested
}

// pruneLocked drops the oldest settled records past the retention limit. An
// operation still in flight is never pruned.
func (m *Manager) pruneLocked() {
	for len(m.order) > m.deps.Retention {
		var removed bool
		for i, id := range m.order {
			op, ok := m.ops[id]
			if !ok {
				m.order = append(m.order[:i], m.order[i+1:]...)
				removed = true
				break
			}
			if !op.Status.Terminal() {
				continue
			}
			if err := m.deps.Store.Delete(id); err != nil {
				klog.Warningf("clusterop: prune operation %s: %v", id, err)
				return
			}
			m.order = append(m.order[:i], m.order[i+1:]...)
			delete(m.ops, id)
			// Only if it is still the one answering. A retired record shares
			// its request id with the retry that replaced it, and is older, so
			// it is pruned first — deleting the mapping unconditionally would
			// unbind the live operation and leave the caller that asks by
			// request id unable to find its own run.
			if current, ok := m.byRequest[op.RequestID]; ok && current == id {
				delete(m.byRequest, op.RequestID)
			}
			removed = true
			break
		}
		if !removed {
			return
		}
	}
}
