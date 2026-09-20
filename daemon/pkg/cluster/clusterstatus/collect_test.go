package clusterstatus

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

const testToken = "test-access-token"

func directory(nodes ...inventory.Node) func(context.Context) ([]inventory.Node, error) {
	return func(context.Context) ([]inventory.Node, error) { return nodes, nil }
}

func masterEntry() inventory.Node {
	return inventory.Node{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true, IsSelf: true}
}

func workerEntry(name, ip string) inventory.Node {
	return inventory.Node{NodeName: name, Role: inventory.RoleWorker, IP: ip, Ready: true}
}

func localRunning() func(context.Context) nodestatus.Status {
	return func(context.Context) nodestatus.Status {
		return nodestatus.Status{
			NodeName:     "master-1",
			Role:         inventory.RoleMaster,
			Health:       nodestatus.HealthHealthy,
			Connectivity: nodestatus.ConnectivityOnline,
			Phase:        nodestatus.PhaseRunning,
		}
	}
}

func remoteRunning(context.Context, inventory.Node, string) (nodestatus.Status, error) {
	return nodestatus.Status{
		Health:       nodestatus.HealthHealthy,
		Connectivity: nodestatus.ConnectivityOnline,
		Phase:        nodestatus.PhaseRunning,
	}, nil
}

func collector(t *testing.T) *Collector {
	t.Helper()
	return &Collector{
		Nodes:    directory(masterEntry()),
		Identify: func(context.Context) Identity { return identity() },
		Local:    localRunning(),
		Remote: func(context.Context, inventory.Node, string) (nodestatus.Status, error) {
			t.Error("a worker was dialled by a test that declared none")
			return nodestatus.Status{}, nil
		},
		Now: func() time.Time { return buildAt },
	}
}

func TestCollectDescribesASingleNodeCluster(t *testing.T) {
	got, err := collector(t).Collect(context.Background(), testToken)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}

	if got.Health != nodestatus.HealthHealthy || got.Nodes.Total != 1 {
		t.Errorf("summary = %+v", got)
	}
	if got.ClusterID != "ns-uid" {
		t.Errorf("clusterId = %q", got.ClusterID)
	}
}

// The control node is the machine running this code. Asking it over HTTP would
// make the cluster's own summary depend on its network being up.
func TestTheControlNodeIsReadLocally(t *testing.T) {
	c := collector(t)
	var localCalls atomic.Int32
	c.Local = func(ctx context.Context) nodestatus.Status {
		localCalls.Add(1)
		return localRunning()(ctx)
	}

	if _, err := c.Collect(context.Background(), testToken); err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if localCalls.Load() != 1 {
		t.Errorf("the local snapshot was read %d times, want once", localCalls.Load())
	}
}

func TestTheAccessTokenReachesTheWorkers(t *testing.T) {
	c := collector(t)
	c.Nodes = directory(masterEntry(), workerEntry("worker-1", "10.0.0.2"))
	var seen atomic.Value
	c.Remote = func(_ context.Context, _ inventory.Node, token string) (nodestatus.Status, error) {
		seen.Store(token)
		return remoteRunning(context.Background(), inventory.Node{}, token)
	}

	if _, err := c.Collect(context.Background(), testToken); err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if got, _ := seen.Load().(string); got != testToken {
		t.Errorf("token forwarded to the worker = %q, want the caller's", got)
	}
}

// Nodes are asked at the same time. Adding them up one timeout at a time is
// how a two-node cluster's overview page becomes a minute long.
func TestWorkersAreQueriedConcurrently(t *testing.T) {
	const perNode = 200 * time.Millisecond
	c := collector(t)
	c.Nodes = directory(
		masterEntry(),
		workerEntry("worker-1", "10.0.0.2"),
		workerEntry("worker-2", "10.0.0.3"),
		workerEntry("worker-3", "10.0.0.4"),
		workerEntry("worker-4", "10.0.0.5"),
	)
	c.Remote = func(ctx context.Context, _ inventory.Node, _ string) (nodestatus.Status, error) {
		select {
		case <-time.After(perNode):
		case <-ctx.Done():
			return nodestatus.Status{}, ctx.Err()
		}
		return remoteRunning(ctx, inventory.Node{}, "")
	}

	start := time.Now()
	got, err := c.Collect(context.Background(), testToken)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	elapsed := time.Since(start)

	if got.Nodes.Total != 5 || got.Nodes.Healthy != 5 {
		t.Fatalf("summary = %+v", got.Nodes)
	}
	if elapsed >= 4*perNode {
		t.Errorf("four workers took %v, which is one after another rather than at once", elapsed)
	}
}

// One node that never answers must not decide how long the page takes, and
// must not stop the others from being reported.
func TestOneHangingNodeIsBoundedAndDoesNotHideTheRest(t *testing.T) {
	c := collector(t)
	c.Timeout = 150 * time.Millisecond
	c.Nodes = directory(masterEntry(), workerEntry("hangs", "10.0.0.2"), workerEntry("answers", "10.0.0.3"))
	c.Remote = func(ctx context.Context, n inventory.Node, _ string) (nodestatus.Status, error) {
		if n.NodeName != "hangs" {
			return remoteRunning(ctx, n, "")
		}
		<-ctx.Done()
		return nodestatus.Status{}, ctx.Err()
	}

	start := time.Now()
	got, err := c.Collect(context.Background(), testToken)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Errorf("a hanging node held the response for %v", elapsed)
	}
	if got.Nodes.Total != 3 || got.Nodes.Healthy != 2 || got.Nodes.Unreachable != 1 {
		t.Errorf("counts = %+v", got.Nodes)
	}
	if !hasCondition(got, ConditionNodeUnreachable, "hangs") {
		t.Errorf("conditions = %+v", got.Conditions)
	}
	if got.Health != nodestatus.HealthDegraded {
		t.Errorf("health = %q, want degraded", got.Health)
	}
}

// A directory entry with no internal address is a node nothing can dial. It is
// reported as such rather than paying a timeout to discover it.
func TestANodeWithNoAddressIsNotDialled(t *testing.T) {
	c := collector(t)
	c.Nodes = directory(masterEntry(), inventory.Node{NodeName: "no-ip", Role: inventory.RoleWorker, Ready: true})

	got, err := c.Collect(context.Background(), testToken)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if !hasCondition(got, ConditionNodeUnaddressable, "no-ip") {
		t.Errorf("conditions = %+v", got.Conditions)
	}
	if got.Nodes.Unreachable != 1 {
		t.Errorf("counts = %+v", got.Nodes)
	}
}

// Whatever the transport said went wrong stays out of the summary: the port
// olaresd listens on, the scheme, why a certificate did not verify. The page
// gets the node name and a code.
//
// The node's address is not on that list, and deliberately so: it is a field of
// the directory the same caller can read next to this, and it is how the page
// tells two nodes with similar names apart. What must not appear is the rest of
// the sentence the transport built around it.
func TestATransportFailureIsNotQuotedToTheCaller(t *testing.T) {
	c := collector(t)
	c.Nodes = directory(masterEntry(), workerEntry("worker-1", "10.0.0.2"))
	c.Remote = func(context.Context, inventory.Node, string) (nodestatus.Status, error) {
		return nodestatus.Status{}, errors.New(`Get "http://10.0.0.2:18088/system/node-status": x509: unknown authority`)
	}

	got, err := c.Collect(context.Background(), testToken)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}

	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, leak := range []string{"x509", "18088", "http://", "authority"} {
		if strings.Contains(string(raw), leak) {
			t.Errorf("the summary leaks %q: %s", leak, raw)
		}
	}
	if !hasCondition(got, ConditionNodeUnreachable, "worker-1") {
		t.Errorf("conditions = %+v", got.Conditions)
	}
}

// A directory that cannot be read is not a cluster with no nodes.
func TestCollectFailsWhenTheDirectoryCannotBeRead(t *testing.T) {
	c := collector(t)
	c.Nodes = func(context.Context) ([]inventory.Node, error) { return nil, errors.New("apiserver unreachable") }

	if _, err := c.Collect(context.Background(), testToken); err == nil {
		t.Fatal("an unreadable directory was summarized as an empty cluster")
	}
}

func TestAnOperationInFlightReachesTheSummary(t *testing.T) {
	c := collector(t)
	c.ActivePhase = func() (nodestatus.Phase, bool) { return nodestatus.PhaseShuttingDown, true }

	got, err := c.Collect(context.Background(), testToken)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if got.Phase != nodestatus.PhaseShuttingDown {
		t.Errorf("phase = %q", got.Phase)
	}
}

// Parallel bounds how many nodes are dialled at once. A hundred-node cluster
// must not open a hundred sockets because somebody loaded a page.
func TestConcurrencyIsBounded(t *testing.T) {
	c := collector(t)
	c.Parallel = 2
	nodes := []inventory.Node{masterEntry()}
	for _, name := range []string{"w1", "w2", "w3", "w4", "w5", "w6"} {
		nodes = append(nodes, workerEntry(name, "10.0.0.9"))
	}
	c.Nodes = directory(nodes...)

	var mu sync.Mutex
	var inFlight, peak int
	c.Remote = func(ctx context.Context, n inventory.Node, _ string) (nodestatus.Status, error) {
		mu.Lock()
		inFlight++
		if inFlight > peak {
			peak = inFlight
		}
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		mu.Lock()
		inFlight--
		mu.Unlock()
		return remoteRunning(ctx, n, "")
	}

	if _, err := c.Collect(context.Background(), testToken); err != nil {
		t.Fatalf("Collect: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if peak > 2 {
		t.Errorf("%d workers were dialled at once, want at most 2", peak)
	}
}
