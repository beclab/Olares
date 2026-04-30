package gpu

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunTask is the cmd-side entry point for `dashboard overview gpu
// task <name> <pod-uid>`. Returns a single flat task-detail Item;
// no gauge / trend fan-out (use `gpu task-detail` for that).
func RunTask(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, name, podUID, sharemode string) error {
	now := time.Now()
	advisoryNote, _ := pkgdashboard.GPUAdvisory(ctx, c, cf, os.Stderr)
	detail, err := pkgdashboard.FetchTaskDetail(ctx, c, name, podUID, sharemode)
	env := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewGPUTaskDet,
		Meta: pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
	}
	if advisoryNote != "" {
		env.Meta.Note = advisoryNote
	}
	if err != nil {
		if he, ok := pkgdashboard.IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			if cf.Output == pkgdashboard.OutputJSON {
				return pkgdashboard.WriteJSON(os.Stdout, env)
			}
			fmt.Fprintln(os.Stdout, "(task not found — HAMI integration absent or pod-uid invalid)")
			return nil
		}
		if unavail, ok := pkgdashboard.VgpuUnavailableFromError(c, cf, err, pkgdashboard.KindOverviewGPUTaskDet, now, os.Stderr); ok {
			if advisoryNote != "" {
				unavail.Meta.Note = advisoryNote + " | " + unavail.Meta.Note
			}
			if cf.Output == pkgdashboard.OutputJSON {
				return pkgdashboard.WriteJSON(os.Stdout, unavail)
			}
			return nil
		}
		return err
	}
	if len(detail) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		if cf.Output == pkgdashboard.OutputJSON {
			return pkgdashboard.WriteJSON(os.Stdout, env)
		}
		fmt.Fprintln(os.Stdout, "(no detail returned for this task)")
		return nil
	}
	env.Items = []pkgdashboard.Item{{Raw: detail, Display: detail}}
	if cf.Output == pkgdashboard.OutputJSON {
		return pkgdashboard.WriteJSON(os.Stdout, env)
	}
	return pkgdashboard.EmitDefault(env, cf.Output)
}
