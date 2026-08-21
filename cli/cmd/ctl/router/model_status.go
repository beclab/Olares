package router

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/utils"
)

// The lifecycle of a model that runs on this machine: what phase it is in, how
// far through it is, and the two ways to push it along.
//
// `router model status`   — GET  /healthz, /api/progress, /api/build-info
// `router model progress` — GET  /api/progress
// `router model retry`    — POST /api/retry
// `router model restart`  — POST /console/api/engine/restart, or the
//                           application's own /api/engine/restart with --app
//
// Status and progress overlap on purpose. Status answers "can this model be
// called, and if not, what is it doing"; progress answers "how far along, how
// fast, how much longer", which is a different question asked repeatedly.

type localStatusReport struct {
	App      *llmInit        `json:"app"`
	Health   *localHealth    `json:"health,omitempty"`
	Progress *localProgress  `json:"progress,omitempty"`
	Build    *localBuildInfo `json:"build,omitempty"`
	Spec     *localSpecBrief `json:"spec,omitempty"`
	Errors   []string        `json:"errors,omitempty"`
}

// localSpecBrief is the part of the model card that belongs in a status
// summary: what the model is called and what kind of work it takes. The rest is
// `router model spec`.
type localSpecBrief struct {
	Name string `json:"name"`
	Mode string `json:"mode"`
}

func newModelStatusCommand(f *cmdutil.Factory, how localAddressing) *cobra.Command {
	var output string
	target := newModelTarget(how)
	cmd := &cobra.Command{
		Use:   "status " + target.arg(),
		Short: "what a local model is doing, and whether it can be called",
		Long: `Report the lifecycle of a model that runs on this Olares.

Router says whether a model answers. This says why it does not yet, by reading
the application serving it: the composite health document, the lifecycle
progress snapshot, the model card and the Model Console's own build, together,
because they fail independently and the combination is what identifies a
problem.

Phases mean different things to a caller:

  init       the process is up, nothing fetched yet
  download   fetching weights; calls are refused meanwhile
  loading    weights are on disk and the engine is starting on them
  ready      serving
  degraded   serving in a reduced state, with a reason in progress
  failed     it stopped trying; "router model retry" restarts the loop

A probe that fails is reported rather than ending the command, so a half-up
application still produces a usable picture.

Naming a model asks Router which application serves it. --app names the
application instead and skips Router entirely, which is the point of it: a
model that is not answering is exactly when Router may be unreachable too, and
this reads the application either way. Only a model running here has any of
this; a cloud provider's models have no lifecycle to report.

Examples:
  olares-cli router model status Olares/qwen3-4b
  olares-cli router model status qwen3-4b -o json
  olares-cli router model status --app llamacppqwen3627bggufv3
`,
		RunE: func(c *cobra.Command, args []string) error {
			ref, err := target.appRef(c.Context(), f, args)
			if err != nil {
				return err
			}
			return runLocalStatus(c.Context(), f, ref, output)
		},
	}
	target.bind(cmd)
	addOutputFlag(cmd, &output)
	return cmd
}

func runLocalStatus(ctx context.Context, f *cmdutil.Factory, ref, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	li, err := openLocal(ctx, f, ref)
	if err != nil {
		return err
	}

	report := &localStatusReport{App: li}
	report.Build = li.Console
	var (
		health localHealth
		prog   localProgress
		brief  localSpecBrief
	)
	probes := []struct {
		what string
		path string
		into any
		keep func()
	}{
		{"health", epHealth, &health, func() { report.Health = &health }},
		{"progress", epLocalProgress, &prog, func() { report.Progress = &prog }},
		{"the model card", epLocalModelSpec, &brief, func() { report.Spec = &brief }},
	}
	// One failure is a fact about that document; four identical ones are a fact
	// about the connection, and printing the same sentence four times buries
	// the part that differs.
	failures := map[string][]string{}
	for _, p := range probes {
		if err := li.client.doJSON(ctx, "GET", p.path, nil, p.into); err != nil {
			failures[err.Error()] = append(failures[err.Error()], p.what)
			continue
		}
		p.keep()
	}
	for msg, what := range failures {
		report.Errors = append(report.Errors, strings.Join(what, ", ")+": "+msg)
	}
	sort.Strings(report.Errors)

	if format == FormatJSON {
		return printJSON(os.Stdout, report)
	}
	return renderLocalStatus(os.Stdout, report)
}

func renderLocalStatus(w io.Writer, r *localStatusReport) error {
	t := newTable(w)
	t.row("APPLICATION", r.App.AppName)
	t.row("TITLE", nonEmpty(r.App.Title))
	if r.App.Provider != "" && r.App.Provider != r.App.Title {
		t.row("ROUTER PROVIDER", r.App.Provider)
	}
	if r.App.State != "" {
		t.row("OLARES STATE", r.App.State)
	}
	t.row("MODEL CONSOLE", r.App.BaseURL)
	if r.Spec != nil {
		t.row("MODEL", nonEmpty(r.Spec.Name))
		t.row("MODE", nonEmpty(r.Spec.Mode))
	}
	if h := r.Health; h != nil {
		phase := h.Phase
		if note := phaseNote(h.Phase, h); note != "" {
			phase += " — " + note
		}
		t.row("HEALTH", h.Status)
		t.row("PHASE", phase)
		t.row("READY", boolStr(h.Ready))
		t.row("ENGINE ALIVE", boolStr(h.EngineAlive))
		t.row("MODEL ON ENGINE", boolStr(h.ModelExists))
		t.row("LAST VERIFY", verifyLabel(h.LastVerify, h.LastVerifyOK))
	}
	if p := r.Progress; p != nil {
		if f := p.fraction(); f >= 0 {
			t.row("DOWNLOADED", fmt.Sprintf("%.1f%% of %s", f*100, utils.FormatBytes(p.BytesTotal)))
		} else if p.BytesCompleted > 0 {
			t.row("DOWNLOADED", utils.FormatBytes(p.BytesCompleted)+" (total unknown)")
		}
		if p.RetryCount > 0 || p.TransportRetries > 0 {
			t.row("RETRIES", fmt.Sprintf("%d lifecycle, %d transport", p.RetryCount, p.TransportRetries))
		}
		if e := strings.TrimSpace(p.LastError); e != "" {
			t.row("LAST ERROR", e)
		}
	}
	if b := r.Build; b != nil {
		t.row("CONSOLE VERSION", nonEmpty(b.Version))
	}
	if err := t.flush(); err != nil {
		return err
	}

	for _, e := range r.Errors {
		if _, err := fmt.Fprintf(w, "\ncould not read %s\n", e); err != nil {
			return err
		}
	}
	if r.Health != nil && !r.Health.Ready {
		_, err := fmt.Fprintln(w, "\nThis model refuses calls until it is ready. "+
			"`olares-cli router model progress --app "+r.App.AppName+" --watch` follows it.")
		return err
	}
	return nil
}

// verifyLabel folds the pair of nullable verification fields into one cell.
// They are only meaningful together: a timestamp with no verdict says a check
// ran and nothing about how it went.
func verifyLabel(at *string, ok *bool) string {
	if at == nil || strings.TrimSpace(*at) == "" {
		return "never"
	}
	verdict := "unknown"
	if ok != nil {
		verdict = "failed"
		if *ok {
			verdict = "passed"
		}
	}
	return verdict + " at " + *at
}

func newModelProgressCommand(f *cmdutil.Factory, how localAddressing) *cobra.Command {
	var (
		output   string
		watch    bool
		interval time.Duration
	)
	target := newModelTarget(how)
	cmd := &cobra.Command{
		Use:   "progress " + target.arg(),
		Short: "how far a local model is through getting itself ready",
		Long: `Show a model application's download and load snapshot.

Bytes, speed and estimate are only populated while something is being fetched;
a model already on disk reports its phase and little else, which is the correct
answer rather than a gap.

--watch redraws until the phase settles — ready, degraded or failed. There is no
event stream to subscribe to here, so this polls, and the interval is yours to
choose.

--app names the application directly instead of asking Router which one serves
the model.

Examples:
  olares-cli router model progress qwen3-4b
  olares-cli router model progress qwen3-4b --watch
  olares-cli router model progress --app llamacppqwen3627bggufv3 --watch --interval 5s
`,
		RunE: func(c *cobra.Command, args []string) error {
			ref, err := target.appRef(c.Context(), f, args)
			if err != nil {
				return err
			}
			return runLocalProgress(c.Context(), f, ref, watch, interval, output)
		},
	}
	target.bind(cmd)
	cmd.Flags().BoolVar(&watch, "watch", false, "keep polling until the phase settles")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "how often to poll with --watch")
	addOutputFlag(cmd, &output)
	return cmd
}

func runLocalProgress(ctx context.Context, f *cmdutil.Factory, ref string, watch bool, interval time.Duration, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if interval < time.Second {
		interval = time.Second
	}
	li, err := openLocal(ctx, f, ref)
	if err != nil {
		return err
	}

	read := func() (*localProgress, error) {
		var p localProgress
		if err := li.client.doJSON(ctx, "GET", epLocalProgress, nil, &p); err != nil {
			return nil, err
		}
		return &p, nil
	}

	p, err := read()
	if err != nil {
		return err
	}
	if format == FormatJSON && !watch {
		return printJSON(os.Stdout, p)
	}
	if err := renderLocalProgress(os.Stdout, li, p, format); err != nil {
		return err
	}
	if !watch || settledPhase(p.Phase) {
		return nil
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			next, err := read()
			if err != nil {
				return err
			}
			if err := renderLocalProgress(os.Stdout, li, next, format); err != nil {
				return err
			}
			if settledPhase(next.Phase) {
				return nil
			}
		}
	}
}

// settledPhase is where the lifecycle stops moving on its own. `degraded` and
// `failed` count: waiting past them would wait forever, and the reason to stop
// watching is that a person now has to decide something.
func settledPhase(phase string) bool {
	switch strings.ToLower(strings.TrimSpace(phase)) {
	case "ready", "degraded", "failed":
		return true
	}
	return false
}

func renderLocalProgress(w io.Writer, li *llmInit, p *localProgress, format Format) error {
	if format == FormatJSON {
		return printJSON(w, p)
	}
	line := li.AppName + "  " + p.Phase
	if f := p.fraction(); f >= 0 {
		line += fmt.Sprintf("  %.1f%%  %s/%s", f*100,
			utils.FormatBytes(p.BytesCompleted), utils.FormatBytes(p.BytesTotal))
	} else if p.BytesCompleted > 0 {
		line += "  " + utils.FormatBytes(p.BytesCompleted)
	}
	// Speed and estimate are only about a transfer in flight. The snapshot
	// keeps the last values it saw, so a finished download still reports the
	// rate it managed at the end — true, and read as current.
	if !settledPhase(p.Phase) {
		if p.SpeedBytesPerSec > 0 {
			line += fmt.Sprintf("  %s/s", utils.FormatBytes(int64(p.SpeedBytesPerSec)))
		}
		if p.ETASeconds > 0 {
			line += "  eta " + fmtDuration(p.ETASeconds)
		}
	}
	if note := phaseNote(p.Phase, nil); note != "" && p.fraction() < 0 {
		line += "  (" + note + ")"
	}
	if _, err := fmt.Fprintln(w, line); err != nil {
		return err
	}
	if e := strings.TrimSpace(p.LastError); e != "" {
		if _, err := fmt.Fprintln(w, "  last error: "+e); err != nil {
			return err
		}
	}
	return nil
}

func newModelRetryCommand(f *cmdutil.Factory, how localAddressing) *cobra.Command {
	var (
		output string
		force  bool
		level  string
	)
	target := newModelTarget(how)
	cmd := &cobra.Command{
		Use:   "retry " + target.arg(),
		Short: "re-enter a local model's download and load loop now",
		Long: `Tell a model application to try again immediately.

An application that failed, or is waiting between attempts, sleeps before its
next one. This wakes it. It is the right response to a download that failed for a
reason since fixed — a network outage, a token that has been replaced, a disk
that had no room.

--force re-downloads everything, ignoring the manifest's account of what is
already good. That is a full re-pull, potentially many gigabytes, and it is the
only thing that gets past a manifest that wrongly believes the files are intact.

--level chooses how thoroughly this one attempt verifies: size is quick, sha256
reads every byte.

--app names the application directly instead of asking Router which one serves
the model.

Calls are rate limited by the application, so a loop of these earns a refusal
rather than faster progress.

Examples:
  olares-cli router model retry qwen3-4b
  olares-cli router model retry qwen3-4b --level sha256
  olares-cli router model retry --app llamacppqwen3627bggufv3 --force
`,
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			switch level {
			case "", "size", "sha256":
			default:
				return fmt.Errorf("--level is size or sha256, not %q", level)
			}
			li, err := target.open(ctx, f, args)
			if err != nil {
				return err
			}
			q := url.Values{}
			if force {
				q.Set("force", "true")
			}
			if level != "" {
				q.Set("level", level)
			}
			path := withQuery(epLocalRetry, q)
			var accepted struct {
				Status        string `json:"status"`
				PreviousPhase string `json:"previous_phase"`
			}
			if err := li.client.doJSON(ctx, "POST", path, nil, &accepted); err != nil {
				return retryErr(err)
			}
			if format == FormatJSON {
				return printJSON(os.Stdout, accepted)
			}
			what := "re-entering the loop"
			if force {
				what = "re-downloading every file"
			}
			_, err = fmt.Printf("%s: %s, from phase %s.\n"+
				"`olares-cli router model progress --app %s --watch` follows it.\n",
				li.AppName, what, nonEmpty(accepted.PreviousPhase), li.AppName)
			return err
		},
	}
	target.bind(cmd)
	cmd.Flags().BoolVar(&force, "force", false, "re-download every file, ignoring the manifest")
	cmd.Flags().StringVar(&level, "level", "", "verification for this attempt: size or sha256")
	addOutputFlag(cmd, &output)
	return cmd
}

func retryErr(err error) error {
	var re *RouterError
	if err == nil || !errors.As(err, &re) {
		return err
	}
	switch {
	case re.Status == 429:
		return fmt.Errorf("%w\nThe application limits how often this can be asked. "+
			"Waiting a second or two is enough; the retry it is already doing is not affected", err)
	case re.Status == 503:
		return fmt.Errorf("%w\nThis application has no lifecycle to retry — it is not managing a download "+
			"at all", err)
	}
	return err
}

// newModelRestartCommand relaunches the inference process, by either of the two
// roads to it.
//
// Router's own route is the default, because it is the one that works from
// anywhere and takes the name a caller already knows the model by. --app goes
// straight at the application's Model Console instead, which is what is left
// when Router cannot resolve the model or cannot be reached at all — the engine
// being wedged and the gateway being confused about it are not independent
// events.
//
// Neither reads or writes the card. Both mean the same thing to a user of the
// model: it stops answering until the weights have loaded again.
func newModelRestartCommand(f *cmdutil.Factory, how localAddressing) *cobra.Command {
	var (
		output string
		yes    bool
	)
	target := newModelTarget(how)
	cmd := &cobra.Command{
		Use:   "restart " + target.arg(),
		Short: "relaunch a local model's engine on the card it already has",
		Long: `Relaunch the inference process without changing anything.

This is for an engine that has stopped behaving rather than one that is
configured wrongly — wedged, leaking, answering slowly after a long session. The
card is not read and not written; the process is told to come back on the
document already on disk.

A success means the relaunch was signalled. The engine then shuts down and
loads the model again, and the model does not answer while it does, which for a
large one is minutes. In-flight requests end. An application whose engine is a
sidecar — OCR, audio, embedding — has no process to signal, and the call
succeeds having changed nothing.

Naming the model goes through Router, which is the road to prefer. --app goes
straight at the application, for when Router cannot resolve the model or cannot
be reached.

Examples:
  olares-cli router model restart Olares/qwen3-4b
  olares-cli router model restart Olares/qwen3-4b -y
  olares-cli router model restart --app llamacppqwen3627bggufv3
`,
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			if target.direct(args) {
				return runConsoleRestart(ctx, f, target, args, yes, format)
			}
			return runRouterRestart(ctx, f, target.model(args), yes, format)
		},
	}
	target.bind(cmd)
	addConfirmFlag(cmd, &yes)
	addOutputFlag(cmd, &output)
	return cmd
}

func runRouterRestart(ctx context.Context, f *cmdutil.Factory, model string, yes bool, format Format) error {
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin, fmt.Sprintf(
		"Relaunch the engine serving %s? It stops answering until the model has loaded again.",
		model), yes); err != nil {
		return err
	}
	var res struct {
		Restarted bool `json:"restarted"`
	}
	if err := pc.router.doJSON(ctx, "POST", epForModel(epEngineRestart, model), nil, &res); err != nil {
		return specErr(err, model)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, res)
	}
	if !res.Restarted {
		_, werr := fmt.Println("the application accepted the request and reports no relaunch, " +
			"which is what an application with a sidecar engine does.")
		return werr
	}
	_, werr := fmt.Printf("relaunching. The model answers again once it has loaded; "+
		"`olares-cli router model status %s` reports how far along that is.\n", model)
	return werr
}

func runConsoleRestart(ctx context.Context, f *cmdutil.Factory, target *modelTarget, args []string,
	yes bool, format Format) error {
	li, err := target.open(ctx, f, args)
	if err != nil {
		return err
	}
	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin, fmt.Sprintf(
		"Relaunch the engine in %s? In-flight requests end, and the model is unavailable "+
			"until the weights load again.", li.AppName), yes); err != nil {
		return err
	}
	var res struct {
		OK     bool   `json:"ok"`
		RunDir string `json:"run_dir"`
	}
	if err := li.client.doJSON(ctx, "POST", epLocalEngineRestart, nil, &res); err != nil {
		return consoleRestartErr(ctx, li, err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, res)
	}
	_, err = fmt.Printf("%s: the engine was told to relaunch. It is unavailable until the weights "+
		"are loaded again — `olares-cli router model status --app %s` says when.\n", li.AppName, li.AppName)
	return err
}

func consoleRestartErr(ctx context.Context, li *llmInit, err error) error {
	var re *RouterError
	if err == nil || !errors.As(err, &re) {
		return err
	}
	if re.Status == 404 {
		if served := consoleServes(ctx, li, "POST", epLocalEngineRestart); served != nil && !*served {
			return fmt.Errorf("this application's Model Console (%s) cannot relaunch its engine on "+
				"request — the route arrived in a later version. Restarting the application does the "+
				"same thing from outside: `olares-cli market restart %s`",
				consoleVersion(li), li.AppName)
		}
	}
	switch re.Code {
	case "no_engine":
		return fmt.Errorf("%w\nThis application runs no engine of its own, so there is nothing to relaunch", err)
	case "run_dir_unset":
		return fmt.Errorf("%w\nThe Model Console cannot reach the supervisor that would do the relaunch. "+
			"Restarting the application itself is the way round it: `olares-cli market restart <app>`", err)
	}
	return err
}
