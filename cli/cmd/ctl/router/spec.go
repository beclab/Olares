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

// `olares-cli router spec` — the model card of a model application, through
// Router.
//
// GET   /console/api/model-spec?model=<ref>
// PATCH /console/api/model-spec?model=<ref>
// POST  /console/api/engine/restart?model=<ref>
//
// The card is where a model's mode, capabilities, prices and engine flags are
// declared, and the application that runs the model owns it. Router holds a
// projection of it — that is what `router list` reports and what billing is
// calculated against — and reads it when the application reaches running, so
// the two can be a restart apart.
//
// There are two ways to reach the card and they are not interchangeable.
// `router local spec` goes straight at the application's Model Console, which
// is the only option when Router does not have the application as a provider,
// and its write is a whole-document replace: a field left out of what you send
// is gone, engine flags included. This subtree goes through Router, which
// merges the change onto the card the application currently serves, keeps every
// key neither side has heard of, and stores what the application confirms — so
// the projection is right immediately rather than at the next restart.
//
// Prefer this one. `router local spec set` exists for the case Router cannot
// answer for.
//
// Admin only, the read included, because the same route serves the engine
// flags an inference process is relaunched with.

// specEnvelope is a read. `source` is the interesting field: a paused
// application cannot be asked, so Router answers with its own projection and
// says so rather than pretending it spoke to the app.
type specEnvelope struct {
	Source string          `json:"source"`
	Card   json.RawMessage `json:"card"`
}

const (
	specSourceLive  = "live"
	specSourceCache = "cache"
)

// specWriteResult is what a write answers with. Spec is the card the
// application confirmed keeping, which is not the card that was sent: the Model
// Console adds the flags a mode requires and files capability keys it does not
// know under extensions.
type specWriteResult struct {
	Spec           json.RawMessage `json:"spec,omitempty"`
	Restarted      bool            `json:"restarted"`
	RestartWarning string          `json:"restart_warning,omitempty"`
}

func NewSpecCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "the model card of a model application, through Router",
		Long: `Read and change the card that declares what a local model is.

Mode, capabilities, prices and engine flags all live in the card, and the
application running the model owns it. Router keeps a projection to route and
bill against, and refreshes it when the application reaches running — so a card
edited any other way is a restart away from being what Router believes.

  spec show <model>      the card, and whether it came from the app or the cache
  spec edit <model>      change some of it and leave the rest alone
  spec restart <model>   relaunch the engine on the card already on disk

The model is named the way a caller names it: "<provider>/<model>", or the bare
model name when only one application serves it. A bare name that matches two is
refused with the qualified candidates rather than guessed at.

This works on Olares model applications only — a cloud provider has no card and
no control plane. "olares-cli router local spec" is the other route to the same
document, straight at the application, and replaces it wholesale; this one
merges, and is the one to reach for.

Admin only, reading included.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newSpecShowCommand(f))
	cmd.AddCommand(newSpecEditCommand(f))
	cmd.AddCommand(newSpecRestartCommand(f))
	return cmd
}

func newSpecShowCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:     "show <model>",
		Aliases: []string{"get"},
		Short:   "the card, and where the answer came from",
		Long: `Show a model's card as Router sees it.

A running application is asked directly. One that is stopped or still
installing cannot be, so Router answers with its own projection and marks it
"cache" — which is the answer to "why does Router advertise something this
model cannot do": the projection is what routing and billing use, and it trails
the card until the application next comes up.

JSON output is the card itself rather than a summary, including fields this CLI
does not know about, which is what makes it the thing to edit and feed back to
"spec edit --from".

Examples:
  olares-cli router spec show Olares/qwen3-4b
  olares-cli router spec show qwen3-4b -o json > card.json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runSpecShow(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runSpecShow(ctx context.Context, f *cmdutil.Factory, model, outputRaw string) error {
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
	var env specEnvelope
	if err := pc.router.doJSON(ctx, "GET", epForModel(epModelSpec, model), nil, &env); err != nil {
		return specErr(err, model)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, env.Card)
	}
	var card localSpec
	if err := json.Unmarshal(env.Card, &card); err != nil {
		return fmt.Errorf("this card is not a shape this version understands: %w\n"+
			"-o json prints it as it came", err)
	}
	return renderSpecCard(os.Stdout, model, env.Source, &card)
}

// renderSpecCard prints the same fields `router local spec show` does, and says
// which copy this is. The two verbs deliberately look alike: they are reads of
// one document, and a reader comparing them should be comparing values rather
// than layouts.
func renderSpecCard(w io.Writer, model, source string, s *localSpec) error {
	t := newTable(w)
	t.row("MODEL", nonEmpty(s.Name))
	t.row("ASKED FOR", model)
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
		if _, err := fmt.Fprintln(w, "\nParameter rules are set; -o json prints them."); err != nil {
			return err
		}
	}
	return specSourceNote(w, source)
}

func specSourceNote(w io.Writer, source string) error {
	switch source {
	case specSourceLive:
		_, err := fmt.Fprintln(w, "\nRead from the application itself, so this is what it is running.")
		return err
	case specSourceCache:
		_, err := fmt.Fprintln(w, "\nThe application could not be asked, so this is Router's stored copy — "+
			"what routing and billing use. It is refreshed when the application next reaches running.")
		return err
	}
	if strings.TrimSpace(source) == "" {
		return nil
	}
	_, err := fmt.Fprintf(w, "\nSource: %s\n", source)
	return err
}

func newSpecEditCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output     string
		from       string
		mode       string
		engineArgs string
		yes        bool
	)
	cmd := &cobra.Command{
		Use:   "edit <model>",
		Short: "change part of the card and leave the rest",
		Long: `Change some of a model's card, keeping everything you do not mention.

The change is a JSON object of the keys to replace, from --from or standard
input, and --mode and --engine-args are shorthands for the two people actually
reach for. Router reads the card the application is serving, merges yours onto
it, sends the result back, and stores what the application confirms — so keys
this CLI has never heard of survive, and Router's own copy is right at once
rather than after the next restart.

The merge is one level deep. "pricing", "supports" and "parameter_rules" are
each replaced whole, so changing one price means sending the pricing object you
want, not just the member that changed. A key set to null is dropped from the
card.

Two things follow from the application owning the document. Changing
--engine-args relaunches the inference process, which takes as long as loading
the model takes; the reply says whether the relaunch was signalled, not whether
it has finished. And the application has to be running: a card cannot be
written to something that is not there, and "spec show" will say "cache" when
that is the case.

An application whose engine is a sidecar — OCR, audio, embedding — takes no
engine flags at all, and there --mode is the field that matters, because it is
the gate the data plane routes on.

Examples:
  olares-cli router spec edit Olares/qwen3-4b --mode chat
  olares-cli router spec edit Olares/qwen3-4b --engine-args "--ctx-size 8192 --n-gpu-layers 99"
  olares-cli router spec edit Olares/qwen3-4b --from patch.json
  echo '{"supports":{"supports_vision":true}}' | olares-cli router spec edit Olares/qwen3-4b
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			patch, err := specPatch(c, from, mode, engineArgs)
			if err != nil {
				return err
			}
			return runSpecEdit(c.Context(), f, args[0], patch, yes, output)
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "read the keys to change from this file, or `-` for standard input")
	cmd.Flags().StringVar(&mode, "mode", "", "declare the model's mode, e.g. chat, embedding, ocr, audio")
	cmd.Flags().StringVar(&engineArgs, "engine-args", "", "flags to relaunch the inference engine with")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	addOutputFlag(cmd, &output)
	return cmd
}

// specPatch assembles the object to send. A document and the two shorthands can
// be combined — a patch file plus a mode is a reasonable thing to want — and the
// flags win, because they are the more specific instruction.
func specPatch(c *cobra.Command, from, mode, engineArgs string) (map[string]any, error) {
	patch := map[string]any{}
	if raw, ok, err := readSpecPatchDocument(c, from); err != nil {
		return nil, err
	} else if ok {
		if err := json.Unmarshal(raw, &patch); err != nil {
			return nil, fmt.Errorf("what you sent is not valid JSON: %w", err)
		}
		if len(patch) == 0 {
			return nil, fmt.Errorf("the document names no keys to change; " +
				"`olares-cli router spec show <model> -o json` prints the whole card to start from")
		}
	}
	if c.Flags().Changed("mode") {
		if strings.TrimSpace(mode) == "" {
			return nil, fmt.Errorf("--mode needs a value; a model with no mode cannot be routed to")
		}
		patch["mode"] = strings.TrimSpace(mode)
	}
	if c.Flags().Changed("engine-args") {
		// Kept exactly as typed, including empty. Clearing the flags is a
		// real thing to want, and it relaunches the engine bare.
		patch["engine_args"] = engineArgs
	}
	if len(patch) == 0 {
		return nil, fmt.Errorf("nothing to change: pass --mode, --engine-args, or a document with " +
			"--from. `olares-cli router spec show <model> -o json` prints the current card")
	}
	return patch, nil
}

// readSpecPatchDocument reads the document, if there is one. A piped-in patch
// with no --from is the ordinary shape, so standard input counts — but only when
// it is not a terminal, since `spec edit --mode chat` at a prompt is not asking
// to be waited on.
//
// Nothing on standard input is "no document" rather than an error, and the
// caller's own refusal is the one that names this verb. `--from` is different: a
// file was named, so an unreadable or empty one is a mistake to report.
func readSpecPatchDocument(c *cobra.Command, from string) ([]byte, bool, error) {
	if from != "" && from != "-" {
		raw, err := os.ReadFile(from)
		if err != nil {
			return nil, false, err
		}
		if len(bytes.TrimSpace(raw)) == 0 {
			return nil, false, fmt.Errorf("%s is empty; it should hold the keys to change, as a JSON "+
				"object. `olares-cli router spec show <model> -o json` prints the whole card", from)
		}
		return raw, true, nil
	}
	if isTerminal(os.Stdin) {
		return nil, false, nil
	}
	if from == "" && (c.Flags().Changed("mode") || c.Flags().Changed("engine-args")) {
		// A shorthand was given and stdin happens to be a pipe — a log
		// redirect, a CI runner. Reading it would send whatever it holds as
		// a card patch.
		return nil, false, nil
	}
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, false, fmt.Errorf("read the change from standard input: %w", err)
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, false, nil
	}
	return raw, true, nil
}

func runSpecEdit(ctx context.Context, f *cmdutil.Factory, model string, patch map[string]any,
	yes bool, outputRaw string) error {
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
	if !yes {
		if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin, fmt.Sprintf(
			"Change %s on %s? %s",
			strings.Join(sortedKeys(patch), ", "), model, specEditConsequence(patch)), false); err != nil {
			return err
		}
	}
	var res specWriteResult
	if err := pc.router.doJSON(ctx, "PATCH", epForModel(epModelSpec, model), patch, &res); err != nil {
		return specErr(err, model)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, res)
	}
	return renderSpecWrite(os.Stdout, model, &res)
}

// specEditConsequence names the one edit with a cost. Everything else in the
// card is a declaration Router reads; engine flags restart the process serving
// the model, and a large model takes minutes to load again.
func specEditConsequence(patch map[string]any) string {
	if _, ok := patch["engine_args"]; ok {
		return "Changing the engine flags relaunches the inference process, " +
			"which is unavailable until the model has loaded again."
	}
	return "The application keeps every key you did not mention."
}

func sortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func renderSpecWrite(w io.Writer, model string, res *specWriteResult) error {
	if len(res.Spec) > 0 {
		var card localSpec
		if err := json.Unmarshal(res.Spec, &card); err == nil {
			if err := renderSpecCard(w, model, "", &card); err != nil {
				return err
			}
		}
	}
	switch {
	case res.RestartWarning == "run_dir_unset":
		_, err := fmt.Fprintln(w, "\nThe card is written, but this application cannot signal its engine, "+
			"so the flags take effect only when the application itself restarts: "+
			"`olares-cli market restart <app>`.")
		return err
	case res.RestartWarning != "":
		_, err := fmt.Fprintf(w, "\nThe card is written. The engine was not relaunched: %s\n", res.RestartWarning)
		return err
	case res.Restarted:
		_, err := fmt.Fprintln(w, "\nWritten, and the engine is relaunching. It answers again once the "+
			"model has loaded; `olares-cli router local status <app>` reports how far along that is.")
		return err
	}
	_, err := fmt.Fprintln(w, "\nWritten. Nothing needed relaunching, so this is in effect now. "+
		"Router's own copy was updated with what the application confirmed.")
	return err
}

func newSpecRestartCommand(f *cmdutil.Factory) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "restart <model>",
		Short: "relaunch the engine on the card it already has",
		Long: `Relaunch the inference process without changing anything.

This is for an engine that has stopped behaving rather than one that is
configured wrongly — wedged, leaking, answering slowly after a long session. The
card is not read and not written; the process is told to come back on the
document already on disk.

A success means the relaunch was signalled. The engine then shuts down and
loads the model again, and the model does not answer while it does, which for a
large one is minutes. An application whose engine is a sidecar — OCR, audio,
embedding — has no process to signal here, and the call succeeds having changed
nothing.

Examples:
  olares-cli router spec restart Olares/qwen3-4b
  olares-cli router spec restart Olares/qwen3-4b -y
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			pc, err := prepare(ctx, f)
			if err != nil {
				return err
			}
			if !yes {
				if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin, fmt.Sprintf(
					"Relaunch the engine serving %s? It stops answering until the model has loaded again.",
					args[0]), false); err != nil {
					return err
				}
			}
			var res struct {
				Restarted bool `json:"restarted"`
			}
			if err := pc.router.doJSON(ctx, "POST",
				epForModel(epEngineRestart, args[0]), nil, &res); err != nil {
				return specErr(err, args[0])
			}
			if !res.Restarted {
				_, werr := fmt.Println("the application accepted the request and reports no relaunch, " +
					"which is what an application with a sidecar engine does.")
				return werr
			}
			_, werr := fmt.Println("relaunching. The model answers again once it has loaded; " +
				"`olares-cli router local status <app>` reports how far along that is.")
			return werr
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	return cmd
}

// specErr explains the refusals particular to this surface. Every one of them
// is about which application is meant or whether it can be spoken to, and none
// of them is fixed by retrying.
func specErr(err error, model string) error {
	var re *RouterError
	if err == nil || !errors.As(err, &re) {
		return err
	}
	switch re.Code {
	case "model_spec_model_not_found":
		return fmt.Errorf("%w\n`olares-cli router list` shows the configured models. A model that is "+
			"disabled is not resolvable here either", err)
	case "model_spec_model_ambiguous":
		return fmt.Errorf("%w\nQualify it as <provider>/<model>; `olares-cli router list` shows which "+
			"providers serve %s. This route refuses to guess, because the wrong guess reconfigures the "+
			"wrong application", err, model)
	case "model_spec_unsupported_provider":
		return fmt.Errorf("%w\nOnly a model application on this Olares has a card — a cloud provider's "+
			"models are described by Router's own catalogue, and `olares-cli router provider models "+
			"update` is what changes those", err)
	case "model_spec_app_unavailable":
		return fmt.Errorf("%w\n`olares-cli router provider get <provider>` shows the application's state, "+
			"and `olares-cli market resume <app>` starts it. `router spec show` still answers meanwhile, "+
			"from Router's stored copy", err)
	case "model_spec_app_not_ready":
		return fmt.Errorf("%w\nThe application has an entry but no address yet, which is where it sits "+
			"while installing. `olares-cli router local status <app>` reports the phase", err)
	case "model_spec_rejected":
		return fmt.Errorf("%w\nThe application rejected the card and kept the one it had; the message "+
			"above is its own and names the field", err)
	case "model_spec_unsupported_app":
		return fmt.Errorf("%w\nThis application is too old to serve its card. Upgrading it through the "+
			"Market is what adds the route", err)
	case "invalid_model_spec_body":
		return fmt.Errorf("%w\nThe change has to be a JSON object of top-level keys. "+
			"`olares-cli router spec show %s -o json` prints the card its keys come from", err, model)
	case "model_spec_too_large":
		return fmt.Errorf("%w\nThe merged card is larger than the application accepts", err)
	case "engine_restart_unavailable":
		return fmt.Errorf("%w\nThis application cannot signal its engine, so restarting the application "+
			"itself is the way: `olares-cli market restart <app>`", err)
	case "model_spec_upstream_unreachable", "model_spec_upstream_timeout", "model_spec_upstream_error":
		return fmt.Errorf("%w\nRouter reached the application's entry but not its control plane. "+
			"`olares-cli router local status <app>` says whether it is up, and a model still loading "+
			"is the usual reason", err)
	}
	return err
}
