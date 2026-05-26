package authz

import "testing"

func TestInClusterIdentity(t *testing.T) {
	host := "a1b2c3d4.alice.olares.com"
	l5d := "default.user-space-alice.serviceaccount.identity.linkerd.cluster.local"

	if d := InClusterIdentity("other.example.com", nil); d.Action != ActionPass {
		t.Fatalf("non-shared pass: %v", d.Action)
	}
	if d := InClusterIdentity(host, map[string]string{}); d.Action != ActionPass {
		t.Fatalf("missing l5d pass: %v", d.Action)
	}
	if d := InClusterIdentity(host, map[string]string{"l5d-client-id": "bad"}); d.Action != ActionDeny {
		t.Fatalf("bad l5d deny: %v", d)
	}
	if d := InClusterIdentity(host, map[string]string{"l5d-client-id": l5d}); d.Action != ActionAllow || d.Viewer != "alice" {
		t.Fatalf("valid l5d: %+v", d)
	}
	if d := InClusterIdentity("a1b2c3d4.bob.olares.com", map[string]string{"l5d-client-id": l5d}); d.Action != ActionDeny || d.Code != CodeInvalidHostUser {
		t.Fatalf("viewer mismatch: %+v", d)
	}
}

func TestInClusterSharedAllow(t *testing.T) {
	if d := InClusterSharedAllow("a1b2c3d4.alice.olares.com"); d.Action != ActionAllow {
		t.Fatalf("shared allow: %v", d.Action)
	}
	if d := InClusterSharedAllow("api.example.com"); d.Action != ActionPass {
		t.Fatalf("other pass: %v", d.Action)
	}
}

func TestHeadersWithDerivedUser(t *testing.T) {
	h := map[string]string{"other": "1"}
	out := HeadersWithDerivedUser(h, Decision{Viewer: "alice"})
	if out["x-bfl-user"] != "alice" {
		t.Fatalf("injected = %q", out["x-bfl-user"])
	}
	if h["x-bfl-user"] != "" {
		t.Fatal("must not mutate input map")
	}
	existing := map[string]string{"X-BFL-USER": "bob"}
	if out := HeadersWithDerivedUser(existing, Decision{Viewer: "alice"}); out["X-BFL-USER"] != "bob" {
		t.Fatal("must not override existing header")
	}
	if out := HeadersWithDerivedUser(h, Decision{}); len(out) != len(h) {
		t.Fatal("empty viewer must pass through")
	}
}
