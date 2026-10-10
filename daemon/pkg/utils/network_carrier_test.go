//go:build linux
// +build linux

package utils

import (
	"context"
	"testing"

	"github.com/vishvananda/netlink"
)

func TestHandleCarrierLinkUpdateClosedChannel(t *testing.T) {
	ctx := context.Background()
	called := false
	stop := handleCarrierLinkUpdate(ctx, "enp3s0", netlink.LinkUpdate{}, false, func() {
		called = true
	})
	if !stop {
		t.Fatal("closed channel should stop watcher")
	}
	if called {
		t.Fatal("downCallback must not run on closed channel")
	}
}

func TestHandleCarrierLinkUpdateNilLink(t *testing.T) {
	ctx := context.Background()
	// Zero-value LinkUpdate has nil Link — must not panic via Attrs().
	stop := handleCarrierLinkUpdate(ctx, "enp3s0", netlink.LinkUpdate{}, true, func() {
		t.Fatal("downCallback must not run for nil Link")
	})
	if stop {
		t.Fatal("nil Link should continue, not stop")
	}
}

func TestHandleCarrierLinkUpdateCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stop := handleCarrierLinkUpdate(ctx, "enp3s0", netlink.LinkUpdate{}, true, func() {
		t.Fatal("downCallback must not run when ctx cancelled")
	})
	if !stop {
		t.Fatal("cancelled context should stop watcher")
	}
}

func linkUpdateFor(name string, flags uint32) netlink.LinkUpdate {
	u := netlink.LinkUpdate{Link: &netlink.Device{LinkAttrs: netlink.LinkAttrs{Name: name}}}
	u.IfInfomsg.Flags = flags
	return u
}

func TestHandleCarrierLinkUpdateIgnoresOtherDevices(t *testing.T) {
	stop := handleCarrierLinkUpdate(context.Background(), "enp3s0", linkUpdateFor("wlo1", 0), true, func() {
		t.Fatal("downCallback must not run for a device that is not the overlay parent")
	})
	if stop {
		t.Fatal("unrelated device must not stop the watcher")
	}
}

func TestHandleCarrierLinkUpdateFiresWhenParentLosesCarrier(t *testing.T) {
	called := false
	handleCarrierLinkUpdate(context.Background(), "enp3s0", linkUpdateFor("enp3s0", 0), true, func() { called = true })
	if !called {
		t.Fatal("downCallback must run when the overlay parent goes down")
	}
	called = false
	handleCarrierLinkUpdate(context.Background(), "enp3s0", linkUpdateFor("enp3s0", 1|0x10000), true, func() { called = true })
	if called {
		t.Fatal("downCallback must not run while the parent is up with carrier")
	}
}
