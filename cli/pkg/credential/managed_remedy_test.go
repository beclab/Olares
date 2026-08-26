package credential

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// The remedy for a managed credential is never `profile login`: the command
// refuses it, and it would not help if it did, because nothing local can mint
// the grant. Every failure that reaches a user has to say so.
func TestManagedFailuresRedirectTheRemedy(t *testing.T) {
	at := time.Date(2026, 8, 26, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		err  error
	}{
		{"no token", &ErrNotLoggedIn{OlaresID: managedID, Managed: true, AppName: "lares"}},
		{"refused grant", &ErrTokenInvalidated{OlaresID: managedID, InvalidatedAt: at, Managed: true, AppName: "lares"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := tc.err.Error()
			if strings.Contains(msg, "profile login") || strings.Contains(msg, "profile import") {
				t.Errorf("message sends the user down a road that is closed to them: %s", msg)
			}
			if !strings.Contains(msg, `"lares"`) {
				t.Errorf("message does not name the application to repair: %s", msg)
			}
			if !strings.Contains(msg, managedID) {
				t.Errorf("message does not name the identity: %s", msg)
			}
		})
	}
}

// A grant we cannot attribute still gets the right remedy, minus the name.
func TestManagedFailureWithoutAnApplicationStillRedirects(t *testing.T) {
	msg := (&ErrNotLoggedIn{OlaresID: managedID, Managed: true}).Error()
	if strings.Contains(msg, "profile login") {
		t.Errorf("message sends the user down a road that is closed to them: %s", msg)
	}
	if !strings.Contains(msg, "reinstall or repair") {
		t.Errorf("message has no remedy at all: %s", msg)
	}
}

// Nothing changes for an account somebody logged into by hand.
func TestLocalFailuresKeepTheLoginCTA(t *testing.T) {
	for _, err := range []error{
		&ErrNotLoggedIn{OlaresID: managedID},
		&ErrTokenInvalidated{OlaresID: managedID, InvalidatedAt: time.Now()},
	} {
		if !strings.Contains(err.Error(), "profile login") {
			t.Errorf("message lost its call to action: %s", err)
		}
	}
}

func TestRequireNotManaged(t *testing.T) {
	if err := RequireNotManaged(nil, "remove"); err != nil {
		t.Errorf("a profile that does not exist is not managed: %v", err)
	}
	if err := RequireNotManaged(&cliconfig.ProfileConfig{OlaresID: managedID}, "remove"); err != nil {
		t.Errorf("a hand-made profile must pass: %v", err)
	}

	err := RequireNotManaged(&cliconfig.ProfileConfig{
		OlaresID: managedID, Managed: true, AppName: "lares",
	}, "remove")
	var managed *ErrManagedProfile
	if !errors.As(err, &managed) {
		t.Fatalf("err = %v, want *ErrManagedProfile", err)
	}
	// Naming the real revocation matters: the CLI cannot revoke anything,
	// and a message that only said "not allowed" would leave a user
	// believing the grant is stuck on the machine forever.
	if !strings.Contains(err.Error(), "uninstall") {
		t.Errorf("message does not say what actually revokes the grant: %s", err)
	}
}

// The application is filled in on failures the refresher raised, which knows
// the identity and the token but not which install they belong to.
func TestNameAppFillsInWhatTheRefresherCouldNotKnow(t *testing.T) {
	err := nameApp(&ErrNotLoggedIn{OlaresID: managedID, Managed: true}, "lares")
	var notLoggedIn *ErrNotLoggedIn
	if !errors.As(err, &notLoggedIn) || notLoggedIn.AppName != "lares" {
		t.Fatalf("err = %v, want the application named", err)
	}

	// An answer the caller already gave is not overwritten.
	err = nameApp(&ErrTokenInvalidated{OlaresID: managedID, Managed: true, AppName: "wise"}, "lares")
	var invalidated *ErrTokenInvalidated
	if !errors.As(err, &invalidated) || invalidated.AppName != "wise" {
		t.Fatalf("err = %v, want the original application kept", err)
	}
}
