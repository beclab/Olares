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
	stop := handleCarrierLinkUpdate(ctx, netlink.LinkUpdate{}, false, func() {
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
	stop := handleCarrierLinkUpdate(ctx, netlink.LinkUpdate{}, true, func() {
		t.Fatal("downCallback must not run for nil Link")
	})
	if stop {
		t.Fatal("nil Link should continue, not stop")
	}
}

func TestHandleCarrierLinkUpdateCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stop := handleCarrierLinkUpdate(ctx, netlink.LinkUpdate{}, true, func() {
		t.Fatal("downCallback must not run when ctx cancelled")
	})
	if !stop {
		t.Fatal("cancelled context should stop watcher")
	}
}
