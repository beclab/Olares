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
		Short: "download-server global settings (aria2 / yt-dlp / seeding)",
		Long: `Read or update download-server global settings such as aria2 and
yt-dlp concurrency and seeding limits (GET/PUT /api/system/settings).

These are server-wide, not per-user, so changing them may require
administrator privileges — the CLI does not pre-check this and defers to
the server, which returns an error if the caller is not allowed.`,
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
	fmt.Fprintf(w, "%-26s %d\n", "aria2_max_concurrent:", s.Aria2MaxConcurrent)
	fmt.Fprintf(w, "%-26s %d\n", "aria2_max_conn_per_server:", s.Aria2MaxConnPerServer)
	fmt.Fprintf(w, "%-26s %d\n", "aria2_split:", s.Aria2Split)
	fmt.Fprintf(w, "%-26s %d\n", "ytdlp_concurrent:", s.YtdlpConcurrent)
	fmt.Fprintf(w, "%-26s %g\n", "seed_ratio_limit:", s.SeedRatioLimit)
	fmt.Fprintf(w, "%-26s %d\n", "seed_time_limit:", s.SeedTimeLimit)
	if s.UpdatedAt > 0 {
		fmt.Fprintf(w, "%-26s %s\n", "updated_at:", formatUnix(s.UpdatedAt))
	}
	return nil
}

func newSettingsSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		aria2MaxConcurrent    int
		aria2MaxConnPerServer int
		aria2Split            int
		ytdlpConcurrent       int
		seedRatioLimit        float64
		seedTimeLimit         int64
		output                string
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "update one or more download-server global settings",
		Long: `Update download-server global settings (PUT /api/system/settings).

Only the flags you explicitly pass are sent; unset fields are left
unchanged. Provide at least one flag.`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			changed := map[string]bool{
				"aria2-max-concurrent":      c.Flags().Changed("aria2-max-concurrent"),
				"aria2-max-conn-per-server": c.Flags().Changed("aria2-max-conn-per-server"),
				"aria2-split":               c.Flags().Changed("aria2-split"),
				"ytdlp-concurrent":          c.Flags().Changed("ytdlp-concurrent"),
				"seed-ratio-limit":          c.Flags().Changed("seed-ratio-limit"),
				"seed-time-limit":           c.Flags().Changed("seed-time-limit"),
			}
			req, err := buildSettingsPatch(changed, aria2MaxConcurrent, aria2MaxConnPerServer, aria2Split, ytdlpConcurrent, seedRatioLimit, seedTimeLimit)
			if err != nil {
				return err
			}
			return runSettingsSet(c.Context(), f, req, output)
		},
	}
	addOutputFlag(cmd, &output)
	cmd.Flags().IntVar(&aria2MaxConcurrent, "aria2-max-concurrent", 0, "max concurrent aria2 downloads")
	cmd.Flags().IntVar(&aria2MaxConnPerServer, "aria2-max-conn-per-server", 0, "max aria2 connections per server")
	cmd.Flags().IntVar(&aria2Split, "aria2-split", 0, "aria2 split (segments per download)")
	cmd.Flags().IntVar(&ytdlpConcurrent, "ytdlp-concurrent", 0, "max concurrent yt-dlp downloads")
	cmd.Flags().Float64Var(&seedRatioLimit, "seed-ratio-limit", 0, "seeding share ratio limit")
	cmd.Flags().Int64Var(&seedTimeLimit, "seed-time-limit", 0, "seeding time limit in seconds")
	return cmd
}

// buildSettingsPatch assembles the partial update body from the set of flags
// the user actually changed, so unset fields stay nil (unchanged). It errors
// when nothing was provided.
func buildSettingsPatch(changed map[string]bool, aria2MaxConcurrent, aria2MaxConnPerServer, aria2Split, ytdlpConcurrent int, seedRatioLimit float64, seedTimeLimit int64) (UpdateSystemSettingsReq, error) {
	var req UpdateSystemSettingsReq
	any := false
	if changed["aria2-max-concurrent"] {
		v := aria2MaxConcurrent
		req.Aria2MaxConcurrent = &v
		any = true
	}
	if changed["aria2-max-conn-per-server"] {
		v := aria2MaxConnPerServer
		req.Aria2MaxConnPerServer = &v
		any = true
	}
	if changed["aria2-split"] {
		v := aria2Split
		req.Aria2Split = &v
		any = true
	}
	if changed["ytdlp-concurrent"] {
		v := ytdlpConcurrent
		req.YtdlpConcurrent = &v
		any = true
	}
	if changed["seed-ratio-limit"] {
		v := seedRatioLimit
		req.SeedRatioLimit = &v
		any = true
	}
	if changed["seed-time-limit"] {
		v := seedTimeLimit
		req.SeedTimeLimit = &v
		any = true
	}
	if !any {
		return UpdateSystemSettingsReq{}, fmt.Errorf("nothing to set: provide at least one --... flag")
	}
	return req, nil
}

func runSettingsSet(ctx context.Context, f *cmdutil.Factory, req UpdateSystemSettingsReq, outputRaw string) error {
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
