package router

// Paths that moved, kept answering for one release.
//
// Every command here is hidden and delegates to whatever replaced it, so there
// is one implementation and not two drifting copies. Cobra prints the
// `Deprecated` line before running, which is the only difference a caller sees.
// The file is meant to be deleted whole rather than pruned entry by entry.

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// newDeprecatedDefaultCommand is `router default`, which turned out to be three
// ways of saying `router route`. A category is a route whose kind Router owns
// rather than a noun of its own: the old subtree read the same rows through the
// same filter and wrote the same PATCH, so it was a `--kind` flag wearing a
// verb, and the price was a help text on each side explaining it was not the
// other one.
// newDeprecatedLocalCommand is `router local`, which became verbs on `router
// model`. It was named after the plane it reached rather than the thing it
// reached about, so a user with a question about a model had to know that the
// answer lived behind an application, and know the application's Olares app id
// to ask. Both are now the model's own verbs, and the app id survives as
// `--app` — no longer the only way in, and still the only way in when Router
// cannot be asked.
//
// These keep the old shape, where the positional was always the application.
func newDeprecatedLocalCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "local",
		Short:  "moved onto `router model`",
		Hidden: true,
	}
	cmd.SilenceUsage = true

	moved := []struct {
		cmd *cobra.Command
		to  string
	}{
		{newModelStatusCommand(f, addressByApp), "model status <model>"},
		{newModelProgressCommand(f, addressByApp), "model progress <model>"},
		{newModelRetryCommand(f, addressByApp), "model retry <model>"},
		{newModelRestartCommand(f, addressByApp), "model restart <model>"},
		{newModelGPUCommand(f, addressByApp), "model diag gpu <model>"},
		{newModelPerfCommand(f, addressByApp), "model diag perf <model>"},
		{newModelConfigCommand(f, addressByApp), "model diag config <model>"},
		{newModelEndpointsCommand(f, addressByApp), "model diag endpoints <model>"},
	}
	for _, m := range moved {
		m.cmd.Hidden = true
		m.cmd.Deprecated = "use `olares-cli router " + m.to + "` instead, or `--app <app_id>` to keep " +
			"naming the application."
		cmd.AddCommand(m.cmd)
	}
	cmd.AddCommand(newDeprecatedLocalSpecCommand(f))
	return cmd
}

func newDeprecatedLocalSpecCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "spec",
		Short:  "moved to `router model spec`",
		Hidden: true,
	}
	cmd.SilenceUsage = true
	moved := []struct {
		cmd *cobra.Command
		to  string
	}{
		{newModelSpecShowCommand(f, addressByApp), "model spec show <model>"},
		{newModelSpecFileCommand(f, addressByApp), "model spec file <model>"},
		{newModelSpecSetCommand(f, addressByApp), "model spec set <model>"},
	}
	for _, m := range moved {
		m.cmd.Hidden = true
		m.cmd.Deprecated = "use `olares-cli router " + m.to + "` instead, or `--app <app_id>` to keep " +
			"naming the application."
		cmd.AddCommand(m.cmd)
	}
	return cmd
}

// newDeprecatedSpecCommand is `router spec`, which became `router model spec`
// and `router model restart`. A card is a property of a model rather than a
// noun beside it, and standing at the top level it read as a third thing to
// know about alongside `list` and `local`.
func newDeprecatedSpecCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "spec",
		Short:  "moved to `router model spec`",
		Hidden: true,
	}
	cmd.SilenceUsage = true

	show := newModelSpecShowCommand(f, addressByModel)
	show.Hidden = true
	show.Deprecated = "use `olares-cli router model spec show <model>` instead."
	cmd.AddCommand(show)

	edit := newSpecEditCommand(f)
	edit.Hidden = true
	edit.Deprecated = "use `olares-cli router model spec edit <model>` instead."
	cmd.AddCommand(edit)

	restart := newModelRestartCommand(f, addressByModel)
	restart.Hidden = true
	restart.Deprecated = "use `olares-cli router model restart <model>` instead."
	cmd.AddCommand(restart)

	return cmd
}

// newDeprecatedListCommand is `router list`, which became `router model list`.
// It stood at the top level as though listing models were a different activity
// from managing them, so the noun a reader wanted was reached two ways
// depending on which verb they needed next.
func newDeprecatedListCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := newModelListCommand(f)
	cmd.Hidden = true
	cmd.Short = "moved to `router model list`"
	cmd.Deprecated = "use `olares-cli router model list` instead."
	return cmd
}

// newDeprecatedModelsCommand is `router models`, which became `router call
// models`. Sitting beside `router list` it read as a second listing of the same
// thing, when what it answers is a question about a call: these are the names
// the credential `router call` uses may put in `--model`. It is served by the
// data plane, so it belongs with the other verbs that go there.
func newDeprecatedModelsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := newCallModelsCommand(f)
	cmd.Hidden = true
	cmd.Short = "moved to `router call models`"
	cmd.Deprecated = "use `olares-cli router call models` instead."
	return cmd
}

// newDeprecatedKeyLocalCommand is `router key local`. Every other Olares tree
// spells "local" as "on this machine, not the platform", and this one meant
// "the one in use", which is not the same claim.
func newDeprecatedKeyLocalCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := newKeyCurrentCommand(f)
	cmd.Use = "local"
	cmd.Hidden = true
	cmd.Short = "moved to `router key current`"
	cmd.Deprecated = "use `olares-cli router key current` instead."
	return cmd
}

func newDeprecatedDefaultCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "default",
		Short:  "moved to `router route`",
		Hidden: true,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newDeprecatedDefaultShowCommand(f))
	cmd.AddCommand(newDeprecatedDefaultToggleCommand(f, true))
	cmd.AddCommand(newDeprecatedDefaultToggleCommand(f, false))
	return cmd
}

func newDeprecatedDefaultShowCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:        "show",
		Short:      "moved to `router route list --kind default`",
		Deprecated: "use `olares-cli router route list --kind default` instead.",
		Hidden:     true,
		Args:       cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runRouteList(c.Context(), f, routeKindDefault, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// newDeprecatedProviderModelsCommand is `router provider models`, which became
// `router model`. Reaching a model through the provider serving it made naming
// the provider compulsory for every verb, including the four that could have
// worked it out from the model — and it split the model noun across two
// subtrees, since listing models was never under here at all.
//
// These keep the old two-argument shape, provider first.
func newDeprecatedProviderModelsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "models",
		Short:  "moved to `router model`",
		Hidden: true,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newDeprecatedProviderModelsGetCommand(f))
	cmd.AddCommand(newDeprecatedProviderModelsImportCommand(f))
	cmd.AddCommand(newDeprecatedProviderModelsAddCommand(f))
	cmd.AddCommand(newDeprecatedProviderModelsUpdateCommand(f))
	cmd.AddCommand(newDeprecatedProviderModelsDeleteCommand(f))
	return cmd
}

func newDeprecatedProviderModelsGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:        "get <provider> <model>",
		Short:      "moved to `router model get`",
		Deprecated: "use `olares-cli router model get <model>` instead.",
		Hidden:     true,
		Args:       cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runModelGet(c.Context(), f, args[1], args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func newDeprecatedProviderModelsImportCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:        "import <provider> <model>...",
		Short:      "moved to `router model import`",
		Deprecated: "use `olares-cli router model import --provider <provider> <model>...` instead.",
		Hidden:     true,
		Args:       cobra.MinimumNArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runModelImport(c.Context(), f, args[0], args[1:], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func newDeprecatedProviderModelsAddCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output       string
		mode         string
		contextSize  int
		maxOutTokens int
		supports     []string
		pricing      []string
	)
	cmd := &cobra.Command{
		Use:        "add <provider> <model-name>",
		Short:      "moved to `router model add`",
		Deprecated: "use `olares-cli router model add <model> --provider <provider>` instead.",
		Hidden:     true,
		Args:       cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			req := addModelRequest{
				Name: strings.TrimSpace(args[1]),
				Mode: strings.ToLower(strings.TrimSpace(mode)),
			}
			if c.Flags().Changed("context-size") {
				req.ContextSize = &contextSize
			}
			if c.Flags().Changed("max-output-tokens") {
				req.MaxOutputTokens = &maxOutTokens
			}
			return runModelAdd(c.Context(), f, args[0], req, supports, pricing, output)
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "chat", "endpoint family this model answers on")
	cmd.Flags().IntVar(&contextSize, "context-size", 0, "context window in tokens")
	cmd.Flags().IntVar(&maxOutTokens, "max-output-tokens", 0, "upper bound on generated tokens")
	cmd.Flags().StringArrayVar(&supports, "supports", nil, "capability as key=true|false; repeatable")
	cmd.Flags().StringArrayVar(&pricing, "pricing", nil, "price as key=value; repeatable")
	addOutputFlag(cmd, &output)
	return cmd
}

func newDeprecatedProviderModelsUpdateCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output       string
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
		Use:        "update <provider> <model>",
		Short:      "moved to `router model update`",
		Deprecated: "use `olares-cli router model update <model>` instead.",
		Hidden:     true,
		Args:       cobra.ExactArgs(2),
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
			return runModelUpdate(c.Context(), f, args[1], args[0], req, supports, pricing, replaceDesc, output)
		},
	}
	cmd.Flags().BoolVar(&enable, "enable", false, "offer this model to callers")
	cmd.Flags().BoolVar(&disable, "disable", false, "stop offering this model, keeping its configuration")
	cmd.Flags().StringVar(&status, "status", "", "active or disabled")
	cmd.Flags().IntVar(&contextSize, "context-size", 0, "context window in tokens")
	cmd.Flags().IntVar(&maxOutTokens, "max-output-tokens", 0, "upper bound on generated tokens")
	cmd.Flags().StringArrayVar(&supports, "supports", nil, "capability as key=true|false; repeatable")
	cmd.Flags().StringArrayVar(&pricing, "pricing", nil, "price as key=value; repeatable")
	cmd.Flags().BoolVar(&replaceDesc, "replace-description", false,
		"send only the capabilities, prices and sizes given here, discarding the rest")
	addOutputFlag(cmd, &output)
	return cmd
}

func newDeprecatedProviderModelsDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:        "delete <provider> <model>",
		Short:      "moved to `router model remove`",
		Deprecated: "use `olares-cli router model remove <model>` instead.",
		Hidden:     true,
		Args:       cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runModelRemove(c.Context(), f, args[1], args[0], assumeYes, output)
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt")
	addOutputFlag(cmd, &output)
	return cmd
}

func newDeprecatedDefaultToggleCommand(f *cmdutil.Factory, on bool) *cobra.Command {
	var output string
	verb := "disable"
	if on {
		verb = "enable"
	}
	cmd := &cobra.Command{
		Use:        verb + " <category>",
		Short:      "moved to `router route " + verb + "`",
		Deprecated: "use `olares-cli router route " + verb + " <category>` instead.",
		Hidden:     true,
		Args:       cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			enabled := on
			return runRoutePatch(c.Context(), f, args[0], "", &enabled, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}
