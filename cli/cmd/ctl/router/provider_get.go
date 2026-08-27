package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router provider get <name|id>`
//
// GET /console/api/providers/:id, which carries the provider's models inline.
//
// One behaviour is worth knowing before reading the output: for a model
// application Router re-probes the app's own /v1/models on every detail read,
// and reports an empty model list when that probe fails rather than the last
// values it cached. An empty list under a running app therefore means the app
// is not answering right now, not that it serves nothing.

func newProviderGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <name|id>",
		Short: "show one provider and the models it serves",
		Long: `Show a provider with its models inline.

The argument is the provider name as "provider list" shows it; an id works
too.

For a model application, Router probes the application's own model list while
serving this request, so what you see is current availability rather than a
cached copy. An empty model list on a running application means it is not
answering right now.

The models table summarises the capabilities that change what you can send —
vision, tools, reasoning, streaming, audio. Router tracks around thirty more,
including per-parameter support and pricing; pass --output json for those.

Examples:
  olares-cli router provider get openai-prod
  olares-cli router provider get "Qwen3.6-27B (llama.cpp)" -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProviderGet(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderGet(ctx context.Context, f *cmdutil.Factory, ref, outputRaw string) error {
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
	found, err := resolveProvider(ctx, pc, ref)
	if err != nil {
		return err
	}
	detail, err := getProvider(ctx, pc, found.ID)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, detail)
	}
	return renderProviderGet(os.Stdout, detail)
}

func renderProviderGet(w io.Writer, d *providerDetail) error {
	if err := renderProviderRow(w, &d.providerRow); err != nil {
		return err
	}

	if len(d.Models) == 0 {
		note := "\nno models"
		if d.isMarketSourced() {
			note += " — Router could not reach this application's model list just now"
		}
		_, err := fmt.Fprintln(w, note)
		return err
	}

	if _, err := fmt.Fprintf(w, "\nMODELS (%d)\n", len(d.Models)); err != nil {
		return err
	}
	// AT ONCE only when something declared it. A cloud model has no engine of
	// ours behind it, so the column would be a dash on every row of most
	// providers and read as a missing figure rather than an inapplicable one.
	wide := false
	for i := range d.Models {
		wide = wide || d.Models[i].MaxConcurrency > 0
	}
	headers := []string{"NAME", "MODE", "ENABLED", "STATUS", "CONTEXT"}
	if wide {
		headers = append(headers, "AT ONCE")
	}
	headers = append(headers, "CAPABILITIES")
	mt := newTable(w, headers...)
	for i := range d.Models {
		m := &d.Models[i]
		cells := []string{
			nonEmpty(m.Name),
			nonEmpty(m.Mode),
			boolStr(m.Enabled),
			nonEmpty(m.Status),
			intOrDash(m.ContextSize),
		}
		if wide {
			cells = append(cells, intOrDash(m.MaxConcurrency))
		}
		cells = append(cells, summarizeSupports(m.Supports))
		mt.row(cells...)
	}
	if err := mt.flush(); err != nil {
		return err
	}
	if wide {
		_, err := fmt.Fprintln(w, "\nAT ONCE is how many requests the engine was launched to work on at "+
			"the same time. A request beyond that waits its turn, which looks like a slow model rather "+
			"than a queue — ENGINE LOAD above is what tells the two apart.")
		return err
	}
	return nil
}

func intOrDash(v int) string {
	if v <= 0 {
		return "-"
	}
	return strconv.Itoa(v)
}
