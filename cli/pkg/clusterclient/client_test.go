package clusterclient

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

// TestHTTPErrorIsHTTPStatus pins the unwrap behavior the watch loops
// (cluster pod get -w, cluster pod logs -f, cluster workload
// rollout-status -w) rely on to distinguish "transient blip" from
// "terminal HTTP response". The wrap layer that watch verbs use is
// `fmt.Errorf("…: %w", err)`, so the helper MUST unwrap through that.
func TestHTTPErrorIsHTTPStatus(t *testing.T) {
	base := &HTTPError{Status: http.StatusNotFound, Message: "boom"}

	if !IsHTTPStatus(base, http.StatusNotFound) {
		t.Fatal("direct *HTTPError should match its status")
	}
	if IsHTTPStatus(base, http.StatusOK) {
		t.Fatal("status mismatch should not match")
	}
	if !IsNotFound(base) {
		t.Fatal("IsNotFound should recognize 404")
	}

	// Mirror the watch-loop wrap path verbatim.
	wrapped := fmt.Errorf("get pod ns/x: %w", base)
	if !IsHTTPStatus(wrapped, http.StatusNotFound) {
		t.Fatal("wrapped error should still match via errors.As")
	}
	if !IsNotFound(wrapped) {
		t.Fatal("IsNotFound should unwrap through fmt.Errorf %%w")
	}

	if IsHTTPStatus(errors.New("plain"), http.StatusNotFound) {
		t.Fatal("non-HTTPError should not match any status")
	}
	if IsHTTPStatus(nil, http.StatusNotFound) {
		t.Fatal("nil error should not match")
	}
}

// TestIsClientError checks the 4xx-is-terminal rule the watch loops
// use to short-circuit retries. 408 and 429 are deliberately excluded
// (both can resolve on a retry); other 4xx and 5xx behave as
// expected.
func TestIsClientError(t *testing.T) {
	cases := []struct {
		status int
		want   bool
	}{
		{http.StatusBadRequest, true},          // 400
		{http.StatusUnauthorized, true},        // 401
		{http.StatusForbidden, true},           // 403
		{http.StatusNotFound, true},            // 404
		{http.StatusRequestTimeout, false},     // 408 - retryable
		{http.StatusConflict, true},            // 409
		{http.StatusGone, true},                // 410
		{http.StatusTooManyRequests, false},    // 429 - retryable
		{http.StatusInternalServerError, false},// 500
		{http.StatusBadGateway, false},         // 502
		{http.StatusServiceUnavailable, false}, // 503
		{http.StatusGatewayTimeout, false},     // 504
		{200, false},
	}
	for _, c := range cases {
		he := &HTTPError{Status: c.status, Message: "x"}
		if got := IsClientError(he); got != c.want {
			t.Errorf("status %d: IsClientError=%v, want %v", c.status, got, c.want)
		}
		wrapped := fmt.Errorf("wrapped: %w", he)
		if got := IsClientError(wrapped); got != c.want {
			t.Errorf("wrapped status %d: IsClientError=%v, want %v", c.status, got, c.want)
		}
	}
	if IsClientError(errors.New("plain")) {
		t.Fatal("non-HTTPError must not be treated as a client error")
	}
	if IsClientError(nil) {
		t.Fatal("nil error must not be treated as a client error")
	}
}

// TestHTTPErrorErrorString covers both branches of the Error()
// renderer — Message wins when set, otherwise the generic dump.
func TestHTTPErrorErrorString(t *testing.T) {
	withMsg := (&HTTPError{Status: 404, Message: "hand-crafted"}).Error()
	if withMsg != "hand-crafted" {
		t.Fatalf("Message should win, got %q", withMsg)
	}
	bare := (&HTTPError{Status: 500, Method: "GET", URL: "http://x/y", Body: "boom"}).Error()
	want := "GET http://x/y: HTTP 500: boom"
	if bare != want {
		t.Fatalf("bare error string mismatch\nwant: %q\ngot:  %q", want, bare)
	}
	var nilErr *HTTPError
	if got := nilErr.Error(); got != "" {
		t.Fatalf("nil receiver should yield empty string, got %q", got)
	}
}
