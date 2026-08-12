package clusterop

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

func TestParseTypeAcceptsOnlyRegisteredOperations(t *testing.T) {
	for _, in := range []string{"reboot", "shutdown", "set-ssh-password"} {
		got, err := ParseType(in)
		if err != nil {
			t.Errorf("ParseType(%q) errored: %v", in, err)
		}
		if string(got) != in {
			t.Errorf("ParseType(%q) = %q", in, got)
		}
	}
	for _, in := range []string{"", "Reboot", "poweroff", "restart", "uninstall"} {
		if _, err := ParseType(in); err == nil {
			t.Errorf("ParseType(%q) accepted an operation this daemon cannot perform", in)
		}
	}
}

// The six statuses are a wire contract shared with user-service and TermiPass.
// A client that receives one it has never heard of shows it verbatim, so
// renaming one silently is worse than adding one.
func TestEveryDeclaredStatusIsOnTheWire(t *testing.T) {
	want := map[Status]bool{
		StatusPending:         false,
		StatusRunning:         false,
		StatusSucceeded:       true,
		StatusPartiallyFailed: true,
		StatusFailed:          true,
		StatusCommandIssued:   true,
	}
	for status, terminal := range want {
		if status == "" {
			t.Fatalf("a status constant is empty")
		}
		if status.Terminal() != terminal {
			t.Errorf("%q terminal = %v, want %v", status, status.Terminal(), terminal)
		}
	}
	if len(want) != 6 {
		t.Fatalf("the status set changed; update the clients before the daemon")
	}
}

// A power operation that reaches the control node can never be verified from
// the inside: the machine that would confirm it is the one going down. Saying
// "succeeded" there would be a claim nothing observed.
func TestCommandIssuedIsTerminalButNotSuccess(t *testing.T) {
	if !StatusCommandIssued.Terminal() {
		t.Error("command_issued must be terminal, or a second operation is blocked forever")
	}
	if StatusCommandIssued == StatusSucceeded {
		t.Error("command_issued must stay distinct from succeeded")
	}
}

func TestPhaseFollowsTheActiveOperation(t *testing.T) {
	for _, tc := range []struct {
		name string
		op   *Operation
		want nodestatus.Phase
		ok   bool
	}{
		{name: "no operation", op: nil},
		{name: "reboot running", op: &Operation{Type: TypeReboot, Status: StatusRunning}, want: nodestatus.PhaseRestarting, ok: true},
		{name: "shutdown running", op: &Operation{Type: TypeShutdown, Status: StatusRunning}, want: nodestatus.PhaseShuttingDown, ok: true},
		{name: "reboot pending", op: &Operation{Type: TypeReboot, Status: StatusPending}, want: nodestatus.PhaseRestarting, ok: true},
		{name: "finished reboot", op: &Operation{Type: TypeReboot, Status: StatusCommandIssued}},
		{name: "failed shutdown", op: &Operation{Type: TypeShutdown, Status: StatusFailed}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := PhaseFor(tc.op)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if ok && got != tc.want {
				t.Errorf("phase = %q, want %q", got, tc.want)
			}
		})
	}
}

// Health is measured; phase is what the cluster is doing. An operation in
// flight must not be allowed to answer the other question.
func TestPhaseCarriesNoHealthOpinion(t *testing.T) {
	_, ok := PhaseFor(&Operation{Type: TypeReboot, Status: StatusRunning})
	if !ok {
		t.Fatal("an active operation should yield a phase")
	}
	// The signature is the guarantee: nothing here can return a Health.
	var _ func(*Operation) (nodestatus.Phase, bool) = PhaseFor
}

func TestOperationWireFieldNames(t *testing.T) {
	at := time.Unix(1700000000, 0).UTC()
	op := Operation{
		ID:        "op-1",
		Type:      TypeReboot,
		RequestID: "client-1",
		Owner:     "alice@olares.com",
		Status:    StatusRunning,
		CreatedAt: at,
		UpdatedAt: at,
		Steps:     []Step{{Name: StepPrecheck, Status: StepSucceeded, StartedAt: &at, FinishedAt: &at}},
		Nodes: []NodeResult{{
			NodeName: "worker-1",
			Role:     "worker",
			Status:   NodeCommandIssued,
			Code:     CodeNodeUnreachable,
			Error:    "no route to host",
		}},
	}

	raw, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"id", "type", "requestId", "owner", "status", "createdAt", "updatedAt", "steps", "nodes"} {
		if _, ok := got[key]; !ok {
			t.Errorf("field %q missing: %s", key, raw)
		}
	}

	steps, _ := got["steps"].([]any)
	if len(steps) != 1 {
		t.Fatalf("steps missing: %s", raw)
	}
	step, _ := steps[0].(map[string]any)
	for _, key := range []string{"name", "status", "startedAt", "finishedAt"} {
		if _, ok := step[key]; !ok {
			t.Errorf("step field %q missing: %s", key, raw)
		}
	}

	nodes, _ := got["nodes"].([]any)
	node, _ := nodes[0].(map[string]any)
	for _, key := range []string{"nodeName", "role", "status", "code", "error"} {
		if _, ok := node[key]; !ok {
			t.Errorf("node field %q missing: %s", key, raw)
		}
	}
}

// The credentials that authorize a power operation are held for the length of
// one run. Anything reachable from the operation record is written to disk and
// read back by whoever can read the state directory.
func TestOperationCarriesNoCredentials(t *testing.T) {
	raw, err := json.Marshal(Operation{ID: "op-1", Type: TypeShutdown})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, forbidden := range []string{"token", "signature", "jws", "auth", "authorization", "credentials"} {
		if _, ok := fields[forbidden]; ok {
			t.Errorf("operation record carries %q: %s", forbidden, raw)
		}
	}
}

func TestCloneDoesNotShareStepsOrNodes(t *testing.T) {
	op := Operation{
		ID:    "op-1",
		Steps: []Step{{Name: StepPrecheck, Status: StepRunning}},
		Nodes: []NodeResult{{NodeName: "worker-1", Status: NodePending}},
	}

	clone := op.Clone()
	clone.Steps[0].Status = StepFailed
	clone.Nodes[0].Status = NodeFailed

	if op.Steps[0].Status != StepRunning || op.Nodes[0].Status != NodePending {
		t.Errorf("a handed-out copy wrote back into the stored operation: %+v", op)
	}
}

// TestCloneDeepCopiesTimestamps is a Minor fix: Clone must copy the value
// behind every *time.Time, not just the pointer, or a caller that writes
// through a timestamp it was handed reaches back into the stored record.
func TestCloneDeepCopiesTimestamps(t *testing.T) {
	at := time.Unix(1700000000, 0).UTC()
	op := Operation{
		ID:         "op-1",
		StartedAt:  &at,
		FinishedAt: &at,
		Steps:      []Step{{Name: StepPrecheck, StartedAt: &at, FinishedAt: &at}},
		Nodes:      []NodeResult{{NodeName: "worker-1", StartedAt: &at, FinishedAt: &at}},
	}

	clone := op.Clone()
	later := at.Add(time.Hour)
	*clone.StartedAt = later
	*clone.FinishedAt = later
	*clone.Steps[0].StartedAt = later
	*clone.Steps[0].FinishedAt = later
	*clone.Nodes[0].StartedAt = later
	*clone.Nodes[0].FinishedAt = later

	if !op.StartedAt.Equal(at) || !op.FinishedAt.Equal(at) {
		t.Errorf("mutating the clone's own timestamps changed the original operation: %+v", op)
	}
	if !op.Steps[0].StartedAt.Equal(at) || !op.Steps[0].FinishedAt.Equal(at) {
		t.Errorf("mutating a cloned step's timestamps changed the original: %+v", op.Steps[0])
	}
	if !op.Nodes[0].StartedAt.Equal(at) || !op.Nodes[0].FinishedAt.Equal(at) {
		t.Errorf("mutating a cloned node's timestamps changed the original: %+v", op.Nodes[0])
	}
}
