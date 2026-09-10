package appearance

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func sampleView() appearanceView {
	return appearanceView{
		Locale: &localeConfig{
			Language: "zh-CN",
			Timezone: "Asia/Shanghai",
		},
		Widget: &widgetPreferences{
			ShowWidgets:    true,
			Is24HourFormat: false,
			DateFormat:     "M/D/YY",
			ShowDashboard:  false,
		},
		Wallpaper: &wallpaperConfig{
			Desktop:                  "/bg/3.jpg",
			DesktopStyle:             "cover",
			Login:                    "https://files.example/greeter.jpg",
			LoginStyle:               "fill",
			UploadDesktopBackgrounds: []string{"https://files.example/a.jpg"},
		},
	}
}

func TestRenderAppearanceCoversAllThreeSections(t *testing.T) {
	var buf bytes.Buffer
	if err := renderAppearance(&buf, sampleView()); err != nil {
		t.Fatalf("renderAppearance errored: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		// The code plus the name the picker shows for it, so an unfamiliar
		// locale does not have to be looked up elsewhere.
		"Locale", "Language:", "zh-CN (简体中文)", "Asia/Shanghai",
		"Widget", "Show widgets:", "24-hour format:", "Date format:", "M/D/YY",
		// The pattern alone does not say what the clock reads, which is
		// how Settings shows it: picker plus a live preview.
		"(" + renderDateFormat("M/D/YY", time.Now()) + ")",
		// The wallpaper rows are named as `wallpaper set` and Settings
		// name them: a built-in by number, a fill mode by its label.
		"Wallpaper", "Desktop:", "built-in 3", "Desktop style:", "Stretch",
		"Login:", "https://files.example/greeter.jpg", "Login style:", "Fill",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output is missing %q:\n%s", want, out)
		}
	}

	// The stored /bg/<n>.jpg token is an upstream detail; only -o json
	// carries it.
	if strings.Contains(out, "/bg/3.jpg") {
		t.Errorf("table leaked the stored wallpaper token:\n%s", out)
	}

	// Location is empty upstream; a blank line would read as a missing
	// row rather than an unset value.
	if !strings.Contains(out, "Location:  -") {
		t.Errorf("empty Location is not rendered as \"-\":\n%s", out)
	}

	// The uploaded lists are what `wallpaper delete` takes, so they are
	// printed in full rather than counted.
	if !strings.Contains(out, "https://files.example/a.jpg") {
		t.Errorf("uploaded desktop background is not listed:\n%s", out)
	}
	if !strings.Contains(out, "Uploaded login backgrounds\n  none") {
		t.Errorf("an empty uploaded list is not rendered as \"none\":\n%s", out)
	}
}

// 1.12.5 has no widget API, so that section arrives nil and the rest of
// the page still has to render.
func TestRenderAppearanceMarksAGatedSection(t *testing.T) {
	view := sampleView()
	view.Widget = nil

	var buf bytes.Buffer
	if err := renderAppearance(&buf, view); err != nil {
		t.Fatalf("renderAppearance errored on a nil section: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "Widget\n  requires Olares >= "+appearanceMinOlaresVersion) {
		t.Errorf("gated widget section is not called out with its minimum version:\n%s", out)
	}
	for _, want := range []string{"zh-CN", "built-in 3", "https://files.example/a.jpg"} {
		if !strings.Contains(out, want) {
			t.Errorf("section that does exist lost %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Date format:") {
		t.Errorf("absent section rendered zeroed rows:\n%s", out)
	}
}

// A nil wallpaper section must not print the uploaded-gallery lists,
// which would read as "the galleries are empty".
func TestRenderAppearanceSkipsGalleriesWhenWallpaperIsAbsent(t *testing.T) {
	view := sampleView()
	view.Wallpaper = nil

	var buf bytes.Buffer
	if err := renderAppearance(&buf, view); err != nil {
		t.Fatalf("renderAppearance errored: %v", err)
	}
	if out := buf.String(); strings.Contains(out, "Uploaded desktop backgrounds") {
		t.Errorf("gallery list printed for an absent wallpaper section:\n%s", out)
	}
}

func TestAppearanceViewJSONEmitsNullForAGatedSection(t *testing.T) {
	view := sampleView()
	view.Widget = nil
	raw, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// The key stays present so a scripted caller can tell "gated" from
	// "this CLI does not know about the section".
	got, ok := decoded["widget"]
	if !ok {
		t.Fatalf("widget key dropped entirely: %s", raw)
	}
	if string(got) != "null" {
		t.Errorf("widget = %s; want null", got)
	}
}

// The JSON view is a scripted-caller contract: three section keys, each
// holding the upstream field names verbatim.
func TestAppearanceViewJSONShape(t *testing.T) {
	raw, err := json.Marshal(sampleView())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded) != 3 {
		t.Errorf("top-level keys = %d; want exactly locale, widget, wallpaper: %s", len(decoded), raw)
	}
	for _, key := range []string{"locale", "widget", "wallpaper"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("missing top-level key %q: %s", key, raw)
		}
	}

	// The upstream locale payload carries a theme this page leaves out
	// (see theme.go); emitting it would advertise a setting no interface
	// reads.
	var locale map[string]interface{}
	if err := json.Unmarshal(decoded["locale"], &locale); err != nil {
		t.Fatalf("unmarshal locale: %v", err)
	}
	if _, ok := locale["theme"]; ok {
		t.Errorf("locale still carries a theme key: %s", raw)
	}
	// The picker's name for the locale belongs to the table only.
	if got := locale["language"]; got != "zh-CN" {
		t.Errorf("locale.language = %v; want the stored zh-CN", got)
	}

	var widget map[string]interface{}
	if err := json.Unmarshal(decoded["widget"], &widget); err != nil {
		t.Fatalf("unmarshal widget: %v", err)
	}
	// showWeight is the upstream name of the widget master switch.
	if got, ok := widget["showWeight"]; !ok || got != true {
		t.Errorf("widget.showWeight = %v (present=%v); want true", got, ok)
	}
	// The preview belongs to the table only: a script reading this back
	// has to get the pattern the upstream stores.
	if got := widget["dateFormat"]; got != "M/D/YY" {
		t.Errorf("widget.dateFormat = %v; want the stored M/D/YY", got)
	}

	// The table renames values for the reader; JSON must keep the stored
	// ones, or a script round-tripping them would write nonsense.
	var wallpaper map[string]interface{}
	if err := json.Unmarshal(decoded["wallpaper"], &wallpaper); err != nil {
		t.Fatalf("unmarshal wallpaper: %v", err)
	}
	for key, want := range map[string]interface{}{"desktop": "/bg/3.jpg", "desktopStyle": "cover"} {
		if got := wallpaper[key]; got != want {
			t.Errorf("wallpaper.%s = %v; want the stored %v", key, got, want)
		}
	}
	if _, ok := wallpaper["upload_desktop_backgrounds"]; !ok {
		t.Errorf("wallpaper is missing upload_desktop_backgrounds: %s", decoded["wallpaper"])
	}
}
