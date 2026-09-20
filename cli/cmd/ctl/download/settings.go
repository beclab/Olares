package download

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewSettingsCommand assembles `olares-cli knowledge download settings`.
func NewSettingsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "download-server global settings (aria2 max concurrency)",
		Long: `Read or update global download settings.

Currently, aria2_max_concurrent controls the number of concurrent aria2
downloads. This setting applies to the whole server, not just the active
user, and changing it may require administrator privileges.`,
	}
	cmd.AddCommand(newSettingsGetCommand(f))
	cmd.AddCommand(newSettingsSetCommand(f))
	return cmd
}

func newSettingsGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "show the current download-server global settings",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runSettingsGet(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runSettingsGet(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	var s SystemSettings
	if err := doGet(ctx, pc.doer, "/api/system/settings", &s); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, s)
	default:
		return renderSystemSettings(os.Stdout, s)
	}
}

func renderSystemSettings(w io.Writer, s SystemSettings) error {
	fmt.Fprintf(w, "%-22s %d\n", "aria2_max_concurrent:", s.Aria2MaxConcurrent)
	return nil
}

func newSettingsSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		aria2MaxConcurrent int
		output             string
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "update a download-server global setting",
		Long: `Set the maximum number of concurrent aria2 downloads.

The accepted range is 1 through 16. The updated setting is printed on
success.`,
		Example: `  olares-cli knowledge download settings set --aria2-max-concurrent 4`,
		Args:    cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			if !c.Flags().Changed("aria2-max-concurrent") {
				return fmt.Errorf("nothing to set: provide --aria2-max-concurrent")
			}
			req := SystemSettingUpdateReq{
				Key:   systemSettingAria2MaxConcurrent,
				Value: aria2MaxConcurrent,
			}
			return runSettingsSet(c.Context(), f, req, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().IntVar(&aria2MaxConcurrent, "aria2-max-concurrent", 0, "max concurrent aria2 downloads (range [1, 16])")
	return cmd
}

func runSettingsSet(ctx context.Context, f *cmdutil.Factory, req SystemSettingUpdateReq, outputRaw string) error {
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
	var s SystemSettings
	if err := doMutate(ctx, pc.doer, "PUT", "/api/system/settings", req, &s); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, s)
	default:
		return renderSystemSettings(os.Stdout, s)
	}
}
