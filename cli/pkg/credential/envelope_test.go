package credential

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/clierr"
)

// The three credential failures read alike in prose and want three
// different things done, which is the case a `.error.code` branch exists
// for. This test is the contract: each one arrives under its own code,
// says it is not worth retrying, and names a next step that is not a
// restatement of the message.
func TestEachCredentialFailureArrivesUnderItsOwnCode(t *testing.T) {
	invalidatedAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name       string
		err        error
		wantCode   string
		wantAction string
	}{
		{
			name:       "nothing configured",
			err:        ErrNoProfile,
			wantCode:   clierr.CodeAuthNoProfile,
			wantAction: "olares-cli profile login --olares-id <id>",
		},
		{
			name:       "a profile with no token",
			err:        &ErrNotLoggedIn{OlaresID: "alice@olares.com"},
			wantCode:   clierr.CodeAuthNotLoggedIn,
			wantAction: "olares-cli profile login --olares-id alice@olares.com",
		},
		{
			name:       "a grant the server rejected",
			err:        &ErrTokenInvalidated{OlaresID: "alice@olares.com", InvalidatedAt: invalidatedAt},
			wantCode:   clierr.CodeAuthTokenInvalidated,
			wantAction: "olares-cli profile login --olares-id alice@olares.com",
		},
		{
			name:       "an access token that aged out",
			err:        &ErrTokenExpired{OlaresID: "alice@olares.com", ExpiredAt: invalidatedAt},
			wantCode:   clierr.CodeAuthTokenExpired,
			wantAction: "olares-cli profile login --olares-id alice@olares.com",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := envelopeFor(t, tc.err)
			if body["code"] != tc.wantCode {
				t.Fatalf("code = %v, want %q", body["code"], tc.wantCode)
			}
			if body["action"] != tc.wantAction {
				t.Fatalf("action = %v, want %q", body["action"], tc.wantAction)
			}
			// Not "unknown": nothing about running the same command
			// again mints a credential, and a caller that retries a
			// login failure on a schedule is the outcome this
			// prevents.
			if body["retryable"] != false {
				t.Fatalf("retryable = %v, want false", body["retryable"])
			}
		})
	}
}

// A platform-issued credential is minted by the platform: `profile
// login` is refused for it, so pointing a caller at it would send them
// somewhere that cannot help.
func TestAManagedCredentialIsNotSentToProfileLogin(t *testing.T) {
	for _, err := range []error{
		&ErrNotLoggedIn{OlaresID: "alice@olares.com", Managed: true, AppName: "wise"},
		&ErrTokenInvalidated{OlaresID: "alice@olares.com", Managed: true, AppName: "wise"},
	} {
		body := envelopeFor(t, err)
		action, _ := body["action"].(string)
		if action == "" {
			t.Fatalf("%T left the caller with nothing to do", err)
		}
		if bytes.Contains([]byte(action), []byte("profile login")) {
			t.Fatalf("%T points at a command that is refused for managed credentials: %q", err, action)
		}
	}
}

// ErrNoProfile stopped being errors.New so it could carry a code. Both
// comparisons callers already use have to survive that.
func TestErrNoProfileStillCompares(t *testing.T) {
	if !errors.Is(fmt.Errorf("resolve profile: %w", ErrNoProfile), ErrNoProfile) {
		t.Fatal("errors.Is no longer matches through a wrap")
	}
	//nolint:errorlint // the point of the test is that == still holds
	if err := error(noProfileError{}); err != ErrNoProfile {
		t.Fatal("the singleton is no longer comparable to itself")
	}
}

func envelopeFor(t *testing.T, err error) map[string]any {
	t.Helper()
	var out bytes.Buffer
	if !clierr.WriteEnvelope(&out, err) {
		t.Fatalf("no envelope was written for %T", err)
	}
	var document map[string]map[string]any
	if unmarshalErr := json.Unmarshal(out.Bytes(), &document); unmarshalErr != nil {
		t.Fatalf("the envelope is not JSON: %v\n%s", unmarshalErr, out.String())
	}
	return document["error"]
}
