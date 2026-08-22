package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/audio/vad, /v1/audio/diarization, /v1/audio/enhance, /v1/audio/align
//
// Four things done to a recording that are not transcription, and the reason
// they are four separate verbs rather than flags on `transcribe` is that they
// are four separate engines. One audio mode covers all of them at the gateway,
// but an installed application serves one capability: the model that recognises
// speech cannot tell you who was speaking, and the one that can does not
// transcribe.
//
// That is also why each has its own default category, alignment included.
// Sending an audio file to the recognition default and asking for diarization
// would reach a model that answers 404 for the route, so `--model` is left alone
// and the category picks a model that declares the capability.
//
// Alignment looked like an exception for a while and is not one: the aligner is
// `audioqwenalignerv3`, an application declaring `align` and nothing else, so
// falling back to the recognition default reached an engine with no /align at
// all. There is one category per capability here because there is one
// application per capability in the Market.

// vadResponse is where the speech is. The engine also reports the parameters it
// used, which matter when a threshold is being tuned, and which -o json carries.
type vadResponse struct {
	Model         string  `json:"model"`
	SamplingRate  int     `json:"sampling_rate"`
	Duration      float64 `json:"duration"`
	SpeechSeconds float64 `json:"speech_seconds"`
	NumSegments   int     `json:"num_segments"`
	Segments      []struct {
		Start float64 `json:"start"`
		End   float64 `json:"end"`
	} `json:"segments"`
}

type diarizationResponse struct {
	Model       string   `json:"model"`
	Device      string   `json:"device"`
	NumSpeakers int      `json:"num_speakers"`
	Speakers    []string `json:"speakers"`
	NumSegments int      `json:"num_segments"`
	Segments    []struct {
		Start   float64 `json:"start"`
		End     float64 `json:"end"`
		Speaker string  `json:"speaker"`
	} `json:"segments"`
}

type alignResponse struct {
	Model    string `json:"model"`
	Device   string `json:"device"`
	Language string `json:"language"`
	Units    []struct {
		Text  string  `json:"text"`
		Start float64 `json:"start"`
		End   float64 `json:"end"`
	} `json:"units"`
}

func newCallVADCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		model     string
		threshold float64
		async     bool
		apiKey    string
	)
	cmd := &cobra.Command{
		Use:   "vad <audio-file>",
		Short: "find where the speech is in a recording",
		Long: `Detect the stretches of a recording that contain speech.

The answer is a list of intervals in seconds. This is what makes a long
recording cheap to transcribe: an hour of audio with four minutes of speech in it
is four minutes of work, and the intervals are what say which four.

--threshold trades the two mistakes against each other. Higher misses quiet
speech; lower calls a cough speech. The default is the engine's.

This needs a model that declares voice activity detection, which is a different
engine from the one that transcribes. Leaving --model off finds one.

Examples:
  olares-cli router call vad interview.wav
  olares-cli router call vad noisy.m4a --threshold 0.7 -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			fields := map[string]string{"model": callModel(model, categoryVAD)}
			if c.Flags().Changed("threshold") {
				fields["threshold"] = strconv.FormatFloat(threshold, 'f', -1, 64)
			}
			return runAudioAnalysis(c.Context(), f, audioAnalysis{
				Path: args[0], Route: epAudioVAD, Fields: fields, Async: async,
				Model: fields["model"], APIKey: apiKey, OutputIn: output, Render: renderVAD,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryVAD))
	cmd.Flags().Float64Var(&threshold, "threshold", 0, "how confident the engine must be to call something speech")
	cmd.Flags().BoolVar(&async, "async", false, audioAsyncFlagUsage)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func newCallDiarizeCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output      string
		model       string
		numSpeakers int
		minSpeakers int
		maxSpeakers int
		stream      bool
		sampleRate  int
		async       bool
		apiKey      string
	)
	cmd := &cobra.Command{
		Use:   "diarize [audio-file]",
		Short: "work out who spoke when",
		Long: `Split a recording by speaker.

The answer is intervals labelled with a speaker. The labels are arbitrary —
"SPEAKER_00" and so on — because the engine has no way to know anyone's name;
what it establishes is that two stretches are the same voice.

--num-speakers is worth setting when the count is known, since guessing it is
the hardest part of the job and getting it wrong splits one person in two.
--min-speakers and --max-speakers bound the guess instead.

This needs a model that declares diarization, which is a different engine from
the one that transcribes. Leaving --model off finds one. Combining the two is a
matter of transcribing separately and lining the intervals up.

--stream does the same job over a socket, against a model that declares
streaming diarization — a separate application again, and one that had no way to
be called from here at all before this flag. It takes raw 16-bit PCM from the
file or standard input rather than a container, and reports the segmentation as
it goes; the speaker-count flags do not apply, since the engine discovers them
as it listens.

Examples:
  olares-cli router call diarize meeting.wav
  olares-cli router call diarize call.m4a --num-speakers 2
  olares-cli router call diarize panel.wav --min-speakers 3 --max-speakers 6
  olares-cli router call diarize long-panel.wav --async
  ffmpeg -i meeting.m4a -f s16le -ar 16000 -ac 1 - | olares-cli router call diarize --stream
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if stream {
				for _, name := range []string{"num-speakers", "min-speakers", "max-speakers", "async"} {
					if c.Flags().Changed(name) {
						return fmt.Errorf("--%s does not apply to --stream: the streaming engine "+
							"discovers speakers as it listens, and a socket is already the "+
							"asynchronous shape", name)
					}
				}
				path := ""
				if len(args) > 0 {
					path = args[0]
				}
				return runDiarizeStream(c.Context(), f, path, model, sampleRate, apiKey)
			}
			if len(args) == 0 {
				return fmt.Errorf("name a recording to split, or read PCM from a pipe with --stream")
			}
			fields := map[string]string{"model": callModel(model, categoryDiarization)}
			if c.Flags().Changed("num-speakers") {
				fields["num_speakers"] = strconv.Itoa(numSpeakers)
			}
			if c.Flags().Changed("min-speakers") {
				fields["min_speakers"] = strconv.Itoa(minSpeakers)
			}
			if c.Flags().Changed("max-speakers") {
				fields["max_speakers"] = strconv.Itoa(maxSpeakers)
			}
			return runAudioAnalysis(c.Context(), f, audioAnalysis{
				Path: args[0], Route: epAudioDiarization, Fields: fields, Async: async,
				Model: fields["model"], APIKey: apiKey, OutputIn: output,
				Render: renderDiarization,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "",
		modelFlagHelp(categoryDiarization)+", or "+categoryDiarStream+" with --stream")
	cmd.Flags().IntVar(&numSpeakers, "num-speakers", 0, "how many people are speaking, when it is known")
	cmd.Flags().IntVar(&minSpeakers, "min-speakers", 0, "the fewest people the engine should consider")
	cmd.Flags().IntVar(&maxSpeakers, "max-speakers", 0, "the most people the engine should consider")
	cmd.Flags().BoolVar(&stream, "stream", false,
		"send PCM over a socket and report the segmentation as it arrives; "+
			"resolves "+categoryDiarStream+" instead of "+categoryDiarization+" when --model is omitted")
	cmd.Flags().IntVar(&sampleRate, "sample-rate", 16000, "with --stream, the sample rate of the PCM being sent")
	cmd.Flags().BoolVar(&async, "async", false, audioAsyncFlagUsage)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func newCallAlignCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		model    string
		text     string
		language string
		async    bool
		apiKey   string
	)
	cmd := &cobra.Command{
		Use:   "align <audio-file>",
		Short: "line a transcript up with the audio",
		Long: `Find where each part of a known transcript occurs in the audio.

Unlike transcription this is given the words: --text is what was said, and the
answer is when each piece of it was said. That is what subtitles are made of, and
what lets a recording be searched by text.

The transcript comes from --text or from standard input, so the output of
"router call transcribe" can be piped straight in.

This needs a model that declares alignment, which is a different application
from the one that transcribes even though the two jobs look related. Leaving
--model off finds one.

Examples:
  olares-cli router call align talk.wav --text "the words that were spoken"
  olares-cli router call transcribe talk.wav | olares-cli router call align talk.wav
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			transcript := strings.TrimSpace(text)
			if transcript == "" {
				read, err := readPromptArgs(nil, "transcript")
				if err != nil {
					return fmt.Errorf("%w; pass it with --text", err)
				}
				transcript = read
			}
			return runAudioAnalysis(c.Context(), f, audioAnalysis{
				Path:  args[0],
				Route: epAudioAlign,
				Fields: map[string]string{
					"model":    callModel(model, categoryAlign),
					"text":     transcript,
					"language": strings.TrimSpace(language),
				},
				Async: async, Model: callModel(model, categoryAlign),
				APIKey: apiKey, OutputIn: output, Render: renderAlign,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryAlign))
	cmd.Flags().StringVar(&text, "text", "", "the transcript to line up; read from standard input when omitted")
	cmd.Flags().StringVar(&language, "language", "", "language of the audio; detected when omitted")
	cmd.Flags().BoolVar(&async, "async", false, audioAsyncFlagUsage)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

// enhance is the odd one out: it answers with audio rather than with a
// description of the audio, so it shares nothing with the three above but the
// upload.
func newCallEnhanceCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		model   string
		outPath string
		respFmt string
		async   bool
		apiKey  string
		output  string
	)
	cmd := &cobra.Command{
		Use:   "enhance <audio-file>",
		Short: "clean up a noisy recording",
		Long: `Remove noise from a recording.

The answer is audio, not a description of it, so it goes to --out or to standard
output when that is a pipe. Writing audio into a terminal is refused rather than
done.

Worth doing before transcribing a difficult recording rather than instead of it:
a cleaned-up file transcribes better, and this returns no text of its own.

This needs a model that declares enhancement, which is a different engine from
the one that transcribes. Leaving --model off finds one.

Examples:
  olares-cli router call enhance noisy.wav --out clean.wav
  olares-cli router call enhance street.m4a --format flac --out clean.flac
  olares-cli router call enhance noisy.wav | olares-cli router call transcribe /dev/stdin
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCallEnhance(c.Context(), f, args[0], enhanceOptions{
				Model:    callModel(model, categoryEnhance),
				Out:      outPath,
				Format:   respFmt,
				Async:    async,
				APIKey:   apiKey,
				OutputIn: output,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryEnhance))
	cmd.Flags().StringVar(&outPath, "out", "", "write the audio here instead of standard output")
	cmd.Flags().StringVar(&respFmt, "format", "", "container format, e.g. wav, flac or ogg")
	cmd.Flags().BoolVar(&async, "async", false, audioAsyncFlagUsage)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

// audioAnalysis is the shape the three describing verbs share: upload a file
// with some fields, decode JSON, render it.
type audioAnalysis struct {
	Path     string
	Route    string
	Fields   map[string]string
	Async    bool
	Model    string
	APIKey   string
	OutputIn string
	Render   func(io.Writer, []byte) error
}

func runAudioAnalysis(ctx context.Context, f *cmdutil.Factory, a audioAnalysis) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(a.OutputIn)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, a.APIKey)
	fields := a.Fields
	if a.Async {
		// These are multipart routes, so the flag is a field. As a query
		// parameter it would be read by nothing and the call would simply
		// wait — see the note at the top of call_audio_task.go.
		fields = make(map[string]string, len(a.Fields)+1)
		for k, v := range a.Fields {
			fields[k] = v
		}
		fields[audioAsyncFormField] = "1"
	}
	body, contentType, err := multipartFile(a.Path, "file", fields)
	if err != nil {
		return err
	}
	resp, err := dp.do(ctx, "POST", a.Route, body, contentType)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read the answer: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return callErr(dp.formatErr("POST", a.Route, resp.StatusCode, raw))
	}
	if task, ok := receiptFrom(resp.StatusCode, raw); ok {
		return printReceipt(os.Stdout, task, a.Model, format)
	}
	if format == FormatJSON {
		return printRawJSON(os.Stdout, raw)
	}
	return a.Render(os.Stdout, raw)
}

func renderVAD(w io.Writer, raw []byte) error {
	var resp vadResponse
	if err := decodeAudioAnalysis(raw, &resp); err != nil {
		return err
	}
	if len(resp.Segments) == 0 {
		_, err := fmt.Fprintf(w, "no speech found in %s of audio.\n", seconds(resp.Duration))
		return err
	}
	t := newTable(w, "#", "START", "END", "LENGTH")
	for i := range resp.Segments {
		s := &resp.Segments[i]
		t.row(strconv.Itoa(i+1), seconds(s.Start), seconds(s.End), seconds(s.End-s.Start))
	}
	if err := t.flush(); err != nil {
		return err
	}
	share := ""
	if resp.Duration > 0 {
		share = fmt.Sprintf(" of %s (%.0f%%)", seconds(resp.Duration), resp.SpeechSeconds/resp.Duration*100)
	}
	_, err := fmt.Fprintf(os.Stderr, "\n%s speech%s across %d segments  %s\n",
		seconds(resp.SpeechSeconds), share, resp.NumSegments, nonEmpty(resp.Model))
	return err
}

func renderDiarization(w io.Writer, raw []byte) error {
	var resp diarizationResponse
	if err := decodeAudioAnalysis(raw, &resp); err != nil {
		return err
	}
	if len(resp.Segments) == 0 {
		_, err := fmt.Fprintln(w, "the engine found nobody speaking.")
		return err
	}
	t := newTable(w, "SPEAKER", "START", "END", "LENGTH")
	for i := range resp.Segments {
		s := &resp.Segments[i]
		t.row(nonEmpty(s.Speaker), seconds(s.Start), seconds(s.End), seconds(s.End-s.Start))
	}
	if err := t.flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(os.Stderr, "\n%d speakers across %d segments  %s\n",
		resp.NumSpeakers, resp.NumSegments, nonEmpty(resp.Model))
	return err
}

func renderAlign(w io.Writer, raw []byte) error {
	var resp alignResponse
	if err := decodeAudioAnalysis(raw, &resp); err != nil {
		return err
	}
	if len(resp.Units) == 0 {
		_, err := fmt.Fprintln(w, "the engine lined nothing up; the transcript may not match the audio.")
		return err
	}
	t := newTable(w, "START", "END", "TEXT")
	for i := range resp.Units {
		u := &resp.Units[i]
		t.row(seconds(u.Start), seconds(u.End), clip(u.Text, 60))
	}
	if err := t.flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(os.Stderr, "\n%d units  %s  %s\n", len(resp.Units),
		nonEmpty(resp.Language), nonEmpty(resp.Model))
	return err
}

// decodeAudioAnalysis hands the engine's own answer back when it does not fit.
// Router forwards these routes untouched, so the shape belongs to whichever
// engine image answered, and quoting what it said is more use than naming the
// field that did not parse.
func decodeAudioAnalysis(raw []byte, out any) error {
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("the engine answered in a shape this verb does not know: %s\n"+
			"-o json prints it as it came", truncate(strings.TrimSpace(string(raw)), 300))
	}
	return nil
}

// seconds is the unit every one of these routes reports in. Rendered to
// milliseconds because a segment boundary is what a subtitle is cut on.
func seconds(v float64) string {
	return strconv.FormatFloat(v, 'f', 3, 64) + "s"
}

type enhanceOptions struct {
	Model    string
	Out      string
	Format   string
	Async    bool
	APIKey   string
	OutputIn string
}

func runCallEnhance(ctx context.Context, f *cmdutil.Factory, path string, opts enhanceOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(opts.OutputIn)
	if err != nil {
		return err
	}
	if strings.TrimSpace(opts.Out) == "" && !opts.Async && isTerminal(os.Stdout) {
		return fmt.Errorf("audio would be written to the terminal; name a file with --out, or pipe the output")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, opts.APIKey)
	fields := map[string]string{
		"model":  strings.TrimSpace(opts.Model),
		"format": strings.TrimSpace(opts.Format),
	}
	if opts.Async {
		fields[audioAsyncFormField] = "1"
	}
	body, contentType, err := multipartFile(path, "file", fields)
	if err != nil {
		return err
	}
	return streamAudioAnswer(ctx, dp, audioAnswer{
		Method: "POST", Route: epAudioEnhance,
		Body: body, ContentType: contentType,
		Model: opts.Model, Out: opts.Out, Async: opts.Async, Format: format,
	})
}
