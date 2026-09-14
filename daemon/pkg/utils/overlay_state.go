package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// OverlayParentAltname is the alternative interface name the underlay
	// NetworkAttachmentDefinition uses as its macvlan master. Adding it to the
	// wired NIC is the whole node-side work of enabling the overlay gateway.
	OverlayParentAltname = "olares-lan"

	cniDhcpUnit = "cni-dhcp.service"
)

// Paths are variables so tests can redirect them.
var (
	// OverlayDesiredStateFile exists when the user has enabled the overlay
	// gateway on this node. It is the single source of truth for the switch;
	// the runtime state is converged towards it on boot.
	OverlayDesiredStateFile = "/var/lib/olares/overlay-gateway/enabled"
	// OverlayLinkFile persists the alternative name across reboots through
	// systemd-udevd, independently of olaresd.
	OverlayLinkFile = "/etc/systemd/network/10-olares-lan.link"
)

// ErrNoWiredInterface is returned when no connected ethernet interface can
// serve as the overlay parent.
var ErrNoWiredInterface = errors.New("no connected wired interface for the overlay gateway")

// ErrAltnameBound is returned when olares-lan already names a device other
// than the one selected as parent. The name is never moved silently: doing so
// would move every overlay Pod to another network on its next restart.
var ErrAltnameBound = errors.New("overlay parent alternative name is bound to another device")

func OverlayGatewayDesired() bool {
	_, err := os.Stat(OverlayDesiredStateFile)
	return err == nil
}

func SetOverlayGatewayDesired(enabled bool) error {
	if !enabled {
		if err := os.Remove(OverlayDesiredStateFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("clear overlay gateway desired state: %w", err)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(OverlayDesiredStateFile), 0o755); err != nil {
		return fmt.Errorf("create overlay gateway state dir: %w", err)
	}
	if err := os.WriteFile(OverlayDesiredStateFile, nil, 0o644); err != nil {
		return fmt.Errorf("write overlay gateway desired state: %w", err)
	}
	return nil
}

// overlayLinkFileContent renders the systemd.link unit that re-adds the
// alternative name whenever the NIC appears. Matching on MAC plus type keeps
// the rule stable across kernel interface renames.
func overlayLinkFileContent(mac string) string {
	return fmt.Sprintf("[Match]\nMACAddress=%s\nType=ether\n\n[Link]\nAlternativeName=%s\n", mac, OverlayParentAltname)
}

// kernelSupportsAltname reports whether the release string (as in
// /proc/sys/kernel/osrelease) names a kernel with alternative interface names
// (5.5 and later).
func kernelSupportsAltname(release string) bool {
	var major, minor int
	if _, err := fmt.Sscanf(strings.TrimSpace(release), "%d.%d", &major, &minor); err != nil {
		return false
	}
	return major > 5 || (major == 5 && minor >= 5)
}

// IsCniDhcpActive reports whether the CNI DHCP daemon unit is running. The
// daemon is infrastructure that stays up regardless of the switch, so this is
// exposed as its own health flag and never folded into the switch state.
func IsCniDhcpActive(ctx context.Context) bool {
	out, err := exec.CommandContext(ctx, "systemctl", "is-active", cniDhcpUnit).Output()
	return err == nil && strings.TrimSpace(string(out)) == "active"
}

// EnsureCniDhcpActive starts and enables the CNI DHCP daemon. It is called on
// boot convergence only; the switch never stops the daemon.
func EnsureCniDhcpActive(ctx context.Context) error {
	if IsCniDhcpActive(ctx) {
		return nil
	}
	cmd := exec.CommandContext(ctx, "systemctl", "enable", "--now", cniDhcpUnit)
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("enable %s: %w: %s", cniDhcpUnit, err, strings.TrimSpace(string(out)))
	}
	return nil
}
