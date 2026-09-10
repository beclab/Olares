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

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/utils"
)

// `router model diag …` — the three questions about a local model that are not
// about whether it works.
//
// gpu       GET  /api/diag/gpu
// config    GET  /api/config
// endpoints GET  /api/endpoints
//
// gpu answers "the model replies, but slowly", which Router cannot see: from
// the gateway a model on the CPU and a model on the GPU differ only in latency.
// The figures come from the engine's own introspection rather than from
// nvidia-smi, so what is reported varies by engine and an absent field means
// "this engine does not say" rather than zero.
//
// config and endpoints answer the question underneath that one — what this
// deployment was actually asked to be, and which routes it therefore serves.
//
// These are separated from the lifecycle verbs because they answer a different
// question. `model status` says whether a model is usable; these say why a model
// that is usable behaves the way it does.

// newModelDiagCommand groups them under one noun rather than four more verbs on
// `model`, which already carries the ones a person reaches for daily.
func newModelDiagCommand(f *cmdutil.Factory, how localAddressing) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diag",
		Short: "why a working local model behaves the way it does",
		Long: `Look inside a model application that is running.

  diag gpu <model>        how much of the model is resident on the GPU
  diag config <model>     the effective configuration, secrets redacted
  diag endpoints <model>  which routes this deployment actually serves

Whether a model is usable at all is "router model status". These are for the
one that is: slow, on the wrong device, or missing a route you expected.

Every verb takes a model, and --app to name the application directly instead of
asking Router which one serves it.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newModelGPUCommand(f, how))
	cmd.AddCommand(newModelConfigCommand(f, how))
	cmd.AddCommand(newModelEndpointsCommand(f, how))
	return cmd
}

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

func newModelGPUCommand(f *cmdutil.Factory, how localAddressing) *cobra.Command {
	var (
		output string
		native bool
	)
	target := newModelTarget(how)
	cmd := &cobra.Command{
		Use:   "gpu " + target.arg(),
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
  olares-cli router model diag gpu qwen3-4b
  olares-cli router model diag gpu --app llamacppqwen3627bggufv3 --engine-native -o json
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
			li, err := target.open(ctx, f, args)
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
	target.bind(cmd)
	cmd.Flags().BoolVar(&native, "engine-native", false,
		"include the engine's raw introspection dump (JSON output only)")
	addOutputFlag(cmd, &output)
	return cmd
}

func renderGPUReport(w io.Writer, li *llmInit, r *gpuReport) error {
	t := newTable(w)
	t.row("APPLICATION", li.AppName)
	t.row("ENGINE", nonEmpty(r.EngineKind))
	t.row("MODEL", nonEmpty(r.Model.Name))
	if r.Model.Bytes > 0 {
		size := utils.FormatBytes(r.Model.Bytes)
		if r.Model.Quantize != "" {
			size += " (" + r.Model.Quantize + ")"
		}
		t.row("WEIGHTS", size)
	}
	t.row("PLACEMENT", r.placement())
	if r.GPU.VRAMBytes > 0 {
		t.row("RESIDENT", utils.FormatBytes(r.GPU.VRAMBytes))
	}
	if r.GPU.TotalLayers > 0 {
		t.row("LAYERS ON GPU", fmt.Sprintf("%d of %d", r.GPU.GPULayers, r.GPU.TotalLayers))
	}
	if r.GPU.CPUOffloadGB > 0 {
		t.row("CPU OFFLOAD", fmt.Sprintf("%d GB", r.GPU.CPUOffloadGB))
	}
	if p := r.GPU.GPUMemoryUtilization; p != nil {
		t.row("MEMORY RESERVED", fmt.Sprintf("%.0f%% of the GPU", *p*100))
	}
	if p := r.GPU.KVCacheUsagePerc; p != nil {
		t.row("KV CACHE IN USE", fmt.Sprintf("%.1f%% right now", *p*100))
	}
	t.row("MEASURED BY", nonEmpty(r.GPU.Source))
	if r.GeneratedAt != "" {
		t.row("AT", r.GeneratedAt)
	}
	if err := t.flush(); err != nil {
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
			"`olares-cli router model diag config --app %s`.\n", li.AppName)
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
			"`olares-cli router model diag endpoints <model>` lists what it does serve", err)
	case re.Status == 503:
		return fmt.Errorf("%w\nThe engine did not answer its own introspection probe, which usually means "+
			"it is not up yet; `olares-cli router model status <model>` says which phase it is in", err)
	}
	return err
}

func newModelConfigCommand(f *cmdutil.Factory, how localAddressing) *cobra.Command {
	var output string
	target := newModelTarget(how)
	cmd := &cobra.Command{
		Use:   "config " + target.arg(),
		Short: "the effective configuration, secrets redacted",
		Long: `Show the configuration a model application is actually running with.

Secrets are masked by the Model Console before they leave it — a token becomes a
length, a source URL keeps only its scheme and host — so this is safe to paste
into a bug report.

It is the answer to "why is it downloading from there" and "which engine flags
are in force", both of which are set at install time and easy to misremember.

Examples:
  olares-cli router model diag config qwen3-4b
  olares-cli router model diag config --app llamacppqwen3627bggufv3 -o json
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
			li, err := target.open(ctx, f, args)
			if err != nil {
				return err
			}
			var cfg map[string]any
			if err := li.client.doJSON(ctx, "GET", epLocalConfig, nil, &cfg); err != nil {
				return err
			}
			if format == FormatJSON {
				return printJSON(os.Stdout, cfg)
			}
			return renderFlatMap(os.Stdout, cfg)
		},
	}
	target.bind(cmd)
	addOutputFlag(cmd, &output)
	return cmd
}

// renderFlatMap prints an unknown-shaped document as sorted key/value lines,
// descending into nested objects with dotted keys. The configuration's fields
// are the Model Console's own and will grow; enumerating them here would mean
// silently dropping whatever it learns to report next.
func renderFlatMap(w io.Writer, doc map[string]any) error {
	t := newTable(w)
	writeFlat(t, "", doc)
	return t.flush()
}

func writeFlat(t *table, prefix string, doc map[string]any) {
	keys := make([]string, 0, len(doc))
	for k := range doc {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		name := k
		if prefix != "" {
			name = prefix + "." + k
		}
		switch v := doc[k].(type) {
		case map[string]any:
			if len(v) == 0 {
				continue
			}
			writeFlat(t, name, v)
		case nil:
			continue
		default:
			t.row(name, fmt.Sprintf("%v", v))
		}
	}
}

type localEndpoint struct {
	Method         string   `json:"method"`
	Path           string   `json:"path"`
	Category       string   `json:"category"`
	Group          string   `json:"group,omitempty"`
	Description    string   `json:"description"`
	Available      bool     `json:"available"`
	AsyncSupported *bool    `json:"async_supported,omitempty"`
	Reasons        []string `json:"reasons,omitempty"`
}

func newModelEndpointsCommand(f *cmdutil.Factory, how localAddressing) *cobra.Command {
	var (
		output      string
		unavailable bool
	)
	target := newModelTarget(how)
	cmd := &cobra.Command{
		Use:   "endpoints " + target.arg(),
		Short: "which routes this deployment actually serves",
		Long: `List the routes a model application serves, and which are switched off.

A Model Console mounts different routes depending on what it is running: an
embedding deployment has no chat completions, a translation one moves its routes
to the host root, and the engine-native passthrough exists only for the engines
that have one. So "the endpoint does not exist" and "the endpoint is not mounted
in this deployment" are different answers, and this is where they are told apart
— each unavailable route carries the reason it is absent.

Examples:
  olares-cli router model diag endpoints qwen3-4b
  olares-cli router model diag endpoints --app llamacppqwen3627bggufv3 --unavailable
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
			li, err := target.open(ctx, f, args)
			if err != nil {
				return err
			}
			var env struct {
				EngineKind string          `json:"engine_kind"`
				Endpoints  []localEndpoint `json:"endpoints"`
			}
			if err := li.client.doJSON(ctx, "GET", epLocalEndpoints, nil, &env); err != nil {
				return err
			}
			rows := env.Endpoints
			if unavailable {
				kept := rows[:0:0]
				for _, r := range rows {
					if !r.Available {
						kept = append(kept, r)
					}
				}
				rows = kept
			}
			if format == FormatJSON {
				return printJSON(os.Stdout, map[string]any{
					"engine_kind": env.EngineKind,
					"endpoints":   rows,
				})
			}
			return renderLocalEndpoints(os.Stdout, env.EngineKind, rows, unavailable)
		},
	}
	target.bind(cmd)
	cmd.Flags().BoolVar(&unavailable, "unavailable", false, "only the routes this deployment does not serve")
	addOutputFlag(cmd, &output)
	return cmd
}

func renderLocalEndpoints(w io.Writer, engineKind string, rows []localEndpoint, onlyUnavailable bool) error {
	if len(rows) == 0 {
		if onlyUnavailable {
			_, err := fmt.Fprintln(w, "every route this Model Console knows about is mounted.")
			return err
		}
		_, err := fmt.Fprintln(w, "this Model Console reported no routes at all, which should not happen.")
		return err
	}
	if engineKind != "" {
		if _, err := fmt.Fprintf(w, "engine: %s\n\n", engineKind); err != nil {
			return err
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Category != rows[j].Category {
			return rows[i].Category < rows[j].Category
		}
		return rows[i].Path < rows[j].Path
	})
	hasAsyncMetadata := false
	for _, r := range rows {
		if r.AsyncSupported != nil {
			hasAsyncMetadata = true
			break
		}
	}
	headers := []string{"METHOD", "PATH", "CATEGORY", "SERVED"}
	if hasAsyncMetadata {
		headers = append(headers, "ASYNC")
	}
	headers = append(headers, "WHY NOT")
	t := newTable(w, headers...)
	for _, r := range rows {
		why := ""
		if !r.Available {
			why = clip(strings.Join(r.Reasons, "; "), 60)
			if why == "" {
				why = "no reason given"
			}
		}
		values := []string{r.Method, r.Path, r.Category, boolStr(r.Available)}
		if hasAsyncMetadata {
			async := "-"
			if r.AsyncSupported != nil {
				async = boolStr(*r.AsyncSupported)
			}
			values = append(values, async)
		}
		values = append(values, why)
		t.row(values...)
	}
	return t.flush()
}
