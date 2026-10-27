package state

import (
	"sync"
	"testing"
	"time"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/pkg/utils"
)

func restoreState(t *testing.T) {
	t.Helper()
	prevPublished := published.Load()
	TerminusStateMu.Lock()
	prevState := CurrentState
	TerminusStateMu.Unlock()

	t.Cleanup(func() {
		TerminusStateMu.Lock()
		CurrentState = prevState
		TerminusStateMu.Unlock()
		published.Store(prevPublished)
	})
}

func setCurrentState(t *testing.T, s clistate.State) {
	t.Helper()
	TerminusStateMu.Lock()
	defer TerminusStateMu.Unlock()
	CurrentState = s
}

// A refresh holds TerminusStateMu across every probe it makes: NetworkManager,
// the filesystem, and several apiserver round trips. A reader that waits for
// that mutex waits for the slowest of them, and the aggregator upstream reads
// the timeout as a node that is gone.
func TestSnapshotDoesNotWaitForARefreshHoldingTheStateLock(t *testing.T) {
	restoreState(t)
	publishObservation(true, time.Now())

	TerminusStateMu.Lock()
	defer TerminusStateMu.Unlock()

	done := make(chan struct{})
	go func() {
		defer close(done)
		Snapshot()
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Snapshot blocked while a refresh held the state mutex")
	}
}

func TestSnapshotReportsWhenTheStateWasLastRefreshed(t *testing.T) {
	restoreState(t)
	refreshedAt := time.Now().Add(-42 * time.Second)

	publishObservation(true, refreshedAt)

	_, observedAt := Snapshot()
	if !observedAt.Equal(refreshedAt) {
		t.Errorf("observedAt = %v, want the time of the last refresh %v", observedAt, refreshedAt)
	}
}

func TestSnapshotBeforeTheFirstRefreshHasNoObservation(t *testing.T) {
	restoreState(t)
	published.Store(nil) // a process that has just started

	if _, observedAt := Snapshot(); !observedAt.IsZero() {
		t.Errorf("observedAt = %v, want the zero time before anything is published", observedAt)
	}

	// Whatever the first refresh did, it did not complete.
	publishObservation(false, time.Now())

	_, observedAt := Snapshot()
	if !observedAt.IsZero() {
		t.Errorf("observedAt = %v, want the zero time so callers can say they do not know", observedAt)
	}
}

// A refresh that gave up halfway — no route to the apiserver, no host IP —
// still moves the state machine, and that much is worth publishing. What it
// must not do is restamp the rest of the data as freshly observed.
func TestObservationTimeOnlyMovesOnACompletedRefresh(t *testing.T) {
	restoreState(t)
	completedAt := time.Now().Add(-time.Minute)
	setCurrentState(t, clistate.State{TerminusState: clistate.TerminusRunning, CpuInfo: "NVIDIA Grace"})
	publishObservation(true, completedAt)

	setCurrentState(t, clistate.State{TerminusState: clistate.NetworkNotReady, CpuInfo: "NVIDIA Grace"})
	publishObservation(false, time.Now())

	snap, observedAt := Snapshot()
	if !observedAt.Equal(completedAt) {
		t.Errorf("observedAt = %v, want it pinned to the last completed refresh %v", observedAt, completedAt)
	}
	if snap.TerminusState != clistate.NetworkNotReady {
		t.Errorf("state = %q, want the failing refresh to still be visible", snap.TerminusState)
	}
}

func TestPublishedSnapshotIsNotChangedByLaterWrites(t *testing.T) {
	restoreState(t)
	setCurrentState(t, clistate.State{
		TerminusState: clistate.TerminusRunning,
		GPUList:       []string{"NVIDIA GB10"},
		Pressure:      []utils.NodePressure{{Type: "MemoryPressure"}},
	})
	publishObservation(true, time.Now())

	TerminusStateMu.Lock()
	CurrentState.TerminusState = clistate.Restarting
	CurrentState.GPUList[0] = "tampered"
	CurrentState.GPUList = append(CurrentState.GPUList, "another")
	CurrentState.Pressure[0].Type = "tampered"
	TerminusStateMu.Unlock()

	snap, _ := Snapshot()
	if snap.TerminusState != clistate.TerminusRunning {
		t.Errorf("published state followed a later change: %q", snap.TerminusState)
	}
	if len(snap.GPUList) != 1 || snap.GPUList[0] != "NVIDIA GB10" {
		t.Errorf("published GPU list shares memory with the live state: %+v", snap.GPUList)
	}
	if len(snap.Pressure) != 1 || snap.Pressure[0].Type != "MemoryPressure" {
		t.Errorf("published pressures share memory with the live state: %+v", snap.Pressure)
	}
}

func TestSnapshotDoesNotShareSlicesWithTheLiveState(t *testing.T) {
	restoreState(t)
	setCurrentState(t, clistate.State{
		GPUList:  []string{"NVIDIA GB10"},
		Pressure: []utils.NodePressure{{Type: "MemoryPressure"}},
	})
	publishObservation(true, time.Now())

	snap, _ := Snapshot()
	snap.GPUList[0] = "tampered"
	snap.Pressure[0].Type = "tampered"

	again, _ := Snapshot()
	if again.GPUList[0] != "NVIDIA GB10" || again.Pressure[0].Type != "MemoryPressure" {
		t.Errorf("one caller mutating its snapshot changed what the next one reads: %+v", again)
	}

	TerminusStateMu.Lock()
	defer TerminusStateMu.Unlock()
	if CurrentState.GPUList[0] != "NVIDIA GB10" || CurrentState.Pressure[0].Type != "MemoryPressure" {
		t.Errorf("a caller mutating its snapshot changed the live state: %+v", CurrentState)
	}
}

func TestSnapshotStaysReadableWhilePublishesLandUnderIt(t *testing.T) {
	restoreState(t)
	setCurrentState(t, clistate.State{TerminusState: clistate.TerminusRunning, GPUList: []string{"NVIDIA GB10"}})
	publishObservation(true, time.Now())

	var wg sync.WaitGroup
	stop := make(chan struct{})
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					snap, _ := Snapshot()
					if len(snap.GPUList) != 1 {
						t.Errorf("torn snapshot: %+v", snap)
						return
					}
				}
			}
		}()
	}

	for i := 0; i < 50; i++ {
		publishObservation(i%2 == 0, time.Now())
	}
	close(stop)
	wg.Wait()
}
