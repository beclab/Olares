package rag

import (
	"net/http"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Files RAG roots, progress, policy, and worker state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pc, err := prepare(cmd.Context(), f)
			if err != nil {
				return err
			}
			response, err := call(cmd.Context(), pc.doer, http.MethodGet, "/api/rag/control", nil)
			if err != nil {
				return err
			}
			return printData(response.Data, output)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "table", "output format: table, json")
	return cmd
}
