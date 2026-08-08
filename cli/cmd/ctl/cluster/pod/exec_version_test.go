package pod

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func TestRequireExecBackendVersion(t *testing.T) {
	previous := viper.GetString(cmdutil.FlagOlaresVersion)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, previous) })

	viper.Set(cmdutil.FlagOlaresVersion, "1.12.6")
	err := requireExecBackendVersion(context.Background(), cmdutil.NewFactory())
	if err == nil {
		t.Fatal("expected version gate error")
	}
	for _, want := range []string{"cluster exec", "Olares >= 1.12.7", "this backend is 1.12.6"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err, want)
		}
	}

	viper.Set(cmdutil.FlagOlaresVersion, "1.12.7-20260808")
	if err := requireExecBackendVersion(context.Background(), cmdutil.NewFactory()); err != nil {
		t.Fatalf("daily build should pass: %v", err)
	}

	viper.Set(cmdutil.FlagOlaresVersion, "not-a-version")
	err = requireExecBackendVersion(context.Background(), cmdutil.NewFactory())
	if err == nil || !strings.Contains(err.Error(), "version could not be determined") {
		t.Fatalf("expected fail-closed version error, got %v", err)
	}
}
