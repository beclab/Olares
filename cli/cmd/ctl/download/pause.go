package download

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewPauseCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "pause <id>",
		Short: "pause a download task",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runLifecycle(c.Context(), f, "pause", args[0])
		},
	}
}

func NewResumeCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "resume <id>",
		Short: "resume a download task",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runLifecycle(c.Context(), f, "resume", args[0])
		},
	}
}

func NewCancelCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <id>",
		Short: "cancel a download task",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runLifecycle(c.Context(), f, "cancel", args[0])
		},
	}
}

func runLifecycle(ctx context.Context, f *cmdutil.Factory, verb, idRaw string) error {
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
	path := fmt.Sprintf("/api/download/%s/%d", verb, id)
	if err := doMutate(ctx, pc.doer, "PUT", path, nil, nil); err != nil {
		return err
	}
	fmt.Printf("%s task %d\n", verb, id)
	return nil
}
