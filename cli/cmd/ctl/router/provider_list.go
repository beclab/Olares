package router

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router provider list`
//
// GET /console/api/providers. The list is the inventory question: what can
// this Olares route to, and is any of it disabled or broken.

func newProviderListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list every provider Router can route to",
		Long: `List the configured providers, newest first.

Columns: the provider's routing name, its display title, the application behind
it, its type, whether it is active or disabled, whether it came from an admin
("manual") or from an installed model application ("olares"), and its base URL.

NAME is what a caller writes in front of a model name, and it does not identify
a row: every locally installed model application answers to "Olares", so that a
caller can write "Olares/<model>" without knowing which application serves it.
APP is what identifies those rows, and what the verbs here accept.

A Market-sourced provider is listed while its application is installing, running
or tearing down, and only then. An absent one is stopped or gone — check it with
"olares-cli router app catalog" or "olares-cli market status".

Pass --output json for every field, including the credential version that
"provider rollback" selects on.

Examples:
  olares-cli router provider list
  olares-cli router provider list -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runProviderList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	rows, err := listProviders(ctx, pc)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		if rows == nil {
			rows = []providerRow{}
		}
		return printJSON(os.Stdout, rows)
	}
	return renderProviderList(os.Stdout, rows)
}

func renderProviderList(w io.Writer, rows []providerRow) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintf(w, "no provider is listed. %s\nTo add a cloud account, use `olares-cli router provider create`.\n",
			hiddenProviderNote)
		return err
	}
	t := newTable(w, "NAME", "TITLE", "APP", "TYPE", "STATUS", "SOURCE", "BASE URL")
	for i := range rows {
		r := &rows[i]
		app := nonEmpty(strDeref(r.OlaresAppName))
		if s := strDeref(r.OlaresStatus); s != "" {
			app += " (" + s + ")"
		}
		t.row(
			nonEmpty(r.Name),
			clip(r.title(), 28),
			app,
			nonEmpty(r.ProviderType),
			nonEmpty(r.Status),
			nonEmpty(r.Source),
			clip(r.BaseURL, 44),
		)
	}
	return t.flush()
}
