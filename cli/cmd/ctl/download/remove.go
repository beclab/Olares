package download

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewRemoveCommand(f *cmdutil.Factory) *cobra.Command {
	var removeFile bool
	cmd := &cobra.Command{
		Use:   "remove <id>",
		Short: "remove a download task",
		Long:  `Remove a download task (DELETE /api/download/remove). Pass --remove-file to also delete the downloaded artefact.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runRemove(c.Context(), f, args[0], removeFile)
		},
	}
	cmd.Flags().BoolVar(&removeFile, "remove-file", false, "also delete the downloaded file (remove_flag)")
	return cmd
}

func runRemove(ctx context.Context, f *cmdutil.Factory, idRaw string, removeFile bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	id, err := parseTaskID(idRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	req := RemoveReq{TaskID: id, RemoveFlag: removeFile}
	if err := doMutate(ctx, pc.doer, "DELETE", "/api/download/remove", req, nil); err != nil {
		return err
	}
	if removeFile {
		fmt.Printf("removed task %d (and file)\n", id)
	} else {
		fmt.Printf("removed task %d\n", id)
	}
	return nil
}
