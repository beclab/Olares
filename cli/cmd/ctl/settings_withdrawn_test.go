package ctl

import (
	"io"
	"strings"
	"testing"
)

// Withdrawn commands must be unreachable, including when invoked directly
// rather than discovered through help or shell completion.
func TestWithdrawnCommandsAreRejected(t *testing.T) {
	for _, path := range []string{
		"settings me check-update",
		"settings me sso list",
		"settings backup password set example",
		"settings integration cookie import --file example.txt",
		"settings integration cookie list",
		"settings integration cookie rm example.com",
		"settings integration cookie validate example.com",
		"settings advanced env system list",
		"settings advanced env user list",
		"settings integration accounts get awss3 example",
		"cluster middleware list -o json --show-passwords",
		"cluster middlewares list -o json --show-passwords",
		"cluster mw list -o json --show-passwords",
	} {
		t.Run(path, func(t *testing.T) {
			root := NewDefaultCommand()
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			root.SetArgs(strings.Fields(path))
			if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "unknown") {
				t.Fatalf("withdrawn command must fail as unknown before authentication: %v", err)
			}
		})
	}
}

func TestOtherSettingsCommandsRemainRegistered(t *testing.T) {
	root := NewDefaultCommand()
	for _, path := range []string{
		"settings me whoami", "settings me version", "settings me password",
		"settings backup plans list", "settings backup snapshots list",
		"settings integration accounts list", "settings integration accounts list-by-type",
		"settings integration accounts add awss3", "settings integration accounts delete",
		"settings advanced env system set", "settings advanced env user set",
		"cluster middleware list",
		"settings apps domain get", "settings apps domain list",
	} {
		t.Run(path, func(t *testing.T) {
			cmd, rest, err := root.Find(strings.Fields(path))
			if err != nil || len(rest) != 0 || cmd.CommandPath() != "olares-cli "+path {
				t.Fatalf("retained command does not resolve: command=%v rest=%v err=%v", cmd, rest, err)
			}
		})
	}
}
