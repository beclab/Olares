package terminus

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/stretchr/testify/assert"
)

func healthyMaster() *MasterInfo {
	return &MasterInfo{
		JuiceFSEnabled:      true,
		KubernetesInstalled: true,
		OlaresInstalled:     true,
		KubernetesType:      common.K3s,
		OlaresVersion:       "1.12.7",
		MasterNodeName:      "olares",
		AllNodes:            []string{"olares", "worker-1"},
	}
}

func TestCheckJoinEligibilityAcceptsHealthyMaster(t *testing.T) {
	assert.NoError(t, CheckJoinEligibility(healthyMaster(), "1.12.7", "worker-2", "", false))
}

func TestCheckJoinEligibilityRejectsUnusableMaster(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*MasterInfo)
		expects string
	}{
		{"juicefs disabled", func(m *MasterInfo) { m.JuiceFSEnabled = false }, "[JuiceFS]"},
		{"kubernetes missing", func(m *MasterInfo) { m.KubernetesInstalled = false }, "[Kubernetes]"},
		{"olares missing", func(m *MasterInfo) { m.OlaresInstalled = false }, "[Olares]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := healthyMaster()
			tt.mutate(info)
			err := CheckJoinEligibility(info, "1.12.7", "worker-2", "", false)
			assert.ErrorContains(t, err, tt.expects)
			assert.ErrorContains(t, err, "unable to add current node to the cluster")
		})
	}
}

// Every reason must be reported in one go, so the operator does not discover
// them one failed run at a time.
func TestCheckJoinEligibilityAggregatesAllReasons(t *testing.T) {
	info := healthyMaster()
	info.JuiceFSEnabled = false
	err := CheckJoinEligibility(info, "1.12.6", "olares", "", false)
	assert.ErrorContains(t, err, "[JuiceFS]")
	assert.ErrorContains(t, err, "[Version]")
	assert.ErrorContains(t, err, "[NodeName]")
}

// The node name is checked only once the join flow has had its chance to rename
// this machine; before that, a collision is expected rather than fatal.
func TestCheckJoinEligibilityNodeNameOnlyAfterBootstrap(t *testing.T) {
	assert.NoError(t, CheckJoinEligibility(healthyMaster(), "1.12.7", "olares", "", true))
	assert.ErrorContains(t,
		CheckJoinEligibility(healthyMaster(), "1.12.7", "OLARES", "", false),
		"[NodeName]")
}

// Matching versions means matching olares-cli binaries, which only the bootstrap
// script can arrange, so the remedy is the same however far the flow got.
func TestCheckJoinEligibilityVersionMismatch(t *testing.T) {
	for _, bootstrapping := range []bool{true, false} {
		err := CheckJoinEligibility(healthyMaster(), "1.12.6", "worker-2", "https://cdn.olares.cn", bootstrapping)
		assert.ErrorContains(t, err, "the master node is running Olares 1.12.7, but this node's installer is 1.12.6")
		assert.ErrorContains(t, err, "node join-command")
	}

	// An unknown version on either side is not a mismatch worth reporting.
	unknownMaster := healthyMaster()
	unknownMaster.OlaresVersion = ""
	assert.NoError(t, CheckJoinEligibility(unknownMaster, "1.12.7", "worker-2", "", true))
	assert.NoError(t, CheckJoinEligibility(healthyMaster(), "", "worker-2", "", true))
}
