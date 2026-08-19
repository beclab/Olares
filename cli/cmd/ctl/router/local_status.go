package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/utils"
)

// `router local status` — GET /healthz, /api/progress, /api/build-info
// `router local progress` — GET /api/progress
//
// The two overlap on purpose. Status answers "can this model be called, and if
// not, what is it doing"; progress answers "how far along, how fast, how much
// longer", which is a different question asked repeatedly.

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
// `router local spec`.
type localSpecBrief struct {
	Name string `json:"name"`
	Mode string `json:"mode"`
}

func newLocalStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "status <app>",
		Short: "lifecycle phase, engine, and last verification",
		Long: `Report what a model application is doing.

Four things are read together, because they fail independently and the
combination is what identifies a problem: the composite health document, the
lifecycle progress snapshot, the model card, and the Model Console's own build.

Phases mean different things to a caller:

  init       the process is up, nothing fetched yet
  download   fetching weights; calls are refused meanwhile
  loading    weights are on disk and the engine is starting on them
  ready      serving
  degraded   serving in a reduced state, with a reason in progress
  failed     it stopped trying; "router local retry" restarts the loop

A probe that fails is reported rather than ending the command, so a half-up
application still produces a usable picture.

Examples:
  olares-cli router local status llamacppqwen3627bggufv3
  olares-cli router local status "Qwen3.6-27B (llama.cpp)"
  olares-cli router local status llamacppqwen3627bggufv3 -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runLocalStatus(c.Context(), f, args[0], output)
		},
	}
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
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	rows := [][2]string{
		{"APPLICATION", r.App.AppName},
		{"TITLE", nonEmpty(r.App.Title)},
	}
	if r.App.Provider != "" && r.App.Provider != r.App.Title {
		rows = append(rows, [2]string{"ROUTER PROVIDER", r.App.Provider})
	}
	if r.App.State != "" {
		rows = append(rows, [2]string{"OLARES STATE", r.App.State})
	}
	rows = append(rows, [2]string{"MODEL CONSOLE", r.App.BaseURL})
	if r.Spec != nil {
		rows = append(rows, [2]string{"MODEL", nonEmpty(r.Spec.Name)})
		rows = append(rows, [2]string{"MODE", nonEmpty(r.Spec.Mode)})
	}
	if h := r.Health; h != nil {
		phase := h.Phase
		if note := phaseNote(h.Phase, h); note != "" {
			phase += " — " + note
		}
		rows = append(rows,
			[2]string{"HEALTH", h.Status},
			[2]string{"PHASE", phase},
			[2]string{"READY", boolStr(h.Ready)},
			[2]string{"ENGINE ALIVE", boolStr(h.EngineAlive)},
			[2]string{"MODEL ON ENGINE", boolStr(h.ModelExists)},
			[2]string{"LAST VERIFY", verifyLabel(h.LastVerify, h.LastVerifyOK)},
		)
	}
	if p := r.Progress; p != nil {
		if f := p.fraction(); f >= 0 {
			rows = append(rows, [2]string{"DOWNLOADED", fmt.Sprintf("%.1f%% of %s",
				f*100, utils.FormatBytes(p.BytesTotal))})
		} else if p.BytesCompleted > 0 {
			rows = append(rows, [2]string{"DOWNLOADED", utils.FormatBytes(p.BytesCompleted) + " (total unknown)"})
		}
		if p.RetryCount > 0 || p.TransportRetries > 0 {
			rows = append(rows, [2]string{"RETRIES", fmt.Sprintf("%d lifecycle, %d transport",
				p.RetryCount, p.TransportRetries)})
		}
		if e := strings.TrimSpace(p.LastError); e != "" {
			rows = append(rows, [2]string{"LAST ERROR", e})
		}
	}
	if b := r.Build; b != nil {
		rows = append(rows, [2]string{"CONSOLE VERSION", nonEmpty(b.Version)})
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1]); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	for _, e := range r.Errors {
		if _, err := fmt.Fprintf(w, "\ncould not read %s\n", e); err != nil {
			return err
		}
	}
	if r.Health != nil && !r.Health.Ready {
		_, err := fmt.Fprintln(w, "\nThis model refuses calls until it is ready. "+
			"`olares-cli router local progress "+r.App.AppName+" --watch` follows it.")
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

func newLocalProgressCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		watch    bool
		interval time.Duration
	)
	cmd := &cobra.Command{
		Use:   "progress <app>",
		Short: "the download and load snapshot",
		Long: `Show how far a model application is through getting itself ready.

Bytes, speed and estimate are only populated while something is being fetched;
a model already on disk reports its phase and little else, which is the correct
answer rather than a gap.

--watch redraws until the phase settles — ready, degraded or failed. There is no
event stream to subscribe to here, so this polls, and the interval is yours to
choose.

Examples:
  olares-cli router local progress llamacppqwen3627bggufv3
  olares-cli router local progress llamacppqwen3627bggufv3 --watch
  olares-cli router local progress llamacppqwen3627bggufv3 --watch --interval 5s
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runLocalProgress(c.Context(), f, args[0], watch, interval, output)
		},
	}
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
