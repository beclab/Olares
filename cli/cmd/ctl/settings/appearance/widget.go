package appearance

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings appearance widget set ...`
//
// Backed by user-service's widget.controller.ts:
//
//	GET  /api/widget  -> WidgetPreferences
//	POST /api/widget  <- { "widget": WidgetPreferences }
//
// The POST replaces the whole preferences object, so a partial change is
// a read-modify-write. Current values are shown by `appearance get`.
const widgetPath = "/api/widget"

// widgetPreferences mirrors the SPA's WidgetPreferences
// (apps/.../stores/settings/widgetPreferences.ts:19-24). The JSON tags
// keep the upstream spelling — "showWeight" is the wire name for the
// widget master switch, despite what it reads like.
type widgetPreferences struct {
	ShowWidgets    bool   `json:"showWeight"`
	Is24HourFormat bool   `json:"is24HourFormat"`
	DateFormat     string `json:"dateFormat"`
	ShowDashboard  bool   `json:"showDashboard"`
}

// dateFormats mirrors the DateFormat union in
// apps/.../stores/settings/widgetPreferences.ts:6-17. The upstream stores
// whatever string it is handed, so an unlisted value would be accepted
// and then fail to render in the desktop clock.
var dateFormats = []string{
	"YYYY/MM/DD",
	"D/M/YY",
	"M/D/YY",
	"DD/MM/YYYY",
	"DD.MM.YYYY",
	"DD-MM-YYYY",
	"YYYY.MM.DD",
	"YYYY-MM-DD",
	"YY/MM/DD",
	"YY-M-D",
	"YY.M.D",
}

func NewWidgetCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "widget",
		Short: "desktop widget preferences",
		Long: `Update the desktop widget preferences (Settings -> Appearance >
Widgets and Date & Time).

Current values are shown by "settings appearance get".

Subcommands:
  set [flags]           update one or more preferences
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newWidgetSetCommand(f))
	return cmd
}

// widgetPatch carries only the preferences the caller actually named.
// The upstream POST replaces the whole object, so anything left nil has
// to be sent back exactly as it was read.
type widgetPatch struct {
	showWidgets    *bool
	is24HourFormat *bool
	dateFormat     *string
	showDashboard  *bool
}

func (p widgetPatch) empty() bool {
	return p.showWidgets == nil && p.is24HourFormat == nil &&
		p.dateFormat == nil && p.showDashboard == nil
}

func (p widgetPatch) apply(cur widgetPreferences) widgetPreferences {
	if p.showWidgets != nil {
		cur.ShowWidgets = *p.showWidgets
	}
	if p.is24HourFormat != nil {
		cur.Is24HourFormat = *p.is24HourFormat
	}
	if p.dateFormat != nil {
		cur.DateFormat = *p.dateFormat
	}
	if p.showDashboard != nil {
		cur.ShowDashboard = *p.showDashboard
	}
	return cur
}

func newWidgetSetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		showWidgets    bool
		is24HourFormat bool
		dateFormat     string
		showDashboard  bool
	)
	cmd := &cobra.Command{
		Use:   "set [flags]",
		Short: "update one or more desktop widget preferences",
		Long: fmt.Sprintf(`Update one or more desktop widget preferences. Only the flags you
pass are changed; the rest are read from the upstream and sent back
unchanged, because the endpoint replaces the whole object.

Allowed --date-format values:
  %s

Examples:
  olares-cli settings appearance widget set --show-widgets=false
  olares-cli settings appearance widget set --24-hour=false --date-format M/D/YY
  olares-cli settings appearance widget set --show-dashboard=true
`, strings.Join(dateFormats, "\n  ")),
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			patch := widgetPatch{}
			if c.Flags().Changed("show-widgets") {
				patch.showWidgets = &showWidgets
			}
			if c.Flags().Changed("24-hour") {
				patch.is24HourFormat = &is24HourFormat
			}
			if c.Flags().Changed("show-dashboard") {
				patch.showDashboard = &showDashboard
			}
			if c.Flags().Changed("date-format") {
				resolved, err := resolveDateFormat(dateFormat)
				if err != nil {
					return err
				}
				patch.dateFormat = &resolved
			}
			return runWidgetSet(c.Context(), f, patch)
		},
	}
	cmd.Flags().BoolVar(&showWidgets, "show-widgets", true, "show the desktop widgets")
	cmd.Flags().BoolVar(&is24HourFormat, "24-hour", true, "use 24-hour time in the clock widget")
	cmd.Flags().StringVar(&dateFormat, "date-format", "", "date format used by the clock widget")
	cmd.Flags().BoolVar(&showDashboard, "show-dashboard", true, "show the dashboard widget")
	return cmd
}

func resolveDateFormat(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("--date-format requires a value (allowed: %s)", strings.Join(dateFormats, ", "))
	}
	for _, f := range dateFormats {
		if value == f {
			return value, nil
		}
	}
	return "", fmt.Errorf("unsupported --date-format %q (allowed: %s)", raw, strings.Join(dateFormats, ", "))
}

func runWidgetSet(ctx context.Context, f *cmdutil.Factory, patch widgetPatch) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if patch.empty() {
		return fmt.Errorf("widget set requires at least one of --show-widgets, --24-hour, --date-format, --show-dashboard")
	}
	if err := requireAppearanceBackendVersion(ctx, f,
		"settings appearance widget set", "the widget preferences API"); err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var current widgetPreferences
	if err := doGetEnvelope(ctx, pc.doer, widgetPath, &current); err != nil {
		return err
	}
	body := map[string]widgetPreferences{"widget": patch.apply(current)}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", widgetPath, body, nil); err != nil {
		return err
	}
	fmt.Println("Widget preferences updated.")
	return nil
}
