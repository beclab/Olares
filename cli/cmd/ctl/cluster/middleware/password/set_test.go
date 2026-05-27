package password

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/clusterclient"
)

// TestNormalizeType pins the client-side --type validation. The
// validator runs BEFORE the password prompt, so a typo must fail
// loudly instead of asking the operator to type a secret that's
// going to land in a 404 anyway.
func TestNormalizeType(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "lowercase canonical", in: "postgres", want: "postgres"},
		{name: "uppercase normalized", in: "REDIS", want: "redis"},
		{name: "leading/trailing space stripped", in: "  mysql  ", want: "mysql"},
		{name: "mixed case + space", in: "  MariaDB ", want: "mariadb"},
		{name: "every supported entry", in: "elasticsearch", want: "elasticsearch"},
		{name: "empty string rejected", in: "", wantErr: true},
		{name: "whitespace-only rejected", in: "   ", wantErr: true},
		{name: "unknown type rejected", in: "sqlite", wantErr: true},
		{name: "near-miss typo rejected", in: "postgresql", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeType(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("normalizeType(%q): want error, got %q", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeType(%q): unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("normalizeType(%q): got %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestFormatSetErrorNotImplemented pins the 501 friendly-error
// branch. The server only rotates postgres today
// (platform/tapr/cmd/middleware/app/handler.go), so every other
// --type ends in a *HTTPError{Status: 501}. We MUST translate that
// into a sentence that says "the server doesn't support this yet"
// rather than echoing the raw "HTTP 501: Not Implemented" — the
// operator otherwise has no way to tell a known server gap apart
// from a transient routing fault.
//
// The asserts deliberately string-match three pieces so the wording
// can be tweaked without breaking the test, but the signal stays:
//   - the offending type name appears (operator confirms we got it)
//   - "501" or "Not Implemented" appears (preserves the diagnostic)
//   - "postgres" appears (tells the operator what DOES work)
func TestFormatSetErrorNotImplemented(t *testing.T) {
	raw := &clusterclient.HTTPError{
		Status: http.StatusNotImplemented,
		Method: "POST",
		URL:    "https://control-hub.example/middleware/v1/redis/password",
		Body:   "Not Implemented",
	}
	got := formatSetError(raw, "redis", "os-platform", "redis-demo", "admin")
	if got == nil {
		t.Fatal("formatSetError on a 501 must not return nil")
	}
	msg := got.Error()
	if !strings.Contains(msg, "redis") {
		t.Errorf("error message should name the offending type; got %q", msg)
	}
	if !strings.Contains(msg, "501") && !strings.Contains(strings.ToLower(msg), "not implemented") {
		t.Errorf("error message should preserve the 501 signal; got %q", msg)
	}
	if !strings.Contains(msg, "postgres") {
		t.Errorf("error message should name the supported type(s); got %q", msg)
	}
	// The 501 branch deliberately does NOT wrap the original error:
	// the raw "POST <url>: HTTP 501: ..." dump is noise we just
	// replaced. errors.As on *HTTPError should therefore NOT find
	// the original HTTPError — the helpful sentence is what the
	// operator sees, end of story.
	var he *clusterclient.HTTPError
	if errors.As(got, &he) {
		t.Errorf("501 branch should swallow the *HTTPError, but errors.As recovered it: %+v", he)
	}
}

// TestFormatSetErrorGenericPreservesChain pins the non-501 branch:
// any other error MUST be wrapped with %w so downstream callers can
// still errors.As / errors.Is on the typed *HTTPError (clusterclient
// callers like watch loops rely on this). Concretely: a 404 should
// still be detectable as such after formatSetError mutates the
// outer wording.
func TestFormatSetErrorGenericPreservesChain(t *testing.T) {
	raw := &clusterclient.HTTPError{
		Status:  http.StatusNotFound,
		Method:  "POST",
		URL:     "https://control-hub.example/middleware/v1/postgres/password",
		Message: "POST .../password: HTTP 404: instance not found",
	}
	got := formatSetError(raw, "postgres", "prod", "pg-1", "admin")
	if got == nil {
		t.Fatal("formatSetError on a 404 must not return nil")
	}
	// The wrapper must mention what we were trying to do — the
	// existing "set <type> password for <ns>/<name> user=<u>"
	// envelope is the contract.
	want := "set postgres password for prod/pg-1 user=admin"
	if !strings.Contains(got.Error(), want) {
		t.Errorf("envelope wording missing: got %q, want substring %q", got.Error(), want)
	}
	if !clusterclient.IsNotFound(got) {
		t.Errorf("errors.As / IsNotFound must still recognize the wrapped 404; got %v", got)
	}
}

// TestFormatSetErrorNil keeps the trivial guard: a nil input must
// return nil. Without it the caller — runSet's `if err := ...; err
// != nil` — would still be safe, but the helper is also used in
// tests / future call sites where someone might forget the nil
// check.
func TestFormatSetErrorNil(t *testing.T) {
	if got := formatSetError(nil, "postgres", "prod", "pg-1", "admin"); got != nil {
		t.Fatalf("formatSetError(nil) should return nil; got %v", got)
	}
}

// TestFormatSetErrorNonHTTPError pins behavior for plain errors
// that aren't *HTTPError (e.g. network failures from
// refreshingTransport before the request lands). The 501 branch
// must NOT match — IsHTTPStatus is well-typed — and the generic
// wrap must fire.
func TestFormatSetErrorNonHTTPError(t *testing.T) {
	plain := fmt.Errorf("dial control-hub.example:443: i/o timeout")
	got := formatSetError(plain, "postgres", "prod", "pg-1", "admin")
	if got == nil {
		t.Fatal("formatSetError on a plain error must not return nil")
	}
	if !strings.Contains(got.Error(), "i/o timeout") {
		t.Errorf("plain error must be wrapped, not swallowed; got %q", got.Error())
	}
	if !errors.Is(got, plain) {
		t.Errorf("errors.Is should unwrap to the original network error; got %v", got)
	}
}
