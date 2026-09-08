package router

import (
	"context"
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

// `router call history …` — what a synthesis model has read out, with the
// audio still attached.
//
// GET    /v1/history           the readings, newest first
// GET    /v1/history/:id       one reading
// GET    /v1/history/:id/audio its audio, again
// DELETE /v1/history/:id
//
// The point of this surface is that fetching a reading again is free. The
// engine kept the bytes, so `history download` costs nothing and is not billed
// a second time — synthesis was paid for when it happened.
//
// Two things bound what is visible. A reading belongs to the caller that made
// it, so history is per-principal and one key cannot read another's. And
// Router has no default category for history, so every verb here takes
// --model: a reading lives in the engine that performed it, and "the history"
// is not something to resolve.

func newCallHistoryCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "what a synthesis model has read out",
		Long: `Read back what a speech model has spoken.

Every reading a synthesis engine performs is kept with its audio, so a file
that was lost, or never written because --out was forgotten, can be fetched
again. Fetching costs nothing: the synthesis was billed when it happened and
the bytes are already on disk.

--model is required on every verb here. Router resolves a default category for
speaking, cloning and designing, but not for history — a reading exists inside
the engine that performed it, so there is no model to guess.

A reading is visible to the caller that made it. Another key's readings are not
missing, they are not yours to read.

Subcommands:
  list                  the readings, newest first
  get <id>              one reading and what it was asked to say
  download <id>         its audio, again
  delete <id>           forget one
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newHistoryListCommand(f))
	cmd.AddCommand(newHistoryGetCommand(f))
	cmd.AddCommand(newHistoryDownloadCommand(f))
	cmd.AddCommand(newHistoryDeleteCommand(f))
	return cmd
}

// historyItem is one reading. `state` is "processing" while the engine is
// still speaking it, which is why a download can arrive as it is produced.
type historyItem struct {
	ID         string `json:"history_item_id"`
	DateUnix   int64  `json:"date_unix"`
	State      string `json:"state"`
	VoiceID    string `json:"voice_id"`
	VoiceName  string `json:"voice_name"`
	ModelID    string `json:"model_id"`
	Text       string `json:"text"`
	Source     string `json:"source"`
	FormatName string `json:"output_format"`
	MediaType  string `json:"content_type"`
}

func newHistoryListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		model  string
		apiKey string
		voice  string
		after  string
		limit  int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "the readings, newest first",
		Long: `List what this model has read out for you.

--voice narrows to one voice, which is how to find the take that came out right
when several were tried. --after continues from an id, since a busy engine has
more readings than one page.

Examples:
  olares-cli router call history list --model Olares/<tts-model>
  olares-cli router call history list --model Olares/<tts-model> --voice vc_01H… --limit 20
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			if err := requireHistoryModel(model); err != nil {
				return err
			}
			return runHistoryList(c.Context(), f, model, voice, after, limit, apiKey, output)
		},
	}
	addHistoryModelFlag(cmd, &model)
	cmd.Flags().StringVar(&voice, "voice", "", "only readings spoken with this voice")
	cmd.Flags().StringVar(&after, "after", "", "continue from this reading")
	cmd.Flags().IntVar(&limit, "limit", 0, "how many readings to ask for")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func runHistoryList(ctx context.Context, f *cmdutil.Factory, model, voice, after string, limit int,
	apiKey, outputRaw string) error {
	format, dp, err := voicePlane(ctx, f, outputRaw, apiKey)
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("model", strings.TrimSpace(model))
	if v := strings.TrimSpace(voice); v != "" {
		q.Set("voice_id", v)
	}
	if a := strings.TrimSpace(after); a != "" {
		q.Set("start_after_history_item_id", a)
	}
	if limit > 0 {
		q.Set("page_size", strconv.Itoa(limit))
	}
	var resp struct {
		History []historyItem `json:"history"`
		HasMore bool          `json:"has_more"`
		Last    string        `json:"last_history_item_id"`
	}
	if err := dp.doJSON(voiceCtx(ctx), http.MethodGet, withQuery(epHistory, q), nil, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	if len(resp.History) == 0 {
		_, err := fmt.Println("nothing read out yet. History is per caller, so another key's readings " +
			"are not listed here.")
		return err
	}
	t := newTable(os.Stdout, "READING", "WHEN", "VOICE", "STATE", "SAID")
	for i := range resp.History {
		h := &resp.History[i]
		t.row(h.ID, historyWhen(h.DateUnix), nonEmpty(historyVoice(h)), nonEmpty(h.State),
			clip(strings.TrimSpace(h.Text), 40))
	}
	if err := t.flush(); err != nil {
		return err
	}
	if resp.HasMore && resp.Last != "" {
		_, err = fmt.Printf("\nmore readings follow; continue with `--after %s`.\n", resp.Last)
	}
	return err
}

func newHistoryGetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		model  string
		apiKey string
	)
	cmd := &cobra.Command{
		Use:   "get <reading>",
		Short: "one reading and what it was asked to say",
		Long: `Show one reading in full: the text, the voice, and the settings it used.

"state" is "processing" while the engine is still speaking it. A reading in
that state can still be downloaded — the audio arrives as it is produced.

Example:
  olares-cli router call history get hi_01H… --model Olares/<tts-model>
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := requireHistoryModel(model); err != nil {
				return err
			}
			return runHistoryGet(c.Context(), f, args[0], model, apiKey, output)
		},
	}
	addHistoryModelFlag(cmd, &model)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func runHistoryGet(ctx context.Context, f *cmdutil.Factory, id, model, apiKey, outputRaw string) error {
	format, dp, err := voicePlane(ctx, f, outputRaw, apiKey)
	if err != nil {
		return err
	}
	var h historyItem
	if err := dp.doJSON(voiceCtx(ctx), http.MethodGet,
		voicePath(epHistoryItem(strings.TrimSpace(id)), model), nil, &h); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, h)
	}
	t := newTable(os.Stdout)
	t.row("READING", h.ID)
	t.row("WHEN", historyWhen(h.DateUnix))
	t.row("STATE", nonEmpty(h.State))
	t.row("VOICE", nonEmpty(historyVoice(&h)))
	t.row("MODEL", nonEmpty(h.ModelID))
	t.row("FORMAT", nonEmpty(h.FormatName))
	if err := t.flush(); err != nil {
		return err
	}
	if strings.TrimSpace(h.Text) != "" {
		if _, err := fmt.Printf("\nSAID:\n%s\n", strings.TrimSpace(h.Text)); err != nil {
			return err
		}
	}
	_, err = fmt.Printf("\n`olares-cli router call history download %s --model %s --out <file>` "+
		"fetches the audio again, at no cost.\n", h.ID, strings.TrimSpace(model))
	return err
}

func newHistoryDownloadCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		model   string
		apiKey  string
		outPath string
	)
	cmd := &cobra.Command{
		Use:   "download <reading>",
		Short: "its audio, again",
		Long: `Write a reading's audio to a file.

This is not a second synthesis. The engine kept the bytes when it spoke them,
so this is a copy: no model runs and nothing is billed.

--out is required, because audio is not something to put on a terminal.

Example:
  olares-cli router call history download hi_01H… --model Olares/<tts-model> --out again.wav
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := requireHistoryModel(model); err != nil {
				return err
			}
			if strings.TrimSpace(outPath) == "" {
				return fmt.Errorf("--out is required: audio does not belong on a terminal")
			}
			return runHistoryDownload(c.Context(), f, args[0], model, apiKey, outPath)
		},
	}
	addHistoryModelFlag(cmd, &model)
	cmd.Flags().StringVar(&outPath, "out", "", "write the audio here")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	return cmd
}

func runHistoryDownload(ctx context.Context, f *cmdutil.Factory, id, model, apiKey, outPath string) error {
	ctx = voiceCtx(ctx)
	pc, err := prepareLongRequest(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, apiKey)
	route := voicePath(epHistoryAudio(strings.TrimSpace(id)), model)
	resp, err := dp.do(ctx, http.MethodGet, route, nil, "")
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)
		return callErr(dp.formatErr(http.MethodGet, route, resp.StatusCode, raw))
	}
	file, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	written, err := io.Copy(file, resp.Body)
	if cerr := file.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	_, err = fmt.Printf("wrote %s (%s). Nothing was synthesised, so nothing was billed.\n",
		outPath, humanBytes(written))
	return err
}

func newHistoryDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		model  string
		apiKey string
		yes    bool
	)
	cmd := &cobra.Command{
		Use:   "delete <reading>",
		Short: "forget one reading",
		Long: `Delete a reading and its audio.

A reading still being spoken stops when it is deleted, so this is also how to
abandon a long piece of text partway through.

The spend row stays: what was billed happened, and throwing away the result
does not unbill it.

Example:
  olares-cli router call history delete hi_01H… --model Olares/<tts-model> --yes
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := requireHistoryModel(model); err != nil {
				return err
			}
			return runHistoryDelete(c.Context(), f, args[0], model, apiKey, yes)
		},
	}
	addHistoryModelFlag(cmd, &model)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addConfirmFlag(cmd, &yes)
	return cmd
}

func runHistoryDelete(ctx context.Context, f *cmdutil.Factory, id, model, apiKey string, yes bool) error {
	ctx = voiceCtx(ctx)
	if !yes {
		if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin, fmt.Sprintf(
			"Delete reading %s? Its audio goes with it and cannot be fetched again.",
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
		voicePath(epHistoryItem(strings.TrimSpace(id)), model), nil, &ignored); err != nil {
		return callErr(err)
	}
	_, err = fmt.Printf("%s is gone. The spend row it produced stays.\n", strings.TrimSpace(id))
	return err
}

// requireHistoryModel refuses early rather than letting the dispatcher answer.
// Router's error for an unresolvable reference is accurate and unhelpful here:
// the caller did not misname a model, they assumed history had a default the
// way speaking does.
func requireHistoryModel(model string) error {
	if strings.TrimSpace(model) != "" {
		return nil
	}
	return fmt.Errorf("--model is required: a reading lives in the engine that performed it, and " +
		"Router has no default category for history to resolve. " +
		"`olares-cli router model list --mode tts` names the candidates")
}

func addHistoryModelFlag(cmd *cobra.Command, model *string) {
	cmd.Flags().StringVar(model, "model", "",
		"required: the synthesis model whose history to read, as <provider>/<model> or a route name")
}

func historyVoice(h *historyItem) string {
	if name := strings.TrimSpace(h.VoiceName); name != "" {
		return name
	}
	return strings.TrimSpace(h.VoiceID)
}

func historyWhen(unix int64) string {
	if unix <= 0 {
		return "-"
	}
	return time.Unix(unix, 0).Local().Format("2006-01-02 15:04")
}
