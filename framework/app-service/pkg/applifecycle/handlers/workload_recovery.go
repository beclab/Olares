package handlers

import "github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"

// recoverWorkload preserves a running or workloadResultQueued claim.
// If no claim survived (for example after a process restart), it re-dispatches
// only the phase's workload and leaves the persisted AM status untouched.
func recoverWorkload(w domain.World, call domain.CallWorkload) domain.Result {
	if w.WorkloadInFlight {
		return domain.Result{}
	}
	return domain.Result{Effects: []domain.Effect{call}}
}

// startUnlessInFlight suppresses a fresh-phase kickoff when this
// process already holds a workload claim. A second Tick can still
// see an empty phase Payload if PatchAM from the first Tick has
// not been re-read yet; re-dispatching would Preempt the in-flight
// cancel/uninstall/wait.
// startUnlessSameKind suppresses a fresh-phase kickoff when this
// process already holds a claim for the same WorkloadKind. A second
// Tick can still see an empty phase Payload if PatchAM from the
// first Tick has not been re-read yet; re-dispatching would Preempt
// the in-flight cancel. A different Kind (image.Pull while entering
// DownloadingCanceling) must still dispatch so Preempts can steal
// the slot.
func startUnlessSameKind(w domain.World, kind domain.WorkloadKind, start domain.Result) domain.Result {
	if w.WorkloadInFlight && w.InFlightKind == kind {
		return domain.Result{}
	}
	return start
}
