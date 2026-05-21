package authz

import "testing"

func TestHostUser_Disabled_Passes(t *testing.T) {
	d := HostUser("anything", nil, HostUserConfig{Enabled: false})
	if d.Action != ActionPass {
		t.Fatalf("disabled must Pass, got %+v", d)
	}
}

func TestHostUser_Allow_HappyPath(t *testing.T) {
	d := HostUser("01234567.brucedai.olares.com",
		map[string]string{"x-bfl-user": "brucedai"},
		DefaultHostUserConfig())
	if d.Action != ActionAllow {
		t.Fatalf("expected Allow, got %+v", d)
	}
	if d.Viewer != "brucedai" || d.Username != "brucedai" {
		t.Fatalf("viewer/username mismatch: %+v", d)
	}
}

func TestHostUser_CaseInsensitiveCompare(t *testing.T) {
	d := HostUser("01234567.BRUCEDAI.olares.com",
		map[string]string{"x-bfl-user": "BruceDai"},
		DefaultHostUserConfig())
	if d.Action != ActionAllow {
		t.Fatalf("expected Allow, got %+v", d)
	}
}

func TestHostUser_HostWithPort_Allow(t *testing.T) {
	d := HostUser("01234567.alice.olares.com:443",
		map[string]string{"x-bfl-user": "alice"},
		DefaultHostUserConfig())
	if d.Action != ActionAllow {
		t.Fatalf("expected Allow despite :443, got %+v", d)
	}
}

func TestHostUser_DenyMissingXBflUser(t *testing.T) {
	d := HostUser("01234567.brucedai.olares.com", map[string]string{}, DefaultHostUserConfig())
	if d.Action != ActionDeny || d.Code != "INVALID_HOST_USER" {
		t.Fatalf("expected Deny INVALID_HOST_USER, got %+v", d)
	}
}

func TestHostUser_DenyMismatch(t *testing.T) {
	d := HostUser("01234567.brucedai.olares.com",
		map[string]string{"x-bfl-user": "alice"},
		DefaultHostUserConfig())
	if d.Action != ActionDeny || d.Code != "INVALID_HOST_USER" {
		t.Fatalf("expected Deny INVALID_HOST_USER, got %+v", d)
	}
	if d.Viewer != "brucedai" || d.Username != "alice" {
		t.Fatalf("decision must carry both sides: %+v", d)
	}
}

func TestHostUser_DenyTooFewLabels(t *testing.T) {
	d := HostUser("olares.com",
		map[string]string{"x-bfl-user": "brucedai"},
		DefaultHostUserConfig())
	if d.Action != ActionDeny || d.Code != "INVALID_HOST_USER" {
		t.Fatalf("expected Deny INVALID_HOST_USER for 2-label host, got %+v", d)
	}
}

func TestHostUser_NonV2Host_Passes(t *testing.T) {
	for _, host := range []string{"demo.agw.local", "ZZZZZZZZ.alice.olares.com"} {
		d := HostUser(host, map[string]string{}, DefaultHostUserConfig())
		if d.Action != ActionPass {
			t.Fatalf("non-v2 host %q should Pass to allow-all baseline, got %+v", host, d)
		}
	}
}

func TestHostUser_SkipPrefix(t *testing.T) {
	cfg := HostUserConfig{Enabled: true, SkipPrefixes: []string{"admin"}}
	d := HostUser("01234567.admin.olares.com",
		map[string]string{"x-bfl-user": "service-account"},
		cfg)
	if d.Action != ActionAllow {
		t.Fatalf("expected Allow via skip-list, got %+v", d)
	}
}

func TestHostUser_EmptyAuthority(t *testing.T) {
	d := HostUser("", map[string]string{"x-bfl-user": "alice"}, DefaultHostUserConfig())
	if d.Action != ActionDeny || d.Code != "INVALID_HOST_USER" {
		t.Fatalf("expected Deny INVALID_HOST_USER for empty :authority, got %+v", d)
	}
}

func TestNormalizeHost(t *testing.T) {
	cases := map[string]string{
		"":                             "",
		"  Foo.Bar.com  ":              "foo.bar.com",
		"Foo.Bar.com:443":              "foo.bar.com",
		"Foo.Bar.com.":                 "foo.bar.com",
		"olares.com:NOTANUMBER":        "olares.com:notanumber",
		"01234567.alice.olares.com:65": "01234567.alice.olares.com",
	}
	for in, want := range cases {
		if got := normalizeHost(in); got != want {
			t.Fatalf("normalizeHost(%q) = %q, want %q", in, got, want)
		}
	}
}
