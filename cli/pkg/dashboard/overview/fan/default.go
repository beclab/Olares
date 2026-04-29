package fan

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunDefault is the cmd-side entry point for `dashboard overview
// fan` without a subverb. Emits a sections envelope: live + curve.
// Both sections share the same Olares-One capability gate (the
// curve is hardware-specific reference data, not portable). When
// gated, the parent envelope mirrors per-section gating so JSON
// consumers can demux either at the top or per-section.
//
// One-shot only: live and curve have different recommended
// cadences (5s vs static) and a unified --watch-interval would lie
// to consumers.
func RunDefault(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	now := time.Now()
	if gated, ok := pkgdashboard.GateOlaresOne(ctx, c, cf, pkgdashboard.KindOverviewFan, now, os.Stderr); ok {
		liveGated := gated
		liveGated.Kind = pkgdashboard.KindOverviewFanLive
		curveGated := gated
		curveGated.Kind = pkgdashboard.KindOverviewFanCurve
		gated.Sections = map[string]pkgdashboard.Envelope{
			"live":  liveGated,
			"curve": curveGated,
		}
		if cf.Output == pkgdashboard.OutputJSON {
			return pkgdashboard.WriteJSON(os.Stdout, gated)
		}
		return nil
	}

	env, liveErr := BuildSectionsEnvelope(ctx, c, cf, now)
	if cf.Output == pkgdashboard.OutputJSON {
		return pkgdashboard.WriteJSON(os.Stdout, env)
	}
	return WriteSectionsTable(os.Stdout, env, liveErr)
}

// BuildSectionsEnvelope assembles { live, curve } in one fan-out.
// The live section's transport error is propagated separately so
// the table renderer can print "(error: ...)" instead of an empty
// table; JSON consumers see the error via Meta.Error on the live
// section. Curve never errors (it's pure data).
func BuildSectionsEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, now time.Time) (pkgdashboard.Envelope, error) {
	live, lerr := BuildLiveEnvelope(ctx, c, cf, now)
	if lerr != nil {
		live.Kind = pkgdashboard.KindOverviewFanLive
		live.Meta.Error = lerr.Error()
		live.Meta.ErrorKind = pkgdashboard.ClassifyTransportErr(lerr)
	}
	curve := BuildCurveEnvelope(cf, c.OlaresID(), now)
	env := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewFan,
		Meta: pkgdashboard.NewMeta(time.Now().In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Sections: map[string]pkgdashboard.Envelope{
			"live":  live,
			"curve": curve,
		},
	}
	return env, lerr
}

// WriteSectionsTable renders the human-readable sections layout:
//
//	== LIVE ==
//	<live row>
//
//	== CURVE ==
//	<curve table>
//
// liveErr (when non-nil) replaces the live table with an "(error:
// ...)" line so the curve still emits even if the live fetch
// failed. Mirrors the cmd-side pre-refactor behavior.
func WriteSectionsTable(w io.Writer, env pkgdashboard.Envelope, liveErr error) error {
	live, hasLive := env.Sections["live"]
	curve, hasCurve := env.Sections["curve"]

	fmt.Fprintln(w, "== LIVE ==")
	switch {
	case liveErr != nil:
		fmt.Fprintf(w, "(error: %s)\n", liveErr)
	case !hasLive:
		fmt.Fprintln(w, "(missing)")
	default:
		if err := WriteLiveTable(w, live); err != nil {
			return err
		}
	}
	fmt.Fprintln(w, "\n== CURVE ==")
	if !hasCurve {
		fmt.Fprintln(w, "(missing)")
		return nil
	}
	return WriteCurveTable(w, curve)
}
