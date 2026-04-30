package appearance

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings appearance get`
//
// Backed by user-service's /api/wallpaper/config/system, which forwards
// /bfl/settings/v1alpha1/config-system. The body is BFL's PostLocale
// (it's the same struct used for both GET and POST):
//
//   { "language": "en-US", "location": "..." }
//
// Other appearance bits (login background, desktop style, theme picker,
// wallpaper image upload) are intentionally browser-bound and stay out
// of CLI scope — they're either visual blob workflows or only meaningful
// inside an active SPA session.
func NewGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "show language + locale",
		Long: `Show the language + location values used for Olares localization
(the same fields the Settings -> Appearance > Language picker writes).
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runGet(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

type appearanceConfig struct {
	Language string `json:"language"`
	Location string `json:"location"`
}

func runGet(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var cfg appearanceConfig
	if err := doGetEnvelope(ctx, pc.doer, "/api/wallpaper/config/system", &cfg); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, cfg)
	default:
		return renderAppearance(os.Stdout, cfg)
	}
}

func renderAppearance(w io.Writer, c appearanceConfig) error {
	rows := [][2]string{
		{"Language", nonEmpty(c.Language)},
		{"Location", nonEmpty(c.Location)},
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-10s %s\n", r[0]+":", r[1]); err != nil {
			return err
		}
	}
	return nil
}
