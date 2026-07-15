package download

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewPrefsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prefs",
		Short: "user download preferences (yt-dlp quality)",
		Long:  `Read or write per-(user, app) yt-dlp quality preferences.`,
	}
	cmd.AddCommand(newPrefsGetCommand(f))
	cmd.AddCommand(newPrefsSetCommand(f))
	return cmd
}

func newPrefsGetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		app    string
		output string
	)
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get yt-dlp quality preference for an app",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runPrefsGet(c.Context(), f, app, output)
		},
	}
	addAppFlag(cmd, &app)
	addOutputFlag(cmd, &output)
	return cmd
}

func newPrefsSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		app     string
		quality string
		output  string
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "set yt-dlp quality preference for an app",
		Long:  `PUT /api/user/preferences. --quality must be one of: best, 2160p, 1080p, 720p, 480p, 360p, audio.`,
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runPrefsSet(c.Context(), f, app, quality, output)
		},
	}
	addAppFlag(cmd, &app)
	addOutputFlag(cmd, &output)
	cmd.Flags().StringVar(&quality, "quality", "", "yt-dlp quality preset (required)")
	_ = cmd.MarkFlagRequired("quality")
	return cmd
}

func runPrefsGet(ctx context.Context, f *cmdutil.Factory, app, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	app = strings.TrimSpace(app)
	if app == "" {
		app = defaultApp
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("app", app)
	var pref UserPreference
	if err := doGet(ctx, pc.doer, "/api/user/preferences"+encodeQuery(q), &pref); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, pref)
	default:
		fmt.Printf("App:      %s\n", orDash(pref.App))
		fmt.Printf("Quality:  %s\n", orDash(pref.YtdlpQuality))
		if !pref.UpdatedAt.IsZero() {
			fmt.Printf("Updated:  %s\n", formatTime(pref.UpdatedAt))
		}
		return nil
	}
}

func runPrefsSet(ctx context.Context, f *cmdutil.Factory, app, quality, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	app = strings.TrimSpace(app)
	if app == "" {
		app = defaultApp
	}
	quality = strings.TrimSpace(quality)
	if quality == "" {
		return fmt.Errorf("--quality is required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	body := map[string]string{
		"app":           app,
		"ytdlp_quality": quality,
	}
	var pref UserPreference
	if err := doMutate(ctx, pc.doer, "PUT", "/api/user/preferences", body, &pref); err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, pref)
	default:
		fmt.Printf("set app=%s quality=%s\n", pref.App, pref.YtdlpQuality)
		return nil
	}
}
