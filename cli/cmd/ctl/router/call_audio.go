package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
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
// The live audio routes (/v1/audio/stream and /v1/audio/diarize/stream) are
// WebSocket and have no verb here. A command that opened a socket to relay
// microphone frames would be a different program.

func newCallTranscribeCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		model     string
		language  string
		prompt    string
		respFmt   string
		translate bool
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
English regardless of what was spoken.

Plain text goes to standard output, so this pipes. --response-format asks the
engine for something structured — "verbose_json" carries timings, "srt" and
"vtt" are subtitle files — and whatever comes back is printed as it arrived.

Examples:
  olares-cli router call transcribe meeting.m4a
  olares-cli router call transcribe clip.wav --language zh
  olares-cli router call transcribe talk.mp3 --response-format srt > talk.srt
  olares-cli router call transcribe interview.mp3 --translate
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCallTranscribe(c.Context(), f, args[0], transcribeOptions{
				Model:      model,
				Language:   language,
				Prompt:     prompt,
				RespFormat: respFmt,
				Translate:  translate,
				APIKey:     apiKey,
				OutputIn:   output,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", "model to use; the workspace audio default when omitted")
	cmd.Flags().StringVar(&language, "language", "", "language of the audio, as an ISO-639-1 code")
	cmd.Flags().StringVar(&prompt, "prompt", "", "text that biases spelling and vocabulary")
	cmd.Flags().StringVar(&respFmt, "response-format", "", "json, text, verbose_json, srt or vtt, as the engine supports")
	cmd.Flags().BoolVar(&translate, "translate", false, "translate to English instead of transcribing verbatim")
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
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp, _, err := dataPlane(ctx, pc, opts.APIKey)
	if err != nil {
		return err
	}

	fields := map[string]string{
		"model":           strings.TrimSpace(opts.Model),
		"language":        strings.TrimSpace(opts.Language),
		"prompt":          strings.TrimSpace(opts.Prompt),
		"response_format": strings.TrimSpace(opts.RespFormat),
	}
	body, contentType, err := multipartFile(path, "file", fields)
	if err != nil {
		return err
	}

	route := dataPlaneAPI + "/audio/transcriptions"
	if opts.Translate {
		route = dataPlaneAPI + "/audio/translations"
	}
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
	if format == FormatJSON {
		return printRawJSON(os.Stdout, raw)
	}
	return printTranscript(os.Stdout, raw)
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

// multipartFile builds an upload with the file plus whatever text fields have a
// value. The file part comes last on purpose: Router reads the model field by
// scanning the upload, and a model named after the audio makes it hold the whole
// file in memory to find it.
func multipartFile(path, fieldName string, fields map[string]string) (io.Reader, string, error) {
	fh, err := os.Open(path)
	if err != nil {
		// os.Open's own error already names the operation and the path.
		return nil, "", err
	}
	defer fh.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		if strings.TrimSpace(v) == "" {
			continue
		}
		if err := mw.WriteField(k, v); err != nil {
			return nil, "", fmt.Errorf("write field %s: %w", k, err)
		}
	}
	part, err := mw.CreateFormFile(fieldName, filepath.Base(path))
	if err != nil {
		return nil, "", fmt.Errorf("create the file part: %w", err)
	}
	if _, err := io.Copy(part, fh); err != nil {
		return nil, "", fmt.Errorf("read %s: %w", path, err)
	}
	if err := mw.Close(); err != nil {
		return nil, "", fmt.Errorf("finish the upload: %w", err)
	}
	return &buf, mw.FormDataContentType(), nil
}

func newCallSpeakCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		model   string
		voice   string
		outPath string
		respFmt string
		speed   float64
		apiKey  string
	)
	cmd := &cobra.Command{
		Use:   "speak [text]",
		Short: "text to speech",
		Long: `Synthesise speech from text.

The text is the argument or standard input. The answer is audio, so it goes to
the file named by --out, or to standard output when that is a pipe. Writing
audio bytes into a terminal is refused rather than done.

--voice and --response-format are passed through untouched: which voices exist
and which container formats they come in is the engine's business, and Router
does not translate either.

Examples:
  olares-cli router call speak "your build finished" --out done.mp3
  olares-cli router call speak "hello" --voice alloy --out hello.wav --response-format wav
  echo "read this aloud" | olares-cli router call speak --out out.mp3
  olares-cli router call speak "piped" | ffplay -
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			text, err := readPromptArgs(args, "text")
			if err != nil {
				return err
			}
			var speedPtr *float64
			if c.Flags().Changed("speed") {
				speedPtr = &speed
			}
			return runCallSpeak(c.Context(), f, text, speakOptions{
				Model:      model,
				Voice:      voice,
				OutPath:    outPath,
				RespFormat: respFmt,
				Speed:      speedPtr,
				APIKey:     apiKey,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", "model to use; the workspace audio default when omitted")
	cmd.Flags().StringVar(&voice, "voice", "", "voice name, as the engine names it")
	cmd.Flags().StringVar(&outPath, "out", "", "write the audio here instead of standard output")
	cmd.Flags().StringVar(&respFmt, "response-format", "", "container format, e.g. mp3 or wav")
	cmd.Flags().Float64Var(&speed, "speed", 1, "playback rate, if the engine supports it")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	return cmd
}

type speakOptions struct {
	Model      string
	Voice      string
	OutPath    string
	RespFormat string
	Speed      *float64
	APIKey     string
}

func runCallSpeak(ctx context.Context, f *cmdutil.Factory, text string, opts speakOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(opts.OutPath) == "" && isTerminal(os.Stdout) {
		return fmt.Errorf("audio would be written to the terminal; name a file with --out, or pipe the output")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp, _, err := dataPlane(ctx, pc, opts.APIKey)
	if err != nil {
		return err
	}

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
	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}

	route := dataPlaneAPI + "/audio/speech"
	resp, err := dp.do(ctx, "POST", route, bytes.NewReader(buf), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)
		return callErr(dp.formatErr("POST", route, resp.StatusCode, raw))
	}

	dst := io.Writer(os.Stdout)
	if p := strings.TrimSpace(opts.OutPath); p != "" {
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
	if p := strings.TrimSpace(opts.OutPath); p != "" {
		fmt.Fprintf(os.Stderr, "wrote %s (%s bytes)\n", p, strconv.FormatInt(n, 10))
	}
	return nil
}
