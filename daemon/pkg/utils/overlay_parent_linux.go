//go:build linux
// +build linux

package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	lookPath     func(file string) (string, error)
	writeFile    func(path string, data []byte) error
	removeFile   func(path string) error
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
		lookPath: exec.LookPath,
		writeFile: func(path string, data []byte) error {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			return os.WriteFile(path, data, 0o644)
		},
		removeFile: os.Remove,
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
	bound, boundErr := ops.linkByName(OverlayParentAltname)
	dev, selErr := ops.selectParent(ctx)
	switch {
	case selErr != nil && boundErr == nil:
		// The name is already in place (the udev rule re-adds it at boot) but
		// the wired NIC cannot be identified yet, typically because DHCP has
		// not finished. Keep the bound device instead of failing.
		dev = bound.Attrs().Name
		klog.Warningf("overlay-parent: wired interface not identified (%v); keeping %s on %s", selErr, OverlayParentAltname, dev)
	case selErr != nil:
		klog.Errorf("overlay-parent: select wired interface failed: %v", selErr)
		return "", selErr
	case boundErr == nil && bound.Attrs().Name != dev:
		klog.Errorf("overlay-parent: %s is bound to %s but the wired interface is %s", OverlayParentAltname, bound.Attrs().Name, dev)
		return "", fmt.Errorf("%w: %s is on %s, wired interface is %s", ErrAltnameBound, OverlayParentAltname, bound.Attrs().Name, dev)
	case boundErr != nil:
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
	ipPath, err := ops.lookPath("ip")
	if err != nil {
		klog.Errorf("overlay-parent: locate the ip command failed: %v", err)
		return "", fmt.Errorf("locate the ip command for the udev rule: %w", err)
	}
	if err := ops.writeFile(OverlayUdevRuleFile, []byte(overlayUdevRuleContent(mac, ipPath))); err != nil {
		klog.Errorf("overlay-parent: persist %s failed: %v", OverlayUdevRuleFile, err)
		return "", fmt.Errorf("persist alternative name: %w", err)
	}
	if err := ops.removeFile(legacyOverlayLinkFile); err != nil && !os.IsNotExist(err) {
		klog.Warningf("overlay-parent: remove legacy %s failed: %v", legacyOverlayLinkFile, err)
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

const (
	overlayConvergeAttempts = 30
	overlayConvergeInterval = 10 * time.Second
)

// ConvergeOverlayGateway brings the node to the state the desired-state file
// asks for. It runs at daemon start: the CNI DHCP daemon is kept running in
// every case; when the switch is on, the alternative name is re-added and
// overlay Pods that lost their LAN interface are restarted once. Both steps
// retry in the background because the daemon may start before the wired NIC
// has its address or before the API server answers.
func ConvergeOverlayGateway(ctx context.Context) {
	if err := EnsureCniDhcpActive(ctx); err != nil {
		klog.Errorf("overlay-converge: %v", err)
	}
	if !OverlayGatewayDesired() {
		return
	}
	go convergeOverlayParent(ctx, EnsureOverlayParentAltname, HealOverlayPodsWithoutNet1, time.After)
}

// convergeOverlayParent retries the alternative name until it is in place,
// then heals the overlay Pods once the API server answers. It gives up after
// a bounded number of attempts and logs why, so a node that never gets a
// wired address does not spin forever.
func convergeOverlayParent(ctx context.Context, ensure func(context.Context) (string, error), heal func(context.Context) (int, error), after func(time.Duration) <-chan time.Time) {
	wait := func() bool {
		select {
		case <-ctx.Done():
			return false
		case <-after(overlayConvergeInterval):
			return true
		}
	}
	var lastErr error
	for attempt := 1; ; attempt++ {
		dev, err := ensure(ctx)
		if err == nil {
			if attempt > 1 {
				klog.Infof("overlay-converge: %s restored on %s after %d attempts", OverlayParentAltname, dev, attempt)
			}
			break
		}
		lastErr = err
		if attempt >= overlayConvergeAttempts {
			klog.Errorf("overlay-converge: alternative name not restored after %d attempts, giving up: %v", attempt, lastErr)
			return
		}
		klog.V(4).Infof("overlay-converge: attempt %d to restore the alternative name failed: %v", attempt, err)
		if !wait() {
			return
		}
	}
	for attempt := 1; attempt <= overlayConvergeAttempts; attempt++ {
		restarted, err := heal(ctx)
		if err == nil {
			if restarted > 0 {
				klog.Infof("overlay-converge: restarted %d overlay pod(s) that had no %s", restarted, overlayLANInterface)
			}
			return
		}
		klog.V(4).Infof("overlay-converge: heal attempt %d not ready: %v", attempt, err)
		if !wait() {
			return
		}
	}
	klog.Errorf("overlay-converge: overlay pods not checked for %s after %d attempts", overlayLANInterface, overlayConvergeAttempts)
}
