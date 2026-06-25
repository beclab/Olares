package olares

import (
	"net/url"
	"testing"
)

func TestLocationConnectionDerivation(t *testing.T) {
	cases := []struct {
		loc          Location
		valid        bool
		localDomain  bool
		clusterDNS   bool
		scheme       string
	}{
		{LocationExternal, true, false, false, "https"},
		{LocationLAN, true, true, false, "http"},
		{LocationHost, true, false, true, "https"},
		{LocationCluster, true, false, false, "https"},
		{Location("bogus"), false, false, false, "https"},
		{Location(""), false, false, false, "https"},
	}
	for _, c := range cases {
		if got := c.loc.Valid(); got != c.valid {
			t.Errorf("%q.Valid() = %v, want %v", c.loc, got, c.valid)
		}
		if got := c.loc.UsesLocalDomain(); got != c.localDomain {
			t.Errorf("%q.UsesLocalDomain() = %v, want %v", c.loc, got, c.localDomain)
		}
		if got := c.loc.UsesClusterResolver(); got != c.clusterDNS {
			t.Errorf("%q.UsesClusterResolver() = %v, want %v", c.loc, got, c.clusterDNS)
		}
		if got := c.loc.scheme(); got != c.scheme {
			t.Errorf("%q.scheme() = %q, want %q", c.loc, got, c.scheme)
		}
	}
}

func TestEndpointsFourPositions(t *testing.T) {
	id, err := ParseID("alice@olares.com")
	if err != nil {
		t.Fatalf("ParseID: %v", err)
	}

	ext := id.Endpoints(LocationExternal, "")
	host := id.Endpoints(LocationHost, "")
	cluster := id.Endpoints(LocationCluster, "")
	lan := id.Endpoints(LocationLAN, "")

	// external / host / cluster produce byte-identical URLs — they differ only
	// in the transport's resolver, not the URL.
	if ext != host || ext != cluster {
		t.Fatalf("external/host/cluster endpoints must match:\n ext=%+v\nhost=%+v\nclus=%+v", ext, host, cluster)
	}
	if ext.Auth != "https://auth.alice.olares.com" {
		t.Errorf("external Auth = %q", ext.Auth)
	}
	if ext.Vault != "https://vault.alice.olares.com/server" {
		t.Errorf("external Vault = %q (expected /server suffix)", ext.Vault)
	}

	// lan uses plain http + the olares.local suffix.
	if lan.Auth != "http://auth.alice.olares.local" {
		t.Errorf("lan Auth = %q", lan.Auth)
	}
	if lan.Desktop != "http://desktop.alice.olares.local" {
		t.Errorf("lan Desktop = %q", lan.Desktop)
	}
}

func TestEndpointsLocalPrefix(t *testing.T) {
	id, _ := ParseID("alice@olares.com")
	ep := id.Endpoints(LocationExternal, "dev.")
	if ep.Auth != "https://auth.dev.alice.olares.com" {
		t.Errorf("Auth with localPrefix = %q", ep.Auth)
	}
}

func TestEndpointsCustomDomainLANStillOlaresLocal(t *testing.T) {
	id, _ := ParseID("bob@example.com")

	ext := id.Endpoints(LocationExternal, "")
	if ext.Auth != "https://auth.bob.example.com" {
		t.Errorf("external Auth for custom domain = %q", ext.Auth)
	}
	// The user's real domain is dropped on the LAN; only the local part is used.
	lan := id.Endpoints(LocationLAN, "")
	if lan.Auth != "http://auth.bob.olares.local" {
		t.Errorf("lan Auth for custom domain = %q, want olares.local suffix", lan.Auth)
	}
}

func TestEndpointsInvalidLocationFallsBackToExternal(t *testing.T) {
	id, _ := ParseID("alice@olares.com")
	if got := id.Endpoints(Location("nope"), ""); got != id.Endpoints(LocationExternal, "") {
		t.Errorf("invalid Location should behave as external, got %+v", got)
	}
}

func TestRebaseURL(t *testing.T) {
	id, _ := ParseID("alice@olares.com")
	src, _ := url.Parse("https://files.alice.olares.com/api/list?x=1#frag")

	// external↔host↔cluster: URL stays byte-identical.
	for _, loc := range []Location{LocationExternal, LocationHost, LocationCluster} {
		got := id.RebaseURL(src, loc, "")
		if got.String() != src.String() {
			t.Errorf("RebaseURL(%q) = %q, want unchanged %q", loc, got.String(), src.String())
		}
	}

	// lan flips scheme and domain suffix but keeps path/query/fragment.
	got := id.RebaseURL(src, LocationLAN, "")
	want := "http://files.alice.olares.local/api/list?x=1#frag"
	if got.String() != want {
		t.Errorf("RebaseURL(lan) = %q, want %q", got.String(), want)
	}

	// lan → external flips back.
	lanURL, _ := url.Parse("http://files.alice.olares.local/api/list?x=1")
	back := id.RebaseURL(lanURL, LocationExternal, "")
	if back.String() != "https://files.alice.olares.com/api/list?x=1" {
		t.Errorf("RebaseURL(lan→external) = %q", back.String())
	}

	if id.RebaseURL(nil, LocationLAN, "") != nil {
		t.Error("RebaseURL(nil) should return nil")
	}
}
