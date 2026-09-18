// `router call ocr <file>` used to hang until its own timeout on a job the
// engine had already finished, printing nothing.
//
// A submission is answered with a receipt — the task nested under one key,
// alongside the same id one level up as `job_id` — and the code read that
// receipt as a bare task. Nothing failed: it got a task with an empty id and an
// empty status, then polled `/v1/ocr/tasks/` + "" , which is the list route, and
// read no top-level status off the board it got back. So the wait loop never saw
// a terminal state, and every poll asked the wrong URL for it.

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

// ocrStub answers the two routes a wait cycle needs and records what was asked.
// The task settles on the poll numbered settleAfter, so a test can distinguish
// "read the terminal state" from "gave up".
type ocrStub struct {
	mu         sync.Mutex
	paths      []string
	polls      int
	settleFrom int
}

func (s *ocrStub) serve(w http.ResponseWriter, r *http.Request) {
	// An id-less lookup addresses `/v1/ocr/tasks/`, and Router's trailing-slash
	// redirect lands it on the list route — the empty segment is not an id, so
	// there is no per-task route to reach. Modelling that is the whole point:
	// the list answer parses fine and says nothing about any one task.
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/v1/ocr/tasks"), "/")

	s.mu.Lock()
	s.paths = append(s.paths, r.URL.String())
	if r.Method == http.MethodGet && id != "" {
		s.polls++
	}
	polls := s.polls
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodPost:
		// The shape OCRAdapter actually answers a submission with.
		_, _ = w.Write([]byte(`{"job_id":"tsk_abc","status":"queued",` +
			`"task":{"object":"task","id":"tsk_abc","status":"queued","model":"paddle-hybrid"}}`))
	case id == "":
		_, _ = w.Write([]byte(`{"object":"list","capacity":{},"data":[]}`))
	default:
		status := "running"
		if polls >= s.settleFrom {
			status = "succeeded"
		}
		_, _ = w.Write([]byte(`{"object":"task","id":"tsk_abc","status":"` + status +
			`","model":"paddle-hybrid","result":{"text":"hello"}}`))
	}
}

func newOCRStub(t *testing.T, settleFrom int) (*routerClient, *ocrStub) {
	t.Helper()
	stub := &ocrStub{settleFrom: settleFrom}
	srv := httptest.NewServer(http.HandlerFunc(stub.serve))
	t.Cleanup(srv.Close)
	return newRouterClient(srv.Client(), srv.URL, "alice@example.com"), stub
}

func (s *ocrStub) asked() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.paths...)
}

// The receipt has to be read as a receipt. Decoded as a task it has no id, and
// the id is the whole of what the wait loop needs.
func TestOCRSubmissionReceiptYieldsTheTaskID(t *testing.T) {
	const receipt = `{"job_id":"tsk_abc","status":"queued",` +
		`"task":{"object":"task","id":"tsk_abc","status":"queued"}}`

	var flat ocrTask
	if err := json.Unmarshal([]byte(receipt), &flat); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if flat.ID != "" {
		t.Fatalf("the receipt is not a task after all; this test is describing the wrong bug")
	}

	var accepted ocrAccepted
	if err := json.Unmarshal([]byte(receipt), &accepted); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	task := accepted.task()
	if task.ID != "tsk_abc" {
		t.Errorf("id=%q want %q", task.ID, "tsk_abc")
	}
	if task.Status != "queued" {
		t.Errorf("status=%q want %q", task.Status, "queued")
	}
}

// An engine that answers with job_id and no nested task still named the job, so
// there is still something to come back for.
func TestOCRSubmissionFallsBackToJobID(t *testing.T) {
	var accepted ocrAccepted
	if err := json.Unmarshal([]byte(`{"job_id":"tsk_only","status":"queued"}`), &accepted); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	task := accepted.task()
	if task.ID != "tsk_only" {
		t.Errorf("id=%q want %q", task.ID, "tsk_only")
	}
	if task.Status != "queued" {
		t.Errorf("status=%q want %q", task.Status, "queued")
	}
}

// The loop has to reach a terminal state and stop. With an empty id every poll
// went to the list route, so this both pins the exit and pins the URL: an id in
// the path is the difference between reading a status and reading a board.
func TestOCRWaitStopsWhenTheTaskSettles(t *testing.T) {
	dp, stub := newOCRStub(t, 2)
	task := ocrTask{ID: "tsk_abc", Status: "queued"}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := waitForOCRTask(ctx, dp, &task, "Olares/paddle-hybrid", 20*time.Second, false); err != nil {
		t.Fatalf("wait: %v", err)
	}
	if task.Status != "succeeded" {
		t.Errorf("status=%q want %q", task.Status, "succeeded")
	}
	for _, p := range stub.asked() {
		if id := strings.Trim(strings.TrimPrefix(strings.SplitN(p, "?", 2)[0],
			"/v1/ocr/tasks"), "/"); id == "" {
			t.Errorf("a poll went to the list route, which carries no status: %q", p)
		}
	}
}

// A task belongs to one engine. Router routes an OCR task lookup by ?model=,
// so a poll that leaves it out asks whichever model answers the default OCR
// route — not necessarily the one holding the task.
func TestOCRPollCarriesTheModel(t *testing.T) {
	dp, stub := newOCRStub(t, 1)
	task := ocrTask{ID: "tsk_abc", Status: "queued"}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := waitForOCRTask(ctx, dp, &task, "Olares/paddle-hybrid", 20*time.Second, false); err != nil {
		t.Fatalf("wait: %v", err)
	}

	asked := stub.asked()
	if len(asked) == 0 {
		t.Fatal("no request reached the stub")
	}
	for _, p := range asked {
		if !strings.Contains(p, "model=Olares%2Fpaddle-hybrid") {
			t.Errorf("the poll dropped the model: %q", p)
		}
	}
}

// What the original bug actually cost, kept executable: an id-less task polls
// the list route and never settles, so the only thing that ends the wait is the
// timeout. This is the symptom that read as "the engine is slow" for six
// minutes on a job that finished in under two seconds, and it is why decoding
// the receipt correctly is not a cosmetic fix.
func TestOCRWaitWithNoTaskIDCanOnlyTimeOut(t *testing.T) {
	dp, stub := newOCRStub(t, 1)
	task := ocrTask{} // what reading the receipt as a task used to produce

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := waitForOCRTask(ctx, dp, &task, "Olares/paddle-hybrid", 2*time.Second, false)
	if err == nil {
		t.Fatal("an id-less wait settled, so it is reading a status from somewhere")
	}
	if !strings.Contains(err.Error(), "still") {
		t.Errorf("the wait ended some way other than the timeout: %v", err)
	}
	for _, p := range stub.asked() {
		if id := strings.Trim(strings.TrimPrefix(strings.SplitN(p, "?", 2)[0],
			"/v1/ocr/tasks"), "/"); id != "" {
			t.Errorf("an id-less poll somehow addressed task %q: %s", id, p)
		}
	}
}

// An empty model is left out rather than sent blank, the same way audioTaskPath
// does it: a bare lookup is a different request from one naming no model.
func TestOCRTaskPathOmitsAnEmptyModel(t *testing.T) {
	if got := ocrTaskPath("/v1/ocr/tasks/tsk_abc", "  "); got != "/v1/ocr/tasks/tsk_abc" {
		t.Errorf("path=%q want it unchanged", got)
	}
	if got := ocrTaskPath("/v1/ocr/tasks/tsk_abc", "Olares/paddle-hybrid"); !strings.Contains(got, "model=") {
		t.Errorf("path=%q carries no model", got)
	}
}
