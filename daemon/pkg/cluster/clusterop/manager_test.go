package clusterop

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// fakeClock is the only clock the orchestrator sees under test. Sleeping moves
// it forward, so a wait loop bounded by a deadline terminates without the test
// waiting for anything.
type fakeClock struct {
	mu sync.Mutex
	at time.Time
}

func newClock() *fakeClock {
	return &fakeClock{at: time.Unix(1700000000, 0).UTC()}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.at = c.at.Add(time.Millisecond)
	return c.at
}

func (c *fakeClock) Sleep(_ context.Context, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.at = c.at.Add(d)
	return nil
}

// cluster is the fake the orchestrator is driven against. Every seam that
// would touch a machine, a cluster or a wire is here, so a unit test can never
// power down the host running it.
type cluster struct {
	mu sync.Mutex

	nodes        []inventory.Node
	inventoryErr error

	capabilities map[string]map[string]nodestatus.Capability
	inspectErr   map[string]error

	dispatchErr map[string]error

	// obs is what the cluster currently says about each node. A name missing
	// from it is a node that has left the directory. obsSeq scripts changes:
	// one entry is applied per Observe call, and a nil entry means the node
	// disappears.
	obs        map[string]inventory.Observation
	obsSeq     map[string][]*inventory.Observation
	observeErr error

	// hostBootID is the boot the machine running this daemon is on.
	hostBootID string

	// localPowerErr is what this machine's own execution point says when
	// asked whether it can power itself.
	localPowerErr error

	powerSelfErr error

	events    []string
	dispatchN int
	powerN    int
	observeN  int

	onPowerSelf func()
}

func newCluster(nodes ...inventory.Node) *cluster {
	c := &cluster{
		nodes:        nodes,
		capabilities: map[string]map[string]nodestatus.Capability{},
		inspectErr:   map[string]error{},
		dispatchErr:  map[string]error{},
		obs:          map[string]inventory.Observation{},
		obsSeq:       map[string][]*inventory.Observation{},
		hostBootID:   "host-boot-1",
	}
	for _, n := range nodes {
		c.capabilities[n.NodeName] = map[string]nodestatus.Capability{
			nodestatus.CapPowerReboot:   {Supported: true},
			nodestatus.CapPowerShutdown: {Supported: true},
		}
		c.obs[n.NodeName] = inventory.Observation{Ready: n.Ready, BootID: "boot-" + n.NodeName + "-1"}
	}
	return c
}

// The scripts a restarting node can follow. One entry is applied per Observe
// call and the last one persists; a nil entry removes the node.
func (c *cluster) script(name string, seq ...*inventory.Observation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.obsSeq[name] = seq
}

func seen(ready bool, bootID string) *inventory.Observation {
	return &inventory.Observation{Ready: ready, BootID: bootID}
}

func (c *cluster) goesDownAndComesBack(name string) {
	c.script(name, seen(false, "boot-"+name+"-1"), seen(true, "boot-"+name+"-2"))
}

// The node left the directory entirely, which is the usual way a reboot looks.
func (c *cluster) vanishesAndComesBack(name string) {
	c.script(name, nil, seen(true, "boot-"+name+"-2"))
}

// kubelet flapped: the node is Ready again on the boot it started on.
func (c *cluster) comesBackOnTheSameBoot(name string) {
	c.script(name, seen(false, "boot-"+name+"-1"), seen(true, "boot-"+name+"-1"))
}

func (c *cluster) neverComesBack(name string) {
	c.script(name, nil)
}

func (c *cluster) record(e string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, e)
}

func (c *cluster) log() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.events...)
}

func (c *cluster) counts() (dispatch, power int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dispatchN, c.powerN
}

func (c *cluster) observeCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.observeN
}

func (c *cluster) deps(t *testing.T, dir string) Deps {
	t.Helper()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	clock := newClock()
	var ids int
	return Deps{
		Store: store,
		Now:   clock.Now,
		Sleep: clock.Sleep,
		NewID: func() string {
			c.mu.Lock()
			defer c.mu.Unlock()
			ids++
			return "op-" + string(rune('a'+ids-1))
		},
		Inventory: func(context.Context) ([]inventory.Node, error) {
			c.record("inventory")
			c.mu.Lock()
			defer c.mu.Unlock()
			if c.inventoryErr != nil {
				return nil, c.inventoryErr
			}
			return append([]inventory.Node(nil), c.nodes...), nil
		},
		Inspect: func(_ context.Context, n inventory.Node, _ Credentials) (nodestatus.Status, error) {
			c.record("inspect " + n.NodeName)
			c.mu.Lock()
			defer c.mu.Unlock()
			if err := c.inspectErr[n.NodeName]; err != nil {
				return nodestatus.Status{}, err
			}
			return nodestatus.Status{NodeName: n.NodeName, Capabilities: c.capabilities[n.NodeName]}, nil
		},
		Dispatch: func(_ context.Context, nodes []inventory.Node, _ PeerRequest, _ Credentials) []DispatchOutcome {
			out := make([]DispatchOutcome, 0, len(nodes))
			for _, n := range nodes {
				c.record("dispatch " + n.NodeName)
				c.mu.Lock()
				c.dispatchN++
				err := c.dispatchErr[n.NodeName]
				c.mu.Unlock()
				o := DispatchOutcome{NodeName: n.NodeName}
				if err != nil {
					o.Code, o.Err = CodeDispatchFailed, err.Error()
				}
				out = append(out, o)
			}
			return out
		},
		Observe: func(context.Context) (map[string]inventory.Observation, error) {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.observeN++
			if c.observeErr != nil {
				return nil, c.observeErr
			}
			// A restart script describes what happens once the node has been
			// told to reboot, so it does not run over the baseline the
			// precheck reads.
			if c.dispatchN == 0 {
				out := make(map[string]inventory.Observation, len(c.obs))
				for k, v := range c.obs {
					out[k] = v
				}
				return out, nil
			}
			for name, seq := range c.obsSeq {
				if len(seq) == 0 {
					continue
				}
				if seq[0] == nil {
					delete(c.obs, name)
				} else {
					c.obs[name] = *seq[0]
				}
				if len(seq) > 1 {
					c.obsSeq[name] = seq[1:]
				}
			}
			out := make(map[string]inventory.Observation, len(c.obs))
			for k, v := range c.obs {
				out[k] = v
			}
			return out, nil
		},
		LocalPowerSupport: func(ty Type) error {
			c.record("local power check " + string(ty))
			c.mu.Lock()
			defer c.mu.Unlock()
			return c.localPowerErr
		},
		HostBootID: func() (string, error) {
			c.mu.Lock()
			defer c.mu.Unlock()
			return c.hostBootID, nil
		},
		PowerSelf: func(_ context.Context, ty Type) error {
			c.record("power self " + string(ty))
			c.mu.Lock()
			c.powerN++
			err := c.powerSelfErr
			hook := c.onPowerSelf
			c.mu.Unlock()
			if hook != nil {
				hook()
			}
			return err
		},
		Timeouts: Timeouts{Poll: time.Second, Down: time.Minute, Ready: 10 * time.Minute},
	}
}

func master(name, ip string) inventory.Node {
	return inventory.Node{NodeName: name, Role: inventory.RoleMaster, IP: ip, Ready: true, IsSelf: true}
}

func worker(name, ip string) inventory.Node {
	return inventory.Node{NodeName: name, Role: inventory.RoleWorker, IP: ip, Ready: true}
}

func newManager(t *testing.T, c *cluster) (*Manager, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "operations")
	return newManagerAt(t, c, dir), dir
}

// newManagerAt reopens the records in dir, which is what the next daemon does
// after the machine it was running on went down.
func newManagerAt(t *testing.T, c *cluster, dir string) *Manager {
	t.Helper()
	m, err := NewManager(c.deps(t, dir))
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return m
}

func createOp(t *testing.T, m *Manager, ty Type, requestID string) Operation {
	t.Helper()
	op, err := m.Create(context.Background(), CreateRequest{
		Type:      ty,
		RequestID: requestID,
		Owner:     "alice@olares.com",
		Creds:     Credentials{Token: "token", Signature: "jws"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	return op
}

func TestCreatePersistsTheOperationBinding(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)

	op, err := m.Create(context.Background(), CreateRequest{
		Type:      TypeReboot,
		RequestID: "request-1",
		Scope:     ScopeNode,
		Target:    "master-1",
		ClusterID: "cluster-1",
		Owner:     "alice@olares.com",
		Creds:     Credentials{Signature: "jws"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if op.Scope != ScopeNode || op.Target != "master-1" || op.ClusterID != "cluster-1" {
		t.Fatalf("operation binding = %+v", op)
	}
	// Create starts the operation, and its run goroutine writes to the store
	// this test's temporary directory holds. Waiting it out is what keeps
	// that write from racing the framework's own cleanup.
	awaitTerminal(t, m, op.ID)
}

func TestNodeOperationDispatchesOnlyItsTarget(t *testing.T) {
	c := newCluster(
		master("master-1", "10.0.0.1"),
		worker("worker-1", "10.0.0.2"),
		worker("worker-2", "10.0.0.3"),
	)
	m, _ := newManager(t, c)
	op, err := m.Create(context.Background(), CreateRequest{
		Type:      TypeShutdown,
		RequestID: "request-1",
		Scope:     ScopeNode,
		Target:    "worker-1",
		ClusterID: "cluster-1",
		Owner:     "alice@olares.com",
		Creds:     Credentials{Signature: "jws"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := awaitTerminal(t, m, op.ID)
	if len(got.Nodes) != 1 || got.Nodes[0].NodeName != "worker-1" {
		t.Fatalf("node operation recorded %v, want only worker-1", got.Nodes)
	}
	for _, event := range c.log() {
		if event == "dispatch worker-2" || event == "power self shutdown" {
			t.Fatalf("node operation touched another node: %v", c.log())
		}
	}
}

func TestNodeWorkerShutdownReleasesTransientAfterTargetBecomesUnreachable(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.neverComesBack("worker-1")
	m, _ := newManager(t, c)
	op, err := m.Create(context.Background(), CreateRequest{
		Type: TypeShutdown, RequestID: "request-1", Scope: ScopeNode, Target: "worker-1",
		ClusterID: "cluster-1", Owner: "alice@olares.com", Creds: Credentials{Signature: "jws"},
	})
	if err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(time.Second)
	for {
		if _, active := m.ActivePhase(); !active {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("unreachable worker did not release the shutdown transient")
		}
		time.Sleep(time.Millisecond)
	}
	got, _ := m.Get(op.ID)
	if got.Status != StatusCommandIssued {
		t.Fatalf("status = %q, want command_issued", got.Status)
	}
	if c.observeCount() == 0 {
		t.Fatal("worker shutdown was never observed")
	}
}

func TestNodeScopeControlRebootUsesPersistedMasterCommand(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)
	op, err := m.Create(context.Background(), CreateRequest{
		Type:      TypeReboot,
		RequestID: "request-control",
		Scope:     ScopeNode,
		Target:    "master-1",
		ClusterID: "cluster-1",
		Owner:     "alice@olares.com",
		Creds:     Credentials{Signature: "jws"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := awaitTerminal(t, m, op.ID)
	if got.Status != StatusCommandIssued || got.HostBootID == "" || len(got.Nodes) != 1 ||
		got.Nodes[0].NodeName != "master-1" || got.Nodes[0].Status != NodeCommandIssued {
		t.Fatalf("control node operation = %+v", got)
	}
	if dispatch, power := c.counts(); dispatch != 0 || power != 1 {
		t.Fatalf("dispatches=%d powerSelf=%d, want local master reboot only", dispatch, power)
	}
}

func TestNodeScopeControlShutdownIsRejected(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)
	op, err := m.Create(context.Background(), CreateRequest{
		Type: TypeShutdown, RequestID: "request-control", Scope: ScopeNode, Target: "master-1",
		ClusterID: "cluster-1", Owner: "alice@olares.com", Creds: Credentials{Signature: "jws"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := awaitTerminal(t, m, op.ID)
	if got.Status != StatusFailed || got.Code != CodePowerUnsupported {
		t.Fatalf("control node shutdown = %+v", got)
	}
	if _, power := c.counts(); power != 0 {
		t.Fatalf("control node shutdown powered the host %d times", power)
	}
}

func awaitTerminal(t *testing.T, m *Manager, id string) Operation {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		op, ok := m.Get(id)
		if !ok {
			t.Fatalf("operation %s disappeared", id)
		}
		if op.Status.Terminal() {
			return op
		}
		time.Sleep(2 * time.Millisecond)
	}
	op, _ := m.Get(id)
	t.Fatalf("operation %s never settled, last status %q", id, op.Status)
	return Operation{}
}

func TestCreateRejectsWhatItCannotActdOn(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)

	for _, tc := range []struct {
		name string
		req  CreateRequest
	}{
		{"no type", CreateRequest{RequestID: "r1", Owner: "alice"}},
		{"unknown type", CreateRequest{Type: Type("halt"), RequestID: "r1", Owner: "alice"}},
		{"no request id", CreateRequest{Type: TypeReboot, Owner: "alice"}},
		{"no owner", CreateRequest{Type: TypeReboot, RequestID: "r1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := m.Create(context.Background(), tc.req); err == nil {
				t.Fatal("want a refusal")
			}
		})
	}
	if d, p := c.counts(); d != 0 || p != 0 {
		t.Errorf("a refused request touched the cluster: dispatch=%d power=%d", d, p)
	}
}

// A retry of the same confirmation must not power the cluster twice. The key
// is the owner, the operation type and the caller's request id together.
func TestCreateIsIdempotentPerRequestID(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	c.onPowerSelf = func() { time.Sleep(20 * time.Millisecond) }
	m, _ := newManager(t, c)

	first := createOp(t, m, TypeReboot, "client-1")
	second := createOp(t, m, TypeReboot, "client-1")

	if first.ID != second.ID {
		t.Fatalf("retry created a second operation: %s vs %s", first.ID, second.ID)
	}
	awaitTerminal(t, m, first.ID)

	third := createOp(t, m, TypeReboot, "client-1")
	if third.ID != first.ID {
		t.Errorf("a retry after the run finished created a new operation: %s", third.ID)
	}
	if _, p := c.counts(); p != 1 {
		t.Errorf("the host was powered %d times for one request id", p)
	}
}

func TestCreateRejectsAReusedRequestIDWithDifferentIntent(t *testing.T) {
	base := CreateRequest{
		Type: TypeReboot, RequestID: "client-1", Owner: "alice@olares.com",
		Scope: ScopeNode, Target: "master-1", ClusterID: "cluster-1",
	}
	for _, tc := range []struct {
		name   string
		change func(*CreateRequest)
	}{
		{"owner", func(req *CreateRequest) { req.Owner = "bob@olares.com" }},
		{"type", func(req *CreateRequest) { req.Type = TypeShutdown }},
		// A whole-cluster operation names no node, so the scope and the
		// target change together: the module refuses a cluster scope that
		// still names one before the request id is ever looked at.
		{"scope", func(req *CreateRequest) { req.Scope, req.Target = ScopeCluster, "" }},
		{"target", func(req *CreateRequest) { req.Target = "worker-1" }},
		{"cluster", func(req *CreateRequest) { req.ClusterID = "cluster-2" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newCluster(master("master-1", "10.0.0.1"))
			m, _ := newManager(t, c)
			first, err := m.Create(context.Background(), base)
			if err != nil {
				t.Fatal(err)
			}
			awaitTerminal(t, m, first.ID)

			changed := base
			tc.change(&changed)
			_, err = m.Create(context.Background(), changed)
			var conflict *RequestConflictError
			if !errors.As(err, &conflict) {
				t.Fatalf("err = %v, want a requestId conflict", err)
			}
			if conflict.RequestID != "client-1" || conflict.ExistingID != first.ID {
				t.Fatalf("conflict = %+v, want request client-1 and operation %s", conflict, first.ID)
			}
		})
	}
}

func TestCreateRejectsSameRequestIDWithDifferentParams(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)
	first := CreateRequest{
		Type:      TypeReboot,
		RequestID: "client-params",
		Scope:     ScopeNode,
		Target:    "master-1",
		ClusterID: "cluster-1",
		Owner:     "alice@olares.com",
		Params:    json.RawMessage(`{"value":1}`),
	}
	created, err := m.Create(context.Background(), first)
	if err != nil {
		t.Fatal(err)
	}
	// The first request is still being carried out in the background, and it
	// writes to the same directory the test is about to take away.
	awaitTerminal(t, m, created.ID)

	second := first
	second.Params = json.RawMessage(`{"value":2}`)
	_, err = m.Create(context.Background(), second)
	var conflict *RequestConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("Create() error = %v, want RequestConflictError", err)
	}
}

func TestCreatePersistsParamsDigestWithoutRawParams(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, dir := newManager(t, c)
	req := CreateRequest{
		Type:      TypeReboot,
		RequestID: "client-digest",
		Scope:     ScopeNode,
		Target:    "master-1",
		ClusterID: "cluster-1",
		Owner:     "alice@olares.com",
		Params:    json.RawMessage(`{"password":"secret","value":1}`),
	}
	op, err := m.Create(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if op.ParamsDigest == "" {
		t.Fatal("Create() returned an empty params digest")
	}
	if _, ok := reflect.TypeOf(op).FieldByName("Params"); ok {
		t.Fatal("Operation must not carry raw params")
	}

	data, err := os.ReadFile(filepath.Join(dir, op.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"paramsDigest"`) {
		t.Fatalf("persisted JSON = %s, want paramsDigest", data)
	}
	if strings.Contains(string(data), `"params":`) {
		t.Fatalf("persisted JSON leaked params field: %s", data)
	}
	if strings.Contains(string(data), "secret") || strings.Contains(string(data), `"value":1`) {
		t.Fatalf("persisted JSON leaked raw params: %s", data)
	}
}

func TestGetByRequestSurvivesRestartAndReturnsACopy(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, dir := newManager(t, c)
	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client/request 1").ID)

	restarted := newManagerAt(t, c, dir)
	got, ok := restarted.GetByRequest("client/request 1")
	if !ok || got.ID != op.ID {
		t.Fatalf("operation = %+v ok = %v, want %s", got, ok, op.ID)
	}
	got.RequestID = "tampered"
	again, _ := restarted.GetByRequest("client/request 1")
	if again.RequestID != "client/request 1" {
		t.Fatal("a caller wrote back into the request index")
	}
}

func TestOperationJSONBackwardCompatibleWithoutModuleState(t *testing.T) {
	data := []byte(`{"id":"op-1","type":"reboot","requestId":"request-1","owner":"alice@olares.com","status":"pending","createdAt":"2024-01-01T00:00:00Z","updatedAt":"2024-01-01T00:00:00Z","steps":[],"nodes":[]}`)
	var op Operation
	if err := json.Unmarshal(data, &op); err != nil {
		t.Fatal(err)
	}
	if len(op.ModuleState) != 0 {
		t.Fatalf("moduleState = %s, want empty", op.ModuleState)
	}

	op.ModuleState = json.RawMessage(`{"resume":"ok"}`)
	encoded, err := json.Marshal(op)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"moduleState":{"resume":"ok"}`) {
		t.Fatalf("marshaled JSON = %s, want moduleState", encoded)
	}
}

func TestOperationCloneCopiesModuleState(t *testing.T) {
	op := Operation{ModuleState: json.RawMessage(`{"resume":"ok"}`)}
	cloned := op.Clone()
	if string(cloned.ModuleState) != `{"resume":"ok"}` {
		t.Fatalf("Clone() moduleState = %s", cloned.ModuleState)
	}
	cloned.ModuleState[0] = '['
	if string(op.ModuleState) != `{"resume":"ok"}` {
		t.Fatalf("Clone() shared moduleState backing bytes: %s", op.ModuleState)
	}
}

func TestCreateRefusesASecondPowerOperationWhileOneIsInFlight(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	release := make(chan struct{})
	c.onPowerSelf = func() { <-release }
	m, _ := newManager(t, c)

	first := createOp(t, m, TypeReboot, "client-1")

	_, err := m.Create(context.Background(), CreateRequest{
		Type: TypeShutdown, RequestID: "client-2", Owner: "alice@olares.com",
	})
	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("err = %v, want a conflict", err)
	}
	if conflict.ActiveID != first.ID {
		t.Errorf("conflict names %q, want the operation in flight %q", conflict.ActiveID, first.ID)
	}
	if conflict.ActiveType != TypeReboot {
		t.Errorf("conflict type = %q, want reboot", conflict.ActiveType)
	}

	close(release)
	awaitTerminal(t, m, first.ID)
}

// Same id, different operation: a shutdown must never be answered with a
// reboot that happens to carry the caller's request id.
func TestCreateDoesNotJoinADifferentOperationTypeOnTheSameRequestID(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	release := make(chan struct{})
	c.onPowerSelf = func() { <-release }
	m, _ := newManager(t, c)

	first := createOp(t, m, TypeReboot, "client-1")

	_, err := m.Create(context.Background(), CreateRequest{
		Type: TypeShutdown, RequestID: "client-1", Owner: "alice@olares.com",
	})
	var conflict *RequestConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("err = %v, want a requestId conflict rather than the reboot back", err)
	}

	close(release)
	awaitTerminal(t, m, first.ID)
}

// A cluster reboot outlives the HTTP request that asked for it, and the
// server recycles that request's context the moment the response is written.
// Running on it would abandon the operation part way through the cluster.
func TestCreateDoesNotRunOnTheCallersContext(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	m, _ := newManager(t, c)

	ctx, cancel := context.WithCancel(context.Background())
	op, err := m.Create(ctx, CreateRequest{
		Type: TypeShutdown, RequestID: "client-1", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	cancel()

	settled := awaitTerminal(t, m, op.ID)
	if settled.Status != StatusCommandIssued {
		t.Errorf("status = %q, want the operation carried out anyway: %+v", settled.Status, settled)
	}
	if _, power := c.counts(); power != 1 {
		t.Errorf("the control node was powered %d times", power)
	}
}

func TestCommandIssuedTransientBlocksAnotherOwner(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)

	first := createOp(t, m, TypeReboot, "client-1")
	awaitTerminal(t, m, first.ID)

	_, err := m.Create(context.Background(), CreateRequest{
		Type: TypeReboot, RequestID: "client-1", Owner: "bob@olares.com",
	})
	if err == nil {
		t.Fatal("another owner bypassed the command_issued transient")
	}
}

func TestCreateAllowsANewOperationAfterTheTransientDeadline(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)

	first := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)
	if first.Status != StatusCommandIssued {
		t.Fatalf("status = %q, want command_issued", first.Status)
	}

	if err := m.deps.Sleep(context.Background(), 2*m.deps.Timeouts.Ready); err != nil {
		t.Fatal(err)
	}
	next, err := m.Create(context.Background(), CreateRequest{
		Type: TypeShutdown, RequestID: "client-2", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("an expired transient still blocks the next one: %v", err)
	}
	awaitTerminal(t, m, next.ID)
}

func TestCreatedOperationIsQueryableAfterTheProcessRestarts(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, dir := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)

	restarted, err := NewManager(c.deps(t, dir))
	if err != nil {
		t.Fatalf("NewManager after restart: %v", err)
	}
	got, ok := restarted.Get(op.ID)
	if !ok {
		t.Fatalf("operation %s not queryable after a restart", op.ID)
	}
	if got.Status != op.Status || got.Type != TypeShutdown {
		t.Errorf("operation changed across the restart: %+v", got)
	}
}

// olaresd going down mid-operation is exactly what a cluster reboot does to
// it. Nothing observed how that operation ended, so it is reported as failed —
// and, just as importantly, it stops holding the single-operation lock.
func TestAnOperationInterruptedByARestartIsReportedFailed(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	at := time.Unix(1700000000, 0).UTC()
	stranded := Operation{
		ID: "op-stranded", Type: TypeReboot, RequestID: "client-1", Owner: "alice@olares.com",
		Status: StatusRunning, CreatedAt: at, UpdatedAt: at, StartedAt: &at,
		Steps: []Step{{Name: StepWorkerRestart, Status: StepRunning, StartedAt: &at}},
	}
	if err := store.Save(stranded); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m, err := NewManager(c.deps(t, dir))
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	got, ok := m.Get("op-stranded")
	if !ok {
		t.Fatal("the stranded operation is gone")
	}
	if got.Status != StatusFailed || got.Code != CodeDaemonRestarted {
		t.Errorf("status = %q code = %q, want failed/%s", got.Status, got.Code, CodeDaemonRestarted)
	}
	if got.FinishedAt == nil {
		t.Error("a settled operation must say when it settled")
	}
	if len(got.Steps) == 0 || got.Steps[0].Status != StepFailed {
		t.Errorf("the step that was in flight is still running: %+v", got.Steps)
	}
	next, err := m.Create(context.Background(), CreateRequest{
		Type: TypeShutdown, RequestID: "client-2", Owner: "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("a stranded operation still blocks the cluster: %v", err)
	}
	awaitTerminal(t, m, next.ID)

	reloaded, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	persisted, _, err := reloaded.Load("op-stranded")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if persisted.Status != StatusFailed {
		t.Errorf("the correction was not written back: %q", persisted.Status)
	}
}

// Nothing in this package may reach a machine unless it was given the means to.
// A test that forgets to inject the power executor must fail to build a
// manager, not reboot the host it runs on.
func TestNewManagerRefusesMissingSeams(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	c := newCluster(master("master-1", "10.0.0.1"))
	complete := c.deps(t, dir)

	for name, breakIt := range map[string]func(*Deps){
		"store":      func(d *Deps) { d.Store = nil },
		"inventory":  func(d *Deps) { d.Inventory = nil },
		"inspect":    func(d *Deps) { d.Inspect = nil },
		"dispatch":   func(d *Deps) { d.Dispatch = nil },
		"observe":    func(d *Deps) { d.Observe = nil },
		"hostBootID": func(d *Deps) { d.HostBootID = nil },
		"powerSelf":  func(d *Deps) { d.PowerSelf = nil },
	} {
		t.Run(name, func(t *testing.T) {
			broken := complete
			breakIt(&broken)
			if _, err := NewManager(broken); err == nil {
				t.Fatalf("NewManager accepted a manager with no %s", name)
			}
		})
	}
}

// A store field that holds a nil *Store compares unequal to nil as an
// OperationStore interface — the interface has a type, even though the
// pointer inside it is nil — so the plain nil check above lets it through.
// The very first call through it would then panic instead of refusing to
// build the manager.
func TestNewManagerRefusesTypedNilStore(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")
	deps := c.deps(t, dir)

	var nilStore *Store
	deps.Store = nilStore

	if _, err := NewManager(deps); err == nil {
		t.Fatal("NewManager(deps with a typed-nil Store) = nil error, want a refusal")
	}
}

func TestManagerPrunesOldRecords(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")
	deps := c.deps(t, dir)
	deps.Retention = 2
	m, err := NewManager(deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	var ids []string
	for _, req := range []string{"r1", "r2", "r3", "r4"} {
		op := createOp(t, m, TypeReboot, req)
		awaitTerminal(t, m, op.ID)
		ids = append(ids, op.ID)
		if err := m.deps.Sleep(context.Background(), 2*m.deps.Timeouts.Ready); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) > 2 {
		t.Errorf("want at most 2 records retained, got %d", len(entries))
	}
	if _, ok := m.Get(ids[len(ids)-1]); !ok {
		t.Error("the newest operation was pruned")
	}
	if _, ok := m.GetByRequest("r1"); ok {
		t.Error("a pruned operation remains in the request index")
	}
}

func TestManagerKeepsIndexesWhenStoreDeleteFails(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)
	op := &Operation{
		ID: "invalid/id", Type: TypeReboot, RequestID: "request-retained",
		Owner: "alice@olares.com", Status: StatusSucceeded,
	}
	m.ops[op.ID] = op
	m.order = []string{op.ID}
	m.byRequest[op.RequestID] = op.ID
	m.deps.Retention = 0

	m.mu.Lock()
	m.pruneLocked()
	m.mu.Unlock()

	if _, ok := m.Get(op.ID); !ok {
		t.Fatal("operation was removed from memory after Store.Delete failed")
	}
	if got, ok := m.GetByRequest(op.RequestID); !ok || got.ID != op.ID {
		t.Fatal("requestId was released after Store.Delete failed")
	}
	if len(m.order) != 1 || m.order[0] != op.ID {
		t.Fatalf("order = %v, want the failed record retained", m.order)
	}
	same, err := m.Create(context.Background(), CreateRequest{
		Type: op.Type, RequestID: op.RequestID, Owner: op.Owner,
	})
	if err != nil || same.ID != op.ID {
		t.Fatalf("same request after failed prune = %+v, %v", same, err)
	}
	_, err = m.Create(context.Background(), CreateRequest{
		Type: TypeShutdown, RequestID: op.RequestID, Owner: op.Owner,
	})
	var conflict *RequestConflictError
	if !errors.As(err, &conflict) || conflict.ExistingID != op.ID {
		t.Fatalf("changed request after failed prune err = %v, want conflict with %s", err, op.ID)
	}
}

func TestRestoredRequestIDKeepsCreateIdempotentAndUnambiguous(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, dir := newManager(t, c)
	req := CreateRequest{
		Type: TypeReboot, RequestID: "request-restored", Scope: ScopeCluster,
		ClusterID: "cluster-1", Owner: "alice@olares.com",
	}
	first, err := m.Create(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	awaitTerminal(t, m, first.ID)

	restarted := newManagerAt(t, c, dir)
	same, err := restarted.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("same intent after restart: %v", err)
	}
	if same.ID != first.ID {
		t.Fatalf("same intent created %s, want %s", same.ID, first.ID)
	}

	changed := req
	changed.Scope, changed.Target = ScopeNode, "worker-1"
	_, err = restarted.Create(context.Background(), changed)
	var conflict *RequestConflictError
	if !errors.As(err, &conflict) || conflict.ExistingID != first.ID {
		t.Fatalf("different intent err = %v, want conflict with %s", err, first.ID)
	}
}

func TestRestoredEmptyParamsDigestMatchesOnlyEmptyParams(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")
	deps := c.deps(t, dir)
	now := deps.Now()
	if err := deps.Store.Save(Operation{
		ID:        "op-legacy",
		Type:      TypeReboot,
		RequestID: "request-legacy",
		Scope:     ScopeCluster,
		ClusterID: "cluster-1",
		Owner:     "alice@olares.com",
		Status:    StatusSucceeded,
		CreatedAt: now,
		UpdatedAt: now,
		Steps:     []Step{},
		Nodes:     []NodeResult{},
	}); err != nil {
		t.Fatal(err)
	}

	m, err := NewManager(deps)
	if err != nil {
		t.Fatal(err)
	}

	same, err := m.Create(context.Background(), CreateRequest{
		Type:      TypeReboot,
		RequestID: "request-legacy",
		Scope:     ScopeCluster,
		ClusterID: "cluster-1",
		Owner:     "alice@olares.com",
	})
	if err != nil {
		t.Fatalf("empty params retry after restart: %v", err)
	}
	if same.ID != "op-legacy" {
		t.Fatalf("same request returned %s, want op-legacy", same.ID)
	}

	_, err = m.Create(context.Background(), CreateRequest{
		Type:      TypeReboot,
		RequestID: "request-legacy",
		Scope:     ScopeCluster,
		ClusterID: "cluster-1",
		Owner:     "alice@olares.com",
		Params:    json.RawMessage(`{"value":1}`),
	})
	var conflict *RequestConflictError
	if !errors.As(err, &conflict) || conflict.ExistingID != "op-legacy" {
		t.Fatalf("non-empty params err = %v, want conflict with op-legacy", err)
	}
}

func TestActivePhaseFollowsTheOperationInFlight(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	release := make(chan struct{})
	c.onPowerSelf = func() { <-release }
	m, _ := newManager(t, c)

	if _, ok := m.ActivePhase(); ok {
		t.Fatal("an idle cluster must not impose a phase")
	}

	op := createOp(t, m, TypeShutdown, "client-1")
	phase, ok := m.ActivePhase()
	if !ok || phase != nodestatus.PhaseShuttingDown {
		t.Errorf("phase = %q ok = %v, want shutting_down", phase, ok)
	}

	close(release)
	awaitTerminal(t, m, op.ID)
}

func TestPersistedCommandIssuedOperationBlocksUntilItsDeadline(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	dir := filepath.Join(t.TempDir(), "operations")
	deps := c.deps(t, dir)
	now := deps.Now()
	if err := deps.Store.Save(Operation{
		ID:                 "op-issued",
		Type:               TypeShutdown,
		RequestID:          "request-issued",
		Scope:              ScopeCluster,
		Owner:              "alice@olares.com",
		Status:             StatusCommandIssued,
		CreatedAt:          now,
		UpdatedAt:          now,
		CommandIssuedUntil: now.Add(time.Minute),
	}); err != nil {
		t.Fatal(err)
	}
	m, err := NewManager(deps)
	if err != nil {
		t.Fatal(err)
	}
	if phase, ok := m.ActivePhase(); !ok || phase != nodestatus.PhaseShuttingDown {
		t.Fatalf("phase = %q ok = %v, want persisted shutting_down", phase, ok)
	}
	if _, err := m.Create(context.Background(), CreateRequest{
		Type: TypeReboot, RequestID: "request-next", Scope: ScopeCluster, Owner: "alice@olares.com",
	}); err == nil {
		t.Fatal("command_issued operation did not block another power operation")
	}
}

// An operation that failed changes nothing about the machine, so it imposes
// nothing on the phase. An operation that reached command_issued is the other
// case and is covered in phase_test.go: there the machine really is on its way
// down.
func TestAFailedOperationImposesNoPhase(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	c.localPowerErr = errors.New("this node cannot power itself")
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeShutdown, "client-1").ID)
	if op.Status != StatusFailed {
		t.Fatalf("status = %q, want failed", op.Status)
	}

	if phase, ok := m.ActivePhase(); ok {
		t.Errorf("phase = %q, want none after an operation that did nothing", phase)
	}
}

func TestGetIsACopy(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"))
	m, _ := newManager(t, c)
	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	got, _ := m.Get(op.ID)
	if len(got.Nodes) == 0 {
		t.Fatal("no node results recorded")
	}
	got.Nodes[0].NodeName = "tampered"

	again, _ := m.Get(op.ID)
	if again.Nodes[0].NodeName == "tampered" {
		t.Error("a caller wrote back into the stored operation")
	}
}

func TestPersistenceFailureStopsBeforePoweringTheControlNode(t *testing.T) {
	dir := t.TempDir()
	control := inventory.Node{NodeName: "master", Role: inventory.RoleMaster, Ready: true, IsSelf: true}
	c := newCluster(control)
	deps := c.deps(t, dir)
	m, err := NewManager(deps)
	if err != nil {
		t.Fatal(err)
	}
	op := &Operation{
		ID:        "op-persist",
		Type:      TypeReboot,
		RequestID: "request-persist",
		Owner:     "alice@olares.com",
		Status:    StatusRunning,
		Nodes:     []NodeResult{{NodeName: control.NodeName, Role: control.Role, Status: NodePending}},
	}
	if err := deps.Store.Save(*op); err != nil {
		t.Fatal(err)
	}
	m.ops[op.ID] = op
	m.activeID = op.ID
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}

	m.powerMaster(context.Background(), op.ID, plan{master: control}, rebootSpec, false)

	_, powered := c.counts()
	if powered != 0 {
		t.Fatalf("control node powered %d times after persistence failed", powered)
	}
	got, ok := m.Get(op.ID)
	if !ok || got.Status != StatusFailed || got.Code != CodeStatePersistenceFailed {
		t.Fatalf("operation = %+v, want failed/%s", got, CodeStatePersistenceFailed)
	}
	if _, err := m.Create(context.Background(), CreateRequest{
		Type: TypeShutdown, RequestID: "request-after-failure", Owner: "alice@olares.com",
	}); err == nil {
		t.Fatal("manager accepted another operation after persistence failed")
	}
}
