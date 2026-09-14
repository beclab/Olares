//go:build linux
// +build linux

package utils

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/vishvananda/netlink"
)

type fakeParentOps struct {
	links   map[string]string // name or altname -> primary name
	added   []string
	written map[string]string
	macErr  error
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
		readMAC: func(dev string) (string, error) { return "d8:43:ae:af:5a:33", f.macErr },
		writeFile: func(path string, data []byte) error {
			f.written[path] = string(data)
			return nil
		},
	}
}

func newFakeParentOps(links map[string]string) *fakeParentOps {
	return &fakeParentOps{links: links, written: map[string]string{}}
}

func TestEnsureOverlayParentAltnameAddsNameAndPersists(t *testing.T) {
	orig := OverlayLinkFile
	defer func() { OverlayLinkFile = orig }()
	OverlayLinkFile = filepath.Join(t.TempDir(), "10-olares-lan.link")

	f := newFakeParentOps(map[string]string{"enp3s0": "enp3s0"})
	dev, err := ensureOverlayParentAltname(context.Background(), f.ops("enp3s0", nil))
	if err != nil || dev != "enp3s0" {
		t.Fatalf("ensure = %q,%v", dev, err)
	}
	if len(f.added) != 1 || f.added[0] != "enp3s0="+OverlayParentAltname {
		t.Fatalf("alternative name must be added to enp3s0, got %v", f.added)
	}
	if !strings.Contains(f.written[OverlayLinkFile], "MACAddress=d8:43:ae:af:5a:33") {
		t.Fatalf("link file must be persisted, got %v", f.written)
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
		t.Fatal("link file must still be refreshed")
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
		t.Fatal("no link file without a parent")
	}
}
