package clusterctx

import "testing"

func TestCanExecNamespace_Admin(t *testing.T) {
	const admin = "admin"
	sys := []string{"kube-system", "os-system"}

	cases := []struct {
		name      string
		namespace string
		want      bool
	}{
		{"own user-space", "user-space-admin", true},
		{"own user-system", "user-system-admin", true},
		{"sub-account user-space denied", "user-space-alice", false},
		{"sub-account user-system denied", "user-system-alice", false},
		{"system namespace allowed", "kube-system", true},
		{"unlisted system-looking namespace denied", "os-network", false},
		{"shared suffix allowed", "app-shared", true},
		{"os-protected allowed", "os-protected", true},
		{"empty namespace denied", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CanExecNamespace(tc.namespace, admin, GlobalRoleAdmin, sys)
			if got != tc.want {
				t.Fatalf("CanExecNamespace(%q, admin) = %v, want %v", tc.namespace, got, tc.want)
			}
		})
	}
}

func TestCanExecNamespace_NonAdmin(t *testing.T) {
	const user = "alice"
	sys := []string{"kube-system", "os-system"}

	cases := []struct {
		name      string
		namespace string
		want      bool
	}{
		{"own user-space", "user-space-alice", true},
		{"own user-system", "user-system-alice", true},
		{"other user's namespace denied", "user-space-admin", false},
		{"system namespace denied for non-admin", "kube-system", false},
		{"shared suffix denied for non-admin", "app-shared", false},
		{"os-protected denied for non-admin", "os-protected", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CanExecNamespace(tc.namespace, user, "platform-regular", sys)
			if got != tc.want {
				t.Fatalf("CanExecNamespace(%q, alice) = %v, want %v", tc.namespace, got, tc.want)
			}
		})
	}
}

func TestCanExecNamespace_EmptyUsernameDenied(t *testing.T) {
	if CanExecNamespace("user-space-alice", "", GlobalRoleAdmin, nil) {
		t.Fatal("empty username must never permit exec")
	}
}

func TestInfoCanExec(t *testing.T) {
	info := Info{
		Username:         "admin",
		GlobalRole:       GlobalRoleAdmin,
		SystemNamespaces: []string{"kube-system"},
	}
	if !info.CanExec("user-space-admin") {
		t.Fatal("admin should be able to exec into their own namespace")
	}
	if info.CanExec("user-space-alice") {
		t.Fatal("admin must not be able to exec into a sub-account's namespace")
	}
}
