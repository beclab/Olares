package rag

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newTaskCommand(f *cmdutil.Factory, paused bool) *cobra.Command {
	name, description := "pause", "Pause this user's Files RAG task"
	if !paused {
		name, description = "resume", "Resume this user's Files RAG task"
	}
	return &cobra.Command{
		Use: name, Short: description, Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pc, err := prepare(cmd.Context(), f)
			if err != nil {
				return err
			}
			response, err := call(cmd.Context(), pc.doer, http.MethodPut, "/api/rag/task", map[string]bool{"paused": paused})
			if err != nil {
				return err
			}
			if message := envelopeMessage(response.Message); message != "" {
				fmt.Fprintln(os.Stdout, message)
			} else {
				fmt.Fprintf(os.Stdout, "Files RAG task %sd.\n", name)
			}
			return nil
		},
	}
}
