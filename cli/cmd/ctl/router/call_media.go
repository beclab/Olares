package router

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// The media create routes, and the two routes each of them leaves behind:
// GET …/{id} and GET …/{id}/content.
//
// Three routes create: /v1/images/generations, /v1/videos and /v1/generations.
// They write the same record and answer with the same shape, so everything past
// the submission is shared — the wait, the download, the file name. Which route
// a verb uses, and why, is in call_media_canonical.go.
//
// These are the only calls in this tree that outlive the request that made them.
// An image takes seconds and a video takes minutes, so Router records the work
// as a generation, hands back its id, and answers questions about it afterwards.
// The bytes are never in that record — a video in a JSON field is a video nobody
// can stream — so the file comes from the /content route.
//
// Both verbs therefore have the same shape: submit, wait, write the file. And
// both take --no-wait, which prints the id and stops, and --id, which picks a
// generation up later. A generation expires; --id after that is a 404 rather
// than a file, which is why --no-wait says when.
//
// Images have one wrinkle worth stating. Router serves image generation two
// ways: synchronously, forwarding the provider's answer as it arrives, and
// asynchronously, persisting a generation. This verb asks for the second, since
// that is what gives a file to write and an id to come back for — and falls back
// to the first on a Router with no persistent media API, where the bytes arrive
// inline instead. Video has only the asynchronous path.
//
// Router also serves /v1/images/edits and /v1/images/variations, and neither has
// a verb here. Both take an input image and a mask, which is a file-handling
// surface of its own rather than another flag on this one, and an edit is
// something people reach for in a picture editor. Left to a direct call.

// generationView is Router's record of one piece of work, and every media
// family answers with it on all three routes.
//
// What a generation produced is stated in outputs rather than left inside the
// provider's own JSON, and progress and usage are Router's fields rather than
// the vendor's. provider_response is whatever the upstream said that Router has
// no field for: a whitelist, optional, and not needed to use a generation.
type generationView struct {
	ID        string             `json:"id"`
	Object    string             `json:"object"`
	MediaType string             `json:"media_type"`
	Operation string             `json:"operation"`
	Status    string             `json:"status"`
	Progress  *float64           `json:"progress,omitempty"`
	Model     string             `json:"model"`
	Outputs   []generationOutput `json:"outputs,omitempty"`
	Usage     *generationUsage   `json:"usage,omitempty"`
	Response  json.RawMessage    `json:"provider_response,omitempty"`
	ErrorCode *string            `json:"error_code,omitempty"`
	Error     *string            `json:"error,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	ExpiresAt time.Time          `json:"expires_at"`
}

// generationOutput is one artifact. A zero measurement means the provider did
// not report it, never that it is zero, and content_url addresses this Router
// rather than anything upstream.
type generationOutput struct {
	ID              string  `json:"id"`
	ContentType     string  `json:"content_type,omitempty"`
	Width           int     `json:"width,omitempty"`
	Height          int     `json:"height,omitempty"`
	DurationSeconds float64 `json:"duration_seconds,omitempty"`
	ContentURL      string  `json:"content_url"`
}

// generationUsage is what the provider said it produced, and it is absent
// unless the provider said something.
type generationUsage struct {
	Count   int     `json:"count,omitempty"`
	Seconds float64 `json:"seconds,omitempty"`
}

func (g *generationView) done() bool {
	switch strings.ToLower(g.Status) {
	case "completed", "failed":
		return true
	}
	return false
}

func (g *generationView) failed() bool { return strings.EqualFold(g.Status, "failed") }

// progressNote is the one thing worth showing while waiting. Not every provider
// reports it, and its absence is not worth a word.
func (g *generationView) progressNote() string {
	if g.Progress == nil {
		return ""
	}
	return fmt.Sprintf(", %.0f%%", *g.Progress)
}

// outputIDs are the pieces this generation produced, named. Asking for a named
// output is how a caller avoids depending on which one happens to be first —
// and a provider that settled without naming anything is reported as one
// unnamed output, which is nothing to choose between and so nothing to list.
func (g *generationView) outputIDs() []string {
	out := make([]string, 0, len(g.Outputs))
	for _, o := range g.Outputs {
		if s := strings.TrimSpace(o.ID); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func (g *generationView) reason() string {
	switch {
	case g.Error != nil && strings.TrimSpace(*g.Error) != "":
		return *g.Error
	case g.ErrorCode != nil && strings.TrimSpace(*g.ErrorCode) != "":
		return *g.ErrorCode
	}
	return "the provider did not say why"
}

type mediaOptions struct {
	Out      string
	OutputID string
	Wait     bool
	Timeout  time.Duration
	APIKey   string
	OutputIn string
	ID       string
	// Body is the request, built by the verb. Which shape it is follows from
	// the route: the released routes take their own flat keys, and
	// /v1/generations takes the canonical body. Neither is assembled here,
	// because a verb knows which fields its family admits and this does not.
	Body any
}

func newCallImageCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		model    string
		out      string
		outputID string
		noWait   bool
		id       string
		timeout  time.Duration
		apiKey   string
		flags    mediaFlags
	)
	cmd := &cobra.Command{
		Use:   "image [prompt…]",
		Short: "generate an image",
		Long: `Generate an image from a description.

The image is written to --out, or to a file named after the generation in the
current directory. What comes back is a file rather than a URL: Router holds the
bytes, so the picture does not depend on a provider's link staying alive.

--size asks for pixels and --aspect-ratio for a shape; they describe the same
thing two ways, so give one. Values are checked against the resolved model's own
parameters before anything is billed, and a field it has no parameter for is
refused rather than dropped.

--no-wait prints the generation id and stops; "--id <id>" collects it later, and
also reports one that is still running. A generation expires, and --no-wait says
when: after that the id is gone along with the image.

Image generation needs a model whose mode is image_generation; "olares-cli router
list --mode image_generation" shows the ones that qualify.

Examples:
  olares-cli router call image "a red bicycle in the rain"
  olares-cli router call image "a logo for a coffee shop" --out logo.png
  olares-cli router call image "a wide landscape" --size 1792x1024
  olares-cli router call image "a portrait" --aspect-ratio 2:3 --quality high --seed 7
  olares-cli router call image "slow one" --no-wait
  olares-cli router call image --id gen_01H…
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runLegacyMedia(c, f, imageKind, legacyVerb{
				model: callModel(model, categoryImage), id: id, out: out, outputID: outputID,
				wait: !noWait, timeout: timeout, apiKey: apiKey, format: output,
				flags: &flags, args: args,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryImage))
	cmd.Flags().StringVar(&out, "out", "", "write the image here instead of a name derived from the generation")
	cmd.Flags().StringVar(&outputID, "output-id", "", "which of the generation's outputs to write; the first when omitted")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "print the generation id instead of waiting for the image")
	cmd.Flags().StringVar(&id, "id", "", "collect a generation submitted earlier")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "give up waiting after this long; the work continues")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	flags.register(cmd, imageFields...)
	addOutputFlag(cmd, &output)
	return cmd
}

func newCallVideoCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		model    string
		out      string
		outputID string
		noWait   bool
		id       string
		timeout  time.Duration
		apiKey   string
		flags    mediaFlags
	)
	cmd := &cobra.Command{
		Use:   "video [prompt…]",
		Short: "generate a video",
		Long: `Generate a video from a description.

Video generation is always asynchronous — minutes, not seconds — so this submits
the work and then waits, writing the file to --out or to a name derived from the
generation. --no-wait prints the id instead, and "--id <id>" collects it later.

Waiting costs nothing and stopping the wait cancels nothing: the work continues
at the provider either way, and it is billed either way. The generation expires,
and the id stops working when it does.

--operation asks for something other than generating from text: edit, extend,
lip_sync, first_last_frame and the rest. Which inputs an operation reads follows
from the operation — a lip sync needs --audio, an extend needs
--source-generation — and a model that does not declare the one you named is
refused before anything is billed.

--size and --aspect-ratio and --resolution overlap: give --size or one of the
other two, not both.

Video generation needs a model whose mode is video_generation; "olares-cli router
list --mode video_generation" shows the ones that qualify.

Examples:
  olares-cli router call video "a paper plane over a city at dusk"
  olares-cli router call video "waves on a beach" --out waves.mp4 --timeout 20m
  olares-cli router call video "a slow pan" --resolution 1080p --duration 8 --fps 24
  olares-cli router call video --operation lip_sync --image face.png --audio line.wav
  olares-cli router call video "long one" --no-wait
  olares-cli router call video --id gen_01H…
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runLegacyMedia(c, f, videoKind, legacyVerb{
				model: callModel(model, categoryVideo), id: id, out: out, outputID: outputID,
				wait: !noWait, timeout: timeout, apiKey: apiKey, format: output,
				flags: &flags, args: args,
				// An operation names what to do with the inputs, and several of
				// them describe the whole request: extending a clip or syncing
				// a mouth to a recording needs no words.
				inputHint: "name what to work from with --" + flagImage +
					", --" + flagAudioIn + " or --" + flagSource,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryVideo))
	cmd.Flags().StringVar(&out, "out", "", "write the video here instead of a name derived from the generation")
	cmd.Flags().StringVar(&outputID, "output-id", "", "which of the generation's outputs to write; the first when omitted")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "print the generation id instead of waiting for the video")
	cmd.Flags().StringVar(&id, "id", "", "collect a generation submitted earlier")
	cmd.Flags().DurationVar(&timeout, "timeout", 20*time.Minute, "give up waiting after this long; the work continues")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	flags.register(cmd, videoFields...)
	addOutputFlag(cmd, &output)
	return cmd
}

// legacyVerb is what image and video differ in, which past the field list is
// only how long a caller is expected to wait.
type legacyVerb struct {
	model     string
	id        string
	out       string
	outputID  string
	wait      bool
	timeout   time.Duration
	apiKey    string
	format    string
	flags     *mediaFlags
	args      []string
	inputHint string
}

// runLegacyMedia submits on the route the family was released on.
//
// The body is the released spelling rather than the canonical one, and that is
// not a compatibility shim: Router lifts every one of those keys onto the
// canonical field it means before anything is gated or billed, so the same
// request is described either way. What the released image route has that the
// unified one cannot is a synchronous answer for a provider that keeps no
// generations, and moving off it would refuse the most widely installed
// provider there is.
func runLegacyMedia(
	c *cobra.Command, f *cmdutil.Factory, kind mediaKind, verb legacyVerb,
) error {
	opts := mediaOptions{
		Out: verb.out, OutputID: verb.outputID, Wait: verb.wait, Timeout: verb.timeout,
		APIKey: verb.apiKey, OutputIn: verb.format, ID: strings.TrimSpace(verb.id),
	}
	if opts.ID != "" {
		if len(verb.args) > 0 {
			return fmt.Errorf("--id collects a generation that already exists; it takes no prompt")
		}
		return runMedia(c.Context(), f, kind, opts)
	}
	prompt, err := resolvePrompt(c, verb.flags, verb.args, verb.inputHint)
	if err != nil {
		return err
	}
	body, err := verb.flags.legacyBody(c, verb.model, prompt)
	if err != nil {
		return err
	}
	opts.Body = body
	return runMedia(c.Context(), f, kind, opts)
}

// mediaKind is the little that differs between the two verbs: where to submit,
// where to look, and what to call the file.
type mediaKind struct {
	noun       string
	verb       string
	submitPath string
	get        func(id string) string
	content    func(id string) string
	defaultExt string
}

var imageKind = mediaKind{
	noun: "image", verb: "image", submitPath: epImageGenerations,
	get: epImageGeneration, content: epImageGenerationContent, defaultExt: ".png",
}

var videoKind = mediaKind{
	noun: "video", verb: "video", submitPath: epVideos,
	get: epVideo, content: epVideoContent, defaultExt: ".mp4",
}

// The two families with no released route of their own. They create, poll and
// download exactly like the other two — the record is the same and so is the
// content proxy — which is why they are two more rows here rather than a
// surface of their own.
var (
	musicKind = mediaKind{
		noun: "track", verb: "music", submitPath: epGenerations,
		get: epGeneration, content: epGenerationContent, defaultExt: ".mp3",
	}
	model3DKind = mediaKind{
		noun: "model", verb: "3d", submitPath: epGenerations,
		get: epGeneration, content: epGenerationContent, defaultExt: ".glb",
	}
)

func runCallImage(ctx context.Context, f *cmdutil.Factory, opts mediaOptions) error {
	return runMedia(ctx, f, imageKind, opts)
}

func runCallVideo(ctx context.Context, f *cmdutil.Factory, opts mediaOptions) error {
	return runMedia(ctx, f, videoKind, opts)
}

func runMedia(ctx context.Context, f *cmdutil.Factory, kind mediaKind, opts mediaOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(opts.OutputIn)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, opts.APIKey)

	var gen generationView
	switch {
	case opts.ID != "":
		if err := dp.doJSON(ctx, "GET", kind.get(opts.ID), nil, &gen); err != nil {
			return callErr(err)
		}
	default:
		inline, err := submitMedia(ctx, dp, kind, opts, &gen)
		if err != nil {
			return err
		}
		if inline != nil {
			// A Router with no persistent media API answered with the picture
			// itself. There is no id to come back for, so --no-wait has nothing
			// to hand over and the only thing left is to write the file.
			return writeInlineImage(inline, kind, opts, format)
		}
		if !opts.Wait {
			if format == FormatJSON {
				return printJSON(os.Stdout, gen)
			}
			fmt.Printf("submitted as %s\n", gen.ID)
			fmt.Printf("`olares-cli router call %s --id %s` collects it; it expires %s\n",
				kind.verb, gen.ID, gen.ExpiresAt.Local().Format(time.RFC1123))
			return nil
		}
	}

	if !gen.done() && opts.Wait {
		if err := waitForGeneration(ctx, dp, kind, &gen, opts.Timeout, format == FormatTable); err != nil {
			return err
		}
	}
	if gen.failed() {
		return fmt.Errorf("%s %s failed: %s", kind.noun, gen.ID, gen.reason())
	}
	if !gen.done() {
		if format == FormatJSON {
			return printJSON(os.Stdout, gen)
		}
		_, err := fmt.Printf("%s is %s%s; `olares-cli router call %s --id %s` collects it\n",
			gen.ID, nonEmpty(gen.Status), gen.progressNote(), kind.verb, gen.ID)
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, gen)
	}
	return fetchGenerationContent(ctx, dp, kind, &gen, opts)
}

// submitMedia posts the request. A non-nil first result is the picture arriving
// inline, which only happens for an image on a Router that keeps no generations.
func submitMedia(ctx context.Context, dp *routerClient, kind mediaKind, opts mediaOptions, gen *generationView) ([]byte, error) {
	// Router persists a generation only when the caller asks it to. For video
	// the header is redundant and harmless; for an image it is the difference
	// between a record to come back to and a one-shot answer.
	async := dp.withHeader("Prefer", "respond-async")
	err := async.doJSON(ctx, "POST", kind.submitPath, opts.Body, gen)
	if err == nil {
		return nil, nil
	}
	var re *RouterError
	if kind.noun == "image" && errors.As(err, &re) && re.Code == "image_generation_async_reserved" {
		var sync imageSyncResponse
		if serr := dp.doJSON(ctx, "POST", kind.submitPath, opts.Body, &sync); serr != nil {
			return nil, callErr(serr)
		}
		raw, derr := sync.bytes()
		if derr != nil {
			return nil, derr
		}
		return raw, nil
	}
	return nil, callErr(err)
}

// imageSyncResponse is the OpenAI Images answer, which Router forwards as it
// arrives when it is not persisting a generation.
type imageSyncResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
		URL     string `json:"url"`
	} `json:"data"`
}

func (r *imageSyncResponse) bytes() ([]byte, error) {
	if len(r.Data) == 0 {
		return nil, fmt.Errorf("the provider returned no image")
	}
	first := r.Data[0]
	if s := strings.TrimSpace(first.B64JSON); s != "" {
		raw, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("decode the image the provider sent: %w", err)
		}
		return raw, nil
	}
	if s := strings.TrimSpace(first.URL); s != "" {
		return nil, fmt.Errorf("this Router keeps no generations, and the provider answered with a link "+
			"rather than the image: %s\nThe link is the provider's and usually expires; download it "+
			"before it does", s)
	}
	return nil, fmt.Errorf("the provider's answer carried neither an image nor a link to one")
}

func writeInlineImage(raw []byte, kind mediaKind, opts mediaOptions, format Format) error {
	path := strings.TrimSpace(opts.Out)
	if path == "" {
		path = "image-" + time.Now().Format("20060102-150405") + kind.defaultExt
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"path": path, "bytes": len(raw)})
	}
	_, err := fmt.Printf("wrote %s (%s)\nThis Router keeps no generations, so there is no id to come back for.\n",
		path, humanBytes(int64(len(raw))))
	return err
}

func waitForGeneration(ctx context.Context, dp *routerClient, kind mediaKind, gen *generationView,
	timeout time.Duration, verbose bool) error {
	deadline := time.Now().Add(timeout)
	last := ""
	for {
		if verbose {
			note := gen.ID + ": " + nonEmpty(gen.Status) + gen.progressNote()
			if note != last {
				fmt.Fprintln(os.Stderr, note)
				last = note
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s %s is still %s after %s; the provider keeps working on it — "+
				"`olares-cli router call %s --id %s` collects it",
				kind.noun, gen.ID, nonEmpty(gen.Status), timeout, kind.verb, gen.ID)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
		var next generationView
		if err := dp.doJSON(ctx, "GET", kind.get(gen.ID), nil, &next); err != nil {
			return callErr(err)
		}
		*gen = next
		if gen.done() {
			return nil
		}
	}
}

func fetchGenerationContent(ctx context.Context, dp *routerClient, kind mediaKind,
	gen *generationView, opts mediaOptions) error {
	path := kind.content(gen.ID)
	if id := strings.TrimSpace(opts.OutputID); id != "" {
		q := url.Values{}
		q.Set("outputId", id)
		path = withQuery(path, q)
	}
	resp, err := dp.do(ctx, "GET", path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)
		return callErr(dp.formatErr("GET", path, resp.StatusCode, raw))
	}

	target := strings.TrimSpace(opts.Out)
	if target == "" {
		target = gen.ID + extForContentType(resp.Header.Get("Content-Type"), kind.defaultExt)
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create %s: %w", target, err)
	}
	written, copyErr := io.Copy(file, resp.Body)
	closeErr := file.Close()
	if copyErr != nil {
		return fmt.Errorf("write %s: %w", target, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close %s: %w", target, closeErr)
	}
	if _, err := fmt.Printf("wrote %s (%s)\n", target, humanBytes(written)); err != nil {
		return err
	}
	if others := gen.outputIDs(); len(others) > 1 {
		_, err := fmt.Fprintf(os.Stderr, "this generation has %d outputs (%s); --output-id names one\n",
			len(others), strings.Join(others, " "))
		return err
	}
	return nil
}

// extForContentType names the file after what arrived rather than after what was
// asked for. A provider that answered with a JPEG when PNG was requested has
// still answered with a JPEG, and a file named otherwise is a file that opens
// wrong.
func extForContentType(contentType, fallback string) string {
	ct, _, err := mime.ParseMediaType(contentType)
	if err != nil || ct == "" {
		return fallback
	}
	switch ct {
	// Pinned, because mime.ExtensionsByType is ordered by the platform's own
	// database: it answers ".jpe" for image/jpeg on some machines.
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	case "video/x-matroska":
		return ".mkv"
	case "audio/mpeg":
		return ".mp3"
	case "audio/wav", "audio/x-wav":
		return ".wav"
	case "audio/flac":
		return ".flac"
	case "audio/ogg":
		return ".ogg"
	case "model/gltf-binary":
		return ".glb"
	case "model/gltf+json":
		return ".gltf"
	case "model/obj":
		return ".obj"
	case "model/stl":
		return ".stl"
	case "application/zip":
		// Several 3D workflows answer with a bundle: the mesh, its textures
		// and a material file are one download.
		return ".zip"
	}
	if exts, eerr := mime.ExtensionsByType(ct); eerr == nil && len(exts) > 0 {
		return exts[0]
	}
	return fallback
}
