package cmdutil

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

type statusErr struct {
	code int
	msg  string
}

func (e *statusErr) Error() string   { return e.msg }
func (e *statusErr) HTTPStatus() int { return e.code }

func TestIsVersionSuspect(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain error", errors.New("boom"), false},
		{"marked suspect", MarkVersionSuspect(errors.New("shape mismatch")), true},
		{"wrapped marked suspect", fmt.Errorf("op failed: %w", MarkVersionSuspect(errors.New("x"))), true},
		{"http 404", &statusErr{code: http.StatusNotFound, msg: "not found"}, true},
		{"http 501", &statusErr{code: http.StatusNotImplemented, msg: "not implemented"}, true},
		{"http 401 not suspect", &statusErr{code: http.StatusUnauthorized, msg: "unauth"}, false},
		{"http 500 not suspect", &statusErr{code: http.StatusInternalServerError, msg: "boom"}, false},
		{"wrapped http 404", fmt.Errorf("call: %w", &statusErr{code: http.StatusNotFound, msg: "nf"}), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isVersionSuspect(c.err); got != c.want {
				t.Fatalf("isVersionSuspect(%v)=%v, want %v", c.err, got, c.want)
			}
		})
	}
}

func TestMarkVersionSuspectNilPassthrough(t *testing.T) {
	if MarkVersionSuspect(nil) != nil {
		t.Fatal("MarkVersionSuspect(nil) must be nil")
	}
	wrapped := MarkVersionSuspect(errors.New("inner"))
	if wrapped.Error() != "inner" {
		t.Fatalf("Error()=%q, want %q", wrapped.Error(), "inner")
	}
	if !errors.Is(wrapped, wrapped) {
		t.Fatal("sanity")
	}
}
