package router

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router app …` — the model applications that run models here.
//
// GET  /console/api/market/catalog
// POST /console/api/market/install
// POST /console/api/market/providers/:id/upgrade
// POST /console/api/market/providers/:id/uninstall
//
// A model application packages a model and the engine that serves it.
// Installing one through Router rather than through the Market on its own is
// what ties the two together: the provider row exists from the moment the
// install starts, so the application is addressable as soon as it runs.
//
// Nothing here streams. Each of the three lifecycle routes hands the request to
// the Market and answers with the provider row to watch; how far the work has
// got is that row's olares_status, which Router's application directory keeps
// current whether or not anyone is watching. So --watch is a poll of the row,
// and there is no task list and no event stream to ask for.
//
// The catalog carries what this machine knows about each row as well as what
// the Market publishes, which is why every refusal below is decided from one
// read: whether the app is a template, whether its name is already taken here,
// and which source publishes it.

// marketApp is one row of Router's Market catalog: what the Market publishes,
// plus what this machine knows about it.
//
// Title and Description arrive as whole locale maps — the Market is answering a
// request whose reader it cannot see, so it does not pick one.
type marketApp struct {
	AppName string `json:"app_name"`
	// Source is the Market source that published this copy. One app name
	// arrives once per source, and only one copy can be installed.
	Source        string             `json:"source,omitempty"`
	Title         map[string]string  `json:"title"`
	Description   map[string]string  `json:"description"`
	IconURL       string             `json:"icon_url"`
	Version       string             `json:"version"`
	Category      string             `json:"category"`
	TemplateOnly  bool               `json:"template_only"`
	ModelMode     string             `json:"model_mode"`
	ModelSupports []string           `json:"model_supports"`
	Install       marketInstallState `json:"install"`
}

// marketInstallState answers, for one catalog row, whether that row is
// something this machine can install. Router decides it: the Market's catalog
// says what could be installed on this Olares and deliberately says nothing
// about this machine.
type marketInstallState struct {
	// Installed reports that this row's copy occupies the app's namespace
	// here. That is wider than running: a queued, downloading, stopped or
	// failed install has taken the name just as firmly.
	Installed bool `json:"installed"`
	// Status is the provider row's olares_status, and ProviderID the row to
	// watch. Both empty when Installed is false, because they would otherwise
	// be another copy's.
	Status      string `json:"status"`
	ProviderID  string `json:"provider_id"`
	EntranceURL string `json:"entrance_url"`
	// TakenBySource names the source whose copy holds this app name, when it
	// is not this row's. One namespace per app name means this row cannot be
	// installed at all until that copy goes.
	TakenBySource string `json:"taken_by_source"`
	// SourceAmbiguous reports that the install could not be attributed to a
	// source, so it has been attributed to every copy of the name rather than
	// to none. What it says about this row may belong to a sibling.
	SourceAmbiguous bool `json:"source_ambiguous"`
}

// marketActionResponse is what the three lifecycle routes answer with: the
// provider row to watch, named both ways because a route takes the id and the
// application directory keys on the app name.
type marketActionResponse struct {
	ProviderID    string `json:"provider_id"`
	OlaresAppName string `json:"olares_app_name"`
}

// The three lifecycle actions, spelled as the last path segment of their route,
// which is also how they read in a sentence.
const (
	marketActionInstall   = "install"
	marketActionUpgrade   = "upgrade"
	marketActionUninstall = "uninstall"
)

func (a *marketApp) title() string {
	if t := i18nText(a.Title); t != "" {
		return t
	}
	return a.AppName
}

func NewAppCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "model applications that run models on this machine",
		Long: `Install and manage the model applications Router routes to.

A model application packages a model and the engine that serves it. Installing
one through Router, rather than through the Market on its own, is what connects
the two: the provider row appears as the install starts, and its address and
status are filled in once the application is running.

Subcommands:
  catalog              the model applications available to install
  install <app>        install one, creating the provider Router will route to
  upgrade <provider>   upgrade the application behind a provider
  uninstall <provider> remove the application, and with it the provider

Each of the three lifecycle verbs returns as soon as the Market accepts the
request; --watch follows the application instead, which is usually what you
want for an install that takes minutes. There is nothing to catch up on
afterwards, because what is being watched is the provider's own status:
"olares-cli router provider get <name>" reports it at any later moment.

These routes exist only on a Router configured with an Olares Market address. If
they are missing, the model applications on this machine are still usable —
Router discovers them on its own — and "olares-cli market install" installs them.

Admin only.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newAppCatalogCommand(f))
	cmd.AddCommand(newAppInstallCommand(f))
	cmd.AddCommand(newAppUpgradeCommand(f))
	cmd.AddCommand(newAppUninstallCommand(f))
	return cmd
}

func newAppCatalogCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		category string
	)
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "the model applications available to install",
		Long: `List the model applications Router can install.

This is the Olares Market's own view, asked for on your behalf, with what this
machine knows about each row attached — so what is already installed says so,
and what cannot be installed says why.

SERVES is the mode the application's manifest declares: chat, embedding, audio,
ocr, translate. It is the only thing that separates one model application from
another before it is installed, since every one of them is filed under AI.

TAKES says which verb the row accepts. An engine template has no installable
form — "install" refuses it, and "olares-cli market clone" creates an instance
from it, which is also where the model and the engine arguments are chosen.

Examples:
  olares-cli router app catalog
  olares-cli router app catalog --category AI
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runAppCatalog(c.Context(), f, category, output)
		},
	}
	cmd.Flags().StringVar(&category, "category", "", "only this Market category")
	addOutputFlag(cmd, &output)
	return cmd
}

func runAppCatalog(ctx context.Context, f *cmdutil.Factory, category, outputRaw string) error {
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
	apps, err := marketCatalog(ctx, pc, category)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": apps})
	}
	return renderCatalog(os.Stdout, apps)
}

// marketCatalog reads the catalog, sorted so two runs of the same command print
// the same table. The Market answers in its own order, and one app name arrives
// once per source that publishes it.
func marketCatalog(ctx context.Context, pc *preparedClient, category string) ([]marketApp, error) {
	q := url.Values{}
	if c := strings.TrimSpace(category); c != "" {
		q.Set("category", c)
	}
	apps, err := collection[marketApp](ctx, pc, withQuery(epMarketCatalog, q))
	if err != nil {
		return nil, marketRouteErr(err)
	}
	sort.SliceStable(apps, func(i, j int) bool {
		if apps[i].AppName != apps[j].AppName {
			return apps[i].AppName < apps[j].AppName
		}
		return apps[i].Source < apps[j].Source
	})
	return apps, nil
}

func renderCatalog(w io.Writer, items []marketApp) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "the Market offers no model applications to this account.")
		return err
	}
	var anyTemplate, anyTaken, anyAmbiguous bool
	t := newTable(w, "APP", "TITLE", "VERSION", "SERVES", "SOURCE", "STATE", "TAKES")
	for i := range items {
		it := &items[i]
		takes := "install"
		switch {
		case it.TemplateOnly:
			takes = "clone"
			anyTemplate = true
		case it.Install.Installed:
			takes = "upgrade, uninstall"
		case it.Install.TakenBySource != "":
			takes = "-"
		}
		state := "not installed"
		switch {
		case it.Install.Installed:
			state = nonEmpty(it.Install.Status)
			if it.Install.SourceAmbiguous {
				state += " (?)"
				anyAmbiguous = true
			}
		case it.Install.TakenBySource != "":
			state = "taken by " + it.Install.TakenBySource
			anyTaken = true
		}
		t.row(
			nonEmpty(it.AppName), clip(it.title(), 28), nonEmpty(it.Version),
			nonEmpty(it.ModelMode), nonEmpty(it.Source), state, takes)
	}
	if err := t.flush(); err != nil {
		return err
	}
	notes := []string{"APP is the name `app install` takes."}
	if anyTemplate {
		notes = append(notes, "TAKES clone means an engine template: it has no installable form, and "+
			"`olares-cli market clone <app> --title <name>` creates an instance from it, choosing the model there.")
	}
	if anyTaken {
		notes = append(notes, "A row taken by another source cannot be installed: one app name occupies one "+
			"namespace on this machine, and the copy holding it has to go first.")
	}
	if anyAmbiguous {
		notes = append(notes, "A (?) after the state means the install could not be attributed to a source, so "+
			"every copy of that name is reporting it — the status may belong to a sibling row.")
	}
	for _, note := range notes {
		if _, err := fmt.Fprintln(w, "\n"+note); err != nil {
			return err
		}
	}
	return nil
}

func newAppInstallCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		source string
		watch  bool
	)
	cmd := &cobra.Command{
		Use:   "install <app>",
		Short: "install a model application and create its provider",
		Long: `Install a model application, with Router following the install.

Two things happen at once. The Market starts installing the application, and
Router creates the provider row that will route to it — addressed at the
in-cluster shared entrance, marked pending until the application is running.

The provider row survives a failed install. That is deliberate: it is what says
an install was attempted and how it ended. Removing it means uninstalling the
application, which is what "app uninstall" does.

--source picks between sources that publish the same application name. Without
it, a name published once is installed from wherever it comes, and a name
published by several sources is refused rather than guessed at.

This installs a published application as it stands; it chooses nothing. An
engine template is refused, because a template has no installable form and the
model, engine arguments and compute mode are chosen while an instance is made
from it: "olares-cli market clone" is that verb. "app catalog" says which of the
two a row takes.

Examples:
  olares-cli router app install qwen3-8b --watch
  olares-cli router app install qwen3-8b --source market.olares -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppInstall(c.Context(), f, args[0], source, watch, output)
		},
	}
	cmd.Flags().StringVar(&source, "source", "", "the Market source to install from, when several publish the name")
	cmd.Flags().BoolVar(&watch, "watch", false, "follow the install to its end instead of returning immediately")
	addOutputFlag(cmd, &output)
	return cmd
}

func runAppInstall(ctx context.Context, f *cmdutil.Factory, appName, source string, watch bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	appName, err = requireRef(appName, "an application name")
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	// The whole catalog rather than one category: an application is named
	// without saying where the Market files it.
	apps, err := marketCatalog(ctx, pc, "")
	if err != nil {
		return err
	}
	app, err := pickCatalogApp(apps, appName, source)
	if err != nil {
		return err
	}
	if err := refuseUninstallable(app); err != nil {
		return err
	}
	body := map[string]string{"app_name": app.AppName}
	if app.Source != "" {
		body["source"] = app.Source
	}
	var started marketActionResponse
	if err := pc.router.doJSON(ctx, http.MethodPost, epMarketInstall, body, &started); err != nil {
		return marketRouteErr(err)
	}
	return reportAction(ctx, pc, started, marketActionInstall, app.AppName, watch, format)
}

// pickCatalogApp finds the catalog row to install.
//
// One app name arrives once per source that publishes it, and installing is a
// choice between those copies rather than a lookup: they can be different
// versions of different builds. A name published once needs no choice; a name
// published several times and not narrowed by --source is refused, because
// picking for someone here means installing software they did not choose.
func pickCatalogApp(apps []marketApp, appName, source string) (*marketApp, error) {
	byName := make([]*marketApp, 0, 2)
	for i := range apps {
		if strings.EqualFold(apps[i].AppName, appName) {
			byName = append(byName, &apps[i])
		}
	}
	if len(byName) == 0 {
		known := make([]string, 0, len(apps))
		seen := map[string]bool{}
		for i := range apps {
			if name := apps[i].AppName; !seen[name] {
				seen[name] = true
				known = append(known, name)
			}
		}
		return nil, missing{
			noun:  "model application",
			ref:   appName,
			known: known,
			have:  "the Market offers",
			none:  "the Market offers none to this account",
			note:  "`olares-cli router app catalog` lists them with what each one serves.",
		}.err()
	}
	if want := strings.TrimSpace(source); want != "" {
		for _, app := range byName {
			if strings.EqualFold(app.Source, want) {
				return app, nil
			}
		}
		return nil, fmt.Errorf("no source named %q publishes %s; it comes from %s",
			want, appName, strings.Join(sourcesOf(byName), ", "))
	}
	if len(byName) > 1 {
		return nil, fmt.Errorf("%s is published by %s, and only one copy of an application name can be "+
			"installed on this machine; name the one you want with --source",
			appName, strings.Join(sourcesOf(byName), ", "))
	}
	return byName[0], nil
}

// providerIDFromMarket finds a provider id by application name in the Market
// catalog, which is where a row hidden from every other list still appears.
//
// Empty when the name is unknown there, when the copy holding it belongs to
// another source, or when there is no Market to ask — this is a fallback, and
// its failure has to read as "not found" rather than replace the caller's error
// with one about a different route.
func providerIDFromMarket(ctx context.Context, pc *preparedClient, appName string) string {
	apps, err := marketCatalog(ctx, pc, "")
	if err != nil {
		return ""
	}
	for i := range apps {
		if strings.EqualFold(apps[i].AppName, appName) && apps[i].Install.ProviderID != "" {
			return apps[i].Install.ProviderID
		}
	}
	return ""
}

func sourcesOf(apps []*marketApp) []string {
	out := make([]string, 0, len(apps))
	for _, app := range apps {
		out = append(out, nonEmpty(app.Source))
	}
	sort.Strings(out)
	return out
}

// refuseUninstallable stops an install the Market would refuse, or would accept
// and then fail a second or two later.
//
// Both cases are worth catching here rather than after the fact. A template is
// refused upstream with "template apps cannot be installed directly; clone it
// instead", which names the reason but not the way through — which command, that
// it needs a title, and that the model is chosen in the same breath. An app
// whose name is already taken here fails after the request has been accepted,
// leaving a failed install to explain instead of an answer.
func refuseUninstallable(app *marketApp) error {
	if app.TemplateOnly {
		return fmt.Errorf("%s is an engine template rather than an installable application; "+
			"an instance is created from it with `olares-cli market clone %s --title <name> "+
			"--compute-mode nvidia --env ...`, which is where the model is chosen. Only "+
			"--title is enforced there, so a clone missing the rest of the template's "+
			"published environment is created and then fails to serve: MODEL_SOURCE, "+
			"MODEL_NAME, MODEL_MODE, MODEL_SUPPORTS, ENGINE_ARGS and the engine's own "+
			"<ENGINE>_REQUIRED_GPU_MEMORY all belong on that command, and the olares-chart "+
			"skill's LLM model workflow gives the per-engine values. Router picks the "+
			"instance up once it is running, and `olares-cli router local spec set` is what "+
			"changes what it serves afterwards",
			app.AppName, app.AppName)
	}
	if taken := app.Install.TakenBySource; taken != "" {
		return fmt.Errorf("the name %s is already taken on this machine by the copy from %s, and one "+
			"application name occupies one namespace; that copy has to be uninstalled before this "+
			"one can be installed. `olares-cli router app catalog` shows both",
			app.AppName, taken)
	}
	if app.Install.Installed {
		status := nonEmpty(app.Install.Status)
		return fmt.Errorf("%s is already installed on this machine and is %s; "+
			"`olares-cli router app upgrade %s` moves it to a newer version, "+
			"`olares-cli router app uninstall %s` removes it, and "+
			"`olares-cli market resume %s` starts it if it is stopped",
			app.AppName, status, app.AppName, app.AppName, app.AppName)
	}
	return nil
}

func newAppUpgradeCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		watch  bool
	)
	cmd := &cobra.Command{
		Use:   "upgrade <provider>",
		Short: "upgrade the application behind a provider",
		Long: `Upgrade the model application a provider routes to.

The provider is named rather than the application, because the provider is what
Router knows: it holds the application name and is what the upgrade reconciles
when it finishes. The application's own name identifies the provider too, since
every locally installed one answers to the same routing name.

A provider an admin entered by hand has no application behind it and cannot be
upgraded. Its models come from an upstream somebody else operates.

Examples:
  olares-cli router app upgrade qwen3-8b --watch
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppLifecycle(c.Context(), f, args[0], marketActionUpgrade, false, watch, output)
		},
	}
	cmd.Flags().BoolVar(&watch, "watch", false, "follow the upgrade to its end instead of returning immediately")
	addOutputFlag(cmd, &output)
	return cmd
}

func newAppUninstallCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output    string
		watch     bool
		assumeYes bool
	)
	cmd := &cobra.Command{
		Use:   "uninstall <provider>",
		Short: "remove the application behind a provider",
		Long: `Uninstall the model application a provider routes to.

The application goes, and the provider with it — including the models attached
to it, any route that pointed at one of them, and the keys' permission to call
them. A model downloaded by the application is part of the application and does
not survive it.

Anything still calling this provider starts failing as soon as the application
stops. Check with "olares-cli router usage" before removing something a
workspace depends on.

Confirmation is required. --yes skips the prompt and is mandatory when stdin is
not a terminal.

Examples:
  olares-cli router app uninstall qwen3-8b --yes --watch
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppLifecycle(c.Context(), f, args[0], marketActionUninstall, !assumeYes, watch, output)
		},
	}
	cmd.Flags().BoolVar(&watch, "watch", false, "follow the uninstall to its end instead of returning immediately")
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt (required when stdin is not a terminal)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runAppLifecycle(ctx context.Context, f *cmdutil.Factory, ref, action string, confirm, watch bool, outputRaw string) error {
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
	found, err := resolveProvider(ctx, pc, ref)
	if err != nil {
		return err
	}
	if !found.isMarketSourced() {
		return fmt.Errorf("%s was entered by hand and has no application behind it, so there is nothing to %s; "+
			"`olares-cli router provider delete %s` removes the provider itself",
			found.handle(), action, found.handle())
	}
	app := "its application"
	if found.OlaresAppName != nil && *found.OlaresAppName != "" {
		app = *found.OlaresAppName
	}
	if confirm {
		if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
			fmt.Sprintf("Uninstall %q, removing its provider along with its models, the routes pointing at them, and the models the application downloaded?",
				app),
			false); err != nil {
			return err
		}
	}
	var started marketActionResponse
	if err := pc.router.doJSON(ctx, http.MethodPost, epMarketProviderAction(found.ID, action), nil, &started); err != nil {
		return marketRouteErr(err)
	}
	return reportAction(ctx, pc, started, action, app, watch, format)
}

// reportAction says what was started, and follows it when asked to.
func reportAction(ctx context.Context, pc *preparedClient, started marketActionResponse, action, what string, watch bool, format Format) error {
	if !watch {
		if format == FormatJSON {
			return printJSON(os.Stdout, started)
		}
		if _, err := fmt.Fprintf(os.Stdout, "%s of %s accepted; its provider is %s\n",
			action, nonEmpty(what), started.ProviderID); err != nil {
			return err
		}
		_, err := fmt.Fprintf(os.Stdout, "`olares-cli router provider get %s` says how far it has got, "+
			"and --watch on this command would have followed it here.\n", started.ProviderID)
		return err
	}
	if format == FormatTable {
		if _, err := fmt.Fprintf(os.Stdout, "%s of %s accepted; watching its provider %s. "+
			"Interrupting stops the watch, not the %s.\n",
			action, nonEmpty(what), started.ProviderID, action); err != nil {
			return err
		}
	}
	return watchAction(ctx, pc, started.ProviderID, action, what, format)
}

// marketLifecycle is what "still working", "done" and "failed" mean for one
// action.
//
// Ported from the console's own table (frontend/src/lib/marketLifecycle.ts in
// the router repo), because the two read the same column and would otherwise
// disagree about whether the same application has finished. The table is keyed
// by action rather than being one list of terminal states, since the same
// status means opposite things depending on what was asked: `unreachable` is a
// successful uninstall and a failed install.
type marketLifecycle struct {
	success []string
	failed  []string
	// departsFrom is a status the application is expected to be sitting in
	// when the action starts, and must be seen leaving before success can be
	// believed. Only upgrade needs it: the app is already running when the
	// request goes out, so the first poll would otherwise read success off the
	// state being replaced.
	departsFrom string
	// tracksModel makes model_console_status part of the verdict. It is, for
	// the two actions that end with an application expected to serve a model,
	// and is not for an uninstall — there the phase is a leftover from the app
	// being removed, and reading it would hold a finished uninstall open.
	tracksModel bool
}

var marketLifecycles = map[string]marketLifecycle{
	marketActionInstall: {
		success:     []string{"running"},
		failed:      []string{"failed", "unreachable"},
		tracksModel: true,
	},
	marketActionUpgrade: {
		success:     []string{"running"},
		failed:      []string{"failed", "unreachable"},
		departsFrom: "running",
		tracksModel: true,
	},
	// A finished uninstall reports `unreachable`, which the Market means as
	// "gone" and every other action means as "lost". `archived` is accepted
	// too: nothing writes it today, but it is the honest name for this end
	// state.
	marketActionUninstall: {
		success: []string{"unreachable", "archived"},
		failed:  []string{"failed"},
	},
}

func lifecycleFor(action string) marketLifecycle {
	if lc, ok := marketLifecycles[action]; ok {
		return lc
	}
	return marketLifecycles[marketActionInstall]
}

// Phases of model_console_status that mean the application is still fetching or
// loading what it serves, and the one that means it could not.
var modelPhasesInFlight = []string{"init", "download", "loading"}

const modelPhaseFailed = "failed"

// verdict decides whether the action is finished, and how. Empty while it is
// still running — including when there is no status yet, which is what the
// first poll of a freshly created row sees.
func (lc marketLifecycle) verdict(appStatus, modelPhase string, sawDeparture bool, elapsed time.Duration) string {
	if appStatus == "" {
		return ""
	}
	if contains(lc.failed, appStatus) {
		return "failed"
	}
	if lc.tracksModel && modelPhase == modelPhaseFailed {
		return "failed"
	}
	if !contains(lc.success, appStatus) {
		return ""
	}
	if lc.departsFrom == appStatus && !sawDeparture && elapsed < marketDepartureGrace {
		return ""
	}
	// A running application is not finished while it is still fetching the
	// weights it serves. One reporting no phase at all has nothing to wait for.
	if lc.tracksModel && contains(modelPhasesInFlight, modelPhase) {
		return ""
	}
	return "done"
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

const (
	// marketPollInterval is how often the provider row is re-read. The
	// application directory polls the Market faster while anything is in
	// flight, so this is the interval at which the row can actually change.
	marketPollInterval = 2 * time.Second
	// marketDepartureGrace is how long to wait for an upgrade to leave
	// `running` before believing it finished. A short upgrade can complete
	// between the request and the first poll, leaving no departure to observe,
	// and a watch that never ends on a finished upgrade is worse than one that
	// reports it a few seconds late. Same value as the console's.
	marketDepartureGrace = 10 * time.Second
)

// watchFrame is one observed change to the row an action is moving.
type watchFrame struct {
	ProviderID   string `json:"provider_id"`
	OlaresStatus string `json:"olares_status"`
	ModelPhase   string `json:"model_console_status,omitempty"`
	// Outcome is set on the last frame only: `done` or `failed`.
	Outcome string `json:"outcome,omitempty"`
}

// watchAction polls the provider row until the action reaches its end.
//
// The row is read directly rather than through the memoized collection: nothing
// is written while a watch waits, so a cached answer would be the same on every
// turn of the loop and the watch would never end.
func watchAction(ctx context.Context, pc *preparedClient, providerID, action, what string, format Format) error {
	lc := lifecycleFor(action)
	started := time.Now()
	var (
		lastApp, lastPhase string
		sawDeparture       bool
		first              = true
	)
	for {
		detail, err := getProvider(ctx, pc, providerID)
		if err != nil {
			var re *RouterError
			if errors.As(err, &re) && re.Status == http.StatusNotFound {
				// Nothing deletes the row today; an uninstall leaves it
				// behind reporting `unreachable`. If one ever does, that is
				// the uninstall having succeeded and any other action having
				// lost what it was moving.
				if action == marketActionUninstall {
					return reportVerdict(os.Stdout, providerID, action, what, "", "", "done", format)
				}
				return fmt.Errorf("the provider Router created for this %s no longer exists, so there is "+
					"nothing left to watch: %w", action, err)
			}
			return err
		}
		appStatus := strings.TrimSpace(strDeref(detail.OlaresStatus))
		phase := strings.TrimSpace(strDeref(detail.ModelConsoleStatus))
		if lc.departsFrom != "" && appStatus != "" && appStatus != lc.departsFrom {
			sawDeparture = true
		}
		verdict := lc.verdict(appStatus, phase, sawDeparture, time.Since(started))
		if verdict != "" {
			return reportVerdict(os.Stdout, providerID, action, what, appStatus, phase, verdict, format)
		}
		if first || appStatus != lastApp || phase != lastPhase {
			if err := printFrame(os.Stdout, watchFrame{
				ProviderID:   providerID,
				OlaresStatus: appStatus,
				ModelPhase:   phase,
			}, format); err != nil {
				return err
			}
		}
		lastApp, lastPhase, first = appStatus, phase, false

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(marketPollInterval):
		}
	}
}

func printFrame(w io.Writer, frame watchFrame, format Format) error {
	if format == FormatJSON {
		return printJSON(w, frame)
	}
	line := "  " + nonEmpty(frame.OlaresStatus)
	if frame.ModelPhase != "" {
		line += ", model " + frame.ModelPhase
	}
	_, err := fmt.Fprintln(w, line)
	return err
}

func reportVerdict(w io.Writer, providerID, action, what, appStatus, phase, verdict string, format Format) error {
	if format == FormatJSON {
		return printJSON(w, watchFrame{
			ProviderID:   providerID,
			OlaresStatus: appStatus,
			ModelPhase:   phase,
			Outcome:      verdict,
		})
	}
	if verdict == "done" {
		_, err := fmt.Fprintf(w, "%s of %s finished; the application is %s.\n",
			action, nonEmpty(what), nonEmpty(appStatus))
		return err
	}
	detail := nonEmpty(appStatus)
	if phase == modelPhaseFailed {
		detail += ", and the model it serves could not be loaded"
	}
	_, err := fmt.Fprintf(w, "%s of %s failed: the application is %s. "+
		"`olares-cli market status %s` says what the Market recorded, and "+
		"`olares-cli router local status` reports what the application itself says.\n",
		action, nonEmpty(what), detail, nonEmpty(what))
	return err
}

func strDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// marketRouteErr explains a missing route rather than reporting it as a bare
// 404. Router mounts the whole Market subtree only when it has been given a
// Market address, so absence is a configuration fact about this Router and not
// something the caller did wrong.
func marketRouteErr(err error) error {
	if err == nil {
		return nil
	}
	var re *RouterError
	if errors.As(err, &re) && re.Status == http.StatusNotFound && re.Code == "" {
		return fmt.Errorf("this Router has no Olares Market configured, so it cannot install applications itself. " +
			"The model applications already on this machine still work — Router finds them on its own — and " +
			"`olares-cli market install <app>` installs new ones")
	}
	return err
}
