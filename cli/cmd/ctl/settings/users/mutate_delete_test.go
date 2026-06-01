package users

import (
	"strings"
	"testing"
)

func TestUserIsOwner(t *testing.T) {
	tests := []struct {
		name string
		info *userInfo
		want bool
	}{
		{"owner role", &userInfo{Roles: []string{"owner"}}, true},
		{"owner with spaces", &userInfo{Roles: []string{" owner "}}, true},
		{"admin role", &userInfo{Roles: []string{"admin"}}, false},
		{"normal role", &userInfo{Roles: []string{"normal"}}, false},
		{"empty roles", &userInfo{Roles: []string{}}, false},
		{"nil info", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := userIsOwner(tt.info); got != tt.want {
				t.Fatalf("userIsOwner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateUserDeletable(t *testing.T) {
	if err := validateUserDeletable("missing", nil); err == nil {
		t.Fatal("expected error for nil user info")
	} else if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("err = %q, want username context", err.Error())
	}
	if err := validateUserDeletable("alice", &userInfo{Roles: []string{"owner"}}); err == nil {
		t.Fatal("expected error for owner")
	}
	if err := validateUserDeletable("bob", &userInfo{Roles: []string{"admin"}, State: "Deleting"}); err == nil {
		t.Fatal("expected error for Deleting state")
	}
	if err := validateUserDeletable("carol", &userInfo{Roles: []string{"normal"}, State: "Created"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
