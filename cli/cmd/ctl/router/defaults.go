package router

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router default …` — which model answers a request that names a
// category rather than a model.
//
// GET   /console/api/model-routes?kind=default   any console session
// PATCH /console/api/model-routes/:id            admin
//
// A default is a route like any other, and callers reach one by name:
// `default-chat`, `default-tts`, one per kind of request. What makes it
// different is who decides what it answers with. The categories themselves come
// from a registry in Router's own code, and each one is pointed at an installed
// model by reconciliation, against what that model says it can do. So there is
// nothing here that sets a target — an admin's decision about a category is
// whether it is answered at all.
//
// Two things this deliberately no longer has. There is no fallback layer: a
// category with nothing behind it is refused rather than answered by the oldest
// enabled model of that kind, which was a value nobody chose and which moved as
// models came and went. And there is no per-user override; every caller in a
// deployment gets the same answer, and one that wants a specific model names it.

func NewDefaultCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default",
		Short: "the categories a caller can ask for instead of a model",
		Long: `Show and switch the default categories.

A caller that does not want to choose a model names a category instead:
"default-chat" for a conversation, "default-tts" to read something aloud, one
per kind of request. Router keeps the list of categories itself and points each
one at an installed model that can serve it, so what a category answers with is
not configured here — installing, enabling and disabling models is what moves it.

A category with nothing behind it is refused rather than answered by something
approximate. That is the honest failure: a request for speech served by a chat
model is worse than a request that was turned down.

Subcommands:
  show               every category, what answers it, and whether it is on
  enable <category>  answer that kind of request again
  disable <category> refuse it, without uninstalling anything

A category may be named in full ("default-chat") or by the part that varies
("chat"). Reading is open to everyone, since the name is what a user types into
their client; switching one is admin-only.

Aliases and groups — names of your own — are "olares-cli router route".
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newDefaultShowCommand(f))
	cmd.AddCommand(newDefaultToggleCommand(f, true))
	cmd.AddCommand(newDefaultToggleCommand(f, false))
	return cmd
}

func newDefaultShowCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "every category and what currently answers it",
		Long: `Show the default categories.

ANSWERS WITH is what Router has chosen for that category right now. It changes
on its own as models are installed, enabled and disabled — that is the design,
not a bug: a category is a kind of request, and which model serves it best is a
function of what is available.

CALLABLE is the question a caller cares about. A category is reachable only if
it is switched on and what it points at can take traffic, so a category can be
on and still be refused because the model behind it is stopped.

VIA appears when a category points at a route of its own rather than straight at
a model, which is how several models come to share one category.

Examples:
  olares-cli router default show
  olares-cli router default show -o json
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runDefaultShow(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runDefaultShow(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	routes, err := listRoutes(ctx, pc, routeKindDefault)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": routes})
	}
	return renderDefaults(os.Stdout, routes)
}

func renderDefaults(w io.Writer, routes []modelRoute) error {
	if len(routes) == 0 {
		_, err := fmt.Fprintln(w, "this Router registers no default categories, which means the deployment "+
			"is older than the ones this command reads. Every model is still callable as "+
			"<provider>/<model>; `olares-cli router list` shows them.")
		return err
	}
	var anyEmpty, anyOff bool
	t := newTable(w, "CATEGORY", "MODE", "CALLABLE", "ANSWERS WITH", "VIA")
	for i := range routes {
		r := &routes[i]
		if !r.Enabled {
			anyOff = true
		}
		if len(r.Members) == 0 {
			anyEmpty = true
		}
		t.row(r.Name, nonEmpty(r.Mode), boolStr(r.callable()),
			clip(r.answersWith(), 44), nonEmpty(r.via()))
	}
	if err := t.flush(); err != nil {
		return err
	}
	if anyEmpty {
		if _, err := fmt.Fprintln(w, "\nA category nothing serves is refused, not approximated. Install or "+
			"enable a model of that kind and Router points the category at it on its own — "+
			"`olares-cli market install <app>` is where local models come from, and "+
			"`olares-cli router provider create` is where a cloud vendor does."); err != nil {
			return err
		}
	}
	if anyOff {
		if _, err := fmt.Fprintln(w, "\nA category switched off is refused even when something could serve it. "+
			"`olares-cli router default enable <category>` answers it again."); err != nil {
			return err
		}
	}
	return nil
}

func newDefaultToggleCommand(f *cmdutil.Factory, on bool) *cobra.Command {
	var output string
	verb, short, long := "disable", "refuse a kind of request", `Switch a default category off.

Callers naming it are refused. Nothing is uninstalled and nothing is
reconfigured: the models that were serving the category keep running and stay
callable as "<provider>/<model>", and Router goes on maintaining what the
category would point at.

That is the difference between this and disabling the model itself. Disabling
the model takes it away from every caller; disabling the category takes away the
convenience of not naming one.

Example:
  olares-cli router default disable chat
`
	if on {
		verb, short, long = "enable", "answer a kind of request again", `Switch a default category on.

Whether it then answers is a second question: a category with nothing behind it
is enabled and still refused. "default show" distinguishes the two.

Example:
  olares-cli router default enable chat
`
	}
	cmd := &cobra.Command{
		Use:   verb + " <category>",
		Short: short,
		Long:  long,
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runDefaultToggle(c.Context(), f, args[0], on, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runDefaultToggle(ctx context.Context, f *cmdutil.Factory, ref string, on bool, outputRaw string) error {
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
	found, err := resolveDefaultCategory(ctx, pc, ref)
	if err != nil {
		return err
	}
	var updated modelRoute
	if err := pc.router.doJSON(ctx, http.MethodPatch, epModelRoute(found.ID),
		map[string]any{"enabled": on}, &updated); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, updated)
	}
	switch {
	case !on:
		_, err = fmt.Fprintf(os.Stdout, "%s is refused from now on; %s\n", updated.Name,
			"the models that were serving it keep running and stay callable by name")
	case updated.callable():
		_, err = fmt.Fprintf(os.Stdout, "%s answers again, with %s\n", updated.Name, updated.answersWith())
	default:
		_, err = fmt.Fprintf(os.Stdout, "%s is switched on, but %s, so calls naming it are still refused\n",
			updated.Name, updated.answersWith())
	}
	return err
}

// resolveDefaultCategory finds one category, accepting both the name a caller
// sends and the part of it a person thinks in. Only the categories are read, so
// naming an alias here says so rather than switching off something else that
// happens to share the word.
func resolveDefaultCategory(ctx context.Context, pc *preparedClient, ref string) (*modelRoute, error) {
	ref, err := requireRef(ref, "a category, such as chat or default-chat")
	if err != nil {
		return nil, err
	}
	routes, err := listRoutes(ctx, pc, routeKindDefault)
	if err != nil {
		return nil, err
	}
	want := strings.ToLower(ref)
	if !strings.HasPrefix(want, defaultNamePrefix) {
		want = defaultNamePrefix + want
	}
	for i := range routes {
		if strings.EqualFold(routes[i].Name, want) || routes[i].ID == ref {
			return &routes[i], nil
		}
	}
	known := make([]string, 0, len(routes))
	for i := range routes {
		known = append(known, strings.TrimPrefix(routes[i].Name, defaultNamePrefix))
	}
	return nil, missing{
		noun:  "default category",
		ref:   ref,
		known: known,
		have:  "the categories are",
		none:  "this Router registers none",
		note: "A name of your own is an alias or a group: `olares-cli router route create <name> " +
			"--kind alias --model <model>`.",
	}.err()
}

// modelNames maps model ids to a readable label, for the routes and settings
// that carry an id where a person expects a name. A failed lookup is not worth
// failing a command over, so the id is shown bare instead.
func modelNames(ctx context.Context, pc *preparedClient) map[string]string {
	rows, err := listAllModels(ctx, pc)
	if err != nil {
		return nil
	}
	out := make(map[string]string, len(rows))
	for i := range rows {
		out[rows[i].ProviderModelID] = rows[i].label()
	}
	return out
}

func listAllModels(ctx context.Context, pc *preparedClient) ([]adminModelRow, error) {
	return collection[adminModelRow](ctx, pc, withQuery(epProviderModels, url.Values{"limit": {"1000"}}))
}

func modelLabel(names map[string]string, id string) string {
	if id == "" {
		return "-"
	}
	if label, ok := names[id]; ok && label != "" {
		return label
	}
	return "(unknown)"
}
