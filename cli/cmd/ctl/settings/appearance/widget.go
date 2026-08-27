package appearance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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

// dateFormatLayout translates the SPA's pattern vocabulary into a Go
// layout. The replacer is single-pass and tries these in order, so YYYY
// is never read as two YY and no substitution rewrites another's output.
var dateFormatLayout = strings.NewReplacer(
	"YYYY", "2006", "YY", "06", "MM", "01", "DD", "02", "M", "1", "D", "2",
)

// renderDateFormat shows what a pattern produces. Settings pairs its
// picker with a live preview of the current date, and without one
// YY-M-D and YY.M.D can only be told apart by setting both.
func renderDateFormat(pattern string, at time.Time) string {
	return at.Format(dateFormatLayout.Replace(pattern))
}

// describeDateFormat is how a stored pattern reads in the table: the
// pattern the page shows, plus what the clock currently makes of it.
func describeDateFormat(pattern string, at time.Time) string {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nonEmpty(pattern)
	}
	return fmt.Sprintf("%s (%s)", pattern, renderDateFormat(pattern, at))
}

// dateFormatChoices lists every pattern beside that preview, two per
// line to keep the help short enough to read.
func dateFormatChoices(at time.Time) string {
	var b strings.Builder
	for i := 0; i < len(dateFormats); i += 2 {
		b.WriteString(fmt.Sprintf("  %-11s %-11s", dateFormats[i], renderDateFormat(dateFormats[i], at)))
		if i+1 < len(dateFormats) {
			b.WriteString(fmt.Sprintf("  %-11s %s", dateFormats[i+1], renderDateFormat(dateFormats[i+1], at)))
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), " \n")
}

func NewWidgetCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "widget",
		Short: "desktop widget preferences",
		Long: `Update the desktop widget preferences (Settings -> Appearance >
Widget and Date & Time).

Current values are shown by "settings appearance get".

Subcommands:
  set --show-widgets=<true|false>    show the desktop widgets
  set --24-hour=<true|false>         24-hour time in the clock widget
  set --date-format YYYY/MM/DD       date format in the clock widget,
                                     one of 11 patterns ("set --help")
  set --show-dashboard=<true|false>  show the dashboard widget

One "set" takes any combination of these, and a preference you leave
out keeps its current value. The desktop shows the last three only
while --show-widgets is true.
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

The boolean flags need the equals sign: "--24-hour=false" turns the
setting off, while "--24-hour false" turns it on and then fails on the
leftover word.

--date-format patterns, each shown as today's date:
%s

Case does not matter: "yyyy/mm/dd" sets YYYY/MM/DD.

The desktop shows the clock and the dashboard only while
--show-widgets is true, though the other preferences are still stored
while it is false.

Examples:
  olares-cli settings appearance widget set --show-widgets=false
  olares-cli settings appearance widget set --24-hour=false --date-format M/D/YY
  olares-cli settings appearance widget set --show-dashboard=true
`, dateFormatChoices(time.Now())),
		Args: rejectBooleanAsArgument,
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
	// Declared false so the help does not print "(default true)": omitting
	// a flag keeps the current value rather than setting it to anything.
	cmd.Flags().BoolVar(&showWidgets, "show-widgets", false, "show the desktop widgets (unchanged if omitted)")
	cmd.Flags().BoolVar(&is24HourFormat, "24-hour", false, "use 24-hour time in the clock widget (unchanged if omitted)")
	cmd.Flags().StringVar(&dateFormat, "date-format", "",
		"one of the 11 `YYYY/MM/DD`-style patterns listed above (unchanged if omitted)")
	cmd.Flags().BoolVar(&showDashboard, "show-dashboard", false, "show the dashboard widget (unchanged if omitted)")
	return cmd
}

// A boolean flag consumes no following word, so "--24-hour false" sets the
// preference to true and leaves "false" as a positional, which NoArgs alone
// would report as an unknown command.
func rejectBooleanAsArgument(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		if value := strings.ToLower(args[0]); value == "true" || value == "false" {
			return fmt.Errorf("%q was read as an argument, not as a flag value: a boolean flag needs the equals sign, as in --%s=%s",
				args[0], exampleBooleanFlag(cmd), value)
		}
	}
	return cobra.NoArgs(cmd, args)
}

// The flag to put in that example: the boolean the caller just passed, when
// there is only one, so the suggestion is the line they meant to type.
func exampleBooleanFlag(cmd *cobra.Command) string {
	var passed []string
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if f.Value.Type() == "bool" {
			passed = append(passed, f.Name)
		}
	})
	if len(passed) == 1 {
		return passed[0]
	}
	return "show-widgets"
}

// resolveDateFormat accepts a pattern in any case and returns the spelling
// the desktop clock renders, matching how this subtree reads surfaces and
// fill modes. A pattern a newer Settings adds needs a newer CLI to preview
// it, so there is no flag to bypass the list.
func resolveDateFormat(raw string) (string, error) {
	value := strings.ToUpper(strings.TrimSpace(raw))
	if value == "" {
		return "", fmt.Errorf("--date-format requires one of these patterns (shown as today's date):\n%s",
			dateFormatChoices(time.Now()))
	}
	for _, f := range dateFormats {
		if value == f {
			return f, nil
		}
	}
	return "", fmt.Errorf("unsupported --date-format %q; allowed patterns (shown as today's date):\n%s",
		raw, dateFormatChoices(time.Now()))
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
