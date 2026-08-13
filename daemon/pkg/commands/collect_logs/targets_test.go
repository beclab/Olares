package collectlogs

import (
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

func TestReadyCollectTargetsSkipsNotReadyAndNoIP(t *testing.T) {
	nodes := []inventory.Node{
		{NodeName: "ready", Role: inventory.RoleWorker, IP: "10.0.0.1", Ready: true},
		{NodeName: "not-ready", Role: inventory.RoleWorker, IP: "10.0.0.2", Ready: false},
		{NodeName: "no-ip", Role: inventory.RoleMaster, IP: "", Ready: true, IsSelf: true},
		{NodeName: "master", Role: inventory.RoleMaster, IP: "10.0.0.3", Ready: true, IsSelf: true},
	}

	got := readyCollectTargets(nodes)
	if len(got) != 2 {
		t.Fatalf("targets = %+v, want 2 ready+IP nodes", got)
	}
	if got[0].Name != "ready" || got[0].IP != "10.0.0.1" || got[0].IsMaster {
		t.Errorf("first = %+v", got[0])
	}
	if got[1].Name != "master" || !got[1].IsMaster || !got[1].IsSelf {
		t.Errorf("second = %+v", got[1])
	}
}

func TestNodeCollectTimeoutRemainsFifteenMinutes(t *testing.T) {
	if nodeCollectTimeout != 15*time.Minute {
		t.Fatalf("nodeCollectTimeout = %v, want 15m", nodeCollectTimeout)
	}
}
