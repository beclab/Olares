package router

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"

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

Examples:
  olares-cli router call models
  olares-cli router call models --include-not-ready
  olares-cli router call models --api-key sk-… -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runModels(c.Context(), f, apiKey, includeNotReady, output)
		},
	}
	cmd.Flags().BoolVar(&includeNotReady, "include-not-ready", false,
		"also list models whose weights are still loading or failed to load")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "list what this `sk-*` key may call, rather than the credential this machine uses")
	addOutputFlag(cmd, &output)
	return cmd
}

func runModels(ctx context.Context, f *cmdutil.Factory, apiKey string, includeNotReady bool, outputRaw string) error {
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
	dp, _, err := dataPlane(ctx, pc, apiKey)
	if err != nil {
		return err
	}
	var resp modelsListResponse
	if err := dp.doJSON(ctx, "GET", modelsPath(includeNotReady), nil, &resp); err != nil {
		return callErr(err)
	}
	items := resp.Data
	sortModels(items)
	if format == FormatJSON {
		return printJSON(os.Stdout, modelsListResponse{Object: nonEmpty(resp.Object), Data: items})
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
func modelsPath(includeNotReady bool) string {
	if !includeNotReady {
		return epDataPlaneModels
	}
	return withQuery(epDataPlaneModels, url.Values{"include_not_ready": {"true"}})
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
	t := newTable(w, "NAME", "MODE", "SUPPORTS", "READINESS", "SERVED BY")
	for i := range items {
		m := &items[i]
		t.row(
			m.ID,
			nonEmpty(m.Mode),
			summarizeSupportNames(m.Supports),
			nonEmpty(m.Readiness),
			clip(nonEmpty(m.OwnedBy), 24),
		)
	}
	return t.flush()
}
