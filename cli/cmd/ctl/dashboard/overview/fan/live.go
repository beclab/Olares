package fan

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newOverviewFanLiveCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "live",
		Short:         "1-row real-time fan / temperature / power snapshot (Olares One)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewFanLive(c.Context(), f)
		},
	}
	return cmd
}

func runOverviewFanLive(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       common,
		Recommended: 5 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			// Capability gate runs each iteration so a `--watch`
			// stream against the wrong device terminates with a
			// clear empty envelope per tick rather than silent zeros.
			if gated, ok := gateOlaresOne(ctx, c, KindOverviewFanLive, now); ok {
				return gated, nil
			}
			env, err := buildFanLiveEnvelope(ctx, c, now)
			if err != nil {
				return env, err
			}
			env.Meta = NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)
			env.Meta.RecommendedPollSeconds = 5
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeFanLiveTable(env)
		},
	}
	return r.Run(ctx)
}

func buildFanLiveEnvelope(ctx context.Context, c *Client, now time.Time) (Envelope, error) {
	fan, err := fetchSystemFan(ctx, c)
	if err != nil {
		// 404 → no fan integration. Surface the empty envelope so the
		// caller's three-state branch works.
		if he, ok := IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env := Envelope{Kind: KindOverviewFanLive}
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_fan_integration"
			env.Meta.HTTPStatus = he.Status
			return env, nil
		}
		return Envelope{Kind: KindOverviewFanLive}, err
	}
	gpuPower, gpuPowerLimit := 0.0, 0.0
	if list, _ := fetchGraphicsList(ctx, c, nil); len(list) > 0 {
		if v, ok := list[0]["power"].(float64); ok {
			gpuPower = v
		}
		if v, ok := list[0]["powerLimit"].(float64); ok {
			gpuPowerLimit = v
		}
	}

	raw := map[string]any{
		"cpu_fan_rpm":     fan.CPUFanSpeed,
		"cpu_fan_rpm_max": fanSpeedMaxCPU,
		"cpu_temp_c":      fan.CPUTemperature,
		"gpu_fan_rpm":     fan.GPUFanSpeed,
		"gpu_fan_rpm_max": fanSpeedMaxGPU,
		"gpu_temp_c":      fan.GPUTemperature,
		"gpu_power":       gpuPower,
		"gpu_power_limit": gpuPowerLimit,
	}
	disp := map[string]any{
		"cpu_fan":       fmt.Sprintf("%.0f / %d RPM", fan.CPUFanSpeed, fanSpeedMaxCPU),
		"cpu_temp":      renderTemperature(fan.CPUTemperature, common.TempUnit),
		"gpu_fan":       fmt.Sprintf("%.0f / %d RPM", fan.GPUFanSpeed, fanSpeedMaxGPU),
		"gpu_temp":      renderTemperature(fan.GPUTemperature, common.TempUnit),
		"gpu_power":     fmt.Sprintf("%.2f W", gpuPower),
		"gpu_power_lim": fmt.Sprintf("%.0f W", gpuPowerLimit),
	}
	return Envelope{
		Kind:  KindOverviewFanLive,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: []Item{{Raw: raw, Display: disp}},
	}, nil
}

func writeFanLiveTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "CPU_FAN", Get: func(it Item) string { return DisplayString(it, "cpu_fan") }},
		{Header: "CPU_TEMP", Get: func(it Item) string { return DisplayString(it, "cpu_temp") }},
		{Header: "GPU_FAN", Get: func(it Item) string { return DisplayString(it, "gpu_fan") }},
		{Header: "GPU_TEMP", Get: func(it Item) string { return DisplayString(it, "gpu_temp") }},
		{Header: "GPU_POWER", Get: func(it Item) string { return DisplayString(it, "gpu_power") }},
		{Header: "POWER_LIM", Get: func(it Item) string { return DisplayString(it, "gpu_power_lim") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}
