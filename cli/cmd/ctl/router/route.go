package router

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router route …` — the names a caller may send that are not a
// qualified `<provider>/<model>` reference.
//
// GET    /console/api/model-routes                        any console session
// GET    /console/api/model-routes/:id                    any console session
// POST   /console/api/model-routes                        admin
// PATCH  /console/api/model-routes/:id                    admin
// DELETE /console/api/model-routes/:id                    admin
// PUT    /console/api/model-routes/:id/members/:model_id  admin
// DELETE /console/api/model-routes/:id/members/:model_id  admin
//
// There are two shapes of name and no others. One containing a slash is a
// qualified reference, split at the first slash, so `openrouter/openai/gpt-5`
// names the model `openai/gpt-5`. One without a slash is a route, and is looked
// up here. A bare model name is not a calling path, and neither is an empty
// `model` field: both were ways of guessing which model somebody meant.
//
// Three kinds share the table and its single unique name, which is what lets a
// name be resolved by looking it up rather than by trying each kind in turn.
// The kinds differ in who owns them: an admin creates and points an alias or a
// group, while a default category is created by Router from a registry in its
// own code and pointed by reconciliation at whatever installed model can serve
// it. So a default takes no target from anyone, and the one thing an admin
// decides about a category is whether it is answered at all — which is the same
// enable and disable the other two kinds take.

// Route kinds, spelled as the wire spells them.
const (
	routeKindAlias   = "alias"
	routeKindGroup   = "group"
	routeKindDefault = "default"
)

// routeKinds is the vocabulary --kind and the list filter accept.
var routeKinds = []string{routeKindAlias, routeKindGroup, routeKindDefault}

// defaultNamePrefix marks a default category, and only a default category:
// Router refuses an alias or group that carries it and a default that does not.
// The CLI knows it so a verb can take the category a person thinks in ("chat")
// as readily as the name a caller sends ("default-chat").
const defaultNamePrefix = "default-"

// modelRoute is one row: the name, what kind of name it is, and what it
// currently answers with.
//
// Members carry the flattened backends for every kind, so "what does this name
// answer with" has one shape: an alias reports its single model here, a group
// its membership in candidate order, and a default whatever its target
// resolves to.
type modelRoute struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Mode    string `json:"mode"`
	Enabled bool   `json:"enabled"`

	Members []routeMember `json:"members"`
	// Target is set on a default only, and only once reconciliation has found
	// something for it. Absent means the category is empty — a state to
	// report rather than an error, since nothing installed can serve it.
	Target *routeTarget `json:"target,omitempty"`
}

// routeMember is one backend behind a name.
//
// Priority orders the candidates, lowest first, and Weight splits traffic
// between the ones sharing a priority. Servable is the dispatcher's own
// liveness verdict, read from the same SQL it filters on, so a route can be
// enabled and still answer nothing.
type routeMember struct {
	ProviderModelID string  `json:"provider_model_id"`
	ProviderID      string  `json:"provider_id"`
	ProviderName    string  `json:"provider_name"`
	ProviderTitle   *string `json:"provider_title,omitempty"`
	OlaresAppName   *string `json:"olares_app_name,omitempty"`
	ModelName       string  `json:"model_name"`
	QualifiedName   string  `json:"qualified_name"`
	Priority        int     `json:"priority"`
	Weight          int     `json:"weight"`
	Servable        bool    `json:"servable"`
}

// routeTarget says what a default resolves to. Exactly one of the two is set:
// a model directly, or another route whose members are the ones listed above.
type routeTarget struct {
	ProviderModelID string `json:"provider_model_id,omitempty"`
	RouteID         string `json:"route_id,omitempty"`
	RouteKind       string `json:"route_kind,omitempty"`
	RouteName       string `json:"route_name,omitempty"`
}

func (r *modelRoute) isDefault() bool { return r.Kind == routeKindDefault }

// live counts the members that can take traffic now. A route with none is
// reachable by name and answers nothing, which is the state worth seeing.
func (r *modelRoute) live() int {
	n := 0
	for i := range r.Members {
		if r.Members[i].Servable {
			n++
		}
	}
	return n
}

// callable reports whether sending this name right now would reach a model.
func (r *modelRoute) callable() bool { return r.Enabled && r.live() > 0 }

// answersWith is the one-cell summary of a route's backends.
func (r *modelRoute) answersWith() string {
	if len(r.Members) == 0 {
		if r.isDefault() {
			return "nothing installed serves it"
		}
		return "nothing"
	}
	first := r.Members[0].label()
	if len(r.Members) == 1 {
		return first
	}
	return fmt.Sprintf("%s and %d more", first, len(r.Members)-1)
}

// via names the hop between a default and its models, which is a route of its
// own when an admin has built one. Empty for the direct case, where the
// category points at a model and the member row already says which.
func (r *modelRoute) via() string {
	if r.Target == nil {
		return ""
	}
	if r.Target.RouteName != "" {
		return r.Target.RouteKind + " " + r.Target.RouteName
	}
	if r.Target.RouteID != "" {
		// The row it points at is gone and reconciliation has not run yet.
		return "a route that no longer exists (" + r.Target.RouteID + ")"
	}
	return ""
}

// label is how a member reads in one cell.
//
// The qualified name alone is not enough for a local application: every one of
// them is a provider called `Olares`, so two members of the same group render
// as the same string, and the one an admin wants to reorder or drop is a
// specific application.
func (m *routeMember) label() string {
	name := m.QualifiedName
	if name == "" {
		name = m.ProviderName + "/" + m.ModelName
	}
	if app := strDeref(m.OlaresAppName); app != "" {
		return name + " (" + app + ")"
	}
	if title := strDeref(m.ProviderTitle); title != "" && title != m.ProviderName {
		return name + " (" + title + ")"
	}
	return name
}

func NewRouteCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route",
		Short: "the names callers may send instead of a provider and model",
		Long: `Manage the names a caller may put in the "model" field.

A name containing a slash is a qualified reference to one model,
"<provider>/<model>", and needs nothing set up. Every other name is a route,
and this is where routes are made.

  alias    a second name for exactly one model
  group    one name served by several models, in priority order with weights
  default  a category Router maintains itself: default-chat, default-tts, and
           one per kind of request

A route name is lowercase letters, digits, '-' and '_', up to 64 characters,
and never contains a slash. That is what keeps the two kinds of name from
overlapping. Names beginning "default-" belong to Router.

The three kinds differ in who owns them. An alias and a group are yours: you
make them, point them and name them. A category is Router's — it comes from a
registry in Router's own code and is pointed at an installed model by
reconciliation, against what that model says it can do. So a category is not
created, renamed, deleted, or given members here, and installing, enabling and
disabling models is what moves it. What an admin does decide about a category
is whether it is answered at all, which is the same enable and disable the
other kinds take. A category may be named in full ("default-chat") or by the
part that varies ("chat").

Subcommands:
  list                    every route, with what it answers with
  get <route>             one route and its backends in candidate order
  create <name>           make an alias or a group
  rename <route> <name>   give a route a different name
  enable <route>          let callers reach it again
  disable <route>         take it out of service without deleting it
  delete <route>          remove it
  add <route> <model>     put a model behind a group
  remove <route> <model>  take one out

Reading is open to everyone: the name is what a user types into their client.
Every change is admin-only.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newRouteListCommand(f))
	cmd.AddCommand(newRouteGetCommand(f))
	cmd.AddCommand(newRouteCreateCommand(f))
	cmd.AddCommand(newRouteRenameCommand(f))
	cmd.AddCommand(newRouteEnableCommand(f, true))
	cmd.AddCommand(newRouteEnableCommand(f, false))
	cmd.AddCommand(newRouteDeleteCommand(f))
	cmd.AddCommand(newRouteAddCommand(f))
	cmd.AddCommand(newRouteRemoveCommand(f))
	return cmd
}

func newRouteListCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		kind   string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "every route, with what it answers with",
		Long: `List the routes configured on this Router.

CALLABLE is the question worth asking of a row: a route is reachable only if it
is switched on and at least one of its backends can take traffic. A group whose
members are all stopped is enabled and answers nothing, which reads as a
working name until somebody calls it.

BACKENDS counts the live ones against the total. ANSWERS WITH names the first
candidate, which for a group is the one traffic goes to first.

--kind default asks the narrower question of what the categories currently
stand at, and answers it in the terms categories are in: what each one answers
with, and the route it goes through when several models share one.

Unpaginated: a deployment has as many routes as it has publicly callable names.

Examples:
  olares-cli router route list
  olares-cli router route list --kind group
  olares-cli router route list --kind default
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runRouteList(c.Context(), f, kind, output)
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "only this kind: "+strings.Join(routeKinds, ", "))
	addOutputFlag(cmd, &output)
	return cmd
}

func runRouteList(ctx context.Context, f *cmdutil.Factory, kind, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind != "" && !containsString(routeKinds, kind) {
		return fmt.Errorf("--kind must be one of %s, not %q", strings.Join(routeKinds, ", "), kind)
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	routes, err := listRoutes(ctx, pc, kind)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": routes})
	}
	// Asking for the categories alone is a different question from asking for
	// the routes, and it wants different columns: the kind is known, and what
	// a category points through matters more than how many backends it has.
	if kind == routeKindDefault {
		return renderDefaults(os.Stdout, routes)
	}
	return renderRoutes(os.Stdout, routes, kind)
}

// listRoutes reads the routes, sorted so the three kinds do not interleave and
// two runs print the same order.
func listRoutes(ctx context.Context, pc *preparedClient, kind string) ([]modelRoute, error) {
	q := url.Values{}
	if kind != "" {
		q.Set("kind", kind)
	}
	routes, err := collection[modelRoute](ctx, pc, withQuery(epModelRoutes, q))
	if err != nil {
		return nil, err
	}
	order := map[string]int{routeKindDefault: 0, routeKindAlias: 1, routeKindGroup: 2}
	sort.SliceStable(routes, func(i, j int) bool {
		if order[routes[i].Kind] != order[routes[j].Kind] {
			return order[routes[i].Kind] < order[routes[j].Kind]
		}
		return routes[i].Name < routes[j].Name
	})
	return routes, nil
}

func renderRoutes(w io.Writer, routes []modelRoute, kind string) error {
	if len(routes) == 0 {
		what := "route"
		if kind != "" {
			what = kind
		}
		_, err := fmt.Fprintf(w, "no %s exists yet. Every model is still reachable as "+
			"<provider>/<model>; `olares-cli router model list` shows those.\n", what)
		return err
	}
	var anyEmptyDefault, anyOff bool
	t := newTable(w, "NAME", "KIND", "MODE", "CALLABLE", "BACKENDS", "ANSWERS WITH")
	for i := range routes {
		r := &routes[i]
		if !r.Enabled {
			anyOff = true
		}
		if r.isDefault() && len(r.Members) == 0 {
			anyEmptyDefault = true
		}
		backends := "-"
		if len(r.Members) > 0 {
			backends = fmt.Sprintf("%d/%d live", r.live(), len(r.Members))
		}
		t.row(r.Name, r.Kind, nonEmpty(r.Mode), boolStr(r.callable()),
			backends, clip(r.answersWith(), 44))
	}
	if err := t.flush(); err != nil {
		return err
	}
	if anyOff {
		if _, err := fmt.Fprintln(w, "\nA route that is not callable because it is switched off is still "+
			"holding its name: `route enable <name>` brings it back with its members intact."); err != nil {
			return err
		}
	}
	if anyEmptyDefault {
		if _, err := fmt.Fprintln(w, "\nA default with no backends is a kind of request nothing installed can "+
			"answer. Router fills it in on its own once a model of that kind exists — `olares-cli router "+
			"route list --kind default` says which categories are waiting."); err != nil {
			return err
		}
	}
	return nil
}

// renderDefaults prints the categories in the terms categories are in. The
// kind column is dropped because it is the question, VIA is added because a
// category pointing through a route of its own is how several models come to
// share one, and BACKENDS goes because what a category answers with matters
// more than how many things could.
func renderDefaults(w io.Writer, routes []modelRoute) error {
	if len(routes) == 0 {
		_, err := fmt.Fprintln(w, "this Router registers no default categories, which means the deployment "+
			"is older than the ones this command reads. Every model is still callable as "+
			"<provider>/<model>; `olares-cli router model list` shows them.")
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
			"`olares-cli router route enable <category>` answers it again."); err != nil {
			return err
		}
	}
	return nil
}

func newRouteGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get <route>",
		Short: "one route and the backends behind it",
		Long: `Show one route: what it is, whether it is reachable, and what answers it.

Backends are listed in the order the dispatcher considers them — by priority,
lowest first, then split by weight among the ones sharing a priority. A member
that cannot take traffic is skipped rather than tried, so LIVE is the column
that explains a name that answers nothing.

A route may be named or given by id.

Examples:
  olares-cli router route get fast
  olares-cli router route get default-chat -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runRouteGet(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runRouteGet(ctx context.Context, f *cmdutil.Factory, ref, outputRaw string) error {
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
	found, err := resolveRoute(ctx, pc, ref)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, found)
	}
	return renderRoute(os.Stdout, found)
}

func renderRoute(w io.Writer, r *modelRoute) error {
	t := newTable(w)
	t.row("NAME", r.Name)
	t.row("KIND", r.Kind)
	t.row("MODE", nonEmpty(r.Mode))
	t.row("SWITCHED ON", boolStr(r.Enabled))
	t.row("CALLABLE NOW", boolStr(r.callable()))
	if v := r.via(); v != "" {
		t.row("VIA", v)
	}
	t.row("ID", r.ID)
	if err := t.flush(); err != nil {
		return err
	}

	if len(r.Members) == 0 {
		if r.isDefault() {
			_, err := fmt.Fprintf(w, "\nNothing installed answers %s requests, so this category is empty. "+
				"Router points it at a model on its own as soon as one exists.\n", nonEmpty(r.Mode))
			return err
		}
		if r.Kind == routeKindGroup {
			_, err := fmt.Fprintf(w, "\nThis group has no backends, so calling %q is answered with a 404. "+
				"`olares-cli router route add %s <model>` puts one behind it.\n", r.Name, r.Name)
			return err
		}
		_, err := fmt.Fprintln(w, "\nThis route names no model.")
		return err
	}

	if _, err := fmt.Fprintln(w, "\nBACKENDS, in the order they are tried:"); err != nil {
		return err
	}
	mt := newTable(w, "MODEL", "SERVED BY", "PRIORITY", "WEIGHT", "LIVE", "MODEL ID")
	for i := range r.Members {
		m := &r.Members[i]
		served := strDeref(m.OlaresAppName)
		if served == "" {
			served = strDeref(m.ProviderTitle)
		}
		if served == "" {
			served = m.ProviderName
		}
		mt.row(nonEmpty(m.QualifiedName), served,
			fmt.Sprintf("%d", m.Priority), fmt.Sprintf("%d", m.Weight),
			boolStr(m.Servable), m.ProviderModelID)
	}
	if err := mt.flush(); err != nil {
		return err
	}
	if r.live() == 0 {
		if _, err := fmt.Fprintf(w, "\nNone of these can take traffic now, so %q is answered with a 404 "+
			"even though it exists. `olares-cli router model list` says why each model is out.\n", r.Name); err != nil {
			return err
		}
	}
	return nil
}

func newRouteCreateCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		kind     string
		modelRef string
		mode     string
		disabled bool
	)
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "make an alias or a group",
		Long: `Create a route callers can name.

An alias is a second name for one model, and takes --model. Its kind of request
comes from that model, so there is nothing to declare: an alias for an
embedding model answers embedding requests.

A group is a name served by several models, and takes --mode. It is created
empty and filled with "route add", which keeps one path for membership rather
than two that have to agree. Until it has a backend the name answers a 404 —
that is the honest state, not a half-finished one.

Every member of a group must answer the same kind of request as the group, so
--mode fixes what the group is for. It cannot be changed afterwards: changing
it would invalidate every member at once, so the honest operation is to delete
the group and make another.

A default category is not created here. Router keeps the list of categories in
its own code and points each one at an installed model itself.

Examples:
  olares-cli router route create fast --kind alias --model claude/claude-sonnet-4-5
  olares-cli router route create house --kind group --mode chat
  olares-cli router route create house --kind group --mode chat --disabled
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runRouteCreate(c.Context(), f, args[0], kind, modelRef, mode, disabled, output)
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "alias or group")
	cmd.Flags().StringVar(&modelRef, "model", "", "the model an alias names")
	cmd.Flags().StringVar(&mode, "mode", "", "the kind of request a group answers: "+strings.Join(providerModelModes, ", "))
	cmd.Flags().BoolVar(&disabled, "disabled", false, "create it switched off, so the name is held but not callable")
	addOutputFlag(cmd, &output)
	return cmd
}

func runRouteCreate(ctx context.Context, f *cmdutil.Factory, name, kind, modelRef, mode string, disabled bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	name, err = requireRef(name, "a route name")
	if err != nil {
		return err
	}
	if err := checkRouteName(name); err != nil {
		return err
	}
	kind = strings.ToLower(strings.TrimSpace(kind))
	switch kind {
	case routeKindAlias, routeKindGroup:
	case routeKindDefault:
		return fmt.Errorf("a default category is not created here: Router keeps the list of them in its own "+
			"code and points each one at an installed model itself. `olares-cli router route list --kind "+
			"default` lists the categories that exist, and `olares-cli router route create %s --kind alias "+
			"--model <model>` is how a name of your own is made", name)
	case "":
		return fmt.Errorf("--kind is required: an alias names one model, a group is served by several")
	default:
		return fmt.Errorf("--kind must be alias or group, not %q", kind)
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	body := map[string]any{"name": name, "kind": kind}
	if disabled {
		body["enabled"] = false
	}

	if kind == routeKindAlias {
		if strings.TrimSpace(modelRef) == "" {
			return fmt.Errorf("an alias needs --model: it is a second name for exactly one model")
		}
		if strings.TrimSpace(mode) != "" {
			return fmt.Errorf("--mode does not apply to an alias: it answers whatever its model answers")
		}
		model, merr := resolveModel(ctx, pc, modelRef)
		if merr != nil {
			return merr
		}
		body["provider_model_id"] = model.ProviderModelID
		body["mode"] = model.Model.Mode
	} else {
		if strings.TrimSpace(modelRef) != "" {
			return fmt.Errorf("--model does not apply to a group: its backends are attached with "+
				"`olares-cli router route add %s <model>`, which is also where priority and weight are set", name)
		}
		mode = strings.ToLower(strings.TrimSpace(mode))
		if mode == "" {
			return fmt.Errorf("a group needs --mode: every member has to answer the same kind of request, "+
				"and the group is what fixes which. One of %s", strings.Join(providerModelModes, ", "))
		}
		if !containsString(providerModelModes, mode) {
			return fmt.Errorf("--mode must be one of %s, not %q", strings.Join(providerModelModes, ", "), mode)
		}
		body["mode"] = mode
	}

	var created modelRoute
	if err := pc.router.doJSON(ctx, http.MethodPost, epModelRoutes, body, &created); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, created)
	}
	if _, err := fmt.Fprintf(os.Stdout, "created %s %s\n\n", created.Kind, created.Name); err != nil {
		return err
	}
	return renderRoute(os.Stdout, &created)
}

func newRouteRenameCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "rename <route> <new-name>",
		Short: "give a route a different name",
		Long: `Rename a route.

Nothing stores a route by value, so the rename is complete the moment it lands:
a key's allowed list holds the name and follows it, and a default's target holds
an id. What does not follow is whatever software is already sending the old
name — those calls start being refused, which is the point of renaming rather
than the risk of it.

A default category cannot be renamed. Its name is the category, and Router owns
the list.

Example:
  olares-cli router route rename fast quick
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runRoutePatch(c.Context(), f, args[0], args[1], nil, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func newRouteEnableCommand(f *cmdutil.Factory, on bool) *cobra.Command {
	var output string
	verb, short, long := "disable", "take a route out of service", `Switch a route off.

The name stays taken and its members stay attached; callers sending it are
refused until it is switched back on. That is the reversible way to stop
traffic to a name — "route delete" is not.

A default category is switched off the same way, and that is the one decision
an admin makes about a category: whether that kind of request is answered at
all. Nothing is uninstalled and nothing is reconfigured — the models serving it
keep running and stay callable as "<provider>/<model>", and Router goes on
maintaining what the category would point at. Disabling the model takes it away
from every caller; disabling the category takes away the convenience of not
naming one. A category may be named in full or by the part that varies.

Examples:
  olares-cli router route disable fast
  olares-cli router route disable chat
`
	if on {
		verb, short, long = "enable", "let callers reach a route again", `Switch a route on.

Whether it then answers is a second question: a route with no live backend is
enabled and still refused. "route get" says which of the two you are looking
at, and for a category "route list --kind default" does.

Examples:
  olares-cli router route enable fast
  olares-cli router route enable chat
`
	}
	cmd := &cobra.Command{
		Use:   verb + " <route>",
		Short: short,
		Long:  long,
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			enabled := on
			return runRoutePatch(c.Context(), f, args[0], "", &enabled, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// runRoutePatch is the one write behind rename, enable and disable: they are
// two columns of the same PATCH, and Router writes both in one statement so a
// rename that lands with an enable flip that fails cannot happen.
func runRoutePatch(ctx context.Context, f *cmdutil.Factory, ref, newName string, enabled *bool, outputRaw string) error {
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
	found, err := resolveRoute(ctx, pc, ref)
	if err != nil {
		return err
	}
	body := map[string]any{}
	if newName != "" {
		name := strings.TrimSpace(newName)
		if err := checkRouteName(name); err != nil {
			return err
		}
		if found.isDefault() {
			return fmt.Errorf("%s is a category Router maintains, and its name is the category, so it "+
				"cannot be renamed. `olares-cli router route disable %s` takes it out of service, and an "+
				"alias or group of your own carries a name you choose", found.Name, found.Name)
		}
		body["name"] = name
	}
	if enabled != nil {
		body["enabled"] = *enabled
	}
	var updated modelRoute
	if err := pc.router.doJSON(ctx, http.MethodPatch, epModelRoute(found.ID), body, &updated); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, updated)
	}
	// A category and a name of your own read differently when switched: a
	// category is never emptied by being refused, because Router goes on
	// maintaining what it would point at.
	switch {
	case newName != "":
		_, err = fmt.Fprintf(os.Stdout, "%s is now called %s; callers sending %q are refused from now on\n",
			found.Name, updated.Name, found.Name)
	case *enabled && updated.isDefault():
		if updated.callable() {
			_, err = fmt.Fprintf(os.Stdout, "%s answers again, with %s\n", updated.Name, updated.answersWith())
		} else {
			_, err = fmt.Fprintf(os.Stdout, "%s is switched on, but %s, so calls naming it are still refused\n",
				updated.Name, updated.answersWith())
		}
	case *enabled:
		if updated.callable() {
			_, err = fmt.Fprintf(os.Stdout, "%s is callable again\n", updated.Name)
		} else {
			_, err = fmt.Fprintf(os.Stdout, "%s is switched on, but nothing behind it can take traffic yet, "+
				"so calls naming it are still refused. `olares-cli router route get %s` says what it holds\n",
				updated.Name, updated.Name)
		}
	case updated.isDefault():
		_, err = fmt.Fprintf(os.Stdout, "%s is refused from now on; the models that were serving it keep "+
			"running and stay callable by name\n", updated.Name)
	default:
		_, err = fmt.Fprintf(os.Stdout, "%s is switched off; its name and its backends are kept\n", updated.Name)
	}
	return err
}

func newRouteDeleteCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:   "delete <route>",
		Short: "remove a route",
		Long: `Delete a route.

The name stops resolving, and anything still sending it is refused. Nothing else
goes: the models behind it are untouched, and remain reachable as
"<provider>/<model>".

If the goal is to stop traffic rather than to give up the name, "route disable"
does that reversibly and keeps the membership.

A default category cannot be deleted — Router would create it again on its next
pass. Switch it off instead.

Confirmation is required. --yes skips the prompt and is mandatory when stdin is
not a terminal.

Example:
  olares-cli router route delete fast --yes
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runRouteDelete(c.Context(), f, args[0], assumeYes, output)
		},
	}
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt (required when stdin is not a terminal)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runRouteDelete(ctx context.Context, f *cmdutil.Factory, ref string, assumeYes bool, outputRaw string) error {
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
	found, err := resolveRoute(ctx, pc, ref)
	if err != nil {
		return err
	}
	if found.isDefault() {
		return fmt.Errorf("%s is a category Router maintains, so deleting it would only last until its next "+
			"pass. `olares-cli router route disable %s` refuses that kind of request instead, which is "+
			"what deleting one would have meant", found.Name, found.Name)
	}
	if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
		fmt.Sprintf("Delete %s %q? Callers sending that name will be refused; the %d models behind it are untouched.",
			found.Kind, found.Name, len(found.Members)),
		assumeYes); err != nil {
		return err
	}
	if err := pc.router.doJSON(ctx, http.MethodDelete, epModelRoute(found.ID), nil, nil); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"id": found.ID, "name": found.Name, "deleted": true})
	}
	_, err = fmt.Fprintf(os.Stdout, "deleted %s %s\n", found.Kind, found.Name)
	return err
}

func newRouteAddCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		priority int
		weight   int
	)
	cmd := &cobra.Command{
		Use:   "add <route> <model>",
		Short: "put a model behind a group",
		Long: `Attach a model to a group, or change how much traffic it takes.

Priority orders the candidates, lowest first: everything at priority 10 is tried
before anything at 20, and a request only reaches the second tier when every
member of the first refused it. Weight splits traffic between the members
sharing one priority, in proportion — two members at weight 1 and 3 see a
quarter and three quarters of it.

Adding a model that is already a member rewrites its priority and weight rather
than failing, so this command states the membership you want without having to
read it first.

The model has to answer the same kind of request as the group. Only a group
takes members: an alias names one model by definition, and a default is pointed
by Router.

Examples:
  olares-cli router route add house claude/claude-sonnet-4-5
  olares-cli router route add house olares/qwen3-8b --priority 20
  olares-cli router route add house openai/gpt-5 --priority 10 --weight 3
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			flags := c.Flags()
			var p, wt *int
			if flags.Changed("priority") {
				p = &priority
			}
			if flags.Changed("weight") {
				wt = &weight
			}
			return runRouteAdd(c.Context(), f, args[0], args[1], p, wt, output)
		},
	}
	cmd.Flags().IntVar(&priority, "priority", 100, "tier this member sits in; lower is tried first")
	cmd.Flags().IntVar(&weight, "weight", 1, "share of the traffic within its priority")
	addOutputFlag(cmd, &output)
	return cmd
}

func runRouteAdd(ctx context.Context, f *cmdutil.Factory, routeRef, modelRef string, priority, weight *int, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	if priority != nil && *priority <= 0 {
		return fmt.Errorf("--priority must be greater than 0")
	}
	if weight != nil && *weight <= 0 {
		return fmt.Errorf("--weight must be greater than 0")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	found, err := resolveRoute(ctx, pc, routeRef)
	if err != nil {
		return err
	}
	if err := refuseNonGroupMembership(found, "attach a model to"); err != nil {
		return err
	}
	model, err := resolveModel(ctx, pc, modelRef)
	if err != nil {
		return err
	}
	if !strings.EqualFold(model.Model.Mode, found.Mode) {
		return fmt.Errorf("%s answers %s requests and %s is a %s route, so it cannot serve it; every member "+
			"of a group answers the same kind of request",
			model.label(), nonEmpty(model.Model.Mode), found.Name, nonEmpty(found.Mode))
	}
	body := map[string]any{}
	if priority != nil {
		body["priority"] = *priority
	}
	if weight != nil {
		body["weight"] = *weight
	}
	var updated modelRoute
	if err := pc.router.doJSON(ctx, http.MethodPut,
		epModelRouteMember(found.ID, model.ProviderModelID), body, &updated); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, updated)
	}
	if _, err := fmt.Fprintf(os.Stdout, "%s now serves %s\n\n", model.label(), updated.Name); err != nil {
		return err
	}
	return renderRoute(os.Stdout, &updated)
}

func newRouteRemoveCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "remove <route> <model>",
		Short: "take a model out of a group",
		Long: `Detach a model from a group.

Removing a model that is not a member succeeds: the state asked for already
holds. The model itself is untouched and stays reachable as
"<provider>/<model>".

A group emptied this way keeps its name and answers a 404, which is worth
knowing before removing the last member of something in use.

Example:
  olares-cli router route remove house olares/qwen3-8b
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runRouteRemove(c.Context(), f, args[0], args[1], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runRouteRemove(ctx context.Context, f *cmdutil.Factory, routeRef, modelRef, outputRaw string) error {
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
	found, err := resolveRoute(ctx, pc, routeRef)
	if err != nil {
		return err
	}
	if err := refuseNonGroupMembership(found, "take a model out of"); err != nil {
		return err
	}
	// The member is looked for in the route itself first. A model already
	// removed from Router leaves the membership behind on nothing, and the
	// aggregate model list can no longer name it.
	modelID := memberIDInRoute(found, modelRef)
	if modelID == "" {
		model, merr := resolveModel(ctx, pc, modelRef)
		if merr != nil {
			return merr
		}
		modelID = model.ProviderModelID
	}
	var updated modelRoute
	if err := pc.router.doJSON(ctx, http.MethodDelete,
		epModelRouteMember(found.ID, modelID), nil, &updated); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, updated)
	}
	if _, err := fmt.Fprintf(os.Stdout, "%s no longer serves %s\n\n", modelRef, updated.Name); err != nil {
		return err
	}
	return renderRoute(os.Stdout, &updated)
}

// memberIDInRoute matches a typed reference against the members a route
// already carries, by qualified name, model name, application name or id.
func memberIDInRoute(r *modelRoute, ref string) string {
	ref = strings.TrimSpace(ref)
	for i := range r.Members {
		m := &r.Members[i]
		switch {
		case m.ProviderModelID == ref,
			strings.EqualFold(m.QualifiedName, ref),
			strings.EqualFold(m.ModelName, ref):
			return m.ProviderModelID
		}
	}
	return ""
}

// refuseNonGroupMembership explains why the other two kinds take no members,
// which is a property of what they are rather than a permission.
func refuseNonGroupMembership(r *modelRoute, doing string) error {
	if r.Kind == routeKindGroup {
		return nil
	}
	if r.Kind == routeKindAlias {
		target := "one model"
		if len(r.Members) > 0 {
			target = r.Members[0].label()
		}
		return fmt.Errorf("%s is an alias, which names %s and nothing else, so there is no membership to %s. "+
			"A name served by several models is a group: `olares-cli router route create <name> --kind group "+
			"--mode %s`", r.Name, target, doing, nonEmpty(r.Mode))
	}
	return fmt.Errorf("%s is a category Router points itself, against whatever is installed, so its backends "+
		"are not yours to %s. A group of your own can be built and then made the answer for that kind of "+
		"request by installing or enabling what it holds", r.Name, doing)
}

// checkRouteName applies Router's own name rule before the round trip. The
// rule is worth restating locally because the reason for it is not obvious
// from a 422: a name containing a slash would be indistinguishable from a
// qualified `<provider>/<model>` reference, and a name beginning `default-`
// would impersonate a category Router maintains.
func checkRouteName(name string) error {
	if name == "" {
		return fmt.Errorf("a route name is required")
	}
	if len(name) > 64 {
		return fmt.Errorf("a route name is at most 64 characters; %q is %d", name, len(name))
	}
	if strings.HasPrefix(name, defaultNamePrefix) {
		return fmt.Errorf("names beginning %q belong to the categories Router maintains, so %q would "+
			"impersonate one; `olares-cli router route list --kind default` lists them",
			defaultNamePrefix, name)
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
		case r == '/':
			return fmt.Errorf("a route name cannot contain '/': that is what tells a route name apart from a "+
				"%q reference, which needs no route at all", "<provider>/<model>")
		case r >= 'A' && r <= 'Z':
			return fmt.Errorf("a route name is lowercase: %q and %q would be two different names, and only "+
				"one of them would be the one clients send", name, strings.ToLower(name))
		default:
			return fmt.Errorf("a route name may contain lowercase letters, digits, '-' and '_' only; %q has %q",
				name, r)
		}
	}
	return nil
}

// resolveRoute finds one route by name or id, and returns it with its members.
//
// The name is what people have: it is the string clients send, it is unique
// across all three kinds, and the id appears nowhere a person would read. So
// the list is consulted first and the detail route is used only to confirm an
// id — which also means a name that does not exist is answered with the names
// that do, rather than with a 404.
func resolveRoute(ctx context.Context, pc *preparedClient, ref string) (*modelRoute, error) {
	ref, err := requireRef(ref, "a route name or id")
	if err != nil {
		return nil, err
	}
	routes, err := listRoutes(ctx, pc, "")
	if err != nil {
		return nil, err
	}
	for i := range routes {
		if strings.EqualFold(routes[i].Name, ref) {
			return &routes[i], nil
		}
	}
	// A category is also findable by the part of it that varies. People think
	// in "chat" and clients send "default-chat", and only Router mints these
	// names, so the expansion cannot collide with an alias or a group.
	if !strings.HasPrefix(strings.ToLower(ref), defaultNamePrefix) {
		want := defaultNamePrefix + strings.ToLower(ref)
		for i := range routes {
			if routes[i].isDefault() && strings.EqualFold(routes[i].Name, want) {
				return &routes[i], nil
			}
		}
	}
	if entityID.MatchString(ref) {
		for i := range routes {
			if routes[i].ID == ref {
				return &routes[i], nil
			}
		}
		var one modelRoute
		if err := pc.router.doJSON(ctx, http.MethodGet, epModelRoute(ref), nil, &one); err != nil {
			return nil, err
		}
		return &one, nil
	}
	known := make([]string, 0, len(routes))
	for i := range routes {
		known = append(known, routes[i].Name)
	}
	return nil, missing{
		noun:  "route",
		ref:   ref,
		known: known,
		have:  "the names that exist are",
		none:  "no route exists yet",
		note: "A model can also be called as `<provider>/<model>` with no route at all; " +
			"`olares-cli router model list` shows those.",
	}.err()
}
