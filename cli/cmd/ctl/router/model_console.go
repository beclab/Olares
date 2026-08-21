package router

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// The Model Console inside one model application.
//
// Most of this tree talks to Router, which is the right level for "which models
// can this Olares reach". It is the wrong level for "why is this model not
// answering yet": Router sees an upstream that does or does not reply, while
// the application itself knows it is 40% through a download, that the weights
// are on disk but the engine has not loaded them, or that the last verification
// failed a checksum. Those facts live in the Model Console — the llm-init
// process every model application ships alongside its engine — and the verbs
// that read them are the only way to reach it.
//
// Addressing. A model application publishes its Model Console on its own
// entrance, so the app is named and its entrance looked up the same way Router
// itself is found. An application Router knows as a provider can be named
// either way: `router provider get` prints both the provider title and the
// Olares app id, and both resolve here.
//
// Identity. The entrance is `authLevel: internal`, which the active profile
// already satisfies; the Model Console has no credentials of its own. So there
// is nothing to configure here, and a 401 means the profile rather than the
// application.

// llmInit wraps the client and the addressing decision, so a verb reports which
// application it reached rather than only what it found there.
type llmInit struct {
	client   *routerClient
	AppName  string          `json:"app_name"`
	Title    string          `json:"title"`
	State    string          `json:"state"`
	BaseURL  string          `json:"base_url"`
	Provider string          `json:"provider,omitempty"`
	Console  *localBuildInfo `json:"console,omitempty"`
}

// localAddressing is how a verb that reaches into a model application decides
// which one, and there are exactly two ways.
//
// A model reference is the normal one, because a model is what the user has a
// question about; the application serving it is an implementation detail they
// should not have to know. Resolving it costs a read of Router's model list.
//
// An app id is the other, and it exists for the moment the first one cannot be
// paid for. A model that is not answering is precisely when Router may also be
// unreachable, misconfigured, or answering with a stale row — and being unable
// to look at the application then would take the diagnostic away exactly when
// it is needed. So `--app` skips Router entirely: `prepareLocal` builds a
// client against the application's own entrance whether Router is there or not.
type localAddressing int

const (
	// addressByModel is `<verb> <model>` with `--app` as the escape.
	addressByModel localAddressing = iota
	// addressByApp is the retired `router local <verb> <app>` shape, where the
	// positional was always the application.
	addressByApp
)

// modelTarget carries the addressing decision and the `--app` flag it may need.
type modelTarget struct {
	how localAddressing
	app string
}

func newModelTarget(how localAddressing) *modelTarget { return &modelTarget{how: how} }

// arg is what the positional is called in a Use line.
func (t *modelTarget) arg() string {
	if t.how == addressByApp {
		return "<app>"
	}
	return "<model>"
}

func (t *modelTarget) bind(cmd *cobra.Command) {
	if t.how == addressByApp {
		cmd.Args = cobra.ExactArgs(1)
		return
	}
	cmd.Args = cobra.MaximumNArgs(1)
	cmd.Flags().StringVar(&t.app, "app", "",
		"reach the application directly by its Olares app id, without asking Router which one serves the model")
}

// appRef turns what was typed into the application reference openLocal takes.
func (t *modelTarget) appRef(ctx context.Context, f *cmdutil.Factory, args []string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if t.how == addressByApp {
		return args[0], nil
	}
	app := strings.TrimSpace(t.app)
	switch {
	case app != "" && len(args) > 0:
		return "", fmt.Errorf("name a model or pass --app, not both: %q is a model and %q is an "+
			"application, and if they disagree there is no right answer", args[0], app)
	case app != "":
		return app, nil
	case len(args) == 0:
		return "", errNoModelTarget
	}
	return appServingModel(ctx, f, args[0])
}

var errNoModelTarget = errors.New(
	"name the model to act on, or the application serving it with --app; " +
		"`olares-cli router list` shows the models and `olares-cli router provider list` the applications")

// direct reports whether this invocation bypasses Router. It is asked by the
// two verbs that have a Router road as well as a direct one; everything else
// here only has the direct one.
func (t *modelTarget) direct(args []string) bool {
	return t.how == addressByApp || strings.TrimSpace(t.app) != "" || len(args) == 0
}

// model is the reference to hand a Router route, and is only meaningful when
// direct said no.
func (t *modelTarget) model(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

// appServingModel asks Router which application serves a model.
//
// The failure worth spelling out is a model that resolves to a cloud provider:
// there is no Model Console at the other end of one, and nothing here applies
// to it. The other is Router being unavailable, which is why the refusal names
// `--app` — the whole point of that flag is that this lookup can be skipped.
func appServingModel(ctx context.Context, f *cmdutil.Factory, modelRef string) (string, error) {
	pc, err := prepare(ctx, f)
	if err != nil {
		return "", fmt.Errorf("%w\nRouter is what says which application serves a model. To look at an "+
			"application without asking it, name the application: `--app <app_id>`", err)
	}
	row, err := resolveModel(ctx, pc, modelRef)
	if err != nil {
		return "", err
	}
	app := strings.TrimSpace(derefOr(row.OlaresAppName, ""))
	if app == "" {
		return "", fmt.Errorf("%s is served by %s, which is a %s upstream rather than an application on "+
			"this machine. Only a model running here has a Model Console to read; what Router believes "+
			"about a cloud model is the row itself, and `olares-cli router model get %s` shows it",
			row.label(), row.ProviderName, row.ProviderType, modelRef)
	}
	return app, nil
}

// open is what every verb here begins with: work out which application, then
// reach its Model Console.
func (t *modelTarget) open(ctx context.Context, f *cmdutil.Factory, args []string) (*llmInit, error) {
	ref, err := t.appRef(ctx, f, args)
	if err != nil {
		return nil, err
	}
	return openLocal(ctx, f, ref)
}

// openLocal resolves an application reference to its Model Console.
//
// Olares is asked first and Router second, because they answer different
// questions: Olares knows what is installed, Router knows what it has been
// configured to route to. A model application that Router has never been told
// about still has a Model Console worth reading, and an application Router
// knows by a provider title is one whose app id the user may never have seen.
//
// Router is therefore not required to get here. Everything else in this tree
// needs it and says so; a model that is not answering is exactly when a user
// should not have to have a working gateway to find out why.
func openLocal(ctx context.Context, f *cmdutil.Factory, ref string) (*llmInit, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("name the application; `olares-cli router provider list` shows both the " +
			"provider title and the Olares app it runs as")
	}
	pc, err := prepareLocal(ctx, f)
	if err != nil {
		return nil, err
	}

	installed, err := listMyApps(ctx, pc.desktop)
	if err != nil {
		return nil, err
	}
	var li *llmInit
	if app := findInstalledApp(installed, ref); app != nil {
		li, err = localFromApp(pc, *app, "")
	} else {
		// Not an app id. Try Router's providers, where the same application may
		// be listed under a human title.
		li, err = localFromProvider(ctx, pc, ref, installed)
	}
	if err != nil {
		return nil, err
	}
	if err := confirmConsole(ctx, li); err != nil {
		return nil, err
	}
	return li, nil
}

// confirmConsole checks that something on the other end is a Model Console
// before any verb reads it.
//
// Without this the tree happily addresses any application at all, and the
// result is worse than an error: an app's own /healthz can answer with a
// document that overlaps this one enough to render, and the fields it does not
// have appear as a model that is not ready. Build metadata is the right probe —
// every Model Console serves it, and serves it regardless of engine or
// lifecycle phase, so a refusal here is about what the application is rather
// than about how far along it is.
func confirmConsole(ctx context.Context, li *llmInit) error {
	var build localBuildInfo
	err := li.client.doJSON(ctx, "GET", epLocalBuildInfo, nil, &build)
	switch {
	case err == nil && strings.TrimSpace(build.Version) != "":
		li.Console = &build
		return nil
	case err == nil:
		// Valid JSON, no version in it. Some other service that happens to
		// answer this path.
	default:
		var re *RouterError
		if errors.As(err, &re) && re.Status >= 500 {
			return fmt.Errorf("application %q is running, but nothing inside it is answering yet: %w\n"+
				"An application that has just started does this for a while; `olares-cli doctor %s` "+
				"says whether it is stuck", li.AppName, err, li.AppName)
		}
	}
	return fmt.Errorf("application %q does not run a Model Console, so it has no model lifecycle to "+
		"read. This subtree is for the applications that serve models on this machine — "+
		"`olares-cli router provider list` shows which those are", li.AppName)
}

// prepareLocal is prepare without the requirement that Router be installed.
// The Router client is still built when discovery finds it, because a provider
// title is one of the two ways an application may be named here.
func prepareLocal(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: the Model Console verbs are not wired with cmdutil.Factory")
	}
	if err := cmdutil.RequireMinVersion(ctx, f, cmdutil.MinVersionGate{
		Verb:       "router model status",
		MinVersion: minOlaresVersion,
		Reason:     "the Model Console inside model applications",
	}); err != nil {
		return nil, err
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	pc := &preparedClient{
		profile: rp,
		desktop: whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID),
		hc:      hc,
	}
	if found, derr := discoverRouter(ctx, pc.desktop, rp); derr == nil {
		pc.found = found
		pc.router = newRouterClient(hc, found.BaseURL, rp.OlaresID)
	}
	return pc, nil
}

// notServing names the states in which no process is listening, and says what
// to do about each.
//
// Installing and upgrading are deliberately absent: that is when a Model
// Console is up and downloading weights, and reading its progress is the main
// reason this subtree exists.
func notServing(state, app string) string {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "stopped", "suspend", "suspended":
		return "`olares-cli market resume " + app + "` starts it"
	case "uninstalling", "uninstalled":
		return "it is being removed"
	case "crash", "crashloopbackoff":
		return "it is failing to start, which is a matter for `olares-cli doctor " + app + "`"
	}
	return ""
}

// consoleServes asks the application's own catalogue whether a route exists on
// it.
//
// Model Consoles ship with the applications that embed them, so an Olares has
// as many versions of this API as it has model applications, and a route added
// upstream is simply absent on the older ones. A bare "404 page not found" then
// reads as a fault in the request. This is the discriminator: the catalogue is
// built from the routes actually registered on that process.
//
// A catalogue that cannot be read leaves the question open, which is why the
// answer is a pointer rather than a bool — callers keep the original 404 in
// that case instead of guessing either way.
func consoleServes(ctx context.Context, li *llmInit, method, path string) *bool {
	var env struct {
		Endpoints []struct {
			Method    string `json:"method"`
			Path      string `json:"path"`
			Available bool   `json:"available"`
		} `json:"endpoints"`
	}
	if err := li.client.doJSON(ctx, "GET", epLocalEndpoints, nil, &env); err != nil {
		return nil
	}
	for _, e := range env.Endpoints {
		if strings.EqualFold(e.Method, method) && e.Path == path {
			served := e.Available
			return &served
		}
	}
	no := false
	return &no
}

func findInstalledApp(installed []myApp, ref string) *myApp {
	for i := range installed {
		if strings.EqualFold(strings.TrimSpace(installed[i].Name), ref) {
			return &installed[i]
		}
	}
	return nil
}

func localFromApp(pc *preparedClient, app myApp, providerName string) (*llmInit, error) {
	_, base := pickEntrance(app)
	if base == "" {
		return nil, fmt.Errorf("application %q exposes no reachable entrance (state %q), so its Model "+
			"Console cannot be read from here; `olares-cli market status %s` says why",
			app.Name, app.State, app.Name)
	}
	if reason := notServing(app.State, app.Name); reason != "" {
		// An application that is not running still has an entrance, and every
		// probe against it comes back as the same connection failure from the
		// edge. Reporting the state once is the same information without
		// four copies of a message about a socket.
		return nil, fmt.Errorf("application %q is %s, so nothing is listening on its Model Console; %s",
			app.Name, app.State, reason)
	}
	return &llmInit{
		client:   newRouterClient(pc.hc, base, pc.profile.OlaresID),
		AppName:  app.Name,
		Title:    app.Title,
		State:    app.State,
		BaseURL:  base,
		Provider: providerName,
	}, nil
}

// localFromProvider matches a Router provider by name, title or id and follows
// it back to the application it runs as. The provider row also carries an
// entrance URL, which is the fallback when Olares reports the application under
// a name the provider does not repeat.
func localFromProvider(ctx context.Context, pc *preparedClient, ref string, installed []myApp) (*llmInit, error) {
	if pc.router == nil {
		return nil, fmt.Errorf("no application called %q is installed here, and Router — which is where a "+
			"provider title could be matched — is not installed either; name the application by its "+
			"Olares app id, which `olares-cli market list` shows", ref)
	}
	rows, err := listProviders(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("no application called %q is installed, and Router's providers could not "+
			"be read to match it by title: %w", ref, err)
	}
	for i := range rows {
		r := &rows[i]
		if !strings.EqualFold(r.Name, ref) && !strings.EqualFold(r.title(), ref) && r.ID != ref {
			continue
		}
		appName := derefOr(r.OlaresAppName, "")
		if strings.TrimSpace(appName) == "" {
			return nil, fmt.Errorf("provider %q is not an application on this machine — it is a %s "+
				"upstream, and only a model running here has a Model Console", r.title(), r.ProviderType)
		}
		if app := findInstalledApp(installed, appName); app != nil {
			return localFromApp(pc, *app, r.title())
		}
		if base := normalizeBaseURL(derefOr(r.EntranceURL, "")); base != "" {
			return &llmInit{
				client:   newRouterClient(pc.hc, base, pc.profile.OlaresID),
				AppName:  appName,
				Title:    r.title(),
				BaseURL:  base,
				Provider: r.title(),
			}, nil
		}
		return nil, fmt.Errorf("provider %q names the application %q, which Olares does not list as "+
			"installed; `olares-cli market status %s` is the place to start",
			r.title(), appName, appName)
	}
	return nil, fmt.Errorf("nothing called %q is installed here or listed as a Router provider; "+
		"`olares-cli router provider list` shows the model applications and the app ids they run as", ref)
}

// Wire shapes. These mirror the Model Console's own documented responses; the
// fields this tree renders, and nothing more.

type localHealth struct {
	Status       string  `json:"status"`
	Ready        bool    `json:"ready"`
	EngineAlive  bool    `json:"engine_alive"`
	ModelExists  bool    `json:"model_exists"`
	Phase        string  `json:"phase"`
	LastVerify   *string `json:"last_verify"`
	LastVerifyOK *bool   `json:"last_verify_ok"`
}

type localProgress struct {
	Phase            string  `json:"phase"`
	StartedAt        string  `json:"started_at"`
	UpdatedAt        string  `json:"updated_at"`
	BytesTotal       int64   `json:"bytes_total"`
	BytesCompleted   int64   `json:"bytes_completed"`
	SpeedBytesPerSec float64 `json:"speed_bytes_per_sec"`
	ETASeconds       int64   `json:"eta_seconds"`
	LastError        string  `json:"last_error"`
	RetryCount       int     `json:"retry_count"`
	TransportRetries int     `json:"transport_retries"`
	LastVerifyAt     *string `json:"last_verify_at"`
	LastVerifyOK     *bool   `json:"last_verify_ok"`
}

type localBuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// fraction is the share of the download that is done, or -1 when the total is
// not known yet. A download that has not been sized reports 0 bytes of 0, and
// calling that 100% would announce a finished download at the moment it starts.
func (p localProgress) fraction() float64 {
	if p.BytesTotal <= 0 {
		return -1
	}
	f := float64(p.BytesCompleted) / float64(p.BytesTotal)
	if f > 1 {
		return 1
	}
	return f
}

// phaseNote translates the lifecycle phase into what it means for a caller
// waiting on the model. The phase names are the Model Console's; the
// consequence is what a reader actually wants.
func phaseNote(phase string, h *localHealth) string {
	switch strings.ToLower(strings.TrimSpace(phase)) {
	case "init":
		return "starting up; nothing has been fetched yet"
	case "download":
		return "fetching the weights; calls are refused until this finishes"
	case "loading":
		return "the engine is starting on weights that are already on disk"
	case "ready":
		if h != nil && !h.ModelExists {
			return "serving, but the engine no longer reports the configured model"
		}
		return "serving"
	case "degraded":
		return "serving in a reduced state; `router model progress` carries the reason"
	case "failed":
		return "stopped trying; `router model retry` re-enters the loop"
	}
	return ""
}

func fmtDuration(seconds int64) string {
	if seconds <= 0 {
		return "—"
	}
	d := time.Duration(seconds) * time.Second
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
	default:
		return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
	}
}
