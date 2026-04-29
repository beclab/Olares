package network

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings network hosts-file ...`
//
// Backed by user-service's /api/system/hosts-file, which proxies the
// olaresd daemon's /system/hosts-file. The wire shape (after
// returnSucceed wraps the daemon body) is a BFL envelope around an
// array of `{ip, host}`:
//
//   { "code": 0, "message": "success", "data": [{"ip":"...", "host":"..."}, ...] }
//
// The daemon side is JWS-signed: when the SPA hits this endpoint it
// also passes an X-Signature header; without one, user-service falls
// back to the access token (see olaresd/utils.ts). The CLI sends only
// X-Authorization, so this verb works as long as olaresd accepts
// authorization-header callers — if a future release tightens that,
// we'll have to surface a JWS-signing path here.
func NewHostsFileCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hosts-file",
		Short: "system hosts-file (Settings -> Network)",
		Long: `Inspect the system hosts-file used for custom DNS resolution
inside the Olares mesh.

Subcommands:
  get   show the current hosts-file entries

Out of scope until a JWS key sourcing path exists:
  set   replace the entire hosts-file (atomic write)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newHostsFileGetCommand(f))
	return cmd
}

func newHostsFileGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "show the current hosts-file entries",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runHostsFileGet(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

type hostEntry struct {
	IP   string `json:"ip"`
	Host string `json:"host"`
}

func runHostsFileGet(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var rows []hostEntry
	if err := doGetEnvelope(ctx, pc.doer, "/api/system/hosts-file", &rows); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, rows)
	default:
		return renderHostsTable(os.Stdout, rows)
	}
}

func renderHostsTable(w io.Writer, rows []hostEntry) error {
	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, "no hosts-file entries")
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "IP\tHOST"); err != nil {
		return err
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", nonEmpty(r.IP), nonEmpty(r.Host)); err != nil {
			return err
		}
	}
	return tw.Flush()
}
