package rag

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewRAGCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rag",
		Short: "Manage Files RAG directories, tasks, failures, and rebuilds",
		Long: `Manage the same Files RAG control plane exposed by Settings.

All requests go through user-service using the active Olares profile. The CLI
cannot select another owner. Directory and task operations are user-scoped;
the service enforces Owner-only access to a global rebuild.`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newStatusCommand(f))
	cmd.AddCommand(newDirsCommand(f))
	cmd.AddCommand(newTaskCommand(f, true))
	cmd.AddCommand(newTaskCommand(f, false))
	cmd.AddCommand(newFailedCommand(f))
	cmd.AddCommand(newRetryCommand(f))
	cmd.AddCommand(newRebuildCommand(f))
	return cmd
}
