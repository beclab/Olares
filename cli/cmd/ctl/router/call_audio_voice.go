package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/audio/speech/clone, POST /v1/audio/speech (dialogue shape)
//
// Two ways of synthesising speech that a preset voice cannot do, and both need
// a recording of the voice they are imitating. They are separate verbs from
// `speak` because they are separate applications: `audioqwen3ttsclonev3`
// declares tts_clone and `audiosoulxdialogv3` declares tts_dialogue, and
// neither declares plain tts — so a call that resolved the synthesis default
// would reach an engine with no cloning route at all.
//
// Dialogue shares its path with `speak`, which is why it cannot be a flag on
// it: /v1/audio/speech carries a completely different body here, and the model
// is the only thing that decides which body the engine will accept. Router
// cannot tell them apart by path, so `speak` and `dialogue` resolve different
// categories and that is the whole distinction.

func newCallCloneCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		model    string
		refText  string
		language string
		outPath  string
		respFmt  string
		async    bool
		apiKey   string
		output   string
	)
	cmd := &cobra.Command{
		Use:   "clone <reference-audio> [text]",
		Short: "text to speech in a voice from a recording",
		Long: `Synthesise speech in the voice of a reference recording.

The recording is a sample of the voice to imitate — a few seconds of clear
speech is enough — and the text is what that voice should say. The text is the
second argument or standard input.

--ref-text is the transcript of the reference recording, and it is worth
supplying: the model conditions on the pair, and without it the voice comes out
noticeably less like the original. Being wrong about it degrades the voice
silently, so an approximate transcript is worse than none.

The answer is audio, so it goes to --out or to standard output when that is a
pipe. Writing audio bytes into a terminal is refused rather than done.

This needs a model that declares voice cloning, which is a different application
from the one that speaks preset voices. Leaving --model off finds one. Asking
such a model for its voice menu returns nothing: there is no menu to pick from,
the recording is the voice.

Examples:
  olares-cli router call clone me.wav "your build finished" --out done.wav
  olares-cli router call clone me.wav "hello" --ref-text "this is my voice" --out hi.wav
  echo "read this in my voice" | olares-cli router call clone me.wav --out out.wav
  olares-cli router call clone me.wav "a long script" --async
`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(c *cobra.Command, args []string) error {
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			text, err := readPromptArgs(args[1:], "text")
			if err != nil {
				return err
			}
			return runCallClone(c.Context(), f, args[0], cloneOptions{
				Text:       text,
				Model:      callModel(model, categoryTTSClone),
				RefText:    refText,
				Language:   language,
				Out:        outPath,
				RespFormat: respFmt,
				Async:      async,
				APIKey:     apiKey,
				Format:     format,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryTTSClone))
	cmd.Flags().StringVar(&refText, "ref-text", "", "the verbatim transcript of the reference recording")
	cmd.Flags().StringVar(&language, "language", "", "language of the text, if the engine takes one")
	cmd.Flags().StringVar(&outPath, "out", "", "write the audio here instead of standard output")
	cmd.Flags().StringVar(&respFmt, "response-format", "", "container format, e.g. wav or mp3")
	cmd.Flags().BoolVar(&async, "async", false, audioAsyncFlagUsage)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

type cloneOptions struct {
	Text       string
	Model      string
	RefText    string
	Language   string
	Out        string
	RespFormat string
	Async      bool
	APIKey     string
	Format     Format
}

func runCallClone(ctx context.Context, f *cmdutil.Factory, refAudio string, opts cloneOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if !opts.Async && strings.TrimSpace(opts.Out) == "" && isTerminal(os.Stdout) {
		return fmt.Errorf("audio would be written to the terminal; name a file with --out, " +
			"pipe the output, or submit it with --async")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, opts.APIKey)

	fields := map[string]string{
		"model":           strings.TrimSpace(opts.Model),
		"input":           opts.Text,
		"ref_text":        strings.TrimSpace(opts.RefText),
		"language":        strings.TrimSpace(opts.Language),
		"response_format": strings.TrimSpace(opts.RespFormat),
	}
	if opts.Async {
		fields[audioAsyncFormField] = "1"
	}
	body, contentType, err := multipartFile(refAudio, "file", fields)
	if err != nil {
		return err
	}
	return streamAudioAnswer(ctx, dp, audioAnswer{
		Method: "POST", Route: epAudioSpeechClone, Body: body, ContentType: contentType,
		Model: opts.Model, Out: opts.Out, Async: opts.Async, Format: opts.Format,
	})
}

// dialogueScript is what a script file holds. It is the engine's own body
// rather than a shape invented here, with one change: `ref_audio` may name a
// local file, because a path is what somebody writing a script has and a
// base64 data URL is what the engine needs.
type dialogueScript struct {
	Speakers []struct {
		RefAudio string `json:"ref_audio"`
		RefText  string `json:"ref_text"`
	} `json:"speakers"`
	Turns []struct {
		Speaker int    `json:"speaker"`
		Text    string `json:"text"`
	} `json:"turns"`
}

func newCallDialogueCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		model   string
		outPath string
		respFmt string
		perTurn bool
		seed    int
		async   bool
		apiKey  string
		output  string
	)
	cmd := &cobra.Command{
		Use:   "dialogue <script-file>",
		Short: "a multi-speaker conversation from a script",
		Long: `Synthesise a conversation between several voices.

The script is a JSON file naming the speakers and what each one says:

  {
    "speakers": [
      {"ref_audio": "alice.wav", "ref_text": "what alice actually says in that clip"},
      {"ref_audio": "bob.wav",   "ref_text": "what bob actually says in that clip"}
    ],
    "turns": [
      {"speaker": 0, "text": "did the build pass?"},
      {"speaker": 1, "text": "it did, eventually."}
    ]
  }

Every speaker needs a reference recording: this is a cloning model with no preset
voices at all, so there is no voice to fall back on. "ref_audio" may be a path
to a local file — it is read and encoded on the way out — or a data URL already.
"ref_text" must be the verbatim transcript of that clip.

The whole script is synthesised in one pass, and that is the reason to use this
rather than calling "clone" once per line: each turn is conditioned on the ones
before it, which is what makes the turn-taking sound like a conversation instead
of a list of sentences. --per-turn splits the answer into one clip per turn as
JSON, without giving up that conditioning.

At most four speakers, which is the engine's limit rather than this verb's.

Examples:
  olares-cli router call dialogue podcast.json --out episode.wav
  olares-cli router call dialogue podcast.json --per-turn -o json > turns.json
  olares-cli router call dialogue long-script.json --async
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			opts := dialogueOptions{
				Model:      callModel(model, categoryTTSDialogue),
				Out:        outPath,
				RespFormat: respFmt,
				PerTurn:    perTurn,
				Async:      async,
				APIKey:     apiKey,
				Format:     format,
			}
			if c.Flags().Changed("seed") {
				opts.Seed = &seed
			}
			return runCallDialogue(c.Context(), f, args[0], opts)
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryTTSDialogue))
	cmd.Flags().StringVar(&outPath, "out", "", "write the audio here instead of standard output")
	cmd.Flags().StringVar(&respFmt, "response-format", "", "container format, e.g. wav, mp3 or flac")
	cmd.Flags().BoolVar(&perTurn, "per-turn", false, "answer with one clip per turn as JSON rather than one recording")
	cmd.Flags().IntVar(&seed, "seed", 0, "make the sampling reproducible")
	cmd.Flags().BoolVar(&async, "async", false, audioAsyncFlagUsage)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

type dialogueOptions struct {
	Model      string
	Out        string
	RespFormat string
	PerTurn    bool
	Seed       *int
	Async      bool
	APIKey     string
	Format     Format
}

func runCallDialogue(ctx context.Context, f *cmdutil.Factory, scriptPath string, opts dialogueOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// --per-turn answers with JSON rather than audio, so the terminal guard
	// does not apply to it.
	if !opts.Async && !opts.PerTurn && strings.TrimSpace(opts.Out) == "" && isTerminal(os.Stdout) {
		return fmt.Errorf("audio would be written to the terminal; name a file with --out, " +
			"pipe the output, or ask for --per-turn")
	}
	req, err := buildDialogueBody(scriptPath, opts)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, opts.APIKey)
	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}
	return streamAudioAnswer(ctx, dp, audioAnswer{
		Method: "POST", Route: asyncQuery(epAudioSpeech, opts.Async),
		Body: bytes.NewReader(buf), ContentType: "application/json",
		Model: opts.Model, Out: opts.Out, Async: opts.Async, Format: opts.Format,
		// --per-turn is JSON on a route that usually answers audio, so the
		// answer is read rather than streamed to a file.
		ExpectJSON: opts.PerTurn,
	})
}

func buildDialogueBody(scriptPath string, opts dialogueOptions) (map[string]any, error) {
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}
	var script dialogueScript
	if err := json.Unmarshal(raw, &script); err != nil {
		return nil, fmt.Errorf("read %s as a dialogue script: %w\n"+
			"`olares-cli router call dialogue --help` shows the shape", scriptPath, err)
	}
	if len(script.Speakers) == 0 {
		return nil, fmt.Errorf("%s names no speakers; this model clones every voice from a "+
			"reference recording, so each speaker needs one", scriptPath)
	}
	if len(script.Turns) == 0 {
		return nil, fmt.Errorf("%s has no turns; there is nothing to say", scriptPath)
	}
	speakers := make([]map[string]any, 0, len(script.Speakers))
	for i, spk := range script.Speakers {
		ref, err := dialogueRefAudio(spk.RefAudio, i)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(spk.RefText) == "" {
			return nil, fmt.Errorf("speakers[%d].ref_text is missing; it has to be the verbatim "+
				"transcript of that clip, and a wrong one degrades the voice without saying so", i)
		}
		speakers = append(speakers, map[string]any{"ref_audio": ref, "ref_text": spk.RefText})
	}
	turns := make([]map[string]any, 0, len(script.Turns))
	for i, turn := range script.Turns {
		if strings.TrimSpace(turn.Text) == "" {
			return nil, fmt.Errorf("turns[%d].text is empty", i)
		}
		if turn.Speaker < 0 || turn.Speaker >= len(speakers) {
			return nil, fmt.Errorf("turns[%d].speaker is %d, but the script names %d speaker(s)",
				i, turn.Speaker, len(speakers))
		}
		turns = append(turns, map[string]any{"speaker": turn.Speaker, "text": turn.Text})
	}
	req := map[string]any{"speakers": speakers, "turns": turns}
	if m := strings.TrimSpace(opts.Model); m != "" {
		req["model"] = m
	}
	if v := strings.TrimSpace(opts.RespFormat); v != "" {
		req["response_format"] = v
	}
	if opts.PerTurn {
		req["per_turn"] = true
	}
	if opts.Seed != nil {
		req["seed"] = *opts.Seed
	}
	return req, nil
}

// dialogueRefAudio accepts either of the two things a script can hold. A data
// URL is passed through — somebody who already encoded a clip has done the
// work — and anything else is read off this machine, because a path is
// meaningful here and nowhere else.
func dialogueRefAudio(ref string, index int) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("speakers[%d].ref_audio is missing; name a recording of that voice", index)
	}
	if strings.HasPrefix(ref, "data:") {
		return ref, nil
	}
	encoded, err := dataURL(ref)
	if err != nil {
		return "", fmt.Errorf("read the reference recording for speakers[%d]: %w", index, err)
	}
	return encoded, nil
}

// audioAnswer is the shape every audio route that can answer with bytes shares:
// send something, and then either write the audio out, print the JSON, or —
// when the submission was async — report the task instead.
type audioAnswer struct {
	Method      string
	Route       string
	Body        io.Reader
	ContentType string
	Model       string
	Out         string
	Async       bool
	ExpectJSON  bool
	Format      Format
}

func streamAudioAnswer(ctx context.Context, dp *routerClient, a audioAnswer) error {
	resp, err := dp.do(ctx, a.Method, a.Route, a.Body, a.ContentType)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// An async submission and a rejection both answer small and both have to be
	// read whole; only a result is streamed.
	if a.Async || resp.StatusCode/100 != 2 {
		raw, rerr := io.ReadAll(resp.Body)
		if rerr != nil {
			return fmt.Errorf("read the answer: %w", rerr)
		}
		if resp.StatusCode/100 != 2 {
			return callErr(dp.formatErr(a.Method, a.Route, resp.StatusCode, raw))
		}
		if task, ok := receiptFrom(resp.StatusCode, raw); ok {
			return printReceipt(os.Stdout, task, a.Model, a.Format)
		}
		// --async was asked for and the engine answered with the result
		// anyway. Reporting that as a failure would throw away work that was
		// actually done, so it is printed as what it is.
		fmt.Fprintln(os.Stderr, "the engine answered directly rather than queueing this")
		return writeAudioResult(resp, raw, a)
	}
	if a.ExpectJSON {
		raw, rerr := io.ReadAll(resp.Body)
		if rerr != nil {
			return fmt.Errorf("read the answer: %w", rerr)
		}
		return printRawJSON(os.Stdout, raw)
	}
	dst := io.Writer(os.Stdout)
	if p := strings.TrimSpace(a.Out); p != "" {
		fh, ferr := os.Create(p)
		if ferr != nil {
			return ferr
		}
		defer fh.Close()
		dst = fh
	}
	n, err := io.Copy(dst, resp.Body)
	if err != nil {
		return fmt.Errorf("write the audio: %w", err)
	}
	if p := strings.TrimSpace(a.Out); p != "" {
		fmt.Fprintf(os.Stderr, "wrote %s (%s)\n", p, humanBytes(n))
	}
	return nil
}

// writeAudioResult puts an already-read body where the audio was meant to go.
// Only reached when --async did not take effect.
func writeAudioResult(resp *http.Response, raw []byte, a audioAnswer) error {
	if a.ExpectJSON || strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "json") {
		return printRawJSON(os.Stdout, raw)
	}
	if p := strings.TrimSpace(a.Out); p != "" {
		if err := os.WriteFile(p, raw, 0o600); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "wrote %s (%s)\n", p, humanBytes(int64(len(raw))))
		return nil
	}
	if isTerminal(os.Stdout) {
		return fmt.Errorf("the engine answered with %d bytes of audio and there is nowhere to "+
			"put it; name a file with --out", len(raw))
	}
	_, err := os.Stdout.Write(raw)
	return err
}

// Cobra reads a backquoted word in a usage string as the flag's value
// placeholder, so the command this points at is spelled plainly.
const audioAsyncFlagUsage = "hand back a task id instead of waiting; " +
	"router call task reads it"
