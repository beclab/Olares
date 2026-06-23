package preflight

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// factoryWithVersion returns a fresh Factory whose backend version resolves to
// `v` via the --olares-version override (which short-circuits before any
// profile/network lookup). Pass an unparseable string (e.g. "bad") to simulate
// an undetectable version. The override is reset when the test finishes.
func factoryWithVersion(t *testing.T, v string) *cmdutil.Factory {
	t.Helper()
	prev := viper.GetString(cmdutil.FlagOlaresVersion)
	viper.Set(cmdutil.FlagOlaresVersion, v)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, prev) })
	return cmdutil.NewFactory()
}

func TestRequireMinVersion(t *testing.T) {
	gate := MinVersionGate{
		Verb:       "settings compute",
		MinVersion: "1.12.6",
		Reason:     "compute-resources APIs",
		Fallback:   "use the legacy `olares-cli settings gpu list` on 1.12.5",
	}

	t.Run("at or above min passes", func(t *testing.T) {
		if err := RequireMinVersion(context.Background(), factoryWithVersion(t, "1.12.6"), gate); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		// A daily build of the same line counts as >= the line.
		if err := RequireMinVersion(context.Background(), factoryWithVersion(t, "1.12.6-20260603"), gate); err != nil {
			t.Fatalf("daily build: expected nil, got %v", err)
		}
	})

	t.Run("below min is rejected with reason + fallback", func(t *testing.T) {
		err := RequireMinVersion(context.Background(), factoryWithVersion(t, "1.12.5"), gate)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		msg := err.Error()
		for _, want := range []string{"settings compute", "requires Olares >= 1.12.6", "compute-resources APIs", "1.12.5", gate.Fallback} {
			if !strings.Contains(msg, want) {
				t.Fatalf("error %q missing %q", msg, want)
			}
		}
	})

	t.Run("undetectable version is fail-closed and suggests the flag", func(t *testing.T) {
		err := RequireMinVersion(context.Background(), factoryWithVersion(t, "not-a-version"), gate)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "could not be determined") ||
			!strings.Contains(err.Error(), cmdutil.FlagOlaresVersion) {
			t.Fatalf("error %q should mention undetectable + --%s", err.Error(), cmdutil.FlagOlaresVersion)
		}
	})

	t.Run("nil factory passes", func(t *testing.T) {
		if err := RequireMinVersion(context.Background(), nil, gate); err != nil {
			t.Fatalf("nil factory: expected nil, got %v", err)
		}
	})
}

func TestRejectIfRemoved(t *testing.T) {
	gate := RemovedGate{
		Verb:        "settings gpu list",
		Detail:      "legacy HAMI /api/gpu/list",
		RemovedIn:   "1.12.6",
		Replacement: "olares-cli settings compute resources list",
	}

	t.Run("at or after removal is rejected with replacement", func(t *testing.T) {
		err := RejectIfRemoved(context.Background(), factoryWithVersion(t, "1.12.6"), gate)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		msg := err.Error()
		for _, want := range []string{"settings gpu list", "legacy HAMI /api/gpu/list", "removed in Olares 1.12.6", gate.Replacement} {
			if !strings.Contains(msg, want) {
				t.Fatalf("error %q missing %q", msg, want)
			}
		}
	})

	t.Run("before removal passes (legacy still works)", func(t *testing.T) {
		if err := RejectIfRemoved(context.Background(), factoryWithVersion(t, "1.12.5"), gate); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("undetectable version is fail-open", func(t *testing.T) {
		if err := RejectIfRemoved(context.Background(), factoryWithVersion(t, "not-a-version"), gate); err != nil {
			t.Fatalf("undetectable should fail-open (nil), got %v", err)
		}
	})

	t.Run("nil factory passes", func(t *testing.T) {
		if err := RejectIfRemoved(context.Background(), nil, gate); err != nil {
			t.Fatalf("nil factory: expected nil, got %v", err)
		}
	})
}
