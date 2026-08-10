package clusterop

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
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
const (
	defaultRetention   = 20
	observationTimeout = 10 * time.Second
)

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

// Timeouts bound the waits a reboot performs.
type Timeouts struct {
	// Poll is how often a restarting node is probed.
	Poll time.Duration
	// Down is how long a node may take to stop answering after being told to
	// reboot. A node that never stops answering never rebooted.
	Down time.Duration
	// Ready is how long a node may take to come back and be Ready again.
	Ready time.Duration
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
	return t
}

// Deps are every side effect the orchestrator can have. All of them are
// injected: a unit test drives a whole cluster reboot without a cluster, a
// network, a clock or a machine that can be powered off.
type Deps struct {
	Store *Store

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
}

func (d Deps) validate() error {
	var missing []string
	if d.Store == nil {
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
	Creds     Credentials
}

// Errors a caller can act on.
var (
	ErrRequestIDRequired = errors.New("requestId is required")
	ErrOwnerRequired     = errors.New("owner is required")
)

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

	mu            sync.Mutex
	ops           map[string]*Operation
	order         []string
	byRequest     map[string]string
	activeID      string
	persistFailed map[string]bool
}

// NewManager loads the recorded operations and returns a manager over them.
func NewManager(deps Deps) (*Manager, error) {
	if err := deps.validate(); err != nil {
		return nil, err
	}
	deps = deps.withDefaults()

	m := &Manager{
		deps:          deps,
		ops:           map[string]*Operation{},
		byRequest:     map[string]string{},
		persistFailed: map[string]bool{},
	}

	stored, err := deps.Store.List()
	if err != nil {
		return nil, fmt.Errorf("load cluster operations: %w", err)
	}
	boot, err := deps.HostBootID()
	if err != nil {
		// Without a boot id nothing is promoted. Reporting a reboot that may
		// not have happened is worse than leaving it at command_issued.
		klog.Warningf("clusterop: read this machine's boot id: %v", err)
	}
	var pendingReboots []string
	for i := range stored {
		op := stored[i]
		switch {
		case !op.Status.Terminal():
			m.markInterrupted(&op)
			if err := deps.Store.Save(op); err != nil {
				klog.Warningf("clusterop: record interrupted operation %s: %v", op.ID, err)
			}
		case rebootChangedBoot(&op, boot):
			pendingReboots = append(pendingReboots, op.ID)
		}
		if id, ok := m.byRequest[op.RequestID]; ok && id != op.ID {
			return nil, fmt.Errorf("load cluster operations: %w", &RequestConflictError{
				RequestID: op.RequestID, ExistingID: id,
			})
		}
		m.ops[op.ID] = &op
		m.order = append(m.order, op.ID)
		m.byRequest[op.RequestID] = op.ID
		if m.operationActive(&op, deps.Now()) {
			m.activeID = op.ID
		}
	}
	for _, id := range pendingReboots {
		go m.confirmRebootWhenReady(id, boot)
	}
	return m, nil
}

func rebootChangedBoot(op *Operation, boot string) bool {
	return op.Type == TypeReboot && op.Status == StatusCommandIssued &&
		boot != "" && op.HostBootID != "" && boot != op.HostBootID
}

// markInterrupted settles an operation that was still moving when olaresd
// stopped. Nothing watched how it ended, so it is not reported as anything but
// failed — and settling it is what stops it from holding the cluster's
// single-operation lock forever.
func (m *Manager) markInterrupted(op *Operation) {
	at := m.deps.Now()
	op.Status = StatusFailed
	op.Code = CodeDaemonRestarted
	op.Error = "olaresd restarted while this operation was in progress"
	op.UpdatedAt = at
	op.FinishedAt = &at
	for i := range op.Steps {
		if op.Steps[i].Status == StepRunning || op.Steps[i].Status == StepPending {
			op.Steps[i].Status = StepFailed
			op.Steps[i].Code = CodeDaemonRestarted
			op.Steps[i].FinishedAt = &at
		}
	}
	for i := range op.Nodes {
		if op.Nodes[i].Status == NodePending {
			op.Nodes[i].Status = NodeFailed
			op.Nodes[i].Code = CodeDaemonRestarted
			op.Nodes[i].FinishedAt = &at
		}
	}
}

// confirmReboot promotes a control-node reboot that this daemon can prove
// happened. The proof is the machine being on a different boot than the one
// recorded before the command was issued; olaresd restarting on the same boot
// proves nothing, and a shutdown is never promoted at all — the machine being
// on again means somebody turned it back on, not that the operation succeeded.
func (m *Manager) confirmReboot(op *Operation, boot string,
	observed map[string]inventory.Observation) bool {
	if !rebootChangedBoot(op, boot) {
		return false
	}
	if !controlNodeReady(op, boot, observed) {
		return false
	}

	at := m.deps.Now()
	op.Status = StatusSucceeded
	op.UpdatedAt = at
	op.FinishedAt = &at
	for i := range op.Steps {
		if op.Steps[i].Name == StepMasterCommand && op.Steps[i].Status == StepCommandIssued {
			op.Steps[i].Status = StepSucceeded
			op.Steps[i].FinishedAt = &at
		}
	}
	for i := range op.Nodes {
		if op.Nodes[i].Status == NodeCommandIssued && op.Nodes[i].Role == inventory.RoleMaster {
			op.Nodes[i].Status = NodeRestarted
			op.Nodes[i].FinishedAt = &at
		}
	}
	return true
}

func controlNodeReady(op *Operation, boot string,
	observed map[string]inventory.Observation) bool {
	for _, node := range op.Nodes {
		if node.Role != inventory.RoleMaster {
			continue
		}
		obs, ok := observed[node.NodeName]
		return ok && obs.Ready && obs.BootID == boot
	}
	return false
}

func (m *Manager) confirmRebootWhenReady(id, boot string) {
	deadline := m.deps.Now().Add(m.deps.Timeouts.Ready)
	for m.deps.Now().Before(deadline) {
		observeCtx, cancel := context.WithTimeout(m.deps.Base, observationTimeout)
		observed, err := m.deps.Observe(observeCtx)
		cancel()
		if err == nil {
			op, ok := m.Get(id)
			if !ok || op.Status != StatusCommandIssued {
				return
			}
			if controlNodeReady(&op, boot, observed) {
				m.update(id, func(stored *Operation) {
					m.confirmReboot(stored, boot, observed)
				})
				return
			}
		}
		if err := m.deps.Sleep(m.deps.Base, m.deps.Timeouts.Poll); err != nil {
			return
		}
	}
}

func sameIntent(op *Operation, req CreateRequest, opType Type) bool {
	return op.Owner == req.Owner &&
		op.Type == opType &&
		op.Scope == req.Scope &&
		op.Target == req.Target &&
		op.ClusterID == req.ClusterID
}

// Create starts a cluster power operation, or returns the one this request
// already started. It returns as soon as the operation is recorded: the power
// commands themselves are issued by the run it launches.
// The caller's context is deliberately unused: see Deps.Base.
func (m *Manager) Create(_ context.Context, req CreateRequest) (Operation, error) {
	opType, err := ParseType(string(req.Type))
	if err != nil {
		return Operation{}, err
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return Operation{}, ErrRequestIDRequired
	}
	if strings.TrimSpace(req.Owner) == "" {
		return Operation{}, ErrOwnerRequired
	}

	m.mu.Lock()
	if id, ok := m.byRequest[req.RequestID]; ok {
		existing := m.ops[id]
		if !sameIntent(existing, req, opType) {
			m.mu.Unlock()
			return Operation{}, &RequestConflictError{RequestID: req.RequestID, ExistingID: id}
		}
		cloned := existing.Clone()
		m.mu.Unlock()
		return cloned, nil
	}
	if active := m.activeOperationLocked(); active != nil {
		err := &ConflictError{ActiveID: active.ID, ActiveType: active.Type}
		m.mu.Unlock()
		return Operation{}, err
	}

	at := m.deps.Now()
	op := &Operation{
		ID:        m.deps.NewID(),
		Type:      opType,
		RequestID: req.RequestID,
		Scope:     req.Scope,
		Target:    req.Target,
		ClusterID: req.ClusterID,
		Owner:     req.Owner,
		Status:    StatusPending,
		CreatedAt: at,
		UpdatedAt: at,
		Steps:     []Step{},
		Nodes:     []NodeResult{},
	}
	m.ops[op.ID] = op
	m.order = append(m.order, op.ID)
	m.byRequest[req.RequestID] = op.ID
	m.activeID = op.ID
	if err := m.deps.Store.Save(*op); err != nil {
		delete(m.ops, op.ID)
		m.order = m.order[:len(m.order)-1]
		delete(m.byRequest, req.RequestID)
		m.activeID = ""
		m.mu.Unlock()
		return Operation{}, fmt.Errorf("record cluster operation: %w", err)
	}
	m.pruneLocked()
	created := op.Clone()
	m.mu.Unlock()

	// Detached from the request: the caller polls by id, and a browser that
	// navigates away must not cancel a reboot half way through.
	go m.run(m.deps.Base, created.ID, created.Type, req.Creds)

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

// ActivePhase is the cluster phase implied by the operation in flight, or by
// one whose command has been issued and whose machine has not gone down yet. ok
// is false when there is neither, which leaves the caller's own phase alone.
func (m *Manager) ActivePhase() (nodestatus.Phase, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	active := m.activeOperationLocked()
	if active == nil {
		return "", false
	}
	if active.Status == StatusCommandIssued {
		return phaseForType(active.Type)
	}
	return PhaseFor(active)
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

// update applies fn to the stored operation and writes the result out. The
// record is saved under the same lock that changed it, so the file never goes
// backwards relative to what a reader is told in memory.
func (m *Manager) update(id string, fn func(*Operation)) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	op, ok := m.ops[id]
	if !ok {
		return false
	}
	if m.persistFailed[id] {
		return false
	}
	fn(op)
	op.UpdatedAt = m.deps.Now()
	if op.Status.Terminal() {
		if op.FinishedAt == nil {
			at := op.UpdatedAt
			op.FinishedAt = &at
		}
		if m.activeID == id && !m.operationActive(op, op.UpdatedAt) {
			m.activeID = ""
		}
	}
	if err := m.deps.Store.Save(*op); err != nil {
		klog.Errorf("clusterop: persist operation %s: %v", id, err)
		m.persistFailed[id] = true
		op.Status = StatusFailed
		op.Code = CodeStatePersistenceFailed
		op.Error = "the operation stopped because its state could not be recorded"
		at := m.deps.Now()
		op.UpdatedAt = at
		op.FinishedAt = &at
		m.activeID = id
		return false
	}
	return true
}

func (m *Manager) canContinue(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.persistFailed[id]
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
			delete(m.byRequest, op.RequestID)
			removed = true
			break
		}
		if !removed {
			return
		}
	}
}
