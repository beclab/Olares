package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"k8s.io/klog/v2"
)

type Device struct {
	Name        string
	Type        string
	State       string
	Connection  string
	Ipv4Gateway string
	Ipv6Gateway string
	Ipv4DNS     string
	Ipv6DNS     string
	Ipv4Address string
	Ipv4Mask    string
	Ipv6Address string
	Method      string
}

type BridgeConnection struct {
	BridgeName  string
	SlaveName   string
	Active      bool
	Ipv4Address string
}

var (
	cmdPathCache   = make(map[string]string)
	cmdPathCacheMu sync.RWMutex
)

// findCommand resolves the absolute path of cmdName. The result is cached
// because callers (nmcli wrappers) invoke it on every command, and the
// previous implementation forked a `bash -c "command -v ..."` subprocess
// each time. On the 5s state-polling path this doubled the process churn
// against NetworkManager. The executable path is effectively immutable for
// the lifetime of the daemon, so a process-wide cache is safe.
func findCommand(ctx context.Context, cmdName string) (string, error) {
	cmdPathCacheMu.RLock()
	cached, ok := cmdPathCache[cmdName]
	cmdPathCacheMu.RUnlock()
	if ok {
		return cached, nil
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("command -v %s", cmdName))
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		klog.Error("find command error, ", cmdName, ", ", err)
		return "", err
	}

	cmdPath := strings.TrimSpace(string(output))
	if cmdPath != "" {
		cmdPathCacheMu.Lock()
		cmdPathCache[cmdName] = cmdPath
		cmdPathCacheMu.Unlock()
	}

	return cmdPath, nil
}
