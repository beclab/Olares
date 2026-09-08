package router

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router route pin|unpin|candidates …` — choosing what a default
// category answers with.
//
// GET    /console/api/model-routes/:id/candidates  admin
// PUT    /console/api/model-routes/:id/target      admin
// DELETE /console/api/model-routes/:id/target      admin
//
// Every other route verb works on an alias or a group, which are yours to point
// wherever you like. A category is Router's: reconciliation looks at what is
// installed and picks, once per pass, and until these routes existed that pick
// was the only answer available. So "I have three chat models and I want that
// one" had no expression here at all.
//
// Pinning is a standing decision rather than a nudge. Reconciliation stops
// maintaining the category, which is the point and also the cost: a pinned
// category keeps pointing at a model that has been stopped, and keeps pointing
// at it when something better is installed. Unpinning hands it back.

func newRoutePinCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "pin <category> <model>",
		Short: "make one model the answer for a default category",
		Long: `Choose which model a default category answers with.

Router maintains its categories itself: a pass over what is installed picks one
model per category, against what each model says it can do. That is the right
answer on a machine with one chat model and a guess on a machine with three.
This is how to make the choice yourself.

The choice is standing, not a nudge. Once a category is pinned, reconciliation
leaves it alone — which is what makes it survive installing something new, and
also what makes it survive the pinned model being stopped. A pinned category
whose model is not running answers a 404 rather than falling back, and says so
in "route get". That is deliberate: silently answering with a different model is
the thing the pin exists to prevent.

Only a default takes a pin. An alias names one model by definition and a
group's answer is its membership, so both are pointed with their own verbs.

The model has to be one the category accepts: the right kind of request, and
the capabilities that category selects on. "route candidates <category>" lists
the ones that qualify.

A category may be named in full ("default-chat") or by the part that varies
("chat").

Examples:
  olares-cli router route pin chat Olares/qwen3-4b
  olares-cli router route pin default-tts Olares/kokoro-82m
`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runRoutePin(c.Context(), f, args[0], args[1], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runRoutePin(ctx context.Context, f *cmdutil.Factory, routeRef, modelRef, outputRaw string) error {
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
	if err := refuseNonDefaultTarget(found, "pin"); err != nil {
		return err
	}
	model, err := resolveModel(ctx, pc, modelRef)
	if err != nil {
		return err
	}
	var updated modelRoute
	err = pc.router.doJSON(ctx, http.MethodPut, epModelRouteTarget(found.ID),
		map[string]any{"provider_model_id": model.ProviderModelID}, &updated)
	if err != nil {
		return explainTargetMismatch(ctx, pc, found, model.label(), err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, updated)
	}
	if _, err := fmt.Fprintf(os.Stdout, "%s now answers %s, and Router will not change that on its own "+
		"until it is unpinned.\n\n", model.label(), updated.Name); err != nil {
		return err
	}
	return renderRoute(os.Stdout, &updated)
}

func newRouteUnpinCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "unpin <category>",
		Short: "hand a default category back to Router",
		Long: `Stop choosing what a category answers with, and let Router choose again.

Reconciliation runs before the answer comes back, so what this prints is
already Router's own pick rather than an empty category to poll for. That pick
may well be the model that was pinned; the difference is that it is no longer
guaranteed to be, and installing something else can move it.

Unpinning a category nobody pinned succeeds and changes nothing.

Example:
  olares-cli router route unpin chat
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runRouteUnpin(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runRouteUnpin(ctx context.Context, f *cmdutil.Factory, routeRef, outputRaw string) error {
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
	if err := refuseNonDefaultTarget(found, "unpin"); err != nil {
		return err
	}
	var updated modelRoute
	if err := pc.router.doJSON(ctx, http.MethodDelete, epModelRouteTarget(found.ID), nil, &updated); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, updated)
	}
	if _, err := fmt.Fprintf(os.Stdout, "%s is Router's to point again; it now answers with %s.\n\n",
		updated.Name, updated.answersWith()); err != nil {
		return err
	}
	return renderRoute(os.Stdout, &updated)
}

func newRouteCandidatesCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "candidates <category>",
		Short: "the models a default category would accept",
		Long: `List the models that could answer a default category.

A category selects on more than the kind of request: default-tts-clone wants a
speech model that declares voice cloning, and a model of the right mode without
that capability is not a candidate. The rule lives in Router, so this asks
rather than filters "route list" locally.

LIVE is here and is not a filter. Pinning a category to a model whose
application is stopped is a legitimate thing to want — it is how the choice is
made before the application is started — so a stopped model is listed and
marked rather than hidden.

Example:
  olares-cli router route candidates chat
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runRouteCandidates(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runRouteCandidates(ctx context.Context, f *cmdutil.Factory, routeRef, outputRaw string) error {
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
	if err := refuseNonDefaultTarget(found, "list candidates for"); err != nil {
		return err
	}
	items, err := routeCandidates(ctx, pc, found.ID)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": items})
	}
	return renderRouteCandidates(os.Stdout, found, items)
}

func routeCandidates(ctx context.Context, pc *preparedClient, routeID string) ([]routeMember, error) {
	return collection[routeMember](ctx, pc, epModelRouteCandidates(routeID))
}

func renderRouteCandidates(w io.Writer, r *modelRoute, items []routeMember) error {
	if len(items) == 0 {
		_, err := fmt.Fprintf(w, "nothing installed can answer %s. Router leaves the category empty rather "+
			"than approximating it, and there is nothing to pin until a model of that kind exists.\n", r.Name)
		return err
	}
	pinnedID := ""
	if r.TargetPinned && r.Target != nil {
		pinnedID = r.Target.ProviderModelID
	}
	t := newTable(w, "MODEL", "SERVED BY", "LIVE", "PINNED", "MODEL ID")
	for i := range items {
		m := &items[i]
		t.row(nonEmpty(m.QualifiedName), routeMemberServedBy(m), boolStr(m.Servable),
			boolStr(pinnedID != "" && pinnedID == m.ProviderModelID), m.ProviderModelID)
	}
	if err := t.flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "\n`olares-cli router route pin %s <model>` makes one of these the answer.\n", r.Name)
	return err
}

// refuseNonDefaultTarget explains why the other two kinds take no pin. It is
// what they are rather than a permission: an alias has one model by
// definition, and a group's answer is the membership its own verbs edit.
func refuseNonDefaultTarget(r *modelRoute, doing string) error {
	if r.isDefault() {
		return nil
	}
	if r.Kind == routeKindAlias {
		return fmt.Errorf("%s is an alias, which already names exactly one model, so there is nothing to %s. "+
			"`olares-cli router route create %s --kind alias --model <model>` is how an alias is repointed",
			r.Name, doing, r.Name)
	}
	return fmt.Errorf("%s is a group, and what a group answers with is its membership rather than a single "+
		"target, so there is nothing to %s. `olares-cli router route add %s <model> --priority 10` puts a "+
		"model first", r.Name, doing, r.Name)
}

// explainTargetMismatch turns Router's refusal into the list of things that
// would have worked.
//
// The refusal itself already names the reason — the mode the category needs and
// the capability the model is missing — and adding the candidates answers the
// question that follows it. Only for that one refusal: a model that does not
// exist, or a caller who is not an admin, are not helped by a list.
func explainTargetMismatch(ctx context.Context, pc *preparedClient, r *modelRoute, label string, err error) error {
	var re *RouterError
	if !errors.As(err, &re) || re.Code != "model_route_target_mismatch" {
		return err
	}
	items, cerr := routeCandidates(ctx, pc, r.ID)
	if cerr != nil || len(items) == 0 {
		return err
	}
	names := make([]string, 0, len(items))
	for i := range items {
		names = append(names, items[i].label())
	}
	return fmt.Errorf("%w\n%s does accept %s", err, r.Name, strings.Join(names, ", "))
}
