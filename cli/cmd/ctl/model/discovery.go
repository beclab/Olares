package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

// routerAppNames are the Market ids Router ships under, newest first. `router`
// (Olares 1.12.7+) supersedes `llmgatewayv3`, and each lists the other under
// `conflicts`, so at most one is ever installed — we still probe both because
// an instance that has not moved listings yet runs the older id.
var routerAppNames = []string{"router", "llmgatewayv3"}

// discoveredRouter is the addressing decision, kept whole so `model status`
// can report how it was reached rather than just the URL it settled on.
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

func discoverRouter(ctx context.Context, hc *http.Client, rp *credential.ResolvedProfile) (*discoveredRouter, error) {
	if rp == nil {
		return nil, fmt.Errorf("internal error: model discovery called without a resolved profile")
	}
	doer := whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID)

	var env struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := doer.DoJSON(ctx, "GET", "/api/myapps", nil, &env); err != nil {
		return nil, fmt.Errorf("list installed apps to locate Router: %w", err)
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			msg = fmt.Sprintf("code %d", env.Code)
		}
		return nil, fmt.Errorf("list installed apps to locate Router: %s", msg)
	}
	var installed []myApp
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, &installed); err != nil {
			return nil, fmt.Errorf("decode installed apps: %w", err)
		}
	}

	app := pickRouterApp(installed)
	if app == nil {
		return nil, fmt.Errorf(
			"Router is not installed for %s (looked for the app ids %s); install it with `olares-cli market install router`",
			rp.OlaresID, strings.Join(routerAppNames, ", "))
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

// pickRouterApp honours routerAppNames' order rather than the server's, so a
// mid-migration instance that somehow reports both resolves to the newer one.
func pickRouterApp(installed []myApp) *myApp {
	for _, want := range routerAppNames {
		for i := range installed {
			if strings.EqualFold(strings.TrimSpace(installed[i].Name), want) {
				return &installed[i]
			}
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
