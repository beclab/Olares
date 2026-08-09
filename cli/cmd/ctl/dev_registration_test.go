package ctl

import (
	"testing"
)

// hasSubcommand reports whether the command tree resolves the given
// path, e.g. []string{"dev", "push"}.
func hasSubcommand(t *testing.T, path ...string) bool {
	t.Helper()
	cmd, _, err := NewDefaultCommand().Find(path)
	if err != nil {
		return false
	}
	// Find returns the deepest command it could resolve, so a path that
	// stopped short resolves to an ancestor. Compare the leaf name.
	return cmd.Name() == path[len(path)-1]
}

// The `dev` group must be registered in BOTH distributions.
//
// It is deliberately outside root.go's `!remoteOnly` guard: that guard
// marks verbs needing an Olares host filesystem, and anything behind it
// only reaches users when the OS itself is upgraded (the host binary is
// cut by the OS release with publish-npm: false). Registering `dev`
// unconditionally is what lets it ship on the CLI's own npm cadence, so
// a regression that moves it behind the guard would silently delay
// every future fix by an OS release.
func TestDevGroupIsRegisteredInBothDistributions(t *testing.T) {
	verbs := [][]string{
		{"dev"},
		{"dev", "push"},
		{"dev", "deploy"},
		{"dev", "revert"},
		{"dev", "status"},
		{"dev", "node"},
		{"dev", "components"},
	}

	t.Run("host bundle", func(t *testing.T) {
		t.Setenv("OLARES_CLI_REMOTE_ONLY", "")
		for _, path := range verbs {
			if !hasSubcommand(t, path...) {
				t.Errorf("olares-cli %v is not registered on the host bundle", path)
			}
		}
	})

	t.Run("npm/npx build", func(t *testing.T) {
		t.Setenv("OLARES_CLI_REMOTE_ONLY", "1")
		for _, path := range verbs {
			if !hasSubcommand(t, path...) {
				t.Errorf("olares-cli %v is missing under OLARES_CLI_REMOTE_ONLY=1; "+
					"`dev` must not be inside the !remoteOnly guard", path)
			}
		}
	})
}

// Guard the other half of the contract: host verbs stay hidden in the
// npm build. If this ever passes, the guard has been removed and npx
// users get verbs that can only fail on a manifest they do not have.
func TestHostVerbsStayHiddenUnderRemoteOnly(t *testing.T) {
	t.Setenv("OLARES_CLI_REMOTE_ONLY", "1")
	for _, name := range []string{"install", "uninstall", "node", "gpu", "disk", "wizard", "osinfo"} {
		if hasSubcommand(t, name) {
			t.Errorf("host verb %q is exposed under OLARES_CLI_REMOTE_ONLY=1", name)
		}
	}
}

// `cluster workload set-image` is the primitive `dev deploy` calls; it
// lives in the always-registered cluster tree for the same reason.
func TestSetImageIsRegistered(t *testing.T) {
	t.Setenv("OLARES_CLI_REMOTE_ONLY", "1")
	if !hasSubcommand(t, "cluster", "workload", "set-image") {
		t.Error("olares-cli cluster workload set-image is not registered")
	}
}
