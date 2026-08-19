package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router provider models get <provider> <model>`
//
// One model, in full. The model table under `provider get` answers "what does
// this provider offer"; this answers "what exactly does Router believe about
// this one", which is the question that precedes correcting it.
//
// There is no single-model route to call: Router serves models inline on the
// provider detail, so this reads the same response the other model verbs
// resolve through.

func newProviderModelsGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <provider> <model>",
		Short: "everything Router records about one model",
		Long: `Show one model in full.

The model table under "provider get" is a summary: it names the headline
capabilities and leaves the rest out. This prints the whole record — every
capability flag that is on, every price, the context window, and the engine
launch flags when the model is served by a local application.

The model may be named by its upstream name, its alias, or its id.

Engine arguments only appear for a model application's own models. They are
the flags the Model Console starts the inference engine with, mirrored into
Router by "provider sync-models"; the card inside the application is the
original, and "router local spec <app>" reads it there.

Example:
  olares-cli router provider models get lmstudio qwen3-8b
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runProviderModelsGet(c.Context(), f, args[0], args[1], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderModelsGet(ctx context.Context, f *cmdutil.Factory, providerRef, modelRef, outputRaw string) error {
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
	provider, model, err := resolveProviderModel(ctx, pc, providerRef, modelRef)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, model)
	}
	return renderProviderModel(os.Stdout, provider, model)
}

func renderProviderModel(w io.Writer, p *providerRow, m *providerModelRow) error {
	t := newTable(w)
	t.row("MODEL", nonEmpty(m.Name))
	t.row("MODE", nonEmpty(m.Mode))
	t.row("ENABLED", boolStr(m.Enabled))
	t.row("STATUS", nonEmpty(m.Status))
	t.row("PROVIDER", nonEmpty(p.Name)+" ("+nonEmpty(p.Source)+")")
	if m.ContextSize > 0 {
		t.row("CONTEXT", fmt.Sprintf("%d tokens", m.ContextSize))
	}
	if m.MaxOutputTokens > 0 {
		t.row("MAX OUTPUT", fmt.Sprintf("%d tokens", m.MaxOutputTokens))
	}
	if args := strings.TrimSpace(m.EngineArgs); args != "" {
		t.row("ENGINE ARGS", args)
	}
	t.row("ID", nonEmpty(m.ID))
	if err := t.flush(); err != nil {
		return err
	}

	if on := enabledSupports(m.Supports); len(on) > 0 {
		if _, err := fmt.Fprintf(w, "\nSUPPORTS\n  %s\n", strings.Join(on, ", ")); err != nil {
			return err
		}
	} else if len(m.Supports) > 0 {
		if _, err := fmt.Fprintln(w, "\nSUPPORTS\n  none of the declared capabilities are enabled"); err != nil {
			return err
		}
	}

	if len(m.Pricing) > 0 {
		if _, err := fmt.Fprintln(w, "\nPRICING"); err != nil {
			return err
		}
		keys := make([]string, 0, len(m.Pricing))
		for k := range m.Pricing {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		pt := newTable(w)
		for _, k := range keys {
			pt.row("  "+k, m.Pricing[k])
		}
		if err := pt.flush(); err != nil {
			return err
		}
	}

	if len(m.ParameterRules) > 0 && string(m.ParameterRules) != "null" {
		_, err := fmt.Fprintln(w, "\nParameter rules are set; -o json prints them.")
		return err
	}
	return nil
}
