package router

import (
	"bytes"
	"strings"
	"testing"
)

// The note only appears on an instance still running the pre-rename listing,
// which is the one case that cannot be reached from a machine on the current
// one — so it is asserted here rather than left to a live run.
func TestRouterNoteOnlyForThePreRenameID(t *testing.T) {
	t.Parallel()

	t.Run("current listing says nothing", func(t *testing.T) {
		if note := routerNote(&discoveredRouter{AppName: "router"}); note != "" {
			t.Fatalf("note = %q, want empty", note)
		}
	})

	t.Run("pre-rename listing explains the backend", func(t *testing.T) {
		note := routerNote(&discoveredRouter{AppName: legacyRouterAppName})
		if !strings.Contains(note, "2.0.x") {
			t.Fatalf("note does not name the backend line: %q", note)
		}
	})

	t.Run("no discovery, no note", func(t *testing.T) {
		if note := routerNote(nil); note != "" {
			t.Fatalf("note = %q, want empty", note)
		}
	})
}

func TestRenderStatusPrintsTheNote(t *testing.T) {
	t.Parallel()

	report := statusReport{
		Router: &discoveredRouter{AppName: legacyRouterAppName, State: "running"},
		Note:   routerNote(&discoveredRouter{AppName: legacyRouterAppName}),
	}
	var out bytes.Buffer
	if err := renderStatus(&out, report); err != nil {
		t.Fatalf("renderStatus: %v", err)
	}
	if !strings.Contains(out.String(), "NOTE") {
		t.Fatalf("rendered status has no NOTE row:\n%s", out.String())
	}
}
