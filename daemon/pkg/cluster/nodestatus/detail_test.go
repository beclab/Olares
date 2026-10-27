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

func TestBuildReportsHostDetailFields(t *testing.T) {
	ssid := "olares-lab"
	st := clistate.State{
		TerminusState:  clistate.TerminusRunning,
		OsArch:         "amd64",
		OsKernel:       "6.8.0-60-generic",
		HostIP:         "10.0.0.2",
		WiredConnected: true,
		WifiSSID:       &ssid,
	}

	got := Build(Identity{}, st, nil, time.Now())

	if got.OsArch != "amd64" {
		t.Errorf("os_arch = %q, want amd64", got.OsArch)
	}
	if got.OsKernel != "6.8.0-60-generic" {
		t.Errorf("os_kernel = %q, want 6.8.0-60-generic", got.OsKernel)
	}
	if got.HostIP != "10.0.0.2" {
		t.Errorf("hostIp = %q, want 10.0.0.2", got.HostIP)
	}
	if !got.WiredConnected {
		t.Errorf("wiredConnected = false, want true")
	}
	if got.WifiSSID != ssid {
		t.Errorf("wifiSSID = %q, want %q", got.WifiSSID, ssid)
	}
}

func TestBuildOmitsWifiSSIDWhenDisconnected(t *testing.T) {
	st := clistate.State{
		TerminusState:  clistate.TerminusRunning,
		WiredConnected: true,
		HostIP:         "10.0.0.2",
	}

	raw, err := json.Marshal(Build(Identity{}, st, nil, time.Now()))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	if _, ok := fields["wifiSSID"]; ok {
		t.Errorf("wifiSSID must be omitted when disconnected: %s", raw)
	}
	for _, key := range []string{"os_arch", "os_kernel", "hostIp", "wiredConnected"} {
		if _, ok := fields[key]; !ok {
			t.Errorf("field %q missing: %s", key, raw)
		}
	}
}

func TestBuildLeavesUnknownHostDetailsEmpty(t *testing.T) {
	got := Build(Identity{}, clistate.State{TerminusState: clistate.TerminusRunning}, nil, time.Now())

	if got.OsArch != "" || got.OsKernel != "" || got.HostIP != "" || got.WifiSSID != "" {
		t.Errorf("want empty host detail strings: %+v", got)
	}
	if got.WiredConnected {
		t.Errorf("wiredConnected = true, want false when unknown")
	}
}
