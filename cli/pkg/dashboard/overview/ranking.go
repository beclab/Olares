package overview

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunRanking is the cmd-side entry point for `dashboard overview
// ranking`. The leaf only exposes --sort (sortDir); sortBy is hard-
// coded to "cpu" to match the SPA's UsageRanking widget. cf.Head is
// honoured here (a long ranking can be trimmed client-side without a
// re-fetch).
func RunRanking(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, sortDir string) error {
	if sortDir != "asc" && sortDir != "desc" {
		return fmt.Errorf("--sort: %q is not asc/desc", sortDir)
	}
	r := &pkgdashboard.Runner{
		Flags:       cf,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (pkgdashboard.Envelope, error) {
			env, err := BuildRankingEnvelope(ctx, c, cf, sortDir, now)
			if err != nil {
				return env, err
			}
			env.Meta = pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User)
			env.Meta.RecommendedPollSeconds = 60
			env.Items = pkgdashboard.HeadItems(env.Items, cf.Head)
			if cf.Output == pkgdashboard.OutputJSON {
				return env, nil
			}
			return env, WriteRankingTable(os.Stdout, env)
		},
	}
	return r.Run(ctx)
}

// BuildRankingEnvelope is the overview-area's thin wrapper over the
// shared pkgdashboard.BuildRankingEnvelope. Hard-codes sortBy="cpu"
// so the per-leaf wire path matches the SPA's UsageRanking widget;
// `dashboard applications` exposes sortBy as a flag so it stays out
// of this file.
func BuildRankingEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, sortDir string, now time.Time) (pkgdashboard.Envelope, error) {
	return pkgdashboard.BuildRankingEnvelope(ctx, c, cf, cf.User, "cpu", sortDir, now)
}

// WriteRankingTable renders env.Items as the SPA-aligned 7-column
// workload table (RANK / APP / NAMESPACE / CPU / MEMORY / NET_IN /
// NET_OUT). The applications view emits a 9-column variant (state +
// pods); we deliberately keep this leaf trimmed to match the SPA's
// UsageRanking widget.
func WriteRankingTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "RANK", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "rank") }},
		{Header: "APP", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "app") }},
		{Header: "NAMESPACE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "namespace") }},
		{Header: "CPU", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "cpu") }},
		{Header: "MEMORY", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "memory") }},
		{Header: "NET_IN", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "net_in") }},
		{Header: "NET_OUT", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "net_out") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
