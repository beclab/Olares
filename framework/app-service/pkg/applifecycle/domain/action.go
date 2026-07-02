package domain

// ActionKind is the runtime tag used by TransitionTable to dispatch on
// action type. Kept as a string so the tag is stable across binary rebuilds
// and can appear in log lines / metrics.
type ActionKind string

const (
	// KindTick represents a reconciliation tick without a user operation.
	// It is the default action injected by the Runner when a Reconcile is
	// triggered by informer watch events, requeues, or the timeout ticker.
	KindTick ActionKind = "Tick"

	// KindWorkloadReady is fed back into the state machine by the Runner
	// when a long-running CallWorkload succeeds (helm install ok, image
	// pull ok, workload deployment reports Ready, etc.).
	KindWorkloadReady ActionKind = "WorkloadReady"

	// KindWorkloadFailed is the failure counterpart to KindWorkloadReady.
	KindWorkloadFailed ActionKind = "WorkloadFailed"

	// KindTimeoutFired is fed back when the state's TTL elapses without a
	// terminal transition.
	KindTimeoutFired ActionKind = "TimeoutFired"

	// kindOpPrefix is the prefix used to synthesize per-Op ActionKind
	// values. See TriggerOp.Kind for the full format.
	kindOpPrefix ActionKind = "Op:"
)

// OpKind returns the ActionKind that identifies a user-triggered operation
// of type op. TransitionTable.Add stores transitions keyed by this string so
// that (Pending, TriggerOp{InstallOp}) and (Pending, TriggerOp{CancelOp})
// resolve to different edges.
func OpKind(op OpType) ActionKind {
	return kindOpPrefix + ActionKind(op)
}

// Action is the sum type for every input event the state machine can
// receive. The set is closed: introducing a new action requires adding a
// new type in this file and updating the TransitionTable and any handlers
// that need to react to it.
//
// Actions carry ONLY the information required to identify what happened;
// they do NOT carry data derived from the AM or the cluster. Handlers read
// mutable data from the World snapshot, never from the action.
type Action interface {
	// Kind returns the ActionKind used by the dispatch table. Must be
	// pure: the return value depends only on the receiver's type (and its
	// Op field, for TriggerOp).
	Kind() ActionKind

	// isAction is a private marker keeping the set closed. External
	// packages cannot introduce new Action implementations.
	isAction()
}

// Tick is a reconciliation tick. It is the default action the Runner feeds
// into Engine.Step when a Reconcile fires from an informer watch, a
// controller-runtime requeue, or the periodic timeout ticker.
type Tick struct{}

// Kind implements Action.
func (Tick) Kind() ActionKind { return KindTick }
func (Tick) isAction()        {}

// TriggerOp represents a user-initiated operation received through the
// apiserver (or a controller acting as a proxy for a user, e.g.
// pod_abnormal_suspend_app_controller flipping Running to Stopping).
type TriggerOp struct {
	Op OpType
}

// Kind implements Action.
func (t TriggerOp) Kind() ActionKind { return OpKind(t.Op) }
func (TriggerOp) isAction()          {}

// WorkloadReady is the successful completion callback for a CallWorkload
// effect. The Runner fabricates and re-enqueues this action when the
// external workload (helm, image pull, namespace deletion, etc.) reaches
// its success terminal.
type WorkloadReady struct{}

// Kind implements Action.
func (WorkloadReady) Kind() ActionKind { return KindWorkloadReady }
func (WorkloadReady) isAction()        {}

// WorkloadFailed is the failure callback for a CallWorkload effect.
type WorkloadFailed struct {
	// Reason is a short, machine-readable code (matches the string that
	// currently ends up in ApplicationManagerStatus.Reason).
	Reason string

	// Message is a human-readable description (matches the string that
	// currently ends up in ApplicationManagerStatus.Message).
	Message string

	// LandedState, when non-empty, is the state the workload adapter
	// determined the AM must end up in, overriding the handler's
	// default failure target. The Runner copies it from
	// ports.WorkloadResult.LandedState.
	//
	// The motivating case is helm install failing because pods are
	// unschedulable: legacy installing_app.go routes those to
	// Stopping (the release stays deployed at replicas=0 so the user
	// can retry) instead of InstallFailed (which tears the release
	// down). The adapter is the only layer that can tell the two
	// apart, since the distinction lives in the errcode sentinel
	// wrapped inside the helm error.
	LandedState State

	// StopAll mirrors ports.WorkloadResult.StopAll: the pending pods
	// belong to a shared server, so the AM needs the stop-all flag
	// stamped before it lands in LandedState. Only meaningful
	// alongside LandedState == Stopping.
	StopAll bool
}

// Kind implements Action.
func (WorkloadFailed) Kind() ActionKind { return KindWorkloadFailed }
func (WorkloadFailed) isAction()        {}

// TimeoutFired is fed into the state machine when a state's TTL elapses
// without a terminal transition. Handlers decide whether to auto-cancel,
// move to a *Failed state, or extend the deadline.
type TimeoutFired struct{}

// Kind implements Action.
func (TimeoutFired) Kind() ActionKind { return KindTimeoutFired }
func (TimeoutFired) isAction()        {}
