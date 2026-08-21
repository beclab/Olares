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
