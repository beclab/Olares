package clusterop

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

// recordingStore keeps every record a manager writes, in order, so a test can
// ask what a reader of the state directory could have seen — and can make one
// write fail.
type recordingStore struct {
	*Store

	mu      sync.Mutex
	saves   []Operation
	failAll bool
}

func (s *recordingStore) Save(op Operation) error {
	s.mu.Lock()
	fail := s.failAll
	if !fail {
		s.saves = append(s.saves, op.Clone())
	}
	s.mu.Unlock()
	if fail {
		return errors.New("the state directory is unwritable")
	}
	return s.Store.Save(op)
}

func (s *recordingStore) written() []Operation {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Operation(nil), s.saves...)
}

func (s *recordingStore) refuseWrites() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failAll = true
}

// interruptedReboot is the record a daemon leaves behind when the machine it
// was running on was told to reboot: the control node's power command went
// out, and the process that would have confirmed it went down with the
// machine. graceOffset places the deadline relative to the stored moment —
// negative for a record whose grace has long since run out.
func interruptedReboot(t *testing.T, store OperationStore, at time.Time, graceOffset time.Duration) Operation {
	t.Helper()
	finished := at
	op := Operation{
		ID:        "op-interrupted",
		Type:      TypeReboot,
		RequestID: "client-interrupted",
		Scope:     ScopeNode,
		Target:    "master-1",
		ClusterID: "cluster-1",
		Owner:     "alice@olares.com",
		Status:    StatusCommandIssued,
		CreatedAt: at,
		UpdatedAt: at,
		// The record already says when its command went out. Confirming it
		// later has to move this to the moment the outcome was established.
		FinishedAt:         &finished,
		CommandIssuedUntil: at.Add(graceOffset),
		HostBootID:         "host-boot-1",
		Steps: []Step{{
			Name: StepMasterCommand, Status: StepCommandIssued, StartedAt: &at, FinishedAt: &finished,
		}},
		Nodes: []NodeResult{{
			NodeName: "master-1", Role: inventory.RoleMaster, Status: NodeCommandIssued,
			StartedAt: &at, FinishedAt: &finished,
		}},
	}
	if err := store.Save(op); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return op
}

// rebootedControlNode is the cluster as it looks once the machine came back:
// on a different boot, and Ready on it.
func rebootedControlNode(t *testing.T, dir string) (*cluster, *recordingStore, Deps) {
	t.Helper()
	c := newCluster(master("master-1", "10.0.0.1"))
	c.hostBootID = "host-boot-2"
	c.obs["master-1"] = inventory.Observation{Ready: true, BootID: "host-boot-2"}

	deps := c.deps(t, dir)
	store := &recordingStore{Store: deps.Store.(*Store)}
	deps.Store = store
	return c, store, deps
}

// The grace deadline bounds how long the cluster is held for an operation. It
// is not a statute of limitations on the evidence: a control node that comes
// back on a boot other than the one it was told to leave rebooted, whether
// the daemon that finds out is the next one or the one after a long outage.
// Leaving that record saying "command issued" would report a reboot this
// daemon can prove finished as one it never confirmed.
func TestARebootIsConfirmedAfterItsGraceDeadlineHasLongPassed(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	_, _, deps := rebootedControlNode(t, dir)
	stored := interruptedReboot(t, deps.Store, deps.Now(), -time.Hour)

	m, err := NewManager(deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	got := awaitOperationStatus(t, m, stored.ID, StatusSucceeded)
	if got.Status != StatusSucceeded {
		t.Fatalf("status = %q, want succeeded once the machine is back on a new boot", got.Status)
	}
	if node := nodeResult(t, got, "master-1"); node.Status != NodeRestarted {
		t.Errorf("control node = %q, want restarted", node.Status)
	}
	if step := stepResult(t, got, StepMasterCommand); step.Status != StepSucceeded {
		t.Errorf("%s = %q, want succeeded", StepMasterCommand, step.Status)
	}
	if got.FinishedAt == nil || !got.FinishedAt.After(*stored.FinishedAt) {
		t.Errorf("FinishedAt = %v, want the moment the reboot was confirmed (after %v)",
			got.FinishedAt, stored.FinishedAt)
	}
	if !got.CommandIssuedUntil.IsZero() {
		t.Errorf("CommandIssuedUntil = %v, want it cleared once the operation settled", got.CommandIssuedUntil)
	}
}

// Recovery is allowed to settle an outstanding command. It is not a way to
// reopen an operation that already reported what it did, and an ordinary run
// gets no new powers from its existence.
func TestRecoveryCannotReopenAnOperationThatAlreadyReported(t *testing.T) {
	for _, status := range []Status{StatusSucceeded, StatusFailed, StatusPartiallyFailed} {
		t.Run(string(status), func(t *testing.T) {
			rt, _, _ := newTestRuntime(t, status)
			m, _, _ := managerOf(rt)
			recovery := newRecoveryRuntime(m, "op-runtime", context.Background())

			err := recovery.Complete(Outcome{Status: StatusSucceeded})
			if !errors.Is(err, ErrOperationTerminal) {
				t.Fatalf("recovery Complete() on a %s operation = %v, want ErrOperationTerminal", status, err)
			}
		})
	}
}

func TestARunStillCannotChangeAnOperationWhoseGraceHasExpired(t *testing.T) {
	rt, _, _ := buildCommandIssuedRuntime(t, -time.Hour)
	if err := rt.Complete(Outcome{Status: StatusSucceeded}); !errors.Is(err, ErrOperationTerminal) {
		t.Fatalf("Complete() from a run after the grace window = %v, want ErrOperationTerminal", err)
	}
}

// --- I1: the confirmation is one write, or it is nothing. ---

// A reader of the state directory sees whatever was last written. If the
// confirmation reached disk in pieces, one of those pieces would say the
// control node's power command succeeded while the operation it belongs to
// still said the command was outstanding.
func TestConfirmingARebootIsWrittenAllAtOnce(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	_, store, deps := rebootedControlNode(t, dir)
	stored := interruptedReboot(t, deps.Store, deps.Now(), -time.Hour)

	m, err := NewManager(deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	awaitOperationStatus(t, m, stored.ID, StatusSucceeded)

	var confirmations int
	for _, written := range store.written() {
		if written.ID != stored.ID {
			continue
		}
		step := stepResult(t, written, StepMasterCommand)
		node := nodeResult(t, written, "master-1")
		confirmed := step.Status == StepSucceeded || node.Status == NodeRestarted ||
			written.Status == StatusSucceeded
		if !confirmed {
			continue
		}
		confirmations++
		if step.Status != StepSucceeded || node.Status != NodeRestarted ||
			written.Status != StatusSucceeded {
			t.Errorf("a record was written half confirmed: status=%q step=%q node=%q",
				written.Status, step.Status, node.Status)
		}
	}
	if confirmations != 1 {
		t.Errorf("the confirmation was written %d times, want exactly once", confirmations)
	}
}

func TestAConfirmationThatCannotBePersistedConfirmsNothing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	_, store, deps := rebootedControlNode(t, dir)
	stored := interruptedReboot(t, deps.Store, deps.Now(), -time.Hour)

	m, err := NewManager(deps)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	store.refuseWrites()

	got := awaitOperationStatus(t, m, stored.ID, StatusFailed)
	if got.Status != StatusFailed || got.Code != CodeStatePersistenceFailed {
		t.Fatalf("status = %q code = %q, want failed/%s", got.Status, got.Code, CodeStatePersistenceFailed)
	}
	if step := stepResult(t, got, StepMasterCommand); step.Status == StepSucceeded {
		t.Error("the control node's power stage was confirmed by a settlement that was never written")
	}
	if node := nodeResult(t, got, "master-1"); node.Status == NodeRestarted {
		t.Error("the control node was confirmed restarted by a settlement that was never written")
	}

	onDisk, ok, err := store.Store.Load(stored.ID)
	if err != nil || !ok {
		t.Fatalf("Load: ok=%v err=%v", ok, err)
	}
	if onDisk.Status != StatusCommandIssued {
		t.Errorf("on-disk status = %q, want the record left as it was found", onDisk.Status)
	}
	if step := stepResult(t, onDisk, StepMasterCommand); step.Status != StepCommandIssued {
		t.Errorf("on-disk %s = %q, want it left as it was found", StepMasterCommand, step.Status)
	}
}
