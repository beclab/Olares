package model

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

// `olares-cli model local …` — the Model Console inside one model application.
//
// Everything else in this tree talks to Router, which is the right level for
// "which models can this Olares reach". It is the wrong level for "why is this
// model not answering yet": Router sees an upstream that does or does not
// reply, while the application itself knows it is 40% through a download, that
// the weights are on disk but the engine has not loaded them, or that the last
// verification failed a checksum. Those facts live in the Model Console — the
// llm-init process every model application ships alongside its engine — and
// this subtree is the only way to read them.
//
// Addressing. A model application publishes its Model Console on its own
// entrance, so the app is named and its entrance looked up the same way Router
// itself is found. An application Router knows as a provider can be named
// either way: `model provider get` prints both the provider title and the
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

func NewLocalCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local <app> <verb>",
		Short: "the Model Console inside a locally installed model",
		Long: `Read and steer the Model Console behind a model that runs on this Olares.

Router answers whether a model replies. This answers why it does not yet: what
is downloading, whether the engine has the weights, what the model card claims,
and how the last verification went.

The application is named first, either by its Olares app id or by the provider
title Router lists it under — "model provider list" prints both.

  status <app>      lifecycle phase, engine, and last verification
  progress <app>    the download and load snapshot, with --watch
  spec <app>        the model card this application serves to Router
  config <app>      the effective configuration, secrets redacted
  endpoints <app>   which routes this deployment actually serves
  gpu <app>         how much of the model is resident on the GPU
  perf <app>        time to first token and throughput, measured
  retry <app>       re-enter the download and load loop now
  restart <app>     relaunch the engine with the current card

These read the application, not Router: a model missing from "model list" is a
Router configuration matter, and a model listed but silent is one for here.
`,
	}
	cmd.AddCommand(newLocalStatusCommand(f))
	cmd.AddCommand(newLocalProgressCommand(f))
	cmd.AddCommand(newLocalSpecCommand(f))
	cmd.AddCommand(newLocalConfigCommand(f))
	cmd.AddCommand(newLocalEndpointsCommand(f))
	cmd.AddCommand(newLocalGPUCommand(f))
	cmd.AddCommand(newLocalPerfCommand(f))
	cmd.AddCommand(newLocalRetryCommand(f))
	cmd.AddCommand(newLocalRestartCommand(f))
	return cmd
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
		return nil, fmt.Errorf("name the application; `olares-cli model provider list` shows both the " +
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
	err := li.client.doJSON(ctx, "GET", "/api/build-info", nil, &build)
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
		"`olares-cli model provider list` shows which those are", li.AppName)
}

// prepareLocal is prepare without the requirement that Router be installed.
// The Router client is still built when discovery finds it, because a provider
// title is one of the two ways an application may be named here.
func prepareLocal(ctx context.Context, f *cmdutil.Factory) (*preparedClient, error) {
	if f == nil {
		return nil, fmt.Errorf("internal error: model local not wired with cmdutil.Factory")
	}
	if err := cmdutil.RequireMinVersion(ctx, f, cmdutil.MinVersionGate{
		Verb:       "model local",
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
	if err := li.client.doJSON(ctx, "GET", "/api/endpoints", nil, &env); err != nil {
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
		"`olares-cli model provider list` shows the model applications and the app ids they run as", ref)
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
		return "serving in a reduced state; `model local progress` carries the reason"
	case "failed":
		return "stopped trying; `model local retry` re-enters the loop"
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
