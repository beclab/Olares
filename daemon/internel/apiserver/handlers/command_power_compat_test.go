package handlers

import (
	"net/http"
	"testing"
)

// The single-node power commands predate cluster operations and keep working
// unchanged: LarePass and user-service call them today, and a cluster of one
// is still the common case.
func TestSingleNodePowerCommandsStillRequireASignature(t *testing.T) {
	for _, path := range []string{"/command/reboot", "/command/shutdown"} {
		t.Run(path, func(t *testing.T) {
			asAuthorizedUser(t)
			asMaster(t)

			resp, body := callRegisteredMethod(t, http.MethodPost, path, `{}`, authHeaders())

			if resp.StatusCode != http.StatusForbidden {
				t.Fatalf("status = %d, want 403 for a token without a signature: %s", resp.StatusCode, body)
			}
		})
	}
}
