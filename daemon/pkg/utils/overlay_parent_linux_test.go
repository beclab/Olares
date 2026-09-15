//go:build linux
// +build linux

package utils

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/vishvananda/netlink"
)

type fakeParentOps struct {
	links   map[string]string // name or altname -> primary name
	added   []string
	written map[string]string
	removed []string
	macErr  error
	lookErr error
}

func (f *fakeParentOps) ops(dev string, selectErr error) overlayParentOps {
	return overlayParentOps{
		selectParent: func(context.Context) (string, error) { return dev, selectErr },
		linkByName: func(name string) (netlink.Link, error) {
			primary, ok := f.links[name]
			if !ok {
				return nil, errors.New("Link not found")
			}
			return &netlink.Device{LinkAttrs: netlink.LinkAttrs{Name: primary}}, nil
		},
		addAltName: func(link netlink.Link, name string) error {
			f.added = append(f.added, link.Attrs().Name+"="+name)
			f.links[name] = link.Attrs().Name
			return nil
		},
		readMAC:  func(dev string) (string, error) { return "d8:43:ae:af:5a:33", f.macErr },
		lookPath: func(string) (string, error) { return "/bin/ip", f.lookErr },
		writeFile: func(path string, data []byte) error {
			f.written[path] = string(data)
			return nil
		},
		removeFile: func(path string) error {
			f.removed = append(f.removed, path)
			return os.ErrNotExist
		},
	}
}

func newFakeParentOps(links map[string]string) *fakeParentOps {
	return &fakeParentOps{links: links, written: map[string]string{}}
}

func TestEnsureOverlayParentAltnameAddsNameAndPersists(t *testing.T) {
	orig := OverlayUdevRuleFile
	defer func() { OverlayUdevRuleFile = orig }()
	OverlayUdevRuleFile = filepath.Join(t.TempDir(), "80-olares-lan.rules")

	f := newFakeParentOps(map[string]string{"enp3s0": "enp3s0"})
	dev, err := ensureOverlayParentAltname(context.Background(), f.ops("enp3s0", nil))
	if err != nil || dev != "enp3s0" {
		t.Fatalf("ensure = %q,%v", dev, err)
	}
	if len(f.added) != 1 || f.added[0] != "enp3s0="+OverlayParentAltname {
		t.Fatalf("alternative name must be added to enp3s0, got %v", f.added)
	}
	rule := f.written[OverlayUdevRuleFile]
	for _, want := range []string{`ATTR{address}=="d8:43:ae:af:5a:33"`, `RUN+="/bin/ip link property add dev $env{INTERFACE} altname olares-lan"`} {
		if !strings.Contains(rule, want) {
			t.Fatalf("udev rule must be persisted with %q, got %q", want, rule)
		}
	}
	if len(f.removed) != 1 || f.removed[0] != legacyOverlayLinkFile {
		t.Fatalf("legacy link file must be removed, got %v", f.removed)
	}
}

func TestEnsureOverlayParentAltnameFailsWithoutIPCommand(t *testing.T) {
	f := newFakeParentOps(map[string]string{"enp3s0": "enp3s0"})
	f.lookErr = errors.New("executable file not found")
	if _, err := ensureOverlayParentAltname(context.Background(), f.ops("enp3s0", nil)); err == nil {
		t.Fatal("a rule with an unknown ip path must not be written")
	}
	if len(f.written) != 0 {
		t.Fatalf("no rule may be written without the ip path, got %v", f.written)
	}
}

func TestEnsureOverlayParentAltnameIsIdempotent(t *testing.T) {
	f := newFakeParentOps(map[string]string{"enp3s0": "enp3s0", OverlayParentAltname: "enp3s0"})
	if _, err := ensureOverlayParentAltname(context.Background(), f.ops("enp3s0", nil)); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if len(f.added) != 0 {
		t.Fatalf("name already on the right device must not be added again, got %v", f.added)
	}
	if len(f.written) != 1 {
		t.Fatal("udev rule must still be refreshed")
	}
}

func TestEnsureOverlayParentAltnameRefusesToRebind(t *testing.T) {
	f := newFakeParentOps(map[string]string{"enp3s0": "enp3s0", "enp2s0": "enp2s0", OverlayParentAltname: "enp2s0"})
	_, err := ensureOverlayParentAltname(context.Background(), f.ops("enp3s0", nil))
	if !errors.Is(err, ErrAltnameBound) {
		t.Fatalf("expected ErrAltnameBound, got %v", err)
	}
	if len(f.added) != 0 || len(f.written) != 0 {
		t.Fatal("nothing may change when the name is bound elsewhere")
	}
}

func TestEnsureOverlayParentAltnameToleratesEEXIST(t *testing.T) {
	f := newFakeParentOps(map[string]string{"enp3s0": "enp3s0"})
	ops := f.ops("enp3s0", nil)
	ops.addAltName = func(netlink.Link, string) error { return syscall.EEXIST }
	if _, err := ensureOverlayParentAltname(context.Background(), ops); err != nil {
		t.Fatalf("EEXIST from the kernel must be treated as already added: %v", err)
	}
}

func TestEnsureOverlayParentAltnamePropagatesSelectionError(t *testing.T) {
	f := newFakeParentOps(map[string]string{})
	if _, err := ensureOverlayParentAltname(context.Background(), f.ops("", ErrNoWiredInterface)); !errors.Is(err, ErrNoWiredInterface) {
		t.Fatalf("expected ErrNoWiredInterface, got %v", err)
	}
	if len(f.written) != 0 {
		t.Fatal("no udev rule without a parent")
	}
}

func TestEnsureOverlayParentAltnameKeepsBoundDeviceWhenSelectionFails(t *testing.T) {
	orig := OverlayUdevRuleFile
	defer func() { OverlayUdevRuleFile = orig }()
	OverlayUdevRuleFile = filepath.Join(t.TempDir(), "80-olares-lan.rules")

	f := newFakeParentOps(map[string]string{"enp3s0": "enp3s0", OverlayParentAltname: "enp3s0"})
	dev, err := ensureOverlayParentAltname(context.Background(), f.ops("", ErrNoWiredInterface))
	if err != nil || dev != "enp3s0" {
		t.Fatalf("a name already in place must be kept when the NIC has no address yet, got %q,%v", dev, err)
	}
	if len(f.added) != 0 {
		t.Fatalf("nothing to add, got %v", f.added)
	}
	if len(f.written) != 1 {
		t.Fatal("udev rule must be refreshed for the bound device")
	}
}

func TestConvergeOverlayParentRetriesEnsureThenHeals(t *testing.T) {
	ensureCalls, healCalls := 0, 0
	ensure := func(context.Context) (string, error) {
		ensureCalls++
		if ensureCalls < 3 {
			return "", ErrNoWiredInterface
		}
		return "enp3s0", nil
	}
	heal := func(context.Context) (int, error) {
		healCalls++
		if healCalls < 2 {
			return 0, errors.New("apiserver not ready")
		}
		return 1, nil
	}
	fired := make(chan time.Time, 1)
	after := func(time.Duration) <-chan time.Time {
		fired <- time.Time{}
		return fired
	}
	convergeOverlayParent(context.Background(), ensure, heal, after)
	if ensureCalls != 3 {
		t.Fatalf("ensure must be retried until it succeeds, got %d calls", ensureCalls)
	}
	if healCalls != 2 {
		t.Fatalf("heal must run after the name is in place and retry until the API answers, got %d calls", healCalls)
	}
}

func TestConvergeOverlayParentGivesUpAfterBoundedAttempts(t *testing.T) {
	ensureCalls, healCalls := 0, 0
	ensure := func(context.Context) (string, error) { ensureCalls++; return "", ErrNoWiredInterface }
	heal := func(context.Context) (int, error) { healCalls++; return 0, nil }
	fired := make(chan time.Time, 1)
	after := func(time.Duration) <-chan time.Time { fired <- time.Time{}; return fired }
	convergeOverlayParent(context.Background(), ensure, heal, after)
	if ensureCalls != overlayConvergeAttempts {
		t.Fatalf("ensure must stop after %d attempts, got %d", overlayConvergeAttempts, ensureCalls)
	}
	if healCalls != 0 {
		t.Fatal("heal must not run when the name was never restored")
	}
}

func TestConvergeOverlayParentStopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ensureCalls := 0
	ensure := func(context.Context) (string, error) {
		ensureCalls++
		cancel()
		return "", ErrNoWiredInterface
	}
	heal := func(context.Context) (int, error) { t.Fatal("heal must not run"); return 0, nil }
	after := func(time.Duration) <-chan time.Time { return make(chan time.Time) }
	convergeOverlayParent(ctx, ensure, heal, after)
	if ensureCalls != 1 {
		t.Fatalf("ensure must stop once the context is cancelled, got %d calls", ensureCalls)
	}
}
