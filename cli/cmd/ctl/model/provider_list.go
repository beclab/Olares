package model

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli model provider list`
//
// GET /console/api/providers. The list is the inventory question: what can
// this Olares route to, and is any of it disabled or broken.

func newProviderListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list every provider Router can route to",
		Long: `List the configured providers, newest first.

Columns: the provider's name and display title, its type, whether it is
active or disabled, whether it came from an admin ("manual") or from an
installed model application ("olares"), and its base URL.

A Market-sourced provider is listed only while its application is running.
An absent model app is therefore installing, stopped, or unreachable — check
it with "olares-cli model market tasks" or "olares-cli market status".

Pass --output json for every field, including the credential version that
"provider rollback" selects on.

Examples:
  olares-cli model provider list
  olares-cli model provider list -o json
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
		_, err := fmt.Fprintf(w, "no provider is listed. %s\nTo add a cloud account, use `olares-cli model provider create`.\n",
			hiddenProviderNote)
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tTITLE\tTYPE\tSTATUS\tSOURCE\tBASE URL"); err != nil {
		return err
	}
	for i := range rows {
		r := &rows[i]
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(r.Name),
			nonEmpty(r.title()),
			nonEmpty(r.ProviderType),
			nonEmpty(r.Status),
			nonEmpty(r.Source),
			nonEmpty(r.BaseURL),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
