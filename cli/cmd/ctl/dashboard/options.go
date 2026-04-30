package dashboard

import (
	"github.com/spf13/cobra"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// bindPersistentFlags wires every persistent pflag the dashboard tree
// needs onto cmd. The cobra-binding layer lives in the cmd package while
// the underlying CommonFlags struct + Validate logic sit in
// `cli/pkg/dashboard/flags.go`; this is the only place cobra and pkg
// `CommonFlags` meet.
//
// Why a free function (vs. a method on CommonFlags)? Because the data
// type itself stays cobra-free: tests under cli/pkg/dashboard exercise
// Validate / ResolveWindow without ever importing pflag. Any new flag
// gets added in two places: a field (or raw-string field) on
// CommonFlags, and one StringVar / DurationVar / etc. line here.
//
// Flag names + defaults match the pre-refactor surface 1:1 — see the
// SKILL.md "Public flag surface" section for the contract agents rely on.
func bindPersistentFlags(cf *pkgdashboard.CommonFlags, cmd *cobra.Command) {
	pf := cmd.PersistentFlags()
	pf.StringVarP(&cf.OutputRaw, "output", "o", "table",
		"output format (table or json)")
	pf.BoolVar(&cf.Watch, "watch", false,
		"poll the upstream endpoint and emit one envelope per iteration (NDJSON in JSON mode)")
	pf.DurationVar(&cf.WatchInterval, "watch-interval", 0,
		"interval between watch iterations (default: command's recommended-poll-seconds)")
	pf.IntVar(&cf.WatchIterations, "watch-iterations", 0,
		"stop after N iterations (0 = unbounded)")
	pf.DurationVar(&cf.WatchTimeout, "watch-timeout", 0,
		"stop after this much wall-clock time (0 = unbounded)")
	pf.StringVar(&cf.SinceRaw, "since", "",
		"relative window for metric commands; sliding when --watch (e.g. 5m, 1h)")
	pf.StringVar(&cf.StartRaw, "start", "",
		"absolute window start (RFC3339); fixed across iterations when --watch")
	pf.StringVar(&cf.EndRaw, "end", "",
		"absolute window end (RFC3339); fixed across iterations when --watch")
	pf.StringVar(&cf.TimezoneRaw, "timezone", "",
		"timezone for table rendering (IANA name, default: $TZ / system local)")
	pf.StringVar(&cf.TempUnitRaw, "temp-unit", "C",
		"temperature display unit: C, F, or K (JSON raw always Celsius)")
	pf.StringVar(&cf.User, "user", "",
		"target a different user than the active profile (platform-admin only)")
	pf.IntVar(&cf.Limit, "limit", 0,
		"page size for paginated endpoints (0 = upstream default)")
	pf.IntVar(&cf.Page, "page", 0,
		"page index for paginated endpoints (0 = first)")
	pf.IntVar(&cf.Head, "head", 0,
		"truncate output to the first N rows after sorting (0 = no truncation)")
}
