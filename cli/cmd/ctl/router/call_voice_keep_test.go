// `router call voice add` reported a voice that did not exist yet.
//
// A submission that the engine runs as a task is answered with a receipt — the
// task nested under one key — and the code read that receipt as a voice. Nothing
// failed: it got a voice with an empty id, printed "kept as " and a blank, and
// left. The clone was still being made, and the row Router had opened for it was
// waiting for a lookup that this command was never going to perform.

package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// voiceKeepStub answers the poll a wait cycle needs. The task settles from the
// poll numbered settleFrom, carrying finalStatus and finalBody, so a test can
// tell "read the terminal state" apart from "gave up".
type voiceKeepStub struct {
	mu          sync.Mutex
	paths       []string
	polls       int
	settleFrom  int
	finalStatus string
	finalTail   string
}

func (s *voiceKeepStub) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.paths = append(s.paths, r.URL.String())
	s.polls++
	polls := s.polls
	s.mu.Unlock()

	status, tail := "running", ""
	if polls >= s.settleFrom {
		status, tail = s.finalStatus, s.finalTail
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"object":"task","id":"atask_1","kind":"engine-task",` +
		`"cap":"tts_clone","model":"Olares/BreezeBlue/Breeze-TTS-2","status":"` +
		status + `"` + tail + `}`))
}

func newVoiceKeepStub(t *testing.T, settleFrom int, finalStatus, finalTail string) (*routerClient, *voiceKeepStub) {
	t.Helper()
	stub := &voiceKeepStub{settleFrom: settleFrom, finalStatus: finalStatus, finalTail: finalTail}
	srv := httptest.NewServer(http.HandlerFunc(stub.serve))
	t.Cleanup(srv.Close)
	return newRouterClient(srv.Client(), srv.URL, "alice@example.com"), stub
}

func (s *voiceKeepStub) asked() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.paths...)
}

// The receipt has to be read as a receipt. Decoded as a voice it is a voice with
// no id, and that is the whole of the bug: the command printed the blank and
// reported success.
func TestAVoiceReceiptDecodedAsAVoiceHasNoID(t *testing.T) {
	const receipt = `{"task":{"object":"task","id":"atask_1","kind":"engine-task",` +
		`"cap":"tts_clone","status":"queued","model":"Olares/BreezeBlue/Breeze-TTS-2"}}`

	var flat voice
	if err := json.Unmarshal([]byte(receipt), &flat); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if flat.ID != "" {
		t.Fatalf("the receipt is a voice after all; this test is describing the wrong bug")
	}

	var answer keptVoice
	if err := json.Unmarshal([]byte(receipt), &answer); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if answer.Task == nil || answer.Task.ID != "atask_1" {
		t.Fatalf("task=%+v want the id from the receipt", answer.Task)
	}
}

// The wait has to run to the end and read the voice out of the finished task.
// The id in the result is the only place the real voice_id ever appears.
func TestKeepingAVoiceWaitsForTheTaskAndReadsWhatItMade(t *testing.T) {
	dp, stub := newVoiceKeepStub(t, 2, "succeeded", `,"result_kind":"json","result":{"voice_id":"v-mine"}`)
	answer := keptVoice{Task: &audioTask{ID: "atask_1", Status: "queued"}}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	made, err := answer.keep(ctx, dp, "Olares/BreezeBlue/Breeze-TTS-2", 20*time.Second, false)
	if err != nil {
		t.Fatalf("keep: %v", err)
	}
	if made.ID != "v-mine" {
		t.Errorf("voice_id=%q want the id the task reported", made.ID)
	}
	// A task belongs to one engine, and Router routes the lookup by ?model=.
	asked := stub.asked()
	if len(asked) == 0 {
		t.Fatal("no request reached the stub")
	}
	for _, p := range asked {
		if !strings.Contains(p, "model=Olares%2FBreezeBlue%2FBreeze-TTS-2") {
			t.Errorf("the poll dropped the model: %q", p)
		}
		if !strings.Contains(p, "atask_1") {
			t.Errorf("the poll did not address the task: %q", p)
		}
	}
}

// An engine that stores the voice inline is left alone. The receipt is the
// exception, not the shape, and a verb that polled either way would turn one
// request into two against every engine that never needed a task.
func TestAnInlineVoiceIsKeptWithoutPollingAnything(t *testing.T) {
	dp, stub := newVoiceKeepStub(t, 1, "succeeded", "")
	var answer keptVoice
	if err := json.Unmarshal([]byte(`{"voice_id":"v-inline","name":"Night Host"}`), &answer); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	made, err := answer.keep(context.Background(), dp, "Olares/BreezeBlue/Breeze-TTS-2", time.Second, false)
	if err != nil {
		t.Fatalf("keep: %v", err)
	}
	if made.ID != "v-inline" {
		t.Errorf("voice_id=%q want the id the answer carried", made.ID)
	}
	if asked := stub.asked(); len(asked) != 0 {
		t.Errorf("an inline answer was polled anyway: %v", asked)
	}
}

// A task that ends any way but succeeded is a voice that was not kept, and
// saying so beats printing an id the library does not have.
func TestKeepingAVoiceReportsATaskThatFailed(t *testing.T) {
	dp, _ := newVoiceKeepStub(t, 1, "failed",
		`,"error":{"code":500,"message":"voice clone failed"}`)
	answer := keptVoice{Task: &audioTask{ID: "atask_1", Status: "queued"}}

	_, err := answer.keep(context.Background(), dp, "Olares/BreezeBlue/Breeze-TTS-2", 20*time.Second, false)
	if err == nil {
		t.Fatal("a failed task was reported as a kept voice")
	}
	if !strings.Contains(err.Error(), "voice clone failed") {
		t.Errorf("the engine's own reason is missing: %v", err)
	}
}

// A task that succeeded without naming the voice is the empty id again, one step
// later. The library is the place to look, so the message says so rather than
// printing a blank.
func TestASucceededTaskThatNamesNoVoiceIsReportedRatherThanPrintedBlank(t *testing.T) {
	dp, _ := newVoiceKeepStub(t, 1, "succeeded", "")
	answer := keptVoice{Task: &audioTask{ID: "atask_1", Status: "queued"}}

	made, err := answer.keep(context.Background(), dp, "Olares/BreezeBlue/Breeze-TTS-2", 20*time.Second, false)
	if err == nil {
		t.Fatalf("a task that named no voice yielded %+v", made)
	}
	if !strings.Contains(err.Error(), "voice list") {
		t.Errorf("the message does not say where to look: %v", err)
	}
}
