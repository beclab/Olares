package apiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/emicklei/go-restful/v3"
)

func startedQueue(t *testing.T) *OpController {
	t.Helper()
	op := NewQueue(context.Background())
	go op.worker()
	t.Cleanup(func() { op.wq.ShutDown() })
	return op
}

func newRestful() (*restful.Request, *restful.Response, *httptest.ResponseRecorder) {
	httpReq := httptest.NewRequest(http.MethodGet, "/apps/nginx/status", nil)
	req := restful.NewRequest(httpReq)
	rec := httptest.NewRecorder()
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)
	return req, resp, rec
}

// Regression for #3224: a panic in the wrapped handler must not hang the
// waiting HTTP goroutine, must surface a 500, and must not kill the worker.
func TestQueuedRecoversPanicAndReleasesCaller(t *testing.T) {
	op := startedQueue(t)
	h := &Handler{opController: op}

	req, resp, rec := newRestful()
	wrapped := h.queued(func(_ *restful.Request, _ *restful.Response) {
		panic("boom")
	})

	done := make(chan struct{})
	go func() { wrapped(req, resp); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("queued handler hung after panic (caller not released)")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status after panic=%d want 500", rec.Code)
	}

	// The single worker must survive the panic and process a later task.
	req2, resp2, _ := newRestful()
	ran := false
	wrapped2 := h.queued(func(_ *restful.Request, r *restful.Response) {
		ran = true
		r.WriteHeader(http.StatusOK)
	})
	done2 := make(chan struct{})
	go func() { wrapped2(req2, resp2); close(done2) }()
	select {
	case <-done2:
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not survive the panic")
	}
	if !ran {
		t.Error("subsequent task did not run after panic")
	}
}

func TestQueuedPassesThroughNormally(t *testing.T) {
	op := startedQueue(t)
	h := &Handler{opController: op}
	req, resp, rec := newRestful()
	h.queued(func(_ *restful.Request, r *restful.Response) {
		r.WriteHeader(http.StatusCreated)
	})(req, resp)
	if rec.Code != http.StatusCreated {
		t.Errorf("status=%d want 201", rec.Code)
	}
}

// The single worker must serialize all queued tasks: no two run concurrently.
func TestQueuedSerializesTasks(t *testing.T) {
	op := startedQueue(t)
	h := &Handler{opController: op}

	const n = 20
	var mu sync.Mutex
	concurrent, maxConcurrent := 0, 0

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, resp, _ := newRestful()
			h.queued(func(_ *restful.Request, _ *restful.Response) {
				mu.Lock()
				concurrent++
				if concurrent > maxConcurrent {
					maxConcurrent = concurrent
				}
				mu.Unlock()
				time.Sleep(time.Millisecond)
				mu.Lock()
				concurrent--
				mu.Unlock()
			})(req, resp)
		}()
	}
	wg.Wait()
	if maxConcurrent != 1 {
		t.Errorf("max concurrent tasks=%d want 1 (single serial worker)", maxConcurrent)
	}
}
