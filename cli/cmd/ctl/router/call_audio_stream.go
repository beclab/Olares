package router

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// WS /v1/audio/stream, WS /v1/audio/diarize/stream
//
// Two of the audio routes are sockets rather than requests, and they exist for
// the case the HTTP ones cannot serve: audio that is still being produced.
// Recognition and diarization both have applications built for it —
// `audioqwen3asrv3` declares stt_stream, `audiosortformerdiarsv3` declares
// diar_stream — and neither could be reached from here at all before these two
// verbs, so an installed streaming engine had no way to be tried.
//
// The protocol is the same on both, because both engines are the same wrapper:
//
//	server -> {"type":"ready"}
//	client -> {"type":"start","sample_rate":16000}   (optional)
//	client -> binary frames of signed 16-bit little-endian PCM
//	server -> {"type":"partial",...} as it decodes
//	client -> {"type":"stop"}
//	server -> {"type":"final",...} then {"type":"closed","audio_seconds":n}
//
// What a frame carries is the only difference: recognition sends `text`,
// diarization sends `segments` and `speakers`. Everything else here is shared.
//
// The audio has to be PCM already. These sockets take samples, not a container:
// there is no decoder on the other end, so a .mp3 fed in produces a transcript
// of its own header bytes rather than an error. Hence the note in both verbs and
// the ffmpeg line in their examples.

// pcmFrameBytes is how much audio goes in one frame: 32 KiB is a second of
// 16 kHz mono PCM. The engine buffers to its own step size regardless, so this
// only decides how often a live capture is handed over — and a frame per second
// is what makes a `partial` arrive about that often.
const pcmFrameBytes = 32 << 10

// audioStreamFrame is every frame either engine sends. The fields are a union
// of the two shapes; a recognition frame leaves the diarization ones empty and
// the other way round.
type audioStreamFrame struct {
	Type         string   `json:"type"`
	Text         string   `json:"text"`
	Language     string   `json:"language"`
	Detail       string   `json:"detail"`
	AudioSeconds *float64 `json:"audio_seconds"`
	Speakers     []string `json:"speakers"`
	Segments     []struct {
		Start   float64 `json:"start"`
		End     float64 `json:"end"`
		Speaker string  `json:"speaker"`
	} `json:"segments"`
}

func newCallListenCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		model      string
		language   string
		sampleRate int
		stepMillis int
		apiKey     string
		partials   bool
	)
	cmd := &cobra.Command{
		Use:   "listen [pcm-file]",
		Short: "speech to text as the audio arrives",
		Long: `Transcribe audio over a socket, reporting the text as it is decoded.

The difference from "transcribe" is when the answer arrives, not how good it is:
this reports a running transcript while the audio is still being sent, which is
what a live caption needs and what a file upload cannot do. For a recording that
already exists, "transcribe" is the better call — it sees the whole thing at once
and is not asked to commit to words early.

The input is raw signed 16-bit little-endian PCM, mono, from the file named or
from standard input. It is samples rather than a file format: there is no decoder
on the other end, so a compressed file sent here is transcribed as noise instead
of being refused. ffmpeg is what converts one.

The running transcript goes to stderr and the final one to stdout, so this pipes
while still being watchable. --no-partials keeps stderr quiet.

This needs a model that declares streaming recognition, which is not the same as
one that declares recognition: an application does one or the other. Leaving
--model off finds one.

Examples:
  olares-cli router call listen speech.pcm
  ffmpeg -i talk.m4a -f s16le -ar 16000 -ac 1 - | olares-cli router call listen
  ffmpeg -f avfoundation -i ":0" -f s16le -ar 16000 -ac 1 - | olares-cli router call listen
  olares-cli router call listen speech.pcm --language zh > talk.txt
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts := audioStreamOptions{
				Route:      epAudioStreamWS,
				Model:      callModel(model, categorySTTStream),
				Language:   language,
				SampleRate: sampleRate,
				StepMillis: stepMillis,
				APIKey:     apiKey,
				Partials:   partials,
				Render:     renderTranscriptStream,
			}
			if len(args) > 0 {
				opts.Path = args[0]
			}
			return runAudioStream(c.Context(), f, opts)
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categorySTTStream))
	cmd.Flags().StringVar(&language, "language", "", "language of the audio, as an ISO-639-1 code")
	cmd.Flags().IntVar(&sampleRate, "sample-rate", 16000, "sample rate of the PCM being sent")
	cmd.Flags().IntVar(&stepMillis, "step-ms", 0, "how much audio the engine decodes at a time; its own default when 0")
	cmd.Flags().BoolVar(&partials, "partials", true, "report the running transcript on stderr while it decodes")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	return cmd
}

type audioStreamOptions struct {
	Route      string
	Path       string
	Model      string
	Language   string
	SampleRate int
	StepMillis int
	APIKey     string
	Partials   bool
	OutputIn   string
	// Render prints one frame. It is given the frame and whether it is the
	// final one, and writes progress to stderr and the result to stdout.
	Render func(frame *audioStreamFrame, final bool) error
}

func runAudioStream(ctx context.Context, f *cmdutil.Factory, opts audioStreamOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	src, closeSrc, err := openPCMSource(opts.Path)
	if err != nil {
		return err
	}
	defer closeSrc()

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	token, err := f.ValidAccessToken(ctx)
	if err != nil {
		return err
	}
	conn, err := dialAudioStream(ctx, pc, token, opts)
	if err != nil {
		return err
	}
	defer conn.Close()

	// The engine says `ready` before it will decode anything, and a `start`
	// sent ahead of that is read by nothing.
	if err := awaitStreamReady(conn); err != nil {
		return err
	}
	if err := sendStreamStart(conn, opts); err != nil {
		return err
	}

	// Reading runs concurrently with sending, which is the whole point: a
	// partial for the first second arrives while the tenth is being uploaded.
	frames := make(chan error, 1)
	go func() { frames <- readAudioStream(conn, opts) }()

	sendErr := sendPCM(conn, src)
	// `stop` is what makes the engine finalise, so it is sent even after a
	// read failure upstream: the frames goroutine reports what actually broke.
	_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"stop"}`))
	readErr := <-frames
	if readErr != nil {
		return readErr
	}
	return sendErr
}

// openPCMSource is the file or standard input. A terminal is refused: waiting
// for someone to type PCM is a hang rather than an invitation.
func openPCMSource(path string) (io.Reader, func(), error) {
	if p := strings.TrimSpace(path); p != "" {
		fh, err := os.Open(p)
		if err != nil {
			return nil, func() {}, err
		}
		return fh, func() { _ = fh.Close() }, nil
	}
	if isTerminal(os.Stdin) {
		return nil, func() {}, fmt.Errorf("no audio given; name a PCM file or pipe one in, " +
			"e.g. `ffmpeg -i talk.m4a -f s16le -ar 16000 -ac 1 - | olares-cli router call listen`")
	}
	return os.Stdin, func() {}, nil
}

// dialAudioStream opens the socket to Router. The handshake carries the
// profile's session in the same header the HTTP client's transport uses, since
// a WebSocket dial does not go through that transport, plus the data-plane key
// when one was named.
func dialAudioStream(ctx context.Context, pc *preparedClient, token string,
	opts audioStreamOptions) (*websocket.Conn, error) {
	target, err := audioStreamURL(pc.found.BaseURL, opts)
	if err != nil {
		return nil, err
	}
	d := websocket.Dialer{HandshakeTimeout: 30 * time.Second}
	if pc.profile.InsecureSkipVerify {
		d.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicit profile opt-in
	}
	h := http.Header{}
	h.Set("X-Authorization", token)
	h.Set("X-Unauth-Error", "Non-Redirect")
	h.Set("Cookie", "auth_token="+token)
	if named := resolveDataPlaneAuth(opts.APIKey); named.Mode == authKey {
		h.Set("Authorization", "Bearer "+named.Key)
	}
	conn, resp, err := d.DialContext(ctx, target, h)
	if err != nil {
		return nil, audioStreamHandshakeError(err, resp, opts.Route)
	}
	return conn, nil
}

func audioStreamURL(baseURL string, opts audioStreamOptions) (string, error) {
	u, err := url.Parse(strings.TrimRight(baseURL, "/") + opts.Route)
	if err != nil {
		return "", fmt.Errorf("build the stream URL: %w", err)
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	}
	q := u.Query()
	if m := strings.TrimSpace(opts.Model); m != "" {
		q.Set("model", m)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// audioStreamHandshakeError turns a refused upgrade into something actionable.
// A failed handshake is an HTTP response, so Router's own envelope is in it and
// callErr can say what every other verb would have said.
func audioStreamHandshakeError(err error, resp *http.Response, route string) error {
	if resp == nil {
		return fmt.Errorf("open %s: %w", route, err)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusUpgradeRequired {
		// Router mounts the streaming routes as WebSocket only, so this means
		// something between here and it did not forward the upgrade.
		return fmt.Errorf("%s answered %d rather than upgrading; something on the way is not "+
			"forwarding WebSocket", route, resp.StatusCode)
	}
	return callErr(&RouterError{
		Method: "GET", Path: route, Status: resp.StatusCode,
		Code: routerEnvelopeCode(body), Message: routerEnvelopeMessage(body), Body: body,
	})
}

func routerEnvelopeCode(body []byte) string {
	var env struct {
		Error struct{ Code string } `json:"error"`
	}
	_ = json.Unmarshal(body, &env)
	return env.Error.Code
}

func routerEnvelopeMessage(body []byte) string {
	var env struct {
		Error struct{ Message string } `json:"error"`
	}
	_ = json.Unmarshal(body, &env)
	return env.Error.Message
}

func awaitStreamReady(conn *websocket.Conn) error {
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Minute)); err != nil {
		return err
	}
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()
	for {
		kind, payload, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("the engine closed before it was ready: %w", err)
		}
		if kind != websocket.TextMessage {
			continue
		}
		var frame audioStreamFrame
		if json.Unmarshal(payload, &frame) != nil {
			continue
		}
		switch frame.Type {
		case "ready":
			return nil
		case "error":
			// A model still loading answers here rather than at the handshake:
			// the socket is open and the engine is not.
			return fmt.Errorf("the engine is not ready: %s", nonEmpty(frame.Detail))
		}
	}
}

func sendStreamStart(conn *websocket.Conn, opts audioStreamOptions) error {
	start := map[string]any{"type": "start"}
	if opts.SampleRate > 0 {
		start["sample_rate"] = opts.SampleRate
	}
	if opts.StepMillis > 0 {
		start["step_ms"] = opts.StepMillis
	}
	if v := strings.TrimSpace(opts.Language); v != "" {
		start["language"] = v
	}
	buf, err := json.Marshal(start)
	if err != nil {
		return fmt.Errorf("marshal the start frame: %w", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, buf); err != nil {
		return fmt.Errorf("send the start frame: %w", err)
	}
	return nil
}

func sendPCM(conn *websocket.Conn, src io.Reader) error {
	buf := make([]byte, pcmFrameBytes)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if werr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
				// The engine hung up mid-upload. What it said before doing so
				// is the useful part, and the read loop has it.
				return nil
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read the audio: %w", err)
		}
	}
}

// readAudioStream renders frames until the engine closes. The `closed` frame
// carries the audio length the engine actually consumed, which is the one
// number worth reporting: it is what the call is metered on.
func readAudioStream(conn *websocket.Conn, opts audioStreamOptions) error {
	for {
		kind, payload, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure,
				websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				return nil
			}
			return fmt.Errorf("the stream ended unexpectedly: %w", err)
		}
		if kind != websocket.TextMessage {
			continue
		}
		var frame audioStreamFrame
		if json.Unmarshal(payload, &frame) != nil {
			continue
		}
		switch frame.Type {
		case "error":
			return fmt.Errorf("the engine reported: %s", nonEmpty(frame.Detail))
		case "partial":
			if !opts.Partials {
				continue
			}
			if err := opts.Render(&frame, false); err != nil {
				return err
			}
		case "final":
			if err := opts.Render(&frame, true); err != nil {
				return err
			}
		case "closed":
			if frame.AudioSeconds != nil {
				fmt.Fprintf(os.Stderr, "\n%s of audio\n", seconds(*frame.AudioSeconds))
			}
			return nil
		}
	}
}

// renderTranscriptStream rewrites one line while decoding and commits the text
// to stdout at the end. Carriage return rather than a new line per partial: the
// transcript grows in place, and a hundred lines of the same sentence being
// refined is not something to read.
func renderTranscriptStream(frame *audioStreamFrame, final bool) error {
	if !final {
		_, err := fmt.Fprintf(os.Stderr, "\r%s", clip(strings.TrimSpace(frame.Text), 100))
		return err
	}
	fmt.Fprint(os.Stderr, "\r\033[K")
	_, err := fmt.Fprintln(os.Stdout, strings.TrimSpace(frame.Text))
	return err
}

// renderDiarizationStream prints nothing until the end. A diarization partial
// relabels speakers as it learns them, so a running view would show a table
// whose rows change identity — the final frame is the one that means anything,
// and progress is reported as a count instead.
func renderDiarizationStream(frame *audioStreamFrame, final bool) error {
	if !final {
		_, err := fmt.Fprintf(os.Stderr, "\r%d segments, %d speakers so far",
			len(frame.Segments), len(frame.Speakers))
		return err
	}
	fmt.Fprint(os.Stderr, "\r\033[K")
	if len(frame.Segments) == 0 {
		_, err := fmt.Fprintln(os.Stdout, "the engine found nobody speaking.")
		return err
	}
	t := newTable(os.Stdout, "SPEAKER", "START", "END", "LENGTH")
	for i := range frame.Segments {
		s := &frame.Segments[i]
		t.row(nonEmpty(s.Speaker), seconds(s.Start), seconds(s.End), seconds(s.End-s.Start))
	}
	if err := t.flush(); err != nil {
		return err
	}
	speakers := append([]string(nil), frame.Speakers...)
	sort.Strings(speakers)
	_, err := fmt.Fprintf(os.Stderr, "\n%d speakers across %d segments  %s\n",
		len(speakers), len(frame.Segments), strings.Join(speakers, " "))
	return err
}

// runDiarizeStream is `diarize --stream`: the same job over a socket, against a
// model that declares streaming diarization rather than the offline kind.
func runDiarizeStream(ctx context.Context, f *cmdutil.Factory, path, model string,
	sampleRate int, apiKey string) error {
	return runAudioStream(ctx, f, audioStreamOptions{
		Route:      epAudioDiarizeStreamWS,
		Path:       path,
		Model:      callModel(model, categoryDiarStream),
		SampleRate: sampleRate,
		APIKey:     apiKey,
		Partials:   true,
		Render:     renderDiarizationStream,
	})
}
