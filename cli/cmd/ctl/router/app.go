package router

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliutil"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli router app …` — model applications, installed through Router.
//
// GET  /console/api/providers/market-catalog
// POST /console/api/providers/market-install
// POST /console/api/providers/:id/market-upgrade
// POST /console/api/providers/:id/market-uninstall
// GET  /console/api/providers/:id/install-tasks
// GET  /console/api/providers/:id/install-events   (SSE)
//
// A model application runs a model on this machine. Installing one through
// Router rather than through the Market directly is what ties the two together:
// Router creates the provider row as the install starts, follows the Market
// task, and fills in the address and status when it finishes. Installing the
// same app with `olares-cli market install` leaves Router to discover it on its
// own polling cycle, which works but tells you nothing while it happens.
//
// The lifecycle is asynchronous everywhere. Every verb here returns a task, and
// the task is the thing to watch.

// defaultMarketSource is the catalog the official engine templates are
// published in, and the source `olares-cli market` also defaults to.
const defaultMarketSource = "market.olares"

type marketApp struct {
	AppName     string     `json:"app_name"`
	VersionName string     `json:"version_name"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	IconURL     string     `json:"icon_url"`
	Category    string     `json:"category"`
	State       string     `json:"state"`
	InstalledAt *time.Time `json:"installed_at,omitempty"`
}

type installTask struct {
	ID              int64      `json:"id"`
	MarketTaskID    string     `json:"market_task_id"`
	ProviderID      string     `json:"provider_id"`
	Action          string     `json:"action"`
	Status          string     `json:"status"`
	StartedByUserID *string    `json:"started_by_user_id,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
}

// startedTask is what every lifecycle verb answers with. The task id is the
// handle for `app watch`; the provider id is the row that now exists whether or
// not the install ends up succeeding.
type startedTask struct {
	ProviderID string `json:"provider_id"`
	TaskID     int64  `json:"task_id"`
	SSEURL     string `json:"sse_url"`
}

func NewAppCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "model applications that run models on this machine",
		Long: `Install and manage the model applications Router routes to.

A model application packages a model and the engine that serves it. Installing
one through Router, rather than through the Market on its own, is what connects
the two: the provider row appears as the install starts, and its address and
status are filled in when the Market says the app is running.

Every verb here is asynchronous and answers with a task id. "app watch" follows
one to its end.

Subcommands:
  catalog             the model applications available to install
  install <app>       install one, creating the provider Router will route to
  upgrade <provider>  upgrade the application behind a provider
  uninstall <provider> remove the application, and with it the provider
  tasks <provider>    what has been installed, upgraded or removed, and when
  watch <provider>    follow a task as it runs

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
	cmd.AddCommand(newAppTasksCommand(f))
	cmd.AddCommand(newAppWatchCommand(f))
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

This is the Olares Market's own view, asked for on your behalf, so what you see
here is what the Market offers this account — including applications already
installed, whose state says so.

--category narrows the list the way the Market does; without it every category
comes back.

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
	q := url.Values{}
	if c := strings.TrimSpace(category); c != "" {
		q.Set("category", c)
	}
	apps, err := collection[marketApp](ctx, pc, withQuery(epMarketCatalog, q))
	if err != nil {
		return marketRouteErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"items": apps})
	}
	names := make([]string, 0, len(apps))
	for i := range apps {
		names = append(names, apps[i].AppName)
	}
	return renderCatalog(os.Stdout, apps, marketTemplateApps(ctx, pc, names))
}

// renderCatalog prints the catalog with the verb each row actually takes.
//
// Router's catalog does not distinguish an engine template from an installable
// application, and `app install` refuses the former — so the distinction is
// resolved against the Market and shown here, where the choice is made.
func renderCatalog(w io.Writer, items []marketApp, templates map[string]bool) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(w, "the Market offers no model applications for this account.")
		return err
	}
	anyTemplate, anyUnknown := false, false
	t := newTable(w, "APP", "TITLE", "VERSION", "TAKES", "WHAT IT SERVES")
	for i := range items {
		it := &items[i]
		takes := "install"
		if known, ok := templates[it.AppName]; !ok {
			takes = "-"
			anyUnknown = true
		} else if known {
			takes = "clone"
			anyTemplate = true
		}
		// State is empty against a real Market, so it is not a column. The
		// installed ones are known from the provider list instead.
		t.row(
			nonEmpty(it.AppName), clip(nonEmpty(it.Title), 30), nonEmpty(it.VersionName),
			takes, clip(nonEmpty(it.Description), 52))
	}
	if err := t.flush(); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "\nAPP is the name `app install` takes. `olares-cli router provider list` shows which are already installed."); err != nil {
		return err
	}
	if anyTemplate {
		if _, err := fmt.Fprintln(w, "TAKES clone means an engine template: it has no installable form, and "+
			"`olares-cli market clone <app> --title <name>` creates an instance from it, choosing the model there."); err != nil {
			return err
		}
	}
	if anyUnknown {
		if _, err := fmt.Fprintln(w, "A TAKES of - means the Market did not answer for that row; `app install` will find out."); err != nil {
			return err
		}
	}
	return nil
}

func newAppInstallCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		watch  bool
	)
	cmd := &cobra.Command{
		Use:   "install <app>",
		Short: "install a model application and create its provider",
		Long: `Install a model application, with Router following the install.

Two things happen at once. The Market starts installing the application, and
Router creates the provider row that will route to it — addressed at the
in-cluster shared entrance, marked pending until the Market reports the app
running.

The provider row survives a failed install. That is deliberate: it carries the
task history that explains what went wrong. Removing it means uninstalling the
application, which is what "app uninstall" does.

The command returns as soon as the task is accepted. Pass --watch to follow it
instead, which is usually what you want for an install that takes minutes.

This installs a published application as it stands; it chooses nothing. An
engine template is refused, because a template has no installable form and the
model, engine arguments and compute mode are chosen while an instance is made
from it: "olares-cli market clone" is that verb. "app catalog" says which of the
two a row takes.

Examples:
  olares-cli router app install qwen3-8b --watch
  olares-cli router app install qwen3-8b -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppInstall(c.Context(), f, args[0], watch, output)
		},
	}
	cmd.Flags().BoolVar(&watch, "watch", false, "follow the install to its end instead of returning immediately")
	addOutputFlag(cmd, &output)
	return cmd
}

func runAppInstall(ctx context.Context, f *cmdutil.Factory, appName string, watch bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	appName = strings.TrimSpace(appName)
	if appName == "" {
		return fmt.Errorf("application name is required")
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	// Installing what is already installed is accepted by Router and then fails
	// in the Market a second or two later, leaving a failed task on the existing
	// provider. Saying so up front is the difference between an answer and a
	// piece of history to explain.
	if err := refuseReinstall(ctx, pc, appName); err != nil {
		return err
	}
	if err := refuseTemplateInstall(ctx, pc, appName); err != nil {
		return err
	}
	var started startedTask
	body := map[string]string{"app_name": appName}
	if err := pc.router.doJSON(ctx, "POST", epMarketInstall, body, &started); err != nil {
		return marketRouteErr(err)
	}
	return reportStarted(ctx, pc, started, "installing "+appName, watch, format)
}

// refuseReinstall stops an install of something already on this machine.
//
// Olares is asked rather than Router, because Router's provider list omits an
// application that is not running and would answer "not installed" for a
// stopped model app — the exact case where reinstalling looks reasonable and is
// rejected by the Market two seconds later. When a provider does exist the
// message names it, since that is the handle upgrade and uninstall take.
func refuseReinstall(ctx context.Context, pc *preparedClient, appName string) error {
	installed, err := listMyApps(ctx, pc.desktop)
	if err != nil {
		return nil
	}
	var app *myApp
	for i := range installed {
		if strings.EqualFold(strings.TrimSpace(installed[i].Name), appName) {
			app = &installed[i]
			break
		}
	}
	if app == nil {
		return nil
	}
	handle := appName
	if p := providerForApp(ctx, pc, appName); p != nil {
		handle = p.Name
	}
	state := strings.TrimSpace(app.State)
	if state == "" {
		state = "installed"
	}
	return fmt.Errorf("%s is already installed on this machine and is %s; "+
		"`olares-cli router app upgrade %q` moves it to a newer version, "+
		"`olares-cli router app uninstall %q` removes it, and "+
		"`olares-cli market resume %s` starts it if it is stopped",
		appName, state, handle, handle, appName)
}

// refuseTemplateInstall stops an install of an engine template.
//
// The Market will not install a template at all: a template body has no
// installable form, and an instance is made from it by cloning, which is also
// where the model, the engine arguments and the compute mode are chosen.
// Router's catalog offers them anyway — it filters on the manifest's
// LLMGatewaySupported alone, and the official engine bases carry templateOnly
// as well. Left to itself the install is accepted here and refused by the
// Market a moment later with "template apps cannot be installed directly;
// clone it instead", which Router passes through verbatim. That names the
// reason but not the way through: which command, that it needs a title, and
// that the model is chosen in the same breath.
func refuseTemplateInstall(ctx context.Context, pc *preparedClient, appName string) error {
	if !marketTemplateApps(ctx, pc, []string{appName})[appName] {
		return nil
	}
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
		appName, appName)
}

// marketTemplateApps reports which of the named applications the Market holds
// as template bodies, keyed by application name.
//
// The Market is asked directly because that answer is the same on every
// Router: only recent ones report the flag in their own catalog, and one
// request here beats a version check plus two code paths. An application
// missing from the answer is absent from the map rather
// than recorded as installable, and a failed request yields nothing at all:
// this drives a refusal and a column, so not knowing must read as not knowing
// instead of as permission. What each caller then does with "not known" differs
// on purpose: the column prints - and says the install will find out, and the
// install itself goes ahead, because a Market that cannot be reached must not
// stand between someone and a pinned model the Market would have accepted. The
// Market is still the one that refuses a template; this only says so earlier.
//
// Only the default catalog is consulted, which is where the official engine
// templates are published. The flag is the manifest's options.templateOnly,
// surfaced on app_simple_info — the same field `olares-cli market clone` reads.
func marketTemplateApps(ctx context.Context, pc *preparedClient, appNames []string) map[string]bool {
	out := map[string]bool{}
	if pc == nil || pc.profile == nil || strings.TrimSpace(pc.profile.MarketURL) == "" || len(appNames) == 0 {
		return out
	}
	query := make([]map[string]string, 0, len(appNames))
	for _, name := range appNames {
		if n := strings.TrimSpace(name); n != "" {
			query = append(query, map[string]string{"appid": n, "sourceDataName": defaultMarketSource})
		}
	}
	if len(query) == 0 {
		return out
	}
	var env struct {
		Data struct {
			Apps []struct {
				SimpleInfo struct {
					AppName      string `json:"app_name"`
					TemplateOnly bool   `json:"templateOnly"`
				} `json:"app_simple_info"`
			} `json:"apps"`
		} `json:"data"`
	}
	market := whoami.NewHTTPClient(pc.hc, pc.profile.MarketURL, pc.profile.OlaresID)
	if err := market.DoJSON(ctx, "POST", "/app-store/api/v2/apps", map[string]any{"apps": query}, &env); err != nil {
		return out
	}
	for _, app := range env.Data.Apps {
		if name := strings.TrimSpace(app.SimpleInfo.AppName); name != "" {
			out[name] = app.SimpleInfo.TemplateOnly
		}
	}
	return out
}

// providerForApp finds the provider Router keeps for a Market application. Nil
// when there is none, and also when the application's provider is hidden
// because the application is not running.
func providerForApp(ctx context.Context, pc *preparedClient, appName string) *providerRow {
	rows, err := listProviders(ctx, pc)
	if err != nil {
		return nil
	}
	for i := range rows {
		if rows[i].OlaresAppName != nil && strings.EqualFold(*rows[i].OlaresAppName, appName) {
			return &rows[i]
		}
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
when it finishes.

A provider an admin entered by hand has no application behind it and cannot be
upgraded. Its models come from an upstream somebody else operates.

Examples:
  olares-cli router app upgrade "Qwen3.6-27B (llama.cpp)" --watch
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppLifecycle(c.Context(), f, args[0], "market-upgrade", "upgrading", false, watch, output)
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
to it, any default that pointed at one of them, and the keys' permission to call
them. A model downloaded by the application is part of the application and does
not survive it.

Anything still calling this provider starts failing as soon as the application
stops. Check with "olares-cli router usage" before removing something a
workspace depends on.

Confirmation is required. --yes skips the prompt and is mandatory when stdin is
not a terminal.

Examples:
  olares-cli router app uninstall "Qwen3.6-27B (llama.cpp)" --yes --watch
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppLifecycle(c.Context(), f, args[0], "market-uninstall", "uninstalling", !assumeYes, watch, output)
		},
	}
	cmd.Flags().BoolVar(&watch, "watch", false, "follow the uninstall to its end instead of returning immediately")
	cmd.Flags().BoolVar(&assumeYes, "yes", false, "skip the confirmation prompt (required when stdin is not a terminal)")
	addOutputFlag(cmd, &output)
	return cmd
}

func runAppLifecycle(ctx context.Context, f *cmdutil.Factory, ref, action, gerund string, confirm, watch bool, outputRaw string) error {
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
			found.Name, strings.TrimSuffix(gerund, "ing"), found.Name)
	}
	app := "its application"
	if found.OlaresAppName != nil && *found.OlaresAppName != "" {
		app = *found.OlaresAppName
	}
	if confirm {
		if err := cliutil.ConfirmDestructive(os.Stderr, os.Stdin,
			fmt.Sprintf("Uninstall %q, removing provider %q along with its models, the defaults pointing at them, and the models downloaded by the application?",
				app, found.Name),
			false); err != nil {
			return err
		}
	}
	var started startedTask
	path := epMarketProviderAction(found.ID, action)
	if err := pc.router.doJSON(ctx, "POST", path, nil, &started); err != nil {
		return marketRouteErr(err)
	}
	return reportStarted(ctx, pc, started, gerund+" "+app, watch, format)
}

func reportStarted(ctx context.Context, pc *preparedClient, started startedTask, what string, watch bool, format Format) error {
	if format == FormatJSON && !watch {
		return printJSON(os.Stdout, started)
	}
	if _, err := fmt.Fprintf(os.Stdout, "%s: task %d on provider %s\n", what, started.TaskID, started.ProviderID); err != nil {
		return err
	}
	if !watch {
		_, err := fmt.Fprintf(os.Stdout, "follow it with `olares-cli router app watch %s --task %d`\n",
			started.ProviderID, started.TaskID)
		return err
	}
	return followTask(ctx, pc, started.ProviderID, started.TaskID, 0, format)
}

func newAppTasksCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		limit  int
		offset int
	)
	cmd := &cobra.Command{
		Use:   "tasks <provider>",
		Short: "what has been installed, upgraded or removed on a provider",
		Long: `List the lifecycle tasks recorded against a provider, newest first.

This is where a failed install explains itself: the row keeps the error the
Market reported, and "app watch" replays the whole event stream of any task,
including one that finished long ago.

Examples:
  olares-cli router app tasks "Qwen3.6-27B (llama.cpp)"
  olares-cli router app tasks <provider-id> -o json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppTasks(c.Context(), f, args[0], limit, offset, output)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "how many tasks to return (1-200)")
	cmd.Flags().IntVar(&offset, "offset", 0, "how many tasks to skip")
	addOutputFlag(cmd, &output)
	return cmd
}

func runAppTasks(ctx context.Context, f *cmdutil.Factory, ref string, limit, offset int, outputRaw string) error {
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
	env, err := fetchInstallTasks(ctx, pc, found.ID, limit, offset)
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, env)
	}
	return renderInstallTasks(ctx, pc, os.Stdout, found, env.Items, env.Total, env.Offset)
}

func renderInstallTasks(ctx context.Context, pc *preparedClient, w io.Writer, p *providerRow, items []installTask, total, offset int) error {
	if len(items) == 0 {
		_, err := fmt.Fprintf(w, "%s has no lifecycle tasks. Router registered it from a running application "+
			"rather than installing it, which leaves no task history.\n", p.Name)
		return err
	}
	users := userLabels(ctx, pc)
	t := newTable(w, "TASK", "ACTION", "STATUS", "STARTED", "FINISHED", "BY", "ERROR")
	for i := range items {
		it := &items[i]
		by := "-"
		if it.StartedByUserID != nil {
			by = *it.StartedByUserID
			if label, ok := users[by]; ok && label != "" {
				by = label
			}
		}
		finished := "-"
		if it.FinishedAt != nil {
			finished = it.FinishedAt.Local().Format("2006-01-02 15:04:05")
		}
		errText := "-"
		if it.ErrorMessage != nil && *it.ErrorMessage != "" {
			errText = clip(*it.ErrorMessage, 48)
		}
		t.row(
			strconv.FormatInt(it.ID, 10), nonEmpty(it.Action), nonEmpty(it.Status),
			it.StartedAt.Local().Format("2006-01-02 15:04:05"), finished, by, errText)
	}
	if err := t.flush(); err != nil {
		return err
	}
	return pageFooter(w, len(items), total, offset)
}

func newAppWatchCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		taskID int64
		since  int64
	)
	cmd := &cobra.Command{
		Use:   "watch <provider>",
		Short: "follow a lifecycle task as it runs",
		Long: `Follow an install, upgrade or uninstall to its end.

Router replays a task's recorded events, then delivers the live ones as they
happen, and closes the stream when the task reaches its end. The task's own row
is printed before and after, so a task that finished long ago still reports what
it did even when no events were kept for it.

Without --task the newest task on the provider is followed, which is the one you
just started.

--since replays only events after a point, for picking up where an interrupted
watch left off.

Examples:
  olares-cli router app watch "Qwen3.6-27B (llama.cpp)"
  olares-cli router app watch <provider-id> --task 42
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runAppWatch(c.Context(), f, args[0], taskID, since, output)
		},
	}
	cmd.Flags().Int64Var(&taskID, "task", 0, "the task to follow; the newest one by default")
	cmd.Flags().Int64Var(&since, "since", 0, "replay only events after this event id")
	addOutputFlag(cmd, &output)
	return cmd
}

func runAppWatch(ctx context.Context, f *cmdutil.Factory, ref string, taskID, since int64, outputRaw string) error {
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
	if taskID <= 0 {
		latest, err := fetchInstallTasks(ctx, pc, found.ID, 1, 0)
		if err != nil {
			return err
		}
		if len(latest.Items) == 0 {
			return fmt.Errorf("%s has no lifecycle tasks to follow; Router registered it from a running "+
				"application rather than installing it", found.Name)
		}
		taskID = latest.Items[0].ID
	}
	return followTask(ctx, pc, found.ID, taskID, since, format)
}

func fetchInstallTasks(ctx context.Context, pc *preparedClient, providerID string, limit, offset int) (*page[installTask], error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	var env page[installTask]
	if err := pc.router.doJSON(ctx, "GET", withQuery(epMarketInstallTasks(providerID), q), nil, &env); err != nil {
		return nil, marketRouteErr(err)
	}
	return &env, nil
}

// findTask reads one task's row. Nil when Router does not have it, which the
// callers treat as "nothing more to say about it" rather than as a failure —
// this is context around a stream, not the stream itself.
func findTask(ctx context.Context, pc *preparedClient, providerID string, taskID int64) *installTask {
	tasks, err := fetchInstallTasks(ctx, pc, providerID, 200, 0)
	if err != nil {
		return nil
	}
	for i := range tasks.Items {
		if tasks.Items[i].ID == taskID {
			return &tasks.Items[i]
		}
	}
	return nil
}

// followTask brackets the event stream with the task's own row.
//
// Both halves are needed. Router closes the stream immediately for a task that
// has already finished, and it sends no frame at all for one whose events were
// never recorded — a bare stream would then print nothing and read as a broken
// connection. And a stream that ends mid-install because the connection dropped
// looks exactly like one that ended because the install did, so the outcome is
// read back from the row rather than inferred from the last frame.
func followTask(ctx context.Context, pc *preparedClient, providerID string, taskID, since int64, format Format) error {
	before := findTask(ctx, pc, providerID, taskID)
	if format == FormatTable && before != nil {
		if _, err := fmt.Fprintf(os.Stdout, "task %d: %s, %s since %s\n",
			before.ID, nonEmpty(before.Action), nonEmpty(before.Status),
			before.StartedAt.Local().Format("2006-01-02 15:04:05")); err != nil {
			return err
		}
	}
	if err := streamInstallEvents(ctx, pc, providerID, taskID, since, format); err != nil {
		return err
	}
	after := findTask(ctx, pc, providerID, taskID)
	if after == nil {
		if before == nil && format == FormatTable {
			_, err := fmt.Fprintf(os.Stdout, "Router has no task %d on this provider.\n", taskID)
			return err
		}
		return nil
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, map[string]any{"task": after})
	}
	return reportOutcome(os.Stdout, after)
}

func reportOutcome(w io.Writer, t *installTask) error {
	switch t.Status {
	case "succeeded", "success", "completed":
		_, err := fmt.Fprintf(w, "%s succeeded.\n", nonEmpty(t.Action))
		return err
	case "failed", "error":
		reason := "no reason recorded"
		if t.ErrorMessage != nil && *t.ErrorMessage != "" {
			reason = *t.ErrorMessage
		}
		_, err := fmt.Fprintf(w, "%s failed: %s\n", nonEmpty(t.Action), reason)
		return err
	default:
		// The stream ended while the task is still open, which means the
		// connection went rather than the task.
		_, err := fmt.Fprintf(w, "the event stream ended while this %s is still %s; "+
			"`olares-cli router app watch %s --task %d` picks it back up\n",
			nonEmpty(t.Action), nonEmpty(t.Status), t.ProviderID, t.ID)
		return err
	}
}

// sseEvent is one frame of Router's install stream. Router replays a finished
// task's frames before closing, so a stream that ends immediately is a task that
// has already reached its end rather than a stream that failed to open.
type sseEvent struct {
	ID      int64           `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func streamInstallEvents(ctx context.Context, pc *preparedClient, providerID string, taskID, since int64, format Format) error {
	q := url.Values{"task_id": {strconv.FormatInt(taskID, 10)}}
	if since > 0 {
		q.Set("last_event_id", strconv.FormatInt(since, 10))
	}
	resp, err := pc.router.doStream(ctx, withQuery(epMarketInstallEvents(providerID), q))
	if err != nil {
		return marketRouteErr(err)
	}
	defer func() { _ = resp.Body.Close() }()

	return readSSE(resp.Body, func(ev sseEvent) error {
		if format == FormatJSON {
			// Tagged, because the closing task row goes to the same stream and
			// carries an "id" of its own.
			return printJSON(os.Stdout, map[string]any{"event": ev})
		}
		line := describeInstallEvent(ev)
		if line == "" {
			return nil
		}
		_, werr := fmt.Fprintln(os.Stdout, line)
		return werr
	})
}

// readSSE walks the frames of a text/event-stream. Only `id`, `event` and
// `data` matter here; a comment line is the keepalive Router sends to hold an
// idle connection open and carries nothing.
//
// A frame at end of stream without its blank-line terminator is still delivered.
// Router closes the connection the moment a task reaches its end, and dropping
// that last frame would lose the outcome — the one line worth reading.
func readSSE(r io.Reader, onEvent func(sseEvent) error) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var (
		ev  sseEvent
		got bool
	)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		switch {
		case line == "":
			if got {
				if err := onEvent(ev); err != nil {
					return err
				}
			}
			ev, got = sseEvent{}, false
		case strings.HasPrefix(line, ":"):
			// keepalive
		case strings.HasPrefix(line, "id:"):
			if n, err := strconv.ParseInt(strings.TrimSpace(line[3:]), 10, 64); err == nil {
				ev.ID = n
				got = true
			}
		case strings.HasPrefix(line, "event:"):
			ev.Type = strings.TrimSpace(line[6:])
			got = true
		case strings.HasPrefix(line, "data:"):
			ev.Payload = json.RawMessage(strings.TrimSpace(line[5:]))
			got = true
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("read install events: %w", err)
	}
	if got {
		return onEvent(ev)
	}
	return nil
}

// describeInstallEvent turns one frame into a line. An unknown event type is
// printed as itself rather than dropped: a stream that silently omits frames
// misrepresents what happened.
func describeInstallEvent(ev sseEvent) string {
	var body struct {
		Percent      *int   `json:"percent"`
		Message      string `json:"message"`
		State        string `json:"state"`
		OlaresStatus string `json:"olares_status"`
		ProviderID   string `json:"provider_id"`
		Retriable    *bool  `json:"retriable"`
	}
	_ = json.Unmarshal(ev.Payload, &body)

	switch ev.Type {
	case "progress":
		pct := "?"
		if body.Percent != nil {
			pct = strconv.Itoa(*body.Percent) + "%"
		}
		if body.Message != "" {
			return fmt.Sprintf("  %-5s %s", pct, body.Message)
		}
		return "  " + pct
	case "state":
		return "  now " + nonEmpty(body.State)
	case "done":
		if body.OlaresStatus != "" {
			return fmt.Sprintf("done; the application is %s and its provider is ready", body.OlaresStatus)
		}
		return "done"
	case "error":
		msg := body.Message
		if msg == "" {
			msg = "no reason given"
		}
		if body.Retriable != nil && *body.Retriable {
			return "warning: " + msg + " (retrying)"
		}
		return "failed: " + msg
	}
	return fmt.Sprintf("  %s %s", ev.Type, strings.TrimSpace(string(ev.Payload)))
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
	if errors.As(err, &re) && re.Status == 404 && re.Code == "" {
		return fmt.Errorf("this Router has no Olares Market configured, so it cannot install applications itself. " +
			"The model applications already on this machine still work — Router finds them on its own — and " +
			"`olares-cli market install <app>` installs new ones")
	}
	return err
}
