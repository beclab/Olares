package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/chat/completions

type chatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model,omitempty"`
	Messages    []chatMessage `json:"messages"`
	Stream      bool          `json:"stream,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
}

type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role             string `json:"role"`
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *chatUsage `json:"usage,omitempty"`
}

type chatChunk struct {
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *chatUsage `json:"usage,omitempty"`
}

func newCallChatCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output      string
		model       string
		system      string
		images      []string
		maxTokens   int
		temperature float64
		noStream    bool
		quiet       bool
		apiKey      string
		timeout     time.Duration
	)
	cmd := &cobra.Command{
		Use:   "chat [prompt]",
		Short: "a chat completion",
		Long: `Send a prompt and print the answer.

The prompt is the argument, or standard input when there is none, so this reads
from a pipe as happily as from a keyboard.

Output streams as the model produces it, because the first useful thing about a
long answer is that it started. --no-stream waits for the whole response
instead, which is what a script wants when it is going to parse the text rather
than watch it. -o json implies it: a partial JSON object is not JSON.

Under the answer, a line names the model that served it and the tokens it spent
— the same numbers "router usage" will later bill. --quiet drops that line for
piping the answer into something else.

A model running on this Olares serves a fixed number of requests at once and
queues the rest, so an answer can be minutes away through no fault of the
prompt. --timeout is how long to allow; the wait is reported on stderr while it
lasts. Giving up here does not stop the work — the engine finishes the request
and spends the GPU on an answer nobody is left to read.

--image attaches a local file by embedding it in the request. A path would not
help: the gateway forwards the message to an upstream that cannot read this
machine's disk. The model has to support vision, which
"router provider get <provider>" reports.

Examples:
  olares-cli router call chat "summarise the CAP theorem in three lines"
  git diff | olares-cli router call chat --system "write a commit message"
  olares-cli router call chat "what is in this screenshot" --image shot.png
  olares-cli router call chat "list three colours as JSON" --no-stream --quiet
  olares-cli router call chat "hello" --model gpt-4o-mini -o json
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			prompt, err := readPromptArgs(args, "prompt")
			if err != nil {
				return err
			}
			opts := chatOptions{
				Model:    callModel(model, categoryChat),
				System:   system,
				Images:   images,
				Stream:   !noStream,
				Quiet:    quiet,
				APIKey:   apiKey,
				OutputIn: output,
				Timeout:  timeout,
			}
			if c.Flags().Changed("max-tokens") {
				opts.MaxTokens = &maxTokens
			}
			if c.Flags().Changed("temperature") {
				opts.Temperature = &temperature
			}
			return runCallChat(c.Context(), f, prompt, opts)
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryChat))
	cmd.Flags().StringVar(&system, "system", "", "system instruction sent before the prompt")
	cmd.Flags().StringArrayVar(&images, "image", nil, "attach a local image file (repeatable)")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 0, "cap the answer length in tokens")
	cmd.Flags().Float64Var(&temperature, "temperature", 0, "sampling temperature")
	cmd.Flags().BoolVar(&noStream, "no-stream", false, "wait for the whole answer instead of streaming it")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "print only the answer, without the model and token line")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute,
		"give up waiting after this long; the engine finishes the request regardless")
	addOutputFlag(cmd, &output)
	return cmd
}

type chatOptions struct {
	Model       string
	System      string
	Images      []string
	MaxTokens   *int
	Temperature *float64
	Stream      bool
	Quiet       bool
	APIKey      string
	OutputIn    string
	Timeout     time.Duration
}

func runCallChat(ctx context.Context, f *cmdutil.Factory, prompt string, opts chatOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(opts.OutputIn)
	if err != nil {
		return err
	}
	// JSON output and streaming are incompatible in the only way that matters:
	// the caller wants one object to parse, and a stream arrives as many.
	if format == FormatJSON {
		opts.Stream = false
	}
	// A completion is a long request whichever way it is read. The client's
	// own 30-second timeout used to be the shortest budget on a chain whose
	// other links allow 320 seconds, five minutes and forever, so the caller
	// gave up first on every queued request — and giving up here does not
	// recall it: the engine runs it to completion and answers nobody.
	// --timeout replaces that with a deadline the caller chose and can see.
	pc, err := prepareLongRequest(ctx, f)
	if err != nil {
		return err
	}
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}
	dp := dataPlane(pc, opts.APIKey)

	msgs := make([]chatMessage, 0, 2)
	if s := strings.TrimSpace(opts.System); s != "" {
		msgs = append(msgs, chatMessage{Role: "system", Content: s})
	}
	user, err := userMessage(prompt, opts.Images)
	if err != nil {
		return err
	}
	msgs = append(msgs, user)

	req := chatRequest{
		Model:       strings.TrimSpace(opts.Model),
		Messages:    msgs,
		Stream:      opts.Stream,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
	}
	if opts.Stream {
		return streamChat(ctx, dp, req, opts.Quiet, opts.Timeout)
	}

	stop := noticeWhileWaiting(ctx, noticeTarget(opts.Quiet), firstTokenNotice, opts.Timeout)
	var resp chatResponse
	err = dp.doJSON(ctx, "POST", epChatCompletions, req, &resp)
	stop()
	if err != nil {
		return callErr(chatWaitErr(ctx, err, opts.Timeout))
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	return renderChat(os.Stdout, &resp, opts.Quiet)
}

// firstTokenNotice is how long silence has to last before it is worth saying
// something about. Short enough that a queued request is explained before the
// person starts wondering, long enough that a model answering normally never
// prints it.
const firstTokenNotice = 3 * time.Second

// noticeWhileWaiting says once that nothing has arrived yet, and then says
// nothing more. Returns the stop function, which is safe to call repeatedly
// and after the notice has already been printed.
//
// It exists because the failure this whole change is about is silent: a
// request sitting in the engine's queue and a request being answered slowly
// look identical from here, and somebody watching an empty terminal cannot
// tell which is happening, or whether anything is.
//
// The caller passes stderr, so the answer on stdout stays pipeable — the same
// split the token footer already uses — or io.Discard under --quiet, since
// somebody who asked for the answer alone did not ask for company.
func noticeWhileWaiting(ctx context.Context, w io.Writer, after, timeout time.Duration) (stop func()) {
	done := make(chan struct{})
	var once sync.Once
	go func() {
		select {
		case <-done:
		case <-ctx.Done():
		case <-time.After(after):
			fmt.Fprintln(w, waitingNotice(timeout))
		}
	}()
	return func() { once.Do(func() { close(done) }) }
}

// noticeTarget is where the waiting line goes, which is nowhere under --quiet.
func noticeTarget(quiet bool) io.Writer {
	if quiet {
		return io.Discard
	}
	return os.Stderr
}

// waitingNotice names the budget, because "still waiting" without one leaves
// the reader with the same question they already had: how long is this
// allowed to go on.
func waitingNotice(timeout time.Duration) string {
	if timeout <= 0 {
		return "waiting for the model (no deadline; --timeout sets one)"
	}
	return fmt.Sprintf("waiting for the model (up to %s; --timeout changes that)", timeout)
}

// chatWaitErr names the deadline when it is the deadline that ended the wait.
// Without this a --timeout expiry surfaces as "context deadline exceeded",
// which reads like a bug rather than like the choice it was — and hides the
// part that matters: the request was not withdrawn, it is still running.
func chatWaitErr(ctx context.Context, err error, timeout time.Duration) error {
	if err == nil || !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("gave up after %s: the model had not answered yet. The request was not "+
		"withdrawn — a queued completion runs to the end whether or not anyone is still "+
		"reading — so --timeout with a longer budget is the way to see it, and "+
		"`olares-cli router provider get <provider>` shows how deep the queue was", timeout)
}

// userMessage builds the user turn. Text alone stays a plain string, because
// that is what every upstream accepts; images force the parts array, which some
// text-only upstreams reject outright — a difference worth not imposing on a
// request that has no images in it.
func userMessage(prompt string, images []string) (chatMessage, error) {
	if len(images) == 0 {
		return chatMessage{Role: "user", Content: prompt}, nil
	}
	parts := make([]map[string]any, 0, len(images)+1)
	parts = append(parts, map[string]any{"type": "text", "text": prompt})
	for _, path := range images {
		url, err := dataURL(path)
		if err != nil {
			return chatMessage{}, err
		}
		parts = append(parts, map[string]any{
			"type":      "image_url",
			"image_url": map[string]string{"url": url},
		})
	}
	return chatMessage{Role: "user", Content: parts}, nil
}

func renderChat(w io.Writer, resp *chatResponse, quiet bool) error {
	if len(resp.Choices) == 0 {
		_, err := fmt.Fprintln(w, "the model returned no choices.")
		return err
	}
	for i := range resp.Choices {
		ch := &resp.Choices[i]
		if r := strings.TrimSpace(ch.Message.ReasoningContent); r != "" {
			if _, err := fmt.Fprintf(w, "[reasoning]\n%s\n\n", r); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, strings.TrimRight(ch.Message.Content, "\n")); err != nil {
			return err
		}
	}
	if quiet {
		return nil
	}
	return printChatFooter(w, resp.Model, resp.Usage, resp.Choices[0].FinishReason)
}

// printChatFooter goes to stderr so the answer on stdout stays pipeable while
// the accounting stays visible.
func printChatFooter(_ io.Writer, model string, usage *chatUsage, finish string) error {
	line := "\n" + nonEmpty(model)
	if usage != nil {
		line += fmt.Sprintf("  %d prompt + %d completion = %d tokens",
			usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens)
	}
	if finish != "" && finish != "stop" {
		line += "  finish: " + finish
	}
	_, err := fmt.Fprintln(os.Stderr, line)
	return err
}

func streamChat(ctx context.Context, dp *routerClient, req chatRequest, quiet bool, timeout time.Duration) error {
	buf, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}
	// The notice covers the whole silence, not just the connection: on a
	// queued request the response headers arrive promptly and the first token
	// does not, so stopping at the response would explain the shorter half of
	// the wait.
	stopNotice := noticeWhileWaiting(ctx, noticeTarget(quiet), firstTokenNotice, timeout)
	defer stopNotice()

	resp, err := dp.do(ctx, "POST", epChatCompletions, bytes.NewReader(buf), "application/json")
	if err != nil {
		return chatWaitErr(ctx, err, timeout)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		stopNotice()
		return callErr(withRetryAfter(
			dp.formatErr("POST", epChatCompletions, resp.StatusCode, body), resp.Header))
	}

	var (
		model    string
		usage    *chatUsage
		finish   string
		inReason bool
		wrote    bool
	)
	out := os.Stdout
	err = sseLines(resp.Body, func(payload []byte) error {
		var chunk chatChunk
		if uerr := json.Unmarshal(payload, &chunk); uerr != nil {
			// A frame this build cannot parse is not worth losing the answer
			// over; the stream is the product and an unknown field is noise.
			return nil
		}
		if chunk.Model != "" {
			model = chunk.Model
		}
		if chunk.Usage != nil {
			usage = chunk.Usage
		}
		for i := range chunk.Choices {
			c := &chunk.Choices[i]
			if c.FinishReason != nil && *c.FinishReason != "" {
				finish = *c.FinishReason
			}
			if r := c.Delta.ReasoningContent; r != "" {
				if !inReason {
					stopNotice()
					fmt.Fprint(out, "[reasoning] ")
					inReason = true
				}
				fmt.Fprint(out, r)
				wrote = true
			}
			if t := c.Delta.Content; t != "" {
				if inReason {
					fmt.Fprint(out, "\n\n")
					inReason = false
				}
				stopNotice()
				fmt.Fprint(out, t)
				wrote = true
			}
		}
		return nil
	})
	stopNotice()
	if wrote {
		fmt.Fprintln(out)
	}
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return chatWaitErr(ctx, err, timeout)
		}
		return fmt.Errorf("read the answer stream: %w", err)
	}
	if quiet {
		return nil
	}
	return printChatFooter(out, model, usage, finish)
}
