package search

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings search rebuild`
//
// Backed by POST /api/search/task/rebuild (no body). user-service
// `@All`-proxies to search3, which schedules a full reindex of every
// monitored directory.
//
// SPA reference: apps/packages/app/src/api/settings/search.ts:
//   rebuildSearchTask() -> axios.post('/api/search/task/rebuild')
//
// The SPA shows a loading spinner and then re-polls /task/stats/merged
// to report progress; the CLI verb returns as soon as the request is
// accepted — use `settings search status` afterwards to track the task.
//
// Role: Search is in the normal-user menu (admin.ts:107). No
// PreflightRole check.
func NewRebuildCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rebuild",
		Short: "trigger a full reindex of the search index",
		Long: `Trigger a full reindex on search3 (the same action as the SPA's
"Rebuild index" button under Settings -> Search > File Search).

The reindex runs asynchronously; this command returns once search3 has
accepted the request. Use "olares-cli settings search status" afterwards
to watch progress.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runRebuild(c.Context(), f)
		},
	}
	return cmd
}

func runRebuild(ctx context.Context, f *cmdutil.Factory) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/api/search/task/rebuild", nil, nil); err != nil {
		return err
	}
	fmt.Println("Search index rebuild scheduled. Use `olares-cli settings search status` to watch progress.")
	return nil
}
