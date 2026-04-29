package fan

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// ----------------------------------------------------------------------------
// overview fan — sections (live + curve)
// ----------------------------------------------------------------------------

func NewFanCommand(f *cmdutil.Factory, cf *pkgdashboard.CommonFlags) *cobra.Command {
	common = cf
	cmd := &cobra.Command{
		Use:           "fan",
		Short:         "Sections envelope: live = real-time fan/temperature/power; curve = hardcoded fan-curve spec",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownSubcommandRunE(c, args)
			}
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewFanDefault(c.Context(), f)
		},
	}
	cmd.AddCommand(newOverviewFanLiveCommand(f))
	cmd.AddCommand(newOverviewFanCurveCommand(f))
	return cmd
}

func runOverviewFanDefault(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()

	// Capability gate: Fan / cooling features are Olares One-only.
	// Mirrors `FanStore.isOlaresOneDevice` in
	// `Overview2/ClusterResource.vue:238`. Both `live` and `curve`
	// share the gate per the user's policy decision (curve is
	// hardware-specific spec, not portable reference data).
	if gated, ok := gateOlaresOne(ctx, c, KindOverviewFan, now); ok {
		// The aggregate envelope mirrors the live + curve sections,
		// each carrying the same `not_olares_one` reason so consumers
		// can demux either at the top or per-section.
		liveGated := gated
		liveGated.Kind = KindOverviewFanLive
		curveGated := gated
		curveGated.Kind = KindOverviewFanCurve
		gated.Sections = map[string]Envelope{
			"live":  liveGated,
			"curve": curveGated,
		}
		if common.Output == OutputJSON {
			return WriteJSON(os.Stdout, gated)
		}
		return nil
	}

	live, lerr := buildFanLiveEnvelope(ctx, c, now)
	if lerr != nil {
		live.Kind = KindOverviewFanLive
		live.Meta.Error = lerr.Error()
		live.Meta.ErrorKind = ClassifyTransportErr(lerr)
	}
	curve := buildFanCurveEnvelope(now, c.OlaresID())

	env := Envelope{
		Kind: KindOverviewFan,
		Meta: NewMeta(time.Now().In(common.Timezone.Time()), c.OlaresID(), common.User),
		Sections: map[string]Envelope{
			"live":  live,
			"curve": curve,
		},
	}
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	fmt.Fprintln(os.Stdout, "== LIVE ==")
	if lerr != nil {
		fmt.Fprintf(os.Stdout, "(error: %s)\n", lerr)
	} else {
		_ = writeFanLiveTable(live)
	}
	fmt.Fprintln(os.Stdout, "\n== CURVE ==")
	return writeFanCurveTable(curve)
}
