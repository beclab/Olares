package authz

import "testing"

func TestParseL5dClientID(t *testing.T) {
	const suffix = ".serviceaccount.identity.linkerd.cluster.local"
	valid := "default.user-space-alice" + suffix
	sa, ns, err := ParseL5dClientID(valid)
	if err != nil || sa != "default" || ns != "user-space-alice" {
		t.Fatalf("valid: sa=%q ns=%q err=%v", sa, ns, err)
	}
	if _, _, err := ParseL5dClientID(""); err == nil {
		t.Fatal("empty expected error")
	}
	if _, _, err := ParseL5dClientID("not-spiffe"); err == nil {
		t.Fatal("malformed expected error")
	}
	if _, _, err := ParseL5dClientID("onlyns" + suffix); err == nil {
		t.Fatal("short core expected error")
	}
}

func TestDeriveViewer(t *testing.T) {
	cases := []struct {
		ns   string
		want string
		ok   bool
	}{
		{"user-space-alice", "alice", true},
		{"USER-SPACE-BOB", "bob", true},
		{"user-system-svc", "svc", true},
		{"kube-system", "", false},
		{"linkerd", "", false},
	}
	for _, tc := range cases {
		got, ok := DeriveViewer(tc.ns)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("DeriveViewer(%q) = %q,%v want %q,%v", tc.ns, got, ok, tc.want, tc.ok)
		}
	}
}

func TestIsSharedInclusterHost(t *testing.T) {
	if !IsSharedInclusterHost("a1b2c3d4.alice.olares.com") {
		t.Fatal("expected shared host")
	}
	if IsSharedInclusterHost("a1b2c3d.alice.olares.com") {
		t.Fatal("7-char prefix must not match")
	}
	if IsSharedInclusterHost("notshared.example.com") {
		t.Fatal("non-hash8 must not match")
	}
	if IsSharedInclusterHost("") {
		t.Fatal("empty must not match")
	}
}

func TestHostViewerLabel(t *testing.T) {
	if got := HostViewerLabel("a1b2c3d4.alice.olares.com:443"); got != "alice" {
		t.Fatalf("viewer = %q", got)
	}
	if got := HostViewerLabel("short"); got != "" {
		t.Fatalf("short host viewer = %q", got)
	}
}
