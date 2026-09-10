package router

import "testing"

// The three renderings of a phase used to live in three functions, and a phase
// added to one was a phase missing from the others. Now they are one row, and
// a row filled in halfway is the same bug wearing a different shape: a cell
// left empty prints "no · " with nothing after it.
func TestEveryPhaseIsRenderedThreeWays(t *testing.T) {
	for name, p := range modelPhases {
		if p.sentence == "" {
			t.Errorf("%q has no sentence: `model status` would say nothing about it", name)
		}
		if p.readiness == "" {
			t.Errorf("%q has no readiness: `model list -o json` would report an empty string", name)
		}
		switch {
		case !p.servable && p.cell == "":
			t.Errorf("%q refuses calls and has no cell: `model list` would print a bare \"no · \"", name)
		case p.servable && p.cell != "":
			t.Errorf("%q is servable and carries a cell %q, which nothing will ever print", name, p.cell)
		}
	}
}

// servable and readiness are separate axes, but not independent ones: a phase
// that dispatches and reports anything but ready would have `model list`
// promising a call it also says is not warmed up.
func TestServabilityAndReadinessAgree(t *testing.T) {
	for name, p := range modelPhases {
		if p.servable && p.readiness != "ready" {
			t.Errorf("%q dispatches calls but reports readiness %q", name, p.readiness)
		}
		if !p.servable && p.readiness == "ready" {
			t.Errorf("%q reports ready but refuses calls", name)
		}
	}
}

// The vocabulary belongs to llm-init, and a word this build has not been told
// about must not be read as bad news. Three readers depend on that separately,
// so it is asserted on the lookup they share.
func TestAnUnknownPhaseIsTreatedAsWorking(t *testing.T) {
	if phaseBlocksCalls("quiescing") {
		t.Error("a phase this build has never heard of was taken as a reason to hide the model")
	}
	if note := phaseNote("quiescing", nil); note != "" {
		t.Errorf("invented a meaning for an unknown phase: %q", note)
	}
	running, unknownPhase := "running", "quiescing"
	row := adminModelRow{ProviderSource: "olares",
		ProviderOlaresStatus: &running, ModelConsoleStatus: &unknownPhase}
	if got := row.readiness(); got != "ready" {
		t.Errorf("readiness of an unknown phase is %q, want ready", got)
	}
}

// Whitespace and case come off the wire, not out of a literal.
func TestAPhaseIsMatchedRegardlessOfSpelling(t *testing.T) {
	for _, spelling := range []string{"download", "Download", " download ", "DOWNLOAD"} {
		if !phaseBlocksCalls(spelling) {
			t.Errorf("%q was not recognised as download", spelling)
		}
	}
}
