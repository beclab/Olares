package router

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router call models` — every name the `model` field accepts.
//
// GET /v1/models
//
// This is a verb on `call` rather than on `model` because it is answered by the
// data plane, over the same key every other `call` verb uses. `router model
// list` is the management plane's view: one row per configured model, with its
// provider, its mode and whether the provider is healthy — everything an admin
// needs to decide what to change. This is the *caller's* view, and it answers a
// narrower question: what may I put in the `model` field of the call I am about
// to make. Three differences follow from that, and all three are the point.
//
// Only models are here. A route — an alias, a group, or a default category like
// `default-chat` — is callable and is not listed: it has no provider to qualify
// it with and no single backend to describe, so every column below would either
// be empty for it or would pin the caller to the very thing the name exists to
// stop it caring about. `router route list` is where the names live.
//
// Everything listed is sendable right now. A locally installed model
// application owns its `router model list` row from the moment it is installed,
// but it joins this list only once its container is up AND its weights are
// loaded, which are minutes apart. So a name in `router model list` and not here
// is usually that, and `--include-not-ready` is what says so: it widens the read
// to the container alone, and the model appears as `warming` while it downloads
// or as `failed` if it could not load, instead of being indistinguishable from
// one that was never configured.
//
// And what appears depends on the credential. A key with an allowed-models list
// sees only what it may call, so this is the honest answer to "will my client
// work with this key" — which `router model list`, read over the console
// session, cannot give.

// modelObject is one entry of the OpenAI list envelope, plus the three fields
// Router adds to it: the endpoint family the model serves, the capabilities its
// card claims, and whether the weights can answer.
type modelObject struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
	// QualifiedID repeats ID. It predates `id` itself carrying the qualified
	// "<provider>/<model>" reference, and Router still sends both so that
	// clients written against the older shape keep working. Read ID.
	QualifiedID string `json:"qualified_id"`
	// Mode is the endpoint family: chat, embedding, tts, ocr, image and the
	// rest. The list mixes them, and the qualified ids are nothing like the
	// "tts-1" a client pattern-matches against, so this is the only reliable
	// way to tell which endpoint a name belongs to.
	Mode string `json:"mode"`
	// Supports names the capabilities the model card claims, without the
	// `supports_` prefix the card itself uses.
	Supports []string `json:"supports"`
	// Readiness is ready, warming, failed or unknown. Without
	// --include-not-ready only ready and unknown can appear, and both mean
	// the model is sendable: unknown is an honest "nothing here can tell",
	// which is what an application running its own engine — and so reporting
	// no phase for Router to read — looks like. A remote vendor has no
	// weights to wait for and reads ready.
	Readiness string `json:"readiness"`
	// The card's own figures, carried so that -o json here says as much as
	// Router did. ContextSize is the per-request window, MaxOutputTokens the
	// longest reply, and MaxConcurrency how many requests the engine behind
	// the model works on at once — the last one only for a local engine whose
	// launch flags said, since a cloud vendor never tells us.
	ContextSize     int `json:"context_size,omitempty"`
	MaxOutputTokens int `json:"max_output_tokens,omitempty"`
	MaxConcurrency  int `json:"max_concurrency,omitempty"`
	// KVPoolTokens is the whole KV cache the engine serves with, and it is
	// not ContextSize × MaxConcurrency: llama.cpp in unified mode lets every
	// slot promise the full window out of one pool that cannot cover them
	// all. Whoever reaches the model second is then the one whose prompt is
	// refused, so a caller sizing a request needs the pool as well as the
	// pair.
	KVPoolTokens int `json:"kv_pool_tokens,omitempty"`

	// The rest arrive only with --operations, and describe the routes the
	// model actually serves rather than the capabilities it claims. Supports
	// answers "can this engine synthesise speech"; these answer "which URL
	// does it answer, with which method, and what may the body contain" —
	// which is the question a caller has to get right before sending, and
	// until this existed the only way to answer it was to try.
	CapabilitySchemaVersion int                `json:"capability_schema_version,omitempty"`
	ExecutionLocation       string             `json:"execution_location,omitempty"`
	Availability            *modelAvailability `json:"availability,omitempty"`
	Operations              []modelOperation   `json:"operations,omitempty"`
	// Authoritative is a pointer because absent and false are different
	// things here: absent is a Router that does not publish the catalogue at
	// all, false is one that does and is describing a model whose application
	// never declared it. Both mean "do not trust this list", and only the
	// first also means "do not expect one".
	Authoritative   *bool `json:"authoritative,omitempty"`
	CapabilityStale bool  `json:"capability_stale,omitempty"`
}

// modelAvailability is why a model cannot be called, when it cannot. Readiness
// above says the same thing in one word; this says it in a code a script can
// branch on and a flag saying whether waiting is the answer.
type modelAvailability struct {
	State      string `json:"state"`
	ReasonCode string `json:"reason_code,omitempty"`
	Retryable  bool   `json:"retryable"`
}

// modelOperation is one route the model's application declared, mirrored from
// its Model Console `/api/endpoints`. The shape is Router's, and the fields
// this CLI does not render are decoded anyway so that -o json says as much as
// Router did.
type modelOperation struct {
	ID       string `json:"id"`
	Protocol string `json:"protocol"`
	Method   string `json:"method"`
	// PathTemplate is Router-prefixed and may carry `{placeholder}` segments,
	// as in /v1/text-to-speech/{voice_id}. It is matched a segment at a time.
	PathTemplate string `json:"path_template"`
	// Transport is http or websocket. A method of WS goes with the latter.
	Transport        string               `json:"transport"`
	SyncSupported    bool                 `json:"sync_supported"`
	Streaming        bool                 `json:"streaming,omitempty"`
	AsyncSupported   bool                 `json:"async_supported"`
	RequiredSupports []string             `json:"required_supports,omitempty"`
	InputModalities  []string             `json:"input_modalities,omitempty"`
	OutputModalities []string             `json:"output_modalities,omitempty"`
	OutputFormats    []string             `json:"output_formats,omitempty"`
	SampleRates      []int                `json:"sample_rates,omitempty"`
	Parameters       []operationParameter `json:"parameters,omitempty"`
	Limits           map[string]any       `json:"limits,omitempty"`
	ResourceScope    string               `json:"resource_scope,omitempty"`
	Extensions       map[string]any       `json:"extensions,omitempty"`
}

type operationParameter struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Required bool     `json:"required,omitempty"`
	Default  any      `json:"default,omitempty"`
	Minimum  *float64 `json:"minimum,omitempty"`
	Maximum  *float64 `json:"maximum,omitempty"`
	Enum     []string `json:"enum,omitempty"`
	Unit     string   `json:"unit,omitempty"`
}

// declaresOperation answers the question Router's own gate asks: does this
// model name this route. The matching is Router's — a segment at a time, with
// `{placeholder}` accepting any one non-empty segment — because a CLI that
// matched more loosely than the gate would choose a path the gate then refuses.
func (m *modelObject) declaresOperation(method, requestPath, transport string) (*modelOperation, bool) {
	for i := range m.Operations {
		op := &m.Operations[i]
		if !strings.EqualFold(op.Method, method) || !strings.EqualFold(op.Transport, transport) {
			continue
		}
		if operationPathMatches(op.PathTemplate, requestPath) {
			return op, true
		}
	}
	return nil, false
}

func operationPathMatches(pattern, target string) bool {
	p := strings.Split(path.Clean(pattern), "/")
	t := strings.Split(path.Clean(target), "/")
	if len(p) != len(t) {
		return false
	}
	for i := range p {
		if strings.HasPrefix(p[i], "{") && strings.HasSuffix(p[i], "}") {
			if t[i] == "" {
				return false
			}
			continue
		}
		if p[i] != t[i] {
			return false
		}
	}
	return true
}

// catalogueTrust is how much of a routing decision this model's operation list
// may make. Router publishes the two facts behind it separately, and they carry
// different consequences for a caller.
type catalogueTrust int

const (
	// catalogueGuess is a list nothing was declared for: Router built it from
	// the capability flags, enforces nothing against it, and forwards an
	// undeclared route to the engine, which answers with its own bare 404.
	catalogueGuess catalogueTrust = iota
	// catalogueAdvisory is a declaration Router has stopped holding requests
	// to. It waives the gate fifteen minutes after the catalogue was last
	// observed, because a Model Console that has been unreachable since then
	// cannot go on refusing capabilities the user may have installed in the
	// meantime. The declaration is still the best reading of the engine, so
	// it decides which route to try first — but it can no longer make the
	// other one unreachable, and getting it wrong costs the 404 it always
	// cost rather than a refusal.
	catalogueAdvisory
	// catalogueEnforced is a current declaration. Router refuses an operation
	// it does not list before the engine is reached, so a spelling the
	// catalogue omits is not worth sending: the retry that made a wrong guess
	// harmless can only buy a second refusal here.
	catalogueEnforced
)

// catalogueTrust reads the standing off the row. An empty list is nothing to
// enforce however it is labelled — that is the shape Router publishes when
// every endpoint an application declared was refused, and it degrades to the
// supports-derived list anyway.
func (m *modelObject) catalogueTrust() catalogueTrust {
	if m.Authoritative == nil || !*m.Authoritative || len(m.Operations) == 0 {
		return catalogueGuess
	}
	if m.CapabilityStale {
		return catalogueAdvisory
	}
	return catalogueEnforced
}

type modelsListResponse struct {
	Object string        `json:"object"`
	Data   []modelObject `json:"data"`
}

func newCallModelsCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output          string
		apiKey          string
		includeNotReady bool
		operations      bool
	)
	cmd := &cobra.Command{
		Use:   "models",
		Short: "the names the other call verbs accept in --model",
		Long: `List what this credential may send to "router call".

A name in here is spelled <provider>/<model> and can be sent right now. Routes
— an alias, a group, or a default category like "default-chat" — are callable
too and are deliberately not listed, because they describe no single model;
"olares-cli router route list" is where those names live.

This is the data plane's answer, over the same key every other "call" verb uses,
so it is filtered by the credential making the request. A key restricted to a
few models sees only those, which makes this the check to run when a client
reports that a model does not exist: if a name is missing here, that key cannot
call it, whatever "router model list" says.

A locally installed model application can be missing for a second reason. It
keeps its "router model list" row from the moment it is installed, but it
reaches this list only once its container is up and its weights are loaded,
which are minutes apart. --include-not-ready widens the read to the container
alone: a model still fetching its weights then shows as "warming", and one that
could not load them shows as "failed", rather than both looking like nothing was
ever configured.

"olares-cli router model list" is the other view — one row per configured model
with its provider, mode and health, for deciding what to change rather than what
to send.

--operations asks a different question. The capabilities above are what a model
says it can do; the operations are the routes it actually answers — the method,
the path, whether the work can be submitted asynchronously, and what the body
may contain. For audio that is the difference between knowing an engine
synthesises speech and knowing whether it answers "/v1/audio/speech" or
"/v1/text-to-speech/<voice>", which are two spellings of the same job that no
engine serves both of.

A catalogue is marked as declared when the application published it, and
Router then enforces it: an operation it does not list is refused before the
engine sees it. Otherwise it is Router's own reconstruction from the capability
flags, which nothing is enforced against.

A declared catalogue is also stamped with when it was last observed. Router
enforces an old one exactly as it enforces a fresh one, so the age is something
to go and check on the engine rather than a reason to distrust what is here.

Examples:
  olares-cli router call models
  olares-cli router call models --include-not-ready
  olares-cli router call models --operations --api-key sk-…
  olares-cli router call models --operations -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runModels(c.Context(), f, apiKey, includeNotReady, operations, output)
		},
	}
	cmd.Flags().BoolVar(&includeNotReady, "include-not-ready", false,
		"also list models whose weights are still loading or failed to load")
	cmd.Flags().BoolVar(&operations, "operations", false,
		"print the routes each model declares — method, path, async — instead of the summary table")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "list what this `sk-*` key may call, rather than the credential this machine uses")
	addOutputFlag(cmd, &output)
	return cmd
}

func runModels(ctx context.Context, f *cmdutil.Factory, apiKey string, includeNotReady, operations bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, apiKey)
	var resp modelsListResponse
	if err := dp.doJSON(ctx, "GET", modelsPath(includeNotReady, operations), nil, &resp); err != nil {
		return callErr(err)
	}
	items := resp.Data
	sortModels(items)
	if format == FormatJSON {
		return printJSON(os.Stdout, modelsListResponse{Object: nonEmpty(resp.Object), Data: items})
	}
	if operations {
		return renderModelOperations(os.Stdout, items)
	}
	return renderModelsList(os.Stdout, items, includeNotReady)
}

// sortModels puts the list in the order a person reads it in. Router returns
// one entry per name and says nothing about their order, and a caller scanning
// for a name they half-remember is doing so alphabetically.
func sortModels(items []modelObject) {
	sort.SliceStable(items, func(i, j int) bool { return items[i].ID < items[j].ID })
}

// modelsPath asks for the wider read only when it was asked for. Left on by
// default this verb would stop meaning "what can I send", which is the question
// it exists to answer.
func modelsPath(includeNotReady, operations bool) string {
	q := url.Values{}
	if includeNotReady {
		q.Set("include_not_ready", "true")
	}
	// The catalogue is a per-model join Router does not pay for on the plain
	// list, and the plain list is what every other verb in this tree reads.
	if operations {
		q.Set("detail", "capabilities")
	}
	if len(q) == 0 {
		return epDataPlaneModels
	}
	return withQuery(epDataPlaneModels, q)
}

func renderModelsList(w io.Writer, items []modelObject, includeNotReady bool) error {
	if len(items) == 0 {
		msg := "this credential can call nothing. Either no model is configured, or the key's " +
			"allowed list is empty; `olares-cli router model list` says which."
		if !includeNotReady {
			msg += " A model whose weights are still loading is not in this list either — " +
				"`--include-not-ready` shows it."
		}
		_, err := fmt.Fprintln(w, msg)
		return err
	}
	// AT ONCE only where something declared it, which is only ever a local
	// engine. A column of dashes on a list of cloud models would read as a
	// figure nobody filled in.
	wide := false
	for i := range items {
		wide = wide || items[i].MaxConcurrency > 0
	}
	headers := []string{"NAME", "MODE", "SUPPORTS", "READINESS"}
	if wide {
		headers = append(headers, "AT ONCE")
	}
	headers = append(headers, "SERVED BY")
	t := newTable(w, headers...)
	for i := range items {
		m := &items[i]
		cells := []string{
			m.ID,
			nonEmpty(m.Mode),
			summarizeSupportNames(m.Supports),
			nonEmpty(m.Readiness),
		}
		if wide {
			cells = append(cells, atOnceLabelOf(m.ContextSize, m.MaxConcurrency, m.KVPoolTokens))
		}
		cells = append(cells, clip(nonEmpty(m.OwnedBy), 24))
		t.row(cells...)
	}
	if err := t.flush(); err != nil {
		return err
	}
	if wide {
		_, err := fmt.Fprintln(w, "\nAT ONCE is how many requests that model's engine works on at the "+
			"same time. Sending more does not fail: Router waits for a slot, and a call that waited "+
			"looks like a slow model unless you know the width. shared means those slots share one KV "+
			"pool smaller than (window × width), so a long prompt can be refused while a slot is free.")
		return err
	}
	return nil
}

// renderModelOperations prints one block per model rather than one table for
// all of them: the columns that matter differ by transport, and a websocket
// route has nothing to say about async while an HTTP one says little else.
func renderModelOperations(w io.Writer, items []modelObject) error {
	printed := 0
	for i := range items {
		m := &items[i]
		if len(m.Operations) == 0 {
			continue
		}
		printed++
		if printed > 1 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "%s  (%s, %s)\n", m.ID, nonEmpty(m.Mode), catalogueStanding(m)); err != nil {
			return err
		}
		t := newTable(w, "  OPERATION", "METHOD", "PATH", "SYNC", "ASYNC", "NEEDS")
		for j := range m.Operations {
			op := &m.Operations[j]
			t.row(
				"  "+nonEmpty(op.ID),
				nonEmpty(op.Method),
				nonEmpty(op.PathTemplate),
				yesNo(op.SyncSupported),
				yesNo(op.AsyncSupported),
				dashJoin(op.RequiredSupports),
			)
		}
		if err := t.flush(); err != nil {
			return err
		}
	}
	if printed == 0 {
		_, err := fmt.Fprintln(w, "No model here publishes an operation catalogue. Router builds one from "+
			"the capability flags for an application that does not declare its own, so an empty result "+
			"usually means this Router predates the catalogue rather than that the models serve nothing.")
		return err
	}
	_, err := fmt.Fprintf(w, "\nPATH is Router's, and a braced segment accepts one value: a path ending "+
		"in {voice_id} is addressed as %s. Where the catalogue is declared and still current, Router "+
		"refuses an operation it does not list with `audio_operation_not_supported` rather than "+
		"forwarding it; where it is reconstructed, or was last seen over 15 minutes ago, an undeclared "+
		"route is forwarded and the engine's own 404 comes back.\n",
		epSpeakAs("en-f"))
	return err
}

// catalogueStanding is the standing in words, for the reader of the block it
// heads. It is the same judgement `speak` routes on rather than a second one.
func catalogueStanding(m *modelObject) string {
	switch m.catalogueTrust() {
	case catalogueEnforced:
		return "declared by the application"
	case catalogueAdvisory:
		return "declared, last seen over 15 minutes ago"
	default:
		return "reconstructed from capabilities"
	}
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func dashJoin(v []string) string {
	if len(v) == 0 {
		return "-"
	}
	return strings.Join(v, ",")
}
