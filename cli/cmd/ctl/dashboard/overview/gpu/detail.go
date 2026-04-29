package gpu

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newOverviewGPUDetailFullCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "detail <uuid>",
		Short:         "Per-GPU detail page (info + gauges + trends; SPA Overview2/GPU/GPUsDetails)",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUDetailFull(c.Context(), f, args[0])
		},
	}
	return cmd
}

func runOverviewGPUDetailFull(ctx context.Context, f *cmdutil.Factory, uuid string) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       common,
		Recommended: 30 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			start, end, since := resolveGPUDetailWindow(now, gpuDetailDefaultSince)
			env, err := buildGPUDetailFullEnvelope(ctx, c, uuid, start, end, since)
			if err != nil {
				return env, err
			}
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeGPUDetailFullTable(env)
		},
	}
	return r.Run(ctx)
}
