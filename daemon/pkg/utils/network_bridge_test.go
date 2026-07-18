//go:build linux
// +build linux

package utils

import "testing"

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
