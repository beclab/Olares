package router

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// specPatch decides what is sent to an application's model card, and the two
// failures it can have are silent. Sending nothing writes the card back
// unchanged and reads as success; sending a key the caller did not ask to change
// replaces whatever was there. Both are worth a test, because neither shows up
// as an error anywhere.
func TestSpecPatchSendsOnlyWhatWasAsked(t *testing.T) {
	c := specEditFlags(t, "mode", "chat")
	patch, err := specPatch(c, "", "chat", "")
	if err != nil {
		t.Fatalf("mode only: %v", err)
	}
	if len(patch) != 1 || patch["mode"] != "chat" {
		t.Errorf("mode only: got %v", patch)
	}
}

// An empty --engine-args is a real instruction: it clears the flags, and the
// engine comes back bare. Treating it as "not given" would silently refuse the
// one edit that cannot be expressed any other way.
func TestClearingTheEngineFlagsIsExpressible(t *testing.T) {
	c := specEditFlags(t, "engine-args", "")
	patch, err := specPatch(c, "", "", "")
	if err != nil {
		t.Fatalf("empty engine-args: %v", err)
	}
	args, ok := patch["engine_args"]
	if !ok {
		t.Fatalf("empty engine-args was dropped: %v", patch)
	}
	if args != "" {
		t.Errorf("empty engine-args: got %q", args)
	}
}

// A patch with no keys would be a request that changes nothing while reporting
// that it worked. Router accepts it, the application rewrites its own card, and
// the caller is told the edit landed.
func TestAnEmptyPatchIsRefusedBeforeTheRequest(t *testing.T) {
	c := specEditFlags(t, "", "")
	_, err := specPatch(c, "", "", "")
	if err == nil {
		t.Fatal("a patch with nothing in it was accepted")
	}
	// The refusal has to be this verb's own. Falling through to the one the
	// application-side card reader raises would send a reader after
	// `router model spec show --app <app>`, which is a different road to the
	// document and takes a different kind of name.
	if !strings.Contains(err.Error(), "router model spec show <model>") {
		t.Errorf("the refusal does not say where to get the current card: %v", err)
	}
}

// A mode with no value is not a mode. Sent through, it declares a model that no
// category can point at and no dispatcher can route to, which is a worse
// outcome than the edit failing.
func TestABlankModeIsRefused(t *testing.T) {
	c := specEditFlags(t, "mode", "   ")
	if _, err := specPatch(c, "", "   ", ""); err == nil {
		t.Fatal("a blank mode was accepted")
	}
}

// The confirmation has to name the one edit with a cost. An engine relaunch
// takes a large model out of service for minutes, and a person answering "yes"
// to "change engine_args?" has not been told that.
func TestTheConfirmationNamesTheRelaunch(t *testing.T) {
	withArgs := specEditConsequence(map[string]any{"engine_args": "-c 4096"})
	if !strings.Contains(withArgs, "relaunches") {
		t.Errorf("engine_args edit does not mention the relaunch: %q", withArgs)
	}
	withoutArgs := specEditConsequence(map[string]any{"mode": "chat"})
	if strings.Contains(withoutArgs, "relaunches") {
		t.Errorf("a mode edit claims a relaunch: %q", withoutArgs)
	}
}

func TestSpecWriteResultKeepsPendingAppRestart(t *testing.T) {
	var result specWriteResult
	if err := json.Unmarshal([]byte(`{
		"restarted": false,
		"restart_supervision": "confirmed",
		"pending_app_restart": ["extensions.translate"]
	}`), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.RestartSupervision != "confirmed" {
		t.Errorf("restart supervision = %q", result.RestartSupervision)
	}
	if len(result.PendingAppRestart) != 1 ||
		result.PendingAppRestart[0] != "extensions.translate" {
		t.Errorf("pending app restart = %v", result.PendingAppRestart)
	}
}

func TestSpecWriteNamesFieldsWaitingForAnAppRestart(t *testing.T) {
	var out bytes.Buffer
	err := renderSpecWrite(&out, "translator", &specWriteResult{
		PendingAppRestart: []string{"extensions.translate"},
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, want := range []string{"extensions.translate", "market restart <app>"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output does not mention %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "in effect now") {
		t.Errorf("pending write claims immediate effect:\n%s", out.String())
	}
}

// specEditFlags builds the flag set specPatch reads "changed" from. It is
// declared here rather than taken from the command because the command's own
// values live in a closure; what is under test is the merge, and the flag set is
// only how it learns what the caller actually typed.
func specEditFlags(t *testing.T, changed, value string) *cobra.Command {
	t.Helper()
	c := &cobra.Command{Use: "edit"}
	var from, mode, engineArgs string
	c.Flags().StringVar(&from, "from", "", "")
	c.Flags().StringVar(&mode, "mode", "", "")
	c.Flags().StringVar(&engineArgs, "engine-args", "", "")
	if changed == "" {
		return c
	}
	if err := c.Flags().Set(changed, value); err != nil {
		t.Fatalf("set --%s: %v", changed, err)
	}
	return c
}
