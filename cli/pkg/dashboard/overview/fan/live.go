package fan

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunLive is the cmd-side entry point for `dashboard overview fan
// live`. The watch-aware Runner runs the Olares One capability
// gate inside its per-iteration body so a `--watch` stream against
// the wrong device terminates with a clear empty envelope each
// tick rather than silently rendering zero rows.
func RunLive(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	r := &pkgdashboard.Runner{
		Flags:       cf,
		Recommended: 5 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (pkgdashboard.Envelope, error) {
			if gated, ok := pkgdashboard.GateOlaresOne(ctx, c, cf, pkgdashboard.KindOverviewFanLive, now, os.Stderr); ok {
				return gated, nil
			}
			env, err := BuildLiveEnvelope(ctx, c, cf, now)
			if err != nil {
				return env, err
			}
			env.Meta = pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User)
			env.Meta.RecommendedPollSeconds = 5
			if cf.Output == pkgdashboard.OutputJSON {
				return env, nil
			}
			return env, WriteLiveTable(os.Stdout, env)
		},
	}
	return r.Run(ctx)
}

// BuildLiveEnvelope mirrors SPA Overview2 fan card. The driver
// query is the system-fan endpoint (`pkgdashboard.FetchSystemFan`);
// GPU power values come from the optional graphics list (HAMI). A
// 404 from the fan endpoint is the "no fan integration" branch
// (Empty=true + EmptyReason="no_fan_integration") so JSON
// consumers can demux without inspecting the error string.
func BuildLiveEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, now time.Time) (pkgdashboard.Envelope, error) {
	fan, err := pkgdashboard.FetchSystemFan(ctx, c)
	if err != nil {
		if he, ok := pkgdashboard.IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env := pkgdashboard.Envelope{Kind: pkgdashboard.KindOverviewFanLive}
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_fan_integration"
			env.Meta.HTTPStatus = he.Status
			return env, nil
		}
		return pkgdashboard.Envelope{Kind: pkgdashboard.KindOverviewFanLive}, err
	}
	gpuPower, gpuPowerLimit := 0.0, 0.0
	if list, _ := pkgdashboard.FetchGraphicsList(ctx, c, nil); len(list) > 0 {
		if v, ok := list[0]["power"].(float64); ok {
			gpuPower = v
		}
		if v, ok := list[0]["powerLimit"].(float64); ok {
			gpuPowerLimit = v
		}
	}

	raw := map[string]any{
		"cpu_fan_rpm":     fan.CPUFanSpeed,
		"cpu_fan_rpm_max": pkgdashboard.FanSpeedMaxCPU,
		"cpu_temp_c":      fan.CPUTemperature,
		"gpu_fan_rpm":     fan.GPUFanSpeed,
		"gpu_fan_rpm_max": pkgdashboard.FanSpeedMaxGPU,
		"gpu_temp_c":      fan.GPUTemperature,
		"gpu_power":       gpuPower,
		"gpu_power_limit": gpuPowerLimit,
	}
	disp := map[string]any{
		"cpu_fan":       fmt.Sprintf("%.0f / %d RPM", fan.CPUFanSpeed, pkgdashboard.FanSpeedMaxCPU),
		"cpu_temp":      pkgdashboard.RenderTemperature(fan.CPUTemperature, cf.TempUnit),
		"gpu_fan":       fmt.Sprintf("%.0f / %d RPM", fan.GPUFanSpeed, pkgdashboard.FanSpeedMaxGPU),
		"gpu_temp":      pkgdashboard.RenderTemperature(fan.GPUTemperature, cf.TempUnit),
		"gpu_power":     fmt.Sprintf("%.2f W", gpuPower),
		"gpu_power_lim": fmt.Sprintf("%.0f W", gpuPowerLimit),
	}
	return pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewFanLive,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: []pkgdashboard.Item{{Raw: raw, Display: disp}},
	}, nil
}

// WriteLiveTable renders the 1-row live snapshot.
func WriteLiveTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "CPU_FAN", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "cpu_fan") }},
		{Header: "CPU_TEMP", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "cpu_temp") }},
		{Header: "GPU_FAN", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "gpu_fan") }},
		{Header: "GPU_TEMP", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "gpu_temp") }},
		{Header: "GPU_POWER", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "gpu_power") }},
		{Header: "POWER_LIM", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "gpu_power_lim") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
