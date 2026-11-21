package os

import (
	"testing"

	"github.com/beclab/Olares/cli/pkg/pipelines"
)

// An olaresd that is already released calls `upgrade bootstrap-node` on the
// olares-cli it just installed, and so does the Job that stands in for one too
// old to. Neither can be changed after release, so this command line is fixed.
func TestBootstrapNodeCommandLineIsFrozen(t *testing.T) {
	root := NewCmdUpgradeOs()
	cmd, rest, err := root.Find([]string{"bootstrap-node"})
	if err != nil || len(rest) != 0 || cmd.Name() != "bootstrap-node" {
		t.Fatalf("upgrade has no bootstrap-node subcommand: %v %v", err, rest)
	}
	for _, flag := range []string{"version", "base-dir", "cdn-service", "url-override"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("upgrade bootstrap-node has no --%s", flag)
		}
	}
	if err := cmd.Args(cmd, []string{"extra"}); err == nil {
		t.Error("upgrade bootstrap-node accepts positional arguments")
	}
	if pipelines.BootstrapNodeComponent != "olaresd" {
		t.Errorf("bootstrap-node fetches %q, want olaresd", pipelines.BootstrapNodeComponent)
	}
}
