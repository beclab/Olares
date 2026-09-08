package router

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func musicCommand(t *testing.T, line ...string) (*cobra.Command, *mediaFlags) {
	t.Helper()
	return mediaCommand(t, musicFields, line...)
}

// A track is created on the music routes and read back from them. The record
// exists in the same table as every other generation, but it is not
// addressable on the unified route, so a mismatch here is a submission that
// succeeds and an id that cannot be collected.
func TestMusicUsesItsOwnRoutes(t *testing.T) {
	if musicKind.submitPath != epMusicGenerations {
		t.Errorf("music submits to %s", musicKind.submitPath)
	}
	if got := musicKind.get("gen_1"); got != epMusicGeneration("gen_1") {
		t.Errorf("music reads from %s", got)
	}
	if got := musicKind.content("gen_1"); got != epMusicGenerationContent("gen_1") {
		t.Errorf("music downloads from %s", got)
	}
}

// The music route refuses a field it does not know, so the canonical nesting
// every other family sends would be rejected whole rather than partly honored.
func TestATrackIsSpelledFlatTheWayItsRouteReadsIt(t *testing.T) {
	cmd, flags := musicCommand(t,
		"--duration", "30", "--format", "mp3", "--title", "Rain",
		"--lyrics", "la la", "--instrumental=false",
	)
	body, err := musicBody(cmd, "FlowStudio/ace-step", "a waltz", flags)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	const want = `{"model":"FlowStudio/ace-step","prompt":"a waltz","title":"Rain",` +
		`"lyrics":"la la","instrumental":false,"duration_seconds":30,"output_format":"mp3"}`
	if string(encoded) != want {
		t.Errorf("body:\n got %s\nwant %s", encoded, want)
	}
}

// Same reason as the canonical builder: a field nobody asked for is a field
// the resolved model may have no parameter for, and sending it is asking.
func TestAMusicBodyCarriesOnlyWhatWasAskedFor(t *testing.T) {
	cmd, flags := musicCommand(t)
	body, err := musicBody(cmd, "FlowStudio/ace-step", "a waltz", flags)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if string(encoded) != `{"model":"FlowStudio/ace-step","prompt":"a waltz"}` {
		t.Errorf("body carries more than was asked for: %s", encoded)
	}
}

// instrumental=false is a track with vocals asked for, not the absence of a
// preference, and it has to survive.
func TestAMusicBodyKeepsAZeroTheCallerAskedFor(t *testing.T) {
	cmd, flags := musicCommand(t, "--seed", "0", "--instrumental=false")
	body, err := musicBody(cmd, "FlowStudio/ace-step", "a waltz", flags)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if body.Seed == nil || *body.Seed != 0 {
		t.Errorf("seed: got %v, want 0", body.Seed)
	}
	if body.Instrumental == nil || *body.Instrumental {
		t.Errorf("instrumental: got %v, want false", body.Instrumental)
	}
}

// Repainting is the one operation music has besides generating from nothing,
// and it works on a recording. Each half without the other is a request that
// cannot mean anything.
func TestARepaintNeedsARecordingAndARecordingNeedsARepaint(t *testing.T) {
	cmd, flags := musicCommand(t, "--repaint", "--audio", "data:audio/mpeg;base64,AA==")
	body, err := musicBody(cmd, "FlowStudio/ace-step", "brighter chorus", flags)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if body.Operation != "repaint" || body.InputAudio == "" {
		t.Errorf("body: %+v", body)
	}

	cmd, flags = musicCommand(t, "--repaint")
	if _, err := musicBody(cmd, "FlowStudio/ace-step", "brighter", flags); err == nil ||
		!strings.Contains(err.Error(), "--audio") {
		t.Errorf("a repaint with nothing to repaint: %v", err)
	}

	cmd, flags = musicCommand(t, "--audio", "data:audio/mpeg;base64,AA==")
	if _, err := musicBody(cmd, "FlowStudio/ace-step", "a waltz", flags); err == nil ||
		!strings.Contains(err.Error(), "--repaint") {
		t.Errorf("a recording on a plain generate: %v", err)
	}
}

// A canceled generation is terminal. Treating it as unfinished would wait out
// the whole timeout for a state nothing is going to leave.
func TestACanceledTrackIsFinished(t *testing.T) {
	gen := &generationView{Status: "canceled"}
	if !gen.done() {
		t.Fatal("a canceled generation is not going to become anything else")
	}
	if gen.failed() {
		t.Fatal("cancelling is not failing; the difference is who asked")
	}
	if !gen.canceled() {
		t.Fatal("canceled should read as canceled")
	}
}

// Format and draft answer with a document rather than a file, and the text is
// the whole point, so it does not go in a table cell to be clipped.
func TestAFormatShowsBothVersionsAndWhatChanged(t *testing.T) {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(`{
	  "id": "mf_1", "object": "music.format", "status": "completed",
	  "model": "FlowStudio/ace-step", "vocal_language": "en",
	  "draft_prompt": "a waltz", "effective_prompt": "a slow waltz, 3/4, strings",
	  "draft_lyrics": "la la", "effective_lyrics": "[verse]\nla la la",
	  "warnings": ["lyrics were padded to fill 180 seconds"]
	}`), &raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	var buf bytes.Buffer
	if err := renderMusicText(&buf, "format", raw); err != nil {
		t.Fatalf("renderMusicText: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"DRAFT PROMPT", "EFFECTIVE PROMPT", "EFFECTIVE LYRICS",
		"[verse]", "lyrics were padded", "Nothing was generated",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("should show %q, got:\n%s", want, out)
		}
	}
}

// A field this build has never heard of still reaches the reader, because the
// task is held as it arrived rather than decoded into a struct.
func TestAMusicTaskReadsItsLifecycleWithoutDecodingTheRest(t *testing.T) {
	task := musicTask{}
	if err := json.Unmarshal([]byte(
		`{"id":"md_1","status":"failed","error":"the model refused the brief","tempo_bpm":128}`,
	), &task.raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !task.done() || !task.failed() {
		t.Fatalf("a failed task is finished: %+v", task.raw)
	}
	if got := task.reason(); got != "the model refused the brief" {
		t.Fatalf("reason: %q", got)
	}
	if _, ok := task.raw["tempo_bpm"]; !ok {
		t.Fatal("a field this build does not know should survive to the JSON output")
	}
}

// An interrupted submit is the case this exists for, so two of them must not
// collide on one key.
func TestEachSubmitMintsItsOwnIdempotencyKey(t *testing.T) {
	first, err := idempotencyKey()
	if err != nil {
		t.Fatalf("idempotencyKey: %v", err)
	}
	second, err := idempotencyKey()
	if err != nil {
		t.Fatalf("idempotencyKey: %v", err)
	}
	if first == second {
		t.Fatal("two submits sharing a key would be answered with one track")
	}
	// Router accepts one printable-ASCII value of 1 to 255 bytes.
	if len(first) == 0 || len(first) > 255 {
		t.Fatalf("key is %d bytes: %q", len(first), first)
	}
	for i := 0; i < len(first); i++ {
		if first[i] < 0x21 || first[i] > 0x7e {
			t.Fatalf("key carries a byte Router refuses: %q", first)
		}
	}
}
