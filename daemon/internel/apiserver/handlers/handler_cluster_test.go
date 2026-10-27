package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterstatus"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

func decodeSummary(t *testing.T, body []byte) clusterstatus.Summary {
	t.Helper()
	var env struct {
		Data clusterstatus.Summary `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return env.Data
}

// withClusterID replaces the one call the summary makes into Kubernetes for
// the cluster's identifier.
func withClusterID(t *testing.T, id string, err error) {
	t.Helper()
	prev := clusterIDOf
	clusterIDOf = func(context.Context) (string, error) { return id, err }
	t.Cleanup(func() { clusterIDOf = prev })
}

// withRemoteNodeStatus replaces the only network call the summary makes.
func withRemoteNodeStatus(t *testing.T, fn func(context.Context, inventory.Node, string) (nodestatus.Status, error)) {
	t.Helper()
	prev := remoteNodeStatus
	remoteNodeStatus = fn
	t.Cleanup(func() { remoteNodeStatus = prev })
}

func workersMustNotBeDialled(t *testing.T) {
	t.Helper()
	withRemoteNodeStatus(t, func(_ context.Context, n inventory.Node, _ string) (nodestatus.Status, error) {
		t.Errorf("node %q was dialled by a request that should have been refused", n.NodeName)
		return nodestatus.Status{}, nil
	})
}

// asHealthySingleNodeCluster is the ordinary case: one control node, running.
func asHealthySingleNodeCluster(t *testing.T) {
	t.Helper()
	asAuthorizedUser(t)
	asNode(t, "master-1", inventory.RoleMaster, nil)
	name := "alice@olares.com"
	version := "1.12.6-rc.2"
	withCurrentState(t, clistate.State{
		TerminusState:   clistate.TerminusRunning,
		TerminusName:    &name,
		TerminusVersion: &version,
	}, time.Now())
	withClusterID(t, "ns-uid", nil)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return []inventory.Node{
			{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true, IsSelf: true},
		}, nil
	})
	workersMustNotBeDialled(t)
}

func TestClusterSummaryRefusesAnUnauthorizedRequest(t *testing.T) {
	directoryMustNotBeRead(t)
	workersMustNotBeDialled(t)
	asMaster(t)

	resp, body := callRegistered(t, "/cluster", nil)

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 without an authorization header: %s", resp.StatusCode, body)
	}
}

// Only the master can see the whole cluster. A worker's summary would describe
// the part of it that happens to be within reach.
func TestClusterSummaryIsRefusedOnAWorker(t *testing.T) {
	directoryMustNotBeRead(t)
	workersMustNotBeDialled(t)
	asAuthorizedUser(t)
	asWorker(t)

	resp, body := callRegistered(t, "/cluster", authHeaders())

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 on a worker: %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "master") {
		t.Errorf("the refusal should say the route is master-only: %s", body)
	}
}

// Reading the cluster is an ordinary signed-in read, the same as the node
// directory next to it. Nothing here changes anything.
func TestClusterSummaryIsServedToASignedInUser(t *testing.T) {
	asHealthySingleNodeCluster(t)

	resp, body := callRegistered(t, "/cluster", authHeaders())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}

	got := decodeSummary(t, body)
	if got.ClusterID != "ns-uid" {
		t.Errorf("clusterId = %q", got.ClusterID)
	}
	if got.Name != "alice@olares.com" || got.OlaresVersion != "1.12.6-rc.2" {
		t.Errorf("name/version = %q / %q", got.Name, got.OlaresVersion)
	}
	if got.Health != nodestatus.HealthHealthy || got.Phase != nodestatus.PhaseRunning {
		t.Errorf("summary = %+v", got)
	}
	if got.Connectivity != nodestatus.ConnectivityOnline {
		t.Errorf("connectivity = %q, want online in an answered request", got.Connectivity)
	}
	if got.Nodes.Total != 1 || got.Nodes.Healthy != 1 {
		t.Errorf("counts = %+v", got.Nodes)
	}
	if got.UpdatedAt.IsZero() {
		t.Errorf("updatedAt missing: %s", body)
	}
}

func TestClusterSummaryWireFieldNames(t *testing.T) {
	asHealthySingleNodeCluster(t)

	_, body := callRegistered(t, "/cluster", authHeaders())

	var env struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	for _, key := range []string{
		"clusterId", "name", "health", "connectivity", "phase",
		"olaresVersion", "nodes", "nodeList", "updatedAt", "conditions",
	} {
		if _, ok := env.Data[key]; !ok {
			t.Errorf("field %q missing from the cluster wire format: %s", key, body)
		}
	}
	if env.Data["conditions"] == nil {
		t.Errorf("conditions is null rather than []: %s", body)
	}
}

// The worker is asked over the network, and what it says lands in the summary.
func TestClusterSummaryAggregatesTheWorkers(t *testing.T) {
	asHealthySingleNodeCluster(t)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return []inventory.Node{
			{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true, IsSelf: true},
			{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.2", Ready: true},
		}, nil
	})
	withRemoteNodeStatus(t, func(_ context.Context, n inventory.Node, token string) (nodestatus.Status, error) {
		if token != testToken {
			t.Errorf("worker %q was asked with %q rather than the caller's token", n.NodeName, token)
		}
		return nodestatus.Status{Health: nodestatus.HealthDegraded, Phase: nodestatus.PhaseRunning}, nil
	})

	_, body := callRegistered(t, "/cluster", authHeaders())

	got := decodeSummary(t, body)
	if got.Health != nodestatus.HealthDegraded {
		t.Errorf("health = %q, want degraded when a worker says so", got.Health)
	}
	if got.Nodes.Total != 2 || got.Nodes.Healthy != 1 || got.Nodes.Degraded != 1 {
		t.Errorf("counts = %+v", got.Nodes)
	}
}

// The one thing that must not happen when a node is unreachable: the page
// showing a healthy cluster, or claiming the machine is off.
func TestClusterSummaryReportsAnUnreachableWorker(t *testing.T) {
	asHealthySingleNodeCluster(t)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return []inventory.Node{
			{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true, IsSelf: true},
			{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.2", Ready: true},
		}, nil
	})
	withRemoteNodeStatus(t, func(context.Context, inventory.Node, string) (nodestatus.Status, error) {
		return nodestatus.Status{}, errors.New(`Get "http://10.0.0.2:18088/system/node-status": dial tcp: i/o timeout`)
	})

	_, body := callRegistered(t, "/cluster", authHeaders())

	got := decodeSummary(t, body)
	if got.Health != nodestatus.HealthDegraded || got.Nodes.Unreachable != 1 {
		t.Errorf("summary = %+v, counts = %+v", got, got.Nodes)
	}
	// The address is a field of the directory, published next to this. The
	// sentence the transport built around it is not.
	for _, leak := range []string{"18088", "dial tcp", "http://"} {
		if strings.Contains(string(body), leak) {
			t.Errorf("the response leaks %q: %s", leak, body)
		}
	}
}

// The overview names the nodes it counted, and describes each one as far as it
// got an answer: the address always, because the directory knows it, and the
// hardware only from the node that reported it.
func TestClusterSummaryDescribesEachNodeItCounted(t *testing.T) {
	asHealthySingleNodeCluster(t)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return []inventory.Node{
			{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true, IsSelf: true},
			{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.2", Ready: true},
		}, nil
	})
	version := "1.12.6-rc.2"
	withRemoteNodeStatus(t, func(context.Context, inventory.Node, string) (nodestatus.Status, error) {
		return nodestatus.Status{
			Health:         nodestatus.HealthHealthy,
			Phase:          nodestatus.PhaseRunning,
			DeviceType:     nodestatus.DeviceTypeOlaresOne,
			OlaresdVersion: version,
			CPU:            "NVIDIA Grace",
			Memory:         "128 G",
			Disk:           "3725 G",
		}, nil
	})

	_, body := callRegistered(t, "/cluster", authHeaders())

	got := decodeSummary(t, body)
	byName := map[string]clusterstatus.NodeEntry{}
	for _, n := range got.NodeList {
		byName[n.NodeName] = n
	}
	if len(byName) != 2 {
		t.Fatalf("nodeList = %+v, want both nodes: %s", got.NodeList, body)
	}

	worker := byName["worker-1"]
	if worker.IP != "10.0.0.2" || worker.Role != inventory.RoleWorker {
		t.Errorf("worker entry = %+v", worker)
	}
	if worker.DeviceType != nodestatus.DeviceTypeOlaresOne || worker.OlaresdVersion != version {
		t.Errorf("worker did not carry what it reported: %+v", worker)
	}
	if worker.CPU != "NVIDIA Grace" || worker.Memory != "128 G" || worker.Disk != "3725 G" {
		t.Errorf("worker hardware summary = %+v", worker)
	}

	// The control node is read locally, so it is described too rather than
	// being the one entry with nothing in it.
	if master := byName["master-1"]; master.IP != "10.0.0.1" {
		t.Errorf("control node entry = %+v", master)
	}
}

// A node nothing could reach appears with what the directory knows and nothing
// it did not say.
func TestClusterSummaryLeavesAnUnreachableNodeUndescribed(t *testing.T) {
	asHealthySingleNodeCluster(t)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return []inventory.Node{
			{NodeName: "master-1", Role: inventory.RoleMaster, IP: "10.0.0.1", Ready: true, IsSelf: true},
			{NodeName: "worker-1", Role: inventory.RoleWorker, IP: "10.0.0.2", Ready: true},
		}, nil
	})
	withRemoteNodeStatus(t, func(context.Context, inventory.Node, string) (nodestatus.Status, error) {
		return nodestatus.Status{}, errors.New("i/o timeout")
	})

	_, body := callRegistered(t, "/cluster", authHeaders())

	got := decodeSummary(t, body)
	var worker clusterstatus.NodeEntry
	for _, n := range got.NodeList {
		if n.NodeName == "worker-1" {
			worker = n
		}
	}
	if worker.Connectivity != nodestatus.ConnectivityUnreachable {
		t.Errorf("worker entry = %+v, want unreachable", worker)
	}
	if worker.IP != "10.0.0.2" {
		t.Errorf("ip = %q: the directory still knows where the node is", worker.IP)
	}
	if worker.DeviceType != "" || worker.OlaresdVersion != "" || worker.CPU != "" ||
		worker.Memory != "" || worker.Disk != "" {
		t.Errorf("a node that never answered described itself: %+v", worker)
	}
}

// The directory is the one thing the summary cannot do without: an entry
// missing from it is a node nobody would know to be worried about.
func TestClusterSummaryFailureDoesNotLeakClusterInternals(t *testing.T) {
	asHealthySingleNodeCluster(t)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return nil, errors.New(`Get "https://10.0.0.1:6443/api/v1/nodes": x509: certificate signed by unknown authority`)
	})

	resp, body := callRegistered(t, "/cluster", authHeaders())
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", resp.StatusCode, body)
	}

	text := string(body)
	for _, leak := range []string{"x509", "10.0.0.1", "6443", "certificate", "https://"} {
		if strings.Contains(text, leak) {
			t.Errorf("response leaks %q: %s", leak, text)
		}
	}
	if !strings.Contains(text, clusterSummaryUnavailable) {
		t.Errorf("want the stable message %q: %s", clusterSummaryUnavailable, text)
	}
}

// A cluster whose identifier could not be read still describes its nodes.
func TestClusterSummarySurvivesAnUnreadableClusterID(t *testing.T) {
	asHealthySingleNodeCluster(t)
	withClusterID(t, "", errors.New("namespaces \"kube-system\" is forbidden"))

	resp, body := callRegistered(t, "/cluster", authHeaders())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want the rest of the summary anyway: %s", resp.StatusCode, body)
	}

	got := decodeSummary(t, body)
	if got.ClusterID != "" {
		t.Errorf("clusterId = %q, want it left empty", got.ClusterID)
	}
	if got.Nodes.Total != 1 {
		t.Errorf("counts = %+v", got.Nodes)
	}
	if strings.Contains(string(body), "forbidden") {
		t.Errorf("the response quotes the Kubernetes error: %s", body)
	}
}

// A reboot in flight is what the cluster is doing, and it comes from the
// orchestrator rather than from any node's own state machine.
func TestClusterSummaryPhaseFollowsTheOperationInFlight(t *testing.T) {
	asHealthySingleNodeCluster(t)
	withClusterOperations(t, &fakeOperations{activePhase: nodestatus.PhaseRestarting})

	_, body := callRegistered(t, "/cluster", authHeaders())

	got := decodeSummary(t, body)
	if got.Phase != nodestatus.PhaseRestarting {
		t.Errorf("phase = %q, want restarting: %s", got.Phase, body)
	}
	if got.Health != nodestatus.HealthHealthy {
		t.Errorf("health = %q: a cluster rebooting on purpose is not a cluster in trouble", got.Health)
	}
}

// The routes exist before the orchestrator has been wired up. Reading the
// cluster is not an operation, so it answers anyway.
func TestClusterSummaryWorksWithoutAnOrchestrator(t *testing.T) {
	asHealthySingleNodeCluster(t)
	prev := clusterOperations
	clusterOperations = nil
	t.Cleanup(func() { clusterOperations = prev })

	resp, body := callRegistered(t, "/cluster", authHeaders())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}
	if got := decodeSummary(t, body); got.Phase != nodestatus.PhaseRunning {
		t.Errorf("phase = %q", got.Phase)
	}
}

// A cluster with nothing in its directory is one this daemon could not
// describe, not a healthy one with no work to do.
func TestAnEmptyDirectoryIsReportedAsUnknown(t *testing.T) {
	asHealthySingleNodeCluster(t)
	withNodeDirectory(t, func(context.Context) ([]inventory.Node, error) {
		return []inventory.Node{}, nil
	})

	resp, body := callRegistered(t, "/cluster", authHeaders())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, body)
	}

	got := decodeSummary(t, body)
	if got.Health != nodestatus.HealthUnknown {
		t.Errorf("health = %q, want unknown", got.Health)
	}
	if got.Conditions == nil {
		t.Errorf("conditions = null: %s", body)
	}
}

// The route reads. Nothing it touches may start, stop or record an operation.
func TestClusterSummaryStartsNothing(t *testing.T) {
	asHealthySingleNodeCluster(t)
	f := withClusterOperations(t, &fakeOperations{op: sampleOperation()})

	if _, body := callRegistered(t, "/cluster", authHeaders()); len(f.requests()) != 0 {
		t.Errorf("reading the cluster created an operation: %s", body)
	}
}

func TestNodeStatusURLBracketsIPv6Addresses(t *testing.T) {
	got := nodeStatusURL("fd00::12")
	if got != "http://[fd00::12]:18088" {
		t.Fatalf("nodeStatusURL = %q", got)
	}
}
