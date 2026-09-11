package router

import (
	"net/http"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/clierr"
)

// The interface is satisfied by a pointer receiver and the errors travel
// as *RouterError, so this is the assertion that would fail first if the
// receiver or the method set drifted.
var _ clierr.Structured = (*RouterError)(nil)

func TestRouterErrorReportsTheCodeTheMessageNames(t *testing.T) {
	err := &RouterError{Code: "quota_exceeded", Type: "invalid_request_error"}
	if got := err.ErrorCode(); got != "quota_exceeded" {
		t.Fatalf("ErrorCode() = %q", got)
	}
	// Router does not always send a code; the type is what is left, and
	// it is more useful to a caller than nothing.
	typeOnly := &RouterError{Type: "authentication_error"}
	if got := typeOnly.ErrorCode(); got != "authentication_error" {
		t.Fatalf("ErrorCode() = %q", got)
	}
}

func TestRetryableIsAnsweredOnlyWhereTheStatusSettlesIt(t *testing.T) {
	yes, no := true, false
	for _, tc := range []struct {
		name string
		err  *RouterError
		want *bool
	}{
		{"a Retry-After is Router saying to come back", &RouterError{Status: 500, RetryAfter: 5 * time.Second}, &yes},
		{"rate limited", &RouterError{Status: http.StatusTooManyRequests}, &yes},
		{"the model is still loading", &RouterError{Status: http.StatusServiceUnavailable}, &yes},
		{"a refusal earns itself again", &RouterError{Status: http.StatusForbidden}, &no},
		{"a bad request stays bad", &RouterError{Status: http.StatusBadRequest}, &no},
		// A stopped application and one that is still starting both
		// answer this way, and they need opposite responses.
		{"a bare 500 does not say", &RouterError{Status: http.StatusInternalServerError}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.err.Retryable()
			switch {
			case tc.want == nil && got != nil:
				t.Fatalf("claimed retryable=%v where the status says nothing", *got)
			case tc.want != nil && got == nil:
				t.Fatalf("no answer, want %v", *tc.want)
			case tc.want != nil && *got != *tc.want:
				t.Fatalf("retryable = %v, want %v", *got, *tc.want)
			}
		})
	}
}

// The human line joins the hint onto a sentence with "; ". A JSON field
// is not a sentence, so it carries the step alone.
func TestTheRecoveryActionIsNotPunctuatedForProse(t *testing.T) {
	err := &RouterError{Code: "forbidden_admin_required"}
	action := err.RecoveryAction()
	if action == "" {
		t.Fatal("a code with a known next step reported none")
	}
	if action[0] == ';' || action[0] == ' ' {
		t.Fatalf("the prose join survived into the field: %q", action)
	}
	if (&RouterError{Code: "something_router_added_later"}).RecoveryAction() != "" {
		t.Fatal("a code with no known step invented one")
	}
}
