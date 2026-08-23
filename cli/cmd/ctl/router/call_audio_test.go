package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type recordingHTTPClientFactory struct {
	shortCalls int
	longCalls  int
	short      *http.Client
	long       *http.Client
}

func (f *recordingHTTPClientFactory) HTTPClient(context.Context) (*http.Client, error) {
	f.shortCalls++
	return f.short, nil
}

func (f *recordingHTTPClientFactory) HTTPClientWithoutTimeout(context.Context) (*http.Client, error) {
	f.longCalls++
	return f.long, nil
}

func TestAuthenticatedClientSelectionKeepsShortRequestsTimed(t *testing.T) {
	f := &recordingHTTPClientFactory{
		short: &http.Client{Timeout: 30 * time.Second},
		long:  &http.Client{},
	}

	long, err := authenticatedHTTPClient(context.Background(), f, longRequest)
	if err != nil {
		t.Fatal(err)
	}
	if long.Timeout != 0 || f.longCalls != 1 || f.shortCalls != 0 {
		t.Fatalf("long client timeout=%s calls=(short:%d long:%d)",
			long.Timeout, f.shortCalls, f.longCalls)
	}

	short, err := authenticatedHTTPClient(context.Background(), f, shortRequest)
	if err != nil {
		t.Fatal(err)
	}
	if short.Timeout != 30*time.Second || f.shortCalls != 1 || f.longCalls != 1 {
		t.Fatalf("short client timeout=%s calls=(short:%d long:%d)",
			short.Timeout, f.shortCalls, f.longCalls)
	}
}

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

func TestAsyncQueryOnlyAppearsWhenAsked(t *testing.T) {
	if got := audioRequestPath(epAudioSpeech, "", false); got != epAudioSpeech {
		t.Errorf("not async: got %q want %q", got, epAudioSpeech)
	}
	if got := audioRequestPath(epAudioSpeech, "", true); got != epAudioSpeech+"?async=1" {
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

func TestMultipartAudioPathCarriesModel(t *testing.T) {
	for _, route := range []string{
		epAudioTranscriptions,
		epAudioTranslations,
		epAudioVAD,
		epAudioDiarization,
		epAudioAlign,
		epAudioEnhance,
		epAudioSpeechClone,
		epAudioSpeakerEmbeddings,
	} {
		got := audioRequestPath(route, "default-stt", false)
		if got != route+"?model=default-stt" {
			t.Errorf("%s: got %q", route, got)
		}
	}
	got := audioRequestPath(epAudioTranscriptions, "Olares/qwen3-asr", false)
	if want := epAudioTranscriptions + "?model=Olares%2Fqwen3-asr"; got != want {
		t.Errorf("slash model: got %q want %q", got, want)
	}
}

func TestEveryMultipartAudioVerbKeepsAsyncInTheForm(t *testing.T) {
	for _, verb := range []string{
		"transcribe", "translation", "vad", "diarize", "align", "enhance", "clone", "speaker-embed",
	} {
		t.Run(verb, func(t *testing.T) {
			fields := audioMultipartFields("default-audio", true, nil)
			if fields["model"] != "default-audio" || fields[audioAsyncFormField] != "1" {
				t.Fatalf("multipart fields: %#v", fields)
			}
		})
	}
}

func TestAudioUploadSizeBoundaries(t *testing.T) {
	dir := t.TempDir()
	makeSparse := func(name string, size int64) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Truncate(path, size); err != nil {
			t.Fatal(err)
		}
		return path
	}

	var stderr bytes.Buffer
	if err := checkAudioUploadSize(makeSparse("90m", audioUploadWarnBytes), &stderr); err != nil {
		t.Fatalf("90 MiB: %v", err)
	}
	if stderr.Len() != 0 {
		t.Errorf("90 MiB warned: %q", stderr.String())
	}
	if err := checkAudioUploadSize(makeSparse("over-90m", audioUploadWarnBytes+1), &stderr); err != nil {
		t.Fatalf("over 90 MiB: %v", err)
	}
	if !strings.Contains(stderr.String(), "16 kHz mono FLAC") ||
		!strings.Contains(stderr.String(), "split") ||
		!strings.Contains(stderr.String(), humanBytes(audioUploadWarnBytes+1)) ||
		!strings.Contains(stderr.String(), strconv.FormatInt(audioUploadWarnBytes+1, 10)) {
		t.Errorf("warning does not give both remedies: %q", stderr.String())
	}
	if audioFileMaxBytes > audioRequestBudgetBytes-64*1024 {
		t.Fatalf("file allowance %d leaves less than 64 KiB multipart overhead", audioFileMaxBytes)
	}
	if err := checkAudioUploadSize(makeSparse("max", audioFileMaxBytes), io.Discard); err != nil {
		t.Fatalf("file max: %v", err)
	}
	err := checkAudioUploadSize(makeSparse("over-max", audioFileMaxBytes+1), io.Discard)
	if err == nil {
		t.Fatal("file beyond the multipart allowance was accepted")
	}
	if !strings.Contains(err.Error(), strconv.FormatInt(audioFileMaxBytes, 10)) ||
		!strings.Contains(err.Error(), humanBytes(audioFileMaxBytes)) ||
		!strings.Contains(err.Error(), "total request body") {
		t.Errorf("limit is not exact or does not explain Router's request budget: %v", err)
	}
}

func TestOversizeAudioIsRejectedBeforeFactoryOrRequest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "too-large.wav")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(path, audioFileMaxBytes+1); err != nil {
		t.Fatal(err)
	}
	err := runCallTranscribe(context.Background(), nil, path, transcribeOptions{Model: categorySTT})
	if err == nil || !strings.Contains(err.Error(), "total request body") {
		t.Fatalf("got %v; the nil Factory must remain untouched", err)
	}
}

func TestAudioPollBackoffIsBoundedAndResetsOnProgress(t *testing.T) {
	delay := time.Second
	wants := []time.Duration{2 * time.Second, 3 * time.Second, 4 * time.Second, 5 * time.Second, 5 * time.Second}
	for i, want := range wants {
		delay = nextAudioPollDelay(delay, false)
		if delay != want {
			t.Fatalf("step %d: got %s want %s", i, delay, want)
		}
	}
	if got := nextAudioPollDelay(delay, true); got != time.Second {
		t.Errorf("progress reset: got %s want 1s", got)
	}
}

func TestReceiptKeepsRoutingModelInJSON(t *testing.T) {
	task := &audioTask{ID: "tsk_7", Model: "engine-weights", Status: "queued"}
	var out bytes.Buffer
	if err := printReceipt(&out, task, categorySTT, FormatJSON); err != nil {
		t.Fatal(err)
	}
	var got audioTask
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != "tsk_7" || got.Model != categorySTT {
		t.Errorf("receipt: %+v", got)
	}
	if task.Model != "engine-weights" {
		t.Error("printing mutated the server response")
	}
}

func TestTaskResultHelpWarnsAgainstBlindRetry(t *testing.T) {
	help := newCallTaskResultCommand(nil).Long
	for _, want := range []string{"check the task status before retrying", "bill the same result again"} {
		if !strings.Contains(help, want) {
			t.Errorf("task result help is missing %q", want)
		}
	}
}

func TestSpeakerEmbedCommandAndResponse(t *testing.T) {
	var found bool
	for _, cmd := range callVerbs(t) {
		if cmd.Name() == "speaker-embed" {
			found = true
			if cmd.Flag("model") == nil || cmd.Flag("async") == nil || cmd.Flag("output") == nil {
				t.Error("speaker-embed does not match the analysis command flags")
			}
		}
	}
	if !found {
		t.Fatal("speaker-embed command is missing")
	}

	vector := make([]float64, 512)
	vector[0], vector[1], vector[2], vector[511] = 0.25, -0.5, 0.75, 9
	raw, err := json.Marshal(speakerEmbeddingResponse{
		Model: "voice",
		Data:  []speakerEmbeddingData{{Embedding: vector}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := renderSpeakerEmbedding(&out, raw); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "voice") || !strings.Contains(out.String(), "512") ||
		!strings.Contains(out.String(), "9.048") || !strings.Contains(out.String(), "0.2500") {
		t.Errorf("summary is incomplete: %q", out.String())
	}
	if strings.Count(out.String(), "\n") > 5 || strings.Contains(out.String(), "9.0000") {
		t.Errorf("human output expanded the full vector: %q", out.String())
	}
	var jsonOut bytes.Buffer
	if err := renderAudioAnalysisAnswer(&jsonOut, FormatJSON, raw, renderSpeakerEmbedding); err != nil {
		t.Fatal(err)
	}
	var full speakerEmbeddingResponse
	if err := json.Unmarshal(jsonOut.Bytes(), &full); err != nil {
		t.Fatal(err)
	}
	if full.Model != "voice" || len(full.Data) != 1 || len(full.Data[0].Embedding) != 512 {
		t.Errorf("JSON output did not preserve the upstream fields and vector: %+v", full)
	}
}

func TestDialogueUploadUsesModelQueryAndCompleteBodyBudget(t *testing.T) {
	got := audioRequestPath(epAudioSpeech, "Olares/dialogue", true)
	want := epAudioSpeech + "?async=1&model=Olares%2Fdialogue"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}

	req := map[string]any{
		"model": categoryTTSDialogue,
		"speakers": []map[string]any{
			{"ref_audio": "data:audio/wav;base64,AAAA", "ref_text": "alice"},
			{"ref_audio": "data:audio/wav;base64,BBBB", "ref_text": "bob"},
		},
		"turns": []map[string]any{{"speaker": 0, "text": "hello"}, {"speaker": 1, "text": "hi"}},
	}
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkDialogueRequestSize(int64(len(body)), io.Discard); err != nil {
		t.Fatalf("small multi-speaker body: %v", err)
	}
	if err := checkDialogueRequestSize(audioRequestBudgetBytes, io.Discard); err != nil {
		t.Fatalf("exact request budget: %v", err)
	}
	err = checkDialogueRequestSize(audioRequestBudgetBytes+1, io.Discard)
	if err == nil || !strings.Contains(err.Error(), strconv.FormatInt(audioRequestBudgetBytes, 10)) ||
		!strings.Contains(err.Error(), humanBytes(audioRequestBudgetBytes)) {
		t.Fatalf("over-budget body: %v", err)
	}
	var warning bytes.Buffer
	if err := checkDialogueRequestSize(audioUploadWarnBytes+1, &warning); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(warning.String(), "reference clips") ||
		!strings.Contains(warning.String(), "--async") ||
		!strings.Contains(warning.String(), humanBytes(audioUploadWarnBytes+1)) ||
		!strings.Contains(warning.String(), strconv.FormatInt(audioUploadWarnBytes+1, 10)) ||
		strings.Contains(warning.String(), "STT") {
		t.Errorf("warning: %q", warning.String())
	}
}

func TestDialogueOversizeStopsBeforeFactory(t *testing.T) {
	err := submitDialogueBody(context.Background(), nil, dialogueOptions{Model: categoryTTSDialogue},
		strings.NewReader("{}"), audioRequestBudgetBytes+1)
	if err == nil || !strings.Contains(err.Error(), "total request body") {
		t.Fatalf("got %v; the nil Factory must remain untouched", err)
	}
}

func TestSpeechJSONRoutesShareModelAndAsyncQuery(t *testing.T) {
	want := epAudioSpeech + "?async=1&model=Olares%2Fvoice"
	if got := audioRequestPath(epAudioSpeech, "Olares/voice", true); got != want {
		t.Errorf("speech path: got %q want %q", got, want)
	}
	req := buildSpeakRequest("hello", speakOptions{Model: "Olares/voice"})
	if got := req["model"]; got != "Olares/voice" {
		t.Errorf("body model = %#v", got)
	}
}

func TestWaitForAudioTaskLoopWithoutSleeping(t *testing.T) {
	t.Run("deadline truncates sleep", func(t *testing.T) {
		now := time.Unix(0, 0)
		var sleeps []time.Duration
		fetches := 0
		ops := audioTaskWaitOps{
			now: func() time.Time { return now },
			sleep: func(_ context.Context, d time.Duration) error {
				sleeps = append(sleeps, d)
				now = now.Add(d)
				return nil
			},
			fetch: func(_ context.Context, _, _ string, out *audioTask) error {
				fetches++
				*out = audioTask{ID: "tsk", Status: "running"}
				return nil
			},
		}
		err := waitForAudioTaskWith(context.Background(), &audioTask{ID: "tsk", Status: "running"},
			categorySTT, 2500*time.Millisecond, false, ops)
		if err == nil || len(sleeps) != 1 || sleeps[0] != time.Second || fetches != 2 {
			t.Fatalf("err=%v sleeps=%v fetches=%d", err, sleeps, fetches)
		}
	})

	t.Run("final fetch can observe completion", func(t *testing.T) {
		now := time.Unix(0, 0)
		var sleeps []time.Duration
		fetches := 0
		ops := audioTaskWaitOps{
			now: func() time.Time { return now },
			sleep: func(_ context.Context, d time.Duration) error {
				sleeps = append(sleeps, d)
				now = now.Add(d)
				return nil
			},
			fetch: func(_ context.Context, _, _ string, out *audioTask) error {
				fetches++
				*out = audioTask{ID: "tsk", Status: "running"}
				if fetches == 2 {
					out.Status = "succeeded"
				}
				return nil
			},
		}
		err := waitForAudioTaskWith(context.Background(), &audioTask{ID: "tsk", Status: "running"},
			categorySTT, 2500*time.Millisecond, false, ops)
		if err != nil || len(sleeps) != 1 || sleeps[0] != time.Second || fetches != 2 {
			t.Fatalf("err=%v sleeps=%v fetches=%d", err, sleeps, fetches)
		}
	})

	t.Run("late wake does not fetch after deadline", func(t *testing.T) {
		now := time.Unix(0, 0)
		fetches := 0
		ops := audioTaskWaitOps{
			now: func() time.Time { return now },
			sleep: func(_ context.Context, _ time.Duration) error {
				now = now.Add(11 * time.Second)
				return nil
			},
			fetch: func(_ context.Context, _, _ string, _ *audioTask) error {
				fetches++
				return nil
			},
		}
		err := waitForAudioTaskWith(context.Background(), &audioTask{ID: "tsk", Status: "running"},
			categorySTT, 10*time.Second, false, ops)
		if err == nil || fetches != 0 {
			t.Fatalf("err=%v fetches=%d", err, fetches)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ops := audioTaskWaitOps{
			now: time.Now,
			sleep: func(context.Context, time.Duration) error {
				cancel()
				return ctx.Err()
			},
			fetch: func(context.Context, string, string, *audioTask) error {
				t.Fatal("fetched after cancellation")
				return nil
			},
		}
		err := waitForAudioTaskWith(ctx, &audioTask{ID: "tsk", Status: "running"},
			categorySTT, time.Minute, false, ops)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("progress resets and silence reaches cap", func(t *testing.T) {
		now := time.Unix(0, 0)
		var sleeps []time.Duration
		fetches := 0
		ops := audioTaskWaitOps{
			now: func() time.Time { return now },
			sleep: func(_ context.Context, d time.Duration) error {
				sleeps = append(sleeps, d)
				now = now.Add(d)
				return nil
			},
			fetch: func(_ context.Context, _, _ string, out *audioTask) error {
				fetches++
				ratio := 0.5
				*out = audioTask{ID: "tsk", Status: "running"}
				if fetches == 3 {
					out.Progress = &audioProgress{Stage: "decode", Ratio: &ratio}
				}
				if fetches == 10 {
					out.Status = "succeeded"
				}
				return nil
			},
		}
		err := waitForAudioTaskWith(context.Background(), &audioTask{ID: "tsk", Status: "running"},
			categorySTT, time.Minute, false, ops)
		if err != nil {
			t.Fatal(err)
		}
		want := []time.Duration{
			time.Second, 2 * time.Second, 3 * time.Second, time.Second,
			time.Second, 2 * time.Second, 3 * time.Second, 4 * time.Second,
			5 * time.Second, 5 * time.Second,
		}
		if len(sleeps) != len(want) {
			t.Fatalf("sleeps=%v", sleeps)
		}
		for i := range want {
			if sleeps[i] != want[i] {
				t.Fatalf("sleep %d: got %s want %s (%v)", i, sleeps[i], want[i], sleeps)
			}
		}
	})
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
