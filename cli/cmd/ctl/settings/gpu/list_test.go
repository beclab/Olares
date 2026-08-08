package gpu

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func TestListRejectsRemovedBackendVersion(t *testing.T) {
	previous := viper.GetString(cmdutil.FlagOlaresVersion)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, previous) })
	viper.Set(cmdutil.FlagOlaresVersion, "1.12.6")

	cmd := NewListCommand(cmdutil.NewFactory())
	cmd.SetContext(context.Background())
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected removed-command error")
	}
	for _, want := range []string{"was removed in Olares 1.12.6", "olares-cli settings compute list"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err, want)
		}
	}
}
