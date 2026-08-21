package router

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// Locating Router.
//
// Router cannot be addressed the way files.<terminus> or settings.<terminus>
// are: those are system services with subdomains derived from the Olares ID,
// while Router is a Market application whose host app-service assigns per
// install (today a hash such as https://f3395cd5.<terminus>). The
// authoritative answer is the app's own entrance vector — the same data the
// Desktop reads to draw its launcher — so that is what we read.
//
// GET {DesktopURL}/api/myapps returns every installed app with its entrances
// inline, which settles both questions a verb has ("is Router here at all?"
// and "where?") in one round trip. Unlike the rest of this tree, a 401/403
// here really is a profile problem, so it goes through whoami.HTTPClient and
// gets that package's login guidance.

// routerAppName is the Market id Router ships under.
const routerAppName = "router"

// discoveredRouter is the addressing decision, kept whole rather than reduced
// to the URL it settled on: a verb that fails at the far end reports which
// application and entrance it was talking to.
type discoveredRouter struct {
	AppName      string `json:"app_name"`
	Title        string `json:"title"`
	State        string `json:"state"`
	EntranceName string `json:"entrance_name"`
	BaseURL      string `json:"base_url"`
}

// myAppEntrance and myApp are deliberately narrow views of user-service's
// AppInfo (app.service.ts): the fields discovery reads, and nothing more.
// `settings apps` renders the wide shape; duplicating it here would be a
// second thing to keep in step for no gain.
type myAppEntrance struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	State     string `json:"state"`
	AuthLevel string `json:"authLevel"`
	Invisible bool   `json:"invisible"`
	URL       string `json:"url"`
}

type myApp struct {
	Name      string          `json:"name"`
	Title     string          `json:"title"`
	State     string          `json:"state"`
	URL       string          `json:"url"`
	Entrances []myAppEntrance `json:"entrances"`
}

// listMyApps reads every application installed for this account. Discovery
// needs it to find Router; the `local` verbs need it because Olares, not Router,
// is the authority on what is installed — Router hides the provider of an
// application that is not running.
func listMyApps(ctx context.Context, doer *whoami.HTTPClient) ([]myApp, error) {
	var env struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := doer.DoJSON(ctx, "GET", "/api/myapps", nil, &env); err != nil {
		return nil, fmt.Errorf("list installed apps: %w", err)
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			msg = fmt.Sprintf("code %d", env.Code)
		}
		return nil, fmt.Errorf("list installed apps: %s", msg)
	}
	var installed []myApp
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, &installed); err != nil {
			return nil, fmt.Errorf("decode installed apps: %w", err)
		}
	}
	return installed, nil
}

func discoverRouter(ctx context.Context, doer *whoami.HTTPClient, rp *credential.ResolvedProfile) (*discoveredRouter, error) {
	if rp == nil {
		return nil, fmt.Errorf("internal error: router discovery called without a resolved profile")
	}
	installed, err := listMyApps(ctx, doer)
	if err != nil {
		return nil, fmt.Errorf("locate Router: %w", err)
	}

	app := pickRouterApp(installed)
	if app == nil {
		return nil, fmt.Errorf(
			"Router is not installed for %s (no app named %q); install it with `olares-cli market install %s`",
			rp.OlaresID, routerAppName, routerAppName)
	}

	entrance, base := pickEntrance(*app)
	if base == "" {
		return nil, fmt.Errorf(
			"app %q is installed but exposes no reachable entrance (state %q); check it with `olares-cli market status %s`",
			app.Name, app.State, app.Name)
	}
	return &discoveredRouter{
		AppName:      app.Name,
		Title:        app.Title,
		State:        app.State,
		EntranceName: entrance,
		BaseURL:      base,
	}, nil
}

func pickRouterApp(installed []myApp) *myApp {
	for i := range installed {
		if strings.EqualFold(strings.TrimSpace(installed[i].Name), routerAppName) {
			return &installed[i]
		}
	}
	return nil
}

// pickEntrance returns the entrance name and base URL to address. Invisible
// entrances are skipped: Router marks its shared entrance invisible because it
// exists for in-cluster app-to-app calls, and it is not routable from here.
func pickEntrance(app myApp) (name, baseURL string) {
	for _, e := range app.Entrances {
		if e.Invisible || strings.TrimSpace(e.URL) == "" {
			continue
		}
		return e.Name, normalizeBaseURL(e.URL)
	}
	if u := strings.TrimSpace(app.URL); u != "" {
		return "", normalizeBaseURL(u)
	}
	return "", ""
}

// normalizeBaseURL fills in the scheme, which BFL omits — entrance URLs arrive
// as bare hosts. Entrances are TLS-terminated at the edge, so https is the
// only correct guess.
func normalizeBaseURL(raw string) string {
	u := strings.TrimRight(strings.TrimSpace(raw), "/")
	if u == "" {
		return ""
	}
	if !strings.Contains(u, "://") {
		u = "https://" + u
	}
	return u
}
