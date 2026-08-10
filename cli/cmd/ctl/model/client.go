package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/beclab/Olares/cli/pkg/credential"
)

// Route prefixes on the entrance host. The console plane is the management
// surface every configuration verb uses; the data plane is OpenAI-shaped and
// takes an `sk-*` key rather than the session this tree carries.
const (
	consoleAPI   = "/console/api"
	dataPlaneAPI = "/v1"
)

// statusOlaresEdgeAuthFailure is the non-standard code the Olares edge
// (Authelia ext-authz through l4-bfl-proxy) returns when an otherwise valid
// request carries a token it no longer accepts. The Factory transport already
// treats it as a refresh trigger; we only need to recognise it when deciding
// who rejected a request.
const statusOlaresEdgeAuthFailure = 459

// routerClient is the transport for Router's own HTTP surface: the
// /console/api management plane, the /v1 data plane, and /healthz. All three
// reach the backend through the app entrance, because Router's frontend image
// reverse-proxies them (deploy/frontend.nginx.conf.template in the router
// repo).
//
// Why not whoami.HTTPClient, which the other command trees share: it reads
// every 401/403 as "this profile's token is stale, run profile login". Router
// answers a perfectly good non-admin session with 403
// forbidden_admin_required — it is an admin-only app — and reporting that as
// a login problem sends the user off to fix something that is not broken. So
// this client tells the two apart by who produced the body: one of Router's
// error envelopes means Router itself judged an authenticated caller, and
// anything else at those statuses is the edge refusing the token, which is
// where credential.FormatHTTPAuthError's login call-to-action belongs.
type routerClient struct {
	hc       *http.Client
	baseURL  string
	olaresID string
	headers  map[string]string
}

func newRouterClient(hc *http.Client, baseURL, olaresID string) *routerClient {
	return &routerClient{
		hc:       hc,
		baseURL:  strings.TrimRight(baseURL, "/"),
		olaresID: olaresID,
	}
}

// withHeader returns a copy that adds one header to every request. The data
// plane needs an `sk-*` bearer that the management plane must never carry, so
// the two are separate clients over the same session rather than one client
// with a mutable header.
func (c *routerClient) withHeader(name, value string) *routerClient {
	clone := *c
	clone.headers = make(map[string]string, len(c.headers)+1)
	for k, v := range c.headers {
		clone.headers[k] = v
	}
	clone.headers[name] = value
	return &clone
}

func (c *routerClient) applyHeaders(req *http.Request) {
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
}

// RouterError is a non-2xx response that carried one of Router's two error
// envelopes. Both nest under "error" and agree on `code` and `message` — the
// console plane adds `details`, the OpenAI-shaped data plane adds `type` and
// `param` — so a single decoder serves the whole surface. The Model Console
// behind a local model uses the same shape and adds `phase`, which is why this
// client serves that host too.
type RouterError struct {
	Method  string
	Path    string
	Status  int
	Code    string
	Type    string
	Message string
	Phase   string
	Body    []byte
}

func (e *RouterError) Error() string {
	label := e.Code
	if label == "" {
		label = e.Type
	}
	msg := e.Message
	if msg == "" {
		msg = truncate(string(e.Body), 200)
	}
	if strings.TrimSpace(msg) == "" {
		// A status with no body at all. Router's own refusals always carry an
		// envelope, and it passes an upstream's reply through unchanged, so an
		// empty one was written by a proxy on the way rather than by anything
		// that knew what was being asked.
		msg = http.StatusText(e.Status) + ", with an empty response body"
	}
	if p := strings.TrimSpace(e.Phase); p != "" {
		msg += " (phase " + p + ")"
	}
	if label == "" {
		return fmt.Sprintf("%s %s: HTTP %d: %s", e.Method, e.Path, e.Status, msg)
	}
	return fmt.Sprintf("%s %s: HTTP %d (%s): %s%s",
		e.Method, e.Path, e.Status, label, msg, e.recovery())
}

// recovery names the next action for the rejections a caller can actually do
// something about. Everything else reads better without a guess appended.
func (e *RouterError) recovery() string {
	switch e.Code {
	case "forbidden_admin_required":
		return "; Router restricts this to admins — switch with `olares-cli profile use <admin-profile>`"
	case "missing_bfl_user_header":
		return "; the request reached Router without an Olares identity — re-authenticate with `olares-cli profile login`"
	case "invalid_api_key", "key_disabled", "key_expired":
		return "; list and re-issue keys with `olares-cli model key list`"
	case "observability_content_policy_forbids":
		return "; this deployment forbids storing prompt content, so the preference cannot be set either way"
	case "predefined_models_unknown":
		return "; Router does not say which name it rejected — compare against " +
			"`olares-cli model provider types <vendor> --models`"
	case "market_app_not_found":
		return "; the Market will not act on this application in its current state — " +
			"`olares-cli market status <app>` says what that state is, and a stopped app has to be resumed first"
	}
	return ""
}

// doJSON sends an optional JSON body and decodes an optional JSON response.
func (c *routerClient) doJSON(ctx context.Context, method, path string, body, out interface{}) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}
	resp, err := c.do(ctx, method, path, reader, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read %s %s response: %w", method, path, err)
	}
	if resp.StatusCode/100 != 2 {
		return c.formatErr(method, path, resp.StatusCode, respBody)
	}
	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode %s %s response: %w (body=%s)",
			method, path, err, truncate(string(respBody), 200))
	}
	return nil
}

// do issues one request and hands back the live response with its body still
// open, for the callers that stream (install progress, chat completions).
func (c *routerClient) do(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil && contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	c.applyHeaders(req)

	resp, err := c.hc.Do(req)
	if err != nil {
		// The Factory transport raises a typed credential error when
		// /api/refresh itself fails; http.Client buries it in a *url.Error,
		// and errors.As walks that chain. Surfacing the typed value keeps the
		// canonical "run profile login" message intact.
		var invalidated *credential.ErrTokenInvalidated
		if errors.As(err, &invalidated) {
			return nil, invalidated
		}
		var notLoggedIn *credential.ErrNotLoggedIn
		if errors.As(err, &notLoggedIn) {
			return nil, notLoggedIn
		}
		return nil, fmt.Errorf("%s %s: %w", method, url, err)
	}
	return resp, nil
}

// doStream opens a server-sent event stream and hands back the live body. The
// status is checked here rather than by the caller: an error arrives as a JSON
// envelope on the same connection an event stream would have used, and reading
// it as SSE would report a rejection as an empty stream.
func (c *routerClient) doStream(ctx context.Context, path string) (*http.Response, error) {
	fullURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	c.applyHeaders(req)

	resp, err := c.hc.Do(req)
	if err != nil {
		var invalidated *credential.ErrTokenInvalidated
		if errors.As(err, &invalidated) {
			return nil, invalidated
		}
		var notLoggedIn *credential.ErrNotLoggedIn
		if errors.As(err, &notLoggedIn) {
			return nil, notLoggedIn
		}
		return nil, fmt.Errorf("GET %s: %w", fullURL, err)
	}
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, c.formatErr("GET", path, resp.StatusCode, body)
	}
	return resp, nil
}

func (c *routerClient) formatErr(method, path string, status int, body []byte) error {
	var env struct {
		Error struct {
			Code    string `json:"code"`
			Type    string `json:"type"`
			Message string `json:"message"`
			Phase   string `json:"phase"`
		} `json:"error"`
	}
	fromRouter := json.Unmarshal(body, &env) == nil &&
		(env.Error.Message != "" || env.Error.Code != "")

	if !fromRouter {
		switch status {
		case http.StatusUnauthorized, http.StatusForbidden, statusOlaresEdgeAuthFailure:
			// No Router envelope at an auth status: the edge turned the
			// request away before Router ever saw it.
			return credential.FormatHTTPAuthError(status, body, c.olaresID)
		}
		// Still a RouterError, with no code. An unmounted route answers
		// this way — gin's own 404 carries no envelope — and callers that
		// can explain their absence need the status to recognise it.
		return &RouterError{
			Method: method,
			Path:   path,
			Status: status,
			Body:   append([]byte(nil), body...),
		}
	}
	return &RouterError{
		Method:  method,
		Path:    path,
		Status:  status,
		Code:    env.Error.Code,
		Type:    env.Error.Type,
		Message: env.Error.Message,
		Phase:   env.Error.Phase,
		Body:    append([]byte(nil), body...),
	}
}
