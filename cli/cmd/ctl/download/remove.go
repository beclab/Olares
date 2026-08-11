package download

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewRemoveCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		removeFile bool
		output     string
	)
	cmd := &cobra.Command{
		Use:   "remove <id> [id...]",
		Short: "remove one or more download tasks",
		Long: `Remove one or more download tasks.

Pass task ids separated by spaces (not commas or brackets). Up to 500 ids
per call.

By default only the task record is removed and the downloaded file is
kept. Add --remove-file to also delete the file on disk (applies to every
id in the call).`,
		Example: `  # remove a single task, keep the file
  olares-cli knowledge download remove 42

  # remove several tasks and delete their files
  olares-cli knowledge download remove 42 43 44 --remove-file`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runRemove(c.Context(), f, args, removeFile, output)
		},
	}
	cmd.Flags().BoolVar(&removeFile, "remove-file", false, "also delete the downloaded file (remove_flag)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runRemove(ctx context.Context, f *cmdutil.Factory, idRaws []string, removeFile bool, outputRaw string) error {
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
		req := RemoveReq{TaskID: ids[0], RemoveFlag: removeFile}
		if err := doMutate(ctx, pc.doer, "DELETE", "/api/download/remove", req, nil); err != nil {
			return err
		}
		switch format {
		case FormatJSON:
			return printJSON(os.Stdout, map[string]interface{}{
				"verb":        "remove",
				"task_id":     ids[0],
				"remove_file": removeFile,
			})
		default:
			if removeFile {
				fmt.Printf("removed task %d (and file)\n", ids[0])
			} else {
				fmt.Printf("removed task %d\n", ids[0])
			}
			return nil
		}
	}
	var res BatchResult
	if err := doMutate(ctx, pc.doer, "DELETE", "/api/download/batch/remove", BatchReq{
		TaskIDs:    ids,
		RemoveFlag: removeFile,
	}, &res); err != nil {
		return err
	}
	return renderBatchResult(os.Stdout, format, "remove", res)
}
