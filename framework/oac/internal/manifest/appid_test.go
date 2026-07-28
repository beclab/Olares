package manifest

import (
	"strings"
	"testing"
)

// TestAppIDFromName pins the loader-normalization contract: a non-empty name
// yields the first 8 hex characters of md5(name); an empty name yields the
// empty string. Cross-checking against a couple of known md5 prefixes
// guards against accidental hash-truncation bugs (e.g. taking the last 8
// chars or hashing UTF-8 bytes differently).
func TestAppIDFromName(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"", ""},
		// md5("nginx")   = "ee434023cf89d7dfb21f63d64f0f9d74"
		{"nginx", "ee434023"},
		// md5("firefox") = "d6a5c9544eca9b5ce2266d1c34a93222"
		{"firefox", "d6a5c954"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := AppIDFromName(tc.name)
			if got != tc.want {
				t.Fatalf("AppIDFromName(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestAppIDFromName_DeterministicAndLength(t *testing.T) {
	id := AppIDFromName("my-app")
	if len(id) != 8 {
		t.Fatalf("non-empty name must yield an 8-char id, got %q (len=%d)", id, len(id))
	}
	if id != AppIDFromName("my-app") {
		t.Fatal("AppIDFromName must be deterministic across calls")
	}
	if id == AppIDFromName("my-other-app") {
		t.Fatal("distinct names must yield distinct ids")
	}
}

// TestIsReservedSystemAppID covers a representative slice of the reserved
// set (a head, a tail, and the hyphenated middle entry) plus a clearly
// non-reserved name. Whitespace handling is pinned because the lint rule
// trims before checking.
func TestIsReservedSystemAppID(t *testing.T) {
	for _, s := range []string{
		"market", "auth", "search-admin", "olares-app", "control-hub",
	} {
		if !IsReservedSystemAppID(s) {
			t.Errorf("IsReservedSystemAppID(%q) = false, want true", s)
		}
	}
	if !IsReservedSystemAppID("  market  ") {
		t.Error("IsReservedSystemAppID must trim surrounding whitespace")
	}
	for _, s := range []string{
		"", "my-app", "MARKET", " market2", "marketing",
	} {
		if IsReservedSystemAppID(s) {
			t.Errorf("IsReservedSystemAppID(%q) = true, want false", s)
		}
	}
}

// Lint rule: an empty metadata.appid is permitted (the loader will fill it
// in deterministically). The baseline fixture omits the field, so it must
// continue to pass.
func TestValidateMetadataAppID_EmptyAllowed(t *testing.T) {
	c := newValidConfig()
	c.Metadata.AppID = ""
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("empty metadata.appid must be accepted, got: %v", err)
	}
}

// Lint rule: a non-reserved metadata.appid (the existing "hand-author" shape
// from olares-chart-from-compose -- appid==name) must pass even though the
// loader will overwrite it later.
func TestValidateMetadataAppID_CustomAccepted(t *testing.T) {
	c := newValidConfig()
	c.Metadata.AppID = "firefox"
	if err := ValidateAppConfiguration(c); err != nil {
		t.Fatalf("non-reserved metadata.appid must be accepted, got: %v", err)
	}
}

// Lint rule: every reserved system appid must be rejected, and the error
// must mention metadata.appid plus the offending value so the user can
// spot what to change.
func TestValidateMetadataAppID_ReservedRejected(t *testing.T) {
	for _, reserved := range []string{
		"market", "auth", "settings", "search-admin", "olares-app", "control-hub", "nitro",
	} {
		reserved := reserved
		t.Run(reserved, func(t *testing.T) {
			c := newValidConfig()
			c.Metadata.AppID = reserved
			err := ValidateAppConfiguration(c)
			if err == nil {
				t.Fatalf("metadata.appid=%q must be rejected as reserved", reserved)
			}
			if !strings.Contains(err.Error(), "metadata.appid") {
				t.Fatalf("error must mention metadata.appid, got: %v", err)
			}
			if !strings.Contains(err.Error(), reserved) {
				t.Fatalf("error must echo the offending value %q, got: %v", reserved, err)
			}
		})
	}
}
