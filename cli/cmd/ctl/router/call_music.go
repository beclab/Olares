package router

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `router call music …` — a track, and the two things worth doing before
// spending minutes of compute on one.
//
// POST   /v1/music/generations       submit, or repaint an existing recording
// GET    /v1/music/generations/:id   poll
// GET    /v1/music/generations/:id/content
// DELETE /v1/music/generations/:id   stop and settle what ran
// POST   /v1/music/formats, GET /v1/music/formats/:id
// POST   /v1/music/drafts,  GET /v1/music/drafts/:id
//
// Music used to ride /v1/generations with the other families, and the canonical
// body it takes has no room for half of what a track is: a title, the words to
// sing, and whether to sing at all. So these routes take a flat body of their
// own, and refuse a field they do not know rather than dropping it.
//
// The other half is that a music model has a language model in front of the
// audio one, which rewrites a caption and lyrics into what it will actually
// sing. That rewrite used to be invisible — the words came back inside a track
// that took four minutes to make. `music format` runs it on its own, and
// `music draft` runs it from nothing at all, so both are text-speed answers to
// questions that were previously only answerable by generating audio.

// musicRequest is the flat body POST /v1/music/generations takes.
//
// Flat rather than the canonical nesting, and not by preference: this route
// refuses an unknown field, so the canonical body would be rejected wholesale.
// Pointers again mark presence — a model that has no parameter for a field
// refuses it, so "omitted" and "sent as false" are different requests.
type musicRequest struct {
	Model     string `json:"model"`
	Operation string `json:"operation,omitempty"`

	InputAudio     string `json:"input_audio,omitempty"`
	Prompt         string `json:"prompt,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Title          string `json:"title,omitempty"`
	Lyrics         string `json:"lyrics,omitempty"`
	Instrumental   *bool  `json:"instrumental,omitempty"`

	DurationSeconds *float64 `json:"duration_seconds,omitempty"`
	Seed            *int64   `json:"seed,omitempty"`
	OutputFormat    string   `json:"output_format,omitempty"`

	ProviderOptions map[string]json.RawMessage `json:"provider_options,omitempty"`
}

// The two text passes. Both are asynchronous and neither produces audio, so
// they are polled like a generation and rendered like a document.
type musicFormatView struct {
	ID              string `json:"id"`
	Object          string `json:"object"`
	Status          string `json:"status"`
	Model           string `json:"model"`
	DraftPrompt     string `json:"draft_prompt,omitempty"`
	DraftLyrics     string `json:"draft_lyrics,omitempty"`
	EffectivePrompt string `json:"effective_prompt,omitempty"`
	EffectiveLyrics string `json:"effective_lyrics,omitempty"`
	// ConditioningLyrics is the same words spelled for the model rather than
	// for a reader: phoneme hints, section markers, whatever the engine needs
	// to pronounce them. It is separate from the effective lyrics because a
	// player has to show one and sing the other, and a single field made the
	// caller choose which of those to get wrong.
	ConditioningLyrics string         `json:"conditioning_lyrics,omitempty"`
	VocalLanguage      string         `json:"vocal_language,omitempty"`
	Warnings           []string       `json:"warnings"`
	Metrics            map[string]any `json:"metrics"`
	ErrorCode          *string        `json:"error_code,omitempty"`
	Error              *string        `json:"error,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	ExpiresAt          time.Time      `json:"expires_at"`
}

type musicDraftView struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Status string `json:"status"`
	Model  string `json:"model"`
	Brief  string `json:"brief,omitempty"`
	Prompt string `json:"prompt,omitempty"`
	Lyrics string `json:"lyrics,omitempty"`
	// See musicFormatView: the words to show and the words to sing.
	ConditioningLyrics string         `json:"conditioning_lyrics,omitempty"`
	Instrumental       bool           `json:"instrumental"`
	VocalLanguage      string         `json:"vocal_language,omitempty"`
	DurationSeconds    float64        `json:"duration_seconds,omitempty"`
	StylePlan          map[string]any `json:"style_plan,omitempty"`
	Warnings           []string       `json:"warnings"`
	Metrics            map[string]any `json:"metrics"`
	ErrorCode          *string        `json:"error_code,omitempty"`
	Error              *string        `json:"error,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	ExpiresAt          time.Time      `json:"expires_at"`
}

func newCallMusicCommand(f *cmdutil.Factory) *cobra.Command {
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
		Use:   "music [prompt…]",
		Short: "generate a track",
		Long: `Generate music from a description.

The track is written to --out, or to a file named after the generation. Router
holds the bytes, so the file does not depend on a provider's link staying alive.

--lyrics gives the words to sing; --instrumental asks for a track without any;
--title names the piece. A model that cannot honor one of these refuses the
request rather than ignoring the field, which is the point of naming it.

--repaint regenerates part of a recording you already have, given with --audio.
It is the one operation music has besides generating from nothing.

--model is required. Router resolves a default for image and video generation
and none for music: today FlowStudio is the only thing serving it, and a
default would name that one workflow while reading like a choice. "olares-cli
router model list --mode music_generation" lists the names this credential can
send.

--no-wait prints the generation id and stops; "--id <id>" collects it later,
and "music cancel <id>" stops it. A generation expires, and --no-wait says when.

A music model rewrites your caption and lyrics into what it will actually sing.
"music format" shows that rewrite without generating anything, and "music draft"
writes both from a single sentence — two text-speed answers to questions that
otherwise cost a full generation.

Subcommands:
  format   what the model would sing, before it sings it
  draft    a caption and lyrics written from one brief
  cancel   stop a track that is still running

A prompt whose first word is one of those three needs quoting, or the word is
read as the subcommand.

Examples:
  olares-cli router call music "a slow waltz on a rainy afternoon" --model FlowStudio/ace-step
  olares-cli router call music "an upbeat theme" --model FlowStudio/ace-step --duration 30 --instrumental
  olares-cli router call music "a ballad" --model FlowStudio/ace-step --title "Rain" --lyrics "$(cat words.txt)"
  olares-cli router call music "brighter chorus" --model FlowStudio/ace-step --repaint --audio take1.mp3
  olares-cli router call music --id gen_01H… --model FlowStudio/ace-step
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runCallMusic(c, f, musicVerb{
				model: model, id: id, out: out, outputID: outputID,
				wait: !noWait, timeout: timeout, apiKey: apiKey, format: output,
				flags: &flags, args: args,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelRequiredHelp("music_generation"))
	cmd.Flags().StringVar(&out, "out", "", "write the track here instead of a name derived from the generation")
	cmd.Flags().StringVar(&outputID, "output-id", "", "which of the generation's outputs to write; the first when omitted")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "print the generation id instead of waiting for the track")
	cmd.Flags().StringVar(&id, "id", "", "collect a generation submitted earlier")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "give up waiting after this long; the work continues")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	flags.register(cmd, musicFields...)
	addOutputFlag(cmd, &output)
	cmd.AddCommand(newCallMusicFormatCommand(f))
	cmd.AddCommand(newCallMusicDraftCommand(f))
	cmd.AddCommand(newCallMusicAlignmentCommand(f))
	cmd.AddCommand(newCallMusicCancelCommand(f))
	return cmd
}

// musicAlignment is where each line lands in the finished track.
type musicAlignment struct {
	Segments []musicLyricSegment `json:"segments"`
}

type musicLyricSegment struct {
	Text         string  `json:"text"`
	StartSeconds float64 `json:"start_seconds"`
	EndSeconds   float64 `json:"end_seconds"`
}

func newCallMusicAlignmentCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		apiKey string
	)
	cmd := &cobra.Command{
		Use:   "alignment <id>",
		Short: "where each line lands in a finished track",
		Long: `Read the lyric timeline of a track that has already been generated.

A player needs to know when each line is sung — to highlight the current one,
and to seek by tapping one. That is not something a caller can work out from
the lyrics and the duration, because the model decides how the words are laid
over the sections, and it is not in the track either.

Router answers this from the provider and upstream id already sealed onto the
generation, so it goes back to the model that made this track rather than
resolving a model again. Nothing is regenerated and no audio model runs.

Not every track has one. The generation has to have completed, and the
application that made it has to serve the timeline at all — an older one does
not, and says so rather than inventing timings. --lyrics on the generation is
unaffected either way: the words are always there, only their placement is not.

Examples:
  olares-cli router call music alignment gen_01H…
  olares-cli router call music alignment gen_01H… -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCallMusicAlignment(c.Context(), f, args[0], apiKey, output)
		},
	}
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func runCallMusicAlignment(ctx context.Context, f *cmdutil.Factory, id, apiKey, outputRaw string) error {
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
	var alignment musicAlignment
	if err := dataPlane(pc, apiKey).doJSON(ctx, http.MethodGet,
		epMusicLyricsAlignment(strings.TrimSpace(id)), nil, &alignment); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, alignment)
	}
	return renderMusicAlignment(os.Stdout, &alignment)
}

func renderMusicAlignment(w io.Writer, alignment *musicAlignment) error {
	if len(alignment.Segments) == 0 {
		_, err := fmt.Fprintln(w, "This track has no lyric timeline. A completed instrumental has "+
			"nothing to place, and a model application that does not serve timelines returns none.")
		return err
	}
	// Timestamps rather than a table: the point is to read the words in the
	// order they are sung, and a line of lyrics does not fit a cell.
	for i := range alignment.Segments {
		s := &alignment.Segments[i]
		if _, err := fmt.Fprintf(w, "%s  %s\n", musicTimecode(s.StartSeconds), s.Text); err != nil {
			return err
		}
	}
	last := alignment.Segments[len(alignment.Segments)-1]
	_, err := fmt.Fprintf(w, "\n%d lines, through %s. -o json carries the end of each one too.\n",
		len(alignment.Segments), musicTimecode(last.EndSeconds))
	return err
}

// musicTimecode is mm:ss.s — a track is minutes long and a line lands on a
// beat, so seconds alone are too coarse to seek by and a duration string reads
// as an interval rather than a position.
func musicTimecode(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	return fmt.Sprintf("%02d:%04.1f", int(seconds)/60, seconds-float64(int(seconds)/60*60))
}

type musicVerb struct {
	model    string
	id       string
	out      string
	outputID string
	wait     bool
	timeout  time.Duration
	apiKey   string
	format   string
	flags    *mediaFlags
	args     []string
}

func runCallMusic(c *cobra.Command, f *cmdutil.Factory, verb musicVerb) error {
	model := strings.TrimSpace(verb.model)
	if model == "" {
		return fmt.Errorf("--model is required: Router resolves no default for music_generation, so there "+
			"is nothing to fall back to\n`olares-cli router model list --mode %s` lists the names this "+
			"credential can send", "music_generation")
	}
	opts := mediaOptions{
		Out: verb.out, OutputID: verb.outputID, Wait: verb.wait, Timeout: verb.timeout,
		APIKey: verb.apiKey, OutputIn: verb.format, ID: strings.TrimSpace(verb.id),
		Idempotent: true,
	}
	if opts.ID != "" {
		if len(verb.args) > 0 {
			return fmt.Errorf("--id collects a generation that already exists; it takes no prompt")
		}
		return runMedia(c.Context(), f, musicKind, opts)
	}
	hint := ""
	if verb.flags.repaint {
		hint = "say what to change about the --" + flagAudioIn
	}
	prompt, err := resolvePrompt(c, verb.flags, verb.args, hint)
	if err != nil {
		return err
	}
	body, err := musicBody(c, model, prompt, verb.flags)
	if err != nil {
		return err
	}
	opts.Body = body
	return runMedia(c.Context(), f, musicKind, opts)
}

// musicBody spells the request in the keys this route takes.
//
// Only flags the caller gave are sent, for the reason the canonical builder has
// it: the route refuses a field the resolved model has no parameter for, so a
// flag left alone must be absent rather than present and zero.
func musicBody(c *cobra.Command, model, prompt string, m *mediaFlags) (*musicRequest, error) {
	if err := m.check(c); err != nil {
		return nil, err
	}
	given := func(name string) bool { return c.Flags().Changed(name) }
	body := &musicRequest{Model: model, Prompt: prompt}
	if m.repaint {
		if !given(flagAudioIn) {
			return nil, fmt.Errorf("--%s needs the recording to work from: give it with --%s",
				flagRepaint, flagAudioIn)
		}
		body.Operation = "repaint"
	} else if given(flagAudioIn) {
		return nil, fmt.Errorf("--%s is the recording a --%s works from; generating a track from nothing "+
			"takes no input", flagAudioIn, flagRepaint)
	}
	if given(flagAudioIn) {
		audio, err := mediaInput(m.audio, flagAudioIn)
		if err != nil {
			return nil, err
		}
		body.InputAudio = audio
	}
	if given(flagNegative) {
		body.NegativePrompt = m.negative
	}
	if given(flagTitle) {
		body.Title = m.title
	}
	if given(flagLyrics) {
		body.Lyrics = m.lyrics
	}
	if given(flagInstrumental) {
		instrumental := m.instrumental
		body.Instrumental = &instrumental
	}
	if given(flagDuration) {
		duration := m.duration
		body.DurationSeconds = &duration
	}
	if given(flagSeed) {
		seed := m.seed
		body.Seed = &seed
	}
	if given(flagFormat) {
		body.OutputFormat = m.format
	}
	options, err := parseProviderOptions(m.providerOptions)
	if err != nil {
		return nil, err
	}
	body.ProviderOptions = options
	return body, nil
}

func newCallMusicCancelCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		apiKey string
	)
	cmd := &cobra.Command{
		Use:   "cancel <id>",
		Short: "stop a track that is still running",
		Long: `Stop a running generation and settle what it used.

Music is the one family where this is possible: the upstream is told to stop,
and what it did before that is billed. So cancelling is not a refund — it
bounds the cost of a take that is clearly going wrong rather than undoing it.

A generation that has already finished, failed or been cancelled is not
cancellable, and says so rather than pretending. A provider with no way to stop
mid-run is refused too, which is a property of the provider.

Example:
  olares-cli router call music cancel gen_01H…
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCallMusicCancel(c.Context(), f, args[0], apiKey, output)
		},
	}
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func runCallMusicCancel(ctx context.Context, f *cmdutil.Factory, id, apiKey, outputRaw string) error {
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
	var gen generationView
	if err := dataPlane(pc, apiKey).doJSON(ctx, http.MethodDelete,
		epMusicGeneration(strings.TrimSpace(id)), nil, &gen); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, gen)
	}
	_, err = fmt.Printf("%s is %s. What ran before the cancel was still billed; "+
		"`olares-cli router usage list --limit 5` says what it came to.\n", gen.ID, nonEmpty(gen.Status))
	return err
}

func newCallMusicFormatCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		model    string
		lyrics   string
		language string
		duration float64
		temp     float64
		noWait   bool
		id       string
		timeout  time.Duration
		apiKey   string
	)
	cmd := &cobra.Command{
		Use:   "format [caption…]",
		Short: "what the model would sing, before it sings it",
		Long: `Run a music model's text pass without generating any audio.

A music model does not sing the caption and lyrics you send. It rewrites them
first — trimming to the duration, fixing the metre, choosing how the words are
laid over the sections — and until this route existed the only way to see that
rewrite was to wait for the whole track and listen.

The answer carries both: DRAFT is what you sent as the model received it, and
EFFECTIVE is what it would perform. Warnings name what it had to change and
why, which is usually the answer to "why does the second verse not fit".

--lyrics and --vocal-language are required alongside the caption: the rewrite is
of a specific song in a specific language, and there is nothing to normalize
without them.

Text speed rather than generation speed, and priced as the text it is.

Examples:
  olares-cli router call music format "a slow waltz" --model FlowStudio/ace-step \
      --lyrics "$(cat words.txt)" --vocal-language en
  olares-cli router call music format "an anthem" --model FlowStudio/ace-step \
      --lyrics "$(cat words.txt)" --vocal-language zh --duration 180
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			body := map[string]any{}
			if c.Flags().Changed("duration") {
				body["duration_seconds"] = duration
			}
			if c.Flags().Changed("temperature") {
				body["temperature"] = temp
			}
			return runMusicText(c, f, musicTextVerb{
				noun: "format", submit: epMusicFormats, get: epMusicFormat,
				model: model, id: id, wait: !noWait, timeout: timeout, apiKey: apiKey,
				format: output, args: args,
				promptKey: "prompt", promptHint: "the caption to normalize",
				body: body, required: map[string]string{"lyrics": lyrics, "vocal_language": language},
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelRequiredHelp("music_generation"))
	cmd.Flags().StringVar(&lyrics, "lyrics", "", "the words to lay out; required")
	cmd.Flags().StringVar(&language, "vocal-language", "", "the language to sing in, as a short code; required")
	cmd.Flags().Float64Var(&duration, "duration", 0, "the length to fit the words to, in seconds (10-600)")
	cmd.Flags().Float64Var(&temp, "temperature", 0, "how freely to rewrite, 0 to 2")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "print the task id instead of waiting for the rewrite")
	cmd.Flags().StringVar(&id, "id", "", "collect a format task submitted earlier")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "give up waiting after this long; the work continues")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func newCallMusicDraftCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output       string
		model        string
		language     string
		instrumental bool
		temp         float64
		noWait       bool
		id           string
		timeout      time.Duration
		apiKey       string
	)
	cmd := &cobra.Command{
		Use:   "draft [brief…]",
		Short: "a caption and lyrics written from one brief",
		Long: `Have a music model write the song before it performs it.

One sentence in — "a defiant closing credits song about leaving a city" — and
back comes the caption, the lyrics, a tempo and a rough length, in the shape the
generation route takes. It is the step before "call music", not a substitute
for it: nothing is sung here.

--instrumental asks for a piece with no words, in which case what comes back is
a caption and a plan rather than lyrics.

Text speed rather than generation speed, and priced as the text it is.

Examples:
  olares-cli router call music draft "a defiant closing credits song about leaving a city" \
      --model FlowStudio/ace-step --vocal-language en
  olares-cli router call music draft "warm piano under a rainy window" \
      --model FlowStudio/ace-step --instrumental
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			body := map[string]any{}
			if c.Flags().Changed("vocal-language") {
				body["vocal_language"] = language
			}
			if c.Flags().Changed("instrumental") {
				body["instrumental"] = instrumental
			}
			if c.Flags().Changed("temperature") {
				body["temperature"] = temp
			}
			return runMusicText(c, f, musicTextVerb{
				noun: "draft", submit: epMusicDrafts, get: epMusicDraft,
				model: model, id: id, wait: !noWait, timeout: timeout, apiKey: apiKey,
				format: output, args: args,
				promptKey: "brief", promptHint: "one sentence describing the song",
				body: body,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelRequiredHelp("music_generation"))
	cmd.Flags().StringVar(&language, "vocal-language", "", "the language to write the words in, as a short code")
	cmd.Flags().BoolVar(&instrumental, "instrumental", false, "write a piece with no words")
	cmd.Flags().Float64Var(&temp, "temperature", 0, "how freely to write, 0 to 2")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "print the task id instead of waiting for the draft")
	cmd.Flags().StringVar(&id, "id", "", "collect a draft task submitted earlier")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "give up waiting after this long; the work continues")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

// musicTextVerb is what the format and draft passes differ in, which past the
// field names is only what they are called.
type musicTextVerb struct {
	noun       string
	submit     string
	get        func(string) string
	model      string
	id         string
	wait       bool
	timeout    time.Duration
	apiKey     string
	format     string
	args       []string
	promptKey  string
	promptHint string
	body       map[string]any
	required   map[string]string
}

func runMusicText(c *cobra.Command, f *cmdutil.Factory, verb musicTextVerb) error {
	ctx := c.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(verb.format)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, verb.apiKey)

	var task musicTask
	if id := strings.TrimSpace(verb.id); id != "" {
		if len(verb.args) > 0 {
			return fmt.Errorf("--id collects a %s that already exists; it takes no text", verb.noun)
		}
		if err := dp.doJSON(ctx, http.MethodGet, verb.get(id), nil, &task.raw); err != nil {
			return callErr(err)
		}
		task.id = id
	} else {
		model := strings.TrimSpace(verb.model)
		if model == "" {
			return fmt.Errorf("--model is required: Router resolves no default for music_generation, so " +
				"there is nothing to fall back to\n`olares-cli router model list --mode music_generation` " +
				"lists the names this credential can send")
		}
		text, err := readPromptArgs(verb.args, verb.promptHint)
		if err != nil {
			return err
		}
		body := map[string]any{"model": model, verb.promptKey: text}
		for key, value := range verb.body {
			body[key] = value
		}
		for key, value := range verb.required {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("--%s is required: a %s is of a specific song, and there is nothing to "+
					"work on without it", strings.ReplaceAll(key, "_", "-"), verb.noun)
			}
			body[key] = value
		}
		key, err := idempotencyKey()
		if err != nil {
			return err
		}
		if err := dp.withHeader("Idempotency-Key", key).
			doJSON(ctx, http.MethodPost, verb.submit, body, &task.raw); err != nil {
			return callErr(err)
		}
		task.id = task.field("id")
		if !verb.wait {
			if format == FormatJSON {
				return printJSON(os.Stdout, task.raw)
			}
			_, werr := fmt.Printf("submitted as %s\n`olares-cli router call music %s --id %s` collects it\n",
				task.id, verb.noun, task.id)
			return werr
		}
	}

	if verb.wait && !task.done() {
		if err := waitForMusicText(ctx, dp, verb, &task, format == FormatTable); err != nil {
			return err
		}
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, task.raw)
	}
	if task.failed() {
		return fmt.Errorf("%s %s failed: %s", verb.noun, task.id, task.reason())
	}
	if !task.done() {
		_, err := fmt.Printf("%s is %s; `olares-cli router call music %s --id %s` collects it\n",
			task.id, nonEmpty(task.field("status")), verb.noun, task.id)
		return err
	}
	return renderMusicText(os.Stdout, verb.noun, task.raw)
}

// musicTask is a format or draft task held as it arrived.
//
// Held raw because the two views share their lifecycle and nothing else, and
// because -o json should print what Router said rather than the subset this
// build happens to know. The three fields the lifecycle turns on are read by
// name; everything else is rendered generically.
type musicTask struct {
	id  string
	raw map[string]json.RawMessage
}

func (t *musicTask) field(name string) string {
	var s string
	if raw, ok := t.raw[name]; ok {
		_ = json.Unmarshal(raw, &s)
	}
	return s
}

func (t *musicTask) done() bool {
	switch strings.ToLower(t.field("status")) {
	case "completed", "failed", "canceled":
		return true
	}
	return false
}

func (t *musicTask) failed() bool {
	status := strings.ToLower(t.field("status"))
	return status == "failed" || status == "canceled"
}

func (t *musicTask) reason() string {
	if message := t.field("error"); message != "" {
		return message
	}
	if code := t.field("error_code"); code != "" {
		return code
	}
	return "no reason was given"
}

func waitForMusicText(ctx context.Context, dp *routerClient, verb musicTextVerb, task *musicTask, chatty bool) error {
	if chatty {
		fmt.Fprintf(os.Stderr, "waiting for the %s…\n", verb.noun)
	}
	deadline := time.Now().Add(verb.timeout)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
		if err := dp.doJSON(ctx, http.MethodGet, verb.get(task.id), nil, &task.raw); err != nil {
			return callErr(err)
		}
		if task.done() {
			return nil
		}
		if time.Now().After(deadline) {
			// The work is Router's and outlives this command, so this is a
			// stopped wait rather than a stopped task.
			return nil
		}
	}
}

// renderMusicText prints the text a pass produced.
//
// Long values go under their own heading rather than into a table cell: lyrics
// are the point of both of these verbs and a clipped cell would make the answer
// unreadable to save a line.
func renderMusicText(w io.Writer, noun string, raw map[string]json.RawMessage) error {
	short := []string{"id", "status", "model", "vocal_language", "instrumental", "duration_seconds"}
	long := []string{"brief", "prompt", "draft_prompt", "effective_prompt",
		"lyrics", "draft_lyrics", "effective_lyrics", "conditioning_lyrics"}

	t := newTable(w)
	for _, name := range short {
		if value := scalarString(raw[name]); value != "" {
			t.row(strings.ToUpper(strings.ReplaceAll(name, "_", " ")), value)
		}
	}
	if err := t.flush(); err != nil {
		return err
	}
	for _, name := range long {
		value := scalarString(raw[name])
		if strings.TrimSpace(value) == "" {
			continue
		}
		if _, err := fmt.Fprintf(w, "\n%s:\n%s\n", strings.ToUpper(strings.ReplaceAll(name, "_", " ")), value); err != nil {
			return err
		}
	}
	var warnings []string
	if raw["warnings"] != nil {
		_ = json.Unmarshal(raw["warnings"], &warnings)
	}
	if len(warnings) > 0 {
		if _, err := fmt.Fprintf(w, "\nWhat the model changed, and why:\n"); err != nil {
			return err
		}
		for _, warning := range warnings {
			if _, err := fmt.Fprintf(w, "  - %s\n", warning); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintf(w, "\nNothing was generated: this is the %s, not a track. "+
		"`olares-cli router call music` performs it.\n", noun)
	return err
}

// scalarString renders one JSON value for a table cell. A string prints as
// itself rather than quoted; anything else prints as it arrived, so a field
// this build has never heard of is still shown.
func scalarString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return strings.TrimSpace(string(raw))
}

// idempotencyKey mints one key per submit.
//
// Router has no idempotency on most of its writes, and on these routes it does:
// a submit whose answer was lost can be sent again and answered with the track
// the first attempt started, rather than starting a second one. That only works
// if the key is generated here, once, before the first attempt.
func idempotencyKey() (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("could not generate an idempotency key: %w", err)
	}
	return "olares-cli-" + hex.EncodeToString(buf[:]), nil
}
