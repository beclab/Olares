package advanced

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/containerdimages"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings advanced images ...`
//
// Backed by /api/containerd/images?registry=<name>. The body is a BFL
// envelope around an array of RegistryImage:
//
//   { "id": "sha256:...", "size": <int>, "repo_tags": ["..."] }
//
// The SPA passes `registry` as a query string, so we expose it as
// `--registry`. The default (no --registry) aggregates across all
// registries.
func NewImagesCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "containerd images (Settings -> Advanced)",
		Long: `List containerd images, optionally scoped to a single registry.

Subcommands:
  list   list images

Out of scope until a JWS key sourcing path exists:
  delete <id>, prune
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newImagesListCommand(f))
	return cmd
}

func newImagesListCommand(f *cmdutil.Factory) *cobra.Command {
	var output, registry string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list containerd images (optionally scoped by --registry)",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "list containerd images"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runImagesList(ctx, f, registry, output), "list containerd images")
		},
	}
	cmd.Flags().StringVar(&registry, "registry", "", "filter by registry name (matches the SPA's selector)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runImagesList(ctx context.Context, f *cmdutil.Factory, registry, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}

	rows, err := containerdimages.List(ctx, f, registry)
	if err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, rows)
	default:
		return renderImagesTable(os.Stdout, rows)
	}
}

func renderImagesTable(w io.Writer, rows []containerdimages.Image) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no images")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tSIZE\tREPO TAGS"); err != nil {
		return err
	}
	for _, r := range rows {
		tags := "-"
		if len(r.RepoTags) > 0 {
			tags = ""
			for i, t := range r.RepoTags {
				if i > 0 {
					tags += ","
				}
				tags += t
			}
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n",
			nonEmpty(containerdimages.ShortID(r.ID)),
			containerdimages.HumanBytes(r.Size),
			tags,
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
