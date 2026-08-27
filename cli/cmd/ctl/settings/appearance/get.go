package appearance

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// localeConfigPath is user-service's /api/wallpaper/config/system, which
// forwards /bfl/settings/v1alpha1/config-system. The body is BFL's
// PostLocale (the same struct serves GET and POST):
//
//	{ "language": "en-US", "location": "...", "timezone": "UTC", "theme": "light" }
const localeConfigPath = "/api/wallpaper/config/system"

// `olares-cli settings appearance get`
//
// The whole Appearance page in one read, which costs three upstream calls
// because user-service keeps locale, widget preferences and wallpaper in
// three separate endpoints.
func NewGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "show locale, widget preferences and wallpaper",
		Long: `Show the values the Settings -> Appearance page reads: the
localization fields (language, location, timezone), the desktop widget
preferences, and the wallpaper selection.

The theme is left out: it is a browser cookie the SPA never sends
upstream, so the value this backend holds is not what anything displays.

This is the only read verb in the subtree — there is no separate
"widget get" or "wallpaper get".

With --output json the sections arrive as the top-level keys "locale",
"widget" and "wallpaper", each holding the upstream field names verbatim.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runGet(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// localeConfig mirrors the part of BFL's PostLocale this page shows.
// Timezone is read from the per-user UserEnv CRs by BFL's
// HandleGetSysConfig; location still lives on a user annotation. The
// upstream also carries a theme, left out for the reason in theme.go.
type localeConfig struct {
	Language string `json:"language"`
	Location string `json:"location"`
	Timezone string `json:"timezone"`
}

// appearanceView holds a nil section for one this backend is too old to
// serve, so a 1.12.5 reader gets locale and wallpaper instead of an error
// about the widget API it does not have.
type appearanceView struct {
	Locale    *localeConfig      `json:"locale"`
	Widget    *widgetPreferences `json:"widget"`
	Wallpaper *wallpaperConfig   `json:"wallpaper"`
}

// appearanceSection is one upstream read feeding one rendered section.
type appearanceSection struct {
	name  string
	path  string
	out   interface{}
	apply func()
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

	// Locale and wallpaper exist on every supported backend; the widget
	// section is 1.12.6+, so on an older one it is reported as gated
	// rather than failing the whole page.
	widgetSupported, err := f.OlaresBackendAtLeast(ctx, appearanceMinOlaresVersion)
	if err != nil {
		return err
	}

	var view appearanceView
	locale, widget, wallpaper := new(localeConfig), new(widgetPreferences), new(wallpaperConfig)
	sections := []appearanceSection{
		{"locale", localeConfigPath, locale, func() { view.Locale = locale }},
		{"wallpaper", wallpaperPath, wallpaper, func() { view.Wallpaper = wallpaper }},
	}
	if widgetSupported {
		sections = append(sections,
			appearanceSection{"widget preferences", widgetPath, widget, func() { view.Widget = widget }})
	}
	for _, s := range sections {
		if err := doGetEnvelope(ctx, pc.doer, s.path, s.out); err != nil {
			return fmt.Errorf("read %s: %w", s.name, err)
		}
		s.apply()
	}

	if format == FormatJSON {
		return printJSON(os.Stdout, view)
	}
	return renderAppearance(os.Stdout, view)
}

func renderAppearance(w io.Writer, v appearanceView) error {
	sections := []struct {
		title string
		rows  [][2]string
	}{
		{"Locale", nil},
		{"Widget", nil},
		{"Wallpaper", nil},
	}
	if v.Locale != nil {
		sections[0].rows = [][2]string{
			{"Language", describeLocale(v.Locale.Language)},
			{"Location", nonEmpty(v.Locale.Location)},
			{"Timezone", nonEmpty(v.Locale.Timezone)},
		}
	}
	if v.Widget != nil {
		sections[1].rows = [][2]string{
			{"Show widgets", boolStr(v.Widget.ShowWidgets)},
			{"24-hour format", boolStr(v.Widget.Is24HourFormat)},
			{"Date format", describeDateFormat(v.Widget.DateFormat, time.Now())},
			{"Show dashboard", boolStr(v.Widget.ShowDashboard)},
		}
	}
	if v.Wallpaper != nil {
		// Named as the user selects them: a built-in by the number
		// `wallpaper set` takes, a fill mode by its Settings label.
		sections[2].rows = [][2]string{
			{"Desktop", describeWallpaperValue(v.Wallpaper.Desktop)},
			{"Desktop style", contentModeLabel(nonEmpty(v.Wallpaper.DesktopStyle))},
			{"Login", describeWallpaperValue(v.Wallpaper.Login)},
			{"Login style", contentModeLabel(nonEmpty(v.Wallpaper.LoginStyle))},
		}
	}
	for i, s := range sections {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if s.rows == nil {
			if _, err := fmt.Fprintf(w, "%s\n  requires Olares >= %s\n", s.title, appearanceMinOlaresVersion); err != nil {
				return err
			}
			continue
		}
		if err := renderSection(w, s.title, s.rows); err != nil {
			return err
		}
	}

	if v.Wallpaper == nil {
		return nil
	}
	// The uploaded lists are what `wallpaper delete` takes as its <bg>
	// argument, so print them in full rather than as a count.
	for _, l := range []struct {
		title string
		items []string
	}{
		{"Uploaded desktop backgrounds", v.Wallpaper.UploadDesktopBackgrounds},
		{"Uploaded login backgrounds", v.Wallpaper.UploadLoginBackgrounds},
	} {
		if _, err := fmt.Fprintf(w, "\n%s\n", l.title); err != nil {
			return err
		}
		if len(l.items) == 0 {
			if _, err := fmt.Fprintln(w, "  none"); err != nil {
				return err
			}
			continue
		}
		for _, item := range l.items {
			if _, err := fmt.Fprintf(w, "  %s\n", item); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderSection(w io.Writer, title string, rows [][2]string) error {
	if _, err := fmt.Fprintf(w, "%s\n", title); err != nil {
		return err
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, r := range rows {
		if _, err := fmt.Fprintf(tw, "  %s\t%s\n", r[0]+":", r[1]); err != nil {
			return err
		}
	}
	return tw.Flush()
}
