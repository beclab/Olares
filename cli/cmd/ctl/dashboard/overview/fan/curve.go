package fan

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newOverviewFanCurveCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "curve",
		Short:         "10-row hardcoded fan-curve specification (RPM ↔ temperature range)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewFanCurve(c.Context(), f)
		},
	}
	return cmd
}

func runOverviewFanCurve(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	if gated, ok := gateOlaresOne(ctx, c, KindOverviewFanCurve, now); ok {
		if common.Output == OutputJSON {
			return WriteJSON(os.Stdout, gated)
		}
		return nil
	}
	env := buildFanCurveEnvelope(now, c.OlaresID())
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	return writeFanCurveTable(env)
}

func buildFanCurveEnvelope(now time.Time, olaresID string) Envelope {
	items := make([]Item, 0, len(fanCurveTable))
	for _, r := range fanCurveTable {
		raw := map[string]any{
			"step":           r.Step,
			"cpu_fan_rpm":    r.CPUFanRPM,
			"gpu_fan_rpm":    r.GPUFanRPM,
			"cpu_temp_range": r.CPUTempRange,
			"gpu_temp_range": r.GPUTempRange,
		}
		disp := map[string]any{
			"step":           strconv.Itoa(r.Step),
			"cpu_fan_rpm":    strconv.Itoa(r.CPUFanRPM),
			"gpu_fan_rpm":    strconv.Itoa(r.GPUFanRPM),
			"cpu_temp_range": r.CPUTempRange,
			"gpu_temp_range": r.GPUTempRange,
		}
		items = append(items, Item{Raw: raw, Display: disp})
	}
	return Envelope{
		Kind:  KindOverviewFanCurve,
		Meta:  NewMeta(now.In(common.Timezone.Time()), olaresID, common.User),
		Items: items,
	}
}

func writeFanCurveTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "STEP", Get: func(it Item) string { return DisplayString(it, "step") }},
		{Header: "CPU_RPM", Get: func(it Item) string { return DisplayString(it, "cpu_fan_rpm") }},
		{Header: "GPU_RPM", Get: func(it Item) string { return DisplayString(it, "gpu_fan_rpm") }},
		{Header: "CPU_TEMP", Get: func(it Item) string { return DisplayString(it, "cpu_temp_range") }},
		{Header: "GPU_TEMP", Get: func(it Item) string { return DisplayString(it, "gpu_temp_range") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}
