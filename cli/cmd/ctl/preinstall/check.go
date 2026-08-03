package preinstall

import (
	"fmt"

	pkgpreinstall "github.com/beclab/Olares/cli/pkg/preinstall"
	"github.com/spf13/cobra"
)

func NewCmdPreinstallCheck() *cobra.Command {
	var full bool
	cmd := &cobra.Command{
		Use:   "check <path>",
		Short: "Validate a static Market preinstall bundle directory",
		Long: `Validate that <path> is a well-formed static Market preinstall
bundle directory (the directory that contains bundle.json, typically
preinstall/market).

By default the command runs contract-level checks that mirror the
installer prepare path:

  - decode and semantically validate bundle.json
  - build a default install profile and validate it against the bundle
  - verify each chart exists, is a regular file within size limits,
    and matches chartSha256
  - load and validate each artifact manifest

Pass --full to also verify artifact payload trees under each artifact's
source directory against the manifest (entry types, digests, and no
undeclared paths). This can be slow for large model payloads.

Examples:
  olares-cli preinstall check ./preinstall/market
  olares-cli preinstall check ./preinstall/market --full`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := pkgpreinstall.CheckStaticBundle(args[0], pkgpreinstall.CheckOptions{Full: full}); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "ok")
			return nil
		},
	}
	cmd.Flags().BoolVar(&full, "full", false, "also verify artifact payload trees against manifests")
	return cmd
}
