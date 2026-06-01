package gpu

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunGet is the cmd-side entry point for `dashboard overview gpu
// get <uuid>`. Returns a single flat detail Item — no gauge / trend
// fan-out (use `gpu detail <uuid>` for that).
func RunGet(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, uuid string) error {
	now := time.Now()
	advisoryNote, _ := pkgdashboard.GPUAdvisory(ctx, c, cf, os.Stderr)
	detail, err := pkgdashboard.FetchGraphicsDetail(ctx, c, uuid)
	env := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewGPUDetail,
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
			fmt.Fprintln(os.Stdout, "(GPU not found — HAMI integration absent or UUID invalid)")
			return nil
		}
		if unavail, ok := pkgdashboard.VgpuUnavailableFromError(c, cf, err, pkgdashboard.KindOverviewGPUDetail, now, os.Stderr); ok {
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
		fmt.Fprintln(os.Stdout, "(no detail returned for this GPU UUID)")
		return nil
	}
	env.Items = []pkgdashboard.Item{{
		Raw:     detail,
		Display: detail,
	}}
	if cf.Output == pkgdashboard.OutputJSON {
		return pkgdashboard.WriteJSON(os.Stdout, env)
	}
	return pkgdashboard.EmitDefault(env, cf.Output)
}
