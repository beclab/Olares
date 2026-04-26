package advanced

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings advanced registries ...`
//
// Backed by /api/containerd/registries (terminusd proxy). The body is
// a BFL envelope around an array of RegistryMirror:
//
//   { "name": "...", "image_count": <int>, "image_size": <int>,
//     "endpoints": ["..."] | null }
//
// Phase 5 will add `mirrors get/set/delete` and `prune`.
func NewRegistriesCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registries",
		Short: "containerd registry mirrors (Settings -> Advanced)",
		Long: `Inspect containerd registry mirrors / image cache distribution.

Subcommands:
  list                                                    (Phase 1)

Subcommands landing in Phase 5 (JWS-signed):
  mirrors get <registry>, mirrors set <registry> <endpoint>...,
  mirrors delete <registry>, prune
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newRegistriesListCommand(f))
	return cmd
}

func newRegistriesListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list configured containerd registries with image stats",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runRegistriesList(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

type registryMirror struct {
	Name       string   `json:"name"`
	ImageCount int      `json:"image_count"`
	ImageSize  int64    `json:"image_size"`
	Endpoints  []string `json:"endpoints"`
}

func runRegistriesList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var rows []registryMirror
	if err := doGetEnvelope(ctx, pc.doer, "/api/containerd/registries", &rows); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, rows)
	default:
		return renderRegistriesTable(os.Stdout, rows)
	}
}

func renderRegistriesTable(w io.Writer, rows []registryMirror) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no containerd registries")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tIMAGES\tSIZE\tENDPOINTS"); err != nil {
		return err
	}
	for _, r := range rows {
		eps := "-"
		if len(r.Endpoints) > 0 {
			eps = ""
			for i, e := range r.Endpoints {
				if i > 0 {
					eps += ","
				}
				eps += e
			}
		}
		if _, err := fmt.Fprintf(tw, "%s\t%d\t%s\t%s\n",
			nonEmpty(r.Name),
			r.ImageCount,
			humanBytes(r.ImageSize),
			eps,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}

var _ = boolStr // keep helper available for future verbs in this area
