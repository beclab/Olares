package rag

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newFailedCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use: "failed", Short: "List this user's current failed Files RAG jobs", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pc, err := prepare(cmd.Context(), f)
			if err != nil {
				return err
			}
			response, err := call(cmd.Context(), pc.doer, http.MethodGet, "/api/rag/jobs/failed", nil)
			if err != nil {
				return err
			}
			return printData(response.Data, output)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "table", "output format: table, json")
	return cmd
}

func newRetryCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use: "retry <job-id>", Short: "Retry one current failed Files RAG job", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || id < 1 {
				return fmt.Errorf("job-id must be a positive integer")
			}
			pc, err := prepare(cmd.Context(), f)
			if err != nil {
				return err
			}
			response, err := call(cmd.Context(), pc.doer, http.MethodPost, fmt.Sprintf("/api/rag/jobs/%d/retry", id), nil)
			if err != nil {
				return err
			}
			if message := envelopeMessage(response.Message); message != "" {
				fmt.Fprintln(os.Stdout, message)
			}
			return nil
		},
	}
}
