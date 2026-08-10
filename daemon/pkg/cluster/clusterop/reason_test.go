package clusterop

import (
	"strings"
	"testing"
)

// Every code a module settles an operation on reaches the record through
// safeReason, which replaces the module's own text with a reviewed message
// keyed by that code. A code with no entry falls back to a sentence that
// says nothing at all, so an operation that ends on one leaves the caller
// with a stable code and no explanation of it.
func TestEveryOutcomeCodeAPowerOperationSettlesOnHasAReviewedReason(t *testing.T) {
	for _, code := range []string{
		CodeInventoryUnavailable,
		CodeUnsupportedTopology,
		CodeSelfUnresolved,
		CodeNodeIdentityUnknown,
		CodePrecheckFailed,
		CodePowerUnsupported,
		CodeWorkerCommandFailed,
		CodeWorkerRestartFailed,
		CodeHostPowerFailed,
		CodeStatePersistenceFailed,
		CodeUnsupportedOperation,

		// A node-scope operation settles on whatever refused its one node,
		// so every per-node refusal is also an operation outcome.
		CodeNodeNotReady,
		CodeNodeUnaddressable,
		CodeBootIDUnavailable,
		CodeNodeUnreachable,
		CodeDispatchFailed,
		CodeNodeDidNotGoDown,
		CodeRestartTimeout,
	} {
		reason, ok := reasons[code]
		if !ok {
			t.Errorf("code %q has no reviewed reason, so an operation that ends on it explains nothing", code)
			continue
		}
		if strings.TrimSpace(reason) == "" {
			t.Errorf("code %q has an empty reason", code)
		}
	}
}

// The reviewed text stands in for an internal error precisely because it
// cannot carry one. A reason that named an address or a local path would
// put back exactly what suppressing the error took out.
func TestNoReviewedReasonNamesSomethingInternal(t *testing.T) {
	for code, reason := range reasons {
		lower := strings.ToLower(reason)
		for _, fragment := range []string{"10.0.0", "dial", "x509", "/usr/", "tcp", "certificate", "18088", "6443"} {
			if strings.Contains(lower, fragment) {
				t.Errorf("the reason for %q names something internal (%q): %s", code, fragment, reason)
			}
		}
	}
}
