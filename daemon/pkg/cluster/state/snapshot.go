package state

import (
	"sync/atomic"
	"time"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
)

// observation is a published copy of the state and the time it was observed.
// Once stored it is never written to again; a change publishes a new one.
type observation struct {
	state      clistate.State
	observedAt time.Time
}

// published is the last observation readers may see. It is deliberately not
// guarded by TerminusStateMu: a refresh holds that mutex from its first probe
// to its last, and every one of those probes — NetworkManager, the filesystem,
// several apiserver round trips — can take seconds on a node that is already
// in trouble. A status reader that queued behind it would answer only once the
// slow thing finished, and the aggregator upstream reads that delay as a node
// that is gone. So readers take a pointer and never a lock.
var published atomic.Pointer[observation]

// Snapshot returns the last published state and the time it was observed. The
// zero time means no refresh has completed yet, which callers report as
// unknown rather than as "just now".
func Snapshot() (clistate.State, time.Time) {
	obs := published.Load()
	if obs == nil {
		return clistate.State{}, time.Time{}
	}
	// Copied per read as well as per publish: the stored observation is shared
	// by every reader, so handing out its slices would let one caller's edit
	// show up in the next caller's answer.
	return copyState(obs.state), obs.observedAt
}

// publishObservation publishes the live state.
//
// observed says whether the refresh that produced it ran to a conclusion. A
// refresh that gave up halfway still moved the state machine, and that much is
// worth publishing, but the data behind it was never re-read: the observation
// time stays where the last complete refresh left it, so a stale answer reads
// as stale instead of being restamped as current.
func publishObservation(observed bool, at time.Time) {
	TerminusStateMu.Lock()
	defer TerminusStateMu.Unlock()
	publishLocked(observed, at)
}

// publishLocked publishes the live state. The caller must hold
// TerminusStateMu, which is what makes the copy it takes a coherent one.
func publishLocked(observed bool, at time.Time) {
	observedAt := time.Time{}
	if prev := published.Load(); prev != nil {
		observedAt = prev.observedAt
	}
	if observed {
		observedAt = at
	}
	published.Store(&observation{state: copyState(CurrentState), observedAt: observedAt})
}

// copyState returns a copy that shares no slice with s. The pointer fields are
// left as they are: every writer in the daemon replaces them rather than
// writing through them.
func copyState(s clistate.State) clistate.State {
	if s.GPUList != nil {
		s.GPUList = append([]string(nil), s.GPUList...)
	}
	if s.Pressure != nil {
		s.Pressure = append([]clistate.NodePressure(nil), s.Pressure...)
	}
	return s
}
