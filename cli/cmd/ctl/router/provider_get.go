package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

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

// headlineSupports are the capability keys worth a column, in the order they
// change what a caller can send. Router tracks far more (audio, tokenizer
// details, per-parameter support); `--output json` carries all of them.
var headlineSupports = []struct {
	key   string
	label string
}{
	{"supports_vision", "vision"},
	{"supports_function_calling", "tools"},
	{"supports_reasoning", "reasoning"},
	{"supports_native_streaming", "streaming"},
	{"supports_audio_input", "audio-in"},
	{"supports_audio_output", "audio-out"},
	{"supports_web_search", "web-search"},
}

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
	mtw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(mtw, "NAME\tMODE\tENABLED\tSTATUS\tCONTEXT\tCAPABILITIES"); err != nil {
		return err
	}
	for i := range d.Models {
		m := &d.Models[i]
		if _, err := fmt.Fprintf(mtw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			nonEmpty(m.Name),
			nonEmpty(m.Mode),
			boolStr(m.Enabled),
			nonEmpty(m.Status),
			intOrDash(m.ContextSize),
			summarizeSupports(m.Supports),
		); err != nil {
			return err
		}
	}
	return mtw.Flush()
}

// summarizeSupports names the headline capabilities a model has, and says so
// when it has none of them rather than leaving the cell blank — a blank reads
// as missing data, which is a different thing.
func summarizeSupports(supports map[string]bool) string {
	if len(supports) == 0 {
		return "-"
	}
	labels := make([]string, 0, len(headlineSupports))
	for _, h := range headlineSupports {
		if supports[h.key] {
			labels = append(labels, h.label)
		}
	}
	if len(labels) == 0 {
		return "none of the headline set"
	}
	return strings.Join(labels, ",")
}

func intOrDash(v int) string {
	if v <= 0 {
		return "-"
	}
	return strconv.Itoa(v)
}
