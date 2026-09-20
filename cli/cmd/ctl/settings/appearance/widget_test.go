package appearance

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

func TestResolveDateFormat(t *testing.T) {
	for _, allowed := range dateFormats {
		got, err := resolveDateFormat("  " + allowed + "  ")
		if err != nil {
			t.Fatalf("resolveDateFormat(%q) errored: %v", allowed, err)
		}
		if got != allowed {
			t.Errorf("resolveDateFormat(%q) = %q; want %q", allowed, got, allowed)
		}
	}

	// Case-insensitive like this subtree's surfaces and fill modes, and
	// canonicalized because the clock renders the stored spelling.
	for _, raw := range []string{"yyyy/mm/dd", "Yyyy/Mm/Dd"} {
		got, err := resolveDateFormat(raw)
		if err != nil {
			t.Fatalf("resolveDateFormat(%q) errored: %v", raw, err)
		}
		if got != "YYYY/MM/DD" {
			t.Errorf("resolveDateFormat(%q) = %q; want YYYY/MM/DD", raw, got)
		}
	}

	// Both rejections have to carry the list: it is the whole fix, and
	// the patterns are otherwise two levels deep in the help.
	for _, raw := range []string{"", "m/d/y"} {
		_, err := resolveDateFormat(raw)
		if err == nil {
			t.Fatalf("resolveDateFormat(%q) was accepted", raw)
		}
		for _, want := range []string{"YYYY/MM/DD", "YY.M.D", "today's date"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error for %q missing %q:\n%v", raw, want, err)
			}
		}
	}
}

func TestRenderDateFormat(t *testing.T) {
	// A two-digit day and month, then a one-digit pair: the patterns
	// differ in whether they pad, so both cases have to be right.
	padded := time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)
	for pattern, want := range map[string]string{
		"YYYY/MM/DD": "2026/08/27",
		"D/M/YY":     "27/8/26",
		"M/D/YY":     "8/27/26",
		"DD.MM.YYYY": "27.08.2026",
		"YY-M-D":     "26-8-27",
		"YY.M.D":     "26.8.27",
	} {
		if got := renderDateFormat(pattern, padded); got != want {
			t.Errorf("renderDateFormat(%q) = %q; want %q", pattern, got, want)
		}
	}

	single := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	for pattern, want := range map[string]string{
		"YY-M-D":     "26-11-5",
		"D/M/YY":     "5/11/26",
		"DD/MM/YYYY": "05/11/2026",
	} {
		if got := renderDateFormat(pattern, single); got != want {
			t.Errorf("renderDateFormat(%q) = %q; want %q", pattern, got, want)
		}
	}
}

// Every allowed pattern has to survive the layout translation, or the
// help would preview one of them as literal Y/M/D characters.
func TestDateFormatChoicesPreviewsEveryPattern(t *testing.T) {
	at := time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)
	choices := dateFormatChoices(at)
	for _, pattern := range dateFormats {
		rendered := renderDateFormat(pattern, at)
		if strings.ContainsAny(rendered, "YMD") {
			t.Errorf("pattern %q rendered as %q; a token was left untranslated", pattern, rendered)
		}
		if !strings.Contains(choices, pattern) || !strings.Contains(choices, rendered) {
			t.Errorf("choices omit %q or its preview %q:\n%s", pattern, rendered, choices)
		}
	}
	if lines := strings.Count(choices, "\n") + 1; lines != 6 {
		t.Errorf("choices span %d lines; want 6 (two patterns per line)", lines)
	}
}

func TestDescribeDateFormat(t *testing.T) {
	at := time.Date(2026, 8, 27, 0, 0, 0, 0, time.UTC)
	if got, want := describeDateFormat("DD.MM.YYYY", at), "DD.MM.YYYY (27.08.2026)"; got != want {
		t.Errorf("describeDateFormat() = %q; want %q", got, want)
	}
	if got := describeDateFormat("", at); got != nonEmpty("") {
		t.Errorf("describeDateFormat(\"\") = %q; want the placeholder %q", got, nonEmpty(""))
	}
}

// The upstream POST replaces the whole preferences object, so a patch has
// to leave every field the caller did not name exactly as it was read.
func TestWidgetPatchApplyOnlyTouchesNamedFields(t *testing.T) {
	current := widgetPreferences{
		ShowWidgets:    true,
		Is24HourFormat: true,
		DateFormat:     "YYYY/MM/DD",
		ShowDashboard:  true,
	}
	no := false
	format := "M/D/YY"

	t.Run("single_bool", func(t *testing.T) {
		got := widgetPatch{showWidgets: &no}.apply(current)
		want := current
		want.ShowWidgets = false
		if got != want {
			t.Errorf("apply() = %+v; want %+v", got, want)
		}
	})

	t.Run("date_format_only", func(t *testing.T) {
		got := widgetPatch{dateFormat: &format}.apply(current)
		want := current
		want.DateFormat = format
		if got != want {
			t.Errorf("apply() = %+v; want %+v", got, want)
		}
	})

	t.Run("empty_patch_is_identity", func(t *testing.T) {
		if got := (widgetPatch{}).apply(current); got != current {
			t.Errorf("apply() = %+v; want the input unchanged %+v", got, current)
		}
	})

	t.Run("false_is_not_mistaken_for_unset", func(t *testing.T) {
		zeroed := widgetPreferences{DateFormat: "YY-M-D"}
		yes := true
		got := widgetPatch{showDashboard: &yes}.apply(zeroed)
		if !got.ShowDashboard || got.DateFormat != "YY-M-D" {
			t.Errorf("apply() = %+v; want ShowDashboard set and DateFormat preserved", got)
		}
	})
}

// "set" is the only verb here, so the preferences live in its flags. The
// parent help has to name them or the subtree reads as if it did nothing.
func TestWidgetParentHelpNamesEveryPreference(t *testing.T) {
	long := NewWidgetCommand(nil).Long
	newWidgetSetCommand(nil).Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		if !strings.Contains(long, "--"+f.Name) {
			t.Errorf("widget help does not mention --%s", f.Name)
		}
	})
}

// Omitting a flag keeps the current value, so a "(default true)" in the
// help — which pflag prints for any non-zero default — would be a lie.
func TestWidgetSetFlagsAdvertiseNoDefault(t *testing.T) {
	usage := newWidgetSetCommand(nil).Flags().FlagUsages()
	if strings.Contains(usage, "default") {
		t.Errorf("flag help implies a default value:\n%s", usage)
	}
}

func TestRejectBooleanAsArgument(t *testing.T) {
	cmd := newWidgetSetCommand(nil)

	for _, arg := range []string{"false", "TRUE"} {
		err := rejectBooleanAsArgument(cmd, []string{arg})
		if err == nil {
			t.Fatalf("bare %q accepted as an argument", arg)
		}
		for _, want := range []string{"equals sign", strings.ToLower(arg)} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q missing %q", err, want)
			}
		}
	}

	if err := rejectBooleanAsArgument(cmd, nil); err != nil {
		t.Errorf("no arguments rejected: %v", err)
	}
	if err := rejectBooleanAsArgument(cmd, []string{"widgets"}); err == nil {
		t.Error("an unrelated argument was accepted")
	}
}

// The example has to name the flag the caller passed, or "--24-hour false"
// is answered with a line about a flag they never typed.
func TestRejectBooleanAsArgumentNamesTheFlagPassed(t *testing.T) {
	cmd := newWidgetSetCommand(nil)
	if err := cmd.Flags().Parse([]string{"--24-hour"}); err != nil {
		t.Fatalf("parsing --24-hour: %v", err)
	}
	err := rejectBooleanAsArgument(cmd, []string{"false"})
	if err == nil || !strings.Contains(err.Error(), "--24-hour=false") {
		t.Errorf("error = %v; want it to suggest --24-hour=false", err)
	}

	// Ambiguous: two booleans passed, so the example falls back rather
	// than guessing which one was missing its value.
	both := newWidgetSetCommand(nil)
	if err := both.Flags().Parse([]string{"--24-hour", "--show-dashboard"}); err != nil {
		t.Fatalf("parsing two booleans: %v", err)
	}
	if err := rejectBooleanAsArgument(both, []string{"false"}); err == nil ||
		!strings.Contains(err.Error(), "--show-widgets=false") {
		t.Errorf("error = %v; want the fallback example", err)
	}
}

func TestWidgetPatchEmpty(t *testing.T) {
	if !(widgetPatch{}).empty() {
		t.Error("a patch with no fields set reports non-empty")
	}
	flag := false
	format := "YY-M-D"
	for name, p := range map[string]widgetPatch{
		"show_widgets":   {showWidgets: &flag},
		"is_24_hour":     {is24HourFormat: &flag},
		"show_dashboard": {showDashboard: &flag},
		"date_format":    {dateFormat: &format},
	} {
		if p.empty() {
			t.Errorf("patch %s reports empty", name)
		}
	}
}
