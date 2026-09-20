package clusterop

import (
	"errors"
	"runtime/debug"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"k8s.io/klog/v2"
)

// This file is the boundary around the two questions a module is asked
// outside Run: whether it will carry out a request, and what it makes the
// cluster look like while it is happening. Both are reached from a request
// being served — Create from the master's own route, Phase from the cluster
// summary — and neither has an operation record to settle, so a module that
// panics in one of them must leave nothing behind but a log line.
//
// Both boundaries live here so there is one of each. The manager and the
// node endpoint ask the same SafeValidate, rather than each keeping its own
// idea of what a module that could not answer means; and every caller that
// wants an operation's phase goes through phaseOf.

// ErrModuleFailed is what a module that could not answer at all becomes: it
// panicked while being asked whether it will carry out a request, so there
// is no refusal to report and nothing was started.
//
// It is deliberately one fixed sentence. What the module panicked with is
// whatever it happened to be holding — a stack, a lower-level error, the
// params the caller sent — and it goes to this node's log, never to the
// caller. It is distinct from ModuleValidationError because the two are not
// the same answer: a module that refuses a request has judged it, and a
// module that could not answer has not.
var ErrModuleFailed = errors.New("the cluster operation module could not judge the request")

// SafeValidate asks a module whether it will carry out req, inside a panic
// boundary. refusal is the module's own answer; answered is false when it
// could not give one, which is the case a caller must not report as a
// refusal — the request was never judged.
//
// It is exported because the node endpoint asks the same question about the
// same request on a different machine, and the two must not drift: a module
// that cannot answer is a failure to start an operation on the master and a
// failure to carry one out on a node, and neither repeats what it panicked
// with.
func SafeValidate(module OperationModule, req CreateRequest) (refusal error, answered bool) {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("clusterop: module %s panicked while validating a request: %v\n%s",
				req.Type, r, debug.Stack())
			refusal, answered = nil, false
		}
	}()
	return module.Validate(req), true
}

// safePhase asks a module what the operation makes the cluster look like,
// inside the same kind of boundary. A module that cannot say imposes
// nothing: the caller keeps whatever phase it already had, which is exactly
// what an operation with no phase of its own already produces. Nothing is
// settled and nothing is written — describing a cluster is a read.
func safePhase(module OperationModule, op Operation) (phase nodestatus.Phase, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("clusterop: module %s panicked while reporting its phase: %v\n%s",
				op.Type, r, debug.Stack())
			phase, ok = "", false
		}
	}()
	return module.Phase(op)
}
