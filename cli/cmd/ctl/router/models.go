package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router models` — every name the `model` field accepts.
//
// GET /v1/models
//
// This is not `router list` with a different table. `router list` is the
// management plane's view: one row per configured model, with its provider, its
// mode, its prices and whether the provider is healthy — everything an admin
// needs to decide what to change. This is the *caller's* view, on the data
// plane, and it answers a narrower question: what may I put in the `model`
// field. Two differences follow from that, and both are the point.
//
// A route appears here as a name in its own right. `default-chat` and a group
// called `fast` are not models and have no row in the model list, but they are
// exactly what a client is meant to send, so the OpenAI-shaped catalogue lists
// them beside the qualified names.
//
// And what appears depends on the credential. A key with an allowed-models list
// sees only what it may call, so this is the honest answer to "will my client
// work with this key" — which `router list`, read over the console session,
// cannot give.

// modelObject is one entry of the OpenAI list envelope. Router adds
// qualified_id, and it being empty is the only thing on the wire that separates
// a route from a model: a route has no provider to qualify it with.
type modelObject struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Created     int64  `json:"created"`
	OwnedBy     string `json:"owned_by"`
	QualifiedID string `json:"qualified_id"`
}

func (m *modelObject) isRoute() bool { return strings.TrimSpace(m.QualifiedID) == "" }

func (m *modelObject) kind() string {
	if m.isRoute() {
		return "route"
	}
	return "model"
}

type modelsListResponse struct {
	Object string        `json:"object"`
	Data   []modelObject `json:"data"`
}

func newModelsCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		apiKey string
		only   string
	)
	cmd := &cobra.Command{
		Use:   "models",
		Short: "every name the model field accepts, as a caller sees it",
		Long: `List the names this Router answers to.

Two kinds of name are in here. A model is spelled <provider>/<model> and names
one configured model. A route is a name of its own — an alias, a group, or a
default category like "default-chat" — and is what a client is usually meant to
send, because it survives the model behind it being replaced.

This is the data plane's answer, so it is filtered by the credential making the
request. A key restricted to a few models sees only those, which makes this the
check to run when a client reports that a model does not exist: if a name is
missing here, that key cannot call it, whatever "router list" says.

"olares-cli router list" is the other view — one row per configured model with
its provider, mode, prices and health, for deciding what to change rather than
what to send.

Examples:
  olares-cli router models
  olares-cli router models --only routes
  olares-cli router models --api-key sk-… -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			kind, err := parseModelsOnly(only)
			if err != nil {
				return err
			}
			return runModels(c.Context(), f, kind, apiKey, output)
		},
	}
	cmd.Flags().StringVar(&only, "only", "", "narrow to `routes` or `models`")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "list what this `sk-*` key may call, rather than the credential this machine uses")
	addOutputFlag(cmd, &output)
	return cmd
}

func parseModelsOnly(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return "", nil
	case "route", "routes":
		return "route", nil
	case "model", "models":
		return "model", nil
	}
	return "", fmt.Errorf("--only takes routes or models, not %q", raw)
}

func runModels(ctx context.Context, f *cmdutil.Factory, only, apiKey, outputRaw string) error {
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
	if err := dp.doJSON(ctx, "GET", epDataPlaneModels, nil, &resp); err != nil {
		return callErr(err)
	}
	items := resp.Data
	if only != "" {
		kept := make([]modelObject, 0, len(items))
		for i := range items {
			if items[i].kind() == only {
				kept = append(kept, items[i])
			}
		}
		items = kept
	}
	// Routes first, then models, alphabetically within each. A caller reading
	// this is choosing a name to send, and the names they should prefer are
	// the ones that outlive a model being swapped.
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].isRoute() != items[j].isRoute() {
			return items[i].isRoute()
		}
		return items[i].ID < items[j].ID
	})
	if format == FormatJSON {
		return printJSON(os.Stdout, modelsListResponse{Object: nonEmpty(resp.Object), Data: items})
	}
	return renderModelsList(os.Stdout, items, only)
}

func renderModelsList(w io.Writer, items []modelObject, only string) error {
	if len(items) == 0 {
		msg := "this credential can call nothing. Either no model is configured, or the key's " +
			"allowed list is empty; `olares-cli router list` says which."
		if only != "" {
			msg = "no " + only + " is callable with this credential."
		}
		_, err := fmt.Fprintln(w, msg)
		return err
	}
	t := newTable(w, "NAME", "KIND", "SERVED BY")
	for i := range items {
		m := &items[i]
		t.row(m.ID, m.kind(), nonEmpty(m.OwnedBy))
	}
	return t.flush()
}
