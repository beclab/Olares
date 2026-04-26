package auth

import (
	"errors"
	"testing"
	"time"
)

// TestExpiresAt_RealOlaresToken pins the JWT decoder to a real Olares
// access_token captured from a pptest01 session. The payload is:
//
//	{"exp":1777127385,"iat":1777040985,"username":"pptest01",
//	 "groups":["lldap_admin"],"mfa":0,"jid":13962895094395427312}
//
// The test asserts ONLY the `exp` claim — it must NOT depend on any other
// field, because per §7.5 of the design doc the decoder is forbidden from
// surfacing them.
func TestExpiresAt_RealOlaresToken(t *testing.T) {
	const tok = "eyJhbGciOiJIUzUxMiJ9.eyJleHAiOjE3NzcxMjczODUsImlhdCI6MTc3NzA0MDk4NSwidXNlcm5hbWUiOiJwcHRlc3QwMSIsImdyb3VwcyI6WyJsbGRhcF9hZG1pbiJdLCJtZmEiOjAsImppZCI6MTM5NjI4OTUwOTQzOTU0MjczMTJ9.5uEvkvXlUrREuxqK1W2Vruke_OZdiuPdGysiC0XPXVJ9fz_X_-3wPyA4WXdsQKqT9P86yqeb5ZrRGFokCjGkmA"

	got, err := ExpiresAt(tok)
	if err != nil {
		t.Fatalf("ExpiresAt: unexpected error: %v", err)
	}
	want := time.Unix(1777127385, 0)
	if !got.Equal(want) {
		t.Errorf("ExpiresAt: got %s, want %s", got, want)
	}
}

func TestExpiresAt_Errors(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if _, err := ExpiresAt(""); err == nil {
			t.Fatal("expected error for empty token")
		}
	})
	t.Run("not-a-jwt", func(t *testing.T) {
		if _, err := ExpiresAt("not.a.jwt.too.many.dots"); err == nil {
			t.Fatal("expected error for malformed token")
		}
	})
	t.Run("no-exp-claim", func(t *testing.T) {
		// header.payload.sig where payload = base64url("{}") = "e30"
		_, err := ExpiresAt("h.e30.s")
		if !errors.Is(err, ErrNoExpClaim) {
			t.Fatalf("expected ErrNoExpClaim, got %v", err)
		}
	})
}

func TestIsExpired(t *testing.T) {
	const tok = "eyJhbGciOiJIUzUxMiJ9.eyJleHAiOjE3NzcxMjczODUsImlhdCI6MTc3NzA0MDk4NSwidXNlcm5hbWUiOiJwcHRlc3QwMSIsImdyb3VwcyI6WyJsbGRhcF9hZG1pbiJdLCJtZmEiOjAsImppZCI6MTM5NjI4OTUwOTQzOTU0MjczMTJ9.x"

	cases := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"before-exp", time.Unix(1777127385-3600, 0), false},
		{"at-exp", time.Unix(1777127385, 0), true},
		{"after-exp", time.Unix(1777127385+1, 0), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := IsExpired(tok, c.now, 0)
			if err != nil {
				t.Fatalf("IsExpired: %v", err)
			}
			if got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}
