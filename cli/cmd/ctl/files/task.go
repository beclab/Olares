package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/internal/files/archive"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewTaskCommand builds the `olares-cli files task` verb group —
// lifecycle control for the per-node task queue that backs the
// asynchronous files verbs (compress / extract today; the same
// queue also carries cloud uploads and server-side cp/mv).
//
// The async verbs return a task_id; this group is how the user
// acts on a task AFTER it has been queued:
//
//	files task cancel <task-id>      — drop one queued / running task
//	files task cancel --all          — drop EVERY task on the node
//	files task pause  <task-id>      — suspend an in-flight task
//	files task resume <task-id>      — resume a paused task
//
// All four hit the shared `/api/task/<node>/` wire surface (see
// internal/files/archive/task.go for the wire shapes). Tasks are
// per-node: the {node} segment must match the node the task was
// queued on. `compress` / `extract` print that node in their
// queue line ("queued compress task: <id>" + "node=<node>"), and
// these verbs default-resolve the same way `files cp` does
// (--node override → master node from /api/nodes/) when --node is
// omitted.
func NewTaskCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "control the per-node task queue (cancel / pause / resume compress & extract tasks)",
		Long: `Act on tasks queued on the per-user files-backend's per-node task queue.

The asynchronous file verbs — ` + "`files compress`" + ` and ` + "`files extract`" + ` —
return a task_id and run on the server's per-node task queue. This verb
group is how you control a task after it has been queued:

    files task cancel <task-id>      drop one queued / running task
    files task cancel --all          drop EVERY task on the node
    files task pause  <task-id>      suspend an in-flight task
    files task resume <task-id>      resume a paused task

Wire surface (all share /api/task/<node>/):

    DELETE /api/task/<node>/?task_id=<id>          cancel one
    DELETE /api/task/<node>/?all=1                 cancel all
    POST   /api/task/<node>/?task_id=<id>&op=pause
    POST   /api/task/<node>/?task_id=<id>&op=resume

Tasks are per-node — the {node} segment must match the node the task was
queued on. ` + "`files compress`" + ` / ` + "`files extract`" + ` print that node in
their "queued ... task" line. When --node is omitted these verbs resolve
the master node from /api/nodes/, same cascade as ` + "`files cp`" + `.

Examples:

    # Cancel a specific task on the default node.
    olares-cli files task cancel 6f1c2e3a-...

    # Cancel a task on an explicit node.
    olares-cli files task cancel 6f1c2e3a-... --node olares

    # Cancel everything queued on a node (asks for confirmation).
    olares-cli files task cancel --all --node olares

    # Pause / resume a long-running compress.
    olares-cli files task pause 6f1c2e3a-...
    olares-cli files task resume 6f1c2e3a-...
`,
	}
	for _, sub := range []*cobra.Command{
		newTaskCancelCommand(f),
		newTaskPauseCommand(f),
		newTaskResumeCommand(f),
	} {
		sub.SilenceUsage = true
		cmd.AddCommand(sub)
	}
	return cmd
}

// taskCancelOptions holds the flags for `files task cancel`.
type taskCancelOptions struct {
	node  string
	all   bool
	force bool
}

func newTaskCancelCommand(f *cmdutil.Factory) *cobra.Command {
	o := &taskCancelOptions{}
	cmd := &cobra.Command{
		Use:   "cancel [task-id]",
		Short: "cancel one task (DELETE ?task_id=...) or every task on the node (--all)",
		Long: `Cancel a queued / running task on the per-node task queue.

Two modes:

    files task cancel <task-id>      cancel exactly one task
    files task cancel --all          cancel EVERY task on the node

Cancelling a half-built archive is the safe way to abort a compress /
extract that is taking too long or was started by mistake — it stops the
server-side work instead of just detaching the local --wait poll.

--all is destructive (it drops every other user-visible task on the node,
including ones started elsewhere), so it prompts for confirmation unless
-f/--force is passed.

For a single task, the CLI first reads the task's state and refuses up
front when it is already terminal (completed / failed / cancelled) or
when the server reports pause_able=false (not interruptible) — pass
-f/--force to skip that precheck and send the DELETE anyway.

Wire shape:

    DELETE /api/task/<node>/?task_id=<id>     (single)
    DELETE /api/task/<node>/?all=1            (--all)
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := ""
			if len(args) == 1 {
				taskID = strings.TrimSpace(args[0])
			}
			return runTaskCancel(cmd.Context(), f, cmd.OutOrStdout(), taskID, o)
		},
	}
	cmd.Flags().StringVar(&o.node, "node", "",
		"node whose task queue holds the task (defaults to the master node from /api/nodes/)")
	cmd.Flags().BoolVar(&o.all, "all", false,
		"cancel EVERY task on the node instead of a single task-id")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false,
		"skip the --all confirmation prompt AND the single-task pause_able/terminal precheck")
	return cmd
}

func runTaskCancel(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	taskID string,
	o *taskCancelOptions,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// Exactly one of {task-id, --all} must be supplied. Both or
	// neither is a usage error — refuse rather than guess.
	if o.all && taskID != "" {
		return errors.New("task cancel: pass either a <task-id> or --all, not both")
	}
	if !o.all && taskID == "" {
		return errors.New("task cancel: need a <task-id> (or --all to cancel every task on the node)")
	}

	cli, node, err := newTaskClient(ctx, f, o.node)
	if err != nil {
		return err
	}

	if o.all {
		if !o.force {
			// --all is destructive across the whole node queue;
			// mirror `files rm`'s guard — refuse in a non-TTY
			// context, otherwise prompt y/N.
			if !term.IsTerminal(int(syscall.Stdin)) {
				return errors.New("refusing to cancel all tasks without --force in a non-interactive context (no TTY)")
			}
			fmt.Fprintf(out, "Cancel ALL tasks on node %q? This drops every queued / running task, including ones started elsewhere. [y/N]: ", node)
			ok, perr := readYesNo(os.Stdin)
			if perr != nil {
				return perr
			}
			if !ok {
				fmt.Fprintln(out, "aborted; no tasks cancelled")
				return nil
			}
		}
		if err := cli.CancelAllTasks(ctx, node); err != nil {
			return reformatArchiveHTTPErr(err, profileOlaresID(ctx, f), "task cancel --all", node)
		}
		fmt.Fprintf(out, "cancelled all tasks on node %s\n", node)
		return nil
	}

	if err := ensureTaskControllable(ctx, cli, node, taskID, "cancel", o.force); err != nil {
		return err
	}
	if err := cli.CancelTask(ctx, node, taskID); err != nil {
		return reformatArchiveHTTPErr(err, profileOlaresID(ctx, f), "task cancel", taskID)
	}
	fmt.Fprintf(out, "cancelled task %s on node %s\n", taskID, node)
	return nil
}

// ensureTaskControllable preflights a pause / resume / cancel by
// reading the task's current state and refusing client-side when
// the server would reject the op anyway — instead of firing the
// request and surfacing a raw 4xx. Mirrors TermiPass, which hides
// the pause / resume / cancel controls when pause_able is false
// (pauseDisable / cancellable in olaresTask/archive.ts).
//
// Refuses when:
//
//   - the task is already terminal (completed / failed / cancelled)
//     — there is nothing left to act on; or
//   - the server reports pause_able=false — the task's type / phase
//     is not interruptible.
//
// On any lookup error it returns nil (allow the op): a transient
// query failure shouldn't block a legitimate request, and the op
// itself surfaces the real error (a genuinely missing task fails
// there too). force=true skips the precheck entirely.
func ensureTaskControllable(
	ctx context.Context,
	cli *archive.Client,
	node, taskID, op string,
	force bool,
) error {
	if force {
		return nil
	}
	info, err := cli.GetTask(ctx, node, taskID)
	if err != nil {
		return nil
	}
	if info.IsTerminal() {
		return fmt.Errorf(
			"task %s is already %s on node %s; nothing to %s (pass --force to send the request anyway)",
			taskID, info.Status, node, op)
	}
	if !info.PauseAble {
		return fmt.Errorf(
			"task %s on node %s is not controllable: the server reports pause_able=false for its current "+
				"type/phase, so `files task %s` would be rejected. Pass --force to send the request anyway",
			taskID, node, op)
	}
	return nil
}

// taskPauseResumeOptions holds the flags shared by pause / resume.
type taskPauseResumeOptions struct {
	node  string
	force bool
}

func newTaskPauseCommand(f *cmdutil.Factory) *cobra.Command {
	o := &taskPauseResumeOptions{}
	cmd := &cobra.Command{
		Use:   "pause <task-id>",
		Short: "suspend an in-flight task (POST ?task_id=...&op=pause)",
		Long: `Suspend an in-flight compress / extract task. The task keeps its place
in the queue and can be resumed with ` + "`files task resume`" + `.

Only tasks the server reports as pause-able honour this. Before sending
the request the CLI reads the task's state and refuses up front when it
is already terminal (completed / failed / cancelled) or when the server
reports pause_able=false — pass --force to skip that check and send the
request anyway.

Wire shape:

    POST /api/task/<node>/?task_id=<id>&op=pause
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskPauseResume(cmd.Context(), f, cmd.OutOrStdout(), strings.TrimSpace(args[0]), o, "pause")
		},
	}
	cmd.Flags().StringVar(&o.node, "node", "",
		"node whose task queue holds the task (defaults to the master node from /api/nodes/)")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false,
		"skip the pause_able / terminal-status precheck and send the request anyway")
	return cmd
}

func newTaskResumeCommand(f *cmdutil.Factory) *cobra.Command {
	o := &taskPauseResumeOptions{}
	cmd := &cobra.Command{
		Use:   "resume <task-id>",
		Short: "resume a paused task (POST ?task_id=...&op=resume)",
		Long: `Resume a task previously suspended with ` + "`files task pause`" + `.

Before sending the request the CLI reads the task's state and refuses up
front when it is already terminal or when the server reports
pause_able=false — pass --force to skip that check.

Wire shape:

    POST /api/task/<node>/?task_id=<id>&op=resume
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskPauseResume(cmd.Context(), f, cmd.OutOrStdout(), strings.TrimSpace(args[0]), o, "resume")
		},
	}
	cmd.Flags().StringVar(&o.node, "node", "",
		"node whose task queue holds the task (defaults to the master node from /api/nodes/)")
	cmd.Flags().BoolVarP(&o.force, "force", "f", false,
		"skip the pause_able / terminal-status precheck and send the request anyway")
	return cmd
}

func runTaskPauseResume(
	ctx context.Context,
	f *cmdutil.Factory,
	out io.Writer,
	taskID string,
	o *taskPauseResumeOptions,
	op string,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if taskID == "" {
		return fmt.Errorf("task %s: <task-id> must not be empty", op)
	}
	cli, node, err := newTaskClient(ctx, f, o.node)
	if err != nil {
		return err
	}
	if err := ensureTaskControllable(ctx, cli, node, taskID, op, o.force); err != nil {
		return err
	}
	switch op {
	case "pause":
		err = cli.PauseTask(ctx, node, taskID)
	case "resume":
		err = cli.ResumeTask(ctx, node, taskID)
	default:
		return fmt.Errorf("task: unknown op %q", op)
	}
	if err != nil {
		return reformatArchiveHTTPErr(err, profileOlaresID(ctx, f), "task "+op, taskID)
	}
	fmt.Fprintf(out, "%sd task %s on node %s\n", op, taskID, node)
	return nil
}

// newTaskClient builds an archive.Client over the streaming-safe
// HTTP client and resolves the {node} segment via the same cascade
// the archive verbs use (--node override → master node from
// /api/nodes/). Tasks carry no frontend path, so the node resolver
// is fed a nil path slice — it falls through to the /api/nodes/
// default when --node is unset.
func newTaskClient(ctx context.Context, f *cmdutil.Factory, flagNode string) (*archive.Client, string, error) {
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, "", err
	}
	httpClient, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, "", err
	}
	node, err := resolveArchiveNode(ctx, f, rp, nil, flagNode)
	if err != nil {
		return nil, "", err
	}
	return &archive.Client{HTTPClient: httpClient, BaseURL: rp.FilesURL}, node, nil
}

// profileOlaresID best-effort fetches the active profile's Olares
// ID for the error reformatter's "profile login" CTA. Returns ""
// when the profile can't be resolved — the reformatter degrades to
// the generic re-login hint.
func profileOlaresID(ctx context.Context, f *cmdutil.Factory) string {
	rp, err := f.ResolveProfile(ctx)
	if err != nil || rp == nil {
		return ""
	}
	return rp.OlaresID
}
