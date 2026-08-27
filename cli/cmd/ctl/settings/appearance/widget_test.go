package appearance

import (
	"strings"
	"testing"
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

	if _, err := resolveDateFormat("yyyy/mm/dd"); err == nil {
		t.Error("resolveDateFormat accepted a lowercased format; the upstream is case-sensitive")
	}
	_, err := resolveDateFormat("")
	if err == nil || !strings.Contains(err.Error(), "requires a value") {
		t.Errorf("empty --date-format error = %v; want it to mention a required value", err)
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
