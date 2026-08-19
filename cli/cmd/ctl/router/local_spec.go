package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `router local spec` — GET/PUT /api/model-spec, GET /api/model-spec/file
// `router local config` — GET /api/config
// `router local endpoints` — GET /api/endpoints
//
// The model card is the same document Router stores as a model's spec, served
// by the application that runs the model. That makes it the answer to a
// specific confusion: when Router advertises capabilities a model does not
// have, these two copies have diverged, and this is the side that the
// application itself believes.

type localSpec struct {
	Name            string            `json:"name,omitempty"`
	Mode            string            `json:"mode,omitempty"`
	Label           map[string]string `json:"label,omitempty"`
	Supports        map[string]bool   `json:"supports,omitempty"`
	Pricing         map[string]string `json:"pricing,omitempty"`
	ParameterRules  json.RawMessage   `json:"parameter_rules,omitempty"`
	ContextSize     int               `json:"context_size,omitempty"`
	MaxOutputTokens int               `json:"max_output_tokens,omitempty"`
	EngineArgs      string            `json:"engine_args,omitempty"`
	Extensions      map[string]any    `json:"extensions,omitempty"`
}

func newLocalSpecCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "the model card this application serves to Router",
		Long: `Read or replace the model card — the capability document Router reads from
this application and stores as the model's spec.

Capabilities, prices, context window and parameter rules all come from here.
When Router offers a capability a model turns out not to have, this is the copy
the application itself stands behind, and "router provider sync-models" is what
brings the two back together.

  spec show <app>   the card as served, after parsing
  spec file <app>   the bytes on disk, before parsing
  spec set <app>    replace the card wholesale
`,
	}
	cmd.AddCommand(newLocalSpecShowCommand(f))
	cmd.AddCommand(newLocalSpecFileCommand(f))
	cmd.AddCommand(newLocalSpecSetCommand(f))
	return cmd
}

func newLocalSpecShowCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:     "show <app>",
		Aliases: []string{"get"},
		Short:   "the card as this application serves it",
		Long: `Show the model card this application serves, after parsing.

JSON output is the card itself rather than a summary of it, including any field
this CLI does not know about, which is what makes it the thing to edit and hand
back to "spec set".

Examples:
  olares-cli router local spec show llamacppqwen3627bggufv3
  olares-cli router local spec show llamacppqwen3627bggufv3 -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runLocalSpec(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runLocalSpec(ctx context.Context, f *cmdutil.Factory, ref, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	li, err := openLocal(ctx, f, ref)
	if err != nil {
		return err
	}
	// Decoded twice on purpose. The table needs the fields this tree knows
	// about; JSON output has to be the card itself, because it is what a person
	// edits and sends back with `spec set`, and a re-marshalled copy would drop
	// whatever this build of the CLI has not been taught about.
	var raw json.RawMessage
	if err := li.client.doJSON(ctx, "GET", epLocalModelSpec, nil, &raw); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, raw)
	}
	var spec localSpec
	if err := json.Unmarshal(raw, &spec); err != nil {
		return fmt.Errorf("this application's model card is not a shape this version understands: %w", err)
	}
	return renderLocalSpec(os.Stdout, li, &spec)
}

func renderLocalSpec(w io.Writer, li *llmInit, s *localSpec) error {
	t := newTable(w)
	t.row("APPLICATION", li.AppName)
	t.row("MODEL", nonEmpty(s.Name))
	t.row("MODE", nonEmpty(s.Mode))
	if s.ContextSize > 0 {
		t.row("CONTEXT", fmt.Sprintf("%d tokens", s.ContextSize))
	}
	if s.MaxOutputTokens > 0 {
		t.row("MAX OUTPUT", fmt.Sprintf("%d tokens", s.MaxOutputTokens))
	}
	if len(s.Label) > 0 {
		t.row("LABEL", labelOf(s.Label))
	}
	if args := strings.TrimSpace(s.EngineArgs); args != "" {
		t.row("ENGINE ARGS", args)
	}
	if err := t.flush(); err != nil {
		return err
	}

	if on := enabledSupports(s.Supports); len(on) > 0 {
		if _, err := fmt.Fprintf(w, "\nSUPPORTS\n  %s\n", strings.Join(on, ", ")); err != nil {
			return err
		}
	} else if len(s.Supports) > 0 {
		if _, err := fmt.Fprintln(w, "\nSUPPORTS\n  none of the declared capabilities are enabled"); err != nil {
			return err
		}
	}
	if len(s.Pricing) > 0 {
		if _, err := fmt.Fprintln(w, "\nPRICING"); err != nil {
			return err
		}
		keys := make([]string, 0, len(s.Pricing))
		for k := range s.Pricing {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		pt := newTable(w)
		for _, k := range keys {
			pt.row("  "+k, s.Pricing[k])
		}
		if err := pt.flush(); err != nil {
			return err
		}
	}
	if len(s.ParameterRules) > 0 && string(s.ParameterRules) != "null" {
		if _, err := fmt.Fprintf(w, "\nParameter rules are set; -o json prints them.\n"); err != nil {
			return err
		}
	}
	if unknown := unknownExtensions(s.Extensions); len(unknown) > 0 {
		_, err := fmt.Fprintf(w, "\nThe card carries keys this Model Console did not recognise (%s). "+
			"They were kept rather than dropped, and they do nothing.\n", strings.Join(unknown, ", "))
		return err
	}
	return nil
}

func labelOf(label map[string]string) string {
	for _, pref := range []string{"en_US", "en", "zh_Hans", "zh_CN"} {
		if v := strings.TrimSpace(label[pref]); v != "" {
			return v
		}
	}
	keys := make([]string, 0, len(label))
	for k := range label {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return "-"
	}
	return label[keys[0]]
}

// enabledSupports lists only what is on. A card declares every capability key
// it knows about, most of them false, and printing forty booleans to say three
// are true buries the answer.
func enabledSupports(supports map[string]bool) []string {
	on := make([]string, 0, len(supports))
	for k, v := range supports {
		if v {
			on = append(on, strings.TrimPrefix(k, "supports_"))
		}
	}
	sort.Strings(on)
	return on
}

// unknownExtensions surfaces the soft-fail record. The Model Console tolerates
// keys it does not know and files them under extensions, which is the right
// behaviour and silent — a card written for a newer gateway looks like it
// applied.
func unknownExtensions(ext map[string]any) []string {
	var out []string
	for k := range ext {
		if strings.HasPrefix(k, "_unknown") {
			if nested, ok := ext[k].(map[string]any); ok {
				for nk := range nested {
					out = append(out, nk)
				}
				continue
			}
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func newLocalSpecFileCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "file <app>",
		Short: "the model card file on disk, verbatim",
		Long: `Print the model card exactly as it is stored on disk.

"router local spec show" serves the parsed card held in memory; this is the file a
previous write actually left behind. They differ in whitespace, in key order,
and in the record of keys the Model Console did not recognise — so this is the
copy to compare against something you wrote, and that is the whole reason it
exists separately.

Examples:
  olares-cli router local spec file llamacppqwen3627bggufv3
  olares-cli router local spec file llamacppqwen3627bggufv3 > card.json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			li, err := openLocal(ctx, f, args[0])
			if err != nil {
				return err
			}
			resp, err := li.client.do(ctx, "GET", epLocalModelSpecFile, nil, "")
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			raw, rerr := io.ReadAll(resp.Body)
			if rerr != nil {
				return fmt.Errorf("read the card: %w", rerr)
			}
			if resp.StatusCode/100 != 2 {
				return specFileErr(ctx, li,
					li.client.formatErr("GET", epLocalModelSpecFile, resp.StatusCode, raw))
			}
			_, err = os.Stdout.Write(raw)
			return err
		},
	}
}

// specFileErr separates the two things a 404 can mean here, both of which are
// normal: this Model Console does not serve the route at all, or it does and no
// card has been written yet — the file only appears once something writes it,
// and until then the served card is the one seeded at boot.
func specFileErr(ctx context.Context, li *llmInit, err error) error {
	var re *RouterError
	if err == nil || !errors.As(err, &re) || re.Status != 404 {
		return err
	}
	if served := consoleServes(ctx, li, "GET", epLocalModelSpecFile); served != nil && !*served {
		return fmt.Errorf("this application's Model Console (%s) does not serve the card's file, only "+
			"the card it parsed from it. `olares-cli router local spec show %s -o json` is that card, and it "+
			"is what `router local spec set` takes back", consoleVersion(li), li.AppName)
	}
	return fmt.Errorf("%w\nNothing has written a card to disk, so this application is serving the one "+
		"built from its configuration at boot. `olares-cli router local spec show %s` shows it", err, li.AppName)
}

// consoleVersion names the build in a sentence about what it can do. The open
// step read it already, so this costs nothing.
func consoleVersion(li *llmInit) string {
	if li.Console != nil && strings.TrimSpace(li.Console.Version) != "" {
		return li.Console.Version
	}
	return "version unknown"
}

func newLocalSpecSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		from   string
		output string
		yes    bool
	)
	cmd := &cobra.Command{
		Use:   "set <app>",
		Short: "replace the model card wholesale",
		Long: `Replace a model application's card with the document you supply.

This is a replacement, not a merge: what you send becomes the card, and any
field you leave out is gone. Start from the current one —

  olares-cli router local spec show <app> -o json > card.json

— edit that, and send it back. The write is atomic and survives restarts,
because disk wins over the boot-time seed from then on. Until the first write
there is usually no file at all, which is why the starting point above is the
served card rather than "spec file".

Two changes do not take full effect until the engine restarts: the mode, and
capabilities the engine's own behaviour depends on. Router picks the new card up
on its next sync either way, which is what makes this worth doing at all — a
card that undersells a model is why a capability appears unavailable.

Examples:
  olares-cli router local spec set llamacppqwen3627bggufv3 --from card.json
  cat card.json | olares-cli router local spec set llamacppqwen3627bggufv3
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runLocalSpecSet(c.Context(), f, args[0], from, yes, output)
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "read the card from this file; standard input when omitted")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	addOutputFlag(cmd, &output)
	return cmd
}

// readSpecSource takes the card from a file, or from standard input when no
// file is named. Editing the card in a pipeline is the ordinary way to do this —
// read it, change one field, send it back — so an absent --from is a pipe rather
// than a missing argument.
func readSpecSource(from string) ([]byte, error) {
	if from == "" || from == "-" {
		var raw []byte
		if !isTerminal(os.Stdin) {
			var err error
			raw, err = io.ReadAll(os.Stdin)
			if err != nil {
				return nil, fmt.Errorf("read the card from standard input: %w", err)
			}
		}
		if len(bytes.TrimSpace(raw)) == 0 {
			return nil, fmt.Errorf("no card to send: pass --from <file>, or pipe one in. " +
				"`olares-cli router local spec show <app> -o json` prints the current card to start from")
		}
		return raw, nil
	}
	return os.ReadFile(from)
}

func runLocalSpecSet(ctx context.Context, f *cmdutil.Factory, ref, from string, yes bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	raw, err := readSpecSource(from)
	if err != nil {
		return err
	}
	var candidate map[string]any
	if err := json.Unmarshal(raw, &candidate); err != nil {
		return fmt.Errorf("the card is not valid JSON: %w", err)
	}
	if len(candidate) == 0 {
		return fmt.Errorf("the card is empty; a replacement has to carry the whole document, and " +
			"`olares-cli router local spec show <app> -o json` prints the current one to start from")
	}

	li, err := openLocal(ctx, f, ref)
	if err != nil {
		return err
	}
	if !yes {
		name, _ := candidate["name"].(string)
		mode, _ := candidate["mode"].(string)
		if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin, fmt.Sprintf(
			"Replace the model card on %s with the one supplied (model %q, mode %q)? "+
				"Any field absent from it will be gone.",
			li.AppName, nonEmpty(name), nonEmpty(mode)), false); err != nil {
			return err
		}
	}

	var updated localSpec
	if err := li.client.doJSON(ctx, "PUT", epLocalModelSpec, json.RawMessage(raw), &updated); err != nil {
		return specSetErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, updated)
	}
	if err := renderLocalSpec(os.Stdout, li, &updated); err != nil {
		return err
	}
	// Whether Router hears about this depends on whether it has been told the
	// application exists at all. Naming a sync command for a provider that is
	// not there would send someone after an error.
	if p := strings.TrimSpace(li.Provider); p != "" {
		_, err = fmt.Fprintf(os.Stdout, "\nWritten. Router reads this on its next sync; "+
			"`olares-cli router provider sync-models %q` does it now.\n", p)
		return err
	}
	_, err = fmt.Fprintf(os.Stdout, "\nWritten. Router does not have this application as a provider, so "+
		"nothing routes to it yet — `olares-cli router provider register %s` is what adds it.\n", li.AppName)
	return err
}

func specSetErr(err error) error {
	var re *RouterError
	if err == nil || !errors.As(err, &re) {
		return err
	}
	switch re.Status {
	case 400:
		return fmt.Errorf("%w\nThe card was rejected whole; nothing changed. Unknown fields are refused "+
			"here even though unknown capability keys are tolerated, so a typed field name is the "+
			"usual cause", err)
	case 413:
		return fmt.Errorf("%w\nThe card is larger than the Model Console accepts", err)
	case 500:
		return fmt.Errorf("%w\nThe card was valid but could not be written, so the application is still "+
			"serving the old one", err)
	}
	return err
}

func newLocalConfigCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "config <app>",
		Short: "the effective configuration, secrets redacted",
		Long: `Show the configuration a model application is actually running with.

Secrets are masked by the Model Console before they leave it — a token becomes a
length, a source URL keeps only its scheme and host — so this is safe to paste
into a bug report.

It is the answer to "why is it downloading from there" and "which engine flags
are in force", both of which are set at install time and easy to misremember.

Examples:
  olares-cli router local config llamacppqwen3627bggufv3
  olares-cli router local config llamacppqwen3627bggufv3 -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			li, err := openLocal(ctx, f, args[0])
			if err != nil {
				return err
			}
			var cfg map[string]any
			if err := li.client.doJSON(ctx, "GET", epLocalConfig, nil, &cfg); err != nil {
				return err
			}
			if format == FormatJSON {
				return printJSON(os.Stdout, cfg)
			}
			return renderFlatMap(os.Stdout, cfg)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// renderFlatMap prints an unknown-shaped document as sorted key/value lines,
// descending into nested objects with dotted keys. The configuration's fields
// are the Model Console's own and will grow; enumerating them here would mean
// silently dropping whatever it learns to report next.
func renderFlatMap(w io.Writer, doc map[string]any) error {
	t := newTable(w)
	writeFlat(t, "", doc)
	return t.flush()
}

func writeFlat(t *table, prefix string, doc map[string]any) {
	keys := make([]string, 0, len(doc))
	for k := range doc {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		name := k
		if prefix != "" {
			name = prefix + "." + k
		}
		switch v := doc[k].(type) {
		case map[string]any:
			if len(v) == 0 {
				continue
			}
			writeFlat(t, name, v)
		case nil:
			continue
		default:
			t.row(name, fmt.Sprintf("%v", v))
		}
	}
}

type localEndpoint struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Category    string   `json:"category"`
	Group       string   `json:"group,omitempty"`
	Description string   `json:"description"`
	Available   bool     `json:"available"`
	Reasons     []string `json:"reasons,omitempty"`
}

func newLocalEndpointsCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output      string
		unavailable bool
	)
	cmd := &cobra.Command{
		Use:   "endpoints <app>",
		Short: "which routes this deployment actually serves",
		Long: `List the routes a model application serves, and which are switched off.

A Model Console mounts different routes depending on what it is running: an
embedding deployment has no chat completions, a translation one moves its routes
to the host root, and the engine-native passthrough exists only for the engines
that have one. So "the endpoint does not exist" and "the endpoint is not mounted
in this deployment" are different answers, and this is where they are told apart
— each unavailable route carries the reason it is absent.

Examples:
  olares-cli router local endpoints llamacppqwen3627bggufv3
  olares-cli router local endpoints llamacppqwen3627bggufv3 --unavailable
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			li, err := openLocal(ctx, f, args[0])
			if err != nil {
				return err
			}
			var env struct {
				EngineKind string          `json:"engine_kind"`
				Endpoints  []localEndpoint `json:"endpoints"`
			}
			if err := li.client.doJSON(ctx, "GET", epLocalEndpoints, nil, &env); err != nil {
				return err
			}
			rows := env.Endpoints
			if unavailable {
				kept := rows[:0:0]
				for _, r := range rows {
					if !r.Available {
						kept = append(kept, r)
					}
				}
				rows = kept
			}
			if format == FormatJSON {
				return printJSON(os.Stdout, map[string]any{
					"engine_kind": env.EngineKind,
					"endpoints":   rows,
				})
			}
			return renderLocalEndpoints(os.Stdout, env.EngineKind, rows, unavailable)
		},
	}
	cmd.Flags().BoolVar(&unavailable, "unavailable", false, "only the routes this deployment does not serve")
	addOutputFlag(cmd, &output)
	return cmd
}

func renderLocalEndpoints(w io.Writer, engineKind string, rows []localEndpoint, onlyUnavailable bool) error {
	if len(rows) == 0 {
		if onlyUnavailable {
			_, err := fmt.Fprintln(w, "every route this Model Console knows about is mounted.")
			return err
		}
		_, err := fmt.Fprintln(w, "this Model Console reported no routes at all, which should not happen.")
		return err
	}
	if engineKind != "" {
		if _, err := fmt.Fprintf(w, "engine: %s\n\n", engineKind); err != nil {
			return err
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Category != rows[j].Category {
			return rows[i].Category < rows[j].Category
		}
		return rows[i].Path < rows[j].Path
	})
	t := newTable(w, "METHOD", "PATH", "CATEGORY", "SERVED", "WHY NOT")
	for _, r := range rows {
		why := ""
		if !r.Available {
			why = clip(strings.Join(r.Reasons, "; "), 60)
			if why == "" {
				why = "no reason given"
			}
		}
		t.row(r.Method, r.Path, r.Category, boolStr(r.Available), why)
	}
	return t.flush()
}
