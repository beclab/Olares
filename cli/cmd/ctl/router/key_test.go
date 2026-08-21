package router

import (
	"strings"
	"testing"
	"time"
)

// An expired key is active in the raw row and does not work. Reporting the
// status verbatim would put "active" beside a key every call is refusing, and
// the reader would go looking for the refusal somewhere else entirely.
func TestKeyStateFoldsExpiryIntoTheAnswer(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	cases := []struct {
		name string
		key  apiKeyView
		want string
	}{
		{"plain active", apiKeyView{Status: "active"}, "active"},
		{"expired but still marked active", apiKeyView{Status: "active",
			IsExpired: true, ExpiresAt: &past}, "expired"},
		{"disabled outranks expiry", apiKeyView{Status: "disabled", IsExpired: true}, "disabled"},
		{"nothing at all", apiKeyView{}, "-"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := keyState(&tc.key); got != tc.want {
				t.Errorf("keyState() = %q, want %q", got, tc.want)
			}
		})
	}
}

// "never" is the honest word for a key with no expiry and no use. A blank cell
// reads as missing data, which invites somebody to go looking for it.
func TestAKeyWithNoExpiryOrUseSaysNever(t *testing.T) {
	k := apiKeyView{}
	if got := keyExpiry(&k); got != "never" {
		t.Errorf("keyExpiry() = %q", got)
	}
	if got := keyLastUsed(&k); got != "never" {
		t.Errorf("keyLastUsed() = %q", got)
	}
	when := time.Date(2027, 1, 2, 3, 4, 0, 0, time.UTC)
	k = apiKeyView{ExpiresAt: &when, LastUsedAt: &when}
	if got := keyExpiry(&k); got == "never" {
		t.Error("keyExpiry() said never for a key that has one")
	}
	if got := keyLastUsed(&k); got == "never" {
		t.Error("keyLastUsed() said never for a key that has been used")
	}
}

// The prefix has one job: matching a saved key against a row in `key list`
// without reproducing the secret.
func TestTheKeyPrefixShowsEnoughToMatchAndNoMore(t *testing.T) {
	const secret = "sk-abcdefghijklmnopqrstuvwxyz"
	got := keyPrefixOf(secret)
	shown := strings.TrimSuffix(got, "…")
	if got == shown {
		t.Errorf("keyPrefixOf(%q) = %q, with nothing to say it was cut", secret, got)
	}
	if !strings.HasPrefix(secret, shown) {
		t.Errorf("keyPrefixOf(%q) = %q, which is not a prefix of it", secret, got)
	}
	if len([]rune(shown)) != 11 {
		t.Errorf("keyPrefixOf showed %d characters, want the 11 Router's own prefix is",
			len([]rune(shown)))
	}
	short := "sk-abc"
	if got := keyPrefixOf(short); got != short {
		t.Errorf("keyPrefixOf(%q) = %q, want it unchanged", short, got)
	}
}

// A key's lifetime is stated in days, and 720h is nobody's first thought.
func TestTTLTakesDaysAsWellAsGoDurations(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"30d", 30 * 24 * time.Hour},
		{"1d", 24 * time.Hour},
		{"0.5d", 12 * time.Hour},
		{"12h", 12 * time.Hour},
		{"90m", 90 * time.Minute},
		{" 7D ", 7 * 24 * time.Hour},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parseTTL(tc.in)
			if err != nil {
				t.Fatalf("parseTTL(%q): %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("parseTTL(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
	for _, bad := range []string{"", "soon", "0d", "-5d", "-1h", "0h", "d"} {
		if _, err := parseTTL(bad); err == nil {
			t.Errorf("parseTTL(%q) accepted it", bad)
		}
	}
}

// A deadline is typed as a date. Refusing it and asking for RFC3339 is the
// kind of thing that makes a person go and look up the spelling of a timezone
// offset for a key that expires at midnight anyway.
func TestAnInstantTakesABareDate(t *testing.T) {
	got, err := parseInstant("2027-01-01")
	if err != nil {
		t.Fatalf("parseInstant: %v", err)
	}
	if got.Year() != 2027 || got.Month() != time.January || got.Day() != 1 {
		t.Errorf("parseInstant(\"2027-01-01\") = %v", got)
	}
	if h, m, s := got.Clock(); h != 0 || m != 0 || s != 0 {
		t.Errorf("a bare date is not the start of the day: %v", got)
	}
	if _, err := parseInstant("2027-01-01T12:30:00Z"); err != nil {
		t.Errorf("parseInstant refused RFC3339: %v", err)
	}
	for _, bad := range []string{"", "tomorrow", "01/01/2027"} {
		if _, err := parseInstant(bad); err == nil {
			t.Errorf("parseInstant(%q) accepted it", bad)
		}
	}
}

// The example in an error message has to be a name that could work. Falling
// back to a plausible one beats printing an empty string into the sentence.
func TestTheExampleModelPrefersARealOne(t *testing.T) {
	if got := exampleQualified([]string{"openai/gpt-4o", "anthropic/claude"}); got != "openai/gpt-4o" {
		t.Errorf("exampleQualified() = %q, want the first configured model", got)
	}
	if got := exampleQualified(nil); !strings.Contains(got, "/") {
		t.Errorf("exampleQualified(nil) = %q, which is not a qualified name", got)
	}
}
