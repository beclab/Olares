package router

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `router call voice …` and `router call history …` — the durable half of
// speech synthesis.
//
// GET    /v1/voices, /v1/voices/:id, /v1/voices/:id/settings
// GET    /v1/voices/settings/default
// POST   /v1/voices/add                 a voice made from a recording
// DELETE /v1/voices/:id
// POST   /v1/text-to-voice/design       a voice imagined from a description
// POST   /v1/text-to-voice              keep one of the designs
// GET    /v1/history, /v1/history/:id, /v1/history/:id/audio
// DELETE /v1/history/:id
//
// `call speak` and `call clone` synthesize and forget: a reference clip is used
// for one reading and is not a voice afterwards. These are the other half — a
// voice that persists under a name and can be spoken with again, and a log of
// every reading with the audio still attached.
//
// Three different defaults sit behind them and the difference is capability
// rather than tidiness. Reading and editing the table is `default-tts`, because
// a clone and a design land in that table too and an engine with no listable
// voices of its own still has those. Making a voice from a recording is
// `default-tts-clone` and making one from a description is `default-tts-design`
// — an engine that cannot do either has nothing to do with the input, so
// sending it there would trade a clear "no such default" for the engine's own
// confusion.
//
// History reaches no default at all: Router has no category for it, so every
// history verb takes --model. That is not an oversight either — a reading
// belongs to the engine that performed it, and there is no sensible engine to
// resolve "the history" to.

func newCallVoiceCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voice",
		Short: "the voices a synthesis model can speak with",
		Long: `Manage the voice library of a speech model.

A voice here outlives the request that made it. "call speak" and "call clone"
produce audio and keep nothing: a reference recording is used once and is not a
voice afterwards. These verbs build the table that "call speak --voice <id>"
then speaks from.

There are two ways to make one, and they are different capabilities rather than
two spellings. "voice add" needs a recording of somebody and needs the model to
declare cloning. "voice design" needs only a description — "a tired night-shift
radio host" — and needs the model to declare design; it answers with previews
to listen to, and "--save <name>" keeps the one you picked.

Subcommands:
  list                    every voice this model can speak with
  get <voice>             one voice and its settings
  add <name>              make a voice from a recording
  design <description>    imagine a voice, and keep it if it is right
  delete <voice>          remove one
  settings [<voice>]      the settings a voice speaks with, or the defaults
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newVoiceListCommand(f))
	cmd.AddCommand(newVoiceGetCommand(f))
	cmd.AddCommand(newVoiceAddCommand(f))
	cmd.AddCommand(newVoiceDesignCommand(f))
	cmd.AddCommand(newVoiceDeleteCommand(f))
	cmd.AddCommand(newVoiceSettingsCommand(f))
	return cmd
}

// voice is one row of the library, in the shape the synthesis engines answer
// with. Settings are held raw because they are the engine's own knobs and this
// build has no business deciding which of them are worth showing.
type voice struct {
	ID          string            `json:"voice_id"`
	Name        string            `json:"name"`
	Category    string            `json:"category"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels,omitempty"`
	Settings    json.RawMessage   `json:"settings,omitempty"`
}

// keptVoice is what a route that keeps a voice answers with, which is either the
// voice or a receipt for the task that is making it.
//
// One shape reads both because the route does not say in advance which it will
// be, and nothing the caller sends decides: making a voice is minutes of work on
// some engines and instant on others, so `voices/add` and `text-to-voice` each
// answer inline or with a task depending on who is behind the category. The
// discriminator is the receipt's own key rather than the status code — these
// verbs never ask for a task, so a 202 is the engine's choice to report, not the
// caller's to interpret.
type keptVoice struct {
	voice
	Task *audioTask `json:"task,omitempty"`
}

// keep resolves an answer to the voice it made, waiting out the task when the
// answer was a receipt.
//
// Read as a voice, a receipt is a voice with no id: the command reports a name
// that nothing can be spoken with, and — because a submission opens a spend row
// the settling lookup is supposed to close — Router is left holding a row for
// work nobody ever came back for.
func (k keptVoice) keep(ctx context.Context, dp *routerClient, model string,
	timeout time.Duration, verbose bool) (voice, error) {
	if k.Task == nil || strings.TrimSpace(k.Task.ID) == "" {
		return k.voice, nil
	}
	task := k.Task
	if !task.settled() {
		if err := waitForAudioTask(ctx, dp, task, model, timeout, verbose); err != nil {
			return voice{}, err
		}
	}
	if !strings.EqualFold(task.Status, "succeeded") {
		if task.Error != nil && strings.TrimSpace(task.Error.Message) != "" {
			return voice{}, fmt.Errorf("the voice was not kept: %s", task.Error.Message)
		}
		return voice{}, fmt.Errorf("the voice was not kept: task %s ended %s",
			task.ID, nonEmpty(task.Status))
	}
	if len(task.Result) == 0 {
		return voice{}, fmt.Errorf("task %s finished without naming the voice it made; "+
			"`olares-cli router call voice list --model %s` shows whether it is there",
			task.ID, nonEmpty(model))
	}
	var made voice
	if err := json.Unmarshal(task.Result, &made); err != nil {
		return voice{}, fmt.Errorf("read the voice task %s made: %w", task.ID, err)
	}
	return made, nil
}

func newVoiceListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		model  string
		apiKey string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "every voice this model can speak with",
		Long: `List the voices in a synthesis model's library.

CATEGORY is where a voice came from: "premade" ships with the weights, "cloned"
was made from a recording, "generated" was designed from a description. All
three are spoken with the same way.

An empty list is an answer rather than a fault: a model built to speak from a
reference recording has no library to list, and "call clone" is how it is used.

Leaving --model off resolves default-tts, unlike "call history", which refuses.
A library is a property of whichever engine would speak, so the default is the
right guess here; a past reading may belong to another engine entirely.

Examples:
  olares-cli router call voice list
  olares-cli router call voice list --model Olares/<tts-model>
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runVoiceList(c.Context(), f, callModel(model, categoryTTS), apiKey, output)
		},
	}
	addVoiceModelFlag(cmd, &model, categoryTTS)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func runVoiceList(ctx context.Context, f *cmdutil.Factory, model, apiKey, outputRaw string) error {
	format, dp, err := voicePlane(ctx, f, outputRaw, apiKey)
	if err != nil {
		return err
	}
	var resp struct {
		Voices []voice `json:"voices"`
	}
	if err := dp.doJSON(voiceCtx(ctx), http.MethodGet, voicePath(epVoices, model), nil, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	if len(resp.Voices) == 0 {
		_, err := fmt.Println("this model has no voice library. It speaks from a reference recording " +
			"instead, which is what `olares-cli router call clone` sends.")
		return err
	}
	t := newTable(os.Stdout, "VOICE", "NAME", "CATEGORY", "DESCRIPTION")
	for i := range resp.Voices {
		v := &resp.Voices[i]
		t.row(v.ID, nonEmpty(v.Name), nonEmpty(v.Category), clip(v.Description, 48))
	}
	if err := t.flush(); err != nil {
		return err
	}
	_, err = fmt.Println("\n`olares-cli router call speak \"…\" --voice <voice>` speaks with one of these.")
	return err
}

func newVoiceGetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		model  string
		apiKey string
	)
	cmd := &cobra.Command{
		Use:   "get <voice>",
		Short: "one voice and its settings",
		Long: `Show one voice: where it came from, and the settings it speaks with.

Settings are the engine's own knobs and are printed as the engine states them.
"voice settings" with no argument shows the defaults a voice falls back to.

Example:
  olares-cli router call voice get vc_01H…
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runVoiceGet(c.Context(), f, args[0], callModel(model, categoryTTS), apiKey, output)
		},
	}
	addVoiceModelFlag(cmd, &model, categoryTTS)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func runVoiceGet(ctx context.Context, f *cmdutil.Factory, id, model, apiKey, outputRaw string) error {
	format, dp, err := voicePlane(ctx, f, outputRaw, apiKey)
	if err != nil {
		return err
	}
	var v voice
	if err := dp.doJSON(voiceCtx(ctx), http.MethodGet,
		voicePath(epVoice(strings.TrimSpace(id)), model), nil, &v); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, v)
	}
	t := newTable(os.Stdout)
	t.row("VOICE", v.ID)
	t.row("NAME", nonEmpty(v.Name))
	t.row("CATEGORY", nonEmpty(v.Category))
	if v.Description != "" {
		t.row("DESCRIPTION", v.Description)
	}
	for key, value := range v.Labels {
		t.row(strings.ToUpper(key), value)
	}
	if err := t.flush(); err != nil {
		return err
	}
	if len(v.Settings) > 0 {
		if _, err := fmt.Println("\nSETTINGS:"); err != nil {
			return err
		}
		return printJSON(os.Stdout, json.RawMessage(v.Settings))
	}
	return nil
}

func newVoiceAddCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output      string
		model       string
		apiKey      string
		sample      string
		description string
		refText     string
		timeout     time.Duration
	)
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "make a voice from a recording",
		Long: `Keep a recording as a named voice.

This is "call clone" made durable. Cloning speaks one line with a reference
recording and keeps nothing; this puts the same recording in the model's
library under a name, so "call speak --voice <id>" can use it again without
re-uploading anything.

--sample is the recording. --ref-text is what is said in it, which some engines
use to align the clone and all of them ignore harmlessly when they do not.

Some engines store the voice as they answer and some run it as a job; either
way this waits and prints the id the voice actually has. --timeout bounds the
wait, and a timeout leaves the job running rather than losing it.

Needs a model declaring supports_tts_clone; without --model this resolves
default-tts-clone, and a machine with no cloning model says so rather than
sending the recording somewhere that cannot use it.

Example:
  olares-cli router call voice add "Night Host" --sample me.wav --ref-text "the words in the clip"
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if strings.TrimSpace(sample) == "" {
				return fmt.Errorf("--sample is required: a voice made from a recording needs the recording")
			}
			return runVoiceAdd(c.Context(), f, args[0], sample, description, refText,
				callModel(model, categoryTTSClone), apiKey, output, timeout)
		},
	}
	addVoiceModelFlag(cmd, &model, categoryTTSClone)
	cmd.Flags().StringVar(&sample, "sample", "", "the recording to keep as a voice")
	cmd.Flags().StringVar(&description, "description", "", "what this voice is, for the person reading the list")
	cmd.Flags().StringVar(&refText, "ref-text", "", "what is said in the recording")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute,
		"give up waiting after this long; the model keeps making the voice")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func runVoiceAdd(ctx context.Context, f *cmdutil.Factory, name, sample, description, refText,
	model, apiKey, outputRaw string, timeout time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if err := checkAudioUploadSize(sample, os.Stderr); err != nil {
		return err
	}
	pc, err := prepareLongRequest(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, apiKey)
	body, contentType, err := multipartFile(sample, "file", map[string]string{
		"name":        strings.TrimSpace(name),
		"description": strings.TrimSpace(description),
		"ref_text":    strings.TrimSpace(refText),
	})
	if err != nil {
		return err
	}
	route := voicePath(epVoicesAdd, model)
	resp, err := dp.do(ctx, http.MethodPost, route, body, contentType)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read the answer: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return callErr(dp.formatErr(http.MethodPost, route, resp.StatusCode, raw))
	}
	var answer keptVoice
	if err := json.Unmarshal(raw, &answer); err != nil {
		return fmt.Errorf("read the answer to %s: %w (body=%s)",
			route, err, truncate(string(raw), 200))
	}
	made, err := answer.keep(ctx, dp, model, timeout, format == FormatTable)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, made)
	}
	_, err = fmt.Printf("kept as %s. `olares-cli router call speak \"…\" --voice %s` speaks with it.\n",
		nonEmpty(made.ID), nonEmpty(made.ID))
	return err
}

func newVoiceDesignCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output  string
		model   string
		apiKey  string
		text    string
		save    string
		outPath string
		timeout time.Duration
	)
	cmd := &cobra.Command{
		Use:   "design <description…>",
		Short: "imagine a voice, and keep it if it is right",
		Long: `Make a voice out of a description rather than a recording.

"a tired night-shift radio host, warm, slightly hoarse" comes back as one or
more previews: short samples of what that voice sounds like reading a line.
Nothing is kept until you say so.

--text is the line to read in the preview, which is worth setting when the
voice is for a specific piece of writing. --out writes the first preview to a
file to listen to. --save <name> keeps it: the preview becomes a real voice in
the library, with an id "call speak --voice" takes.

Needs a model declaring supports_tts_design; without --model this resolves
default-tts-design.

Examples:
  olares-cli router call voice design "a tired night-shift radio host" --out preview.wav
  olares-cli router call voice design "a bright children's narrator" --save "Narrator"
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			description, err := readPromptArgs(args, "description of the voice")
			if err != nil {
				return err
			}
			return runVoiceDesign(c.Context(), f, description, text, save, outPath,
				callModel(model, categoryTTSDesign), apiKey, output, timeout)
		},
	}
	addVoiceModelFlag(cmd, &model, categoryTTSDesign)
	cmd.Flags().StringVar(&text, "text", "", "the line the preview should read")
	cmd.Flags().StringVar(&save, "save", "", "keep the first preview under this name")
	cmd.Flags().StringVar(&outPath, "out", "", "write the first preview here")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute,
		"give up waiting for --save after this long; the model keeps keeping the voice")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

// voicePreview is one candidate. The audio arrives inline because a preview is
// a few seconds long and exists to be judged, not stored: there is no id to
// come back for until one is kept.
type voicePreview struct {
	GeneratedVoiceID string  `json:"generated_voice_id"`
	AudioBase64      string  `json:"audio_base_64"`
	MediaType        string  `json:"media_type"`
	DurationSecs     float64 `json:"duration_secs"`
}

func runVoiceDesign(ctx context.Context, f *cmdutil.Factory, description, text, save, outPath,
	model, apiKey, outputRaw string, timeout time.Duration) error {
	format, dp, err := voicePlane(ctx, f, outputRaw, apiKey)
	if err != nil {
		return err
	}
	ctx = voiceCtx(ctx)
	body := map[string]any{"voice_description": description}
	if s := strings.TrimSpace(text); s != "" {
		body["text"] = s
	}
	var designed struct {
		Previews []voicePreview `json:"previews"`
	}
	if err := dp.doJSON(ctx, http.MethodPost, voicePath(epTextToVoiceDesign, model), body, &designed); err != nil {
		return callErr(err)
	}
	if len(designed.Previews) == 0 {
		return fmt.Errorf("the model answered with no previews, so there is nothing to listen to or keep")
	}
	first := designed.Previews[0]

	if p := strings.TrimSpace(outPath); p != "" {
		audio, derr := base64.StdEncoding.DecodeString(first.AudioBase64)
		if derr != nil {
			return fmt.Errorf("the preview audio did not decode: %w", derr)
		}
		if werr := os.WriteFile(p, audio, 0o644); werr != nil {
			return fmt.Errorf("write %s: %w", p, werr)
		}
	}

	if name := strings.TrimSpace(save); name != "" {
		var answer keptVoice
		if err := dp.doJSON(ctx, http.MethodPost, voicePath(epTextToVoice, model), map[string]any{
			"generated_voice_id": first.GeneratedVoiceID,
			"voice_name":         name,
			"voice_description":  description,
		}, &answer); err != nil {
			return callErr(err)
		}
		kept, err := answer.keep(ctx, dp, model, timeout, format == FormatTable)
		if err != nil {
			return err
		}
		if format == FormatJSON {
			return printJSON(os.Stdout, kept)
		}
		_, err = fmt.Printf("kept as %s. `olares-cli router call speak \"…\" --voice %s` speaks with it.\n",
			nonEmpty(kept.ID), nonEmpty(kept.ID))
		return err
	}

	if format == FormatJSON {
		return printJSON(os.Stdout, designed)
	}
	t := newTable(os.Stdout, "PREVIEW", "SECONDS", "TYPE")
	for i := range designed.Previews {
		p := &designed.Previews[i]
		t.row(p.GeneratedVoiceID, strconv.FormatFloat(p.DurationSecs, 'f', -1, 64), nonEmpty(p.MediaType))
	}
	if err := t.flush(); err != nil {
		return err
	}
	if p := strings.TrimSpace(outPath); p != "" {
		if _, err := fmt.Printf("\nwrote the first preview to %s\n", p); err != nil {
			return err
		}
	}
	_, err = fmt.Println("\nNothing is kept yet. Run the same description again with `--save <name>` to " +
		"put the voice in the library; a preview is not addressable afterwards.")
	return err
}

func newVoiceDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		model  string
		apiKey string
		yes    bool
	)
	cmd := &cobra.Command{
		Use:   "delete <voice>",
		Short: "remove a voice from the library",
		Long: `Delete a voice.

Only a voice that was made — cloned or designed — can be removed; one that
ships with the weights is part of the model. Readings already in the history
keep their audio, since the recording is what was produced rather than a
reference to the voice.

Example:
  olares-cli router call voice delete vc_01H… --yes
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runVoiceDelete(c.Context(), f, args[0], callModel(model, categoryTTS), apiKey, yes)
		},
	}
	addVoiceModelFlag(cmd, &model, categoryTTS)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addConfirmFlag(cmd, &yes)
	return cmd
}

func runVoiceDelete(ctx context.Context, f *cmdutil.Factory, id, model, apiKey string, yes bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if !yes {
		if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin, fmt.Sprintf(
			"Delete voice %s? Anything that names it stops speaking with it.",
			strings.TrimSpace(id)), false); err != nil {
			return err
		}
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var ignored json.RawMessage
	if err := dataPlane(pc, apiKey).doJSON(ctx, http.MethodDelete,
		voicePath(epVoice(strings.TrimSpace(id)), model), nil, &ignored); err != nil {
		return callErr(err)
	}
	_, err = fmt.Printf("%s is gone.\n", strings.TrimSpace(id))
	return err
}

func newVoiceSettingsCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		model  string
		apiKey string
	)
	cmd := &cobra.Command{
		Use:   "settings [voice]",
		Short: "the settings a voice speaks with, or the defaults",
		Long: `Show the synthesis settings.

With a voice, the settings that voice speaks with. Without one, the defaults
every voice falls back to. They are the engine's own knobs — stability,
similarity, speed, whatever this engine has — and are printed as it states
them, because nothing here knows which of them matter for a given model.

Examples:
  olares-cli router call voice settings
  olares-cli router call voice settings vc_01H…
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := ""
			if len(args) == 1 {
				id = args[0]
			}
			return runVoiceSettings(c.Context(), f, id, callModel(model, categoryTTS), apiKey, output)
		},
	}
	addVoiceModelFlag(cmd, &model, categoryTTS)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func runVoiceSettings(ctx context.Context, f *cmdutil.Factory, id, model, apiKey, outputRaw string) error {
	_, dp, err := voicePlane(ctx, f, outputRaw, apiKey)
	if err != nil {
		return err
	}
	route := epVoiceSettings
	if s := strings.TrimSpace(id); s != "" {
		route = epVoiceOwnSettings(s)
	}
	var settings json.RawMessage
	if err := dp.doJSON(voiceCtx(ctx), http.MethodGet, voicePath(route, model), nil, &settings); err != nil {
		return callErr(err)
	}
	return printJSON(os.Stdout, settings)
}

// addVoiceModelFlag names the category the verb falls back to. It is a
// parameter rather than one constant for the noun because the three are not
// interchangeable: reading the table is `default-tts`, and the two ways to add
// to it each need a capability the other model may not declare.
func addVoiceModelFlag(cmd *cobra.Command, model *string, category string) {
	cmd.Flags().StringVar(model, "model", "", modelFlagHelp(category))
}

// voicePath puts the model on the query, which is where the audio passthrough
// reads it. A body would not do: several of these routes have no body at all.
//
// The category is always sent rather than left for Router to imply. Its gate
// table covers most of these suffixes and not all — `/v1/voices/settings/default`
// is one it does not — and a request that arrives with nothing to resolve is
// refused for a missing model field, which reads as a CLI that forgot rather
// than as a route with no default.
func voicePath(route, model string) string {
	if m := strings.TrimSpace(model); m != "" {
		q := url.Values{}
		q.Set("model", m)
		return withQuery(route, q)
	}
	return route
}

func voiceCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func voicePlane(ctx context.Context, f *cmdutil.Factory, outputRaw, apiKey string) (Format, *routerClient, error) {
	format, err := parseFormat(outputRaw)
	if err != nil {
		return format, nil, err
	}
	pc, err := prepare(voiceCtx(ctx), f)
	if err != nil {
		return format, nil, err
	}
	return format, dataPlane(pc, apiKey), nil
}
