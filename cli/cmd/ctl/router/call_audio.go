package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/audio/transcriptions
// POST /v1/audio/speech
//
// Router forwards these two verbatim: it reads the model name to pick a
// provider and then passes the method, the path, the query, the content type
// and the bytes straight through, streaming the answer back as it arrives. So a
// field this CLI does not know about still reaches the engine, and the engine's
// own answer — including its errors — is what comes back.
//
// The live audio routes are WebSocket rather than HTTP and live in
// call_audio_stream.go; the voice-cloning and dialogue shapes of /v1/audio/speech
// live in call_audio_voice.go.

func newCallTranscribeCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		model     string
		language  string
		prompt    string
		respFmt   string
		translate bool
		async     bool
		apiKey    string
	)
	cmd := &cobra.Command{
		Use:   "transcribe <audio-file>",
		Short: "speech to text",
		Long: `Transcribe an audio file.

The file is uploaded as it is; the engine decides which formats it accepts, and
rejects the rest in its own words. --language is a hint that usually improves
both accuracy and speed when you know the answer, and --prompt biases spelling,
which is how proper nouns and jargon are kept intact.

--translate sends the file to the translation route instead, which returns
English regardless of what was spoken. Only the faster-whisper engines serve
that route, so it is one of the few audio calls worth naming a --model for.

--async hands back a task id rather than waiting, and for anything longer than a
few minutes it is the only thing that works: a synchronous request for an hour
of audio is held open for as long as the engine takes, and something on the way
will cut it first. "router call task" reads the result.

Plain text goes to standard output, so this pipes. --response-format asks the
engine for something structured — "verbose_json" carries timings, "srt" and
"vtt" are subtitle files — and whatever comes back is printed as it arrived.

Examples:
  olares-cli router call transcribe meeting.m4a
  olares-cli router call transcribe clip.wav --language zh
  olares-cli router call transcribe talk.mp3 --response-format srt > talk.srt
  olares-cli router call transcribe interview.mp3 --translate
  olares-cli router call transcribe all-hands.m4a --async
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCallTranscribe(c.Context(), f, args[0], transcribeOptions{
				Model:      callModel(model, categorySTT),
				Language:   language,
				Prompt:     prompt,
				RespFormat: respFmt,
				Translate:  translate,
				Async:      async,
				APIKey:     apiKey,
				OutputIn:   output,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categorySTT))
	cmd.Flags().StringVar(&language, "language", "", "language of the audio, as an ISO-639-1 code")
	cmd.Flags().StringVar(&prompt, "prompt", "", "text that biases spelling and vocabulary")
	cmd.Flags().StringVar(&respFmt, "response-format", "", "json, text, verbose_json, srt or vtt, as the engine supports")
	cmd.Flags().BoolVar(&translate, "translate", false, "translate to English instead of transcribing verbatim")
	cmd.Flags().BoolVar(&async, "async", false, audioAsyncFlagUsage)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

type transcribeOptions struct {
	Model      string
	Language   string
	Prompt     string
	RespFormat string
	Translate  bool
	Async      bool
	APIKey     string
	OutputIn   string
}

func runCallTranscribe(ctx context.Context, f *cmdutil.Factory, path string, opts transcribeOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(opts.OutputIn)
	if err != nil {
		return err
	}
	if err := checkAudioUploadSize(path, os.Stderr); err != nil {
		return err
	}
	pc, err := prepareLongRequest(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, opts.APIKey)

	fields := audioMultipartFields(opts.Model, opts.Async, map[string]string{
		"language":        strings.TrimSpace(opts.Language),
		"prompt":          strings.TrimSpace(opts.Prompt),
		"response_format": strings.TrimSpace(opts.RespFormat),
	})
	body, contentType, err := multipartFile(path, "file", fields)
	if err != nil {
		return err
	}

	route := epAudioTranscriptions
	if opts.Translate {
		route = epAudioTranslations
	}
	route = audioRequestPath(route, opts.Model, opts.Async)
	resp, err := dp.do(ctx, "POST", route, body, contentType)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read the transcription: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return callErr(dp.formatErr("POST", route, resp.StatusCode, raw))
	}
	if task, ok := receiptFrom(resp.StatusCode, raw); ok {
		return printReceipt(os.Stdout, task, opts.Model, format)
	}
	if format == FormatJSON {
		return printRawJSON(os.Stdout, raw)
	}
	return printTranscript(os.Stdout, raw)
}

const (
	audioUploadWarnBytes    = 90 * 1024 * 1024
	audioRequestBudgetBytes = 96 * 1024 * 1024
	audioMultipartAllowance = 64 * 1024
	audioFileMaxBytes       = audioRequestBudgetBytes - audioMultipartAllowance
)

func checkAudioUploadSize(path string, stderr io.Writer) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() > audioFileMaxBytes {
		return fmt.Errorf("%s is %s (%d bytes); audio files cannot exceed %s (%d bytes) "+
			"(Router limits the total request body to %s (%d bytes), with %s (%d bytes) reserved "+
			"for multipart metadata)",
			path, humanBytes(info.Size()), info.Size(),
			humanBytes(audioFileMaxBytes), audioFileMaxBytes,
			humanBytes(audioRequestBudgetBytes), audioRequestBudgetBytes,
			humanBytes(audioMultipartAllowance), audioMultipartAllowance)
	}
	if info.Size() > audioUploadWarnBytes {
		_, err = fmt.Fprintf(stderr, "warning: audio input is %s (%d bytes); convert it to "+
			"16 kHz mono FLAC, or split long STT input into smaller files before uploading\n",
			humanBytes(info.Size()), info.Size())
	}
	return err
}

func checkDialogueRequestSize(size int64, stderr io.Writer) error {
	if size > audioRequestBudgetBytes {
		return fmt.Errorf("dialogue JSON is %s (%d bytes); Router limits the total request body "+
			"to %s (%d bytes)", humanBytes(size), size,
			humanBytes(audioRequestBudgetBytes), audioRequestBudgetBytes)
	}
	if size > audioUploadWarnBytes {
		_, err := fmt.Fprintf(stderr, "warning: dialogue JSON is %s (%d bytes); shorten or "+
			"compress the reference clips, or submit with --async\n", humanBytes(size), size)
		return err
	}
	return nil
}

func audioRequestPath(route, model string, async bool) string {
	q := url.Values{}
	if m := strings.TrimSpace(model); m != "" {
		q.Set("model", m)
	}
	if async {
		q.Set(audioAsyncQueryKey, "1")
	}
	return withQuery(route, q)
}

// printTranscript prefers the text out of a JSON answer and falls back to the
// bytes as they came. The engine chooses the shape — an srt file is not JSON,
// and a verbose_json answer has more in it than a caller reading the transcript
// wants — so this unwraps the common case and gets out of the way otherwise.
func printTranscript(w io.Writer, raw []byte) error {
	var env struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &env) == nil && strings.TrimSpace(env.Text) != "" {
		_, err := fmt.Fprintln(w, strings.TrimSpace(env.Text))
		return err
	}
	_, err := w.Write(raw)
	if err == nil && len(raw) > 0 && raw[len(raw)-1] != '\n' {
		_, err = fmt.Fprintln(w)
	}
	return err
}

func printRawJSON(w io.Writer, raw []byte) error {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		// Not JSON at all — an srt or plain-text answer under -o json. Printing
		// it is more useful than refusing, and quoting it makes it valid JSON.
		return printJSON(w, string(raw))
	}
	return printJSON(w, v)
}

// multipartFile streams a file plus non-empty text fields without buffering the
// complete request. The file part comes last so metadata reaches the receiver
// before the file bytes.
//
// The body is encoded as it is sent rather than into a buffer first. A
// recording long enough to be worth --async is one nobody should have to hold
// in memory, and all a buffer bought was a Content-Length neither Router nor
// the engines need. Only opening the file fails here; anything that goes wrong
// while reading it travels down the pipe, so a file truncated or removed
// mid-upload fails the request instead of quietly sending less than was on
// disk.
//
// A streamed body cannot be replayed, which puts this upload on the same
// footing as `files upload`: an access token close to expiry is rotated before
// the body is handed over rather than after a 401 (see cmdutil's
// preflightSkew).
func multipartFile(path, fieldName string, fields map[string]string) (io.Reader, string, error) {
	fh, err := os.Open(path)
	if err != nil {
		// os.Open's own error already names the operation and the path.
		return nil, "", err
	}
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		defer fh.Close()
		// A request the transport abandons closes pr, which turns the next
		// write into an error and ends this goroutine.
		_ = pw.CloseWithError(writeMultipartFile(mw, fh, path, fieldName, fields))
	}()
	return pr, mw.FormDataContentType(), nil
}

func writeMultipartFile(
	mw *multipart.Writer,
	file io.Reader,
	path, fieldName string,
	fields map[string]string,
) error {
	for k, v := range fields {
		if strings.TrimSpace(v) == "" {
			continue
		}
		if err := mw.WriteField(k, v); err != nil {
			return fmt.Errorf("write field %s: %w", k, err)
		}
	}
	part, err := mw.CreateFormFile(fieldName, filepath.Base(path))
	if err != nil {
		return fmt.Errorf("create the file part: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("finish the upload: %w", err)
	}
	return nil
}

func newCallSpeakCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		model   string
		voice   string
		outPath string
		respFmt string
		speed   float64
		voices  bool
		apiKey  string
		output  string
		soundFX bool
		async   bool
	)
	// The category is the only thing --sound-fx changes: same path, same body,
	// a different model when none was named.
	fallback := func() string {
		if soundFX {
			return categorySoundFX
		}
		return categoryTTS
	}
	cmd := &cobra.Command{
		Use:   "speak [text]",
		Short: "text to speech",
		Long: `Synthesise speech from text.

The text is the argument or standard input. The answer is audio, so it goes to
the file named by --out, or to standard output when that is a pipe. Writing
audio bytes into a terminal is refused rather than done.

--voice and --response-format are passed through untouched: which voices exist
and which container formats they come in is the engine's business, and Router
does not translate either. --voices lists what this model offers and synthesises
nothing; a model built only for voice cloning has no list and answers 404.

--out is where the audio goes. -o names the format of the voice listing, which
is the only thing this verb prints rather than plays, and has no effect on a
synthesis.

--sound-fx generates a sound from a description of it instead of speech from
words. It is the same request to the same endpoint: engines that make sound
effects mount /v1/audio/speech like every other audio model, so the model is
the only thing that decides which you get, and the flag only changes which
default is resolved. Naming a sound-effect model with --model does the same.

A voice cloned from a recording, and a conversation between several of them, are
"router call clone" and "router call dialogue". They are the same endpoint again,
and again separate applications, which is why they are separate verbs.

--async hands back a task id instead of the audio; "router call task" collects
it.

Examples:
  olares-cli router call speak "your build finished" --out done.mp3
  olares-cli router call speak "hello" --voice alloy --out hello.wav --response-format wav
  echo "read this aloud" | olares-cli router call speak --out out.mp3
  olares-cli router call speak "piped" | ffplay -
  olares-cli router call speak --voices
  olares-cli router call speak --sound-fx "rain on a tin roof" --out rain.mp3
  olares-cli router call speak "$(cat chapter.txt)" --async
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if voices {
				if len(args) > 0 {
					return fmt.Errorf("--voices lists what the model offers; it takes no text")
				}
				format, ferr := parseFormat(output)
				if ferr != nil {
					return ferr
				}
				return runListVoices(c.Context(), f, callModel(model, fallback()), apiKey, format)
			}
			text, err := readPromptArgs(args, "text")
			if err != nil {
				return err
			}
			var speedPtr *float64
			if c.Flags().Changed("speed") {
				speedPtr = &speed
			}
			format, ferr := parseFormat(output)
			if ferr != nil {
				return ferr
			}
			return runCallSpeak(c.Context(), f, text, speakOptions{
				Model:      callModel(model, fallback()),
				Voice:      voice,
				OutPath:    outPath,
				RespFormat: respFmt,
				Speed:      speedPtr,
				Async:      async,
				APIKey:     apiKey,
				Format:     format,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "",
		modelFlagHelp(categoryTTS)+", or "+categorySoundFX+" with --sound-fx")
	cmd.Flags().StringVar(&voice, "voice", "", "voice name, as the engine names it")
	cmd.Flags().StringVar(&outPath, "out", "", "write the audio here instead of standard output")
	cmd.Flags().StringVar(&respFmt, "response-format", "", "container format, e.g. mp3 or wav")
	cmd.Flags().Float64Var(&speed, "speed", 1, "playback rate, if the engine supports it")
	cmd.Flags().BoolVar(&voices, "voices", false, "list the voices this model offers and synthesise nothing")
	cmd.Flags().BoolVar(&soundFX, "sound-fx", false,
		"produce a sound effect from the description rather than speech; "+
			"resolves "+categorySoundFX+" instead of "+categoryTTS+" when --model is omitted")
	cmd.Flags().BoolVar(&async, "async", false, audioAsyncFlagUsage)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

// GET /v1/audio/voices
//
// Named voices are one of two ways a TTS engine picks a voice; the other is a
// reference recording, and an engine built for that has nothing to list. So an
// empty list and a 404 both mean "this model is not chosen from a menu" rather
// than a misconfiguration.
func runListVoices(ctx context.Context, f *cmdutil.Factory, model, apiKey string, format Format) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, apiKey)
	path := epAudioVoices
	if m := strings.TrimSpace(model); m != "" {
		q := url.Values{}
		q.Set("model", m)
		path = withQuery(path, q)
	}
	var resp struct {
		Voices []struct {
			ID          string   `json:"id"`
			Name        string   `json:"name"`
			Language    string   `json:"language"`
			Languages   []string `json:"languages"`
			Gender      string   `json:"gender"`
			Description string   `json:"description"`
		} `json:"voices"`
	}
	if err := dp.doJSON(ctx, "GET", path, nil, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	if len(resp.Voices) == 0 {
		_, err := fmt.Println("this model offers no named voices. It is either cloned from a " +
			"reference recording or has a single built-in voice; --voice has nothing to name.")
		return err
	}
	t := newTable(os.Stdout, "VOICE", "NAME", "LANGUAGE", "DESCRIPTION")
	for i := range resp.Voices {
		v := &resp.Voices[i]
		lang := strings.TrimSpace(v.Language)
		if lang == "" {
			lang = strings.Join(v.Languages, " ")
		}
		t.row(v.ID, nonEmpty(v.Name), nonEmpty(lang), clip(v.Description, 48))
	}
	return t.flush()
}

type speakOptions struct {
	Model      string
	Voice      string
	OutPath    string
	RespFormat string
	Speed      *float64
	Async      bool
	APIKey     string
	Format     Format
}

func runCallSpeak(ctx context.Context, f *cmdutil.Factory, text string, opts speakOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if !opts.Async && strings.TrimSpace(opts.OutPath) == "" && isTerminal(os.Stdout) {
		return fmt.Errorf("audio would be written to the terminal; name a file with --out, " +
			"pipe the output, or submit it with --async")
	}
	pc, err := prepareLongRequest(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, opts.APIKey)

	req := buildSpeakRequest(text, opts)
	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}
	return streamAudioAnswer(ctx, dp, audioAnswer{
		Method: "POST", Route: audioRequestPath(epAudioSpeech, opts.Model, opts.Async),
		Body: bytes.NewReader(buf), ContentType: "application/json",
		Model: opts.Model, Out: opts.OutPath, Async: opts.Async, Format: opts.Format,
	})
}

func buildSpeakRequest(text string, opts speakOptions) map[string]any {
	req := map[string]any{"input": text}
	if v := strings.TrimSpace(opts.Model); v != "" {
		req["model"] = v
	}
	if v := strings.TrimSpace(opts.Voice); v != "" {
		req["voice"] = v
	}
	if v := strings.TrimSpace(opts.RespFormat); v != "" {
		req["response_format"] = v
	}
	if opts.Speed != nil {
		req["speed"] = *opts.Speed
	}
	return req
}
