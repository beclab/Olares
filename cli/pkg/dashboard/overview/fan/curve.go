package fan

import (
	"context"
	"io"
	"os"
	"strconv"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunCurve is the cmd-side entry point for `dashboard overview fan
// curve`. The hardware fan-curve table is hardcoded
// (pkgdashboard.FanCurveTable); the leaf still owns the Olares-One
// gate because the curve spec is hardware-specific reference data
// that isn't meaningful on a non-Olares-One device — per the user's
// policy decision the gate fires before the (free) build step.
//
// One-shot only: the curve is static, watch would just re-print the
// same table forever. The cmd-side never triggers Runner.
func RunCurve(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	now := time.Now()
	if gated, ok := pkgdashboard.GateOlaresOne(ctx, c, cf, pkgdashboard.KindOverviewFanCurve, now, os.Stderr); ok {
		if cf.Output == pkgdashboard.OutputJSON {
			return pkgdashboard.WriteJSON(os.Stdout, gated)
		}
		return nil
	}
	env := BuildCurveEnvelope(cf, c.OlaresID(), now)
	if cf.Output == pkgdashboard.OutputJSON {
		return pkgdashboard.WriteJSON(os.Stdout, env)
	}
	return WriteCurveTable(os.Stdout, env)
}

// BuildCurveEnvelope is pure: no HTTP, no clock dependency beyond
// the Meta timestamp. Driven by pkgdashboard.FanCurveTable so the
// curve numbers stay version-controlled in one place.
func BuildCurveEnvelope(cf *pkgdashboard.CommonFlags, olaresID string, now time.Time) pkgdashboard.Envelope {
	items := make([]pkgdashboard.Item, 0, len(pkgdashboard.FanCurveTable))
	for _, r := range pkgdashboard.FanCurveTable {
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
		items = append(items, pkgdashboard.Item{Raw: raw, Display: disp})
	}
	return pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewFanCurve,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), olaresID, cf.User),
		Items: items,
	}
}

// WriteCurveTable renders the 10-row fan-curve spec.
func WriteCurveTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "STEP", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "step") }},
		{Header: "CPU_RPM", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "cpu_fan_rpm") }},
		{Header: "GPU_RPM", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "gpu_fan_rpm") }},
		{Header: "CPU_TEMP", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "cpu_temp_range") }},
		{Header: "GPU_TEMP", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "gpu_temp_range") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
