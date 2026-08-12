package clusterop

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are the shapes a real failure here arrives in. Each one names
// something about the inside of the cluster that the caller asked nothing
// about: which port olaresd listens on, which certificate a node presented,
// which apiserver this daemon talks to, where the shutdown binary lives.
var internalErrors = map[string]error{
	"dial": errors.New(
		"Post \"http://10.0.0.2:18088/command/power-node\": dial tcp 10.0.0.2:18088: connect: connection refused"),
	"x509": errors.New(
		"x509: certificate signed by unknown authority (possibly because of \"crypto/rsa: verification error\" " +
			"while trying to verify candidate authority certificate \"olares-ca\")"),
	"apiserver": errors.New(
		"Get \"https://10.0.0.1:6443/api/v1/nodes\": tls: failed to verify certificate for 10.0.0.1"),
	"exec": errors.New("fork/exec /usr/sbin/shutdown: permission denied"),
}

// leakedFrom names every fragment of an internal error that reached a place a
// caller can read: the JSON of the operation, or the record on disk.
func leakedFrom(t *testing.T, op Operation, dir string) []string {
	t.Helper()

	encoded, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("marshal operation: %v", err)
	}
	surfaces := map[string]string{"json": string(encoded)}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read the state directory: %v", err)
	}
	for _, e := range entries {
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		surfaces["disk:"+e.Name()] = string(raw)
	}

	// Fragments rather than whole messages: a truncated or wrapped internal
	// error leaks exactly as much as the part that survived.
	fragments := []string{
		"18088", "6443", "dial tcp", "x509", "certificate", "connection refused",
		"olares-ca", "crypto/rsa", "/usr/sbin/shutdown", "permission denied",
		"tls:", "10.0.0.1", "10.0.0.2",
	}

	var found []string
	for where, text := range surfaces {
		lower := strings.ToLower(text)
		for _, f := range fragments {
			if strings.Contains(lower, strings.ToLower(f)) {
				found = append(found, where+" contains "+f)
			}
		}
	}
	return found
}

// An operation record is written to the state directory and read back by
// anything that can read the disk, and its JSON goes to whoever asked for the
// operation. Neither is a place for the address of a node or the reason a
// certificate did not verify.
func TestOperationNeverCarriesAnInternalErrorMessage(t *testing.T) {
	for _, tc := range []struct {
		name   string
		break_ func(*cluster)
		opType Type
		code   string
	}{
		{
			name:   "the node directory could not be read",
			break_: func(c *cluster) { c.inventoryErr = internalErrors["apiserver"] },
			opType: TypeReboot,
			code:   CodeInventoryUnavailable,
		},
		{
			name:   "a node could not be inspected",
			break_: func(c *cluster) { c.inspectErr["worker-1"] = internalErrors["dial"] },
			opType: TypeReboot,
			code:   CodePrecheckFailed,
		},
		{
			name:   "a node presented a certificate that did not verify",
			break_: func(c *cluster) { c.inspectErr["worker-1"] = internalErrors["x509"] },
			opType: TypeShutdown,
			code:   CodePrecheckFailed,
		},
		{
			name:   "the cluster could not be observed for the baseline",
			break_: func(c *cluster) { c.observeErr = internalErrors["apiserver"] },
			opType: TypeReboot,
			code:   CodeInventoryUnavailable,
		},
		{
			name:   "a node did not accept the command",
			break_: func(c *cluster) { c.dispatchErr["worker-1"] = internalErrors["dial"] },
			opType: TypeShutdown,
			code:   CodeWorkerCommandFailed,
		},
		{
			name:   "this machine could not be powered",
			break_: func(c *cluster) { c.powerSelfErr = internalErrors["exec"] },
			opType: TypeShutdown,
			code:   CodeHostPowerFailed,
		},
		{
			name:   "this machine's execution point refused",
			break_: func(c *cluster) { c.localPowerErr = internalErrors["exec"] },
			opType: TypeShutdown,
			code:   CodePrecheckFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
			tc.break_(c)
			m, dir := newManager(t, c)

			op := awaitTerminal(t, m, createOp(t, m, tc.opType, "client-1").ID)

			if op.Code != tc.code {
				t.Errorf("operation code = %q, want the stable %q", op.Code, tc.code)
			}
			for _, leak := range leakedFrom(t, op, dir) {
				t.Errorf("internal error reached the caller: %s", leak)
			}
		})
	}
}

// A failure still has to be reported. Suppressing the detail is only correct if
// what remains says which node, what stage, and which stable code — otherwise
// the fix has traded a leak for an operation nobody can act on.
func TestASuppressedErrorStillReportsWhatFailedAndWhere(t *testing.T) {
	c := newCluster(master("master-1", "10.0.0.1"), worker("worker-1", "10.0.0.2"))
	c.inspectErr["worker-1"] = internalErrors["dial"]
	m, _ := newManager(t, c)

	op := awaitTerminal(t, m, createOp(t, m, TypeReboot, "client-1").ID)

	got := nodeResult(t, op, "worker-1")
	if got.Code != CodeNodeUnreachable {
		t.Errorf("node code = %q, want %s", got.Code, CodeNodeUnreachable)
	}
	if strings.TrimSpace(got.Error) == "" {
		t.Error("the node result says nothing about why it failed")
	}
	if s := step(t, op, StepPrecheck); s.Code != CodePrecheckFailed || strings.TrimSpace(s.Error) == "" {
		t.Errorf("the precheck step does not report itself: %+v", s)
	}
}
