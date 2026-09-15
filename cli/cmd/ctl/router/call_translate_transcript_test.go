package router

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A transcript arrives in two shapes and both have to be readable. A file
// written for this route is the request; a file a diarizer produced is the
// turns alone. Requiring the wrapper would mean asking a caller to edit a
// machine-produced file before it can be translated.
func TestATranscriptIsReadWhicheverShapeItArrivedIn(t *testing.T) {
	dir := t.TempDir()
	bare := filepath.Join(dir, "bare.json")
	if err := os.WriteFile(bare, []byte(`[{"id":"1","speaker":"A","text":"对"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	req, err := readTranscript(bare)
	if err != nil {
		t.Fatalf("a bare array of turns was refused: %v", err)
	}
	if len(req.Segments) != 1 || req.Segments[0].Speaker != "A" {
		t.Fatalf("the turns did not survive the read: %+v", req.Segments)
	}

	wrapped := filepath.Join(dir, "wrapped.json")
	body := `{"from":"zh","segments":[{"id":"1","text":"对"}],` +
		`"context":{"before":[{"speaker":"B","text":"ready?"}]},"background":"sprint review"}`
	if err := os.WriteFile(wrapped, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	req, err = readTranscript(wrapped)
	if err != nil {
		t.Fatalf("a request document was refused: %v", err)
	}
	if req.Background != "sprint review" {
		t.Errorf("the background was dropped: %q", req.Background)
	}
	if req.Context == nil || len(req.Context.Before) != 1 {
		t.Fatalf("the context was dropped: %+v", req.Context)
	}
	if req.Context.Before[0].Text != "ready?" {
		t.Errorf("the context line was not read: %+v", req.Context.Before[0])
	}
}

// The answer is keyed by id, so two turns carrying one id is a request whose
// answer cannot be matched. Upstream that is a 400 about the document; here it
// can name both turns.
func TestTwoTurnsCannotShareAnId(t *testing.T) {
	err := checkTranscript(&transcriptRequest{Segments: []transcriptSegment{
		{ID: "a", Text: "one"},
		{ID: "b", Text: "two"},
		{ID: "a", Text: "three"},
	}})
	if err == nil {
		t.Fatal("a duplicate id was accepted")
	}
	msg := err.Error()
	for _, want := range []string{"turns 1 and 3", `"a"`} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal does not say %q: %s", want, msg)
		}
	}
}

func TestATurnWithNoIdIsRefusedBeforeItIsSent(t *testing.T) {
	err := checkTranscript(&transcriptRequest{Segments: []transcriptSegment{
		{ID: "a", Text: "one"},
		{Text: "two"},
	}})
	if err == nil || !strings.Contains(err.Error(), "turn 2 has no id") {
		t.Fatalf("a turn with no id was not named: %v", err)
	}
}

// The ceiling is the upstream's. Refusing here says how to split; refusing
// there returns none of the page translated.
func TestATranscriptLongerThanOneRequestSaysHowToSplitIt(t *testing.T) {
	segments := make([]transcriptSegment, transcriptSegmentLimit+1)
	for i := range segments {
		segments[i] = transcriptSegment{ID: string(rune('a'+i%26)) + string(rune('0'+i/26)), Text: "x"}
	}
	err := checkTranscript(&transcriptRequest{Segments: segments})
	if err == nil {
		t.Fatal("a transcript over the ceiling was accepted")
	}
	if !strings.Contains(err.Error(), "context") {
		t.Errorf("the refusal does not say how to split: %s", err)
	}
	if err := checkTranscript(&transcriptRequest{}); err == nil {
		t.Error("an empty transcript was accepted")
	}
}

func TestAGlossaryTermIsSourceEqualsTarget(t *testing.T) {
	got, err := parseGlossary([]string{"沉浸式=immersive", " a = b "})
	if err != nil {
		t.Fatalf("a well-formed term was refused: %v", err)
	}
	if len(got) != 2 || got[0].Target != "immersive" || got[1].Source != "a" {
		t.Fatalf("the terms were not read: %+v", got)
	}
	for _, bad := range []string{"immersive", "=b", "a="} {
		if _, err := parseGlossary([]string{bad}); err == nil {
			t.Errorf("%q was accepted as a term", bad)
		}
	}
}

// A turn the model failed on is printed as its reason rather than skipped. A
// gap in a translated transcript is worse than a marked hole, because nothing
// in the output says a turn is missing.
func TestAFailedTurnIsPrintedInPlaceRatherThanDropped(t *testing.T) {
	var buf bytes.Buffer
	sent := []transcriptSegment{{ID: "1", Speaker: "Alice"}, {ID: "2", Speaker: "Bob"}}
	got := []transcriptResult{
		{ID: "1", Text: "right"},
		{ID: "2", Error: "context window"},
	}
	if err := renderTranscript(&buf, sent, got); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"Alice", "right", "Bob", "context window", "1 of 2 turns failed"} {
		if !strings.Contains(out, want) {
			t.Errorf("the rendering does not carry %q:\n%s", want, out)
		}
	}
}

// The transcript route is not one of the four MTran-compatible ones, and is
// spelled under /translate rather than beside it.
func TestTheTranscriptRouteIsUnderTranslate(t *testing.T) {
	if epTranslateTranscript != epTranslate+"/transcript" {
		t.Errorf("the transcript route is spelled %q", epTranslateTranscript)
	}
}
