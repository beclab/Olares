package network

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings network frp ...`
//
// Backed by user-service's /api/frp-servers, which proxies the
// upstream Olares-tunnel registry (FRP_LIST_URL/servers). The wire
// shape is a BFL envelope wrapping an array of `{name, host}`:
//
//   { "code": 0, "message": "success", "data": [{"name":"...", "host":"..."}, ...] }
//
// Only `list` is in scope here; the v2 endpoint (POST /api/frp-servers-v2)
// is locale/userName-aware and the SPA only uses it to drive a select
// dropdown — there's no stable scripting use case for it yet.
func NewFRPCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frp",
		Short: "FRP / Olares-tunnel servers (Settings -> Network)",
		Long: `Inspect the upstream FRP server registry that Olares uses for the
Olares-tunnel reverse-proxy mode.

Subcommands:
  list   list the available FRP servers
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newFRPListCommand(f))
	return cmd
}

func newFRPListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list available FRP servers",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "list FRP servers"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runFRPList(ctx, f, output), "list FRP servers")
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

type frpServer struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

func runFRPList(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var rows []frpServer
	if err := doGetEnvelope(ctx, pc.doer, "/api/frp-servers", &rows); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, rows)
	default:
		return renderFRPTable(os.Stdout, rows)
	}
}

func renderFRPTable(w io.Writer, rows []frpServer) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no FRP servers")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tHOST"); err != nil {
		return err
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", nonEmpty(r.Name), nonEmpty(r.Host)); err != nil {
			return err
		}
	}
	return tw.Flush()
}
