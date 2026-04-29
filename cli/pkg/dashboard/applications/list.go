// Package applications hosts the business logic for the
// `olares-cli dashboard applications` cobra leaf — a workload-grain
// table mirroring the SPA's Applications page.
//
// The cmd-side area (cli/cmd/ctl/dashboard/applications/) is a thin
// shell: it owns cobra wiring, persistent-flag inheritance, and the
// HTTP client factory. Every business seam — flag-enum validation,
// upstream fan-out, envelope shaping, table rendering, watch-loop
// assembly — lives here so this package is independently testable
// against an httptest.Server without dragging cobra + cmdutil.Factory
// into the test fixture.
//
// The wire path reuses pkgdashboard.BuildRankingEnvelope (the one
// legitimate horizontal share between cmd-area subpackages, hoisted
// to the top-level pkg per SKILL.md) so the first row of
// `dashboard applications` matches `dashboard overview ranking` by
// construction; we just re-tag the envelope kind and add the
// per-iteration polling cadence.
package applications

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// recommendedPollSeconds is the SPA-side polling cadence for the
// Applications page. Surfaced both via Meta.RecommendedPollSeconds
// (so JSON consumers can mirror the cadence) and as the Runner's
// default --watch-interval.
const recommendedPollSeconds = 60

// RunList is the cmd-side entry point. It owns:
//
//   - business flag-enum validation (--sort, --sort-by) — error
//     wording is 1:1 with the previous cmd-side body so the
//     test-suite-pinned messages don't shift;
//   - watch-aware Runner assembly so cmd-side never sees Runner;
//   - per-iteration envelope construction (delegated to
//     BuildListEnvelope) and table-mode rendering (delegated to
//     WriteListTable). JSON-mode output is handled by Runner.emit.
func RunList(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, sortBy, sortDir string) error {
	if err := validateListFlags(sortBy, sortDir); err != nil {
		return err
	}
	r := &pkgdashboard.Runner{
		Flags:       cf,
		Recommended: recommendedPollSeconds * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (pkgdashboard.Envelope, error) {
			env, err := BuildListEnvelope(ctx, c, cf, sortBy, sortDir, now)
			if err != nil {
				return env, err
			}
			if cf.Output == pkgdashboard.OutputJSON {
				return env, nil
			}
			return env, WriteListTable(os.Stdout, env)
		},
	}
	return r.Run(ctx)
}

// validateListFlags pins the --sort / --sort-by enums. Lifted out of
// RunList so unit tests can hit it without spinning up an
// httptest.Server.
func validateListFlags(sortBy, sortDir string) error {
	if sortDir != "asc" && sortDir != "desc" {
		return fmt.Errorf("--sort: %q is not asc/desc", sortDir)
	}
	switch sortBy {
	case "cpu", "memory", "net_in", "net_out":
		return nil
	default:
		return fmt.Errorf("--sort-by: %q is not cpu|memory|net_in|net_out", sortBy)
	}
}

// BuildListEnvelope is the per-iteration envelope builder. Re-tags the
// shared ranking envelope as KindApplicationsList, threads
// Meta.RecommendedPollSeconds + Meta.User, and applies --head
// truncation. Exported so the package's _test.go can drive it
// directly against a stub upstream without going through the watch
// loop.
func BuildListEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, sortBy, sortDir string, now time.Time) (pkgdashboard.Envelope, error) {
	rankEnv, err := pkgdashboard.BuildRankingEnvelope(ctx, c, cf, cf.User, sortBy, sortDir, now)
	if err != nil {
		return rankEnv, err
	}
	env := pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindApplicationsList,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: rankEnv.Items,
	}
	env.Meta.RecommendedPollSeconds = recommendedPollSeconds
	env.Items = pkgdashboard.HeadItems(env.Items, cf.Head)
	return env, nil
}

// WriteListTable renders env.Items as the SPA-aligned 9-column
// workload table. Exported so the package's _test.go can assert on a
// captured buffer without redirecting os.Stdout. Column order
// (RANK / APP / NAMESPACE / STATE / PODS / CPU / MEMORY / NET_IN /
// NET_OUT) is pinned: it's the documented contract every
// `applications --output table` consumer scrapes.
func WriteListTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "RANK", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "rank") }},
		{Header: "APP", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "app") }},
		{Header: "NAMESPACE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "namespace") }},
		{Header: "STATE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "state") }},
		{Header: "PODS", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "pods") }},
		{Header: "CPU", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "cpu") }},
		{Header: "MEMORY", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "memory") }},
		{Header: "NET_IN", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "net_in") }},
		{Header: "NET_OUT", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "net_out") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
