package router

import (
	"strings"
	"testing"
)

// Every locally installed model application is a provider called `Olares`, so
// the routing name identifies none of them. title is what a reader sees and
// handle is what they type back, and the two differ on purpose: a title reads
// well and an application name is what a reference accepts.
func TestAProviderIsNamedByWhicheverHandleIdentifiesIt(t *testing.T) {
	display, app := "Qwen3 8B", "qwen3-8b-abc"
	cases := []struct {
		name              string
		row               providerRow
		wantTitle, wantIn string
	}{
		{"a hand-entered provider is its own name",
			providerRow{Name: "openai-main"}, "openai-main", "openai-main"},
		{"an application prefers its display title to read",
			providerRow{Name: "Olares", ProviderDisplayTitle: &display, OlaresAppName: &app},
			"Qwen3 8B", "qwen3-8b-abc"},
		{"an application with no title falls back to its own name",
			providerRow{Name: "Olares", OlaresAppName: &app}, "qwen3-8b-abc", "qwen3-8b-abc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.row.title(); got != tc.wantTitle {
				t.Errorf("title() = %q, want %q", got, tc.wantTitle)
			}
			if got := tc.row.handle(); got != tc.wantIn {
				t.Errorf("handle() = %q, want %q", got, tc.wantIn)
			}
		})
	}
}

// handle must never answer `Olares` for a model application. That is the one
// value that would send the reader back with an argument naming every
// application at once.
func TestTheHandleOfAModelApplicationIsNeverTheSharedName(t *testing.T) {
	display, app := "Qwen3 8B", "qwen3-8b-abc"
	row := providerRow{Name: "Olares", ProviderDisplayTitle: &display, OlaresAppName: &app}
	if got := row.handle(); got == "Olares" {
		t.Error("handle() answered the name every model application shares")
	}
	// An empty application name is not a handle either.
	empty := ""
	row = providerRow{Name: "openai-main", OlaresAppName: &empty}
	if got := row.handle(); got != "openai-main" {
		t.Errorf("an empty application name became the handle: %q", got)
	}
}

func TestAMarketSourcedProviderIsRecognisedWhateverTheCase(t *testing.T) {
	for _, source := range []string{"olares", "Olares", "OLARES"} {
		row := providerRow{Source: source}
		if !row.isMarketSourced() {
			t.Errorf("source %q was not recognised as Market-sourced", source)
		}
	}
	for _, source := range []string{"manual", "", "olares-ish"} {
		row := providerRow{Source: source}
		if row.isMarketSourced() {
			t.Errorf("source %q was taken as Market-sourced", source)
		}
	}
}

// Router answers 409 on its own. The value the CLI adds is naming the
// application, because that is what the reader has to go and act on -- a 409
// against a provider called `Olares` does not say which of them.
func TestTheMarketRefusalNamesTheApplicationToActOn(t *testing.T) {
	app := "qwen3-8b-abc"
	row := providerRow{Name: "Olares", Source: "olares", OlaresAppName: &app}
	err := marketOwnedErr(&row, "delete")
	if err == nil {
		t.Fatal("a Market-sourced provider was not refused")
	}
	msg := err.Error()
	for _, want := range []string{"qwen3-8b-abc", "delete", "olares-cli market"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal does not mention %q: %v", want, msg)
		}
	}
}

// A Market-sourced row whose application name never arrived still has to
// produce a sentence, not one with a hole in it.
func TestTheMarketRefusalReadsWithoutAnApplicationName(t *testing.T) {
	row := providerRow{Name: "Olares", Source: "olares"}
	msg := marketOwnedErr(&row, "edit").Error()
	if strings.Contains(msg, "the  application") || strings.Contains(msg, `""`) {
		t.Errorf("the refusal has a gap in it: %v", msg)
	}
	if !strings.Contains(msg, "model application") {
		t.Errorf("the refusal does not say what owns the provider: %v", msg)
	}
}
