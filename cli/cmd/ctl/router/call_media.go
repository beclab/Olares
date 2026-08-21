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

// POST /v1/images/generations, POST /v1/videos, and the two routes each of them
// leaves behind: GET …/{id} and GET …/{id}/content.
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

// generationView is Router's record of one piece of work. Response is the
// provider's own snapshot with its identifiers stripped: it carries the outputs
// and, depending on the provider, a progress percentage.
type generationView struct {
	ID        string          `json:"id"`
	Object    string          `json:"object"`
	MediaType string          `json:"media_type"`
	Operation string          `json:"operation"`
	Status    string          `json:"status"`
	Model     string          `json:"model"`
	Response  json.RawMessage `json:"response,omitempty"`
	ErrorCode *string         `json:"error_code,omitempty"`
	Error     *string         `json:"error,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	ExpiresAt time.Time       `json:"expires_at"`
}

func (g *generationView) done() bool {
	switch strings.ToLower(g.Status) {
	case "completed", "failed":
		return true
	}
	return false
}

func (g *generationView) failed() bool { return strings.EqualFold(g.Status, "failed") }

// progressNote reads the one field of the snapshot worth showing while waiting.
// Not every provider reports it, and its absence is not worth a word.
func (g *generationView) progressNote() string {
	if len(g.Response) == 0 {
		return ""
	}
	var snap struct {
		Progress *float64 `json:"progress"`
	}
	if json.Unmarshal(g.Response, &snap) != nil || snap.Progress == nil {
		return ""
	}
	return fmt.Sprintf(", %.0f%%", *snap.Progress)
}

// outputIDs are the pieces this generation produced. A provider that made one
// image names it anyway, and asking for a named output is how a caller avoids
// depending on which one happens to be first.
func (g *generationView) outputIDs() []string {
	if len(g.Response) == 0 {
		return nil
	}
	var snap struct {
		Outputs []struct {
			ID string `json:"id"`
		} `json:"outputs"`
	}
	if json.Unmarshal(g.Response, &snap) != nil {
		return nil
	}
	out := make([]string, 0, len(snap.Outputs))
	for _, o := range snap.Outputs {
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
	Prompt   string
	Model    string
	Out      string
	OutputID string
	Wait     bool
	Timeout  time.Duration
	APIKey   string
	OutputIn string
	ID       string
	// Extra carries the fields this verb does not interpret. They reach the
	// provider unchanged, which is what lets a size or an aspect ratio be
	// passed without this CLI knowing the vendor's vocabulary.
	Extra map[string]any
}

func newCallImageCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		model    string
		out      string
		outputID string
		size     string
		noWait   bool
		id       string
		timeout  time.Duration
		apiKey   string
	)
	cmd := &cobra.Command{
		Use:   "image [prompt…]",
		Short: "generate an image",
		Long: `Generate an image from a description.

The image is written to --out, or to a file named after the generation in the
current directory. What comes back is a file rather than a URL: Router holds the
bytes, so the picture does not depend on a provider's link staying alive.

--size is passed to the provider untouched, because the accepted values are the
provider's own — "1024x1024" for one, "2K" for another. A rejected value is
reported by whoever rejected it.

--no-wait prints the generation id and stops; "--id <id>" collects it later, and
also reports one that is still running. A generation expires, and --no-wait says
when: after that the id is gone along with the image.

Image generation needs a model whose mode is image_generation; "olares-cli router
list --mode image_generation" shows the ones that qualify.

Examples:
  olares-cli router call image "a red bicycle in the rain"
  olares-cli router call image "a logo for a coffee shop" --out logo.png
  olares-cli router call image "a wide landscape" --size 1792x1024
  olares-cli router call image "slow one" --no-wait
  olares-cli router call image --id gen_01H…
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			opts := mediaOptions{
				Model: callModel(model, categoryImage), Out: out, OutputID: outputID,
				Wait: !noWait, Timeout: timeout, APIKey: apiKey, OutputIn: output,
				ID: strings.TrimSpace(id), Extra: map[string]any{},
			}
			if s := strings.TrimSpace(size); s != "" {
				opts.Extra["size"] = s
			}
			if opts.ID != "" {
				if len(args) > 0 {
					return fmt.Errorf("--id collects a generation that already exists; it takes no prompt")
				}
			} else {
				prompt, err := readPromptArgs(args, "prompt")
				if err != nil {
					return err
				}
				opts.Prompt = prompt
			}
			return runCallImage(c.Context(), f, opts)
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryImage))
	cmd.Flags().StringVar(&out, "out", "", "write the image here instead of a name derived from the generation")
	cmd.Flags().StringVar(&outputID, "output-id", "", "which of the generation's outputs to write; the first when omitted")
	cmd.Flags().StringVar(&size, "size", "", "the size to ask the provider for, in its own vocabulary")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "print the generation id instead of waiting for the image")
	cmd.Flags().StringVar(&id, "id", "", "collect a generation submitted earlier")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "give up waiting after this long; the work continues")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func newCallVideoCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		model     string
		out       string
		outputID  string
		operation string
		noWait    bool
		id        string
		timeout   time.Duration
		apiKey    string
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

--operation names something other than generating from text, for the providers
that offer it. What each accepts is the provider's own; leaving it off generates.

Video generation needs a model whose mode is video_generation; "olares-cli router
list --mode video_generation" shows the ones that qualify.

Examples:
  olares-cli router call video "a paper plane over a city at dusk"
  olares-cli router call video "waves on a beach" --out waves.mp4 --timeout 20m
  olares-cli router call video "long one" --no-wait
  olares-cli router call video --id gen_01H…
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			opts := mediaOptions{
				Model: callModel(model, categoryVideo), Out: out, OutputID: outputID,
				Wait: !noWait, Timeout: timeout, APIKey: apiKey, OutputIn: output,
				ID: strings.TrimSpace(id), Extra: map[string]any{},
			}
			if s := strings.TrimSpace(operation); s != "" {
				opts.Extra["operation"] = s
			}
			if opts.ID != "" {
				if len(args) > 0 {
					return fmt.Errorf("--id collects a generation that already exists; it takes no prompt")
				}
			} else {
				prompt, err := readPromptArgs(args, "prompt")
				if err != nil {
					return err
				}
				opts.Prompt = prompt
			}
			return runCallVideo(c.Context(), f, opts)
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryVideo))
	cmd.Flags().StringVar(&out, "out", "", "write the video here instead of a name derived from the generation")
	cmd.Flags().StringVar(&outputID, "output-id", "", "which of the generation's outputs to write; the first when omitted")
	cmd.Flags().StringVar(&operation, "operation", "", "what to ask the provider for, in its own vocabulary")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "print the generation id instead of waiting for the video")
	cmd.Flags().StringVar(&id, "id", "", "collect a generation submitted earlier")
	cmd.Flags().DurationVar(&timeout, "timeout", 20*time.Minute, "give up waiting after this long; the work continues")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
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
	dp, err := dataPlane(ctx, pc, opts.APIKey)
	if err != nil {
		return err
	}

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
	body := map[string]any{"model": opts.Model, "prompt": opts.Prompt}
	for k, v := range opts.Extra {
		body[k] = v
	}
	// Router persists a generation only when the caller asks it to. For video
	// the header is redundant and harmless; for an image it is the difference
	// between a record to come back to and a one-shot answer.
	async := dp.withHeader("Prefer", "respond-async")
	err := async.doJSON(ctx, "POST", kind.submitPath, body, gen)
	if err == nil {
		return nil, nil
	}
	var re *RouterError
	if kind.noun == "image" && errors.As(err, &re) && re.Code == "image_generation_async_reserved" {
		var sync imageSyncResponse
		if serr := dp.doJSON(ctx, "POST", kind.submitPath, body, &sync); serr != nil {
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
	}
	if exts, eerr := mime.ExtensionsByType(ct); eerr == nil && len(exts) > 0 {
		return exts[0]
	}
	return fallback
}
