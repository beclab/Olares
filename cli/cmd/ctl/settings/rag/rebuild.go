package rag

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newRebuildCommand(f *cmdutil.Factory) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use: "rebuild", Short: "Rebuild the global Files RAG index (Owner only)", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin, "Rebuild the global Files RAG index? Current derived data will be cleared and rebuilt; original files are kept.", yes); err != nil {
				return err
			}
			pc, err := prepare(cmd.Context(), f)
			if err != nil {
				return err
			}
			response, err := call(cmd.Context(), pc.doer, http.MethodPost, "/api/rag/rebuild", nil)
			if err != nil {
				return err
			}
			if message := envelopeMessage(response.Message); message != "" {
				fmt.Fprintln(os.Stdout, message)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "confirm the global rebuild without prompting")
	return cmd
}
