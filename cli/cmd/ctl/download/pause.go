package download

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

const batchMaxTasks = 500

func NewPauseCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "pause <id> [id...]",
		Short: "pause one or more download tasks",
		Long: `Pause one or more download tasks.

Pass task ids separated by spaces (not commas or brackets). Up to 500 ids
per call.`,
		Example: `  # pause a single task
  olares-cli knowledge download pause 42

  # pause several tasks
  olares-cli knowledge download pause 42 43 44`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runLifecycle(c.Context(), f, "pause", args, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func NewResumeCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "resume <id> [id...]",
		Short: "resume one or more download tasks",
		Long: `Resume one or more download tasks.

Pass task ids separated by spaces (not commas or brackets). Up to 500 ids
per call.`,
		Example: `  # resume a single task
  olares-cli knowledge download resume 42

  # resume several tasks
  olares-cli knowledge download resume 42 43 44`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runLifecycle(c.Context(), f, "resume", args, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func NewCancelCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "cancel <id> [id...]",
		Short: "cancel one or more download tasks",
		Long: `Cancel one or more download tasks.

Pass task ids separated by spaces (not commas or brackets). Up to 500 ids
per call.`,
		Example: `  # cancel a single task
  olares-cli knowledge download cancel 42

  # cancel several tasks
  olares-cli knowledge download cancel 42 43 44`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runLifecycle(c.Context(), f, "cancel", args, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runLifecycle(ctx context.Context, f *cmdutil.Factory, verb string, idRaws []string, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	ids, err := parseTaskIDs(idRaws)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if len(ids) == 1 {
		path := fmt.Sprintf("/api/download/%s/%d", verb, ids[0])
		if err := doMutate(ctx, pc.doer, "PUT", path, nil, nil); err != nil {
			return err
		}
		switch format {
		case FormatJSON:
			return printJSON(os.Stdout, map[string]interface{}{"verb": verb, "task_id": ids[0]})
		default:
			fmt.Printf("%s task %d\n", verb, ids[0])
			return nil
		}
	}
	var res BatchResult
	if err := doMutate(ctx, pc.doer, "PUT", "/api/download/batch/"+verb, BatchReq{TaskIDs: ids}, &res); err != nil {
		return err
	}
	return renderBatchResult(os.Stdout, format, verb, res)
}

// parseTaskIDs validates and de-duplicates positional task ids. Empty input,
// non-positive / non-integer tokens, and lists longer than batchMaxTasks are
// rejected locally.
func parseTaskIDs(raws []string) ([]int64, error) {
	if len(raws) == 0 {
		return nil, fmt.Errorf("at least one task id is required")
	}
	if len(raws) > batchMaxTasks {
		return nil, fmt.Errorf("too many task ids (%d); max %d per request", len(raws), batchMaxTasks)
	}
	seen := make(map[int64]struct{}, len(raws))
	ids := make([]int64, 0, len(raws))
	for _, raw := range raws {
		id, err := parseTaskID(raw)
		if err != nil {
			return nil, err
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("at least one task id is required")
	}
	return ids, nil
}

func renderBatchResult(w io.Writer, format Format, verb string, res BatchResult) error {
	switch format {
	case FormatJSON:
		if err := printJSON(w, res); err != nil {
			return err
		}
	default:
		fmt.Fprintf(w, "%s: %d succeeded, %d failed\n", verb, len(res.Succeeded), len(res.Failed))
		for _, f := range res.Failed {
			fmt.Fprintf(w, "  %d: %s\n", f.TaskID, orDash(f.Error))
		}
	}
	// Both table and json print the payload first; any failed id still
	// exits non-zero so automation that trusts the exit code cannot miss
	// leftover tasks.
	if len(res.Failed) > 0 {
		return fmt.Errorf("%s: %d of %d tasks failed", verb, len(res.Failed), len(res.Succeeded)+len(res.Failed))
	}
	return nil
}
