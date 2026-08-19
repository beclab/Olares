package router

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/utils"
)

// `router local gpu`  — GET  /api/diag/gpu
// `router local perf` — POST /api/diag/perf, GET /api/diag/perf/last
//
// Both answer "the model replies, but slowly", which Router cannot see: from
// the gateway a model on the CPU and a model on the GPU differ only in latency.
// The figures come from the engine's own introspection rather than from
// nvidia-smi, so what is reported varies by engine and an absent field means
// "this engine does not say" rather than zero.

type gpuResidency struct {
	VRAMBytes            int64    `json:"vram_bytes,omitempty"`
	ModelBytes           int64    `json:"model_bytes,omitempty"`
	GPULayers            int      `json:"gpu_layers,omitempty"`
	TotalLayers          int      `json:"total_layers,omitempty"`
	CPUOffloadGB         int      `json:"cpu_offload_gb,omitempty"`
	KVCacheUsagePerc     *float64 `json:"kv_cache_usage_perc,omitempty"`
	GPUMemoryUtilization *float64 `json:"gpu_memory_utilization,omitempty"`
	Source               string   `json:"source"`
}

type diagModel struct {
	Name     string `json:"name"`
	Format   string `json:"format,omitempty"`
	Bytes    int64  `json:"bytes,omitempty"`
	Quantize string `json:"quantize,omitempty"`
}

type gpuReport struct {
	GeneratedAt string       `json:"generated_at"`
	EngineKind  string       `json:"engine_kind"`
	Model       diagModel    `json:"model"`
	GPU         gpuResidency `json:"gpu"`
	Warnings    []string     `json:"warnings,omitempty"`
}

// placement is the label the engine deliberately does not compute. Model
// Console dropped its own guess in favour of the byte and layer counts, so each
// client folds them the same way rather than trusting an inference made from
// environment variables that no longer exist.
//
// The last case is the one that matters. Most engines report nothing here, and
// folding silence into "on the CPU" states the opposite of the truth for a
// model that is entirely on the GPU — which is exactly the deployment whose
// operator would come here to check.
func (g gpuReport) placement() string {
	switch {
	case g.GPU.VRAMBytes > 0 && g.GPU.ModelBytes > 0 && g.GPU.VRAMBytes >= g.GPU.ModelBytes:
		return "fully on the GPU"
	case g.GPU.VRAMBytes > 0:
		return "split between GPU and CPU"
	case g.GPU.TotalLayers > 0 && g.GPU.GPULayers >= g.GPU.TotalLayers:
		return "fully on the GPU"
	case g.GPU.TotalLayers > 0 && g.GPU.GPULayers > 0:
		return fmt.Sprintf("split: %d of %d layers on the GPU", g.GPU.GPULayers, g.GPU.TotalLayers)
	case g.GPU.TotalLayers > 0:
		return "on the CPU"
	}
	return "not reported by this engine"
}

func newLocalGPUCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		native bool
	)
	cmd := &cobra.Command{
		Use:   "gpu <app>",
		Short: "how much of the model is resident on the GPU",
		Long: `Show where a model's weights actually are.

The engine is asked about itself, so the fields differ by engine: Ollama reports
resident bytes, vLLM and SGLang report their configured memory fraction and
KV-cache use, llama.cpp reports the layer split. A field that is absent means
this engine does not report it — read it as unknown, not as zero.

Two things read alike and are not: residency is where the weights are, and
KV-cache use is how busy the model is right now. An idle model fully on the GPU
reports 0% cache use.

Examples:
  olares-cli router local gpu llamacppqwen3627bggufv3
  olares-cli router local gpu llamacppqwen3627bggufv3 --engine-native -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			li, err := openLocal(ctx, f, args[0])
			if err != nil {
				return err
			}
			q := url.Values{}
			if native {
				q.Set("include_engine_native", "true")
			}
			path := withQuery(epLocalDiagGPU, q)
			if format == FormatJSON {
				var raw map[string]any
				if err := li.client.doJSON(ctx, "GET", path, nil, &raw); err != nil {
					return diagErr(err)
				}
				return printJSON(os.Stdout, raw)
			}
			var rep gpuReport
			if err := li.client.doJSON(ctx, "GET", path, nil, &rep); err != nil {
				return diagErr(err)
			}
			return renderGPUReport(os.Stdout, li, &rep)
		},
	}
	cmd.Flags().BoolVar(&native, "engine-native", false,
		"include the engine's raw introspection dump (JSON output only)")
	addOutputFlag(cmd, &output)
	return cmd
}

func renderGPUReport(w io.Writer, li *llmInit, r *gpuReport) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	rows := [][2]string{
		{"APPLICATION", li.AppName},
		{"ENGINE", nonEmpty(r.EngineKind)},
		{"MODEL", nonEmpty(r.Model.Name)},
	}
	if r.Model.Bytes > 0 {
		size := utils.FormatBytes(r.Model.Bytes)
		if r.Model.Quantize != "" {
			size += " (" + r.Model.Quantize + ")"
		}
		rows = append(rows, [2]string{"WEIGHTS", size})
	}
	rows = append(rows, [2]string{"PLACEMENT", r.placement()})
	if r.GPU.VRAMBytes > 0 {
		rows = append(rows, [2]string{"RESIDENT", utils.FormatBytes(r.GPU.VRAMBytes)})
	}
	if r.GPU.TotalLayers > 0 {
		rows = append(rows, [2]string{"LAYERS ON GPU", fmt.Sprintf("%d of %d", r.GPU.GPULayers, r.GPU.TotalLayers)})
	}
	if r.GPU.CPUOffloadGB > 0 {
		rows = append(rows, [2]string{"CPU OFFLOAD", fmt.Sprintf("%d GB", r.GPU.CPUOffloadGB)})
	}
	if p := r.GPU.GPUMemoryUtilization; p != nil {
		rows = append(rows, [2]string{"MEMORY RESERVED", fmt.Sprintf("%.0f%% of the GPU", *p*100)})
	}
	if p := r.GPU.KVCacheUsagePerc; p != nil {
		rows = append(rows, [2]string{"KV CACHE IN USE", fmt.Sprintf("%.1f%% right now", *p*100)})
	}
	rows = append(rows, [2]string{"MEASURED BY", nonEmpty(r.GPU.Source)})
	if r.GeneratedAt != "" {
		rows = append(rows, [2]string{"AT", r.GeneratedAt})
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1]); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	for _, warn := range r.Warnings {
		if _, err := fmt.Fprintln(w, "\nwarning: "+warn); err != nil {
			return err
		}
	}
	if r.placement() == "not reported by this engine" {
		_, err := fmt.Fprintf(w, "\nThis engine reports no residency figures, so where the weights sit "+
			"cannot be read from here. What it was asked to do is in its launch flags: "+
			"`olares-cli router local config %s`.\n", li.AppName)
		return err
	}
	return nil
}

type ratePoint struct {
	TokensPerSec float64 `json:"tokens_per_sec"`
	SampleTokens int     `json:"sample_tokens"`
	Source       string  `json:"source"`
}

type perfReport struct {
	StartedAt  string    `json:"started_at"`
	DurationMS int64     `json:"duration_ms"`
	EngineKind string    `json:"engine_kind"`
	Model      diagModel `json:"model"`
	ColdStart  struct {
		EngineColdStartMS  int64  `json:"engine_cold_start_ms"`
		ObservedColdLoadMS int64  `json:"observed_cold_load_ms,omitempty"`
		Source             string `json:"source"`
	} `json:"cold_start"`
	Prefill *ratePoint `json:"prefill,omitempty"`
	Decode  *ratePoint `json:"decode,omitempty"`
	TTFT    struct {
		NoThinkMS   int64  `json:"no_think_ms,omitempty"`
		WithThinkMS int64  `json:"with_think_ms,omitempty"`
		ThinkMode   string `json:"think_mode"`
	} `json:"ttft"`
	GPUAtRunTime gpuResidency `json:"gpu_at_run_time"`
	Warnings     []string     `json:"warnings,omitempty"`
}

type perfRunRequest struct {
	DecodeTokens int    `json:"decode_tokens,omitempty"`
	PrefillPromt string `json:"prefill_prompt,omitempty"`
	WithThink    bool   `json:"with_think,omitempty"`
	WithoutThink *bool  `json:"without_think,omitempty"`
}

func newLocalPerfCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output       string
		last         bool
		decodeTokens int
		prompt       string
		withThink    bool
		onlyThink    bool
	)
	cmd := &cobra.Command{
		Use:   "perf <app>",
		Short: "time to first token and throughput, measured",
		Long: `Measure a model by using it.

This sends one or two real completions at the engine and reports time to first
token and tokens per second, with a GPU snapshot taken at the start so a slow
result can be read against where the weights were.

It costs something. The run occupies a slot in the engine for five to sixty
seconds, only one runs at a time across the application, and the model has to be
ready — an engine still loading refuses rather than recording a meaningless
number. The first run after a load pays the cold-load cost and reads high; the
second is the steady state.

--last reads the previous run instead of starting one, which is free. That cache
is in memory, so a restarted application has none.

Examples:
  olares-cli router local perf llamacppqwen3627bggufv3
  olares-cli router local perf llamacppqwen3627bggufv3 --last
  olares-cli router local perf llamacppqwen3627bggufv3 --decode-tokens 128 --with-think
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			if last && (decodeTokens > 0 || prompt != "" || withThink || onlyThink) {
				return fmt.Errorf("--last reads the run that already happened, so the knobs that shape a " +
					"new one cannot apply to it; drop --last to measure again")
			}
			li, err := openLocal(ctx, f, args[0])
			if err != nil {
				return err
			}

			var rep perfReport
			if last {
				if err := li.client.doJSON(ctx, "GET", epLocalDiagPerfLast, nil, &rep); err != nil {
					return perfLastErr(err)
				}
			} else {
				req := perfRunRequest{DecodeTokens: decodeTokens, PrefillPromt: prompt, WithThink: withThink}
				if onlyThink {
					no := false
					req.WithoutThink = &no
					req.WithThink = true
				}
				if err := li.client.doJSON(ctx, "POST", epLocalDiagPerf, req, &rep); err != nil {
					return perfRunErr(err)
				}
			}
			if format == FormatJSON {
				return printJSON(os.Stdout, rep)
			}
			return renderPerfReport(os.Stdout, li, &rep, last)
		},
	}
	cmd.Flags().BoolVar(&last, "last", false, "read the previous run instead of starting one")
	cmd.Flags().IntVar(&decodeTokens, "decode-tokens", 0, "tokens to sample while measuring decode (default 64, max 256)")
	cmd.Flags().StringVar(&prompt, "prompt", "", "use this prompt for the prefill pass instead of the built-in one")
	cmd.Flags().BoolVar(&withThink, "with-think", false, "add a pass with reasoning enabled")
	cmd.Flags().BoolVar(&onlyThink, "only-think", false, "measure only the reasoning pass")
	addOutputFlag(cmd, &output)
	return cmd
}

func renderPerfReport(w io.Writer, li *llmInit, r *perfReport, cached bool) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	rows := [][2]string{
		{"APPLICATION", li.AppName},
		{"ENGINE", nonEmpty(r.EngineKind)},
		{"MODEL", nonEmpty(r.Model.Name)},
		{"RAN AT", nonEmpty(r.StartedAt)},
		{"TOOK", fmt.Sprintf("%.1fs", float64(r.DurationMS)/1000)},
	}
	if r.TTFT.NoThinkMS > 0 {
		rows = append(rows, [2]string{"FIRST TOKEN", fmt.Sprintf("%dms", r.TTFT.NoThinkMS)})
	}
	switch r.TTFT.ThinkMode {
	case "enabled":
		rows = append(rows, [2]string{"FIRST TOKEN, THINKING", fmt.Sprintf("%dms", r.TTFT.WithThinkMS)})
	case "not_supported_by_model":
		rows = append(rows, [2]string{"THINKING", "the model card does not claim reasoning, so it was not measured"})
	}
	if p := r.Prefill; p != nil {
		rows = append(rows, [2]string{"PREFILL", fmt.Sprintf("%.1f tokens/s over %d tokens (%s)",
			p.TokensPerSec, p.SampleTokens, p.Source)})
	}
	if d := r.Decode; d != nil {
		rows = append(rows, [2]string{"DECODE", fmt.Sprintf("%.1f tokens/s over %d tokens (%s)",
			d.TokensPerSec, d.SampleTokens, d.Source)})
	}
	if r.ColdStart.EngineColdStartMS > 0 {
		rows = append(rows, [2]string{"COLD START", fmt.Sprintf("%.1fs to first ready (%s)",
			float64(r.ColdStart.EngineColdStartMS)/1000, nonEmpty(r.ColdStart.Source))})
	}
	gpu := gpuReport{GPU: r.GPUAtRunTime}
	rows = append(rows, [2]string{"PLACEMENT AT RUN", gpu.placement()})
	for _, row := range rows {
		if _, err := fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1]); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	for _, warn := range r.Warnings {
		if _, err := fmt.Fprintln(w, "\nwarning: "+warn); err != nil {
			return err
		}
	}
	if cached {
		_, err := fmt.Fprintln(w, "\nThis is the stored result of an earlier run, not a fresh measurement.")
		return err
	}
	return nil
}

func diagErr(err error) error {
	var re *RouterError
	if err == nil || !errors.As(err, &re) {
		return err
	}
	switch {
	case re.Status == 404:
		return fmt.Errorf("%w\nThis application does not serve the diagnostic routes; "+
			"`olares-cli router local endpoints <app>` lists what it does serve", err)
	case re.Status == 503:
		return fmt.Errorf("%w\nThe engine did not answer its own introspection probe, which usually means "+
			"it is not up yet; `olares-cli router local status <app>` says which phase it is in", err)
	}
	return err
}

func perfRunErr(err error) error {
	var re *RouterError
	if err == nil || !errors.As(err, &re) {
		return diagErr(err)
	}
	switch re.Code {
	case "engine_not_ready":
		return fmt.Errorf("%w\nMeasuring an engine that is still starting would record timings for work it "+
			"is not doing yet; `olares-cli router local progress <app> --watch` follows it to ready", err)
	case "perf_run_in_progress":
		return fmt.Errorf("%w\nOne run at a time per application. Wait for the other to finish, or read it "+
			"with --last once it has", err)
	}
	return diagErr(err)
}

func perfLastErr(err error) error {
	var re *RouterError
	if err != nil && errors.As(err, &re) && (re.Code == "no_perf_run_yet" || re.Status == 404) {
		return fmt.Errorf("%w\nNothing has been measured in this process yet, and the cache does not survive "+
			"a restart. Drop --last to measure now", err)
	}
	return diagErr(err)
}

func newLocalRetryCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		force  bool
		level  string
	)
	cmd := &cobra.Command{
		Use:   "retry <app>",
		Short: "re-enter the download and load loop now",
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

Calls are rate limited by the application, so a loop of these earns a refusal
rather than faster progress.

Examples:
  olares-cli router local retry llamacppqwen3627bggufv3
  olares-cli router local retry llamacppqwen3627bggufv3 --level sha256
  olares-cli router local retry llamacppqwen3627bggufv3 --force
`,
		Args: cobra.ExactArgs(1),
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
			li, err := openLocal(ctx, f, args[0])
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
			_, err = fmt.Printf("%s: %s, from phase %s.\n`olares-cli router local progress %s --watch` follows it.\n",
				li.AppName, what, nonEmpty(accepted.PreviousPhase), li.AppName)
			return err
		},
	}
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

func newLocalRestartCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		yes    bool
	)
	cmd := &cobra.Command{
		Use:   "restart <app>",
		Short: "relaunch the engine with the current model card",
		Long: `Relaunch a model application's inference engine.

The weights stay where they are; only the engine process is replaced, using the
engine arguments in the card as it stands now. That is what this is for: after
"router local spec set", the changes an engine reads at launch — its mode, and
capabilities its own behaviour depends on — need one.

In-flight requests to that engine end. The model is then unavailable for as long
as it takes to load the weights again, which for a large model on a cold cache
is minutes rather than seconds, and Router reports it as unreachable meanwhile.

Examples:
  olares-cli router local restart llamacppqwen3627bggufv3
  olares-cli router local restart llamacppqwen3627bggufv3 --yes
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := c.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			format, err := parseFormat(output)
			if err != nil {
				return err
			}
			li, err := openLocal(ctx, f, args[0])
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
				return restartErr(ctx, li, err)
			}
			if format == FormatJSON {
				return printJSON(os.Stdout, res)
			}
			_, err = fmt.Printf("%s: the engine was told to relaunch. It is unavailable until the weights "+
				"are loaded again — `olares-cli router local status %s` says when.\n", li.AppName, li.AppName)
			return err
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask for confirmation")
	addOutputFlag(cmd, &output)
	return cmd
}

func restartErr(ctx context.Context, li *llmInit, err error) error {
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
