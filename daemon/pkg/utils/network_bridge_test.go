//go:build linux
// +build linux

package utils

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestParseNMConnections(t *testing.T) {
	output := []byte(
		"bde7eb90-be27-4b5a-a52e-f26fca8ba41f:802-3-ethernet:original-connection\n" +
			"72dca0fd-de3b-4a96-b73f-595f2c2bd661:802-3-ethernet:original-connection\n" +
			"4399b3ff-1d87-4500-b163-6473882fd200:bridge:br-olares\n" +
			"a3ed0548-a413-4410-a34b-29837addbf4e:802-3-ethernet:br-olares-slave-enp129s0\n" +
			"df58d234-88a1-4c9f-8cd4-c3168802cb75:802-11-wireless:BE13000\\:IoT\n" +
			"\n" +
			"malformed-line-without-fields\n",
	)

	conns := parseNMConnections(output)
	if len(conns) != 5 {
		t.Fatalf("expected 5 parsed connections, got %d: %+v", len(conns), conns)
	}

	// escaped ':' inside NAME must be unescaped
	if got := conns[4].Name; got != "BE13000:IoT" {
		t.Errorf("expected unescaped name 'BE13000:IoT', got %q", got)
	}
	if conns[4].Type != "802-11-wireless" {
		t.Errorf("unexpected type for wireless entry: %q", conns[4].Type)
	}

	// two distinct UUIDs share the backup name: this is the duplicate that
	// causes the auto-activation race the fix must prune.
	var dupUUIDs []string
	for _, c := range conns {
		if c.Name == originalConnectionName {
			dupUUIDs = append(dupUUIDs, c.UUID)
		}
	}
	if len(dupUUIDs) != 2 {
		t.Fatalf("expected 2 duplicate %q profiles, got %d", originalConnectionName, len(dupUUIDs))
	}
	if dupUUIDs[0] == dupUUIDs[1] {
		t.Errorf("duplicate profiles should have distinct UUIDs, both %q", dupUUIDs[0])
	}
}

func TestParseNMConnectionsBridgeClassification(t *testing.T) {
	output := []byte(
		"11111111-1111-1111-1111-111111111111:bridge:br-olares\n" +
			"22222222-2222-2222-2222-222222222222:802-3-ethernet:br-olares-slave-eth0\n" +
			"33333333-3333-3333-3333-333333333333:802-3-ethernet:original-connection\n",
	)

	conns := parseNMConnections(output)

	var bridges, slaves int
	for _, c := range conns {
		switch {
		case c.Name == bridgeConnectionName:
			bridges++
		case len(c.Name) >= len(bridgeSlavePrefix) && c.Name[:len(bridgeSlavePrefix)] == bridgeSlavePrefix:
			slaves++
		}
	}
	if bridges != 1 || slaves != 1 {
		t.Fatalf("expected 1 bridge and 1 slave, got bridges=%d slaves=%d", bridges, slaves)
	}
}

func TestBridgeAddArgsClonedMAC(t *testing.T) {
	mac := "aa:bb:cc:dd:ee:ff"
	args := bridgeAddArgs(mac)

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "ethernet.cloned-mac-address "+mac) {
		t.Fatalf("expected cloned-mac-address %q in argv, got %v", mac, args)
	}
	if !strings.Contains(joined, "ipv4.method auto") {
		t.Fatalf("expected ipv4.method auto in argv, got %v", args)
	}
	if mac == "" {
		t.Fatal("test mac must be non-empty")
	}
	// ensure the mac token itself is present as a dedicated argv element
	found := false
	for i, a := range args {
		if a == "ethernet.cloned-mac-address" && i+1 < len(args) && args[i+1] == mac {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("cloned-mac-address value missing or empty in %v", args)
	}
}

func TestClassifyBridgeResetUUIDsIgnoresForeignBridges(t *testing.T) {
	conns := []nmConnection{
		{UUID: "b1", Type: "bridge", Name: bridgeConnectionName},
		{UUID: "s1", Type: "802-3-ethernet", Name: bridgeSlavePrefix + "eth0"},
		{UUID: "o1", Type: "802-3-ethernet", Name: originalConnectionName},
		{UUID: "docker", Type: "bridge", Name: "docker0"},
		{UUID: "br0", Type: "bridge", Name: "br0"},
		{UUID: "other", Type: "802-3-ethernet", Name: "Wired connection 1"},
	}
	bridges, slaves, originals := classifyBridgeResetUUIDs(conns)
	if len(bridges) != 1 || bridges[0] != "b1" {
		t.Fatalf("unexpected bridges: %v", bridges)
	}
	if len(slaves) != 1 || slaves[0] != "s1" {
		t.Fatalf("unexpected slaves: %v", slaves)
	}
	if len(originals) != 1 || originals[0] != "o1" {
		t.Fatalf("unexpected originals: %v", originals)
	}
	for _, u := range append(append(bridges, slaves...), originals...) {
		if u == "docker" || u == "br0" || u == "other" {
			t.Fatalf("foreign UUID %q must not be selected for reset delete", u)
		}
	}
}

func TestActivateBridgeSwitchSkipsUpWhenDownFails(t *testing.T) {
	orig := nmcliCombinedOutput
	t.Cleanup(func() { nmcliCombinedOutput = orig })

	var calls [][]string
	nmcliCombinedOutput = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		cp := append([]string{name}, args...)
		calls = append(calls, cp)
		if len(args) >= 2 && args[1] == "down" {
			return nil, errors.New("down failed")
		}
		return []byte{}, nil
	}

	err := activateBridgeSwitch(context.Background(), "nmcli", "slave-u", "if-u", "br-u")
	if err == nil {
		t.Fatal("expected error when down fails")
	}
	if !strings.Contains(err.Error(), "down original connection") {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range calls {
		joined := strings.Join(c, " ")
		if strings.Contains(joined, "connection up") {
			t.Fatalf("up must not be called after down failure; calls=%v", calls)
		}
	}
	if len(calls) != 2 {
		t.Fatalf("expected modify+down only, got %d calls: %v", len(calls), calls)
	}
}

func TestCommitNMCheckpointDestroyFailureTriggersRollback(t *testing.T) {
	origDestroy := nmCheckpointDestroyFn
	origRollback := nmCheckpointRollbackFn
	t.Cleanup(func() {
		nmCheckpointDestroyFn = origDestroy
		nmCheckpointRollbackFn = origRollback
	})

	destroyCalls := 0
	rollbackCalls := 0
	manualCalls := 0
	nmCheckpointDestroyFn = func(ctx context.Context, conn *dbus.Conn, cp dbus.ObjectPath) error {
		destroyCalls++
		return errors.New("destroy boom")
	}
	nmCheckpointRollbackFn = func(ctx context.Context, conn *dbus.Conn, cp dbus.ObjectPath) error {
		rollbackCalls++
		return nil
	}

	err := commitNMCheckpoint(context.Background(), nil, dbus.ObjectPath("/org/freedesktop/NetworkManager/Checkpoint/1"), func() {
		manualCalls++
	})
	if err == nil {
		t.Fatal("expected commit failure error")
	}
	if !strings.Contains(err.Error(), "checkpoint commit failed") {
		t.Fatalf("unexpected error: %v", err)
	}
	if destroyCalls != checkpointDestroyRetries {
		t.Fatalf("expected %d destroy attempts, got %d", checkpointDestroyRetries, destroyCalls)
	}
	if rollbackCalls != 1 {
		t.Fatalf("expected 1 rollback, got %d", rollbackCalls)
	}
	if manualCalls != 0 {
		t.Fatalf("manual rollback should not run when CheckpointRollback succeeds, got %d", manualCalls)
	}
}

func TestCommitNMCheckpointDestroyFailureManualWhenRollbackFails(t *testing.T) {
	origDestroy := nmCheckpointDestroyFn
	origRollback := nmCheckpointRollbackFn
	t.Cleanup(func() {
		nmCheckpointDestroyFn = origDestroy
		nmCheckpointRollbackFn = origRollback
	})

	nmCheckpointDestroyFn = func(ctx context.Context, conn *dbus.Conn, cp dbus.ObjectPath) error {
		return errors.New("destroy boom")
	}
	nmCheckpointRollbackFn = func(ctx context.Context, conn *dbus.Conn, cp dbus.ObjectPath) error {
		return errors.New("rollback boom")
	}
	manualCalls := 0
	err := commitNMCheckpoint(context.Background(), nil, dbus.ObjectPath("/cp/1"), func() {
		manualCalls++
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if manualCalls != 1 {
		t.Fatalf("expected manual rollback once, got %d", manualCalls)
	}
}
