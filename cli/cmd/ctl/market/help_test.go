package market

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func TestMarketCloneHelpMatchesCloneContract(t *testing.T) {
	long := NewCmdMarketClone(&cmdutil.Factory{}).Long
	for _, unwanted := range []string{"Clone an installed application", "cloneTarget", "'cloneable: true'"} {
		if strings.Contains(long, unwanted) {
			t.Errorf("clone help contains stale contract %q", unwanted)
		}
	}
	for _, required := range []string{"catalog", "allowMultipleInstall", "templateOnly", "targetApp"} {
		if !strings.Contains(long, required) {
			t.Errorf("clone help must describe %q", required)
		}
	}
}

func TestMarketUninstallHelpScopesDeleteData(t *testing.T) {
	long := NewCmdMarketUninstall(&cmdutil.Factory{}).Long
	// "persistent data" reads as "everything the app wrote", which sent at
	// least one user hunting for a bug when drive/Home survived uninstall.
	for _, unwanted := range []string{"the app's persistent data", "remove the app's persistent data"} {
		if strings.Contains(long, unwanted) {
			t.Errorf("uninstall help understates --delete-data's scope: %q", unwanted)
		}
	}
	// Addressing must match the platform storage model: Home and Data live
	// under drive/, the per-app cache is reached as cache/<node>.
	for _, required := range []string{"drive/Data", "cache/<node>", "drive/Home", "permission.userData", "deleteData"} {
		if !strings.Contains(long, required) {
			t.Errorf("uninstall help must describe %q", required)
		}
	}
}

// A caller that parses `-o json` reads the field list here and nowhere
// else, so a field added to OperationResult without a line in the help
// is invisible to every one of them.
//
// The anchored sentence is about the trap that survives knowing the
// field list: `running` is the settling state for three of these eight
// verbs and a failure report for the rest, so a check copied from the
// install example calls a successful stop a failure.
func TestEveryLifecycleVerbDocumentsItsJSONShape(t *testing.T) {
	f := &cmdutil.Factory{}
	verbs := map[string]*cobra.Command{
		"install":   NewCmdMarketInstall(f),
		"upgrade":   NewCmdMarketUpgrade(f),
		"uninstall": NewCmdMarketUninstall(f),
		"clone":     NewCmdMarketClone(f),
		"stop":      NewCmdMarketStop(f),
		"resume":    NewCmdMarketResume(f),
		"restart":   NewCmdMarketRestart(f),
		"cancel":    NewCmdMarketCancel(f),
	}
	for name, cmd := range verbs {
		for _, required := range []string{"-o json", "finalState", "status",
			`"running" is only it for install`} {
			if !strings.Contains(cmd.Long, required) {
				t.Errorf("%s help must describe %q", name, required)
			}
		}
		// The claim these replaced read the two fields the other way
		// round, and it shipped. Refuse it by name so a revert is loud.
		if strings.Contains(cmd.Long, ".finalState, not by .status") {
			t.Errorf("%s help is back to treating .finalState as the command's verdict", name)
		}
		if strings.Contains(cmd.Long, lifecycleJSONShape+"\n\nExamples:") {
			continue
		}
		if index := strings.Index(cmd.Long, "\nExamples:"); index >= 0 &&
			strings.Index(cmd.Long, lifecycleJSONShape) > index {
			t.Errorf("%s help buries the JSON shape below its examples", name)
		}
	}
}

func TestMarketGetHelpDescribesComputedCloneability(t *testing.T) {
	long := NewCmdMarketGet(&cmdutil.Factory{}).Long
	for _, unwanted := range []string{"full upstream payload", "jq '.cloneable'"} {
		if strings.Contains(long, unwanted) {
			t.Errorf("get help contains stale cloneability guidance %q", unwanted)
		}
	}
	for _, required := range []string{"Cloneable", "allowMultipleInstall", "templateOnly"} {
		if !strings.Contains(long, required) {
			t.Errorf("get help must describe %q", required)
		}
	}
}
