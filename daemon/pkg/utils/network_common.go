package utils

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"sync"

	"github.com/google/uuid"
)

const CHECK_CONNECTIVITY_URL = "http://connectivity-check.ubuntu.com/"

var (
	networkChangedNotify map[string]bool
	networkChangedMu     sync.Mutex
)

func CheckInterfaceIPv4Connectivity(ctx context.Context, interfaceName string) bool {
	// try to connect to the CHECK_CONNECTIVITY_URL using the specified interface
	cmd := exec.CommandContext(ctx, "curl", "-4", "--interface", interfaceName, "--connect-timeout", "5", "-s", "-o", "/dev/null", CHECK_CONNECTIVITY_URL)
	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

func CheckInterfaceIPv6Connectivity(ctx context.Context, interfaceName string) bool {
	// try to connect to the CHECK_CONNECTIVITY_URL using the specified interface
	cmd := exec.CommandContext(ctx, "curl", "-6", "--interface", interfaceName, "--connect-timeout", "5", "-s", "-o", "/dev/null", CHECK_CONNECTIVITY_URL)
	if err := cmd.Run(); err == nil {
		return true
	}

	return false
}

func MaskFromCIDR(bits int) (string, error) {
	if bits < 0 || bits > 32 {
		return "", fmt.Errorf("invalid bits: %d", bits)
	}
	mask := net.CIDRMask(bits, 32)
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]), nil
}

func NotifyNetworkChanged() {
	networkChangedMu.Lock()
	defer networkChangedMu.Unlock()
	for key := range networkChangedNotify {
		networkChangedNotify[key] = true
	}
}

func RegisterNetworkChangedNotify() func() bool {
	networkChangedMu.Lock()
	defer networkChangedMu.Unlock()
	if networkChangedNotify == nil {
		networkChangedNotify = make(map[string]bool)
	}
	notifyId := uuid.New().String()
	networkChangedNotify[notifyId] = false
	return func() bool {
		networkChangedMu.Lock()
		defer networkChangedMu.Unlock()
		changed := networkChangedNotify[notifyId]
		networkChangedNotify[notifyId] = false
		return changed
	}
}
