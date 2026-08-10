package model

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

// `olares-cli model call …` — the data plane.
//
// Everything else in this tree configures Router. These verbs use it: one call
// each to the OpenAI-shaped routes it serves, whether the model behind them is
// a cloud provider or a model application on this machine. Which it is does not
// change the request, and that is the point of the gateway.
//
// A call is metered. It lands in `model usage`, it counts against whatever
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

Leaving --model off uses the workspace default for that kind of work, which
"olares-cli model default show" reports and an admin sets.

Credentials resolve on their own. Inside the cluster the platform supplies the
caller identity and no key is needed; anywhere else this machine keeps one key
in the OS keychain, minting it on first use. --api-key overrides both, and
"olares-cli model key local" says which one is in play.

Subcommands:
  chat [prompt]         a chat completion, streamed by default
  embed <text…>         embedding vectors
  transcribe <file>     speech to text
  speak <text>          text to speech
  ocr <file>            text out of an image or PDF

Every call is metered: it appears in "model usage", counts against the quota on
the credential that made it, and may cost money.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newCallChatCommand(f))
	cmd.AddCommand(newCallEmbedCommand(f))
	cmd.AddCommand(newCallTranscribeCommand(f))
	cmd.AddCommand(newCallSpeakCommand(f))
	cmd.AddCommand(newCallOCRCommand(f))
	return cmd
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
