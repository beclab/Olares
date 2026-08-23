package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// GET /v1/audio/tasks, GET /v1/audio/tasks/:id[/result], DELETE /v1/audio/tasks/:id
//
// Every audio verb answers with its result, and takes --async to answer with a
// receipt instead. That is not a convenience: an engine transcribing an hour of
// audio holds the connection for as long as the work takes, and a request held
// open that long is one an edge, a proxy or a laptop lid will cut. The task
// survives the connection; the sync request does not.
//
// Where `--async` goes in the request is the engine's choice rather than this
// tree's, and it differs by route: the multipart routes read an `async` form
// field, the JSON ones read an `async` query parameter. Sending it in the wrong
// place is not an error — it is silently ignored, and the caller waits for the
// long answer they thought they had escaped.
//
// A task lives on the engine that created it and nowhere else, so reading one
// means reaching the same provider. Router remembers which backend answered each
// `--async` submission and routes a bare lookup back to it, so an id is usually
// enough. `--model` is what covers the cases that memory cannot: a gateway that
// has restarted, a task minted through a different one, or a result old enough
// that the id has aged out. When it is given it is sent as `?model=`, exactly as
// the submitting call did.

type audioTask struct {
	ID            string          `json:"id"`
	Kind          string          `json:"kind"`
	Cap           string          `json:"cap"`
	Model         string          `json:"model"`
	Status        string          `json:"status"`
	Created       float64         `json:"created"`
	Started       *float64        `json:"started,omitempty"`
	Finished      *float64        `json:"finished,omitempty"`
	QueuePosition *int            `json:"queue_position,omitempty"`
	Progress      *audioProgress  `json:"progress,omitempty"`
	ResultKind    *string         `json:"result_kind,omitempty"`
	Result        json.RawMessage `json:"result,omitempty"`
	ContentType   string          `json:"content_type,omitempty"`
	ResultBytes   *int64          `json:"result_bytes,omitempty"`
	Error         *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type audioProgress struct {
	Stage string   `json:"stage"`
	Ratio *float64 `json:"ratio,omitempty"`
	Done  *int64   `json:"done,omitempty"`
	Total *int64   `json:"total,omitempty"`
}

// audioAccepted is the 202 body: the task nested under one key, which is what
// separates a receipt from a result on routes that can answer with either.
type audioAccepted struct {
	Task audioTask `json:"task"`
}

func (t *audioTask) settled() bool {
	switch strings.ToLower(t.Status) {
	case "succeeded", "failed", "canceled", "cancelled":
		return true
	}
	return false
}

// binary reports whether the result is bytes rather than JSON, which decides
// whether `task result` can print it or has to be told where to put it.
func (t *audioTask) binary() bool {
	return t.ResultKind != nil && strings.EqualFold(*t.ResultKind, "binary")
}

func newCallTaskCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "read an audio job submitted with --async",
		Long: `Pick up work an audio verb handed back instead of finishing.

Any audio verb takes --async and answers with a task id. These read it: "get"
says how it is going, "result" collects it, "cancel" drops it, and "list" shows
the engine's whole board.

A task exists only inside the engine that is running it, and Router remembers
which one that was: an id is enough for "get", "result" and "cancel". Pass
--model when it is not — after a Router restart, or for work submitted through
another gateway — and it has to be the model the work was submitted to, because
a task read against a different engine is a 404 for a job that is running
perfectly well.

The model shown in a submission receipt is the routing reference needed to find
that engine again, not the engine's canonical model id.

"list" is the exception: a board has no id to remember, so it always needs
--model to say whose queue to show.

Examples:
  olares-cli router call transcribe meeting.m4a --async
  olares-cli router call task get tsk_1f3c
  olares-cli router call task result tsk_1f3c > meeting.txt
  olares-cli router call task list --model default-stt --status running
  olares-cli router call task cancel tsk_1f3c --model default-stt
`,
	}
	cmd.AddCommand(newCallTaskGetCommand(f))
	cmd.AddCommand(newCallTaskResultCommand(f))
	cmd.AddCommand(newCallTaskCancelCommand(f))
	cmd.AddCommand(newCallTaskListCommand(f))
	return cmd
}

func newCallTaskGetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output  string
		model   string
		apiKey  string
		wait    bool
		timeout time.Duration
	)
	cmd := &cobra.Command{
		Use:   "get <task-id>",
		Short: "how an audio job is going",
		Long: `Report a task's status, and its result when it carries one.

A task that finished with JSON — a transcript, a diarization, a set of
intervals — carries the result here, so this is usually the only call needed.
An audio result is bytes and does not travel in a status; "task result"
collects that.

--wait polls until it settles rather than answering once.

Examples:
  olares-cli router call task get tsk_1f3c
  olares-cli router call task get tsk_1f3c --wait
  olares-cli router call task get tsk_1f3c --model default-stt
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			return runAudioTaskGet(c.Context(), f, audioTaskOptions{
				ID: args[0], Model: model, APIKey: apiKey,
				Wait: wait, Timeout: timeout, Format: format,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", audioTaskModelFlagUsage)
	cmd.Flags().BoolVar(&wait, "wait", false, "poll until the task settles")
	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Minute,
		"with --wait, give up watching after this long; the task keeps running")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func newCallTaskResultCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output  string
		model   string
		apiKey  string
		outPath string
	)
	cmd := &cobra.Command{
		Use:   "result <task-id>",
		Short: "collect what an audio job produced",
		Long: `Fetch a finished task's result.

What comes back depends on what was asked for: a transcript or a set of
segments arrives as JSON, and a synthesis or an enhancement arrives as audio.
Audio goes to --out, or to standard output when that is a pipe; writing it into
a terminal is refused rather than done.

A task that has not succeeded has no result, and this says so rather than
waiting — for one still running as well as for one that failed. "task get
--wait" is what waits. A result the engine has already dropped is reported as
gone rather than as missing: it is kept for a while after the work finishes.
If this fetch fails, check the task status before retrying: another successful
fetch can bill the same result again.

Examples:
  olares-cli router call task result tsk_1f3c
  olares-cli router call task result tsk_9a02 --out speech.mp3
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			return runAudioTaskResult(c.Context(), f, audioTaskOptions{
				ID: args[0], Model: model, APIKey: apiKey,
				Out: outPath, Format: format,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", audioTaskModelFlagUsage)
	cmd.Flags().StringVar(&outPath, "out", "", "write an audio result here instead of standard output")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func newCallTaskCancelCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		model  string
		apiKey string
	)
	cmd := &cobra.Command{
		Use:   "cancel <task-id>",
		Short: "drop an audio job",
		Long: `Stop a task, or forget a finished one.

A running task is asked to stop at its next checkpoint; a finished one is
dropped along with whatever it produced. Either way the id stops answering.

Examples:
  olares-cli router call task cancel tsk_1f3c
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAudioTaskCancel(c.Context(), f, audioTaskOptions{
				ID: args[0], Model: model, APIKey: apiKey,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", audioTaskModelFlagUsage)
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	return cmd
}

func newCallTaskListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		model  string
		apiKey string
		status string
		limit  int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "what an audio engine is working on",
		Long: `List an engine's tasks, and say whether it is still taking work.

A full queue refuses a submission outright rather than making it wait, so this
is what tells a rejected call apart from a slow one. The finished rows are
history the engine keeps for a while; "queued" and "running" are the backlog.

The board belongs to one engine, which is why --model is needed here too. There
is no view across all of them: each audio application runs its own queue.

Examples:
  olares-cli router call task list --model default-stt
  olares-cli router call task list --model default-tts --status running
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			return runAudioTaskList(c.Context(), f, audioTaskListOptions{
				Model: model, APIKey: apiKey, Status: status,
				Limit: limit, Format: format,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", audioTaskModelFlagUsage)
	cmd.Flags().StringVar(&status, "status", "", "only queued, running, succeeded, failed or canceled")
	cmd.Flags().IntVar(&limit, "limit", 0, "how many tasks to list")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

const audioTaskModelFlagUsage = "the model the task was submitted to, as <provider>/<model>, " +
	"a route name or the default category the verb used; only needed when Router " +
	"no longer remembers the task"

type audioTaskOptions struct {
	ID      string
	Model   string
	APIKey  string
	Out     string
	Wait    bool
	Timeout time.Duration
	Format  Format
}

type audioTaskListOptions struct {
	Model  string
	APIKey string
	Status string
	Limit  int
	Format Format
}

// audioTaskPath adds the model, when there is one, so the lookup reaches the
// engine holding the task. An empty one is left out rather than sent blank:
// Router routes a bare lookup by the id it remembers, and if it does not
// remember it the refusal then names the missing model instead of an empty one.
func audioTaskPath(path, model string) string {
	q := url.Values{}
	if m := strings.TrimSpace(model); m != "" {
		q.Set("model", m)
	}
	return withQuery(path, q)
}

// audioTaskErr says which of the four ways a task lookup fails happened, since
// the status alone reads the same for a job that is running well and one that
// is gone: a request Router could not place, a task nothing has, a task with no
// result yet, and a result that has been dropped.
func audioTaskErr(err error, id string) error {
	if err == nil {
		return nil
	}
	re := routerErrorOf(err)
	if re == nil {
		return callErr(err)
	}
	switch {
	case re.Code == "model_required":
		return fmt.Errorf("%w\nRouter does not remember task %s — it restarted, or the work was "+
			"submitted through another gateway — and an id alone does not say which engine is "+
			"running it. Name the model the work was submitted to: --model <name>. The verb "+
			"that submitted it printed the whole command", err, id)
	case re.Status == 404:
		return fmt.Errorf("%w\nEither the engine has forgotten task %s — a finished result is kept "+
			"for a while and then dropped — or --model names a different engine from the one "+
			"the work was submitted to", err, id)
	case re.Status == 409:
		// The engine has the task and no result to give: it is still working,
		// or it stopped without producing one. Which of the two is in the
		// status, so that is where this points.
		return fmt.Errorf("%w\nTask %s has no result yet. `olares-cli router call task get %s "+
			"--wait` waits for it and says why if the work failed or was canceled instead",
			err, id, id)
	case re.Status == 410:
		return fmt.Errorf("%w\nTask %s finished, and the engine has since dropped what it "+
			"produced — a result is kept for a while, not forever. Submit the work again",
			err, id)
	}
	return callErr(err)
}

func runAudioTaskGet(ctx context.Context, f *cmdutil.Factory, opts audioTaskOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, opts.APIKey)

	var task audioTask
	if err := fetchAudioTask(ctx, dp, opts.ID, opts.Model, &task); err != nil {
		return err
	}
	if opts.Wait && !task.settled() {
		if err := waitForAudioTask(ctx, dp, &task, opts.Model, opts.Timeout,
			opts.Format == FormatTable); err != nil {
			return err
		}
	}
	if opts.Format == FormatJSON {
		return printJSON(os.Stdout, task)
	}
	return renderAudioTask(os.Stdout, &task)
}

func fetchAudioTask(ctx context.Context, dp *routerClient, id, model string, out *audioTask) error {
	path := audioTaskPath(epAudioTask(id), model)
	if err := dp.doJSON(ctx, "GET", path, nil, out); err != nil {
		return audioTaskErr(err, id)
	}
	return nil
}

// waitForAudioTask polls until the task settles. Progress goes to stderr so a
// result on stdout stays pipeable, and a timeout leaves the task running: the
// work already done is worth collecting later.
func waitForAudioTask(ctx context.Context, dp *routerClient, task *audioTask,
	model string, timeout time.Duration, verbose bool) error {
	return waitForAudioTaskWith(ctx, task, model, timeout, verbose, audioTaskWaitOps{
		now: time.Now,
		sleep: func(ctx context.Context, delay time.Duration) error {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		},
		fetch: func(ctx context.Context, id, model string, out *audioTask) error {
			return fetchAudioTask(ctx, dp, id, model, out)
		},
	})
}

type audioTaskWaitOps struct {
	now   func() time.Time
	sleep func(context.Context, time.Duration) error
	fetch func(context.Context, string, string, *audioTask) error
}

func waitForAudioTaskWith(ctx context.Context, task *audioTask, model string,
	timeout time.Duration, verbose bool, ops audioTaskWaitOps) error {
	deadline := ops.now().Add(timeout)
	lastNote := ""
	delay := time.Second
	timeoutErr := func() error {
		return fmt.Errorf("task %s is still %s after %s; it keeps running — "+
			"`olares-cli router call task get %s --model %s` picks it up",
			task.ID, nonEmpty(task.Status), timeout, task.ID, nonEmpty(model))
	}
	for {
		note := audioTaskNote(task)
		if verbose {
			if note != lastNote {
				fmt.Fprintln(os.Stderr, note)
			}
		}
		lastNote = note
		remaining := deadline.Sub(ops.now())
		if remaining <= 0 {
			return timeoutErr()
		}
		if remaining <= delay {
			if err := ctx.Err(); err != nil {
				return err
			}
			var final audioTask
			if err := ops.fetch(ctx, task.ID, model, &final); err != nil {
				return err
			}
			*task = final
			if task.settled() {
				return nil
			}
			return timeoutErr()
		}
		if err := ops.sleep(ctx, delay); err != nil {
			return err
		}
		if !ops.now().Before(deadline) {
			return timeoutErr()
		}
		var next audioTask
		if err := ops.fetch(ctx, task.ID, model, &next); err != nil {
			return err
		}
		progressed := audioTaskNote(&next) != note
		*task = next
		if task.settled() {
			return nil
		}
		delay = nextAudioPollDelay(delay, progressed)
	}
}

func nextAudioPollDelay(current time.Duration, progressed bool) time.Duration {
	if progressed {
		return time.Second
	}
	current += time.Second
	if current > 5*time.Second {
		return 5 * time.Second
	}
	return current
}

func audioTaskNote(task *audioTask) string {
	note := "task " + task.ID + ": " + nonEmpty(task.Status)
	if task.QueuePosition != nil {
		note += fmt.Sprintf(", %d ahead", *task.QueuePosition)
	}
	if p := task.Progress; p != nil {
		if p.Stage != "" {
			note += ", " + p.Stage
		}
		switch {
		case p.Done != nil && p.Total != nil && *p.Total > 0:
			note += fmt.Sprintf(" %d/%d", *p.Done, *p.Total)
		case p.Ratio != nil:
			note += fmt.Sprintf(" %.0f%%", *p.Ratio*100)
		}
	}
	return note
}

func renderAudioTask(w io.Writer, task *audioTask) error {
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
	if !task.settled() {
		_, err := fmt.Fprintln(w, audioTaskNote(task))
		return err
	}
	if len(task.Result) > 0 {
		if err := printRawJSON(w, task.Result); err != nil {
			return err
		}
	} else if task.binary() {
		size := ""
		if task.ResultBytes != nil {
			size = " (" + humanBytes(*task.ResultBytes) + ")"
		}
		if _, err := fmt.Fprintf(w, "task %s succeeded and produced %s%s; "+
			"`olares-cli router call task result %s --model %s` collects it\n",
			task.ID, nonEmpty(task.ContentType), size, task.ID, nonEmpty(task.Model)); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "task %s succeeded and carries no result.\n", task.ID); err != nil {
			return err
		}
	}
	took := ""
	if task.Started != nil && task.Finished != nil {
		took = "  " + time.Duration((*task.Finished-*task.Started)*float64(time.Second)).
			Round(time.Millisecond).String()
	}
	_, err := fmt.Fprintf(os.Stderr, "\n%s  %s%s\n", nonEmpty(task.Cap), nonEmpty(task.Model), took)
	return err
}

func runAudioTaskResult(ctx context.Context, f *cmdutil.Factory, opts audioTaskOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepareLongRequest(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, opts.APIKey)

	path := audioTaskPath(epAudioTaskResult(opts.ID), opts.Model)
	resp, err := dp.do(ctx, "GET", path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)
		return audioTaskErr(dp.formatErr("GET", path, resp.StatusCode, raw), opts.ID)
	}
	// The engine answers a JSON result with JSON and an audio one with the
	// bytes, and says which in the Content-Type. Reading a transcript into
	// memory is fine; an enhanced recording is streamed.
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "json") {
		raw, rerr := io.ReadAll(resp.Body)
		if rerr != nil {
			return fmt.Errorf("read the result: %w", rerr)
		}
		if opts.Format == FormatJSON {
			return printRawJSON(os.Stdout, raw)
		}
		return printTranscript(os.Stdout, raw)
	}
	dst := io.Writer(os.Stdout)
	if p := strings.TrimSpace(opts.Out); p != "" {
		fh, ferr := os.Create(p)
		if ferr != nil {
			return ferr
		}
		defer fh.Close()
		dst = fh
	} else if isTerminal(os.Stdout) {
		return fmt.Errorf("task %s produced %s; name a file with --out, or pipe the output",
			opts.ID, nonEmpty(resp.Header.Get("Content-Type")))
	}
	n, err := io.Copy(dst, resp.Body)
	if err != nil {
		return fmt.Errorf("write the result: %w", err)
	}
	if p := strings.TrimSpace(opts.Out); p != "" {
		fmt.Fprintf(os.Stderr, "wrote %s (%s)\n", p, humanBytes(n))
	}
	return nil
}

func runAudioTaskCancel(ctx context.Context, f *cmdutil.Factory, opts audioTaskOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, opts.APIKey)
	var out struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Dropped   bool   `json:"dropped"`
		Canceling bool   `json:"canceling"`
	}
	path := audioTaskPath(epAudioTask(opts.ID), opts.Model)
	if err := dp.doJSON(ctx, "DELETE", path, nil, &out); err != nil {
		return audioTaskErr(err, opts.ID)
	}
	if out.Canceling {
		fmt.Printf("asked the engine to stop task %s; it stops at its next checkpoint\n", opts.ID)
		return nil
	}
	fmt.Printf("dropped task %s (%s)\n", opts.ID, nonEmpty(out.Status))
	return nil
}

// audioTaskBoard is one engine's queue. `accepting` is the field a caller acts
// on: a full queue refuses a submission rather than queueing it.
type audioTaskBoard struct {
	Running  string `json:"running"`
	Capacity struct {
		Queued    int  `json:"queued"`
		Running   int  `json:"running"`
		Limit     int  `json:"limit"`
		Accepting bool `json:"accepting"`
	} `json:"capacity"`
	Truncated bool        `json:"truncated"`
	Data      []audioTask `json:"data"`
}

func runAudioTaskList(ctx context.Context, f *cmdutil.Factory, opts audioTaskListOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, opts.APIKey)
	q := url.Values{}
	if m := strings.TrimSpace(opts.Model); m != "" {
		q.Set("model", m)
	}
	if s := strings.TrimSpace(opts.Status); s != "" {
		q.Set("status", s)
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	var board audioTaskBoard
	if err := dp.doJSON(ctx, "GET", withQuery(epAudioTasks, q), nil, &board); err != nil {
		return audioTaskErr(err, "")
	}
	if opts.Format == FormatJSON {
		return printJSON(os.Stdout, board)
	}
	return renderAudioTaskBoard(os.Stdout, &board)
}

func renderAudioTaskBoard(w io.Writer, board *audioTaskBoard) error {
	c := board.Capacity
	state := "accepting work"
	if !c.Accepting {
		state = "full — a submission now is refused, not queued"
	}
	if _, err := fmt.Fprintf(w, "%d waiting, %d running, room for %d: %s\n\n",
		c.Queued, c.Running, c.Limit, state); err != nil {
		return err
	}
	if len(board.Data) == 0 {
		_, err := fmt.Fprintln(w, "no tasks.")
		return err
	}
	t := newTable(w, "TASK", "STATUS", "CAPABILITY", "MODEL", "AHEAD", "AGE")
	for i := range board.Data {
		task := &board.Data[i]
		ahead := "-"
		if task.QueuePosition != nil {
			ahead = strconv.Itoa(*task.QueuePosition)
		}
		t.row(task.ID, nonEmpty(task.Status), nonEmpty(task.Cap), nonEmpty(task.Model),
			ahead, audioTaskAge(task))
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

func audioTaskAge(task *audioTask) string {
	if task.Created <= 0 {
		return "-"
	}
	return time.Since(time.Unix(int64(task.Created), 0)).Round(time.Second).String() + " ago"
}

const (
	audioAsyncFormField = "async"
	audioAsyncQueryKey  = "async"
)

// Engines read async from multipart form data; its query copy only helps Router route the request.
func audioMultipartFields(model string, async bool, fields map[string]string) map[string]string {
	if fields == nil {
		fields = make(map[string]string)
	}
	fields["model"] = strings.TrimSpace(model)
	if async {
		fields[audioAsyncFormField] = "1"
	} else {
		delete(fields, audioAsyncFormField)
	}
	return fields
}

// receiptFrom decodes the 202 an --async submission answers with. A non-202 is
// not an error here: the caller asked for a task and the engine may still have
// answered with the result, which every verb knows how to print.
func receiptFrom(status int, raw []byte) (*audioTask, bool) {
	if status != 202 {
		return nil, false
	}
	var accepted audioAccepted
	if json.Unmarshal(raw, &accepted) != nil || strings.TrimSpace(accepted.Task.ID) == "" {
		return nil, false
	}
	return &accepted.Task, true
}

// printReceipt reports a submitted task and the command that reads it. The
// model printed is the one the caller sent rather than the one the engine
// reports: the engine names its own weights, and what the task verbs need is
// the reference that routes back to it.
func printReceipt(w io.Writer, task *audioTask, model string, format Format) error {
	ref := strings.TrimSpace(model)
	if ref == "" {
		ref = task.Model
	}
	if format == FormatJSON {
		receipt := *task
		receipt.Model = ref
		return printJSON(w, &receipt)
	}
	if _, err := fmt.Fprintf(w, "submitted as task %s (%s)\n", task.ID, nonEmpty(task.Status)); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "`olares-cli router call task get %s --model %s --wait` follows it\n",
		task.ID, ref)
	return err
}
