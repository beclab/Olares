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

// `olares-cli router list` — GET /console/api/provider-models
//
// One request answers "what can this Olares call", across every provider. It is
// the only read in this tree open to a non-admin user, which is deliberate: the
// list carries model and provider metadata and never a credential, and choosing
// which models a key may reach is something a user does for themselves.
//
// A model appearing here is not the same as a model being callable. The row says
// what is configured; whether the upstream answers is what `provider validate`
// and `router call` find out.
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
// renamed there stops being shown rather than failing loudly; `router provider
// models get <provider> <model>` prints whatever the row actually declares.
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
	// why every read of it goes through state().
	ProviderOlaresStatus *string `json:"provider_olares_status,omitempty"`
	// CallableRouteNames are the names a caller may send right now to reach
	// this row; AllRouteNames every name that reaches it, including the ones
	// currently switched off. Router computes both — the callable half depends
	// on liveness and on members outside this page, so a client holding one
	// page cannot work it out.
	CallableRouteNames []string `json:"callable_route_names"`
	AllRouteNames      []string `json:"all_route_names"`
}

// callable mirrors Router's own dispatch predicate: three switches, and for a
// locally installed application the platform phase on top of them. Kept
// identical to the console's isModelRowCallable so the two never disagree about
// a row — a CLI that called something offered where the console called it
// stopped would send its reader looking for the wrong fix.
func (r *adminModelRow) callable() bool {
	if r.ProviderStatus != "active" || !r.Model.Enabled || r.Model.Status != "active" {
		return false
	}
	return !(r.ProviderSource == "olares" && strDeref(r.ProviderOlaresStatus) != "running")
}

// state is the STATE cell: whether a call would reach this row, and when it
// would not, what is in the way.
//
// The phase wins over "disabled" for a local application, because starting an
// application and asking an admin to re-enable a model are different actions and
// a reader who is told the wrong one goes looking for a toggle nobody switched
// off. Where the console stops at one neutral badge this says which switch is
// off, since a line of text is cheaper than a badge and the two have different
// fixes: `provider models update --enable` against `provider update --enable`.
//
// A row whose application has reported no phase keeps the generic verdict.
// Silence is not a state to name, and it is also what an older Router sends.
func (r *adminModelRow) state() string {
	if r.callable() {
		return "callable"
	}
	if phase := strings.TrimSpace(strDeref(r.ProviderOlaresStatus)); r.ProviderSource == "olares" &&
		phase != "" && phase != "running" {
		return phase
	}
	if r.ProviderStatus != "active" {
		return "provider disabled"
	}
	return "disabled"
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

func NewListCommand(f *cmdutil.Factory) *cobra.Command {
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

"STATE" answers whether a call would reach the row. "callable" is yes. Anything
else names what is in the way, and the two kinds ask for different things: a
platform phase — "stopped", "downloading", "failed" — belongs to the model
application and is "olares-cli market" territory, while "disabled" is a switch
an admin threw and "provider models update --enable" restores. A locally
installed application has a row from the moment it is installed, so a model
that has never run is listed here rather than missing.

What a model claims to support, its context window and its prices are not in
this answer — Router keeps them out of the aggregate list. "router provider get
<provider>" carries them.

Being listed does not mean being reachable. This is the configuration; whether
the upstream answers right now is what "provider validate" reports.

Results are capped, newest first. Narrow with --provider, --mode, --search or
--enabled rather than raising --limit when you are looking for one model.

This is the one read here that does not require an admin session: no credential
appears in it, and a user needs it to choose which models their own key may
reach.

Examples:
  olares-cli router list
  olares-cli router list --mode embedding
  olares-cli router list --provider claude --enabled
  olares-cli router list --search qwen -o json
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
	if filter.Mode != "" && !containsString(providerModelModes, filter.Mode) {
		return fmt.Errorf("--mode must be one of %s, not %q", strings.Join(providerModelModes, ", "), filter.Mode)
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
		return printJSON(os.Stdout, env)
	}
	return renderModelList(os.Stdout, env.Items, env.Total, env.Limit, env.Offset)
}

func renderModelList(w io.Writer, items []adminModelRow, total, limit, offset int) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "no models match. If nothing at all is configured, start with "+
			"`olares-cli router provider types` to see what can be added.")
		return err
	}
	t := newTable(w, "MODEL", "PROVIDER", "SERVED BY", "MODE", "SUPPORTS", "STATE", "ROUTES")
	anyRoute := false
	for i := range items {
		it := &items[i]
		if len(it.AllRouteNames) > 0 {
			anyRoute = true
		}
		// The title is the application for a local model and nothing new for a
		// cloud vendor, whose type says more.
		served := strDeref(it.ProviderTitle)
		if served == "" || served == it.ProviderName {
			served = it.ProviderType
		}
		t.row(
			nonEmpty(it.Model.Name),
			nonEmpty(it.ProviderName),
			clip(nonEmpty(served), 24),
			nonEmpty(it.Model.Mode),
			summarizeSupports(it.Model.Supports),
			it.state(),
			clip(it.routeNote(), 30),
		)
	}
	if err := t.flush(); err != nil {
		return err
	}
	if err := pageFooter(w, len(items), total, offset); err != nil {
		return err
	}
	if anyRoute {
		_, err := fmt.Fprintln(w, "\nROUTES are the extra names a caller may send to reach a row — aliases, "+
			"groups and the default categories. Every row is also callable as PROVIDER/MODEL whether or "+
			"not it has any. `olares-cli router route list` shows them.")
		return err
	}
	return nil
}
