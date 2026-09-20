package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/pkg/cluster/clusterop"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/commands"
)

// withReleaseOlaresID replaces what /etc/olares/release says about the owner.
func withReleaseOlaresID(t *testing.T, id string, err error) {
	t.Helper()
	prev := olaresIDFromRelease
	olaresIDFromRelease = func() (string, error) { return id, err }
	t.Cleanup(func() { olaresIDFromRelease = prev })
}

// asInstalledOlares points the install lock at a file that exists, which is
// what state.IsTerminusInstalled reads.
func asInstalledOlares(t *testing.T, installed bool) {
	t.Helper()
	lock := filepath.Join(t.TempDir(), ".installed")
	if installed {
		if err := os.WriteFile(lock, nil, 0o600); err != nil {
			t.Fatalf("write install lock: %v", err)
		}
	}
	prev := commands.INSTALL_LOCK
	commands.INSTALL_LOCK = lock
	t.Cleanup(func() { commands.INSTALL_LOCK = prev })
}

// asJoinedWorker is a compute node that joined an existing cluster. Its
// /etc/olares/release has no OLARES_NAME: that file is written by the installer
// on the machine that was activated, and a node added later never gets one. The
// Olares ID it does have comes from the cluster, through the state snapshot.
func asJoinedWorker(t *testing.T, olaresID string) {
	t.Helper()
	asNode(t, "worker-1", inventory.RoleWorker, nil)
	var name *string
	if olaresID != "" {
		name = &olaresID
	}
	withCurrentState(t, clistate.State{
		TerminusState: clistate.TerminusRunning,
		TerminusName:  name,
	}, time.Now())
	asInstalledOlares(t, true)
	withReleaseOlaresID(t, "", nil)
}

// The master reaches this endpoint on every compute node of a cluster it is
// powering. A worker that refused the owner because its release file names
// nobody would make every cluster power operation fail at the first node.
func TestOwnerGateAcceptsTheOwnerOnAJoinedWorker(t *testing.T) {
	r := withLocalPower(t, &powerRecorder{})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asJoinedWorker(t, testOwner)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 on a worker whose release file has no owner: %s", resp.StatusCode, body)
	}
	if got := r.seen(); len(got) != 1 {
		t.Errorf("powered %v, want one reboot", got)
	}
}

// The fallback names the owner; it does not stop checking who signed.
func TestOwnerGateRefusesAnotherIdentityOnAJoinedWorker(t *testing.T) {
	hostMustNotBePowered(t)
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asJoinedWorker(t, "carol@olares.com")

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 for a signature by somebody else: %s", resp.StatusCode, body)
	}
}

// Neither source names an owner, and Olares is installed: there is somebody to
// be the owner of and this node cannot tell whether the caller is them. The one
// answer that must not happen here is to let the request through.
func TestOwnerGateRefusesWhenNobodyCanBeNamedAsTheOwner(t *testing.T) {
	hostMustNotBePowered(t)
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asJoinedWorker(t, "")

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 when no owner can be named: %s", resp.StatusCode, body)
	}
}

// Before Olares is installed there is no owner to be, and the install-time
// routes are reached by whoever is setting the machine up. That predates
// cluster operations and stays as it is.
func TestOwnerGateStillAdmitsAnUninstalledMachine(t *testing.T) {
	r := withLocalPower(t, &powerRecorder{})
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asNode(t, "worker-1", inventory.RoleWorker, nil)
	withCurrentState(t, clistate.State{TerminusState: clistate.NotInstalled}, time.Now())
	asInstalledOlares(t, false)
	withReleaseOlaresID(t, "", nil)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want the install-time path unchanged: %s", resp.StatusCode, body)
	}
	if got := r.seen(); len(got) != 1 {
		t.Errorf("powered %v", got)
	}
}

// The release file is still the authority where it has an answer. A snapshot
// that disagrees with it must not widen who counts as the owner.
func TestOwnerGatePrefersTheReleaseFileWhereItNamesAnOwner(t *testing.T) {
	hostMustNotBePowered(t)
	asAuthorizedUser(t)
	asOwnerSignature(t)
	asNode(t, "master-1", inventory.RoleMaster, nil)
	snapshotOwner := testOwner
	withCurrentState(t, clistate.State{
		TerminusState: clistate.TerminusRunning,
		TerminusName:  &snapshotOwner,
	}, time.Now())
	asInstalledOlares(t, true)
	withReleaseOlaresID(t, "dave@olares.com", nil)

	resp, body := callRegisteredMethod(t, http.MethodPost, "/command/power-node",
		`{"type":"reboot","operationId":"op-1","requestId":"client-1"}`,
		signedFor(t, clusterop.TypeReboot, "client-1"))

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want the release file to keep deciding: %s", resp.StatusCode, body)
	}
}
