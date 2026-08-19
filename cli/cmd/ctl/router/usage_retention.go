package router

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router usage retention` — how long the per-call rows are kept.
//
// GET /console/api/spend-settings
// PUT /console/api/spend-settings
//
// Router keeps usage in two shapes. A daily total per model, person, provider
// and application is kept for good, and one row per call is kept for a window.
// This is that window, and shortening it takes effect at once: the sweep is
// woken rather than left to its next hour.
//
// So `usage summary` keeps answering for a period whose `usage list` has been
// swept, and the two disagreeing is not a bug — it is what this setting is for.
// Zero is a real value: it keeps no per-call rows at all, which leaves the
// totals, the quotas and the report intact and `usage list` permanently empty.
//
// Admin only, the read included: how long records live is a property of the
// deployment, and somebody reading their own spend has no decision to make with
// it.

// spendSettings is the window plus the range it may be set to, which arrives
// with it so a caller does not need a second source for the bounds.
type spendSettings struct {
	RetentionDays int `json:"retention_days"`
	Min           int `json:"min"`
	Max           int `json:"max"`
	Default       int `json:"default"`
}

func newUsageRetentionCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		days   int
	)
	cmd := &cobra.Command{
		Use:   "retention",
		Short: "how long the individual calls are kept",
		Long: `Report or change how long Router keeps one row per call.

Daily totals are kept for good; the per-call rows behind them are deleted on this
window. That is why "usage summary" can answer for a month that "usage list" no
longer has the calls for.

--days sets it, within the range Router reports. 0 is a real setting and keeps no
per-call rows at all: the summary, the quotas and the export by day still work,
and "usage list" is empty from then on.

A shorter window applies immediately — rows outside it are deleted while you are
reading the answer, not at the next sweep.

Admin only, reading included.

Examples:
  olares-cli router usage retention
  olares-cli router usage retention --days 30
  olares-cli router usage retention --days 0
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			var want *int
			if c.Flags().Changed("days") {
				want = &days
			}
			return runUsageRetention(c.Context(), f, want, output)
		},
	}
	cmd.Flags().IntVar(&days, "days", 0, "keep per-call rows for this many days; 0 keeps none")
	addOutputFlag(cmd, &output)
	return cmd
}

func runUsageRetention(ctx context.Context, f *cmdutil.Factory, days *int, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var cur spendSettings
	if err := pc.router.doJSON(ctx, "GET", epSpendSettings, nil, &cur); err != nil {
		return err
	}
	if days == nil {
		if format == FormatJSON {
			return printJSON(os.Stdout, cur)
		}
		return renderRetention(os.Stdout, &cur, false)
	}
	// Checked here as well as by Router, because the bounds arrived with the
	// current value: refusing locally names the range in the same breath.
	if *days < cur.Min || *days > cur.Max {
		return fmt.Errorf("--days is between %d and %d on this Router, not %d", cur.Min, cur.Max, *days)
	}
	if *days == cur.RetentionDays {
		if format == FormatJSON {
			return printJSON(os.Stdout, cur)
		}
		_, werr := fmt.Fprintf(os.Stdout, "per-call rows are already kept for %s.\n", retentionPhrase(cur.RetentionDays))
		return werr
	}
	shorter := *days < cur.RetentionDays
	var updated spendSettings
	if err := pc.router.doJSON(ctx, "PUT", epSpendSettings,
		map[string]any{"retention_days": *days}, &updated); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, updated)
	}
	return renderRetention(os.Stdout, &updated, shorter)
}

func renderRetention(w io.Writer, s *spendSettings, swept bool) error {
	if _, err := fmt.Fprintf(w, "per-call rows are kept for %s (allowed %d-%d, default %d).\n",
		retentionPhrase(s.RetentionDays), s.Min, s.Max, s.Default); err != nil {
		return err
	}
	if swept {
		if _, err := fmt.Fprintln(w, "Rows outside the new window are being deleted now."); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, "Daily totals are kept regardless, so `usage summary` still answers for "+
		"periods `usage list` no longer holds.")
	return err
}

func retentionPhrase(days int) string {
	switch days {
	case 0:
		return "no time at all"
	case 1:
		return "1 day"
	default:
		return fmt.Sprintf("%d days", days)
	}
}
