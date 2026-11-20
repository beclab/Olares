package rag

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newDirsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{Use: "dirs", Short: "List or change Files RAG root directories"}
	cmd.AddCommand(newDirsListCommand(f), newDirsMutateCommand(f, false), newDirsMutateCommand(f, true))
	return cmd
}

func newDirsListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use: "list", Short: "List this user's Files RAG roots", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pc, err := prepare(cmd.Context(), f)
			if err != nil {
				return err
			}
			response, err := call(cmd.Context(), pc.doer, http.MethodGet, "/api/rag/directories", nil)
			if err != nil {
				return err
			}
			return printData(response.Data, output)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "table", "output format: table, json")
	return cmd
}

func newDirsMutateCommand(f *cmdutil.Factory, remove bool) *cobra.Command {
	verb, method := "add", http.MethodPost
	if remove {
		verb, method = "rm", http.MethodDelete
	}
	var yes bool
	cmd := &cobra.Command{
		Use: verb + " <path>...", Short: verb + " Files RAG root directories", Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			roots := make([]string, 0, len(args))
			seen := map[string]bool{}
			for _, arg := range args {
				root, err := validateRoot(arg)
				if err != nil {
					return err
				}
				if !seen[root] {
					roots = append(roots, root)
					seen[root] = true
				}
			}
			if remove {
				if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin, "Remove these directories from Files RAG indexing (original files are kept)?", yes); err != nil {
					return err
				}
			}
			pc, err := prepare(cmd.Context(), f)
			if err != nil {
				return err
			}
			response, err := call(cmd.Context(), pc.doer, method, "/api/rag/directories", map[string]interface{}{"roots": roots})
			if err != nil {
				return err
			}
			if message := envelopeMessage(response.Message); message != "" {
				fmt.Fprintln(os.Stdout, message)
			}
			return nil
		},
	}
	if remove {
		cmd.Flags().BoolVarP(&yes, "yes", "y", false, "confirm directory removal without prompting")
	}
	return cmd
}
