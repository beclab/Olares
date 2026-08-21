package router

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// One entry as Router sends it today, copied from a live answer. Every key here
// has to survive the trip: what this file is guarding against is the failure it
// was written after, where Router grew mode, supports and readiness, the struct
// was never taught about them, and `-o json` went on printing a five-field
// subset that looked complete.
const liveModelEntry = `{
  "id": "Olares/SenseVoiceSmall",
  "object": "model",
  "created": 1755600000,
  "owned_by": "Olares",
  "qualified_id": "Olares/SenseVoiceSmall",
  "mode": "audio",
  "supports": ["stt", "vad"],
  "readiness": "ready"
}`

func TestModelObjectDecodesEveryFieldOnTheWire(t *testing.T) {
	var got modelObject
	if err := json.Unmarshal([]byte(liveModelEntry), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := modelObject{
		ID:          "Olares/SenseVoiceSmall",
		Object:      "model",
		Created:     1755600000,
		OwnedBy:     "Olares",
		QualifiedID: "Olares/SenseVoiceSmall",
		Mode:        "audio",
		Supports:    []string{"stt", "vad"},
		Readiness:   "ready",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("decoded entry:\n got %+v\nwant %+v", got, want)
	}
}

// An unknown field is dropped silently by encoding/json, so the only way a
// missing one announces itself is a comparison of key sets. `-o json` re-encodes
// the struct rather than passing the body through, which is what turns a field
// nobody modelled into a field the caller never sees.
func TestNoWireFieldIsLostOnTheWayToJSONOutput(t *testing.T) {
	var wire map[string]json.RawMessage
	if err := json.Unmarshal([]byte(liveModelEntry), &wire); err != nil {
		t.Fatalf("decode into a map: %v", err)
	}
	var decoded modelObject
	if err := json.Unmarshal([]byte(liveModelEntry), &decoded); err != nil {
		t.Fatalf("decode into the struct: %v", err)
	}
	reEncoded, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("re-encode: %v", err)
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(reEncoded, &out); err != nil {
		t.Fatalf("decode the re-encoded entry: %v", err)
	}
	for key := range wire {
		if _, ok := out[key]; !ok {
			t.Errorf("%q reached the CLI and did not reach its output; add it to modelObject", key)
		}
	}
	if len(out) != len(wire) {
		t.Errorf("output carries %d fields against %d on the wire: %s", len(out), len(wire), reEncoded)
	}
}

func TestSummarizeSupportNames(t *testing.T) {
	cases := []struct {
		name  string
		given []string
		want  string
	}{
		// A blank cell reads as missing data; a model with no declared
		// capability is a different thing and says so.
		{"nothing declared", nil, "-"},
		{"an empty list", []string{}, "-"},
		{"blanks only", []string{"", "  "}, "-"},
		{"under the cap", []string{"stt", "vad"}, "stt,vad"},
		{"exactly the cap", []string{"stt", "vad", "align"}, "stt,vad,align"},
		{"over the cap counts the rest", []string{"a", "b", "c", "d", "e"}, "a,b,c,+2"},
		// The server has already narrowed the list to what the card claims, so
		// a capability this build predates has to appear as itself rather than
		// be filtered out by a local whitelist.
		{"a capability this build has never heard of",
			[]string{"supersense"}, "supersense"},
		// And the server's order is kept: it is alphabetical there, and
		// re-ordering here would make two tables of the same model disagree.
		{"the server's order is kept", []string{"vision", "reasoning"}, "vision,reasoning"},
	}
	for _, c := range cases {
		if got := summarizeSupportNames(c.given); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestSortModelsIsAlphabeticalByName(t *testing.T) {
	items := []modelObject{
		{ID: "Olares/qwen3-4b"},
		{ID: "Anthropic/claude-opus-4"},
		{ID: "OpenAI/gpt-4o"},
	}
	sortModels(items)
	got := make([]string, 0, len(items))
	for _, m := range items {
		got = append(got, m.ID)
	}
	want := []string{"Anthropic/claude-opus-4", "Olares/qwen3-4b", "OpenAI/gpt-4o"}
	if !sort.StringsAreSorted(got) || !reflect.DeepEqual(got, want) {
		t.Errorf("sorted order: got %v want %v", got, want)
	}
}

// The table's job is to make the three fields Router adds visible. Mode is what
// separates one qualified name from the next when a client cannot pattern-match
// the id, and readiness is what explains a name that answers 503.
func TestRenderModelsListShowsModeSupportsAndReadiness(t *testing.T) {
	var buf bytes.Buffer
	items := []modelObject{
		{ID: "Olares/SenseVoiceSmall", Mode: "audio",
			Supports: []string{"stt", "vad"}, Readiness: "ready", OwnedBy: "Olares"},
		// An application running its own engine reports no phase, so nothing
		// can say whether its weights are loaded: `unknown`, and sendable.
		{ID: "Olares/embeddinggemma", Mode: "embedding", Readiness: "unknown", OwnedBy: "Olares"},
	}
	if err := renderModelsList(&buf, items, false); err != nil {
		t.Fatalf("render: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want a header and two rows:\n%s", len(lines), buf.String())
	}
	for _, column := range []string{"NAME", "MODE", "SUPPORTS", "READINESS", "SERVED BY"} {
		if !strings.Contains(lines[0], column) {
			t.Errorf("header is missing %q: %s", column, lines[0])
		}
	}
	if !strings.Contains(lines[1], "audio") || !strings.Contains(lines[1], "stt,vad") ||
		!strings.Contains(lines[1], "ready") {
		t.Errorf("first row lost a field: %s", lines[1])
	}
	// A model declaring no capability still needs a cell, and "-" is it.
	if !strings.Contains(lines[2], "embedding") || !strings.Contains(lines[2], "-") ||
		!strings.Contains(lines[2], "unknown") {
		t.Errorf("second row lost a field: %s", lines[2])
	}
}

// An empty list has two causes that lead somewhere different: a credential that
// may call nothing, and a model application whose weights are not loaded yet.
// The second is invisible without the flag, so the message names it — but not
// when the flag is already on, where it would be advice the reader has taken.
func TestEmptyModelsListNamesTheReadinessGateOnlyWhenItIsInPlay(t *testing.T) {
	var narrow bytes.Buffer
	if err := renderModelsList(&narrow, nil, false); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(narrow.String(), "--include-not-ready") {
		t.Errorf("the narrow read does not mention the flag that widens it: %s", narrow.String())
	}
	var wide bytes.Buffer
	if err := renderModelsList(&wide, nil, true); err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(wide.String(), "--include-not-ready") {
		t.Errorf("the wide read still suggests its own flag: %s", wide.String())
	}
}

// Off by default: this verb answers "what can I send", and a list padded with
// models that refuse every request would stop answering it.
func TestModelsPathAsksForTheWiderReadOnlyWhenRequested(t *testing.T) {
	if got := modelsPath(false); got != epDataPlaneModels {
		t.Errorf("default: got %q want %q", got, epDataPlaneModels)
	}
	want := epDataPlaneModels + "?include_not_ready=true"
	if got := modelsPath(true); got != want {
		t.Errorf("--include-not-ready: got %q want %q", got, want)
	}
}
