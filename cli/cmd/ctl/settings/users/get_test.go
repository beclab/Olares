package users

import "testing"

func TestUserGetPathsEscapeUsernameConsistently(t *testing.T) {
	username := "alice/team one@olares.com"

	if got, want := userRecordPath(username), "/api/users/alice%2Fteam%20one@olares.com"; got != want {
		t.Fatalf("userRecordPath() = %q, want %q", got, want)
	}
	if got, want := userStatusPath(username), "/api/users/alice%2Fteam%20one@olares.com/status"; got != want {
		t.Fatalf("userStatusPath() = %q, want %q", got, want)
	}
}
