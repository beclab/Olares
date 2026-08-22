package router

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A task lives on one engine and the model reference is the only thing that
// routes back to it. A lookup that dropped it reaches whatever the category
// resolves today and is answered 404 for a task that is running fine.
func TestATaskLookupCarriesTheModel(t *testing.T) {
	got := audioTaskPath(epAudioTask("tsk_1"), "Olares/qwen3-asr")
	want := epAudioTask("tsk_1") + "?model=Olares%2Fqwen3-asr"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
	if got := audioTaskPath(epAudioTask("tsk_1"), "  "); got != epAudioTask("tsk_1") {
		t.Errorf("blank model: got %q, want the bare path", got)
	}
}

// The multipart routes read `async` as a form field and the JSON ones read it
// from the query. Sending it in the wrong place is not refused: the engine
// ignores it and the caller waits for the long answer they meant to escape.
func TestAsyncQueryOnlyAppearsWhenAsked(t *testing.T) {
	if got := asyncQuery(epAudioSpeech, false); got != epAudioSpeech {
		t.Errorf("not async: got %q want %q", got, epAudioSpeech)
	}
	if got := asyncQuery(epAudioSpeech, true); got != epAudioSpeech+"?async=1" {
		t.Errorf("async: got %q", got)
	}
}

// Only a 202 is a receipt. Every one of these verbs can also answer with the
// result itself, and reading that as a task would report an id the caller then
// cannot find.
func TestOnlyA202IsReadAsAReceipt(t *testing.T) {
	body := []byte(`{"task":{"id":"tsk_9","status":"queued"}}`)
	if task, ok := receiptFrom(202, body); !ok || task.ID != "tsk_9" {
		t.Errorf("202 with a task: ok=%v task=%+v", ok, task)
	}
	if _, ok := receiptFrom(200, body); ok {
		t.Error("a 200 was read as a receipt")
	}
	// A 202 whose body is the engine's own result rather than a task envelope.
	if _, ok := receiptFrom(202, []byte(`{"text":"hello"}`)); ok {
		t.Error("a 202 with no task id was read as a receipt")
	}
}

// A stream opens with a different scheme from every other verb in this tree,
// and getting it wrong arrives as a POST that the route does not serve.
func TestAStreamURLSwitchesScheme(t *testing.T) {
	cases := []struct {
		base string
		want string
	}{
		{"https://olares.example", "wss://olares.example" + epAudioStreamWS + "?model=m"},
		{"http://127.0.0.1:8080", "ws://127.0.0.1:8080" + epAudioStreamWS + "?model=m"},
		{"http://127.0.0.1:8080/", "ws://127.0.0.1:8080" + epAudioStreamWS + "?model=m"},
	}
	for _, c := range cases {
		got, err := audioStreamURL(c.base, audioStreamOptions{
			Route: epAudioStreamWS,
			Model: "m",
		})
		if err != nil {
			t.Fatalf("%s: %v", c.base, err)
		}
		if got != c.want {
			t.Errorf("%s: got %q want %q", c.base, got, c.want)
		}
	}
}

// A dialogue script names recordings by path, which means something on this
// machine and nothing to the engine. The conversion is what makes a script
// shareable, and skipping it sends a filename as if it were audio.
func TestADialogueScriptSendsItsRecordingsInline(t *testing.T) {
	dir := t.TempDir()
	clip := filepath.Join(dir, "alice.wav")
	if err := os.WriteFile(clip, []byte("RIFF....WAVEfmt "), 0o600); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "scene.json")
	raw := `{"speakers":[{"ref_audio":"` + clip + `","ref_text":"this is alice"},` +
		`{"ref_audio":"data:audio/wav;base64,AAAA","ref_text":"this is bob"}],` +
		`"turns":[{"speaker":0,"text":"hello"},{"speaker":1,"text":"hi"}]}`
	if err := os.WriteFile(script, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}

	body, err := buildDialogueBody(script, dialogueOptions{Model: categoryTTSDialogue})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), clip) {
		t.Error("the script's local path reached the body; the engine cannot read it")
	}
	speakers, ok := body["speakers"].([]map[string]any)
	if !ok || len(speakers) != 2 {
		t.Fatalf("speakers: %#v", body["speakers"])
	}
	if ref, _ := speakers[0]["ref_audio"].(string); !strings.HasPrefix(ref, "data:audio/") {
		t.Errorf("speakers[0].ref_audio was not encoded: %q", ref)
	}
	// An already-encoded clip is somebody's finished work; re-encoding it would
	// wrap a data URL in another one.
	if ref, _ := speakers[1]["ref_audio"].(string); ref != "data:audio/wav;base64,AAAA" {
		t.Errorf("speakers[1].ref_audio was rewritten: %q", ref)
	}
}

// The four ways a task lookup fails read the same from the status line, and
// three of them are about a task that exists: it is not ready, it is not
// reachable from here, or what it produced has been thrown away. Only one means
// the task is gone.
func TestATaskRefusalSaysWhichOfTheFourItIs(t *testing.T) {
	cases := []struct {
		name string
		err  *RouterError
		want string
	}{
		{"unplaceable", &RouterError{Status: 400, Code: "model_required"}, "--model"},
		{"unknown", &RouterError{Status: 404}, "forgotten task"},
		{"unfinished", &RouterError{Status: 409}, "no result yet"},
		{"dropped", &RouterError{Status: 410}, "Submit the work again"},
	}
	for _, c := range cases {
		got := audioTaskErr(c.err, "tsk_1f3c")
		if got == nil {
			t.Errorf("%s: no error", c.name)
			continue
		}
		if !strings.Contains(got.Error(), c.want) {
			t.Errorf("%s: %q does not explain itself with %q", c.name, got, c.want)
		}
		if !strings.Contains(got.Error(), "tsk_1f3c") {
			t.Errorf("%s: %q does not name the task", c.name, got)
		}
	}
	if audioTaskErr(nil, "tsk_1f3c") != nil {
		t.Error("a lookup that worked was reported as a failure")
	}
}

// An upload is encoded as it is sent. A recording long enough to be worth
// --async is one nobody should have to hold in memory, and a buffer bought
// nothing but a Content-Length neither Router nor the engines read.
func TestAnUploadIsStreamedRatherThanBuffered(t *testing.T) {
	dir := t.TempDir()
	clip := filepath.Join(dir, "meeting.wav")
	audio := strings.Repeat("A", 1<<16)
	if err := os.WriteFile(clip, []byte(audio), 0o600); err != nil {
		t.Fatal(err)
	}
	body, contentType, err := multipartFile(clip, "file", map[string]string{
		"model": categorySTT, "language": "", "async": "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatal(err)
	}
	// Nothing has read the body yet, and the file is already open on the other
	// end of the pipe. Reading it here is what a request does.
	mr := multipart.NewReader(body, params["boundary"])
	fields := map[string]string{}
	var file string
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		if part.FileName() != "" {
			file = string(content)
			continue
		}
		fields[part.FormName()] = string(content)
	}
	if file != audio {
		t.Errorf("the file arrived %d bytes long, want %d", len(file), len(audio))
	}
	if fields["model"] != categorySTT || fields["async"] != "1" {
		t.Errorf("fields: %#v", fields)
	}
	// A field nobody filled in is not a field. Sending it empty is how a
	// language the engine was meant to detect becomes one it was told.
	if _, ok := fields["language"]; ok {
		t.Error("an empty field was sent")
	}
	if _, _, err := multipartFile(filepath.Join(dir, "absent.wav"), "file", nil); err == nil {
		t.Error("a file that does not exist was accepted; the request would fail mid-send")
	}
}

// The two refusals a script gets before any request is made. Both are silent
// failures otherwise: a missing transcript degrades the cloned voice without
// saying so, and a turn pointing past the cast is a 4xx from the engine that
// names a field rather than the line.
func TestADialogueScriptIsCheckedBeforeItIsSent(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	noRefText := write("no-ref-text.json",
		`{"speakers":[{"ref_audio":"data:audio/wav;base64,AA"}],`+
			`"turns":[{"speaker":0,"text":"hello"}]}`)
	if _, err := buildDialogueBody(noRefText, dialogueOptions{}); err == nil {
		t.Error("a speaker with no ref_text was accepted")
	}
	badSpeaker := write("bad-speaker.json",
		`{"speakers":[{"ref_audio":"data:audio/wav;base64,AA","ref_text":"a"}],`+
			`"turns":[{"speaker":3,"text":"hello"}]}`)
	if _, err := buildDialogueBody(badSpeaker, dialogueOptions{}); err == nil {
		t.Error("a turn naming a speaker outside the cast was accepted")
	}
}
