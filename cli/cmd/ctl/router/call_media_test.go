package router

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// One generation as Router answers today, on any of the three routes. Copied
// from a live answer, because this shape has already drifted underneath this
// package once: progress and outputs used to be inside the provider's own
// snapshot under `response`, and reading them from there went on compiling and
// went on returning nothing — no percentage while waiting, and no list of
// outputs to name with --output-id. A wire shape that fails silently has to be
// pinned by a decode rather than by a caller noticing.
const liveGeneration = `{
  "id": "gen_01HZX",
  "object": "media.generation",
  "media_type": "video",
  "operation": "generate",
  "status": "in_progress",
  "progress": 42,
  "model": "FlowStudio/wan2.2",
  "outputs": [
    {
      "id": "output-1",
      "content_type": "video/mp4",
      "width": 1280,
      "height": 720,
      "duration_seconds": 5,
      "content_url": "/v1/generations/gen_01HZX/content?outputId=output-1"
    },
    {
      "id": "output-2",
      "content_type": "video/mp4",
      "content_url": "/v1/generations/gen_01HZX/content?outputId=output-2"
    }
  ],
  "usage": {"count": 2, "seconds": 10},
  "provider_response": {"revised_prompt": "a paper plane"},
  "created_at": "2026-08-27T10:00:00Z",
  "expires_at": "2026-08-27T11:00:00Z"
}`

func TestGenerationDecodesTheFieldsRouterAnswersWith(t *testing.T) {
	var got generationView
	if err := json.Unmarshal([]byte(liveGeneration), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Object != "media.generation" {
		t.Errorf("object: got %q, want media.generation", got.Object)
	}
	if got.Progress == nil || *got.Progress != 42 {
		t.Errorf("progress: got %v, want 42", got.Progress)
	}
	if got.Usage == nil || got.Usage.Count != 2 || got.Usage.Seconds != 10 {
		t.Errorf("usage: got %+v, want {2 10}", got.Usage)
	}
	if len(got.Response) == 0 {
		t.Error("provider_response was dropped; the vendor's own words are the only account of what it said")
	}
	want := []generationOutput{
		{
			ID: "output-1", ContentType: "video/mp4", Width: 1280, Height: 720,
			DurationSeconds: 5,
			ContentURL:      "/v1/generations/gen_01HZX/content?outputId=output-1",
		},
		{
			ID: "output-2", ContentType: "video/mp4",
			ContentURL: "/v1/generations/gen_01HZX/content?outputId=output-2",
		},
	}
	if !reflect.DeepEqual(got.Outputs, want) {
		t.Errorf("outputs:\n got %+v\nwant %+v", got.Outputs, want)
	}
}

// The two readings a waiting caller depends on. Both are computed from the
// decoded view rather than re-parsed, so this is what would have failed when the
// wire moved.
func TestGenerationReportsProgressAndNamesItsOutputs(t *testing.T) {
	var gen generationView
	if err := json.Unmarshal([]byte(liveGeneration), &gen); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if note := gen.progressNote(); note != ", 42%" {
		t.Errorf("progress note: got %q, want %q", note, ", 42%")
	}
	if ids := gen.outputIDs(); !reflect.DeepEqual(ids, []string{"output-1", "output-2"}) {
		t.Errorf("output ids: got %v, want [output-1 output-2]", ids)
	}
}

// A provider that settles without naming its outputs is reported as one output
// with no id, and the content route serves it without one. There is nothing to
// choose between, so nothing to list: a hint pointing at --output-id here would
// name a flag with no argument to give it.
func TestAnUnnamedSingleOutputIsNotOfferedAsAChoice(t *testing.T) {
	const settled = `{
  "id": "gen_01HZY", "object": "media.generation", "media_type": "image",
  "operation": "generate", "status": "completed", "model": "OpenAI/gpt-image-1",
  "outputs": [{"id": "", "content_url": "/v1/generations/gen_01HZY/content"}],
  "created_at": "2026-08-27T10:00:00Z", "expires_at": "2026-08-27T11:00:00Z"
}`
	var gen generationView
	if err := json.Unmarshal([]byte(settled), &gen); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if ids := gen.outputIDs(); len(ids) != 0 {
		t.Errorf("output ids: got %v, want none", ids)
	}
	if !gen.done() || gen.failed() {
		t.Errorf("a completed generation reads as %q", gen.Status)
	}
	if note := gen.progressNote(); note != "" {
		t.Errorf("a provider that reported no progress produced %q", note)
	}
}

// Router's own status vocabulary, which is not the provider's. `queued` and
// `in_progress` are both still running, and treating an unknown literal as
// finished would write a file that is not there yet.
func TestOnlyTerminalStatusesEndTheWait(t *testing.T) {
	for _, status := range []string{"queued", "in_progress", "validating", ""} {
		gen := generationView{Status: status}
		if gen.done() {
			t.Errorf("%q ended the wait", status)
		}
	}
	for _, status := range []string{"completed", "failed"} {
		gen := generationView{Status: status}
		if !gen.done() {
			t.Errorf("%q did not end the wait", status)
		}
	}
	if !(&generationView{Status: "failed"}).failed() {
		t.Error("a failed generation did not report as failed")
	}
}

// A failure has to arrive with a reason a person can act on, and Router sends
// the reason in either of two fields.
func TestAFailedGenerationExplainsItself(t *testing.T) {
	message := "content policy"
	code := "provider_rejected"
	cases := []struct {
		name string
		gen  generationView
		want string
	}{
		{"a message", generationView{Error: &message}, message},
		{"only a code", generationView{ErrorCode: &code}, code},
		{"neither", generationView{}, "the provider did not say why"},
	}
	for _, c := range cases {
		if got := c.gen.reason(); !strings.Contains(got, c.want) {
			t.Errorf("%s: got %q, want it to carry %q", c.name, got, c.want)
		}
	}
}
