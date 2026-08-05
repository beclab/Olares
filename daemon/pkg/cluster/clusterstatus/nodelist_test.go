package clusterstatus

import (
	"encoding/json"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

// The overview lists the nodes it counted. Counts alone say two nodes are
// degraded without saying which, and the page's next move is always to name
// them.
func TestSummaryListsTheNodesItCounted(t *testing.T) {
	master := healthyMaster("master-1")
	master.IP = "10.0.0.1"
	master.DeviceType = nodestatus.DeviceTypeOlaresOne
	master.OlaresdVersion = "1.12.6-rc.2"
	master.CPU = "NVIDIA Grace"
	master.Memory = "128 G"
	master.Disk = "3725 G"

	got := Build(identity(), []NodeView{master, unreachable("worker-1")}, "", buildAt)

	if len(got.NodeList) != 2 {
		t.Fatalf("nodeList = %+v, want one entry per node", got.NodeList)
	}
	first := got.NodeList[0]
	if first.Name != "master-1" || first.Role != inventory.RoleMaster {
		t.Errorf("first entry = %+v", first)
	}
	if first.IP != "10.0.0.1" {
		t.Errorf("ip = %q, want the address from the directory", first.IP)
	}
	if first.DeviceType != nodestatus.DeviceTypeOlaresOne {
		t.Errorf("deviceType = %q", first.DeviceType)
	}
	if first.OlaresdVersion != "1.12.6-rc.2" {
		t.Errorf("olaresdVersion = %q", first.OlaresdVersion)
	}
	if first.CPU != "NVIDIA Grace" || first.Memory != "128 G" || first.Disk != "3725 G" {
		t.Errorf("hardware summary = %+v", first)
	}
	if first.Health != nodestatus.HealthHealthy || first.Connectivity != nodestatus.ConnectivityOnline {
		t.Errorf("first entry lost its status: %+v", first)
	}
}

// A node that did not answer contributes what the directory knows and nothing
// else. Carrying over the last hardware it reported would show a page that
// describes a machine nobody can currently reach as though it just answered.
func TestAnUnreachableNodeContributesNoHardware(t *testing.T) {
	gone := unreachable("worker-1")
	gone.IP = "10.0.0.2"

	got := Build(identity(), []NodeView{healthyMaster("master-1"), gone}, "", buildAt)

	var entry NodeEntry
	for _, n := range got.NodeList {
		if n.Name == "worker-1" {
			entry = n
		}
	}
	if entry.Name == "" {
		t.Fatalf("the unreachable node is missing from the list: %+v", got.NodeList)
	}
	if entry.Connectivity != nodestatus.ConnectivityUnreachable {
		t.Errorf("connectivity = %q, want unreachable", entry.Connectivity)
	}
	if entry.IP != "10.0.0.2" {
		t.Errorf("ip = %q: the directory still knows the address", entry.IP)
	}
	if entry.OlaresdVersion != "" || entry.CPU != "" || entry.Memory != "" || entry.Disk != "" {
		t.Errorf("a node that did not answer described its hardware anyway: %+v", entry)
	}
}

func TestNodeListIsAlwaysAnArray(t *testing.T) {
	raw, err := json.Marshal(Build(identity(), nil, "", buildAt))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	got, ok := fields["nodeList"]
	if !ok {
		t.Fatalf("nodeList missing: %s", raw)
	}
	if string(got) != "[]" {
		t.Errorf("nodeList = %s, want [] rather than null on an empty cluster", got)
	}
}

func TestNodeEntryWireFieldNames(t *testing.T) {
	master := healthyMaster("master-1")
	master.IP = "10.0.0.1"

	raw, err := json.Marshal(Build(identity(), []NodeView{master}, "", buildAt))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded struct {
		NodeList []map[string]any `json:"nodeList"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	if len(decoded.NodeList) != 1 {
		t.Fatalf("nodeList = %+v", decoded.NodeList)
	}
	for _, key := range []string{
		"name", "role", "ready", "health", "connectivity", "phase",
		"ip", "deviceType", "olaresdVersion", "cpu", "memory", "disk",
	} {
		if _, ok := decoded.NodeList[0][key]; !ok {
			t.Errorf("field %q missing from a node entry: %s", key, raw)
		}
	}
}
