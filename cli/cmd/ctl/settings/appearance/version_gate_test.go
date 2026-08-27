package appearance

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// The three verbs whose upstream landed in 1.12.6 share one gate; each
// passes its own verb name and reason so the message names the command
// the user actually typed.
func TestRequireAppearanceBackendVersion(t *testing.T) {
	previous := viper.GetString(cmdutil.FlagOlaresVersion)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, previous) })

	gated := []struct {
		verb   string
		reason string
	}{
		{"settings appearance theme set", "the OLARES_USER_THEME user env"},
		{"settings appearance widget set", "the widget preferences API"},
		{"settings appearance layout reset", "the desktop layout reset route"},
	}

	viper.Set(cmdutil.FlagOlaresVersion, "1.12.5")
	for _, g := range gated {
		err := requireAppearanceBackendVersion(context.Background(), cmdutil.NewFactory(), g.verb, g.reason)
		if err == nil {
			t.Fatalf("%s: expected a version gate error on 1.12.5", g.verb)
		}
		for _, want := range []string{g.verb, "requires Olares >= 1.12.6", g.reason, "1.12.5"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: error %q missing %q", g.verb, err, want)
			}
		}
	}

	// A dated 1.12.6 build is the 1.12.6 line and must pass.
	viper.Set(cmdutil.FlagOlaresVersion, "1.12.6-20260203")
	for _, g := range gated {
		if err := requireAppearanceBackendVersion(context.Background(), cmdutil.NewFactory(), g.verb, g.reason); err != nil {
			t.Fatalf("%s: daily 1.12.6 build should pass: %v", g.verb, err)
		}
	}

	// Fail-closed: an undetectable version is rejected with the shared
	// refresh hint rather than letting the call 404.
	viper.Set(cmdutil.FlagOlaresVersion, "not-a-version")
	err := requireAppearanceBackendVersion(context.Background(), cmdutil.NewFactory(),
		"settings appearance theme set", "the theme user env")
	if err == nil || !strings.Contains(err.Error(), "profile list --refresh-version") {
		t.Fatalf("expected fail-closed refresh hint, got %v", err)
	}
}

// The locale and wallpaper sections predate the gate, so the minimum must
// not be raised to a version that would gate the whole page.
func TestAppearanceMinVersionCoversOnlyTheNewSurfaces(t *testing.T) {
	if appearanceMinOlaresVersion != "1.12.6" {
		t.Fatalf("appearanceMinOlaresVersion = %q, want 1.12.6", appearanceMinOlaresVersion)
	}
}
