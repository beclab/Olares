// nfs_test.go: client-side helpers for `olares-cli files nfs`.
//
// These pin the pure functions that gate the NFS verbs before any
// wire request: target classification (host vs host:/export vs
// invalid), the export → remountable-URL renderer, and the
// favorites-book NFS/SMB filter. The wire side (MountNFS) is tested
// in internal/files/smbmount.
package files

import (
	"strings"
	"testing"
)

func TestNFSTargetKind(t *testing.T) {
	cases := []struct {
		in   string
		want string // "" means expect an error
	}{
		{"192.168.1.10", "host"},
		{"nas.local", "host"},
		{"192.168.1.10:/data", "full"},
		{"192.168.1.10:/", "full"},
		{"nas.local:/export/sub", "full"},
		{"  192.168.1.10:/data  ", "full"}, // trimmed
		{"nfs://192.168.1.10/data", ""},    // scheme URL rejected
		{"//host/share", ""},               // SMB-style rejected
		{"192.168.1.10://data", ""},        // double slash rejected
		{"192.168.1.10/data", ""},          // missing colon
		{"", ""},                           // empty
		{"   ", ""},                        // whitespace only
	}
	for _, c := range cases {
		got, err := nfsTargetKind(c.in)
		if c.want == "" {
			if err == nil {
				t.Errorf("nfsTargetKind(%q): want error, got kind=%q", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("nfsTargetKind(%q): unexpected error %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("nfsTargetKind(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNFSRemountURL(t *testing.T) {
	cases := []struct {
		host, path, want string
	}{
		{"192.168.1.10", "/data", "192.168.1.10:/data"},
		{"192.168.1.10", "data", "192.168.1.10:/data"},               // splice leading slash
		{"192.168.1.10", "", "192.168.1.10:/"},                       // empty export → root
		{"192.168.1.10", "192.168.1.10:/data", "192.168.1.10:/data"}, // already full → unchanged
		{"192.168.1.10:/old", "/new", "192.168.1.10:/new"},           // host carried an export → strip it
	}
	for _, c := range cases {
		if got := nfsRemountURL(c.host, c.path); got != c.want {
			t.Errorf("nfsRemountURL(%q, %q) = %q, want %q", c.host, c.path, got, c.want)
		}
	}
}

func TestIsNFSFavorite(t *testing.T) {
	for _, nfs := range []string{"192.168.1.10:/data", "192.168.1.10", "nas.local:/export"} {
		if !isNFSFavorite(nfs) {
			t.Errorf("isNFSFavorite(%q) = false, want true", nfs)
		}
	}
	for _, smb := range []string{"//host/share", "  //host/share"} {
		if isNFSFavorite(smb) {
			t.Errorf("isNFSFavorite(%q) = true, want false", smb)
		}
	}
}

// TestNFSTargetKind_SMBErrorMentionsSMB ensures the // rejection
// points the user at `files smb` rather than a generic parse error.
func TestNFSTargetKind_SMBErrorMentionsSMB(t *testing.T) {
	_, err := nfsTargetKind("//host/share")
	if err == nil {
		t.Fatal("want error")
	}
	if !strings.Contains(err.Error(), "smb") {
		t.Errorf("error %q should mention smb", err.Error())
	}
}
