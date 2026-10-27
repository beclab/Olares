package dashboard

import (
	"errors"
	"testing"

	"github.com/beclab/Olares/cli/pkg/clierr"
)

// TestEverySectionFailed pins the predicate the aggregate overview verbs
// use to decide their exit code. Partial degradation must stay exit 0 so
// one dead section does not hide the others; a gated or empty section is
// not a failure at all.
func TestEverySectionFailed(t *testing.T) {
	failed := Envelope{Meta: Meta{Error: "HTTP 530: Olares connection error"}}
	ok := Envelope{Items: []Item{{Raw: map[string]any{"used": 1}}}}
	gated := Envelope{Meta: Meta{Empty: true, EmptyReason: "device_not_olares_one"}}

	cases := []struct {
		name     string
		sections map[string]Envelope
		want     bool
	}{
		{"no sections at all", nil, false},
		{"every section failed", map[string]Envelope{"physical": failed, "user": failed}, true},
		{"one section survived", map[string]Envelope{"physical": failed, "user": ok}, false},
		{"gated section is not a failure", map[string]Envelope{"live": gated, "curve": gated}, false},
		{"gated alongside failed still has data-free but non-error section", map[string]Envelope{"live": gated, "curve": failed}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EverySectionFailed(Envelope{Sections: tc.sections})
			if got != tc.want {
				t.Errorf("EverySectionFailed = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestErrAlreadyReportedIsTheSharedSentinel guards the wiring that stops
// "Error: (already reported)" from reaching a user: cmd/main.go matches
// on clierr.ErrAlreadyReported, so the dashboard sentinel has to wrap it
// while keeping a message of its own.
func TestErrAlreadyReportedIsTheSharedSentinel(t *testing.T) {
	if !errors.Is(ErrAlreadyReported, clierr.ErrAlreadyReported) {
		t.Fatal("dashboard.ErrAlreadyReported must satisfy errors.Is against clierr.ErrAlreadyReported")
	}
}
