package os

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// captureStderr swaps os.Stderr for a pipe (never a TTY) and returns whatever
// fn wrote to it.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	w.Close()
	return <-done
}

func TestUploadBodyNonTTYEmitsNoCarriageReturns(t *testing.T) {
	payload := strings.Repeat("x", 512*1024)
	var body io.Reader

	out := captureStderr(t, func() {
		reader, finish := uploadBody(strings.NewReader(payload), int64(len(payload)))
		body = reader
		got, err := io.ReadAll(reader)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if len(got) != len(payload) {
			t.Errorf("body length = %d, want %d", len(got), len(payload))
		}
		finish()
	})

	if strings.Contains(out, "\r") {
		t.Errorf("non-TTY stderr contains carriage returns: %q", out)
	}
	if strings.Contains(out, "\033[") {
		t.Errorf("non-TTY stderr contains ANSI escapes: %q", out)
	}
	if out != "uploading log archive...\n" {
		t.Errorf("non-TTY stderr = %q, want single plain status line", out)
	}
	if _, wrapped := body.(*progressReader); wrapped {
		t.Error("non-TTY stderr should not wrap the body in a progressReader")
	}
}

func TestProgressReaderDrawClearsLine(t *testing.T) {
	p := &progressReader{r: strings.NewReader(""), total: 1000, read: 1000, label: "Uploading", last: -1}
	out := captureStderr(t, func() { p.draw(100) })

	if !strings.HasPrefix(out, "\r\033[K") {
		t.Errorf("draw output = %q, want it to start with \\r\\033[K to clear the line", out)
	}
}

func TestProgressReaderFinishNoOpWhenNothingDrawn(t *testing.T) {
	p := &progressReader{r: strings.NewReader(""), total: 0, label: "Uploading", last: -1}
	if out := captureStderr(t, p.finish); out != "" {
		t.Errorf("finish with nothing drawn wrote %q, want no output", out)
	}

	captureStderr(t, func() { p.draw(0) })
	out := captureStderr(t, p.finish)
	if out != "\n" {
		t.Errorf("finish after a draw wrote %q, want a single newline", out)
	}
}

func TestIsTransientNetErr(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{io.ErrUnexpectedEOF, true},
		{io.EOF, true},
		{fmt.Errorf("upload archive: Put \"https://s3\": net/http: TLS handshake timeout"), true},
		{fmt.Errorf("create ticket: Post \"https://ticket.olares.com/v1/olares-cli/tickets\": unexpected EOF"), true},
		{fmt.Errorf("read tcp 10.0.0.1:443: connection reset by peer"), true},
		{apiError(400, []byte(`{"message":"bad"}`)), false},
	}
	for _, tc := range cases {
		if got := isTransientNetErr(tc.err); got != tc.want {
			t.Errorf("isTransientNetErr(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}

func TestNewIdempotencyKeyFitsTicketAPI(t *testing.T) {
	a := newIdempotencyKey()
	b := newIdempotencyKey()
	if a == b {
		t.Fatal("expected distinct keys")
	}
	if len(a) < 8 || len(a) > 64 {
		t.Fatalf("key length %d outside ticket API 8–64 bound", len(a))
	}
}

func TestPostJSONRetriesTransientThenSucceeds(t *testing.T) {
	orig := sleep
	sleep = func(time.Duration) {}
	defer func() { sleep = orig }()

	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Errorf("hijack: %v", err)
				return
			}
			conn.Close()
			return
		}
		if got := r.Header.Get(headerIdempotencyKey); got != "retry-key-1" {
			t.Errorf("Idempotency-Key = %q, want retry-key-1", got)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ticket_id":"id-1","ticket_number":"TKT-1"}`))
	}))
	defer srv.Close()

	var out ticketResponse
	err := postJSON(srv.Client(), srv.URL, map[string]string{"x": "y"}, map[string]string{headerIdempotencyKey: "retry-key-1"}, &out)
	if err != nil {
		t.Fatalf("postJSON: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if out.TicketNumber != "TKT-1" {
		t.Fatalf("ticket_number = %q, want TKT-1", out.TicketNumber)
	}
}

func TestPostJSONDoesNotRetryClientErrors(t *testing.T) {
	orig := sleep
	sleep = func(time.Duration) {}
	defer func() { sleep = orig }()

	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"bad","message":"nope"}`))
	}))
	defer srv.Close()

	err := postJSON(srv.Client(), srv.URL, map[string]string{}, nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1 (no retry on 400)", attempts)
	}
}

func TestNewAPIClientDisablesKeepAlives(t *testing.T) {
	c := newAPIClient()
	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type %T", c.Transport)
	}
	if !tr.DisableKeepAlives {
		t.Fatal("API client must not reuse connections after a long S3 PUT")
	}
	if tr.TLSHandshakeTimeout != tlsHandshakeTimeout {
		t.Fatalf("TLSHandshakeTimeout = %s, want %s", tr.TLSHandshakeTimeout, tlsHandshakeTimeout)
	}
}

func TestKeepArchiveHintOnlyWhenCollected(t *testing.T) {
	err := fmt.Errorf("upload archive: unexpected EOF")
	if got := keepArchiveHint("/tmp/x.tar.gz", false, err); got != err {
		t.Fatalf("not collected: got %v, want original err", got)
	}
	out := captureStderr(t, func() {
		_ = keepArchiveHint("/tmp/x.tar.gz", true, err)
	})
	if !strings.Contains(out, "--file /tmp/x.tar.gz") {
		t.Fatalf("hint missing --file path: %q", out)
	}
}

// Send successful headers but truncate the JSON body. This exercises a failure
// after Client.Do succeeds, rather than the pre-header disconnect test above.
func TestCreateTicketRetriesTruncatedResponse(t *testing.T) {
	orig := sleep
	sleep = func(time.Duration) {}
	defer func() { sleep = orig }()
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if got := r.Header.Get(headerIdempotencyKey); got != "same-ticket-key" {
			t.Errorf("idempotency key = %q", got)
		}
		if attempts == 1 {
			w.Header().Set("Content-Length", "100")
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"ticket_id":`)
			return
		}
		_, _ = io.WriteString(w, `{"ticket_id":"id-1","ticket_number":"TKT-1"}`)
	}))
	defer srv.Close()
	result, err := createTicket(srv.Client(), srv.URL, &logUploadOptions{}, "attachment-1", "same-ticket-key")
	if err != nil || result.TicketNumber != "TKT-1" || attempts != 2 {
		t.Fatalf("result=%+v err=%v attempts=%d", result, err, attempts)
	}
}

func TestRunLogsUploadUncertainTicketRetainsArchiveWithoutRetryHint(t *testing.T) {
	orig := sleep
	sleep = func(time.Duration) {}
	defer func() { sleep = orig }()
	archive := filepath.Join(t.TempDir(), "logs.tar.gz")
	if err := os.WriteFile(archive, []byte("archive bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	var key string
	var ticketCalls, presignCalls, uploads int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case presignPath:
			presignCalls++
			fmt.Fprintf(w, `{"attachment_id":"attachment-1","upload_url":"http://%s/storage"}`, r.Host)
		case "/storage":
			uploads++
			_, _ = io.Copy(io.Discard, r.Body)
		case ticketPath:
			ticketCalls++
			got := r.Header.Get(headerIdempotencyKey)
			if got == "" || (key != "" && got != key) {
				t.Errorf("key changed or missing: %q -> %q", key, got)
			}
			key = got
			w.Header().Set("Content-Length", "100")
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"ticket_id":`)
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	var runErr error
	output := captureStderr(t, func() {
		runErr = runLogsUpload(&logUploadOptions{File: archive, Endpoint: srv.URL})
	})
	if !errors.Is(runErr, io.ErrUnexpectedEOF) {
		t.Fatalf("lost response read error: %v", runErr)
	}
	if !strings.Contains(runErr.Error(), "Check AssistHub before retrying") || strings.Contains(output, "retry with --file") {
		t.Fatalf("unsafe failure instructions: err=%v output=%q", runErr, output)
	}
	if presignCalls != 1 || uploads != 1 || ticketCalls != maxTransientAttempts {
		t.Fatalf("presign=%d uploads=%d tickets=%d", presignCalls, uploads, ticketCalls)
	}
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("archive not retained: %v", err)
	}
}

func TestCreateTicketRejectsMissingConfirmation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	}))
	defer srv.Close()
	if _, err := createTicket(srv.Client(), srv.URL, &logUploadOptions{}, "attachment-1", "key-1"); err == nil {
		t.Fatal("empty response must not be treated as confirmed ticket creation")
	}
}
