package router

import (
	"strings"
	"testing"
)

// A quota needs exactly one scope, and counting is how both halves of that are
// enforced: none of them means there is nothing to cap, two means the caller
// meant one and typed the other as well.
func TestAScopeIsCountedNotGuessed(t *testing.T) {
	cases := []struct {
		name string
		refs quotaRefs
		want int
	}{
		{"nothing", quotaRefs{}, 0},
		{"only whitespace is nothing", quotaRefs{Key: "  ", User: "\t"}, 0},
		{"one", quotaRefs{Key: "ci"}, 1},
		{"one of each other kind", quotaRefs{App: "wise"}, 1},
		{"two", quotaRefs{Key: "ci", User: "alice"}, 2},
		{"all four", quotaRefs{Key: "ci", User: "alice", Model: "openai/gpt-4o", App: "wise"}, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.refs.given(); got != tc.want {
				t.Errorf("given() = %d, want %d", got, tc.want)
			}
		})
	}
}

// The kind is stored as a wire constant and read by a person. Printing
// max_budget_usd into a sentence about a ceiling is the sort of thing that
// makes a reader wonder whether it is a different setting.
func TestAQuotaKindReadsAsWhatItLimits(t *testing.T) {
	for kind, want := range map[string]string{
		quotaBudget: "spend",
		quotaRPM:    "requests/min",
		quotaTPM:    "tokens/min",
	} {
		if got := quotaKindLabel(kind); got != want {
			t.Errorf("quotaKindLabel(%q) = %q, want %q", kind, got, want)
		}
	}
	// A kind Router adds next is printed as itself rather than dropped.
	if got := quotaKindLabel("max_images_per_day"); got != "max_images_per_day" {
		t.Errorf("an unknown kind rendered as %q instead of itself", got)
	}
}

// Only the budget is money, and only the budget gets the sign. A rate limit
// with a dollar sign in front of it reads as a cost.
func TestOnlyASpendCeilingIsPrintedAsMoney(t *testing.T) {
	budget := quotaRow{QuotaType: quotaBudget, LimitValue: 50}
	if got := quotaLimitLabel(&budget); got != "$50" {
		t.Errorf("quotaLimitLabel(budget) = %q, want $50", got)
	}
	rpm := quotaRow{QuotaType: quotaRPM, LimitValue: 600}
	if got := quotaLimitLabel(&rpm); got != "600" {
		t.Errorf("quotaLimitLabel(rpm) = %q, want a bare number", got)
	}
	fractional := quotaRow{QuotaType: quotaBudget, LimitValue: 12.5}
	if got := quotaLimitLabel(&fractional); got != "$12.5" {
		t.Errorf("quotaLimitLabel(12.5) = %q", got)
	}
}

// The label is what a message tells somebody to type, so it has to be the flag
// as spelled on the command line rather than the wire constant behind it.
func TestTheFlagNameForAKindIsTheOneThatWorks(t *testing.T) {
	for kind, want := range map[string]string{
		quotaBudget: "budget",
		quotaRPM:    "rpm",
		quotaTPM:    "tpm",
	} {
		if got := flagForQuotaKind(kind); got != want {
			t.Errorf("flagForQuotaKind(%q) = %q, want %q", kind, got, want)
		}
	}
}

// The scope flags are named in one string and offered whenever a caller gives
// none, so it has to stay in step with the flags that exist.
func TestTheScopeFlagListNamesEveryScope(t *testing.T) {
	for _, flag := range []string{"--key", "--user", "--model", "--caller-app"} {
		if !strings.Contains(quotaScopeFlags, flag) {
			t.Errorf("quotaScopeFlags does not offer %s: %q", flag, quotaScopeFlags)
		}
	}
}
