package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router model …` — the models this Router is configured with.
//
// POST   /console/api/providers/:id/predefined-models
// POST   /console/api/providers/:id/customizable-models
// PATCH  /console/api/providers/:id/models/:model_id
// DELETE /console/api/providers/:id/models/:model_id
//
// Reads have no route of their own: Router serves a provider's models inline on
// the provider detail, so `model get` and every verb that resolves a model by
// name go through GET /console/api/providers/:id.
//
// Every write route is scoped to a provider, and the verbs here are not. That
// is deliberate: a provider is where a model is *served from*, and naming it is
// only unavoidable when the model does not exist yet. For a model that does,
// the name identifies it and `--provider` is a tiebreak for the one case it
// cannot — every locally installed model application is a provider called
// `Olares`, so two of them can carry the same model name.
//
// A provider having credentials does not make its models available: each one is
// a row you attach on purpose. Where that row comes from depends on the vendor —
// imported from the catalog Router ships, named by hand for an endpoint whose
// catalog is its own, or mirrored from the upstream by `provider sync-models`.
//
// What a model row carries beyond its name is a description of what it can do:
// the capability flags Router checks before dispatching a request, its context
// window, and its prices. Those are what `model update` edits, and getting
// them wrong is visible — a model marked as not supporting vision will refuse
// an image the upstream would have accepted.

// providerModelModes is Router's mode allowlist. The mode decides which
// endpoint family a model answers on, so it cannot be changed later: a row
// created as chat stays chat, and the fix for a wrong one is to delete and add
// it again.
var providerModelModes = []string{
	"chat", "embedding", "rerank", "moderation",
	"audio", "translate", "image_generation", "responses", "ocr",
	"search", "scrape", "video_generation",
}

// NewModelCommand assembles the model noun: one place for every model this
// Router knows about, whichever provider serves it.
func NewModelCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "the models this Router is configured with",
		Long: `Manage the models this Router can reach.

Credentials alone offer nothing: every model is a row attached on purpose, so
that a key does not silently expose a vendor's entire catalog.

Subcommands:
  get <model>       one model in full: every capability, price, and the
                    engine flags of a local model
  import <model>... attach models from Router's vendor catalog, with their
                    capabilities and prices filled in
  add <model>       attach a model you name yourself, for an endpoint whose
                    catalog is its own
  update <model>    enable, disable, or correct what a model claims to support
  remove <model>    detach a model

A model is named the way a caller names it, "<provider>/<model>", or by its
own name when only one row carries it, or by its id. --provider settles the
one case a name cannot: every locally installed model application is a
provider called "Olares", so two of them can offer the same model name.
Attaching a model needs --provider outright, since the row does not exist yet.

For an upstream that publishes its own list — Ollama, a local model application
— none of these is the usual route: "provider sync-models" mirrors the whole
list in one step.

"olares-cli router list" shows every model across every provider.

Admin only, except for reading.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newModelGetCommand(f))
	cmd.AddCommand(newModelImportCommand(f))
	cmd.AddCommand(newModelAddCommand(f))
	cmd.AddCommand(newModelUpdateCommand(f))
	cmd.AddCommand(newModelRemoveCommand(f))
	return cmd
}

func newModelImportCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		provider string
	)
	cmd := &cobra.Command{
		Use:   "import --provider <provider> <model>...",
		Short: "attach models from Router's vendor catalog",
		Long: `Attach one or more of a vendor's catalog models to a provider.

The catalog ships inside Router, so an imported model arrives already
describing itself: its capabilities, context window, parameter rules and
published prices are filled in without a round trip to the vendor.

Model names must match the catalog exactly. If any name is unknown, nothing is
imported at all — a partial import would leave you guessing which of a long list
landed. Router does not say which name it objected to, so check the list with
"provider types <vendor> --models".

A model already attached is skipped rather than reset, so re-running an import
never discards an edit you made to what it claims to support.

Example:
  olares-cli router model import --provider claude \
    claude-sonnet-4-5 claude-haiku-4-5
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runModelImport(c.Context(), f, provider, args, output)
		},
	}
	cmd.Flags().StringVar(&provider, "provider", "", "the provider to attach them to (required)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runModelImport(ctx context.Context, f *cmdutil.Factory, ref string, names []string, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if err := requireProviderFlag(ref, "import"); err != nil {
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
	if found.isMarketSourced() {
		return marketOwnedErr(found, "attach models to")
	}
	var res struct {
		Created int `json:"created"`
		Skipped int `json:"skipped"`
		Total   int `json:"total"`
	}
	path := epProviderPredefinedModels(found.ID)
	body := map[string][]string{"model_names": names}
	if err := pc.router.doJSON(ctx, "POST", path, body, &res); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, res)
	}
	_, err = fmt.Fprintf(os.Stdout, "%s: %d models attached, %d already present (%d requested)\n",
		found.Name, res.Created, res.Skipped, res.Total)
	return err
}

func newModelAddCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output       string
		provider     string
		mode         string
		contextSize  int
		maxOutTokens int
		supports     []string
		pricing      []string
	)
	cmd := &cobra.Command{
		Use:   "add <model-name> --provider <provider>",
		Short: "attach a model you name yourself",
		Long: `Attach a model by name, for an upstream whose catalog is not published.

--provider is required: the row does not exist yet, so there is nothing for a
name to identify.

The model name is sent to the upstream verbatim, so it has to be what that
upstream calls the model rather than a label of your choosing. A name of your
own is a route over the top of it: "olares-cli router route create <name>
--kind alias --model <provider>/<model>".

--mode decides which endpoint family the model answers on and cannot be changed
afterwards; a row created with the wrong mode has to be removed and added again.
Allowed: ` + strings.Join(providerModelModes, ", ") + `.

Capabilities default to none, which is a claim in itself: Router refuses to
dispatch an image to a model that does not declare vision, even when the
upstream would have accepted it. Declare what the model really does with
--supports, and correct it later with "model update".

Prices are per million tokens as the vendor states them, and are what usage
reporting multiplies. A model with none recorded still works; its spend is
simply reported as zero.

Examples:
  olares-cli router model add qwen3-8b --provider lmstudio --mode chat \
    --context-size 32768 --supports supports_function_calling=true

  olares-cli router model add bge-m3 --provider lmstudio --mode embedding
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			req := addModelRequest{
				Name: strings.TrimSpace(args[0]),
				Mode: strings.ToLower(strings.TrimSpace(mode)),
			}
			if c.Flags().Changed("context-size") {
				req.ContextSize = &contextSize
			}
			if c.Flags().Changed("max-output-tokens") {
				req.MaxOutputTokens = &maxOutTokens
			}
			return runModelAdd(c.Context(), f, provider, req, supports, pricing, output)
		},
	}
	cmd.Flags().StringVar(&provider, "provider", "", "the provider that will serve it (required)")
	cmd.Flags().StringVar(&mode, "mode", "chat", "endpoint family this model answers on: "+strings.Join(providerModelModes, ", "))
	cmd.Flags().IntVar(&contextSize, "context-size", 0, "context window in tokens")
	cmd.Flags().IntVar(&maxOutTokens, "max-output-tokens", 0, "upper bound on generated tokens")
	cmd.Flags().StringArrayVar(&supports, "supports", nil, "capability as key=true|false; repeatable (`router model get` prints a row's keys)")
	cmd.Flags().StringArrayVar(&pricing, "pricing", nil, "price as key=value; repeatable")
	addOutputFlag(cmd, &output)
	return cmd
}

type addModelRequest struct {
	Name            string            `json:"name"`
	Mode            string            `json:"mode"`
	Supports        map[string]any    `json:"supports,omitempty"`
	Pricing         map[string]string `json:"pricing,omitempty"`
	ContextSize     *int              `json:"context_size,omitempty"`
	MaxOutputTokens *int              `json:"max_output_tokens,omitempty"`
}

func runModelAdd(ctx context.Context, f *cmdutil.Factory, ref string, req addModelRequest, supports, pricing []string, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if req.Name == "" {
		return fmt.Errorf("model name is required")
	}
	if err := requireProviderFlag(ref, "add"); err != nil {
		return err
	}
	if !containsString(providerModelModes, req.Mode) {
		return fmt.Errorf("--mode must be one of %s, not %q", strings.Join(providerModelModes, ", "), req.Mode)
	}
	req.Supports, err = parseSupportsFlags(supports)
	if err != nil {
		return err
	}
	req.Pricing, err = parseKeyValueFlags(pricing, "--pricing")
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
	if found.isMarketSourced() {
		return marketOwnedErr(found, "attach models to")
	}
	var created providerModelRow
	path := epProviderCustomizableModels(found.ID)
	if err := pc.router.doJSON(ctx, "POST", path, req, &created); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, created)
	}
	_, err = fmt.Fprintf(os.Stdout, "attached %s (%s) to %s as %s\n",
		created.Name, created.Mode, found.Name, created.ID)
	return err
}

func newModelUpdateCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output       string
		provider     string
		enable       bool
		disable      bool
		status       string
		contextSize  int
		maxOutTokens int
		supports     []string
		pricing      []string
		replaceDesc  bool
	)
	cmd := &cobra.Command{
		Use:   "update <model>",
		Short: "enable, disable, or correct what a model claims",
		Long: `Change a model row.

--enable and --disable are the reversible way to control availability. A
disabled model stays configured, keeps its settings and quotas, and is simply
not offered; deleting it discards all of that.

--supports corrects what Router believes the model can do. This is the setting
most worth getting right: Router checks these flags before dispatching, so a
model whose vision flag is unset will refuse images the upstream would have
handled, and one whose flag is set wrongly will forward requests the upstream
rejects. Router keeps a key it does not recognise rather than refusing it, so a
misspelt flag is stored and never honoured; "router model get <model>" prints
the keys a row actually carries.

Only the fields you name change. Router itself replaces the model's whole
description whenever any part of it is sent, so this command reads the model
first and sends back everything it already had with your change folded in;
correcting one capability flag therefore leaves the prices and the context
window alone. Pass --replace-description to skip that and send only what you
typed, which drops everything you did not.

To retract a capability, set it to false rather than omitting it. A price or a
window can only be dropped with --replace-description.

The name and mode cannot change. Both are what the upstream is addressed by, and
a row that means something different is a different row. A second name for
callers to use is a route over the top: "olares-cli router route create <name>
--kind alias --model <provider>/<model>".

Examples:
  olares-cli router model update lmstudio/qwen3-8b --disable
  olares-cli router model update qwen3-8b --provider lmstudio \
    --supports supports_vision=true --context-size 65536
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if enable && disable {
				return errEnableAndDisable
			}
			req := updateModelRequest{}
			flags := c.Flags()
			if enable {
				v := true
				req.Enabled = &v
			}
			if disable {
				v := false
				req.Enabled = &v
			}
			if flags.Changed("status") {
				v := strings.ToLower(strings.TrimSpace(status))
				req.Status = &v
			}
			if flags.Changed("context-size") {
				req.ContextSize = &contextSize
			}
			if flags.Changed("max-output-tokens") {
				req.MaxOutputTokens = &maxOutTokens
			}
			return runModelUpdate(c.Context(), f, args[0], provider, req, supports, pricing, replaceDesc, output)
		},
	}
	cmd.Flags().StringVar(&provider, "provider", "", "which provider's copy, when the name is served by more than one")
	cmd.Flags().BoolVar(&enable, "enable", false, "offer this model to callers")
	cmd.Flags().BoolVar(&disable, "disable", false, "stop offering this model, keeping its configuration")
	cmd.Flags().StringVar(&status, "status", "", "active or disabled")
	cmd.Flags().IntVar(&contextSize, "context-size", 0, "context window in tokens")
	cmd.Flags().IntVar(&maxOutTokens, "max-output-tokens", 0, "upper bound on generated tokens")
	cmd.Flags().StringArrayVar(&supports, "supports", nil, "capability as key=true|false; repeatable, folded into what is stored")
	cmd.Flags().StringArrayVar(&pricing, "pricing", nil, "price as key=value; repeatable, folded into what is stored")
	cmd.Flags().BoolVar(&replaceDesc, "replace-description", false,
		"send only the capabilities, prices and sizes given here, discarding the rest")
	addOutputFlag(cmd, &output)
	return cmd
}

var errEnableAndDisable = errors.New("pass either --enable or --disable, not both")

type updateModelRequest struct {
	Enabled         *bool             `json:"enabled,omitempty"`
	Status          *string           `json:"status,omitempty"`
	Supports        map[string]any    `json:"supports,omitempty"`
	Pricing         map[string]string `json:"pricing,omitempty"`
	ParameterRules  json.RawMessage   `json:"parameter_rules,omitempty"`
	ContextSize     *int              `json:"context_size,omitempty"`
	MaxOutputTokens *int              `json:"max_output_tokens,omitempty"`
}

// touchesDescription reports whether the patch reaches the part of the row
// Router replaces wholesale rather than field by field.
func (r updateModelRequest) touchesDescription() bool {
	return len(r.Supports) > 0 || len(r.Pricing) > 0 ||
		r.ContextSize != nil || r.MaxOutputTokens != nil
}

// carryForward folds a patch into what the model already claims. Router's PATCH
// rebuilds the whole description from whatever spec fields the body carries, so
// sending one capability flag on its own would silently discard the prices and
// the context window. Reading the row first and sending it back complete is the
// difference between correcting a flag and erasing everything beside it.
func (r *updateModelRequest) carryForward(current *providerModelRow) {
	merged := make(map[string]any, len(current.Supports)+len(r.Supports))
	for k, v := range current.Supports {
		merged[k] = v
	}
	for k, v := range r.Supports {
		merged[k] = v
	}
	if len(merged) > 0 {
		r.Supports = merged
	}

	prices := make(map[string]string, len(current.Pricing)+len(r.Pricing))
	for k, v := range current.Pricing {
		prices[k] = v
	}
	for k, v := range r.Pricing {
		prices[k] = v
	}
	if len(prices) > 0 {
		r.Pricing = prices
	}

	if len(current.ParameterRules) > 0 {
		r.ParameterRules = current.ParameterRules
	}
	if r.ContextSize == nil && current.ContextSize > 0 {
		v := current.ContextSize
		r.ContextSize = &v
	}
	if r.MaxOutputTokens == nil && current.MaxOutputTokens > 0 {
		v := current.MaxOutputTokens
		r.MaxOutputTokens = &v
	}
}

func (r updateModelRequest) isEmpty() bool {
	return r.Enabled == nil && r.Status == nil &&
		len(r.Supports) == 0 && len(r.Pricing) == 0 &&
		r.ContextSize == nil && r.MaxOutputTokens == nil
}

func runModelUpdate(ctx context.Context, f *cmdutil.Factory, modelRef, providerRef string, req updateModelRequest, supports, pricing []string, replaceDesc bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if req.Status != nil && *req.Status != "active" && *req.Status != "disabled" {
		return fmt.Errorf("--status must be active or disabled, not %q", *req.Status)
	}
	req.Supports, err = parseSupportsFlags(supports)
	if err != nil {
		return err
	}
	req.Pricing, err = parseKeyValueFlags(pricing, "--pricing")
	if err != nil {
		return err
	}
	if req.isEmpty() {
		return fmt.Errorf("nothing to change; pass at least one of --enable, --disable, --status, --alias, --clear-alias, --context-size, --max-output-tokens, --supports, --pricing")
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	provider, model, err := resolveModelOnProvider(ctx, pc, modelRef, providerRef)
	if err != nil {
		return err
	}
	if req.touchesDescription() && !replaceDesc {
		req.carryForward(model)
	}
	var updated providerModelRow
	path := epProviderModel(provider.ID, model.ID)
	if err := pc.router.doJSON(ctx, "PATCH", path, req, &updated); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, updated)
	}
	_, err = fmt.Fprintf(os.Stdout, "%s on %s: offered=%s status=%s context=%s capabilities=%s\n",
		updated.Name, provider.Name, boolStr(updated.Enabled), nonEmpty(updated.Status),
		intOrDash(updated.ContextSize), summarizeSupports(updated.Supports))
	return err
}

func newModelRemoveCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		provider  string
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:     "remove <model>",
		Aliases: []string{"delete"},
		Short:   "detach a model from its provider",
		Long: `Detach a model from the provider serving it.

The row goes, and with it the default-model choices pointing at this model, its
per-model settings, and any quota scoped to it. A model re-attached afterwards
starts from an empty state.

To stop offering a model while keeping all of that, use "model update
--disable" instead.

Confirmation is required. --yes skips the prompt and is mandatory when stdin is
not a terminal.

Example:
  olares-cli router model remove lmstudio/qwen3-8b --yes
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runModelRemove(c.Context(), f, args[0], provider, assumeYes, output)
		},
	}
	cmd.Flags().StringVar(&provider, "provider", "", "which provider's copy, when the name is served by more than one")
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt (required when stdin is not a terminal)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runModelRemove(ctx context.Context, f *cmdutil.Factory, modelRef, providerRef string, assumeYes bool, outputRaw string) error {
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
	provider, model, err := resolveModelOnProvider(ctx, pc, modelRef, providerRef)
	if err != nil {
		return err
	}
	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Detach model %q from provider %q, dropping the defaults, settings and quotas that point at it?",
			model.Name, provider.Name),
		assumeYes); err != nil {
		return err
	}
	var res map[string]any
	path := epProviderModel(provider.ID, model.ID)
	if err := pc.router.doJSON(ctx, "DELETE", path, nil, &res); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, res)
	}
	_, err = fmt.Fprintf(os.Stdout, "detached %s from %s\n", model.Name, provider.Name)
	return err
}

// requireProviderFlag refuses the two verbs that create a row without being
// told where. Every other verb finds the provider from the model, but a model
// that does not exist yet has nothing to find it by.
func requireProviderFlag(ref, verb string) error {
	if strings.TrimSpace(ref) != "" {
		return nil
	}
	return fmt.Errorf("--provider is required to %s a model: the row does not exist yet, so there is no "+
		"name to find the provider by. `olares-cli router provider list` names them", verb)
}

// resolveModelOnProvider finds a model and the provider serving it, from a
// model reference alone when that is enough and from an explicit provider when
// it is not.
//
// The write routes are all scoped to a provider, so both halves are needed
// either way; what differs is who supplies the first one. A model reference is
// unique often enough to be the normal way in, and resolveModel reports the
// candidates rather than picking one when it is not — which is the case worth
// designing for, since every locally installed model application is a provider
// called `Olares` and two of them can carry the same model name.
func resolveModelOnProvider(ctx context.Context, pc *preparedClient, modelRef, providerRef string) (*providerRow, *providerModelRow, error) {
	if strings.TrimSpace(providerRef) != "" {
		return resolveProviderModel(ctx, pc, providerRef, modelRef)
	}
	row, err := resolveModel(ctx, pc, modelRef)
	if err != nil {
		return nil, nil, err
	}
	// By id on both halves: the reference has already been resolved to exactly
	// one row, and re-matching it by name here would reopen the ambiguity that
	// was just settled.
	return resolveProviderModel(ctx, pc, row.ProviderID, row.ProviderModelID)
}

// resolveProviderModel finds a model on a provider by name or id. The provider
// detail route is the lookup, because a model id alone does not say which
// provider it belongs to and the write routes need both.
func resolveProviderModel(ctx context.Context, pc *preparedClient, providerRef, modelRef string) (*providerRow, *providerModelRow, error) {
	found, err := resolveProvider(ctx, pc, providerRef)
	if err != nil {
		return nil, nil, err
	}
	detail, err := getProvider(ctx, pc, found.ID)
	if err != nil {
		return nil, nil, err
	}
	modelRef, err = requireRef(modelRef, "a model name or id")
	if err != nil {
		return nil, nil, err
	}
	names := make([]string, 0, len(detail.Models))
	for i := range detail.Models {
		m := &detail.Models[i]
		if m.ID == modelRef || strings.EqualFold(m.Name, modelRef) {
			row := detail.providerRow
			return &row, m, nil
		}
		names = append(names, m.Name)
	}
	sort.Strings(names)
	return nil, nil, missing{
		noun:  "model on " + found.Name,
		ref:   modelRef,
		known: names,
		have:  "it offers",
		none:  found.Name + " offers no models at all",
	}.err()
}

// parseSupportsFlags reads key=true|false pairs. Router tolerates a few other
// spellings of a boolean, but accepting them here would let a typo like
// `supports_vision=yes` pass as something other than what was meant.
func parseSupportsFlags(pairs []string) (map[string]any, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	out := make(map[string]any, len(pairs))
	for _, p := range pairs {
		key, value, ok := strings.Cut(p, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("--supports %q is not in key=true|false form", p)
		}
		b, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("--supports %s must be true or false, not %q", key, value)
		}
		out[key] = b
	}
	return out, nil
}

func parseKeyValueFlags(pairs []string, flagName string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(pairs))
	for _, p := range pairs {
		key, value, ok := strings.Cut(p, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("%s %q is not in key=value form", flagName, p)
		}
		out[key] = value
	}
	return out, nil
}

func containsString(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}
