package capi

import (
	"encoding/json"
	"testing"

	"bytetrade.io/web3os/bfl/pkg/constants"
)

func TestGlobalRole(t *testing.T) {
	tests := map[string]string{
		constants.RoleOwner:  "platform-admin",
		constants.RoleAdmin:  "platform-admin",
		constants.RoleNormal: constants.RoleNormal,
		"":                   "",
	}
	for input, want := range tests {
		if got := globalRole(input); got != want {
			t.Fatalf("globalRole(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestUnwrapData(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "envelope", body: `{"code":0,"data":[{"name":"eth0"}]}`, want: `[{"name":"eth0"}]`},
		{name: "raw object", body: `{"name":"eth0"}`, want: `{"name":"eth0"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unwrapData([]byte(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			if !json.Valid(got) || string(got) != tt.want {
				t.Fatalf("unwrapData() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestClusterRolePattern(t *testing.T) {
	for input, want := range map[string]string{
		"clusterRole: host\n":                    "host",
		"multicluster:\n  clusterRole: member\n": "member",
	} {
		match := clusterRolePattern.FindStringSubmatch(input)
		if len(match) != 2 || match[1] != want {
			t.Fatalf("pattern(%q) = %#v, want %q", input, match, want)
		}
	}
}

func TestAppDetailResponseOmitsRemovedFields(t *testing.T) {
	payload, err := json.Marshal(appDetailResponse{})
	if err != nil {
		t.Fatal(err)
	}
	var response map[string]any
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"clusterRole", "localeManifest", "gitopsEngine"} {
		if _, exists := response[field]; exists {
			t.Errorf("app detail response still contains removed field %q", field)
		}
	}
}
