package clusterstatus

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

var buildAt = time.Date(2026, 8, 5, 7, 0, 0, 0, time.UTC)

func identity() Identity {
	return Identity{ClusterID: "ns-uid", Name: "alice@olares.com", Version: "1.12.6-rc.2"}
}

func healthyMaster(name string) NodeView {
	return NodeView{
		Name:         name,
		Role:         inventory.RoleMaster,
		Ready:        true,
		Health:       nodestatus.HealthHealthy,
		Connectivity: nodestatus.ConnectivityOnline,
		Phase:        nodestatus.PhaseRunning,
	}
}

func healthyWorker(name string) NodeView {
	v := healthyMaster(name)
	v.Role = inventory.RoleWorker
	return v
}

func conditionTypes(s Summary) []string {
	out := make([]string, 0, len(s.Conditions))
	for _, c := range s.Conditions {
		out = append(out, c.Type)
	}
	return out
}

func hasCondition(s Summary, ty, node string) bool {
	for _, c := range s.Conditions {
		if c.Type == ty && c.Node == node {
			return true
		}
	}
	return false
}

func TestEveryNodeHealthyMakesTheClusterHealthy(t *testing.T) {
	got := Build(identity(), []NodeView{healthyMaster("master-1"), healthyWorker("worker-1")}, "", buildAt)

	if got.Health != nodestatus.HealthHealthy {
		t.Errorf("health = %q, want healthy: %+v", got.Health, got)
	}
	if got.Nodes.Total != 2 || got.Nodes.Healthy != 2 {
		t.Errorf("counts = %+v", got.Nodes)
	}
	if len(got.Conditions) != 0 {
		t.Errorf("a healthy cluster reported conditions: %v", conditionTypes(got))
	}
	if got.ClusterID != "ns-uid" || got.Name != "alice@olares.com" || got.OlaresVersion != "1.12.6-rc.2" {
		t.Errorf("identity dropped: %+v", got)
	}
	if !got.UpdatedAt.Equal(buildAt) {
		t.Errorf("updatedAt = %v, want the time the summary was assembled", got.UpdatedAt)
	}
}

// The cluster is answering, so the link user-service asked over is up. This is
// the master's own connectivity and never a node's.
func TestASuccessfulSummaryIsOnline(t *testing.T) {
	for _, nodes := range [][]NodeView{
		{healthyMaster("master-1")},
		{},
		{unreachable("worker-1")},
	} {
		if got := Build(identity(), nodes, "", buildAt); got.Connectivity != nodestatus.ConnectivityOnline {
			t.Errorf("connectivity = %q for %v, want online", got.Connectivity, nodes)
		}
	}
}

func unreachable(name string) NodeView {
	return NodeView{
		Name:         name,
		Role:         inventory.RoleWorker,
		Ready:        true,
		Health:       nodestatus.HealthUnknown,
		Connectivity: nodestatus.ConnectivityUnreachable,
		Phase:        nodestatus.PhaseUnknown,
	}
}

func TestOneDegradedNodeDegradesTheCluster(t *testing.T) {
	bad := healthyWorker("worker-1")
	bad.Health = nodestatus.HealthDegraded

	got := Build(identity(), []NodeView{healthyMaster("master-1"), bad}, "", buildAt)

	if got.Health != nodestatus.HealthDegraded {
		t.Errorf("health = %q, want degraded", got.Health)
	}
	if got.Nodes.Healthy != 1 || got.Nodes.Degraded != 1 {
		t.Errorf("counts = %+v", got.Nodes)
	}
	if !hasCondition(got, ConditionNodeDegraded, "worker-1") {
		t.Errorf("the failing node is not named in the conditions: %+v", got.Conditions)
	}
}

func TestANotReadyNodeDegradesTheCluster(t *testing.T) {
	bad := healthyWorker("worker-1")
	bad.Ready = false

	got := Build(identity(), []NodeView{healthyMaster("master-1"), bad}, "", buildAt)

	if got.Health != nodestatus.HealthDegraded {
		t.Errorf("health = %q, want degraded when Kubernetes says a node is NotReady", got.Health)
	}
	if !hasCondition(got, ConditionNodeNotReady, "worker-1") {
		t.Errorf("conditions = %+v", got.Conditions)
	}
}

// A node that did not answer is a node in trouble, but nothing here has seen
// it go down: reporting it as offline would tell the user the machine is off.
func TestAnUnreachableNodeDegradesTheClusterAndIsNotCalledOffline(t *testing.T) {
	got := Build(identity(), []NodeView{healthyMaster("master-1"), unreachable("worker-1")}, "", buildAt)

	if got.Health != nodestatus.HealthDegraded {
		t.Errorf("health = %q, want degraded", got.Health)
	}
	if got.Nodes.Unreachable != 1 {
		t.Errorf("unreachable count = %d, want 1: %+v", got.Nodes.Unreachable, got.Nodes)
	}
	if got.Nodes.Unknown != 1 {
		t.Errorf("an unreachable node's health is unknown: %+v", got.Nodes)
	}
	if !hasCondition(got, ConditionNodeUnreachable, "worker-1") {
		t.Errorf("conditions = %+v", got.Conditions)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), `"offline"`) {
		t.Errorf("a timeout was turned into a confirmed shutdown: %s", raw)
	}
}

// Nothing is wrong and nothing is confirmed either. Calling that healthy would
// be a guess in the one direction the user cannot check.
func TestNodesThatCannotConfirmThemselvesLeaveTheClusterUnknown(t *testing.T) {
	uncertain := healthyWorker("worker-1")
	uncertain.Health = nodestatus.HealthUnknown

	got := Build(identity(), []NodeView{healthyMaster("master-1"), uncertain}, "", buildAt)

	if got.Health != nodestatus.HealthUnknown {
		t.Errorf("health = %q, want unknown", got.Health)
	}
	if !hasCondition(got, ConditionNodeStatusUnknown, "worker-1") {
		t.Errorf("conditions = %+v", got.Conditions)
	}
}

func TestHealthBucketsAlwaysAddUpToTheNodeTotal(t *testing.T) {
	degraded := healthyWorker("worker-1")
	degraded.Health = nodestatus.HealthDegraded

	got := Build(identity(), []NodeView{healthyMaster("master-1"), degraded, unreachable("worker-2")}, "", buildAt)

	if sum := got.Nodes.Healthy + got.Nodes.Degraded + got.Nodes.Unknown; sum != got.Nodes.Total {
		t.Errorf("healthy+degraded+unknown = %d, total = %d: %+v", sum, got.Nodes.Total, got.Nodes)
	}
}

// An empty directory is not an empty, healthy cluster. It is a cluster this
// daemon could not describe, and the caller is told that in as many words.
func TestAnEmptyClusterIsUnknown(t *testing.T) {
	got := Build(identity(), nil, "", buildAt)

	if got.Health != nodestatus.HealthUnknown {
		t.Errorf("health = %q, want unknown for a cluster with no nodes", got.Health)
	}
	if got.Phase != nodestatus.PhaseUnknown {
		t.Errorf("phase = %q, want unknown", got.Phase)
	}
	if got.Nodes.Total != 0 {
		t.Errorf("counts = %+v", got.Nodes)
	}
	if !hasCondition(got, ConditionClusterInventoryEmpty, "") {
		t.Errorf("conditions = %+v", got.Conditions)
	}
}

func TestCollectionsAreNeverNullOnTheWire(t *testing.T) {
	raw, err := json.Marshal(Build(identity(), nil, "", buildAt))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	if wire["conditions"] == nil {
		t.Errorf("conditions is null rather than []: %s", raw)
	}
	if wire["nodes"] == nil {
		t.Errorf("node counts are null rather than zeroes: %s", raw)
	}
}

func TestThePhaseIsTheControlNodes(t *testing.T) {
	master := healthyMaster("master-1")
	master.Phase = nodestatus.PhaseMaintenance
	worker := healthyWorker("worker-1")
	worker.Phase = nodestatus.PhaseRunning

	if got := Build(identity(), []NodeView{worker, master}, "", buildAt); got.Phase != nodestatus.PhaseMaintenance {
		t.Errorf("phase = %q, want the control node's: %+v", got.Phase, got)
	}
}

// An operation in flight is what the cluster is doing, whatever any single
// node's own state machine has reached.
func TestAnOperationInFlightOwnsThePhase(t *testing.T) {
	for _, override := range []nodestatus.Phase{nodestatus.PhaseRestarting, nodestatus.PhaseShuttingDown} {
		got := Build(identity(), []NodeView{healthyMaster("master-1")}, override, buildAt)
		if got.Phase != override {
			t.Errorf("phase = %q, want %q", got.Phase, override)
		}
		if got.Health != nodestatus.HealthHealthy {
			t.Errorf("health = %q: a cluster rebooting on purpose is not a cluster in trouble", got.Health)
		}
	}
}

func TestWithoutAnOperationThePhaseIsLeftAlone(t *testing.T) {
	if got := Build(identity(), []NodeView{healthyMaster("master-1")}, "", buildAt); got.Phase != nodestatus.PhaseRunning {
		t.Errorf("phase = %q, want the control node's own phase", got.Phase)
	}
}

// The reason a node is unusable can be more specific than what its status
// implies: an entry with no address was never dialled at all.
func TestAStatedReasonWinsOverTheDerivedOne(t *testing.T) {
	node := unreachable("worker-1")
	node.Reason = ConditionNodeUnaddressable

	got := Build(identity(), []NodeView{healthyMaster("master-1"), node}, "", buildAt)

	if !hasCondition(got, ConditionNodeUnaddressable, "worker-1") {
		t.Errorf("conditions = %+v", got.Conditions)
	}
	if hasCondition(got, ConditionNodeUnreachable, "worker-1") {
		t.Errorf("a node reported twice: %+v", got.Conditions)
	}
}

// Every failing node keeps its own entry. Collapsing them into "some nodes are
// unreachable" is exactly the sentence that sends somebody to the wrong rack.
func TestEveryFailingNodeKeepsAnEntry(t *testing.T) {
	got := Build(identity(), []NodeView{
		healthyMaster("master-1"),
		unreachable("worker-1"),
		unreachable("worker-2"),
	}, "", buildAt)

	if len(got.Conditions) != 2 {
		t.Fatalf("conditions = %+v, want one per failing node", got.Conditions)
	}
	if !hasCondition(got, ConditionNodeUnreachable, "worker-1") || !hasCondition(got, ConditionNodeUnreachable, "worker-2") {
		t.Errorf("conditions = %+v", got.Conditions)
	}
}

// A cluster whose identity could not be read still describes its nodes, and
// says which fact is missing rather than inventing one.
func TestAMissingClusterIDIsReportedRatherThanFabricated(t *testing.T) {
	id := identity()
	id.ClusterID = ""

	got := Build(id, []NodeView{healthyMaster("master-1")}, "", buildAt)

	if got.ClusterID != "" {
		t.Errorf("clusterId = %q, want it left empty", got.ClusterID)
	}
	if !hasCondition(got, ConditionClusterIdentityUnavailable, "") {
		t.Errorf("conditions = %+v", got.Conditions)
	}
	if got.Health != nodestatus.HealthHealthy {
		t.Errorf("health = %q: the nodes are fine, only the label is missing", got.Health)
	}
}

func TestWireFieldNames(t *testing.T) {
	raw, err := json.Marshal(Build(identity(), []NodeView{healthyMaster("master-1")}, "", buildAt))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	for _, key := range []string{
		"clusterId", "name", "health", "connectivity", "phase",
		"olaresVersion", "nodes", "updatedAt", "conditions",
	} {
		if _, ok := wire[key]; !ok {
			t.Errorf("field %q missing: %s", key, raw)
		}
	}
	counts, ok := wire["nodes"].(map[string]any)
	if !ok {
		t.Fatalf("nodes is not an object: %s", raw)
	}
	for _, key := range []string{"total", "healthy", "degraded", "unknown", "unreachable"} {
		if _, ok := counts[key]; !ok {
			t.Errorf("node count %q missing: %s", key, raw)
		}
	}
}
