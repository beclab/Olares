package gpu

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newOverviewGPUGetCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "get <uuid>",
		Short:         "Per-GPU detail by UUID",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUGet(c.Context(), f, args[0])
		},
	}
	return cmd
}

func runOverviewGPUGet(ctx context.Context, f *cmdutil.Factory, uuid string) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	advisoryNote, _ := gpuAdvisory(ctx, c)
	detail, err := fetchGraphicsDetail(ctx, c, uuid)
	env := Envelope{
		Kind: KindOverviewGPUDetail,
		Meta: NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
	}
	if advisoryNote != "" {
		env.Meta.Note = advisoryNote
	}
	if err != nil {
		if he, ok := IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, env)
			}
			fmt.Fprintln(os.Stdout, "(GPU not found — HAMI integration absent or UUID invalid)")
			return nil
		}
		if unavail, ok := vgpuUnavailableFromError(c, err, KindOverviewGPUDetail, now); ok {
			if advisoryNote != "" {
				unavail.Meta.Note = advisoryNote + " | " + unavail.Meta.Note
			}
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, unavail)
			}
			return nil
		}
		return err
	}
	if len(detail) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		if common.Output == OutputJSON {
			return WriteJSON(os.Stdout, env)
		}
		fmt.Fprintln(os.Stdout, "(no detail returned for this GPU UUID)")
		return nil
	}
	env.Items = []Item{{
		Raw:     detail,
		Display: detail,
	}}
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	return EmitDefault(env, common.Output)
}
