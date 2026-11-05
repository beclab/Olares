package domain

import (
	"time"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// PayloadPendingQueuedAt is the AM.Status.Payload key holding the
// RFC3339Nano timestamp when the AM first entered Pending. The
// download-slot FIFO orders Pending AMs by this value.
const PayloadPendingQueuedAt = "pending.queuedAt"

// World is the immutable snapshot passed to every handler. It is a
// read-only view of the cluster state relevant to the reconcile
// currently in progress. Handlers MUST NOT mutate any field on the
// World or on the objects it points to; the Runner builds a fresh
// World for every Reconcile.
//
// Any decision a handler needs to make about "what is happening
// outside the AM" (workload readiness, cluster-wide constraints,
// injected time-of-day) MUST be encoded as a field on World. Handlers
// MUST NOT call cluster APIs directly. See CONTRACT.md rule 1.
type World struct {
	// AM is the current ApplicationManager. NEVER nil for a legitimate
	// Reconcile; the Runner returns early before invoking Engine.Step
	// if the AM cannot be loaded.
	AM *appsv1.ApplicationManager

	// Now is the wall-clock time as observed by the Runner at the
	// start of the Reconcile. Handlers use this instead of
	// time.Now() so tests can inject deterministic timelines.
	Now time.Time

	// AppExists indicates whether the paired Application CR (which is
	// created by helm and cleaned up on uninstall) is currently
	// present. Multiple states (PendingApp, RunningApp,
	// UninstalledApp) branch on this in the legacy code to detect
	// stale AMs whose Application CR has been GC'd.
	AppExists bool

	// Workload summarizes the runtime status of the workload managed
	// by helm (deployment/statefulset readiness, pod failures,
	// pending unschedulable pods). It is a pre-digested projection so
	// handlers do not need to walk owner references.
	Workload WorkloadStatus

	// WorkloadInFlight is true while this process has a workload
	// goroutine running for the AM or its completion callback is
	// queued and awaiting consumption.
	WorkloadInFlight bool

	// InFlightKind is the WorkloadKind that currently holds the
	// process-local claim. Empty when WorkloadInFlight is false.
	// Canceling Ticks use this to tell "the pull we must preempt"
	// from "the cancel we already dispatched" when Status.Payload
	// has not caught up yet.
	InFlightKind WorkloadKind

	// ClusterStats holds cross-AM aggregate data required by handlers
	// that enforce cluster-wide constraints (e.g. "at most one
	// Downloading app at a time" in pkg/appstate/pending_app.go).
	// See plan §1.5 discussion for why this replaces the process-wide
	// mutex in the legacy code.
	ClusterStats ClusterStats

	// HasWorkloadReplicas mirrors appcfg.ApplicationConfig.
	// HasWorkloadReplicas for the AM's Spec.Config: the app declares
	// per-workload replica counts in its OlaresManifest and therefore
	// follows the two-phase install.
	//
	// Two-phase means the initial helm install renders every workload
	// at replicas=0 (see appinstaller.HelmOps.SetValues), so the
	// resource-pressure validation runs AFTER helm has told us what
	// the release actually costs, and a second helm upgrade
	// (Scale(-1)) brings the workloads up to their declared sizes.
	// Apps without the declaration keep the single-phase flow:
	// validate first, then one helm install that lands the app
	// running.
	//
	// False when Spec.Config is empty or fails to parse — the same
	// fallback the legacy controllers use (see
	// appenv_controller.appMgrHasWorkloadReplicas).
	HasWorkloadReplicas bool

	// Action carries the Action that fired the transition currently
	// being dispatched. Runner populates it just before calling
	// Engine.Step so a handler can inspect Action-specific data
	// (e.g. WorkloadFailed.Reason / Message, TriggerOp.Op) that
	// does not otherwise live on the AM.
	//
	// Nil ONLY when Engine.Step is invoked without an action (which
	// itself returns ErrNoTransition immediately, so a handler
	// running with Action==nil is impossible in production). Tests
	// that call handlers directly with a hand-crafted World may
	// leave Action nil when they don't need Action data.
	Action Action
}

// Current returns the current state of the AM under reconcile, or
// InitialState ("") when the AM has not been written by the state
// machine yet.
func (w World) Current() State {
	if w.AM == nil {
		return InitialState
	}
	return w.AM.Status.State
}

// LastOp returns the operation type most recently requested for the AM
// (through Spec.OpType, which the apiserver writes before flipping the
// state). Returns "" if none has been requested yet.
func (w World) LastOp() OpType {
	if w.AM == nil {
		return ""
	}
	return w.AM.Spec.OpType
}

// WorkloadStatus is the pre-digested projection of the workload's
// runtime state used by handlers to decide readiness and failure
// transitions. Populated by the Runner from informer caches (Deployment,
// StatefulSet, Pod, ImageManager) before invoking Engine.Step.
type WorkloadStatus struct {
	// Ready is true when the workload has reached its steady-state
	// (all replicas ready / job succeeded / stateful set converged).
	Ready bool

	// Failed is true when the workload has failed terminally
	// (CrashLoopBackOff, image pull permanent failure, unschedulable
	// past deadline).
	Failed bool

	// Pending is true when the workload is still progressing but has
	// not yet failed (image pulling, pod creating, etc.).
	Pending bool

	// Reason is the machine-readable reason for a Failed/Pending
	// status, propagated into the eventual WorkloadFailed action or
	// EmitEvent effect.
	Reason string

	// Message is the human-readable message for a Failed/Pending
	// status.
	Message string
}

// ClusterStats holds cross-AM aggregate metrics the Runner computes
// from the informer cache before invoking Engine.Step. Handlers that
// enforce cluster-wide constraints read from here instead of doing a
// List themselves.
type ClusterStats struct {
	// DownloadingCount is the number of AMs currently in the
	// Downloading state cluster-wide. Used by PendingTickHandler to
	// enforce the "at most one concurrent Downloading" constraint
	// that the legacy pkg/appstate/pending_app.go implements with a
	// process-wide mutex.
	DownloadingCount int

	// NextDownloading is the name of the Pending AM that should
	// occupy the download slot once it is free: the oldest
	// pending.queuedAt (FIFO), with CreationTimestamp then name as
	// fallbacks. Empty when no AM is Pending. OnPendingTick must
	// not advance unless this equals World.AM.Name (or is empty,
	// for hand-crafted unit Worlds that only set DownloadingCount).
	NextDownloading string
}
