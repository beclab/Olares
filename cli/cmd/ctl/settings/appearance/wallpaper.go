package appearance

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings appearance wallpaper ...`
//
// Backed by user-service's wallpaper.controller.ts. The desktop and login
// surfaces have separate endpoints throughout, and the login ones fan out
// to BFL's /bfl/settings/v1alpha1/set-login-background so the greeter
// picks the change up.
//
// Current values are shown by `appearance get`.
const wallpaperPath = "/api/wallpaper"

// imageUploadPath is served by tapr's images-uploader, not user-service,
// and is reachable only on the settings ingress — see imageClient.
const imageUploadPath = "/images/upload/v1"

// wallpaperConfig mirrors user-service's Wallpaper type.
type wallpaperConfig struct {
	Desktop                  string   `json:"desktop"`
	Login                    string   `json:"login"`
	DesktopStyle             string   `json:"desktopStyle"`
	LoginStyle               string   `json:"loginStyle"`
	UploadDesktopBackgrounds []string `json:"upload_desktop_backgrounds"`
	UploadLoginBackgrounds   []string `json:"upload_login_backgrounds"`
}

// contentModes mirrors IMG_CONTENT_MODE in apps/.../constant/index.ts:
// the label is what Settings -> Appearance shows, the wire value is what
// user-service stores and validates. Users read the label, so the CLI
// takes either and always prints the label.
var contentModes = []struct {
	label string
	wire  string
}{
	{"Fill", "fill"},
	{"Stretch", "cover"},
	{"Tile", "repeat"},
}

// contentModeLabel maps a stored value back to the label the UI shows,
// so `get` and the CLI's own confirmations read like the page. Unknown
// values pass through: a newer release may add one.
func contentModeLabel(wire string) string {
	for _, m := range contentModes {
		if wire == m.wire {
			return m.label
		}
	}
	return wire
}

func contentModeLabels() []string {
	out := make([]string, 0, len(contentModes))
	for _, m := range contentModes {
		out = append(out, m.label)
	}
	return out
}

// surface holds the per-surface endpoint spellings. The desktop/login
// split is irregular enough (camelCase style routes, a login-only upload
// policy) that keeping it in one table beats branching at each call site.
type surface struct {
	name string
	// set selects an existing image as the active wallpaper.
	set string
	// style sets the fill mode.
	style string
	// register records an uploaded image in the surface's gallery.
	register string
	// remove drops an uploaded image from the gallery.
	remove string
	// uploadPolicy is the `policy` form field the images-uploader needs.
	// Login backgrounds must be public: the greeter fetches them before
	// the user is authenticated.
	uploadPolicy string
	// builtinCount is how many built-in images this surface offers,
	// numbered 0..builtinCount-1. It mirrors picturesCount in the SPA's
	// Appearance page: nothing upstream enumerates them, and the two
	// surfaces ship a different number of images.
	builtinCount int
}

var surfaces = map[string]surface{
	"desktop": {
		name:         "desktop",
		set:          "/api/wallpaper/desktop",
		style:        "/api/wallpaper/desktopStyle",
		register:     "/api/wallpaper/upload/desktop",
		remove:       "/api/wallpaper/delete/desktop",
		builtinCount: 28,
	},
	"login": {
		name:         "login",
		set:          "/api/wallpaper/login",
		style:        "/api/wallpaper/loginStyle",
		register:     "/api/wallpaper/upload/login",
		remove:       "/api/wallpaper/delete/login",
		uploadPolicy: "public",
		builtinCount: 29,
	},
}

// builtinWallpaperPrefix is the token both surfaces store for a built-in
// image. The login surface stores it too: the greeter resolves the value
// under its own asset root (LoginPage.vue renders "auth/" + value), and
// the Settings preview swaps the prefix to /login/ only for display. A
// stored /login/<n>.jpg therefore resolves nowhere.
const builtinWallpaperPrefix = "/bg/"

// builtinWallpapers lists every value `wallpaper set` accepts for s
// without an upload.
func (s surface) builtinWallpapers() []string {
	out := make([]string, 0, s.builtinCount)
	for i := 0; i < s.builtinCount; i++ {
		out = append(out, builtinWallpaperValue(i))
	}
	return out
}

// resolveWallpaperValue turns what the user types into the value to store.
//
// A built-in is addressed by its number, because that is all the Settings
// page offers too — a grid of unnamed thumbnails. The stored /bg/<n>.jpg
// spelling is accepted as well, so a value read back from `get` can be
// fed straight back in.
//
// Validation is on this side because the upstream stores the value
// verbatim and never checks it: an out-of-range number would leave a
// broken background behind with no error. Uploaded URLs are judged on
// their shape rather than looked up, to keep this offline.
func resolveWallpaperValue(s surface, raw string) (string, error) {
	if isUploadedWallpaperURL(raw) {
		return raw, nil
	}
	if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
		if n >= s.builtinCount {
			return "", fmt.Errorf("%s has no built-in wallpaper %d: %s", s.name, n, builtinWallpaperRange(s))
		}
		return builtinWallpaperValue(n), nil
	}
	if n, ok := builtinWallpaperIndex(raw); ok {
		if n >= s.builtinCount {
			return "", fmt.Errorf("%s has no built-in wallpaper %d: %s", s.name, n, builtinWallpaperRange(s))
		}
		return raw, nil
	}
	return "", fmt.Errorf("%q is not a %s wallpaper: pass a built-in number (%s) or an https:// URL from `wallpaper upload`; run `olares-cli settings appearance wallpaper list %s` to see the choices (--force sends it anyway)",
		raw, s.name, builtinWallpaperRange(s), s.name)
}

func isUploadedWallpaperURL(raw string) bool {
	return strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://")
}

func builtinWallpaperValue(n int) string {
	return fmt.Sprintf("%s%d.jpg", builtinWallpaperPrefix, n)
}

func builtinWallpaperRange(s surface) string {
	return fmt.Sprintf("0-%d", s.builtinCount-1)
}

// describeWallpaperValue names a stored value the way the user selects
// it, so output and input use one vocabulary instead of exposing the
// /bg/<n>.jpg token that only the backend cares about.
func describeWallpaperValue(bg string) string {
	if n, ok := builtinWallpaperIndex(bg); ok {
		return fmt.Sprintf("built-in %d", n)
	}
	return nonEmpty(bg)
}

// builtinWallpaperIndex parses "/bg/<n>.jpg" into n.
func builtinWallpaperIndex(bg string) (int, bool) {
	rest, ok := strings.CutPrefix(bg, builtinWallpaperPrefix)
	if !ok {
		return 0, false
	}
	rest, ok = strings.CutSuffix(rest, ".jpg")
	if !ok || rest == "" {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

func resolveSurface(raw string) (surface, error) {
	s, ok := surfaces[strings.ToLower(strings.TrimSpace(raw))]
	if !ok {
		return surface{}, fmt.Errorf("unknown surface %q (allowed: desktop, login)", raw)
	}
	return s, nil
}

func NewWallpaperCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallpaper",
		Short: "desktop and login wallpaper",
		Long: `Change the desktop or login wallpaper (Settings -> Appearance >
Wallpaper).

Current values are shown by "settings appearance get"; the values you can
select are listed by "wallpaper list <surface>".

Subcommands:
  list <desktop|login>                  list the selectable images
  set <desktop|login> <bg>              select an image
  style set <desktop|login> <mode>      set the fill mode
  upload <desktop|login> --file <path>  upload a local image and select it
  delete <desktop|login> <bg>           remove an uploaded image
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newWallpaperListCommand(f))
	cmd.AddCommand(newWallpaperSetCommand(f))
	cmd.AddCommand(newWallpaperStyleCommand(f))
	cmd.AddCommand(newWallpaperUploadCommand(f))
	cmd.AddCommand(newWallpaperDeleteCommand(f))
	return cmd
}

// `olares-cli settings appearance wallpaper list <surface>`
//
// Exists because nothing upstream enumerates the built-in images: their
// count lives only in the SPA, and `get` shows the active value plus the
// uploaded gallery. Without this, picking a wallpaper from the CLI means
// guessing a number at which a wrong guess is accepted silently.
func newWallpaperListCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "list <desktop|login>",
		Short: "show what this surface can be set to",
		Long: `Show what "wallpaper set <surface>" accepts: the range of built-in
image numbers, and the images uploaded to this surface's gallery.

The built-in range differs per surface, so pass the one you mean. The
range is known locally and costs no request; the uploaded images are read
from the backend.

Examples:
  olares-cli settings appearance wallpaper list desktop
  olares-cli settings appearance wallpaper list login -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			s, err := resolveSurface(args[0])
			if err != nil {
				return err
			}
			return runWallpaperList(c.Context(), f, s, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// wallpaperChoices is the --output json contract for `wallpaper list`.
type wallpaperChoices struct {
	Surface  string   `json:"surface"`
	Builtin  []string `json:"builtin"`
	Uploaded []string `json:"uploaded"`
	Active   string   `json:"active"`
}

func runWallpaperList(ctx context.Context, f *cmdutil.Factory, s surface, outputRaw string) error {
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
	var cfg wallpaperConfig
	if err := doGetEnvelope(ctx, pc.doer, wallpaperPath, &cfg); err != nil {
		return err
	}

	choices := wallpaperChoices{
		Surface:  s.name,
		Builtin:  s.builtinWallpapers(),
		Uploaded: cfg.UploadDesktopBackgrounds,
		Active:   cfg.Desktop,
	}
	if s.name == "login" {
		choices.Uploaded, choices.Active = cfg.UploadLoginBackgrounds, cfg.Login
	}
	if choices.Uploaded == nil {
		choices.Uploaded = []string{}
	}

	if format == FormatJSON {
		return printJSON(os.Stdout, choices)
	}
	return renderWallpaperChoices(os.Stdout, choices)
}

// renderWallpaperChoices prints the built-ins as a range rather than one
// line per image: a terminal cannot show the thumbnails that make the
// Settings grid meaningful, so 28 near-identical lines would not help
// anyone choose. The uploaded URLs are listed, since those are distinct.
func renderWallpaperChoices(w io.Writer, c wallpaperChoices) error {
	var active string
	if n, ok := builtinWallpaperIndex(c.Active); ok {
		active = fmt.Sprintf(", currently %d", n)
	}
	if _, err := fmt.Fprintf(w, "Built-in %s wallpapers\n  %s%s\n\n",
		c.Surface, builtinWallpaperNumbers(c), active); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "Uploaded %s wallpapers\n", c.Surface); err != nil {
		return err
	}
	if len(c.Uploaded) == 0 {
		if _, err := fmt.Fprintln(w, "  none"); err != nil {
			return err
		}
	}
	for _, url := range c.Uploaded {
		marker := "  "
		if url == c.Active {
			marker = "* "
		}
		if _, err := fmt.Fprintf(w, "%s%s\n", marker, url); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(w, "\nSelect one with `settings appearance wallpaper set %s <number|url>`.\n", c.Surface)
	return err
}

func builtinWallpaperNumbers(c wallpaperChoices) string {
	if len(c.Builtin) == 0 {
		return "none"
	}
	return fmt.Sprintf("0-%d", len(c.Builtin)-1)
}

func newWallpaperSetCommand(f *cmdutil.Factory) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "set <desktop|login> <number|url>",
		Short: "select an existing image as the wallpaper",
		Long: `Select an image as the desktop or login wallpaper.

Two kinds of value work:

  <number>      a built-in image, numbered as in "wallpaper list
                <surface>"; the range differs between desktop and login
  https://...   an image previously uploaded (see "wallpaper upload")

The upstream stores the value verbatim and never checks it, so an
out-of-range number would leave a broken background behind with no error.
This command therefore validates first; --force skips that for a value a
newer release adds.

Examples:
  olares-cli settings appearance wallpaper set desktop 3
  olares-cli settings appearance wallpaper set login 5
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			s, err := resolveSurface(args[0])
			if err != nil {
				return err
			}
			raw := strings.TrimSpace(args[1])
			if raw == "" {
				return fmt.Errorf("set requires a built-in number or an uploaded URL")
			}
			bg := raw
			if !force {
				if bg, err = resolveWallpaperValue(s, raw); err != nil {
					return err
				}
			}
			return runWallpaperSet(c.Context(), f, s, bg)
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "store the value as given, skipping validation (forward compatibility)")
	return cmd
}

func runWallpaperSet(ctx context.Context, f *cmdutil.Factory, s surface, bg string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", s.set, map[string]string{"bg": bg}, nil); err != nil {
		return err
	}
	fmt.Printf("Set the %s wallpaper to %s.\n", s.name, describeWallpaperValue(bg))
	return nil
}

func newWallpaperStyleCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "style",
		Short: "wallpaper fill mode",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "set <desktop|login> <Fill|Stretch|Tile>",
		Short: "set how the wallpaper fills the screen",
		Long: `Set how the wallpaper fills the screen.

Modes are the ones in Settings -> Appearance: Fill, Stretch, Tile. The
stored values (fill, cover, repeat) are accepted too.

Examples:
  olares-cli settings appearance wallpaper style set desktop Stretch
  olares-cli settings appearance wallpaper style set login Fill
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			s, err := resolveSurface(args[0])
			if err != nil {
				return err
			}
			mode, err := resolveContentMode(args[1])
			if err != nil {
				return err
			}
			return runWallpaperStyleSet(c.Context(), f, s, mode)
		},
	})
	return cmd
}

// resolveContentMode takes what the user saw in Settings (Fill, Stretch,
// Tile) or the stored value (fill, cover, repeat) and returns the stored
// value.
func resolveContentMode(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	for _, m := range contentModes {
		if value == strings.ToLower(m.label) || value == m.wire {
			return m.wire, nil
		}
	}
	return "", fmt.Errorf("unsupported fill mode %q (allowed: %s)", raw, strings.Join(contentModeLabels(), ", "))
}

func runWallpaperStyleSet(ctx context.Context, f *cmdutil.Factory, s surface, mode string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", s.style, map[string]string{"style": mode}, nil); err != nil {
		return err
	}
	fmt.Printf("Set the %s wallpaper fill mode to %s.\n", s.name, contentModeLabel(mode))
	return nil
}

// uploadedImage is the images-uploader's response payload.
type uploadedImage struct {
	ImageURL string `json:"imageUrl"`
	Size     int64  `json:"size"`
}

func newWallpaperUploadCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		file   string
		noSet  bool
		output string
	)
	cmd := &cobra.Command{
		Use:   "upload <desktop|login> --file <path>",
		Short: "upload a local image and make it the wallpaper",
		Long: `Upload a local image, add it to the surface's gallery, and select it
as the wallpaper. Pass --no-set to only upload and register it.

The upload goes to the settings ingress rather than the desktop one,
because the image service is only proxied there. Login backgrounds are
uploaded with a public access policy, which the greeter needs to fetch
them before anyone is signed in.

Accepted image types: png, jpeg, jpg, gif.

Examples:
  olares-cli settings appearance wallpaper upload desktop --file ~/Pictures/bg.jpg
  olares-cli settings appearance wallpaper upload login --file ./greeter.png --no-set
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			s, err := resolveSurface(args[0])
			if err != nil {
				return err
			}
			if strings.TrimSpace(file) == "" {
				return fmt.Errorf("upload requires --file <path>")
			}
			return runWallpaperUpload(c.Context(), f, s, file, noSet, output)
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "path to the local image to upload (required)")
	cmd.Flags().BoolVar(&noSet, "no-set", false, "upload and register the image without selecting it")
	addOutputFlag(cmd, &output)
	return cmd
}

func runWallpaperUpload(ctx context.Context, f *cmdutil.Factory, s surface, file string, noSet bool, outputRaw string) error {
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

	fields := map[string]string{}
	if s.uploadPolicy != "" {
		fields["policy"] = s.uploadPolicy
	}
	var uploaded uploadedImage
	if err := pc.images.upload(ctx, imageUploadPath, file, fields, &uploaded); err != nil {
		return err
	}
	if strings.TrimSpace(uploaded.ImageURL) == "" {
		return fmt.Errorf("upload succeeded but the image service returned no imageUrl")
	}

	if err := doMutateEnvelope(ctx, pc.doer, "POST", s.register,
		map[string]string{"bg": uploaded.ImageURL}, nil); err != nil {
		return fmt.Errorf("register the uploaded image: %w", err)
	}
	if !noSet {
		if err := doMutateEnvelope(ctx, pc.doer, "POST", s.set,
			map[string]string{"bg": uploaded.ImageURL}, nil); err != nil {
			return fmt.Errorf("select the uploaded image: %w", err)
		}
	}

	if format == FormatJSON {
		return printJSON(os.Stdout, uploaded)
	}
	if noSet {
		fmt.Printf("Uploaded %s to the %s gallery: %s\n", file, s.name, uploaded.ImageURL)
		return nil
	}
	fmt.Printf("Uploaded %s and set it as the %s wallpaper: %s\n", file, s.name, uploaded.ImageURL)
	return nil
}

func newWallpaperDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	var assumeYes bool
	cmd := &cobra.Command{
		Use:     "delete <desktop|login> <url>",
		Aliases: []string{"rm"},
		Short:   "remove an uploaded image from a wallpaper gallery",
		Long: `Remove an uploaded image from the desktop or login gallery.

Only uploaded images can be removed, so the value is one of the URLs
listed by "wallpaper list <surface>". If the image being removed is the
active wallpaper, the upstream falls back to a built-in default.

Examples:
  olares-cli settings appearance wallpaper delete desktop https://files.example/bg.jpg
  olares-cli settings appearance wallpaper delete login https://files.example/greeter.png --yes
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			s, err := resolveSurface(args[0])
			if err != nil {
				return err
			}
			bg := strings.TrimSpace(args[1])
			if bg == "" {
				return fmt.Errorf("delete requires the URL of an uploaded image")
			}
			return runWallpaperDelete(c.Context(), f, s, bg, assumeYes)
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the y/N prompt (required for non-TTY stdin)")
	return cmd
}

func runWallpaperDelete(ctx context.Context, f *cmdutil.Factory, s surface, bg string, assumeYes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Remove %q from the %s wallpaper gallery?", bg, s.name),
		assumeYes); err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", s.remove, map[string]string{"bg": bg}, nil); err != nil {
		return err
	}
	fmt.Printf("Removed %q from the %s wallpaper gallery.\n", bg, s.name)
	return nil
}
