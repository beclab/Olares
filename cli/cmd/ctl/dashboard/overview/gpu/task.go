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

func newOverviewGPUTaskCommand(f *cmdutil.Factory) *cobra.Command {
	var sharemode string
	cmd := &cobra.Command{
		Use:           "task <name> <pod-uid>",
		Short:         "Per-task detail (pod-uid from `kubectl get pods -n <ns> -o jsonpath='{.items[*].metadata.uid}'`)",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUTask(c.Context(), f, args[0], args[1], sharemode)
		},
	}
	cmd.Flags().StringVar(&sharemode, "sharemode", "", "task share mode (passed to /v1/container?sharemode=)")
	return cmd
}

func runOverviewGPUTask(ctx context.Context, f *cmdutil.Factory, name, podUID, sharemode string) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	advisoryNote, _ := gpuAdvisory(ctx, c)
	detail, err := fetchTaskDetail(ctx, c, name, podUID, sharemode)
	env := Envelope{
		Kind: KindOverviewGPUTaskDet,
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
			fmt.Fprintln(os.Stdout, "(task not found — HAMI integration absent or pod-uid invalid)")
			return nil
		}
		if unavail, ok := vgpuUnavailableFromError(c, err, KindOverviewGPUTaskDet, now); ok {
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
		fmt.Fprintln(os.Stdout, "(no detail returned for this task)")
		return nil
	}
	env.Items = []Item{{Raw: detail, Display: detail}}
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	return EmitDefault(env, common.Output)
}
