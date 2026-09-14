//go:build linux
// +build linux

package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
)

var kernelReleaseFile = "/proc/sys/kernel/osrelease"

// KernelSupportsAltname reports whether this kernel supports alternative
// interface names.
func KernelSupportsAltname() bool {
	release, err := os.ReadFile(kernelReleaseFile)
	if err != nil {
		klog.Warningf("overlay-parent: read kernel release failed, assuming altname support: %v", err)
		return true
	}
	return kernelSupportsAltname(string(release))
}

// overlayParentOps are the privileged operations behind
// EnsureOverlayParentAltname, injectable so the decision logic is unit-tested
// without netlink.
type overlayParentOps struct {
	selectParent func(ctx context.Context) (string, error)
	linkByName   func(name string) (netlink.Link, error)
	addAltName   func(link netlink.Link, name string) error
	readMAC      func(dev string) (string, error)
	writeFile    func(path string, data []byte) error
}

func defaultOverlayParentOps() overlayParentOps {
	return overlayParentOps{
		selectParent: SelectOverlayParent,
		linkByName:   netlink.LinkByName,
		addAltName:   netlink.LinkAddAltName,
		readMAC: func(dev string) (string, error) {
			b, err := os.ReadFile("/sys/class/net/" + dev + "/address")
			return strings.TrimSpace(string(b)), err
		},
		writeFile: func(path string, data []byte) error {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			return os.WriteFile(path, data, 0o644)
		},
	}
}

// SelectOverlayParent picks the wired NIC that carries the LAN. It reuses the
// ethernet detection the bridge code used for its slave, so the same NIC the
// old topology bridged is the one that now becomes the parent.
func SelectOverlayParent(ctx context.Context) (string, error) {
	iface, _, _, err := GetEthernetConnection(ctx)
	if err != nil {
		return "", err
	}
	if iface == "" {
		return "", ErrNoWiredInterface
	}
	return iface, nil
}

// ResolveOverlayParent returns the primary name of the device currently
// carrying the alternative name.
func ResolveOverlayParent(ctx context.Context) (string, error) {
	link, err := netlink.LinkByName(OverlayParentAltname)
	if err != nil {
		return "", err
	}
	return link.Attrs().Name, nil
}

// EnsureOverlayParentAltname adds olares-lan to the selected wired NIC and
// persists it. It is idempotent and refuses to move the name between devices.
func EnsureOverlayParentAltname(ctx context.Context) (string, error) {
	return ensureOverlayParentAltname(ctx, defaultOverlayParentOps())
}

func ensureOverlayParentAltname(ctx context.Context, ops overlayParentOps) (string, error) {
	dev, err := ops.selectParent(ctx)
	if err != nil {
		klog.Errorf("overlay-parent: select wired interface failed: %v", err)
		return "", err
	}
	if bound, err := ops.linkByName(OverlayParentAltname); err == nil {
		if bound.Attrs().Name != dev {
			klog.Errorf("overlay-parent: %s is bound to %s but the wired interface is %s", OverlayParentAltname, bound.Attrs().Name, dev)
			return "", fmt.Errorf("%w: %s is on %s, wired interface is %s", ErrAltnameBound, OverlayParentAltname, bound.Attrs().Name, dev)
		}
	} else {
		link, err := ops.linkByName(dev)
		if err != nil {
			klog.Errorf("overlay-parent: lookup %s failed: %v", dev, err)
			return "", fmt.Errorf("lookup wired interface %s: %w", dev, err)
		}
		if err := ops.addAltName(link, OverlayParentAltname); err != nil && !errors.Is(err, syscall.EEXIST) {
			klog.Errorf("overlay-parent: add %s to %s failed: %v", OverlayParentAltname, dev, err)
			return "", fmt.Errorf("add alternative name %s to %s: %w", OverlayParentAltname, dev, err)
		}
		klog.Infof("overlay-parent: %s now carries %s", dev, OverlayParentAltname)
	}
	mac, err := ops.readMAC(dev)
	if err != nil {
		klog.Errorf("overlay-parent: read MAC of %s failed: %v", dev, err)
		return "", fmt.Errorf("read MAC of %s: %w", dev, err)
	}
	if err := ops.writeFile(OverlayLinkFile, []byte(overlayLinkFileContent(mac))); err != nil {
		klog.Errorf("overlay-parent: persist %s failed: %v", OverlayLinkFile, err)
		return "", fmt.Errorf("persist alternative name: %w", err)
	}
	return dev, nil
}

// OverlayParentLinkUp reports whether the device carrying olares-lan is
// administratively up with carrier.
func OverlayParentLinkUp(ctx context.Context) (string, bool, error) {
	link, err := netlink.LinkByName(OverlayParentAltname)
	if err != nil {
		return "", false, err
	}
	attrs := link.Attrs()
	up := attrs.Flags&netlinkFlagUp != 0 && attrs.RawFlags&netlinkFlagLowerUp != 0
	return attrs.Name, up, nil
}

const (
	netlinkFlagUp      = 1       // net.FlagUp
	netlinkFlagLowerUp = 0x10000 // IFF_LOWER_UP
)

// ConvergeOverlayGateway brings the node to the state the desired-state file
// asks for. It runs at daemon start: the CNI DHCP daemon is kept running in
// every case, the alternative name is re-added when the switch is on, and
// overlay Pods that lost their LAN interface are restarted once so a node that
// booted before the name existed heals without user action.
func ConvergeOverlayGateway(ctx context.Context) {
	if err := EnsureCniDhcpActive(ctx); err != nil {
		klog.Errorf("overlay-converge: %v", err)
	}
	if !OverlayGatewayDesired() {
		return
	}
	if _, err := EnsureOverlayParentAltname(ctx); err != nil {
		klog.Errorf("overlay-converge: alternative name not restored: %v", err)
		return
	}
	go func() {
		for attempt := 0; attempt < 30; attempt++ {
			restarted, err := HealOverlayPodsWithoutNet1(ctx)
			if err == nil {
				if restarted > 0 {
					klog.Infof("overlay-converge: restarted %d overlay pod(s) that had no %s", restarted, overlayLANInterface)
				}
				return
			}
			klog.V(4).Infof("overlay-converge: heal attempt %d not ready: %v", attempt+1, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
		}
	}()
}
