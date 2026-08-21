package router

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router call …` — the data plane.
//
// Everything else in this tree configures Router. These verbs use it: one call
// each to the OpenAI-shaped routes it serves, whether the model behind them is
// a cloud provider or a model application on this machine. Which it is does not
// change the request, and that is the point of the gateway.
//
// A call is metered. It lands in `router usage`, it counts against whatever
// quota covers the credential, and — unlike every read in this tree — it can
// cost money. So each verb names the model it used and, when the answer carries
// one, the token count behind it.

func NewCallCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call",
		Short: "call a model through Router",
		Long: `Send work to a model.

The routes are OpenAI-shaped, so a call reaches a cloud provider and a model
application on this machine the same way. Which one answers depends on the model
name, and nothing else about the request changes.

Leaving --model off names the default category for that kind of work —
"default-chat", "default-stt", and so on. Router chooses what a category answers
with, against what is installed; "olares-cli router route list --kind default" reports where
each one currently stands. A category with nothing behind it is refused rather
than answered by something approximate, so a call with no --model on a fresh
install fails until a model of that kind exists.

Credentials resolve on their own. Inside the cluster the platform supplies the
caller identity and no key is needed; anywhere else this machine keeps one key
in the OS keychain, minting it on first use. --api-key overrides both, and
"olares-cli router key current" says which one is in play.

Subcommands:
  models                the names these verbs accept in --model, as this
                        credential sees them
  chat [prompt]         a chat completion, streamed by default
  embed <text…>         embedding vectors
  rerank <query>        order documents by how well they answer a query
  search <query>        search the web
  scrape <url>          a page as markdown
  translate <text>      translate, detect a language, list the pairs
  image <prompt>        generate an image
  video <prompt>        generate a video
  transcribe <file>     speech to text
  speak <text>          text to speech
  vad <file>            where the speech is in a recording
  diarize <file>        who spoke when
  enhance <file>        a cleaned-up recording
  align <file>          line a transcript up with the audio
  ocr <file>            text out of an image or PDF

Three of these do not answer with their result. Image and video generation can
hand back a receipt to collect from, and OCR always does; each of those verbs
waits for the work by default and takes --no-wait to hand the id over instead.

Every call is metered: it appears in "router usage", counts against the quota on
the credential that made it, and may cost money.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newCallModelsCommand(f))
	cmd.AddCommand(newCallChatCommand(f))
	cmd.AddCommand(newCallResponsesCommand(f))
	cmd.AddCommand(newCallEmbedCommand(f))
	cmd.AddCommand(newCallRerankCommand(f))
	cmd.AddCommand(newCallSearchCommand(f))
	cmd.AddCommand(newCallScrapeCommand(f))
	cmd.AddCommand(newCallTranslateCommand(f))
	cmd.AddCommand(newCallImageCommand(f))
	cmd.AddCommand(newCallVideoCommand(f))
	cmd.AddCommand(newCallTranscribeCommand(f))
	cmd.AddCommand(newCallSpeakCommand(f))
	cmd.AddCommand(newCallVADCommand(f))
	cmd.AddCommand(newCallDiarizeCommand(f))
	cmd.AddCommand(newCallEnhanceCommand(f))
	cmd.AddCommand(newCallAlignCommand(f))
	cmd.AddCommand(newCallOCRCommand(f))
	return cmd
}

// The default category each verb falls back to. Router no longer resolves an
// empty `model` field — a request that names nothing is refused rather than
// guessed at — so "no --model" has to become a name here, and the name is the
// category for that kind of work.
//
// Audio has no single category, and that is the interesting one. One audio mode
// covers recognition, synthesis, voice activity, diarization, enhancement and
// sound effects, and those are six separate engine images: no installed
// application serves the mode, so a `default-audio` would name a model that
// cannot answer most audio requests, with no way for the caller to tell which.
// Each capability is its own category instead.
//
// Translate has a category but no --model flag to reach it: those routes carry
// no model field at all and resolve the default per call.
//
// These literals are copied from Router's own registry, and nothing here can
// check them: a category renamed there turns every `--model`-less call into a
// route that does not exist. `olares-cli router route list --kind default` lists what this
// deployment actually has, and is the thing to compare against.
//
// Router's registry is longer than this list, and the gap is not an oversight
// in either direction. A category is only useful here if a verb can reach the
// endpoint behind it: `default-moderation` has no `/v1/moderations` at all, so
// there is nothing to name it from, and `default-translate` belongs to a verb
// that sends no model. Adding a constant for either would promise a call this
// tree cannot make.
const (
	categoryChat        = "default-chat"
	categoryEmbedding   = "default-embedding"
	categoryRerank      = "default-rerank"
	categorySearch      = "default-search"
	categoryScrape      = "default-scrape"
	categoryImage       = "default-image-generation"
	categoryVideo       = "default-video-generation"
	categoryOCR         = "default-ocr"
	categorySTT         = "default-stt"
	categoryTTS         = "default-tts"
	categoryVAD         = "default-vad"
	categoryDiarization = "default-diar"
	categoryEnhance     = "default-enhance"
	categorySoundFX     = "default-sound-fx"
)

// Alignment has no category of its own. It is served by the same engine base as
// speech recognition rather than by an image of its own, so the recognition
// default is the one that finds a model able to do it.
const categoryAlign = categorySTT

// Sound effects have a category but no verb of their own, because they have no
// endpoint of their own: the engine serving them mounts `/v1/audio/speech`, the
// same path text-to-speech uses, and which of the two a request is depends
// entirely on the model behind it. Router says so itself — its
// audioDefaultCategory maps `/speech` to default-tts and notes that a
// caption-to-audio request cannot be told from a spoken one by path.
//
// So `call speak --sound-fx` is the whole feature: the same request with the
// other default. A separate verb would have been this one with a different
// constant and every flag repeated.

// Responses has no category either, and that one is Router's decision rather
// than a consequence of a shared path: the mode is provider-model-only, and
// routing/category_test.go asserts the absence deliberately so it "cannot be
// quietly reversed". `call responses` therefore requires --model. Inventing
// `default-responses` here would produce exactly the failure the note above
// describes — a route that does not exist, reported as though the model were
// missing.

// callModel is what goes in the request's `model` field: what was asked for, or
// the category for this kind of work.
func callModel(flagValue, category string) string {
	if v := strings.TrimSpace(flagValue); v != "" {
		return v
	}
	return category
}

// modelFlagHelp keeps every verb saying the same thing about the same flag.
func modelFlagHelp(category string) string {
	return "model to use, as <provider>/<model> or a route name; " + category + " when omitted"
}

// readPromptArgs turns positional arguments, or stdin when there are none, into
// one prompt. Reading stdin is what makes these verbs composable with the rest
// of a shell, and a terminal with no argument is a mistake rather than an
// invitation to hang, so it is refused.
func readPromptArgs(args []string, what string) (string, error) {
	if joined := strings.TrimSpace(strings.Join(args, " ")); joined != "" {
		return joined, nil
	}
	if isTerminal(os.Stdin) {
		return "", fmt.Errorf("no %s given; pass it as an argument or pipe it in", what)
	}
	buf, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read %s from stdin: %w", what, err)
	}
	s := strings.TrimSpace(string(buf))
	if s == "" {
		return "", fmt.Errorf("no %s on stdin", what)
	}
	return s, nil
}

// readLines is the other way a verb takes many inputs: one per line, blanks
// dropped. The buffer is raised because a line here can be a whole document.
func readLines(r io.Reader) ([]string, error) {
	var out []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			out = append(out, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read lines from input: %w", err)
	}
	return out, nil
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// dataURL encodes a local file the way the chat schema takes an image: the
// gateway forwards the message as it arrives, so a path meaningful on this
// machine would reach an upstream that cannot open it.
func dataURL(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	if ct == "" {
		ct = http.DetectContentType(raw)
	}
	return "data:" + ct + ";base64," + base64.StdEncoding.EncodeToString(raw), nil
}

// sseLines yields the payload of each `data:` line in an event stream and stops
// at the `[DONE]` sentinel. Chat streaming and install progress both speak SSE,
// but they disagree about framing — install events carry an event name and a
// JSON object per line — so they do not share a reader.
func sseLines(body io.Reader, onData func([]byte) error) error {
	sc := bufio.NewScanner(body)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			return nil
		}
		if err := onData([]byte(payload)); err != nil {
			return err
		}
	}
	return sc.Err()
}
