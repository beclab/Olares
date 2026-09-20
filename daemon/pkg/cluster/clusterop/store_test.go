package clusterop

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(filepath.Join(t.TempDir(), "operations"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func sampleOp(id string) Operation {
	at := time.Unix(1700000000, 0).UTC()
	return Operation{
		ID:        id,
		Type:      TypeReboot,
		RequestID: "client-" + id,
		Owner:     "alice@olares.com",
		Status:    StatusRunning,
		CreatedAt: at,
		UpdatedAt: at,
		Steps:     []Step{{Name: StepPrecheck, Status: StepSucceeded, StartedAt: &at, FinishedAt: &at}},
		Nodes:     []NodeResult{{NodeName: "worker-1", Role: "worker", Status: NodeCommandIssued}},
	}
}

func TestStoreRoundTrip(t *testing.T) {
	s := newStore(t)
	op := sampleOp("op-1")

	if err := s.Save(op); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, ok, err := s.Load("op-1")
	if err != nil || !ok {
		t.Fatalf("Load: ok=%v err=%v", ok, err)
	}
	if got.ID != op.ID || got.Type != op.Type || got.Owner != op.Owner || got.Status != op.Status {
		t.Errorf("record changed across the disk: %+v", got)
	}
	if len(got.Steps) != 1 || got.Steps[0].Name != StepPrecheck {
		t.Errorf("steps lost: %+v", got.Steps)
	}
	if len(got.Nodes) != 1 || got.Nodes[0].NodeName != "worker-1" {
		t.Errorf("node results lost: %+v", got.Nodes)
	}
	if !got.CreatedAt.Equal(op.CreatedAt) {
		t.Errorf("createdAt = %v, want %v", got.CreatedAt, op.CreatedAt)
	}
}

func TestStoreUnknownOperationIsNotAnError(t *testing.T) {
	s := newStore(t)

	_, ok, err := s.Load("never-existed")
	if err != nil {
		t.Fatalf("a missing operation is a normal answer, got %v", err)
	}
	if ok {
		t.Error("ok = true for an operation that was never created")
	}
}

// The record is what a caller polls after the master rebooted. It survives the
// process that wrote it or it is worth nothing.
func TestStoreSurvivesTheProcessThatWroteIt(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	first, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := first.Save(sampleOp("op-1")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := first.Save(sampleOp("op-2")); err != nil {
		t.Fatalf("Save: %v", err)
	}

	restarted, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore after restart: %v", err)
	}
	ops, err := restarted.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ops) != 2 {
		t.Fatalf("want both operations back after a restart, got %d", len(ops))
	}
	if _, ok, _ := restarted.Load("op-1"); !ok {
		t.Error("op-1 not queryable after a restart")
	}
}

// A reader that catches a half-written record answers a poll with nothing at
// all. Writing through a temp file and renaming is what makes every read see
// either the previous record or the next one.
func TestStoreNeverExposesAPartialRecord(t *testing.T) {
	s := newStore(t)
	op := sampleOp("op-1")
	if err := s.Save(op); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 300; i++ {
			grown := op
			grown.Error = strings.Repeat("x", i*64)
			if err := s.Save(grown); err != nil {
				t.Errorf("Save: %v", err)
				break
			}
		}
		close(stop)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			got, ok, err := s.Load("op-1")
			if err != nil {
				t.Errorf("a concurrent read saw a record mid-write: %v", err)
				return
			}
			if ok && got.ID != "op-1" {
				t.Errorf("read a record that was never written: %+v", got)
				return
			}
		}
	}()

	wg.Wait()
}

func TestStoreLeavesNoTemporaryFilesBehind(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := s.Save(sampleOp("op-1")); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("want exactly one file for one operation, got %d: %v", len(entries), entries)
	}
	if entries[0].Name() != "op-1.json" {
		t.Errorf("unexpected file %q left in the state directory", entries[0].Name())
	}
}

// The state directory is on the host filesystem next to the installer's own
// records. These are operations somebody's Olares performed; they are not for
// every local account to read.
func TestStoreFilesAreNotWorldReadable(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := s.Save(sampleOp("op-1")); err != nil {
		t.Fatalf("Save: %v", err)
	}

	fi, err := os.Stat(filepath.Join(dir, "op-1.json"))
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := fi.Mode().Perm(); perm&0o077 != 0 {
		t.Errorf("record mode = %v, want no group or other access", perm)
	}
	di, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat dir: %v", err)
	}
	if perm := di.Mode().Perm(); perm&0o077 != 0 {
		t.Errorf("state dir mode = %v, want no group or other access", perm)
	}
}

// One unreadable file is a lost operation, not a daemon that cannot list any.
func TestStoreSkipsAnUnreadableRecord(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := s.Save(sampleOp("op-good")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "op-bad.json"), []byte("{ truncated"), 0o600); err != nil {
		t.Fatalf("write corrupt record: %v", err)
	}

	ops, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ops) != 1 || ops[0].ID != "op-good" {
		t.Errorf("want the readable operation back, got %+v", ops)
	}
}

func TestStoreDeleteRemovesTheRecord(t *testing.T) {
	s := newStore(t)
	if err := s.Save(sampleOp("op-1")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := s.Delete("op-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok, _ := s.Load("op-1"); ok {
		t.Error("record still present after Delete")
	}
	if err := s.Delete("op-1"); err != nil {
		t.Errorf("deleting an already-deleted record should be a no-op, got %v", err)
	}
}

// An operation ID reaches the store from a URL path. Nothing derived from a
// request may be able to name a file outside the state directory.
func TestStoreRefusesAnIDThatEscapesTheDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	outside := filepath.Join(filepath.Dir(dir), "escaped.json")
	if err := os.WriteFile(outside, []byte(`{"id":"escaped"}`), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}

	for _, id := range []string{"../escaped", "..", "", "a/b", string(filepath.Separator) + "etc/passwd"} {
		if _, ok, err := s.Load(id); ok || err == nil {
			t.Errorf("Load(%q) = ok:%v err:%v, want a refusal", id, ok, err)
		}
		if err := s.Save(Operation{ID: id}); err == nil {
			t.Errorf("Save(%q) accepted an ID that is not a single path segment", id)
		}
	}
}

func TestStoreRecordOnDiskIsTheWireFormat(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "operations")
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := s.Save(sampleOp("op-1")); err != nil {
		t.Fatalf("Save: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "op-1.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("stored record is not valid JSON: %v", err)
	}
	if fields["requestId"] != "client-op-1" {
		t.Errorf("stored record does not use the wire field names: %s", raw)
	}
}
