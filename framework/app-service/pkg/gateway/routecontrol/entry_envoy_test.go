package routecontrol

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// This opt-in test is required by the app-service CI workflow. It uses the
// shipped tenant auth proxy and generated SecurityPolicy response allowlist in
// a real, digest-pinned Envoy. The authorizer is an explicit policy fixture,
// not a claim that production Authelia sessions have been tested.
const entryEnvoyImage = "envoyproxy/envoy@sha256:a7a56545102f7a682e0cafea2c9b8448af1b09ebb710eab688dfb931e3ec7ff6"

func TestEntryIdentityRealEnvoy(t *testing.T) {
	if os.Getenv("ENTRY_ENVOY_TEST") != "1" {
		t.Skip("ENTRY_ENVOY_TEST=1 enables the isolated Envoy contract test")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	docker := func(args ...string) string {
		t.Helper()
		out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("docker %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	typ := func(name string, fields map[string]any) map[string]any {
		fields["@type"] = "type.googleapis.com/" + name
		return fields
	}
	router := map[string]any{"name": "envoy.filters.http.router", "typed_config": typ("envoy.extensions.filters.http.router.v3.Router", map[string]any{})}
	socket := func(port int) map[string]any {
		return map[string]any{"socket_address": map[string]any{"address": "127.0.0.1", "port_value": port}}
	}
	listener := func(name string, port int, hcm map[string]any) map[string]any {
		addr := socket(port)
		if port == 10000 || port == 10004 {
			addr = map[string]any{"socket_address": map[string]any{"address": "0.0.0.0", "port_value": port}}
		}
		return map[string]any{"name": name, "address": addr, "filter_chains": []any{map[string]any{"filters": []any{map[string]any{"name": "envoy.filters.network.http_connection_manager", "typed_config": hcm}}}}}
	}
	hcm := func(routes []any, filters []any) map[string]any {
		return typ("envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager", map[string]any{"stat_prefix": "fixture", "route_config": map[string]any{"name": "fixture", "virtual_hosts": []any{map[string]any{"name": "fixture", "domains": []any{"*"}, "routes": routes}}}, "http_filters": filters})
	}
	add := func(key, value string) map[string]any {
		return map[string]any{"header": map[string]any{"key": key, "value": value}, "append_action": "OVERWRITE_IF_EXISTS_OR_ADD"}
	}
	cluster := func(name string, port int) map[string]any {
		return map[string]any{"name": name, "type": "STATIC", "connect_timeout": "1s", "load_assignment": map[string]any{"cluster_name": name, "endpoints": []any{map[string]any{"lb_endpoints": []any{map[string]any{"endpoint": map[string]any{"address": socket(port)}}}}}}}
	}
	proxy := tenantAuthBootstrap(t, "alice")
	rawListeners, _, _ := unstructured.NestedSlice(proxy, "static_resources", "listeners")
	var authProxy any
	for _, l := range rawListeners {
		if l.(map[string]any)["name"] == "listener_auth_backend_proxy" {
			authProxy = l
		}
	}
	if authProxy == nil {
		t.Fatal("missing shipped authorization proxy")
	}
	app := desktopEntrance()
	app.SetName("app-olares-app-files")
	policy := desiredApplicationExtAuthSecurityPolicy(app)
	checkPath, _, _ := unstructured.NestedString(policy.Object, "spec", "extAuth", "http", "pathOverride")
	backendHeaders, _, _ := unstructured.NestedStringSlice(policy.Object, "spec", "extAuth", "http", "headersToBackend")
	requestHeaders, _, _ := unstructured.NestedStringSlice(policy.Object, "spec", "extAuth", "headersToExtAuth")
	patterns := func(names []string) []any {
		out := []any{}
		for _, name := range names {
			out = append(out, map[string]any{"exact": name, "ignore_case": true})
		}
		return out
	}
	authFilter := map[string]any{"name": "envoy.filters.http.ext_authz", "typed_config": typ("envoy.extensions.filters.http.ext_authz.v3.ExtAuthz", map[string]any{
		"transport_api_version": "V3", "failure_mode_allow": false, "status_on_error": map[string]any{"code": "ServiceUnavailable"},
		"allowed_headers": map[string]any{"patterns": patterns(requestHeaders)},
		"http_service":    map[string]any{"server_uri": map[string]any{"uri": "http://tenant-auth:9091", "cluster": "tenant-auth", "timeout": "2s"}, "path_override": checkPath, "authorization_response": map[string]any{"allowed_upstream_headers": map[string]any{"patterns": patterns(backendHeaders)}}},
	})}
	token, jwtFilter, sharedRoute := callerJWTEnvoyContract(t)
	northSharedRoute := map[string]any{}
	for k, v := range sharedRoute {
		northSharedRoute[k] = v
	}
	northSharedRoute["typed_per_filter_config"] = map[string]any{"envoy.filters.http.ext_authz": typ("envoy.extensions.filters.http.ext_authz.v3.ExtAuthzPerRoute", map[string]any{"disabled": true})}
	edge := hcm([]any{northSharedRoute, map[string]any{"match": map[string]any{"prefix": "/"}, "route": map[string]any{"cluster": "business"}}}, []any{jwtFilter, authFilter, router})
	internal := hcm([]any{sharedRoute}, []any{jwtFilter, router})
	setEarlyStrip := func(h map[string]any, settings map[string]any) {
		mutations := []any{}
		for _, v := range settings["remove"].([]any) {
			mutations = append(mutations, map[string]any{"remove": v})
		}
		if prefixes, ok := settings["removeOnMatch"].([]any); ok {
			for _, v := range prefixes {
				mutations = append(mutations, map[string]any{"remove_on_match": map[string]any{"key_matcher": map[string]any{"prefix": v.(map[string]any)["value"], "ignore_case": true}}})
			}
		}
		h["early_header_mutation_extensions"] = []any{map[string]any{"name": "envoy.http.early_header_mutation.header_mutation", "typed_config": typ("envoy.extensions.http.early_header_mutation.header_mutation.v3.HeaderMutation", map[string]any{"mutations": mutations})}}
	}
	setEarlyStrip(edge, entryIdentityHeaderSettings())
	setEarlyStrip(internal, internalIdentityHeaderSettings(t))
	// Each fake session represents a policy decision, not a real user's cookie.
	match := func(cookie string, tenant bool) map[string]any {
		headers := []any{map[string]any{"name": "cookie", "string_match": map[string]any{"exact": cookie}}}
		if tenant {
			headers = append(headers, map[string]any{"name": "x-bfl-user", "string_match": map[string]any{"exact": "alice"}})
		}
		return map[string]any{"prefix": "/", "headers": headers}
	}
	decisions := []any{
		// Reproduce the authorizer's slash normalization for the Files host.
		// A prefix here would incorrectly turn the browser navigation into a redirect.
		map[string]any{"match": map[string]any{"prefix": "/", "headers": []any{
			map[string]any{"name": ":authority", "string_match": map[string]any{"exact": "files.alice.olares.local"}},
			map[string]any{"name": ":path", "string_match": map[string]any{"safe_regex": map[string]any{"regex": `^/api/verify(\?.*)?$`}}, "invert_match": true},
		}}, "direct_response": map[string]any{"status": 301}, "response_headers_to_add": []any{add("Location", AutheliaVerifyPath)}},
		// An explicitly authorized different subject must remain that subject.
		// This fixture does not grant cross-tenant Desktop access in production.
		map[string]any{"match": match("session=delegate", true), "direct_response": map[string]any{"status": 200}, "response_headers_to_add": []any{add("Remote-User", "bob"), add("X-BFL-USER", "bob"), add("Remote-Groups", "viewer"), add("Remote-Name", "%REQ(x-forwarded-for)%"), add("Remote-Email", "%REQ(x-original-url)%")}},
		map[string]any{"match": match("session=alice", true), "direct_response": map[string]any{"status": 200}, "response_headers_to_add": []any{add("Remote-User", "alice"), add("X-BFL-USER", "alice"), add("Remote-Groups", "owner"), add("Remote-Name", "%REQ(x-forwarded-for)%"), add("Remote-Email", "%REQ(x-original-url)%")}},
		map[string]any{"match": match("session=bob", false), "direct_response": map[string]any{"status": 403}},
		map[string]any{"match": match("session=expired", false), "direct_response": map[string]any{"status": 401}},
		map[string]any{"match": match("policy=public", true), "direct_response": map[string]any{"status": 200}},
		map[string]any{"match": map[string]any{"prefix": "/", "headers": []any{map[string]any{"name": ":authority", "string_match": map[string]any{"exact": "desktop.alice.olares.local"}}}}, "direct_response": map[string]any{"status": 302}, "response_headers_to_add": []any{add("Location", "http://auth.alice.olares.local/?rd=%REQ(x-original-url)%")}},
		map[string]any{"match": map[string]any{"prefix": "/"}, "direct_response": map[string]any{"status": 302}, "response_headers_to_add": []any{add("Location", "https://auth.alice.example.org/?rd=%REQ(x-original-url)%")}},
	}
	echoScript := `function envoy_on_request(h)
 local r=h:headers()
 h:respond({[":status"]="200",["content-type"]="application/json"},string.format('{"user":%q,"remote":%q,"groups":%q,"cookie":%q,"xff":%q,"authip":%q,"authurl":%q,"caller":%q,"appid":%q,"jwt":%q}',r:get("x-bfl-user") or "",r:get("remote-user") or "",r:get("remote-groups") or "",r:get("cookie") or "",r:get("x-forwarded-for") or "",r:get("remote-name") or "",r:get("remote-email") or "",r:get("x-caller-user") or "",r:get("x-caller-appid") or "",r:get("x-olares-caller-jwt") and "present" or ""))
end`
	echoFilter := map[string]any{"name": "envoy.filters.http.lua", "typed_config": typ("envoy.extensions.filters.http.lua.v3.Lua", map[string]any{"default_source_code": map[string]any{"inline_string": echoScript}})}
	echo := hcm([]any{map[string]any{"match": map[string]any{"prefix": "/"}, "direct_response": map[string]any{"status": 500}}}, []any{echoFilter, router})
	bootstrap := map[string]any{"static_resources": map[string]any{"listeners": []any{listener("entry", 10000, edge), listener("internal", 10004, internal), authProxy, listener("authorizer", 11000, hcm(decisions, []any{router})), listener("business", 10002, echo)}, "clusters": []any{cluster("tenant-auth", 9091), cluster("cluster_auth_backend_proxy", 11000), cluster("business", 10002)}}}
	edgeFixture, err := os.ReadFile(os.Getenv("ENTRY_EDGE_FIXTURE"))
	if err != nil {
		t.Fatal("generate ENTRY_EDGE_FIXTURE from the L4 TestEntryEdgeIdentityContract first", err)
	}
	var generated struct {
		Listener map[string]any `json:"listener"`
		Clusters []any          `json:"clusters"`
	}
	if err = json.Unmarshal(edgeFixture, &generated); err != nil {
		t.Fatal(err)
	}
	resources := bootstrap["static_resources"].(map[string]any)
	resources["listeners"] = append(resources["listeners"].([]any), generated.Listener)
	resources["clusters"] = append(resources["clusters"].([]any), generated.Clusters...)
	terminalCA := addTerminalEntryFixture(t, bootstrap)
	encoded, err := json.Marshal(bootstrap)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.Chmod(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "bootstrap.json")
	if err = os.WriteFile(path, encoded, 0644); err != nil {
		t.Fatal(err)
	}
	id := docker("run", "-d", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges", "--memory=256m", "--pids-limit=64", "--tmpfs", "/tmp:rw,noexec,nosuid,size=16m", "-p", "127.0.0.1::10000", "-p", "127.0.0.1::10003", "-p", "127.0.0.1::10004", "-p", "127.0.0.1::10005", "-p", "127.0.0.1::10006", "-p", "127.0.0.1::10007", "-v", path+":/bootstrap.json:ro", entryEnvoyImage, "--disable-hot-restart", "--concurrency", "1", "-l", "error", "-c", "/bootstrap.json")
	t.Cleanup(func() {
		cleanup, done := context.WithTimeout(context.Background(), 15*time.Second)
		defer done()
		if t.Failed() {
			diagnostic, _ := exec.CommandContext(cleanup, "docker", "logs", "--tail", "20", id).CombinedOutput()
			lines := strings.Split(string(diagnostic), "\n")
			if len(lines) > 20 {
				lines = lines[len(lines)-20:]
			}
			t.Log("credential-free Envoy diagnostics:\n" + strings.Join(lines, "\n"))
		}
		out, err := exec.CommandContext(cleanup, "docker", "rm", "-f", id).CombinedOutput()
		if err != nil {
			t.Errorf("fixture cleanup failed: %v %s", err, out)
		}
	})
	address := docker("port", id, "10000/tcp")
	client := &http.Client{Timeout: 2 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	endpoint := "http://" + address + "/"
	for i := 0; i < 50; i++ {
		r, e := client.Get(endpoint)
		if e == nil {
			r.Body.Close()
			break
		}
		if i == 49 {
			t.Fatal("Envoy fixture not ready", docker("logs", id))
		}
		time.Sleep(100 * time.Millisecond)
	}
	edgeAddress := docker("port", id, "10003/tcp")
	terminalAddress := docker("port", id, "10005/tcp")
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM([]byte(terminalCA)) {
		t.Fatal("missing terminal fixture CA")
	}
	terminalTransport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, terminalAddress)
	}}
	defer terminalTransport.CloseIdleConnections()
	terminalClient := &http.Client{Transport: terminalTransport, Timeout: 5 * time.Second, CheckRedirect: client.CheckRedirect}
	proxyClient := func(port string) *http.Client {
		address := docker("port", id, port+"/tcp")
		transport := &http.Transport{TLSClientConfig: terminalTransport.TLSClientConfig.Clone(),
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				conn, err := (&net.Dialer{}).DialContext(ctx, network, address)
				if err != nil {
					return nil, err
				}
				// Deliberately claim the same IP for both endpoints. Only the physical
				// connection determines whether this is a local HTTP bridge.
				if _, err = io.WriteString(conn, "PROXY TCP4 192.0.2.17 192.0.2.17 54321 10006\r\n"); err != nil {
					conn.Close()
					return nil, err
				}
				return conn, nil
			}}
		t.Cleanup(transport.CloseIdleConnections)
		return &http.Client{Transport: transport, Timeout: 5 * time.Second, CheckRedirect: client.CheckRedirect}
	}
	localBridgeClient := proxyClient("10007")
	remoteProxyClient := proxyClient("10006")
	for _, path := range []string{"/", "/api/items?view=recent"} {
		t.Run("protected application"+path, func(t *testing.T) {
			const host = "files.alice.olares.local"
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+host+path, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Cookie", "session=alice")
			req.Header.Set("X-Original-URL", "https://untrusted.example.org/")
			resp, err := localBridgeClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK || resp.Header.Get("Location") != "" {
				t.Fatalf("application navigation returned HTTP %d, Location %q", resp.StatusCode, resp.Header.Get("Location"))
			}
			var got map[string]string
			if err := json.NewDecoder(io.LimitReader(resp.Body, 8192)).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if want := "http://" + host + path; got["authurl"] != want || got["user"] != "alice" {
				t.Fatalf("unexpected target or identity after authorization: %v", got)
			}
		})
	}
	for _, tc := range []struct {
		name, host, scheme string
		client             *http.Client
	}{
		{"local-http", "desktop.alice.olares.local", "http", localBridgeClient},
		{"local-direct-https", "desktop.alice.olares.local", "https", terminalClient},
		{"remote-proxy-https", "desktop.alice.olares.local", "https", remoteProxyClient},
		{"formal-host-on-local-proxy", "desktop.alice.example.org", "https", localBridgeClient},
	} {
		t.Run("redirect-scheme/"+tc.name, func(t *testing.T) {
			const target = "/files%20and%20folders?tab=recent"
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+tc.host+target, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("X-Forwarded-Proto", "ftp")
			req.Header.Set("X-Original-URL", "https://untrusted.example.org/")
			resp, err := tc.client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			location, err := resp.Location()
			if err != nil || resp.StatusCode != http.StatusFound {
				t.Fatalf("expected redirect, got HTTP %d: %v", resp.StatusCode, err)
			}
			if got, want := location.Query().Get("rd"), tc.scheme+"://"+tc.host+target; got != want {
				// The stub's header formatter does not URL-encode rd, so read
				// its raw value to check preservation of the escaped path too.
				if location.RawQuery != "rd="+want {
					t.Fatalf("rd = %q (%q), want %q", got, location.RawQuery, want)
				}
			}
		})
	}
	for _, mode := range []string{"trusted-entry", "untrusted-edge", "terminal-l4-public", "terminal-l4-local", "terminal-l4-local-http"} {
		requestHost := "desktop.alice.example.org"
		wantScheme := "https"
		requestClient := client
		endpoint = "http://" + address + "/"
		if mode == "untrusted-edge" {
			endpoint = "http://" + edgeAddress + "/"
			wantScheme = "http"
		}
		if strings.HasPrefix(mode, "terminal-") {
			requestClient = terminalClient
			if mode == "terminal-l4-local" || mode == "terminal-l4-local-http" {
				requestHost = "desktop.alice.olares.local"
			}
			if mode == "terminal-l4-local-http" {
				requestClient = localBridgeClient
				wantScheme = "http"
			}
			endpoint = "https://" + requestHost + "/"
		}
		for _, tc := range []struct {
			name, cookie string
			status       int
			user         string
		}{{"authenticated", "session=alice", 200, "alice"}, {"authorized-different-subject", "session=delegate", 200, "bob"}, {"cross-tenant", "session=bob", 403, ""}, {"expired", "session=expired", 401, ""}, {"anonymous", "", 302, ""}, {"public-anonymous", "policy=public", 200, ""}} {
			t.Run(mode+"/"+tc.name, func(t *testing.T) {
				req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
				req.Host = requestHost
				req.Header.Set("Cookie", tc.cookie)
				req.Header.Add("Remote-User", "mallory")
				req.Header.Add("rEmOtE-uSeR", "admin")
				req.Header.Set("X-BFL-USER", "bob")
				req.Header.Set("Remote-Groups", "admin")
				req.Header.Set("X-Olares-Caller-Jwt", token)
				req.Header.Set("X-Caller-User", "mallory")
				req.Header.Set("X-Caller-Appid", "forged-app")
				req.Header.Set("X-Original-URL", "https://desktop.alice.example.org/")
				req.Header.Set("X-Forwarded-Proto", "gopher")
				if mode != "trusted-entry" {
					req.Header.Set("X-Original-URL", "https://evil.example.org/")
				}
				claimedIP := "198.51.100.17"
				if mode != "trusted-entry" {
					claimedIP = "192.168.0.1"
				}
				req.Header.Set("X-Forwarded-For", claimedIP)
				req.Header.Set("X-Real-IP", "127.0.0.1")
				req.Header.Set("Forwarded", "for=127.0.0.1")
				req.Header.Set("X-Original-Forwarded-For", "127.0.0.1")
				req.Header.Set("X-Envoy-External-Address", "127.0.0.1") // already verified at the mTLS edge, not recomputed as a mesh IP
				resp, err := requestClient.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				body, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
				if err != nil {
					t.Fatal(err)
				}
				if resp.StatusCode != tc.status {
					t.Fatalf("HTTP %d want %d: %s", resp.StatusCode, tc.status, body)
				}
				if tc.status == 302 && strings.HasPrefix(mode, "terminal-") {
					want := "https://auth.alice.example.org/"
					if mode == "terminal-l4-local" || mode == "terminal-l4-local-http" {
						want = "http://auth.alice.olares.local/"
					}
					location, err := url.Parse(resp.Header.Get("Location"))
					if err != nil || location.Scheme+"://"+location.Host+location.Path != want || location.Query().Get("rd") != wantScheme+"://"+requestHost+"/" {
						t.Fatal("terminal login redirect lost the platform domain/protocol contract")
					}
				}
				if tc.status == 200 {
					var got map[string]string
					if err = json.Unmarshal(body, &got); err != nil {
						t.Fatal(err, string(body))
					}
					if got["jwt"] != "" || got["caller"] != "" || got["appid"] != "" {
						t.Fatal("internal credential or caller claim leaked to application")
					}
					if got["user"] != tc.user || got["remote"] != tc.user {
						t.Fatalf("identity crossed the authorization boundary: %v", got)
					}
					if tc.user == "" && got["groups"] != "" {
						t.Fatal("anonymous caller retained spoofed group")
					}
					if got["cookie"] != tc.cookie {
						t.Fatal("credential changed", got)
					}
					if mode == "trusted-entry" && got["xff"] != "198.51.100.17" {
						t.Fatal("trusted forwarding context changed", got)
					}
					if mode != "trusted-entry" {
						if _, err := netip.ParseAddr(got["xff"]); err != nil || got["xff"] == claimedIP || got["xff"] == "127.0.0.1" {
							t.Fatal("forged client IP survived", got)
						}
					}
					if tc.user != "" {
						expectedURL := wantScheme + "://" + requestHost + "/"
						if got["authurl"] != expectedURL {
							t.Fatal("forged original target or forwarded scheme survived", got)
						}
					}
					if tc.user != "" && got["authip"] != got["xff"] {
						t.Fatal("authorizer and business server saw different client IP", got)
					}
				}
			})
		}
	}
	for _, host := range []string{"auth.alice.example.org", "auth.alice.olares.local"} {
		t.Run("terminal-login/"+host, func(t *testing.T) {
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+host+"/", nil)
			resp, err := terminalClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				t.Fatal("login route was incorrectly authenticated", resp.StatusCode)
			}
		})
	}
	internalAddress := docker("port", id, "10004/tcp")
	for _, tc := range []struct {
		name, address, credential string
		status                    int
	}{
		{"external-valid-jwt-entry", address, token, 401},
		{"external-valid-jwt-l4", edgeAddress, token, 401},
		{"internal-injected-jwt", internalAddress, token, 200},
		{"internal-missing-jwt", internalAddress, "", 401},
		{"internal-invalid-jwt", internalAddress, "not-a-signed-token", 401},
		{"internal-recovery", internalAddress, token, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+tc.address+"/shared", nil)
			req.Host = "shared.alice.example.org"
			if tc.credential != "" {
				req.Header.Set("X-Olares-Caller-Jwt", tc.credential)
			}
			req.Header.Set("X-Caller-User", "mallory")
			req.Header.Set("X-Caller-Appid", "forged-app")
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(io.LimitReader(resp.Body, 8192))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.status {
				t.Fatalf("HTTP %d want %d", resp.StatusCode, tc.status)
			}
			if tc.status != 200 {
				// The same token succeeds internally. Require a missing-token denial
				// here, rather than accepting an unrelated L4 authorization failure.
				if tc.credential != "not-a-signed-token" && !strings.Contains(string(body), "Jwt is missing") {
					t.Fatal("request was not rejected for the removed or absent JWT")
				}
				if resp.Header.Get("Content-Type") == "application/json" {
					t.Fatal("denied request reached business echo")
				}
				return
			}
			var got map[string]string
			if err = json.Unmarshal(body, &got); err != nil {
				t.Fatal(err)
			}
			if got["caller"] != "alice" || got["appid"] != "test-client" || got["jwt"] != "" {
				t.Fatal("verified caller identity or credential removal failed", got)
			}
		})
	}

	t.Log(fmt.Sprintf("Envoy contract fixture passed with %s; cleanup registered", entryEnvoyImage))
}
