package controllers

import (
	"errors"
	"fmt"
	"testing"
)

// TestIsNotFoundRecognisesTheWorkspaceLookupMisses pins the seam the secret
// read paths rely on to tell "this user has no workspace yet" apart from "the
// lookup failed". IsNotFound matches on message text, so GetWorkspace's two
// miss messages are restated here on purpose: rewording either one without
// updating this test would silently turn a fresh account back into a 500, and
// push consumers into matching the new prose themselves.
func TestIsNotFoundRecognisesTheWorkspaceLookupMisses(t *testing.T) {
	w := &workspaceClient{}

	notFound := []struct {
		name string
		err  error
	}{
		{
			name: "org has no workspaces at all, which is every account before its first secret write",
			err:  errors.New("not found the workspaces of user org"),
		},
		{
			name: "org has workspaces but not this one",
			err:  fmt.Errorf("not found the workspace: %s", "settings-alice"),
		},
	}
	for _, tc := range notFound {
		t.Run(tc.name, func(t *testing.T) {
			if !w.IsNotFound(tc.err) {
				t.Errorf("IsNotFound(%q) = false, want true", tc.err)
			}
		})
	}

	realFailures := []struct {
		name string
		err  error
	}{
		{"transport failure", errors.New("Get \"http://infisical/api/v2\": dial tcp: connection refused")},
		{"upstream rejected the token", errors.New("{\"statusCode\":401,\"message\":\"Unauthorized\"}")},
		{"decode failure", errors.New("invalid character '<' looking for beginning of value")},
	}
	for _, tc := range realFailures {
		t.Run(tc.name, func(t *testing.T) {
			if w.IsNotFound(tc.err) {
				t.Errorf("IsNotFound(%q) = true, want false: a lookup that failed must not read as an empty workspace", tc.err)
			}
		})
	}
}
