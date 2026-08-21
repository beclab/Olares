package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/ocr, GET /v1/ocr/tasks[/:id], DELETE /v1/ocr/tasks/:id
//
// OCR is the one data-plane route that does not answer with its result. A
// submission returns a queued task, and the text arrives later — a scanned book
// is minutes of work, and an HTTP request held open for that long is a request
// that will be cut. So `call ocr` submits and then waits by polling, which is
// what a caller wants often enough to be the default; --no-wait hands back the
// task id for anyone who would rather come back for it.

type ocrTask struct {
	ID            string          `json:"id"`
	Model         string          `json:"model"`
	Status        string          `json:"status"`
	Created       float64         `json:"created"`
	Started       *float64        `json:"started,omitempty"`
	Finished      *float64        `json:"finished,omitempty"`
	QueuePosition *int            `json:"queue_position,omitempty"`
	ETAMillis     *int64          `json:"eta_ms,omitempty"`
	Progress      json.RawMessage `json:"progress,omitempty"`
	ResultKind    *string         `json:"result_kind,omitempty"`
	Result        json.RawMessage `json:"result,omitempty"`
	Error         *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (t *ocrTask) done() bool {
	switch strings.ToLower(t.Status) {
	case "succeeded", "failed", "canceled", "cancelled":
		return true
	}
	return false
}

func newCallOCRCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output      string
		model       string
		format      string
		pages       string
		pdfStrategy string
		noWait      bool
		timeout     time.Duration
		queueStatus string
		queueLimit  int
		apiKey      string
	)
	cmd := &cobra.Command{
		Use:   "ocr <image-or-pdf>",
		Short: "read the text out of an image or PDF",
		Long: `Extract text from a document.

The file is submitted as a task and this waits for it, reporting the queue
position while it waits. A PDF comes back page by page; an image comes back as
one block of text.

--no-wait returns the task id instead of waiting. "router call ocr --task <id>"
picks it up later, and "--task <id> --cancel" drops it.

--queue lists the engine's board and reads nothing: what is waiting, what is
running, and whether it is still accepting work. A queue that is full refuses an
upload, so this is what to look at when a submission comes back rejected rather
than slow.

--pages narrows the work on a long PDF, which is the difference between seconds
and minutes. --pdf-strategy and --format are passed to the engine untouched.

Examples:
  olares-cli router call ocr invoice.png
  olares-cli router call ocr book.pdf --pages 1-10
  olares-cli router call ocr scan.pdf --no-wait
  olares-cli router call ocr --task 9f2c1b40
  olares-cli router call ocr --task 9f2c1b40 --cancel
  olares-cli router call ocr --queue
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			task, _ := c.Flags().GetString("task")
			cancel, _ := c.Flags().GetBool("cancel")
			if queue, _ := c.Flags().GetBool("queue"); queue {
				if len(args) > 0 || strings.TrimSpace(task) != "" {
					return fmt.Errorf("--queue lists the board; it takes neither a file nor --task")
				}
				return runOCRQueue(c.Context(), f, ocrQueueOptions{
					Status: strings.TrimSpace(queueStatus), Limit: queueLimit,
					APIKey: apiKey, OutputIn: output,
				})
			}
			opts := ocrOptions{
				Model:       callModel(model, categoryOCR),
				Format:      format,
				Pages:       pages,
				PDFStrategy: pdfStrategy,
				Wait:        !noWait,
				Timeout:     timeout,
				APIKey:      apiKey,
				OutputIn:    output,
				TaskID:      strings.TrimSpace(task),
				Cancel:      cancel,
			}
			if opts.TaskID == "" && len(args) == 0 {
				return fmt.Errorf("name a file to read, or a task to pick up with --task")
			}
			if opts.TaskID != "" && len(args) > 0 {
				return fmt.Errorf("--task picks up an existing job; it takes no file")
			}
			if len(args) > 0 {
				opts.Path = args[0]
			}
			return runCallOCR(c.Context(), f, opts)
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryOCR))
	cmd.Flags().StringVar(&format, "format", "", "output format the engine should produce")
	cmd.Flags().StringVar(&pages, "pages", "", "which PDF pages to read, e.g. 1-10")
	cmd.Flags().StringVar(&pdfStrategy, "pdf-strategy", "", "how the engine should treat a PDF")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "submit and print the task id instead of waiting")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "give up waiting after this long; the task keeps running")
	cmd.Flags().String("task", "", "pick up a task submitted earlier")
	cmd.Flags().Bool("cancel", false, "with --task, drop it instead of reading it")
	cmd.Flags().Bool("queue", false, "list what the engine is working on and read nothing")
	cmd.Flags().StringVar(&queueStatus, "status", "", "with --queue, only queued, running, succeeded, failed or cancelled")
	cmd.Flags().IntVar(&queueLimit, "limit", 0, "with --queue, how many tasks to list")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

type ocrOptions struct {
	Path        string
	Model       string
	Format      string
	Pages       string
	PDFStrategy string
	Wait        bool
	Timeout     time.Duration
	APIKey      string
	OutputIn    string
	TaskID      string
	Cancel      bool
}

func runCallOCR(ctx context.Context, f *cmdutil.Factory, opts ocrOptions) error {
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
	dp := dataPlane(pc, opts.APIKey)

	if opts.TaskID != "" && opts.Cancel {
		path := epOCRTask(opts.TaskID)
		if err := dp.doJSON(ctx, "DELETE", path, nil, nil); err != nil {
			return callErr(err)
		}
		fmt.Printf("dropped task %s\n", opts.TaskID)
		return nil
	}

	var task ocrTask
	switch {
	case opts.TaskID != "":
		if err := fetchOCRTask(ctx, dp, opts.TaskID, &task); err != nil {
			return err
		}
	default:
		body, contentType, merr := multipartFile(opts.Path, "file", map[string]string{
			"model":        strings.TrimSpace(opts.Model),
			"format":       strings.TrimSpace(opts.Format),
			"pages":        strings.TrimSpace(opts.Pages),
			"pdf_strategy": strings.TrimSpace(opts.PDFStrategy),
		})
		if merr != nil {
			return merr
		}
		resp, derr := dp.do(ctx, "POST", epOCR, body, contentType)
		if derr != nil {
			return derr
		}
		raw, rerr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if rerr != nil {
			return fmt.Errorf("read the submission answer: %w", rerr)
		}
		if resp.StatusCode/100 != 2 {
			return callErr(dp.formatErr("POST", epOCR, resp.StatusCode, raw))
		}
		if uerr := json.Unmarshal(raw, &task); uerr != nil {
			return fmt.Errorf("decode the submission answer: %w (body=%s)", uerr, truncate(string(raw), 200))
		}
		if !opts.Wait {
			if format == FormatJSON {
				return printJSON(os.Stdout, task)
			}
			fmt.Printf("submitted as task %s (%s)\n", task.ID, nonEmpty(task.Status))
			fmt.Printf("`olares-cli router call ocr --task %s` reads it when it is done\n", task.ID)
			return nil
		}
	}

	if !task.done() {
		if err := waitForOCRTask(ctx, dp, &task, opts.Timeout, format == FormatTable); err != nil {
			return err
		}
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, task)
	}
	return renderOCRTask(os.Stdout, &task)
}

// GET /v1/ocr/tasks
//
// The board, and the one thing on it a caller acts on: `accepting`. A full queue
// refuses an upload outright rather than making it wait, so a rejected
// submission and a slow one are different problems and this is what tells them
// apart.
//
// The finished rows are history — the engine keeps them for a while after the
// text has been read — so a long list here is not a backlog. `queued` and
// `running` are the backlog.
type ocrQueue struct {
	Capacity struct {
		Queued    int  `json:"queued"`
		Running   int  `json:"running"`
		Limit     int  `json:"limit"`
		Accepting bool `json:"accepting"`
	} `json:"capacity"`
	Truncated bool      `json:"truncated"`
	Data      []ocrTask `json:"data"`
}

type ocrQueueOptions struct {
	Status   string
	Limit    int
	APIKey   string
	OutputIn string
}

func runOCRQueue(ctx context.Context, f *cmdutil.Factory, opts ocrQueueOptions) error {
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
	dp := dataPlane(pc, opts.APIKey)
	q := url.Values{}
	if opts.Status != "" {
		q.Set("status", opts.Status)
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	var board ocrQueue
	if err := dp.doJSON(ctx, "GET", withQuery(epOCRTasks, q), nil, &board); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, board)
	}
	return renderOCRQueue(os.Stdout, &board)
}

func renderOCRQueue(w io.Writer, board *ocrQueue) error {
	c := board.Capacity
	state := "accepting work"
	if !c.Accepting {
		state = "full — an upload now is refused, not queued"
	}
	if _, err := fmt.Fprintf(w, "%d waiting, %d running, room for %d: %s\n\n",
		c.Queued, c.Running, c.Limit, state); err != nil {
		return err
	}
	if len(board.Data) == 0 {
		_, err := fmt.Fprintln(w, "no tasks.")
		return err
	}
	t := newTable(w, "TASK", "STATUS", "MODEL", "AHEAD", "WAIT", "AGE")
	for i := range board.Data {
		task := &board.Data[i]
		ahead, wait := "-", "-"
		if task.QueuePosition != nil {
			ahead = strconv.Itoa(*task.QueuePosition)
		}
		if task.ETAMillis != nil {
			wait = (time.Duration(*task.ETAMillis) * time.Millisecond).Round(time.Second).String()
		}
		t.row(task.ID, nonEmpty(task.Status), nonEmpty(task.Model), ahead, wait, ocrTaskAge(task))
	}
	if err := t.flush(); err != nil {
		return err
	}
	if board.Truncated {
		_, err := fmt.Fprintln(os.Stderr, "\nfinished tasks were left out to fit the limit; --limit raises it")
		return err
	}
	return nil
}

// ocrTaskAge is how long ago the task arrived. Unix seconds as a float is the
// engine's own unit here, and zero means it did not say.
func ocrTaskAge(task *ocrTask) string {
	if task.Created <= 0 {
		return "-"
	}
	at := time.Unix(int64(task.Created), 0)
	return time.Since(at).Round(time.Second).String() + " ago"
}

func fetchOCRTask(ctx context.Context, dp *routerClient, id string, out *ocrTask) error {
	path := epOCRTask(id)
	if err := dp.doJSON(ctx, "GET", path, nil, out); err != nil {
		return callErr(err)
	}
	return nil
}

// waitForOCRTask polls until the task settles. Progress goes to stderr so the
// text on stdout stays clean, and a timeout leaves the task running rather than
// cancelling it: work already done is worth picking up later.
func waitForOCRTask(ctx context.Context, dp *routerClient, task *ocrTask, timeout time.Duration, verbose bool) error {
	deadline := time.Now().Add(timeout)
	lastNote := ""
	for {
		if verbose {
			if note := ocrWaitNote(task); note != "" && note != lastNote {
				fmt.Fprintln(os.Stderr, note)
				lastNote = note
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("task %s is still %s after %s; it keeps running — "+
				"`olares-cli router call ocr --task %s` picks it up",
				task.ID, nonEmpty(task.Status), timeout, task.ID)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
		var next ocrTask
		if err := fetchOCRTask(ctx, dp, task.ID, &next); err != nil {
			return err
		}
		*task = next
		if task.done() {
			return nil
		}
	}
}

func ocrWaitNote(task *ocrTask) string {
	note := "task " + task.ID + ": " + nonEmpty(task.Status)
	if task.QueuePosition != nil {
		note += fmt.Sprintf(", %d ahead", *task.QueuePosition)
	}
	if task.ETAMillis != nil {
		note += fmt.Sprintf(", about %s to wait", (time.Duration(*task.ETAMillis) * time.Millisecond).Round(time.Second))
	}
	return note
}

func renderOCRTask(w io.Writer, task *ocrTask) error {
	switch strings.ToLower(task.Status) {
	case "failed":
		msg := "the engine did not say why"
		if task.Error != nil && task.Error.Message != "" {
			msg = task.Error.Message
		}
		return fmt.Errorf("task %s failed: %s", task.ID, msg)
	case "canceled", "cancelled":
		return fmt.Errorf("task %s was dropped before it finished", task.ID)
	}
	if len(task.Result) == 0 {
		_, err := fmt.Fprintf(w, "task %s is %s and carries no text.\n", task.ID, nonEmpty(task.Status))
		return err
	}
	if err := printOCRText(w, task.Result); err != nil {
		return err
	}
	took := ""
	if task.Started != nil && task.Finished != nil {
		took = fmt.Sprintf("  %s", (time.Duration((*task.Finished - *task.Started) * float64(time.Second))).Round(time.Millisecond))
	}
	_, err := fmt.Fprintf(os.Stderr, "\n%s%s\n", nonEmpty(task.Model), took)
	return err
}

// printOCRText handles both shapes the engines answer with: one string for an
// image, and page number → string for a PDF. Page order is numeric, because a
// document read in lexical page order is a document read wrong.
func printOCRText(w io.Writer, raw json.RawMessage) error {
	var plain string
	if json.Unmarshal(raw, &plain) == nil {
		_, err := fmt.Fprintln(w, strings.TrimRight(plain, "\n"))
		return err
	}
	var pages map[string]string
	if err := json.Unmarshal(raw, &pages); err != nil {
		return printJSON(w, raw)
	}
	keys := make([]string, 0, len(pages))
	for k := range pages {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ni, erri := strconv.Atoi(keys[i])
		nj, errj := strconv.Atoi(keys[j])
		if erri == nil && errj == nil {
			return ni < nj
		}
		return keys[i] < keys[j]
	})
	// Written straight through: this is the recognised text, and a tab inside it
	// is part of what the page said rather than a column to align.
	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "--- page %s ---\n%s\n\n", k, strings.TrimRight(pages[k], "\n")); err != nil {
			return err
		}
	}
	return nil
}
