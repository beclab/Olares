package router

import (
	"bytes"
	"strings"
	"testing"
)

// A category Router maintains and one an admin pinned answer the same way
// today and make different promises about tomorrow, so a view that showed only
// the answer would be showing half of it.
func TestACategorySaysWhoChoseWhatItAnswersWith(t *testing.T) {
	cases := map[string]struct {
		route modelRoute
		want  string
	}{
		"reconciliation picked it": {modelRoute{Kind: routeKindDefault}, "Router"},
		"an admin pinned it":       {modelRoute{Kind: routeKindDefault, TargetPinned: true}, "admin"},
		"not a category at all":    {modelRoute{Kind: routeKindGroup}, "-"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := tc.route.chosenBy(); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// The cost of pinning is that the category stops following what is installed,
// in both directions. Somebody reading the route has to be told.
func TestAPinnedCategoryExplainsWhatItNoLongerDoes(t *testing.T) {
	var buf bytes.Buffer
	err := renderRoute(&buf, &modelRoute{
		Name: "default-chat", Kind: routeKindDefault, Mode: "chat", Enabled: true, TargetPinned: true,
		Target:  &routeTarget{ProviderModelID: "pm_1"},
		Members: []routeMember{{QualifiedName: "Olares/qwen3-4b", ProviderModelID: "pm_1", Servable: true}},
	})
	if err != nil {
		t.Fatalf("renderRoute: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"CHOSEN BY", "admin", "route unpin default-chat"} {
		if !strings.Contains(out, want) {
			t.Fatalf("should mention %q, got:\n%s", want, out)
		}
	}
}

// A category Router still maintains has nothing to warn about.
func TestAnUnpinnedCategoryCarriesNoWarning(t *testing.T) {
	var buf bytes.Buffer
	err := renderRoute(&buf, &modelRoute{
		Name: "default-chat", Kind: routeKindDefault, Mode: "chat", Enabled: true,
		Members: []routeMember{{QualifiedName: "Olares/qwen3-4b", Servable: true}},
	})
	if err != nil {
		t.Fatalf("renderRoute: %v", err)
	}
	if strings.Contains(buf.String(), "route unpin") {
		t.Fatalf("nothing is pinned here:\n%s", buf.String())
	}
}

// Pinning a model and then deleting it leaves a category nothing will refill,
// which reads exactly like an empty one and is not the same problem.
func TestAPinnedCategoryWithNothingBehindItSaysWhyNothingWillFillIt(t *testing.T) {
	var buf bytes.Buffer
	err := renderRoute(&buf, &modelRoute{
		Name: "default-chat", Kind: routeKindDefault, Mode: "chat", Enabled: true, TargetPinned: true,
	})
	if err != nil {
		t.Fatalf("renderRoute: %v", err)
	}
	if !strings.Contains(buf.String(), "no longer exists") {
		t.Fatalf("should say the pin is what keeps it empty, got:\n%s", buf.String())
	}
}

func TestOnlyACategoryTakesAPin(t *testing.T) {
	alias := &modelRoute{Name: "fast", Kind: routeKindAlias, Mode: "chat"}
	if err := refuseNonDefaultTarget(alias, "pin"); err == nil {
		t.Fatal("an alias names one model by definition")
	} else if !strings.Contains(err.Error(), "route create fast") {
		t.Fatalf("should point at how an alias is repointed, got: %v", err)
	}
	group := &modelRoute{Name: "house", Kind: routeKindGroup, Mode: "chat"}
	if err := refuseNonDefaultTarget(group, "pin"); err == nil {
		t.Fatal("a group's answer is its membership")
	} else if !strings.Contains(err.Error(), "route add house") {
		t.Fatalf("should point at the membership verb, got: %v", err)
	}
	if err := refuseNonDefaultTarget(&modelRoute{Kind: routeKindDefault}, "pin"); err != nil {
		t.Fatalf("a category is exactly what takes a pin: %v", err)
	}
}

// Candidates are not filtered to what is running: pinning before starting the
// application is a legitimate order to do things in, so a stopped model has to
// be listed and marked rather than hidden.
func TestCandidatesShowAStoppedModelRatherThanHidingIt(t *testing.T) {
	var buf bytes.Buffer
	app := "llamacppqwen34b"
	route := &modelRoute{Name: "default-chat", Kind: routeKindDefault, TargetPinned: true,
		Target: &routeTarget{ProviderModelID: "pm_2"}}
	err := renderRouteCandidates(&buf, route, []routeMember{
		{QualifiedName: "Olares/qwen3-4b", ProviderModelID: "pm_1", Servable: false,
			OlaresAppName: &app},
		{QualifiedName: "Olares/qwen3-8b", ProviderModelID: "pm_2", Servable: true},
	})
	if err != nil {
		t.Fatalf("renderRouteCandidates: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Olares/qwen3-4b") {
		t.Fatalf("a stopped candidate should still be listed:\n%s", out)
	}
	if !strings.Contains(out, "llamacppqwen34b") {
		t.Fatalf("a local model is named by its application:\n%s", out)
	}
	if !strings.Contains(out, "route pin default-chat") {
		t.Fatalf("the list should say what to do with it:\n%s", out)
	}
}

func TestACategoryNothingCanServeHasNothingToPin(t *testing.T) {
	var buf bytes.Buffer
	route := &modelRoute{Name: "default-tts-clone", Kind: routeKindDefault}
	if err := renderRouteCandidates(&buf, route, nil); err != nil {
		t.Fatalf("renderRouteCandidates: %v", err)
	}
	if !strings.Contains(buf.String(), "nothing to pin") {
		t.Fatalf("got:\n%s", buf.String())
	}
}

// The category list is where somebody notices a pin they forgot about.
func TestTheCategoryListMarksTheOnesAnAdminChose(t *testing.T) {
	var buf bytes.Buffer
	err := renderDefaults(&buf, []modelRoute{
		{Name: "default-chat", Kind: routeKindDefault, Mode: "chat", Enabled: true, TargetPinned: true,
			Members: []routeMember{{QualifiedName: "Olares/qwen3-4b", Servable: true}}},
		{Name: "default-embedding", Kind: routeKindDefault, Mode: "embedding", Enabled: true,
			Members: []routeMember{{QualifiedName: "Olares/bge-m3", Servable: true}}},
	})
	if err != nil {
		t.Fatalf("renderDefaults: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "CHOSEN BY") {
		t.Fatalf("missing the column:\n%s", out)
	}
	if !strings.Contains(out, "route unpin") {
		t.Fatalf("a list with a pin in it should say how to undo one:\n%s", out)
	}
}
