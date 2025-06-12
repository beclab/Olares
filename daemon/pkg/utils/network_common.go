package utils

import (
	"context"
	"fmt"
	"net"
	"os/exec"
)

const CHECK_CONNECTIVITY_URL = "http://connectivity-check.ubuntu.com/"

func CheckInterfaceIPv4Connectivity(ctx context.Context, interfaceName string) bool {
	// try to connect to the CHECK_CONNECTIVITY_URL using the specified interface
	cmd := exec.CommandContext(ctx, "curl", "--interface", interfaceName, "--connect-timeout", "5", "-s", "-o", "/dev/null", CHECK_CONNECTIVITY_URL)
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
