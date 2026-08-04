package market

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// TestGateInFlightCancel pins the 1.12.7 fail-closed gate for cancelling a
// resuming or an upgrading app. The gate reuses cmdutil.RequireMinVersion,
// whose version resolution short-circuits on the version override viper key
// (FlagOlaresVersion) BEFORE any profile / network access — so a zero-value
// Factory plus a viper override is enough to drive every branch without a
// fake backend. A fresh Factory per scenario avoids the per-Factory
// memoization (backendVersionOnce) leaking a version across cases.
func TestGateInFlightCancel(t *testing.T) {
	// Isolate the global viper override and restore it after the test so
	// other market tests are unaffected.
	prev := viper.GetString(cmdutil.FlagOlaresVersion)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, prev) })

	t.Run("ungated states are never gated", func(t *testing.T) {
		// The rest of the in-flight pipeline must pass regardless of
		// backend version — the gate no-ops for those states (returns
		// before touching the factory), so even an empty override cannot
		// make them fail.
		viper.Set(cmdutil.FlagOlaresVersion, "1.12.5")
		for _, state := range []string{"installing", "downloading", "pending", "initializing", "applyingEnv", ""} {
			if err := gateInFlightCancel(context.Background(), &cmdutil.Factory{}, state); err != nil {
				t.Fatalf("state %q must not be gated, got error: %v", state, err)
			}
		}
	})

	for _, state := range []string{stateResuming, stateUpgrading} {
		minVersion := cancelGateByState[state].minVersion

		t.Run(state+" on 1.12.6 backend is rejected", func(t *testing.T) {
			viper.Set(cmdutil.FlagOlaresVersion, "1.12.6")
			err := gateInFlightCancel(context.Background(), &cmdutil.Factory{}, state)
			if err == nil {
				t.Fatalf("cancelling a %s app on 1.12.6 must be rejected", state)
			}
			if !strings.Contains(err.Error(), minVersion) {
				t.Fatalf("error should name the %s minimum, got %q", minVersion, err.Error())
			}
		})

		t.Run(state+" on 1.12.7 backend is allowed", func(t *testing.T) {
			viper.Set(cmdutil.FlagOlaresVersion, "1.12.7")
			if err := gateInFlightCancel(context.Background(), &cmdutil.Factory{}, state); err != nil {
				t.Fatalf("cancelling a %s app on 1.12.7 must be allowed, got: %v", state, err)
			}
		})

		t.Run(state+" on a newer 1.12.7 daily build is allowed", func(t *testing.T) {
			// Core (major.minor.patch) comparison: a dated build on the
			// 1.12.7 line counts as >= 1.12.7.
			viper.Set(cmdutil.FlagOlaresVersion, "1.12.7-20260604")
			if err := gateInFlightCancel(context.Background(), &cmdutil.Factory{}, state); err != nil {
				t.Fatalf("cancelling a %s app on a 1.12.7 daily build must be allowed, got: %v", state, err)
			}
		})
	}
}
