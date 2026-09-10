package router

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The point of the type is that a caller can branch on it without reading the
// message, and still recognise the cause underneath.
func TestAnUnansweredWriteIsTypedAndKeepsItsCause(t *testing.T) {
	// A server that accepts the connection and answers nothing, so the
	// failure happens after the request was sent — the case where Router may
	// well have applied it.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := newRouterClient(srv.Client(), srv.URL, "someone@olares")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.doJSON(ctx, "POST", epQuotas, map[string]string{"scope": "user"}, nil)
	var uncertain *WriteUncertainError
	if !errors.As(err, &uncertain) {
		t.Fatalf("an unanswered write is not reported as one: %#v", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("the cause did not survive wrapping: %v", err)
	}
	if uncertain.Method != "POST" || uncertain.Path != epQuotas {
		t.Errorf("the error does not name what was attempted: %+v", uncertain)
	}
	msg := uncertain.Error()
	for _, want := range []string{"no idempotency key", "router quota list"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message is missing %q: %s", want, msg)
		}
	}
}

// A GET costs nothing to repeat, and a read that failed is just a read that
// failed. Dressing it up as a possible side effect is noise.
func TestAFailedReadIsJustAFailedRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := newRouterClient(srv.Client(), srv.URL, "someone@olares")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := c.doJSON(ctx, "GET", epQuotas, nil, nil)
	var uncertain *WriteUncertainError
	if errors.As(err, &uncertain) {
		t.Errorf("a read was reported as a possibly-applied write: %v", err)
	}
	if err == nil {
		t.Fatal("a cancelled read reported no error at all")
	}
}

// Nothing was ever sent, so nothing was applied. Warning about it here is what
// teaches a reader to skip the warning on the one occasion it is true.
func TestAConnectionThatWasNeverMadeIsNotUncertain(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"refused", &net.OpError{Op: "dial", Err: errors.New("connection refused")}},
		{"no such host", &net.DNSError{Err: "no such host", Name: "router.example"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if reachedRouter(tc.err) {
				t.Errorf("%v was treated as possibly applied", tc.err)
			}
		})
	}
	// Anything else is uncertain, including a read on an established
	// connection giving up.
	if !reachedRouter(&net.OpError{Op: "read", Err: errors.New("connection reset by peer")}) {
		t.Error("a reset mid-request was treated as never sent")
	}
}

// The verdict per route is the part worth having: whether to retry or to look.
func TestTheNoteSaysWhatASecondRequestWouldDo(t *testing.T) {
	for _, tc := range []struct {
		name         string
		method, path string
		want         string
	}{
		{
			// The one route where a blind retry silently costs something,
			// and the reason this table exists.
			name:   "a second key issue is a second key",
			method: "POST", path: epAPIKeys,
			want: "mints a second key",
		},
		{
			name:   "a duplicate name is refused, so retrying checks itself",
			method: "POST", path: epModelRoutes,
			want: "a 409 means the first attempt landed",
		},
		{
			name:   "attaching models skips what is already there",
			method: "POST", path: epProviderPredefinedModels("p1"),
			want: "retry is safe",
		},
		{
			name:   "a rollback appends a version every time",
			method: "POST", path: epProviderRollback("p1", 3),
			want: "appends another",
		},
		{
			name:   "a model call may already have been billed",
			method: "POST", path: epChatCompletions,
			want: "usage list",
		},
		{
			name:   "setting fields twice lands on the same state",
			method: "PATCH", path: epProvider("p1"),
			want: "same state",
		},
		{
			name:   "an unlisted route promises nothing",
			method: "POST", path: consoleAPI + "/something-new",
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := repeatNote(tc.method, tc.path)
			if tc.want == "" {
				if got != "" {
					t.Errorf("an unknown route was given a promise: %q", got)
				}
				return
			}
			if !strings.Contains(got, tc.want) {
				t.Errorf("repeatNote(%s %s) = %q, want it to mention %q", tc.method, tc.path, got, tc.want)
			}
		})
	}
}

// A route reached with a query string is the same route.
func TestTheNoteIgnoresTheQueryString(t *testing.T) {
	plain := repeatNote("POST", epQuotas)
	withQ := repeatNote("POST", epQuotas+"?user_id=abc")
	if plain == "" || plain != withQ {
		t.Errorf("a query string changed the verdict: %q vs %q", plain, withQ)
	}
}
