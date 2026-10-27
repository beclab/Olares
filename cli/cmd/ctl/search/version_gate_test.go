package search

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// TestRequireSessionAppBackendVersion pins the 1.12.7 fail-closed gate for
// search knowledge. The gate reuses
// cmdutil.RequireMinVersion, whose version resolution short-circuits on the
// version override viper key (FlagOlaresVersion) BEFORE any profile /
// network access — so a zero-value Factory plus a viper override is enough
// to drive every branch without a fake backend. A fresh Factory per
// scenario avoids the per-Factory memoization (backendVersionOnce) leaking
// a version across cases.
func TestRequireSessionAppBackendVersion(t *testing.T) {
	prev := viper.GetString(cmdutil.FlagOlaresVersion)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, prev) })

	verbs := []struct {
		verb   string
		reason string
	}{
		{"search knowledge", "search3 app=knowledge index"},
	}

	for _, v := range verbs {
		v := v

		t.Run(v.verb+" on 1.12.6 backend is rejected", func(t *testing.T) {
			viper.Set(cmdutil.FlagOlaresVersion, "1.12.6")
			err := requireSessionAppBackendVersion(context.Background(), &cmdutil.Factory{}, v.verb, v.reason)
			if err == nil {
				t.Fatalf("%s on 1.12.6 must be rejected", v.verb)
			}
			if !strings.Contains(err.Error(), searchSessionAppMinOlaresVersion) {
				t.Fatalf("error should name the %s minimum, got %q", searchSessionAppMinOlaresVersion, err.Error())
			}
			if !strings.Contains(err.Error(), v.verb) {
				t.Fatalf("error should name the verb %q, got %q", v.verb, err.Error())
			}
		})

		t.Run(v.verb+" on 1.12.7 backend is allowed", func(t *testing.T) {
			viper.Set(cmdutil.FlagOlaresVersion, "1.12.7")
			if err := requireSessionAppBackendVersion(context.Background(), &cmdutil.Factory{}, v.verb, v.reason); err != nil {
				t.Fatalf("%s on 1.12.7 must be allowed, got: %v", v.verb, err)
			}
		})

		t.Run(v.verb+" on a newer 1.12.7 daily build is allowed", func(t *testing.T) {
			viper.Set(cmdutil.FlagOlaresVersion, "1.12.7-20260604")
			if err := requireSessionAppBackendVersion(context.Background(), &cmdutil.Factory{}, v.verb, v.reason); err != nil {
				t.Fatalf("%s on a 1.12.7 daily build must be allowed, got: %v", v.verb, err)
			}
		})

		t.Run(v.verb+" on undetectable version is fail-closed", func(t *testing.T) {
			viper.Set(cmdutil.FlagOlaresVersion, "not-a-version")
			err := requireSessionAppBackendVersion(context.Background(), &cmdutil.Factory{}, v.verb, v.reason)
			if err == nil {
				t.Fatalf("%s with undetectable version must be rejected", v.verb)
			}
		})
	}
}
