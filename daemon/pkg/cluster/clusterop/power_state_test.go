package clusterop

import (
	"sync"
	"testing"

	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

// The state this daemon is in is rewritten by the refresh loop while requests
// are being served, and the only coherent view of it is the published snapshot.
// Reading the live struct instead is a data race, and the value it returns is a
// state nothing ever published: the refresh writes the fields one at a time.
//
// This is the check that decides whether the machine may be powered, so the
// answer being from a half-written state is not a cosmetic problem.
func TestValidatePowerOpReadsThePublishedSnapshot(t *testing.T) {
	before, _ := state.Snapshot()
	t.Cleanup(func() { state.ChangeTerminusStateTo(before.TerminusState) })

	cmd := &fakeCommand{Operation: commands.Operation{Name: commands.Reboot}}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			// The transitions the refresh loop drives the daemon through
			// while a request is in flight.
			state.ChangeTerminusStateTo(state.TerminusRunning)
			state.ChangeTerminusStateTo(state.AddingNode)
		}
	}()

	for i := 0; i < 200; i++ {
		// The return value is not the point: whether this read is safe is.
		_ = validatePowerOp(cmd)
	}
	wg.Wait()
}

// Whatever it reads, it has to keep deciding the same way: a reboot is refused
// while a node is joining the cluster and allowed once the system is running.
func TestValidatePowerOpStillRefusesAStateThatCannotBePowered(t *testing.T) {
	before, _ := state.Snapshot()
	t.Cleanup(func() { state.ChangeTerminusStateTo(before.TerminusState) })

	cmd := &fakeCommand{Operation: commands.Operation{Name: commands.Reboot}}

	state.ChangeTerminusStateTo(state.AddingNode)
	if err := validatePowerOp(cmd); err == nil {
		t.Error("a reboot was allowed while a node was joining the cluster")
	}

	state.ChangeTerminusStateTo(state.TerminusRunning)
	if err := validatePowerOp(cmd); err != nil {
		t.Errorf("a reboot was refused on a running system: %v", err)
	}
}
