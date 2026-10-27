package network

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func TestRequireOverlayBackendVersion(t *testing.T) {
	previous := viper.GetString(cmdutil.FlagOlaresVersion)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, previous) })

	viper.Set(cmdutil.FlagOlaresVersion, "1.12.5")
	err := requireOverlayBackendVersion(context.Background(), cmdutil.NewFactory())
	if err == nil {
		t.Fatal("expected version gate error")
	}
	for _, want := range []string{"settings network overlay", "requires Olares >= 1.12.6", "upgrade the Olares system"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err, want)
		}
	}

	viper.Set(cmdutil.FlagOlaresVersion, "1.12.6-20260808")
	if err := requireOverlayBackendVersion(context.Background(), cmdutil.NewFactory()); err != nil {
		t.Fatalf("daily build should pass: %v", err)
	}

	viper.Set(cmdutil.FlagOlaresVersion, "not-a-version")
	err = requireOverlayBackendVersion(context.Background(), cmdutil.NewFactory())
	if err == nil || !strings.Contains(err.Error(), "profile list --refresh-version") {
		t.Fatalf("expected fail-closed refresh hint, got %v", err)
	}
}
