package router

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router default …` — which model answers when a caller names none.
//
// GET  /console/api/default-models              resolved, any user
// GET  /console/api/default-models?layer=system raw system layer, admin
// POST /console/api/default-models              set the system layer, admin
// POST /console/api/account/default-models      set your own override
//
// A caller that omits the model, or asks for the alias `default`, gets whatever
// is configured here for the mode its request belongs to. There are three places
// an answer can come from and they are consulted in order: your own override,
// the system-wide choice an admin made, and failing both, the oldest enabled
// model of that mode. The third is a fallback rather than a setting, which is
// worth seeing — it changes on its own when models are added or removed.
//
// This is one of the two reads here open to a non-admin: knowing what your own
// blank request will hit is not privileged, and setting your own override is
// self-scoped by the route rather than by a check.

// defaultModeOrder is the set of modes a default can be set for. Router's
// dispatcher does not route a blank `responses` request, so that mode is absent
// here while being a legal mode for a model row.
var defaultModeOrder = []string{
	"chat", "completion", "embedding", "rerank", "moderation",
	"audio", "translate", "image_generation", "ocr",
}

type resolvedDefault struct {
	Mode            string `json:"mode"`
	ProviderModelID string `json:"provider_model_id"`
	Layer           string `json:"layer"`
}

type systemDefault struct {
	Mode            string `json:"mode"`
	ProviderModelID string `json:"provider_model_id"`
	UpdatedBy       string `json:"updated_by"`
	UpdatedAt       string `json:"updated_at"`
}

func NewDefaultCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default",
		Short: "which model answers when a request names none",
		Long: `Show and set the model a request falls back to.

A request that omits the model, or names the alias "default", is answered by
whatever is configured here for that kind of request.

Three sources are consulted in order, and the answer says which one won:

  tenant                  your own override, set with "default set --mine"
  system                  the workspace-wide choice, set by an admin
  fallback_first_enabled   neither was set, so the oldest enabled model of
                          that kind was used

The third is not a setting. It moves on its own as models are added, disabled or
removed, which makes it a reasonable way to get started and a poor thing to rely
on.

Subcommands:
  show          what is in effect for you, and where each answer came from
  set           choose a default, for the workspace or just for yourself
  clear         drop a choice, falling back to the layer beneath it

Reading is open to everyone. Setting the workspace default is admin-only;
setting your own is not.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newDefaultShowCommand(f))
	cmd.AddCommand(newDefaultSetCommand(f))
	cmd.AddCommand(newDefaultClearCommand(f))
	return cmd
}

func newDefaultShowCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		system bool
	)
	cmd := &cobra.Command{
		Use:   "show",
		Short: "the defaults in effect, and which layer each came from",
		Long: `Show the default model per kind of request.

By default this is the resolved view: what your own blank requests will actually
hit, with the layer that decided it. A mode missing from the list has no default
anywhere, and a blank request of that kind is refused rather than guessed.

--system shows the raw workspace layer instead — only what an admin set, with no
fallback filled in and no override of yours applied. That is the view to use
before changing a workspace default, since the resolved view cannot tell you
whether a value came from the workspace or from your own override. Admin only.

Model ids are shown alongside their names where the name is known. The id is
what "default set" takes.

Examples:
  olares-cli router default show
  olares-cli router default show --system
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runDefaultShow(c.Context(), f, system, output)
		},
	}
	cmd.Flags().BoolVar(&system, "system", false, "show the raw workspace layer instead of the resolved view (admin only)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runDefaultShow(ctx context.Context, f *cmdutil.Factory, system bool, outputRaw string) error {
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

	if system {
		var env struct {
			Items []systemDefault `json:"items"`
		}
		path := withQuery(epDefaultModels, url.Values{"layer": {"system"}})
		if err := pc.router.doJSON(ctx, "GET", path, nil, &env); err != nil {
			return err
		}
		if format == FormatJSON {
			return printJSON(os.Stdout, env)
		}
		return renderSystemDefaults(ctx, pc, os.Stdout, env.Items)
	}

	var env struct {
		Items []resolvedDefault `json:"items"`
	}
	if err := pc.router.doJSON(ctx, "GET", epDefaultModels, nil, &env); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, env)
	}
	return renderResolvedDefaults(ctx, pc, os.Stdout, env.Items)
}

func renderResolvedDefaults(ctx context.Context, pc *preparedClient, w io.Writer, items []resolvedDefault) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no default resolves for any kind of request, so a request that names no "+
			"model is refused. Attach a model to a provider first — `olares-cli router list` shows what exists.")
		return err
	}
	names := modelNames(ctx, pc)
	t := newTable(w, "MODE", "MODEL", "FROM", "MODEL ID")
	for i := range items {
		it := &items[i]
		t.row(it.Mode, modelLabel(names, it.ProviderModelID), describeLayer(it.Layer), it.ProviderModelID)
	}
	if err := t.flush(); err != nil {
		return err
	}
	if missing := missingModes(items); len(missing) > 0 {
		if _, err := fmt.Fprintf(w, "\nno default for %s; a request of those kinds must name its model\n",
			strings.Join(missing, ", ")); err != nil {
			return err
		}
	}
	return nil
}

func renderSystemDefaults(ctx context.Context, pc *preparedClient, w io.Writer, items []systemDefault) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "the workspace has no defaults set, so every blank request falls back to "+
			"the oldest enabled model of its kind. Set one with `olares-cli router default set`.")
		return err
	}
	names := modelNames(ctx, pc)
	users := userLabels(ctx, pc)
	t := newTable(w, "MODE", "MODEL", "SET BY", "WHEN", "MODEL ID")
	for i := range items {
		it := &items[i]
		by := it.UpdatedBy
		if label, ok := users[by]; ok && label != "" {
			by = label
		}
		t.row(it.Mode, modelLabel(names, it.ProviderModelID), nonEmpty(by), nonEmpty(it.UpdatedAt), it.ProviderModelID)
	}
	return t.flush()
}

func newDefaultSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		mine   bool
		pairs  []string
	)
	cmd := &cobra.Command{
		Use:   "set --mode <mode>=<model>...",
		Short: "choose the default model for one or more kinds of request",
		Long: `Set the default model for a kind of request.

Without --mine this changes the workspace default, which every user without an
override of their own will hit. With --mine it sets only your own override,
which wins over the workspace choice for you and affects nobody else.

Models are named by id, by provider and name as "<provider>/<model>", or by name
alone when only one provider offers it. "router list" shows all three.

Several modes can be set at once, and either all of them land or none do.

A model has to answer the kind of request you are pointing it at. Router does
not check this and will happily record an embedding model as the chat default,
which fails only when somebody makes a call; this command refuses it up front.

Setting the workspace default is admin-only. Setting your own is not.

Examples:
  olares-cli router default set --mode chat=claude/claude-sonnet-4-5
  olares-cli router default set --mine --mode chat=lmstudio/qwen3-8b
  olares-cli router default set --mode chat=<id> --mode embedding=<id>
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runDefaultSet(c.Context(), f, pairs, mine, output)
		},
	}
	cmd.Flags().StringArrayVar(&pairs, "mode", nil, "assignment as <mode>=<model>; repeatable")
	cmd.Flags().BoolVar(&mine, "mine", false, "set your own override instead of the workspace default")
	addOutputFlag(cmd, &output)
	return cmd
}

type modelSetting struct {
	Mode            string `json:"mode"`
	ProviderModelID string `json:"provider_model_id"`
}

func runDefaultSet(ctx context.Context, f *cmdutil.Factory, pairs []string, mine bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if len(pairs) == 0 {
		return fmt.Errorf("pass at least one --mode <mode>=<model>")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	known, err := listAllModels(ctx, pc)
	if err != nil {
		return err
	}
	settings := make([]modelSetting, 0, len(pairs))
	for _, p := range pairs {
		mode, ref, ok := strings.Cut(p, "=")
		mode = strings.ToLower(strings.TrimSpace(mode))
		ref = strings.TrimSpace(ref)
		if !ok || mode == "" || ref == "" {
			return fmt.Errorf("--mode %q is not in <mode>=<model> form", p)
		}
		if !containsString(defaultModeOrder, mode) {
			return fmt.Errorf("--mode %s: a default cannot be set for %q; Router routes blank requests for %s",
				p, mode, strings.Join(defaultModeOrder, ", "))
		}
		row, rerr := findModelForMode(known, ref, mode)
		if rerr != nil {
			return rerr
		}
		settings = append(settings, modelSetting{Mode: mode, ProviderModelID: row.Model.ID})
	}
	return postDefaults(ctx, pc, settings, mine, format, "set")
}

func newDefaultClearCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		mine   bool
	)
	cmd := &cobra.Command{
		Use:   "clear <mode>...",
		Short: "drop a default, falling back to the layer beneath it",
		Long: `Remove a default so the layer beneath it takes over.

Clearing your own override, with --mine, hands you back to the workspace
default. Clearing the workspace default hands everyone without an override to
the fallback — the oldest enabled model of that kind — which is a value nobody
chose and which moves as models change.

Clearing a mode that was never set is not an error.

Clearing the workspace default is admin-only. Clearing your own is not.

Examples:
  olares-cli router default clear --mine chat
  olares-cli router default clear chat embedding
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runDefaultClear(c.Context(), f, args, mine, output)
		},
	}
	cmd.Flags().BoolVar(&mine, "mine", false, "clear your own override instead of the workspace default")
	addOutputFlag(cmd, &output)
	return cmd
}

func runDefaultClear(ctx context.Context, f *cmdutil.Factory, modes []string, mine bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	settings := make([]modelSetting, 0, len(modes))
	for _, m := range modes {
		mode := strings.ToLower(strings.TrimSpace(m))
		if !containsString(defaultModeOrder, mode) {
			return fmt.Errorf("%q is not a mode a default can be set for; Router routes blank requests for %s",
				m, strings.Join(defaultModeOrder, ", "))
		}
		// The empty id is Router's clear sentinel on the same batch route.
		settings = append(settings, modelSetting{Mode: mode, ProviderModelID: ""})
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	return postDefaults(ctx, pc, settings, mine, format, "cleared")
}

// postDefaults sends one batch to whichever layer was asked for. Router answers
// the system route with the raw system layer and the account route with the
// resolved view, so the echo differs by layer and is rendered accordingly.
func postDefaults(ctx context.Context, pc *preparedClient, settings []modelSetting, mine bool, format Format, verb string) error {
	body := map[string]any{"model_settings": settings}
	if mine {
		var env struct {
			Items []resolvedDefault `json:"items"`
		}
		if err := pc.router.doJSON(ctx, "POST", epAccountDefaultModels, body, &env); err != nil {
			return err
		}
		if format == FormatJSON {
			return printJSON(os.Stdout, env)
		}
		if _, err := fmt.Fprintf(os.Stdout, "%s for you; in effect now:\n\n", verb); err != nil {
			return err
		}
		return renderResolvedDefaults(ctx, pc, os.Stdout, env.Items)
	}

	var env struct {
		Items []systemDefault `json:"items"`
	}
	if err := pc.router.doJSON(ctx, "POST", epDefaultModels, body, &env); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, env)
	}
	if _, err := fmt.Fprintf(os.Stdout, "%s for the workspace; the workspace layer is now:\n\n", verb); err != nil {
		return err
	}
	return renderSystemDefaults(ctx, pc, os.Stdout, env.Items)
}

// findModelForMode resolves a reference against the models Router knows and
// checks that the one found answers the requested kind of request. Router itself
// only checks that the row exists, so an embedding model recorded as the chat
// default is accepted there and fails at call time; catching it here is the
// difference between a rejected flag and a broken workspace.
//
// A bare name is accepted when exactly one provider offers it. Ambiguity is
// reported with the qualified alternatives rather than resolved by picking one,
// because the two rows can be different models with different prices.
func findModelForMode(known []adminModelRow, ref, mode string) (*adminModelRow, error) {
	var matches []*adminModelRow
	for i := range known {
		r := &known[i]
		qualified := r.ProviderName + "/" + r.Model.Name
		switch {
		case r.Model.ID == ref,
			strings.EqualFold(qualified, ref),
			strings.EqualFold(r.Model.Name, ref),
			r.Model.Alias != nil && strings.EqualFold(*r.Model.Alias, ref):
			matches = append(matches, r)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no model %q; `olares-cli router list` shows what is configured", ref)
	}
	if len(matches) > 1 {
		names := make([]string, 0, len(matches))
		for _, m := range matches {
			names = append(names, m.ProviderName+"/"+m.Model.Name)
		}
		sort.Strings(names)
		return nil, fmt.Errorf("%q is offered by more than one provider; name one of %s",
			ref, strings.Join(names, ", "))
	}
	found := matches[0]
	if !strings.EqualFold(found.Model.Mode, mode) {
		return nil, fmt.Errorf("%s/%s answers %s requests, not %s, so it cannot be the %s default",
			found.ProviderName, found.Model.Name, found.Model.Mode, mode, mode)
	}
	if !found.Model.Enabled {
		return nil, fmt.Errorf("%s/%s is not offered to callers, so a request falling back to it would fail; "+
			"enable it with `olares-cli router provider models update %s %s --enable`",
			found.ProviderName, found.Model.Name, found.ProviderName, found.Model.Name)
	}
	return found, nil
}

// modelNames maps model ids to a readable label. The default-model routes carry
// ids alone, and an id is not something a person recognises; a failed lookup is
// not worth failing the command over, so the id is shown bare instead.
func modelNames(ctx context.Context, pc *preparedClient) map[string]string {
	rows, err := listAllModels(ctx, pc)
	if err != nil {
		return nil
	}
	out := make(map[string]string, len(rows))
	for i := range rows {
		r := &rows[i]
		out[r.Model.ID] = r.ProviderName + "/" + r.Model.Name
	}
	return out
}

func listAllModels(ctx context.Context, pc *preparedClient) ([]adminModelRow, error) {
	return collection[adminModelRow](ctx, pc, withQuery(epProviderModels, url.Values{"limit": {"1000"}}))
}

func modelLabel(names map[string]string, id string) string {
	if id == "" {
		return "-"
	}
	if label, ok := names[id]; ok && label != "" {
		return label
	}
	return "(unknown)"
}

// describeLayer spells out the fallback arm. "fallback_first_enabled" is a
// stable wire label rather than something to show a person, and the distinction
// it draws — nobody chose this — is the point of the column.
func describeLayer(layer string) string {
	switch layer {
	case "tenant":
		return "your override"
	case "system":
		return "workspace"
	case "fallback_first_enabled":
		return "fallback (unset)"
	}
	return nonEmpty(layer)
}

func missingModes(items []resolvedDefault) []string {
	have := make(map[string]bool, len(items))
	for i := range items {
		have[items[i].Mode] = true
	}
	var out []string
	for _, m := range defaultModeOrder {
		if !have[m] {
			out = append(out, m)
		}
	}
	sort.Strings(out)
	return out
}
