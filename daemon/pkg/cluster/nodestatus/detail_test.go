package nodestatus

import (
	"encoding/json"
	"testing"
	"time"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
)

// The node detail page asks each node how much memory and disk it has and
// which olaresd it is running. All three are already in the state this daemon
// refreshes; none of them reached the report.
func TestBuildReportsMemoryDiskAndOlaresdVersion(t *testing.T) {
	version := "1.12.6-rc.2"
	st := clistate.State{
		TerminusState:  clistate.TerminusRunning,
		Memory:         "128 G",
		Disk:           "3725 G",
		OlaresdVersion: &version,
	}

	got := Build(Identity{}, st, nil, time.Now())

	if got.Memory != "128 G" {
		t.Errorf("memory = %q, want what the host reported", got.Memory)
	}
	if got.Disk != "3725 G" {
		t.Errorf("disk = %q, want what the host reported", got.Disk)
	}
	if got.OlaresdVersion != version {
		t.Errorf("olaresdVersion = %q, want %q", got.OlaresdVersion, version)
	}
}

// A node that could not work out one of these reports nothing for it. Zero is
// a measurement and "unknown" is a sentence: an empty string is neither.
func TestBuildLeavesUnknownHardwareEmpty(t *testing.T) {
	got := Build(Identity{}, clistate.State{TerminusState: clistate.TerminusRunning}, nil, time.Now())

	if got.Memory != "" || got.Disk != "" || got.OlaresdVersion != "" {
		t.Errorf("want empty fields rather than invented ones: %+v", got)
	}
}

// These are the strings the host formatted, not byte counts. A client that read
// them as numbers would show a machine with 128 bytes of memory, so the wire
// names must not suggest one.
func TestMemoryAndDiskAreNotPresentedAsByteCounts(t *testing.T) {
	st := clistate.State{TerminusState: clistate.TerminusRunning, Memory: "128 G", Disk: "3725 G"}

	raw, err := json.Marshal(Build(Identity{}, st, nil, time.Now()))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	for _, name := range []string{"memoryBytes", "memory_bytes", "diskBytes", "disk_bytes"} {
		if _, ok := fields[name]; ok {
			t.Errorf("field %q claims a byte count for a string the host formatted: %s", name, raw)
		}
	}
	for name, want := range map[string]string{"memory": `"128 G"`, "disk": `"3725 G"`} {
		got, ok := fields[name]
		if !ok {
			t.Errorf("field %q missing: %s", name, raw)
			continue
		}
		if string(got) != want {
			t.Errorf("%s = %s, want %s", name, got, want)
		}
	}
}
