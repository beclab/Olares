package handlers

import (
	"net/http"
	"testing"
)

// Nothing in a test binary may be able to power the machine running it. The
// module set a node acts through is installed by main, not by package
// initialization, so a route reached by a test that forgot to fake it
// refuses rather than reboots.
func TestPowerExecutionIsNotInstalledUntilTheDaemonInstallsIt(t *testing.T) {
	if nodeOperations != nil {
		t.Fatal("the real power modules are reachable from a test binary")
	}
}

func TestPowerNodeRefusesUntilPowerExecutionIsInstalled(t *testing.T) {
	prev := nodeOperations
	nodeOperations = nil
	t.Cleanup(func() { nodeOperations = prev })
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asWorker(t)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		signedFor(t, "reboot", "client-1"))

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503: %s", resp.StatusCode, body)
	}
}

// The orchestrator is installed the same way, with the side effects it needs
// handed to it. A test binary that never installs one cannot start an
// operation, let alone one that reaches a machine.
func TestClusterOperationsAreNotInstalledUntilTheDaemonInstallsThem(t *testing.T) {
	if clusterOperations != nil {
		t.Fatal("an orchestrator wired to the real cluster is reachable from a test binary")
	}
}
