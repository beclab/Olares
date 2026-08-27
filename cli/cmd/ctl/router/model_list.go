package router

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router model list` — GET /console/api/provider-models
//
// One request answers "what is configured on this Olares", across every
// provider. It is the only read in this tree open to a non-admin user, which is
// deliberate: the list carries model and provider metadata and never a
// credential, and choosing which models a key may reach is something a user
// does for themselves.
//
// A model appearing here is not the same as a model being callable. The row says
// what is configured; whether the upstream answers is what `provider validate`
// and `router call` find out. `router call models` is the other question — what
// a caller's own credential may send — and it is a verb on `call` because it is
// answered by the data plane, with the key `call` uses.
//
// A locally installed model application owns a row from the moment it is
// installed, named after the MODEL_NAME its manifest declares, whatever state
// the application is in. So this list carries models of applications that are
// downloading, stopped or failed, and the STATE column is what separates them
// from the ones a call would reach.

// capabilityFlags are the supports keys that say what a model can do, in the
// order Router declares them. The parameter knobs it tracks alongside them —
// temperature, top_p, seed, stop sequences and a dozen more — are deliberately
// absent: a chat model sets nearly all of those, and a column carrying them
// would push vision and tools off the end of it.
//
// These literals are copied from Router's vocabulary and nothing here can check
// them, the same way the default categories in call.go are copied. A key
// renamed there stops being shown rather than failing loudly; `router model get
// <model>` prints whatever the row actually declares.
var capabilityFlags = []string{
	"supports_vision",
	"supports_function_calling",
	"supports_parallel_function_calling",
	"supports_response_schema",
	"supports_reasoning",
	"supports_prompt_caching",
	"supports_web_search",
	"supports_audio_input",
	"supports_audio_output",
	"supports_video_input",
	"supports_pdf_input",
	"supports_computer_use",
	"supports_url_context",
	"supports_embedding_image_input",
	"supports_stt",
	"supports_stt_stream",
	"supports_align",
	"supports_diar",
	"supports_diar_stream",
	"supports_speaker_embed",
	"supports_vad",
	"supports_enhance",
	"supports_tts",
	"supports_tts_clone",
	"supports_tts_dialogue",
	"supports_audio_llm",
	"supports_audio_s2s",
	"supports_sound_fx",
}

// supportsShown is how many capabilities a table cell names before counting the
// rest. Three is what fits beside the other columns on an ordinary terminal.
const supportsShown = 3

// summarizeSupports names what a row declares it can do. This is the column
// that separates one `audio` row from the next — the mode says six different
// jobs and the flags say which of them this model actually serves.
//
// A row declaring none of them prints "-" rather than an empty cell, since a
// blank reads as missing data, which is a different thing.
func summarizeSupports(supports map[string]bool) string {
	on := make([]string, 0, len(capabilityFlags))
	for _, key := range capabilityFlags {
		if supports[key] {
			on = append(on, strings.TrimPrefix(key, "supports_"))
		}
	}
	switch {
	case len(on) == 0:
		return "-"
	case len(on) <= supportsShown:
		return strings.Join(on, ",")
	}
	return fmt.Sprintf("%s,+%d", strings.Join(on[:supportsShown], ","), len(on)-supportsShown)
}

// summarizeSupportNames is the same cell for the data plane's spelling of the
// same fact. GET /v1/models sends the capabilities as a list of names, already
// narrowed to the ones the card claims and already stripped of the `supports_`
// prefix, so the work summarizeSupports does has been done upstream.
//
// Reusing capabilityFlags to order them would undo that: Router adds a
// capability, this build has never heard of it, and the cell quietly stops
// mentioning it — which reads as a model that cannot do the thing. The server's
// own list and order are kept for exactly that reason.
func summarizeSupportNames(names []string) string {
	on := make([]string, 0, len(names))
	for _, name := range names {
		if trimmed := strings.TrimSpace(name); trimmed != "" {
			on = append(on, trimmed)
		}
	}
	switch {
	case len(on) == 0:
		return "-"
	case len(on) <= supportsShown:
		return strings.Join(on, ",")
	}
	return fmt.Sprintf("%s,+%d", strings.Join(on[:supportsShown], ","), len(on)-supportsShown)
}

// The phases the data plane refuses to dispatch to are phaseBlocksCalls in
// model_phase.go, which is also where the reason for treating an unknown phase
// as servable is written down. It is a deny-list here and in Router for the
// same reason: the phase column only has a value while the loop that writes it
// is running, so an allow-list would hide every model of every application
// that has no Model Console to report one.

// adminModelRow is one model with its provider lifted alongside it, which is
// what makes a flat list readable: the same model name can exist on two
// providers, and the provider is what tells them apart.
//
// ProviderTitle is what tells two rows apart when the provider name cannot:
// every locally installed model application is a provider called `Olares`, so
// the routing name is shared and the title is the application.
type adminModelRow struct {
	ProviderModelID string           `json:"provider_model_id"`
	ProviderID      string           `json:"provider_id"`
	ProviderName    string           `json:"provider_name"`
	ProviderTitle   *string          `json:"provider_title,omitempty"`
	ProviderType    string           `json:"provider_type"`
	ProviderSource  string           `json:"provider_source"`
	ProviderStatus  string           `json:"provider_status"`
	Model           providerModelRow `json:"model"`
	// OlaresAppName is the Olares application serving this model, absent on a
	// manual provider. It is the typeable half of the pair the title carries:
	// a title reads well and an application name is what a reference accepts.
	OlaresAppName *string `json:"olares_app_name,omitempty"`
	// ProviderOlaresStatus is the platform's own phase for that application —
	// downloading, installing, running, stopped, failed. Absent on a manual
	// provider, and absent from a Router old enough not to send it, which is
	// why every read of it goes through callableNote().
	ProviderOlaresStatus *string `json:"provider_olares_status,omitempty"`
	// ModelConsoleStatus is what the application's own Model Console says its
	// weights and engine are doing — init, download, loading, ready,
	// degraded, failed. Absent on a manual provider, which has no Model
	// Console, and cleared by Router whenever the application leaves running,
	// since a container that is gone does not get to report a phase.
	//
	// This is a second axis and not a widening of the first. A container is
	// up minutes before the weights it serves can answer, and the two facts
	// have different owners and different fixes: the platform owns one and
	// `olares-cli market` acts on it, the application owns the other and
	// `router model retry` acts on it.
	ModelConsoleStatus *string `json:"model_console_status,omitempty"`
	// CallableRouteNames are the names a caller may send right now to reach
	// this row; AllRouteNames every name that reaches it, including the ones
	// currently switched off. Router computes both — the callable half depends
	// on liveness and on members outside this page, so a client holding one
	// page cannot work it out.
	CallableRouteNames []string `json:"callable_route_names"`
	AllRouteNames      []string `json:"all_route_names"`
}

// callable mirrors Router's own dispatch predicate: two admin switches, and
// for a locally installed application two more gates on top of them — the
// container has to be up and the weights it serves have to be able to answer.
//
// The weights half is easy to leave out and expensive to leave out. A
// container reports running minutes before the model behind it has finished
// loading, and a list that called those minutes "callable" would be offering a
// name that answers every request with `model_not_ready`.
//
// Kept identical to the console's isModelRowCallable and to Router's
// OlaresModelServableSQL. Nothing here can check that: this is a hand copy of a
// predicate that lives in three places, and a fourth reader disagreeing with
// the data plane is worse than not answering, because it sends a reader after
// the wrong fix.
func (r *adminModelRow) callable() bool {
	if r.ProviderStatus != "active" || !r.Model.Enabled || r.Model.Status != "active" {
		return false
	}
	if r.ProviderSource != "olares" {
		return true
	}
	if strings.TrimSpace(strDeref(r.ProviderOlaresStatus)) != "running" {
		return false
	}
	return !phaseBlocksCalls(strDeref(r.ModelConsoleStatus))
}

// readiness is what the weights alone are doing, in the vocabulary
// `router call models` prints. It answers a narrower question than callable():
// an admin switch being off is not a readiness problem, and a remote vendor has
// no weights to wait for.
//
// It is here so that a reader who saw `warming` on one list can find the same
// word on the other, and so that the JSON does not make a script re-derive it
// from two nullable columns and a deny-list.
func (r *adminModelRow) readiness() string {
	if r.ProviderSource != "olares" {
		return "ready"
	}
	if strings.TrimSpace(strDeref(r.ProviderOlaresStatus)) != "running" {
		return "unknown"
	}
	if strings.TrimSpace(strDeref(r.ModelConsoleStatus)) == "" {
		return "unknown"
	}
	if p, ok := lookupPhase(strDeref(r.ModelConsoleStatus)); ok {
		return p.readiness
	}
	// A phase this build does not know is not evidence against the weights;
	// see modelPhases.
	return "ready"
}

// callableNote is the CALLABLE cell: the verdict first, and when it is no, the
// one thing standing in the way.
//
// Leading with the verdict is what makes a hundred rows scannable, and naming
// only one obstacle is what makes the answer actionable. They are ordered by
// what the reader would do about them: the two admin switches first, because
// they are the only ones changeable from here and reporting the platform
// instead would send somebody to Market to fix an application that is fine;
// then the container; then the weights.
//
// The two axes are spelled apart on purpose. "app downloading" is the platform
// fetching an image and "fetching weights" is the model fetching itself, and a
// reader told the wrong one checks the wrong thing — a model stuck loading is
// an engine that will not start, and no amount of looking at disk or network
// explains it.
func (r *adminModelRow) callableNote() string {
	if r.callable() {
		return "yes"
	}
	switch {
	case r.ProviderStatus != "active":
		return "no · provider disabled"
	case !r.Model.Enabled || r.Model.Status != "active":
		return "no · model disabled"
	}
	if r.ProviderSource != "olares" {
		return "no"
	}
	if app := strings.TrimSpace(strDeref(r.ProviderOlaresStatus)); app != "running" {
		if app == "" {
			return "no · app state unknown"
		}
		return "no · app " + app
	}
	if p, ok := lookupPhase(strDeref(r.ModelConsoleStatus)); ok && !p.servable {
		return "no · " + p.cell
	}
	return "no"
}

// label is how one row reads in a sentence: the qualified name Router routes
// on, and the application when that name is shared.
func (r *adminModelRow) label() string {
	name := r.ProviderName + "/" + r.Model.Name
	if title := strDeref(r.ProviderTitle); title != "" && title != r.ProviderName {
		return name + " (" + title + ")"
	}
	return name
}

// routeNote is the ROUTES cell: the names that reach this model, with the ones
// currently unreachable marked. A model with none is not unreachable — every
// row is callable as <provider>/<model> — so the cell is empty rather than
// alarming.
func (r *adminModelRow) routeNote() string {
	if len(r.AllRouteNames) == 0 {
		return "-"
	}
	callable := make(map[string]bool, len(r.CallableRouteNames))
	for _, n := range r.CallableRouteNames {
		callable[n] = true
	}
	out := make([]string, 0, len(r.AllRouteNames))
	for _, n := range r.AllRouteNames {
		if callable[n] {
			out = append(out, n)
			continue
		}
		out = append(out, n+" (off)")
	}
	return strings.Join(out, ", ")
}

func newModelListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output       string
		providerRef  string
		mode         string
		source       string
		status       string
		enabledOnly  bool
		disabledOnly bool
		search       string
		limit        int
		offset       int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "every model configured across every provider",
		Long: `List the models Router has configured, across all providers.

The same model name can exist on more than one provider, so the provider column
is part of the identity rather than decoration. It can also exist twice on the
same provider name: every locally installed model application is a provider
called "Olares", and "SERVED BY" is the application that tells those apart.

A model is called as "PROVIDER/MODEL". "ROUTES" lists the other names that
reach the same row — aliases, groups and the default categories, which
"router route list" manages.

"CALLABLE" answers whether a call would reach the row, and when the answer is
no it names the one thing in the way. There are four, and each asks for
something different:

  no · provider disabled     "provider update <provider> --enable"
  no · model disabled        "model update <model> --enable"
  no · app <phase>           the application is not running — "olares-cli
                             market" territory, and the phase says where it got
                             to: stopped, downloading, installing, failed
  no · fetching weights      the application is running and the model it serves
     · engine loading        is not ready yet — "model progress <model>" follows
     · model load failed     it, "model retry <model>" restarts a failed one

The last two groups are separate facts with separate owners, which is why they
are spelled apart. "app downloading" is the platform fetching an image;
"fetching weights" is the model fetching itself, minutes later, with the
container already up. A model stuck on "engine loading" is an engine that will
not start, so checking disk and network — the right move for a download — finds
nothing.

A locally installed application has a row from the moment it is installed, so a
model that has never run is listed here rather than missing.

What a model claims to support, its context window and its prices are not in
this answer — Router keeps them out of the aggregate list. "router model get
<model>" carries them.

Being listed does not mean being reachable. This is the configuration; whether
the upstream answers right now is what "provider validate" reports, and what a
particular credential may send is "router call models".

-o json adds three derived fields to each row — "callable", a "readiness" of
ready, warming, failed or unknown, and the "callable_note" this table prints in
its CALLABLE cell — beside the two raw columns they come from,
"provider_olares_status" for the container and "model_console_status" for the
weights. The raw pair is never collapsed: they have different owners and
different fixes.

Results are capped, newest first. Narrow with --provider, --mode, --search or
--enabled rather than raising --limit when you are looking for one model.

This is the one read here that does not require an admin session: no credential
appears in it, and a user needs it to choose which models their own key may
reach.

Examples:
  olares-cli router model list
  olares-cli router model list --mode embedding
  olares-cli router model list --provider claude --enabled
  olares-cli router model list --search qwen -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			if enabledOnly && disabledOnly {
				return fmt.Errorf("pass either --enabled or --disabled, not both")
			}
			filter := modelListFilter{
				ProviderRef: strings.TrimSpace(providerRef),
				Mode:        strings.ToLower(strings.TrimSpace(mode)),
				Source:      strings.ToLower(strings.TrimSpace(source)),
				Status:      strings.ToLower(strings.TrimSpace(status)),
				Search:      strings.TrimSpace(search),
				Limit:       limit,
				Offset:      offset,
			}
			if enabledOnly {
				v := true
				filter.Enabled = &v
			}
			if disabledOnly {
				v := false
				filter.Enabled = &v
			}
			return runModelList(c.Context(), f, filter, output)
		},
	}
	cmd.Flags().StringVar(&providerRef, "provider", "", "only models on this provider, by name or id")
	cmd.Flags().StringVar(&mode, "mode", "", "only this endpoint family: "+strings.Join(providerModelModes, ", "))
	cmd.Flags().StringVar(&source, "source", "", "manual or olares")
	cmd.Flags().StringVar(&status, "status", "", "active or disabled")
	cmd.Flags().BoolVar(&enabledOnly, "enabled", false, "only models offered to callers")
	cmd.Flags().BoolVar(&disabledOnly, "disabled", false, "only models not offered to callers")
	cmd.Flags().StringVar(&search, "search", "", "match part of a model or provider name")
	cmd.Flags().IntVar(&limit, "limit", 100, "how many rows to return (1-1000)")
	cmd.Flags().IntVar(&offset, "offset", 0, "how many rows to skip")
	addOutputFlag(cmd, &output)
	return cmd
}

type modelListFilter struct {
	ProviderRef string
	Mode        string
	Source      string
	Status      string
	Enabled     *bool
	Search      string
	Limit       int
	Offset      int
}

func runModelList(ctx context.Context, f *cmdutil.Factory, filter modelListFilter, outputRaw string) error {
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

	q := url.Values{}
	if filter.ProviderRef != "" {
		// The route filters by provider id, so a name has to be resolved first.
		found, rerr := resolveProvider(ctx, pc, filter.ProviderRef)
		if rerr != nil {
			return rerr
		}
		q.Set("provider_id", found.ID)
	}
	if filter.Mode != "" {
		q.Set("mode", filter.Mode)
	}
	if filter.Source != "" {
		q.Set("source", filter.Source)
	}
	if filter.Status != "" {
		q.Set("status", filter.Status)
	}
	if filter.Enabled != nil {
		q.Set("enabled", strconv.FormatBool(*filter.Enabled))
	}
	if filter.Search != "" {
		q.Set("search", filter.Search)
	}
	if filter.Limit > 0 {
		q.Set("limit", strconv.Itoa(filter.Limit))
	}
	if filter.Offset > 0 {
		q.Set("offset", strconv.Itoa(filter.Offset))
	}

	var env page[adminModelRow]
	if err := pc.router.doJSON(ctx, "GET", withQuery(epProviderModels, q), nil, &env); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, page[modelListJSONRow]{
			Items:  modelListJSONRows(env.Items),
			Total:  env.Total,
			Limit:  env.Limit,
			Offset: env.Offset,
		})
	}
	return renderModelList(os.Stdout, env.Items, env.Total, env.Limit, env.Offset)
}

// modelListJSONRow is the row plus the two things every reader of it works out
// for itself. Both raw columns stay exactly where Router put them: the derived
// values are a third field beside them, never a replacement, because collapsing
// two facts with two owners into one is a mistake Router has already had to
// undo once and a client is no better placed to make it.
//
// The derivation is worth carrying because it is not obvious. Callable is four
// gates, Readiness is a deny-list over a vocabulary that belongs to somebody
// else, and both treat a missing value as permissive — a script writing its
// own version of that gets the empty cases backwards and hides working models.
type modelListJSONRow struct {
	adminModelRow
	// Callable is whether a call would reach this row right now. Router
	// itself can be more permissive: the weights half of its gate is only
	// armed while the loop that writes the phase column is running, so a row
	// reported here as warming may still dispatch. It is never the other way
	// round, and `router call models` is the answer with no derivation in it.
	Callable bool `json:"callable"`
	// Readiness is what the weights are doing, in the words `router call
	// models` uses: ready, warming, failed or unknown.
	Readiness string `json:"readiness"`
	// CallableNote is the reason, and the table's own cell verbatim. Callable
	// says a call would fail; this says which of the four things is in the
	// way, and they have four different fixes — an admin switch, the Market,
	// waiting, or `model retry`. Leaving it out made the machine-readable
	// output the one a caller could act on least, which is backwards.
	CallableNote string `json:"callable_note"`
}

func modelListJSONRows(items []adminModelRow) []modelListJSONRow {
	out := make([]modelListJSONRow, 0, len(items))
	for i := range items {
		out = append(out, modelListJSONRow{
			adminModelRow: items[i],
			Callable:      items[i].callable(),
			Readiness:     items[i].readiness(),
			CallableNote:  items[i].callableNote(),
		})
	}
	return out
}

func renderModelList(w io.Writer, items []adminModelRow, total, limit, offset int) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no models match. If nothing at all is configured, start with "+
			"`olares-cli router provider types` to see what can be added.")
		return err
	}
	// AT ONCE appears only when a row carries it, which is only ever a local
	// model: the figure comes from an engine's launch flags and a cloud vendor
	// has none. A permanent column of dashes would read as a figure nobody
	// filled in rather than one that does not apply.
	wide := false
	for i := range items {
		wide = wide || items[i].Model.MaxConcurrency > 0
	}
	headers := []string{"MODEL", "PROVIDER", "SERVED BY", "MODE", "SUPPORTS"}
	if wide {
		headers = append(headers, "AT ONCE")
	}
	headers = append(headers, "CALLABLE", "ROUTES")
	t := newTable(w, headers...)
	anyRoute := false
	anyWarming := false
	for i := range items {
		it := &items[i]
		if len(it.AllRouteNames) > 0 {
			anyRoute = true
		}
		if it.readiness() == "warming" {
			anyWarming = true
		}
		// The title is the application for a local model and nothing new for a
		// cloud vendor, whose type says more.
		served := strDeref(it.ProviderTitle)
		if served == "" || served == it.ProviderName {
			served = it.ProviderType
		}
		cells := []string{
			nonEmpty(it.Model.Name),
			nonEmpty(it.ProviderName),
			clip(nonEmpty(served), 24),
			nonEmpty(it.Model.Mode),
			summarizeSupports(it.Model.Supports),
		}
		if wide {
			cells = append(cells, intOrDash(it.Model.MaxConcurrency))
		}
		cells = append(cells, it.callableNote(), clip(it.routeNote(), 30))
		t.row(cells...)
	}
	if err := t.flush(); err != nil {
		return err
	}
	if err := pageFooter(w, len(items), total, offset); err != nil {
		return err
	}
	if wide {
		if _, err := fmt.Fprintln(w, "\nAT ONCE is how many requests that model's engine was launched to "+
			"work on at the same time. It is only known for a local engine whose launch flags said so; "+
			"`router provider get <app>` reads the engine's current queue beside it."); err != nil {
			return err
		}
	}
	if anyWarming {
		if _, err := fmt.Fprintln(w, "\nA model fetching weights or loading an engine gets there on its "+
			"own; `olares-cli router model progress <model> --watch` follows it. One that says it failed "+
			"will not, and `router model retry <model>` is what re-enters the loop."); err != nil {
			return err
		}
	}
	if anyRoute {
		_, err := fmt.Fprintln(w, "\nROUTES are the extra names a caller may send to reach a row — aliases, "+
			"groups and the default categories. Every row is also callable as PROVIDER/MODEL whether or "+
			"not it has any. `olares-cli router route list` shows them.")
		return err
	}
	return nil
}
