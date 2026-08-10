package clusterop

import (
	"errors"

	"k8s.io/klog/v2"
)

// The text a caller is given for each stable code.
//
// None of it comes from the error that occurred. What fails in this package
// fails at a transport: the messages carry node addresses, the port olaresd
// listens on, the subject of a certificate that did not verify, the path of a
// local binary. An Operation is written to the daemon's state directory and
// returned to whoever asked about the operation, so neither place is somewhere
// that detail belongs. It goes to this node's log instead, where the operator
// who can act on it already has the machine.
//
// The codes are the contract; this text may be reworded.
var reasons = map[string]string{
	CodeInventoryUnavailable: "the cluster node directory could not be read",
	CodeNodeUnreachable:      "this node did not answer",
	CodeDispatchFailed:       "this node did not accept the power command",
	CodePowerUnsupported:     "this node cannot perform this operation",
	CodeHostPowerFailed:      "this node could not be powered",
	CodeNodeUnaddressable:    "node has no internal address",

	// A node-scope operation settles on whatever refused its one node, so
	// every per-node refusal is also something a whole operation can end on.
	// The sentences match the ones the per-node records carry.
	CodeNodeNotReady:      "the cluster does not consider this node ready",
	CodeBootIDUnavailable: "the cluster does not report which boot this node is on",
	CodeNodeDidNotGoDown:  "this node was still up after the power command",
	CodeRestartTimeout:    "this node did not come back ready on a new boot in time",

	// The codes a whole operation settles on. A module reports one of these
	// through its Outcome, and the reviewed sentence is all the record keeps
	// of what the module itself said.
	// The same sentence the node-local power endpoint refuses with, so an
	// operation nothing here can carry out reads the same wherever it is
	// refused.
	CodeUnsupportedOperation: "this daemon does not perform that operation",

	CodeModuleFailed: "the operation stopped without reporting what it did",

	CodeUnsupportedTopology:    "this daemon powers a cluster with exactly one control node",
	CodeSelfUnresolved:         "this daemon could not identify which node it is running on",
	CodeNodeIdentityUnknown:    "the node directory could not identify this node",
	CodePrecheckFailed:         "one or more nodes cannot perform this operation",
	CodeWorkerCommandFailed:    "one or more nodes did not accept the power command",
	CodeWorkerRestartFailed:    "one or more nodes did not come back",
	CodeStatePersistenceFailed: "the operation stopped because its state could not be recorded",
}

// settledWith is how a power operation inside this package reports an
// outcome whose sentence is already reviewed. The sentence is kept on the
// record exactly as written, which is what lets a stage and the operation it
// belongs to say the same thing: both are given this same text, rather than
// one keeping it and the other being handed a summary chosen by code.
//
// It is not a way around safeReason. Text only reaches a record through here
// if it was written in this package, next to the condition it describes and
// reviewed with it. Anything a module supplies from outside is Error, which
// is logged and never persisted.
func settledWith(status Status, code, reason string) Outcome {
	return Outcome{Status: status, Code: code, reason: reason}
}

// failedWith is settledWith for the common case: the operation stops here.
func failedWith(code, reason string) Outcome {
	return settledWith(StatusFailed, code, reason)
}

// persistedReason is the text the record keeps for this outcome. A reviewed
// sentence is kept as written; anything else is replaced by the reviewed
// sentence for the code. A blank code keeps nothing at all, so neither a
// module's detail nor a sentence disconnected from any stable code can be
// left on the record on its own.
func (o Outcome) persistedReason() string {
	if o.Code == "" {
		return ""
	}
	if o.reason != "" {
		if o.Error != "" {
			// Not a leak: reason is already the reviewed sentence this
			// settles on, and Error never overrides or blends with it. It
			// is logged only because a caller inside this package setting
			// both at once is unusual enough to be worth a look.
			klog.Warningf("clusterop: outcome %s carried both a reviewed reason and detail: %s", o.Code, o.Error)
		}
		return o.reason
	}
	return safeReason(o.Code, o.Error)
}

// reasonFor is the message that accompanies a code on the wire and on disk.
func reasonFor(code string) string {
	if r, ok := reasons[code]; ok {
		return r
	}
	return "this operation could not be completed"
}

// safeReason is what a checked Runtime mutation persists for a module-
// supplied code and detail pair. The code chooses a fixed, reviewed message
// through reasonFor; the module's own detail — which may carry a node
// address, a token, or any other text this package cannot vouch for — goes
// only to the log, never to the Operation record or the file it is saved to.
// A blank code always persists an empty message, so a module bug that sends
// detail text without a code cannot make the record carry unreviewed text
// disconnected from any stable code either.
func safeReason(code, detail string) string {
	if code == "" {
		return ""
	}
	if detail != "" {
		klog.Warningf("clusterop: module reported %s: %s", code, detail)
	}
	return reasonFor(code)
}

// suppress records the detail of an internal failure in the log and returns the
// fixed text that stands in for it everywhere a caller can read.
//
// where says which part of the operation was doing what, because the code alone
// does not: node_unreachable is reported both by the precheck reading a node's
// status and by the dispatch handing it a command.
func suppress(code, where string, err error) string {
	if err != nil {
		klog.Warningf("clusterop: %s: %s: %v", where, code, err)
	}
	return reasonFor(code)
}

// powerReason is the same for this machine's own execution point, which already
// separates its fixed message from its detail. Anything that is not a
// PowerError is treated as detail in full.
func powerReason(fallbackCode, where string, err error) (code, reason string) {
	var pe *PowerError
	if errors.As(err, &pe) {
		if pe.Err != nil {
			klog.Warningf("clusterop: %s: %s: %v", where, pe.Code, pe.Err)
		}
		return pe.Code, pe.Message
	}
	return fallbackCode, suppress(fallbackCode, where, err)
}
