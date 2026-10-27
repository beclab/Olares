package appearance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveSurface(t *testing.T) {
	desktop, err := resolveSurface("Desktop")
	if err != nil {
		t.Fatalf("resolveSurface(\"Desktop\") errored: %v", err)
	}
	if desktop.name != "desktop" || desktop.uploadPolicy != "" {
		t.Errorf("desktop surface = %+v; want name=desktop and no upload policy", desktop)
	}

	login, err := resolveSurface("  login ")
	if err != nil {
		t.Fatalf("resolveSurface(\" login \") errored: %v", err)
	}
	// The greeter fetches the login background before the user is
	// authenticated, so it has to be uploaded as public.
	if login.uploadPolicy != "public" {
		t.Errorf("login upload policy = %q; want %q", login.uploadPolicy, "public")
	}

	// Every surface needs all four routes; a missing one would send a
	// write to the empty path.
	for name, s := range surfaces {
		for field, path := range map[string]string{
			"set": s.set, "style": s.style, "register": s.register, "remove": s.remove,
		} {
			if !strings.HasPrefix(path, "/api/wallpaper/") {
				t.Errorf("surface %s %s path = %q; want an /api/wallpaper/ route", name, field, path)
			}
		}
	}

	if _, err := resolveSurface("greeter"); err == nil {
		t.Error("resolveSurface accepted an unknown surface")
	}
}

// A built-in is selected by number, as in the Settings grid, and stored
// as /bg/<n>.jpg on both surfaces: the greeter resolves the value under
// its own asset root, so /login/<n>.jpg — the path the Settings page
// builds for its preview — is not a storable value.
func TestResolveWallpaperValue(t *testing.T) {
	desktop, login := surfaces["desktop"], surfaces["login"]

	for _, s := range []surface{desktop, login} {
		// A number is the documented spelling; the stored path is
		// accepted so `get` output can be fed back in.
		for raw, want := range map[string]string{
			"0":                            "/bg/0.jpg",
			"3":                            "/bg/3.jpg",
			fmt.Sprint(s.builtinCount - 1): builtinWallpaperValue(s.builtinCount - 1),
			"/bg/0.jpg":                    "/bg/0.jpg",
			"https://files.example/a.png":  "https://files.example/a.png",
			"http://files.example/a.png":   "http://files.example/a.png",
		} {
			got, err := resolveWallpaperValue(s, raw, false)
			if err != nil {
				t.Errorf("%s rejected valid %q: %v", s.name, raw, err)
				continue
			}
			if got != want {
				t.Errorf("%s resolveWallpaperValue(%q) = %q; want %q", s.name, raw, got, want)
			}
		}

		for _, bad := range []string{
			fmt.Sprint(s.builtinCount),                // one past the end
			fmt.Sprintf("/bg/%d.jpg", s.builtinCount), // same, spelled as a path
			"999",
			"/login/5.jpg", // a preview path, never stored
			"/bg/3.png",
			"/bg/.jpg",
			"-1",
			"bg/3.jpg",
			"",
		} {
			if _, err := resolveWallpaperValue(s, bad, false); err == nil {
				t.Errorf("%s accepted invalid %q", s.name, bad)
			}
		}
	}

	// The surfaces ship a different number of images, so the boundary
	// must be per-surface rather than shared.
	if desktop.builtinCount == login.builtinCount {
		t.Fatal("test assumes the two surfaces differ in built-in count")
	}
	edge := fmt.Sprint(desktop.builtinCount)
	if _, err := resolveWallpaperValue(desktop, edge, false); err == nil {
		t.Errorf("desktop accepted %q, which is past its own range", edge)
	}
	if _, err := resolveWallpaperValue(login, edge, false); err != nil {
		t.Errorf("login rejected %q, which is inside its range: %v", edge, err)
	}

	// An out-of-range number is a near miss, so the message only needs
	// the range; an unparseable value needs the whole how-to.
	_, err := resolveWallpaperValue(desktop, "999", false)
	if want := builtinWallpaperRange(desktop); !strings.Contains(err.Error(), want) {
		t.Errorf("error %q missing the range %q", err, want)
	}
	_, err = resolveWallpaperValue(desktop, "/bg/3.png", false)
	for _, want := range []string{"wallpaper list desktop", "--force", builtinWallpaperRange(desktop)} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing %q", err, want)
		}
	}
}

// --force relaxes the checks, not the vocabulary. The bare number is not
// a value the backend can resolve, so forcing one has to store
// /bg/<n>.jpg -- otherwise the escape hatch offered by the out-of-range
// message silently writes a broken background.
func TestResolveWallpaperValueForceKeepsTheNumberVocabulary(t *testing.T) {
	s := surfaces["desktop"]

	// The exact path the out-of-range error invites: a built-in a newer
	// release added, past the count this CLI was built with.
	beyond := s.builtinCount
	got, err := resolveWallpaperValue(s, fmt.Sprint(beyond), true)
	if err != nil {
		t.Fatalf("force rejected %d: %v", beyond, err)
	}
	if want := builtinWallpaperValue(beyond); got != want {
		t.Errorf("force resolveWallpaperValue(%d) = %q; want %q", beyond, got, want)
	}

	// In-range numbers resolve the same with or without the flag.
	for _, raw := range []string{"0", "3"} {
		forced, err := resolveWallpaperValue(s, raw, true)
		if err != nil {
			t.Fatalf("force rejected %q: %v", raw, err)
		}
		plain, err := resolveWallpaperValue(s, raw, false)
		if err != nil {
			t.Fatalf("%q rejected without force: %v", raw, err)
		}
		if forced != plain {
			t.Errorf("--force changed %q from %q to %q", raw, plain, forced)
		}
	}

	// A shape this CLI cannot read is the one case force sends verbatim,
	// since there is nothing to resolve it to.
	for _, raw := range []string{"/bg/3.png", "/login/5.jpg"} {
		got, err := resolveWallpaperValue(s, raw, true)
		if err != nil {
			t.Fatalf("force rejected %q: %v", raw, err)
		}
		if got != raw {
			t.Errorf("force resolveWallpaperValue(%q) = %q; want it verbatim", raw, got)
		}
	}

	// Whatever force stores, `get` and the confirmation line must be able
	// to name it the way the user selected it.
	forced, err := resolveWallpaperValue(s, fmt.Sprint(beyond), true)
	if err != nil {
		t.Fatalf("force rejected %d: %v", beyond, err)
	}
	if want := fmt.Sprintf("built-in %d", beyond); describeWallpaperValue(forced) != want {
		t.Errorf("describeWallpaperValue(%q) = %q; want %q", forced, describeWallpaperValue(forced), want)
	}
}

// The table renames values for the reader; the JSON output is a scripted
// caller's contract and must keep the stored spellings, which are also
// what `set` accepts.
func TestWallpaperChoicesJSONKeepsStoredValues(t *testing.T) {
	s := surfaces["desktop"]
	raw, err := json.Marshal(wallpaperChoices{
		Surface:  s.name,
		Builtin:  s.builtinWallpapers(),
		Uploaded: []string{},
		Active:   "/bg/3.jpg",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded struct {
		Surface  string   `json:"surface"`
		Builtin  []string `json:"builtin"`
		Uploaded []string `json:"uploaded"`
		Active   string   `json:"active"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Active != "/bg/3.jpg" {
		t.Errorf("active = %q; want the stored /bg/3.jpg, not a display name", decoded.Active)
	}
	if len(decoded.Builtin) != s.builtinCount || decoded.Builtin[0] != "/bg/0.jpg" {
		t.Errorf("builtin = %v; want the %d stored values", decoded.Builtin, s.builtinCount)
	}
	// An absent gallery must marshal as [] rather than null, so a caller
	// can iterate it unconditionally.
	if !strings.Contains(string(raw), `"uploaded":[]`) {
		t.Errorf("empty gallery is not an empty array: %s", raw)
	}
}

// Output names a value the same way input takes it, so `get` and `set`
// share one vocabulary.
func TestDescribeWallpaperValue(t *testing.T) {
	for raw, want := range map[string]string{
		"/bg/0.jpg":                   "built-in 0",
		"/bg/27.jpg":                  "built-in 27",
		"https://files.example/a.png": "https://files.example/a.png",
		"":                            "-",
	} {
		if got := describeWallpaperValue(raw); got != want {
			t.Errorf("describeWallpaperValue(%q) = %q; want %q", raw, got, want)
		}
	}
}

func TestBuiltinWallpapersEnumeratesTheWholeRange(t *testing.T) {
	s := surfaces["desktop"]
	got := s.builtinWallpapers()
	if len(got) != s.builtinCount {
		t.Fatalf("builtinWallpapers() = %d entries; want %d", len(got), s.builtinCount)
	}
	if got[0] != "/bg/0.jpg" {
		t.Errorf("first entry = %q; want /bg/0.jpg", got[0])
	}
	if last, want := got[len(got)-1], fmt.Sprintf("/bg/%d.jpg", s.builtinCount-1); last != want {
		t.Errorf("last entry = %q; want %q", last, want)
	}
	// Every generated value must be accepted by set, or `list` would be
	// advertising values `set` refuses.
	for _, v := range got {
		if _, err := resolveWallpaperValue(s, v, false); err != nil {
			t.Fatalf("listed value %q is rejected by set: %v", v, err)
		}
	}
}

// The built-ins are printed as a range: a terminal cannot show the
// thumbnails that make the Settings grid meaningful, so one line per
// image would be noise.
func TestRenderWallpaperChoicesShowsTheRangeAndTheActiveOne(t *testing.T) {
	var buf bytes.Buffer
	err := renderWallpaperChoices(&buf, wallpaperChoices{
		Surface:  "desktop",
		Builtin:  []string{"/bg/0.jpg", "/bg/1.jpg", "/bg/2.jpg"},
		Uploaded: []string{"https://files.example/a.png"},
		Active:   "/bg/1.jpg",
	})
	if err != nil {
		t.Fatalf("renderWallpaperChoices errored: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "0-2, currently 1") {
		t.Errorf("built-in range or active number missing:\n%s", out)
	}
	if strings.Contains(out, "/bg/0.jpg") {
		t.Errorf("built-ins are enumerated as paths:\n%s", out)
	}
	if !strings.Contains(out, "  https://files.example/a.png") {
		t.Errorf("uploaded wallpaper missing:\n%s", out)
	}
	// The reader must not have to guess the verb that consumes this.
	if !strings.Contains(out, "wallpaper set desktop <number|url>") {
		t.Errorf("output does not say how to select:\n%s", out)
	}
}

func TestRenderWallpaperChoicesMarksAnActiveUpload(t *testing.T) {
	var buf bytes.Buffer
	const url = "https://files.example/a.png"
	if err := renderWallpaperChoices(&buf, wallpaperChoices{
		Surface:  "desktop",
		Builtin:  []string{"/bg/0.jpg"},
		Uploaded: []string{url, "https://files.example/b.png"},
		Active:   url,
	}); err != nil {
		t.Fatalf("renderWallpaperChoices errored: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "* "+url) {
		t.Errorf("active upload is not marked:\n%s", out)
	}
	// No built-in is active, so no "currently" claim may appear.
	if strings.Contains(out, "currently") {
		t.Errorf("a built-in is reported active while an upload is:\n%s", out)
	}
}

func TestRenderWallpaperChoicesEmptyGallery(t *testing.T) {
	var buf bytes.Buffer
	if err := renderWallpaperChoices(&buf, wallpaperChoices{
		Surface: "login",
		Builtin: []string{"/bg/0.jpg"},
		Active:  "/bg/0.jpg",
	}); err != nil {
		t.Fatalf("renderWallpaperChoices errored: %v", err)
	}
	if !strings.Contains(buf.String(), "Uploaded login wallpapers\n  none") {
		t.Errorf("empty gallery is not rendered as \"none\":\n%s", buf.String())
	}
}

// The fill modes are named in the CLI the way Settings names them, so a
// user typing what they read gets what they expect. The stored values
// stay accepted for scripts.
func TestResolveContentMode(t *testing.T) {
	for _, m := range contentModes {
		for _, raw := range []string{m.label, strings.ToLower(m.label), strings.ToUpper(m.label), m.wire} {
			got, err := resolveContentMode(raw)
			if err != nil {
				t.Errorf("resolveContentMode(%q) errored: %v", raw, err)
				continue
			}
			if got != m.wire {
				t.Errorf("resolveContentMode(%q) = %q; want %q", raw, got, m.wire)
			}
		}
		if got := contentModeLabel(m.wire); got != m.label {
			t.Errorf("contentModeLabel(%q) = %q; want %q", m.wire, got, m.label)
		}
	}

	// An unknown mode is listed by label, since that is what the user saw.
	_, err := resolveContentMode("scaled")
	if err == nil {
		t.Fatal("resolveContentMode accepted an unknown mode")
	}
	if !strings.Contains(err.Error(), "allowed: Fill, Stretch, Tile") {
		t.Errorf("error %q does not list the allowed modes by label", err)
	}

	// A value a newer release adds must still print rather than vanish.
	if got := contentModeLabel("scaled"); got != "scaled" {
		t.Errorf("contentModeLabel passed over an unknown value: %q", got)
	}
}

func TestMultipartImageSendsFieldsThenFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "greeter.png")
	const payload = "not-really-a-png"
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	body, contentType, err := multipartImage(path, map[string]string{"policy": "public", "skipped": "  "})
	if err != nil {
		t.Fatalf("multipartImage errored: %v", err)
	}
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("parse content type %q: %v", contentType, err)
	}

	fields := map[string]string{}
	files := map[string]string{}
	filenames := map[string]string{}
	mr := multipart.NewReader(body, params["boundary"])
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read part: %v", err)
		}
		content, err := io.ReadAll(part)
		if err != nil {
			t.Fatalf("read part body: %v", err)
		}
		if part.FileName() == "" {
			fields[part.FormName()] = string(content)
			continue
		}
		files[part.FormName()] = string(content)
		filenames[part.FormName()] = part.FileName()
	}

	if fields["policy"] != "public" {
		t.Errorf("policy field = %q; want %q", fields["policy"], "public")
	}
	if _, ok := fields["skipped"]; ok {
		t.Error("a blank field was written; blank fields must be dropped")
	}
	// The uploader reads the file from the `image` part.
	if files["image"] != payload {
		t.Errorf("image part = %q; want %q", files["image"], payload)
	}
	if filenames["image"] != "greeter.png" {
		t.Errorf("image filename = %q; want the basename %q", filenames["image"], "greeter.png")
	}
}

func TestMultipartImageMissingFile(t *testing.T) {
	if _, _, err := multipartImage(filepath.Join(t.TempDir(), "absent.png"), nil); err == nil {
		t.Error("multipartImage accepted a path that does not exist")
	}
}
