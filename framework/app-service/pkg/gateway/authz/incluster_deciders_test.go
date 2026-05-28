package authz

import "testing"

func TestInClusterIdentity(t *testing.T) {
	host := "a1b2c3d4.alice.olares.com"
	l5d := "default.user-space-alice.serviceaccount.identity.linkerd.cluster.local"

	if d := InClusterIdentity("other.example.com", nil, nil); d.Action != ActionPass {
		t.Fatalf("non-shared pass: %v", d.Action)
	}
	if d := InClusterIdentity(host, map[string]string{}, nil); d.Action != ActionPass {
		t.Fatalf("missing l5d pass: %v", d.Action)
	}
	if d := InClusterIdentity(host, map[string]string{"l5d-client-id": "bad"}, nil); d.Action != ActionDeny {
		t.Fatalf("bad l5d deny: %v", d)
	}
	if d := InClusterIdentity(host, map[string]string{"l5d-client-id": l5d}, nil); d.Action != ActionAllow || d.Viewer != "alice" {
		t.Fatalf("valid l5d: %+v", d)
	}
	if d := InClusterIdentity("a1b2c3d4.bob.olares.com", map[string]string{"l5d-client-id": l5d}, nil); d.Action != ActionDeny || d.Code != CodeInvalidHostUser {
		t.Fatalf("viewer mismatch: %+v", d)
	}
}

// requirement: WI-27 closes the LiteLLM G-B gap — when the caller namespace is
// of the <app>-<user> form (e.g. litellm-brucedai), InClusterIdentity must
// resolve the viewer via the app_user_fallback path once knownUsers contains
// the trailing user segment.
// behavior: knownUsers={brucedai}, l5d=litellm-brucedai.sa, host=brucedai ->
// Allow with viewer=brucedai (matches host viewer label).
// test: TC-038.
func TestInClusterIdentity_AppUserFallback_Allow_TC038(t *testing.T) {
	host := "a1b2c3d4.brucedai.olares.com"
	l5d := "default.litellm-brucedai.serviceaccount.identity.linkerd.cluster.local"
	known := map[string]struct{}{"brucedai": {}}
	d := InClusterIdentity(host, map[string]string{"l5d-client-id": l5d}, known)
	if d.Action != ActionAllow || d.Viewer != "brucedai" || d.Username != "brucedai" {
		t.Fatalf("TC-038: %+v want Allow viewer=brucedai", d)
	}
}

// requirement: nil knownUsers must keep the legacy InClusterIdentity behavior
// so existing callers that have not been migrated still see the pre-WI-27
// Deny for <app>-<user> caller namespaces.
// behavior: knownUsers=nil + caller_ns=litellm-brucedai -> Deny
// INVALID_CALLER_IDENTITY (path 4 disabled, paths 2/3 miss).
// test: TC-039.
func TestInClusterIdentity_NilKnownUsers_Deny_TC039(t *testing.T) {
	host := "a1b2c3d4.brucedai.olares.com"
	l5d := "default.litellm-brucedai.serviceaccount.identity.linkerd.cluster.local"
	d := InClusterIdentity(host, map[string]string{"l5d-client-id": l5d}, nil)
	if d.Action != ActionDeny || d.Code != CodeInvalidCallerIdentity {
		t.Fatalf("TC-039: %+v want Deny INVALID_CALLER_IDENTITY", d)
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
