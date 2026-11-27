package exposeports

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const nodeSocketDir = "/run/olares/node-sockets"

// Read socket tables from the host network namespace. A TCP dial in a
// meshed Pod can succeed at the local proxy before the node refuses it.
func readNodeSockets(dir string) (portSet, error) {
	busy := make(portSet)
	for _, table := range []struct{ name, protocol string }{
		{"tcp", "tcp"}, {"tcp6", "tcp"}, {"udp", "udp"}, {"udp6", "udp"},
	} {
		f, err := os.Open(filepath.Join(dir, table.name))
		if err != nil {
			return nil, fmt.Errorf("open node socket table %s: %w", table.name, err)
		}
		err = scanSocketTable(f, table.protocol, busy)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("read node socket table %s: %w", table.name, err)
		}
	}
	return busy, nil
}

func scanSocketTable(r io.Reader, protocol string, busy portSet) error {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() || !strings.Contains(scanner.Text(), "local_address") {
		if err := scanner.Err(); err != nil {
			return err
		}
		return fmt.Errorf("missing local_address header")
	}
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		if len(fields) < 2 {
			return fmt.Errorf("invalid socket entry")
		}
		_, hexPort, found := strings.Cut(fields[1], ":")
		if !found {
			return fmt.Errorf("missing local port")
		}
		port, err := strconv.ParseUint(hexPort, 16, 16)
		if err != nil {
			return fmt.Errorf("invalid local port")
		}
		if port != 0 {
			busy[portKey{int32(port), protocol}] = struct{}{}
		}
	}
	return scanner.Err()
}
