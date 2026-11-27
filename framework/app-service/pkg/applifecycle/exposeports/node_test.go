package exposeports

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadNodeSocketsIncludesBothAddressFamilies(t *testing.T) {
	dir := t.TempDir()
	for i, name := range []string{"tcp", "tcp6", "udp", "udp6"} {
		address := "0100007F"
		if strings.HasSuffix(name, "6") {
			address = "00000000000000000000000001000000"
		}
		data := fmt.Sprintf("  sl  local_address rem_address st\n  0: %s:%04X 00000000:0000 0A\n", address, 46800+i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	busy, err := readNodeSockets(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []portKey{{46800, "tcp"}, {46801, "tcp"}, {46802, "udp"}, {46803, "udp"}} {
		if _, exists := busy[key]; !exists {
			t.Errorf("socket not found: %+v", key)
		}
	}
	if len(busy) != 4 {
		t.Fatalf("got %d sockets, want 4", len(busy))
	}
	if err := os.Remove(filepath.Join(dir, "udp6")); err != nil {
		t.Fatal(err)
	}
	if _, err := readNodeSockets(dir); err == nil {
		t.Fatal("missing table must fail closed")
	}
}

func TestScanSocketTableRejectsMalformedInput(t *testing.T) {
	for _, data := range []string{"", "wrong header", "local_address\n0:", "local_address\n0: 00000000", "local_address\n0: 00000000:Z"} {
		if err := scanSocketTable(strings.NewReader(data), "tcp", make(portSet)); err == nil {
			t.Errorf("accepted %q", data)
		}
	}
}
