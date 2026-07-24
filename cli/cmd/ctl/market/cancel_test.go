package market

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// TestGateResumeCancel pins the 1.12.7 fail-closed gate for cancelling a
// resuming app. The gate reuses cmdutil.RequireMinVersion, whose version
// resolution short-circuits on the version override viper key
// (FlagOlaresVersion) BEFORE any profile / network access — so a zero-value
// Factory plus a viper override is enough to drive every branch without a
// fake backend. A fresh Factory per scenario avoids the per-Factory
// memoization (backendVersionOnce) leaking a version across cases.
func TestGateResumeCancel(t *testing.T) {
	// Isolate the global viper override and restore it after the test so
	// other market tests are unaffected.
	prev := viper.GetString(cmdutil.FlagOlaresVersion)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, prev) })

	t.Run("non-resuming state is never gated", func(t *testing.T) {
		// installing must pass regardless of backend version — the gate
		// no-ops for it (returns before touching the factory), so even an
		// empty override cannot make it fail.
		viper.Set(cmdutil.FlagOlaresVersion, "1.12.5")
		for _, state := range []string{"installing", "downloading", "pending", "upgrading", ""} {
			if err := gateResumeCancel(context.Background(), &cmdutil.Factory{}, state); err != nil {
				t.Fatalf("state %q must not be gated, got error: %v", state, err)
			}
		}
	})

	t.Run("resuming on 1.12.6 backend is rejected", func(t *testing.T) {
		viper.Set(cmdutil.FlagOlaresVersion, "1.12.6")
		err := gateResumeCancel(context.Background(), &cmdutil.Factory{}, stateResuming)
		if err == nil {
			t.Fatalf("cancelling a resuming app on 1.12.6 must be rejected")
		}
		if !strings.Contains(err.Error(), resumeCancelMinVersion) {
			t.Fatalf("error should name the %s minimum, got %q", resumeCancelMinVersion, err.Error())
		}
	})

	t.Run("resuming on 1.12.7 backend is allowed", func(t *testing.T) {
		viper.Set(cmdutil.FlagOlaresVersion, "1.12.7")
		if err := gateResumeCancel(context.Background(), &cmdutil.Factory{}, stateResuming); err != nil {
			t.Fatalf("cancelling a resuming app on 1.12.7 must be allowed, got: %v", err)
		}
	})

	t.Run("resuming on a newer 1.12.7 daily build is allowed", func(t *testing.T) {
		// Core (major.minor.patch) comparison: a dated build on the 1.12.7
		// line counts as >= 1.12.7.
		viper.Set(cmdutil.FlagOlaresVersion, "1.12.7-20260604")
		if err := gateResumeCancel(context.Background(), &cmdutil.Factory{}, stateResuming); err != nil {
			t.Fatalf("cancelling a resuming app on a 1.12.7 daily build must be allowed, got: %v", err)
		}
	})
}
