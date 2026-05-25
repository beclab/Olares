package gpu

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunDefault is the cmd-side entry point for `dashboard overview
// gpu` without a subverb. Emits a sections envelope: graphics +
// tasks — same data the SPA renders behind the two tabs of the
// "GPU overview" page (both lists are loaded eagerly when the page
// mounts; the user just toggles between them via the tabs widget).
//
// Concurrency: the two endpoints (HAMI /v1/gpus and /v1/containers)
// are independent, so we fan them out with sync.WaitGroup. Per-
// section transport errors land on the section's Meta.Error rather
// than aborting the parent envelope — partial-failure semantics
// match the rest of the dashboard sections-envelope family.
//
// Soft GPU advisory: the parent envelope's Meta.Note carries the
// SPA-equivalent "would-have-been-hidden" hint (non-admin / no-CUDA
// node) on a single Client cache hit; child sections re-use the
// same note via BuildListEnvelope / BuildTasksEnvelope which each
// call GPUAdvisory once.
//
// One-shot only — graphics + tasks have the same operational nature
// (status snapshots, no monitor windows), but adding --watch on the
// parent would force a single combined cadence for two endpoints
// the SPA refreshes independently. Users wanting watch should pick
// `gpu graphics --watch` or `gpu tasks --watch`.
func RunDefault(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	now := time.Now()
	env := BuildSectionsEnvelope(ctx, c, cf, now)
	if cf.Output == pkgdashboard.OutputJSON {
		return pkgdashboard.WriteJSON(os.Stdout, env)
	}
	return WriteSectionsTable(os.Stdout, env)
}

// BuildSectionsEnvelope assembles { graphics, tasks } in one
// fan-out. A section that fails (transport, decode, gating) keeps
// Meta.Error on itself and does NOT abort sibling fetches — same
// shape as overview/{disk,fan}/default.go.
func BuildSectionsEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, now time.Time) pkgdashboard.Envelope {
	advisoryNote, _ := pkgdashboard.GPUAdvisory(ctx, c, cf, os.Stderr)

	var (
		graphicsEnv pkgdashboard.Envelope
		tasksEnv    pkgdashboard.Envelope
		wg          sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		graphicsEnv = BuildListEnvelope(ctx, c, cf, now)
	}()
	go func() {
		defer wg.Done()
		tasksEnv = BuildTasksEnvelope(ctx, c, cf, now)
	}()
	wg.Wait()

	parent := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewGPU,
		Meta: pkgdashboard.NewMeta(time.Now().In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Sections: map[string]pkgdashboard.Envelope{
			"graphics": graphicsEnv,
			"tasks":    tasksEnv,
		},
	}
	if advisoryNote != "" {
		parent.Meta.Note = advisoryNote
	}
	return parent
}

// WriteSectionsTable renders the human-readable sections layout:
//
//	== GRAPHICS ==
//	<gpu list table>
//
//	== TASKS ==
//	<task list table>
//
// Per-section errors print "(error: ...)" so the other section
// still renders — same pattern as overview/disk/fan default tables.
func WriteSectionsTable(w io.Writer, env pkgdashboard.Envelope) error {
	if env.Meta.Note != "" {
		fmt.Fprintf(os.Stderr, "(advisory) %s\n", env.Meta.Note)
	}
	graphics, hasGraphics := env.Sections["graphics"]
	tasks, hasTasks := env.Sections["tasks"]

	fmt.Fprintln(w, "== GRAPHICS ==")
	switch {
	case !hasGraphics:
		fmt.Fprintln(w, "(missing)")
	case graphics.Meta.Error != "":
		fmt.Fprintf(w, "(error: %s)\n", graphics.Meta.Error)
	case graphics.Meta.Empty:
		fmt.Fprintf(w, "(empty: %s)\n", graphics.Meta.EmptyReason)
	default:
		if err := WriteListTable(w, graphics); err != nil {
			return err
		}
	}

	fmt.Fprintln(w, "\n== TASKS ==")
	switch {
	case !hasTasks:
		fmt.Fprintln(w, "(missing)")
	case tasks.Meta.Error != "":
		fmt.Fprintf(w, "(error: %s)\n", tasks.Meta.Error)
	case tasks.Meta.Empty:
		fmt.Fprintf(w, "(empty: %s)\n", tasks.Meta.EmptyReason)
	default:
		if err := WriteTasksTable(w, tasks); err != nil {
			return err
		}
	}
	return nil
}
