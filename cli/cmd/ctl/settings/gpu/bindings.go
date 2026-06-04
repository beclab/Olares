package gpu

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/edge"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/olaresclient"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings gpu bindings <app>` (Olares 1.12.6+)
//
// Lists the compute (GPU) device bindings held by an app, via
// olaresclient.ComputeOps.GetAppBindings (GET
// /api/apps/<app>/compute-resources/bindings).
func NewBindingsCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "bindings <app>",
		Short: "show an app's compute (GPU) bindings (Olares 1.12.6+)",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "list GPU bindings"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runBindings(ctx, f, args[0], output), "list GPU bindings")
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runBindings(ctx context.Context, f *cmdutil.Factory, appName, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	doer, _, err := edge.New(ctx, f)
	if err != nil {
		return err
	}

	return f.WithOlaresClient(ctx, func(c olaresclient.OlaresClient) error {
		raw, err := c.GetAppBindings(ctx, doer, appName)
		if err != nil {
			return err
		}
		var bindings []computeBinding
		if err := decodeData(raw, &bindings); err != nil {
			return err
		}
		if format == FormatJSON {
			return printJSON(os.Stdout, bindings)
		}
		return renderBindings(os.Stdout, appName, bindings)
	})
}

func renderBindings(w io.Writer, appName string, bindings []computeBinding) error {
	if len(bindings) == 0 {
		_, err := fmt.Fprintf(w, "%s has no compute bindings\n", appName)
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NODE\tDEVICE\tMODE\tMEMORY"); err != nil {
		return err
	}
	for _, b := range bindings {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			nonEmpty(b.NodeName), nonEmpty(b.DeviceID), nonEmpty(b.Mode), formatMem(b.Memory),
		); err != nil {
			return err
		}
	}
	return tw.Flush()
}
