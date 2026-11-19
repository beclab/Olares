package containerd

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pelletier/go-toml"
)

func authTestTree(t *testing.T, server string) (string, *toml.Tree) {
	t.Helper()
	u, err := url.Parse(server)
	if err != nil {
		t.Fatal(err)
	}
	tree, err := toml.Load("")
	if err != nil {
		t.Fatal(err)
	}
	tree.Set("server", server)
	return u.Host, tree
}

func requireAuthError(t *testing.T, err error, code int) {
	t.Helper()
	if code == 200 {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	e, ok := err.(*fiber.Error)
	if !ok || e.Code != code {
		t.Fatalf("got %v, want HTTP %d", err, code)
	}
	if strings.Contains(e.Message, "secret-token") || strings.Contains(e.Message, "temporary-token") {
		t.Fatal("validation error leaked credentials")
	}
}

func TestRegistryAuthBasicBeforeSave(t *testing.T) {
	registryTestDirs(t)
	verifyRegistryAuth = validateRegistryAuth
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/v2/" {
			t.Errorf("unexpected validation path %s", r.URL.Path)
		}
		username, password, ok := r.BasicAuth()
		if ok && username == "alice" && password == "secret-token" {
			w.WriteHeader(200)
			return
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
		w.WriteHeader(401)
	}))
	defer upstream.Close()
	registry, _ := authTestTree(t, upstream.URL)
	path := "/" + registry
	request := func(password string, skip bool) string {
		body, err := json.Marshal(map[string]interface{}{
			"server": upstream.URL, "skip_verify": skip,
			"auth":     RegistryAuth{Username: "alice", Password: password},
			"endpoint": []string{"https://mirror.invalid"},
		})
		if err != nil {
			t.Fatal(err)
		}
		return string(body)
	}
	code, body := registryRequestAt(t, "PUT", path, request("wrong", false))
	if code != 400 {
		t.Fatalf("invalid create: %d %s", code, body)
	}
	if _, err := os.Stat(registryDir(registry)); !os.IsNotExist(err) {
		t.Fatal("failed validation created registry configuration")
	}
	code, body = registryRequestAt(t, "PUT", path, request("secret-token", false))
	if code != 200 || strings.Contains(body, "secret-token") {
		t.Fatalf("valid create: %d %s", code, body)
	}
	before, err := os.ReadFile(registryHostsPath(registry))
	if err != nil {
		t.Fatal(err)
	}
	code, body = registryRequestAt(t, "PUT", path, request("wrong", true))
	if code != 400 {
		t.Fatalf("invalid update: %d %s", code, body)
	}
	after, err := os.ReadFile(registryHostsPath(registry))
	if err != nil || string(before) != string(after) {
		t.Fatal("failed authentication changed existing credentials/TLS/mirrors")
	}
	if calls.Load() != 6 {
		t.Fatalf("expected one challenge and one login per auth submission, got %d calls", calls.Load())
	}
	// Updates without credentials and auth:null work even with the server offline.
	upstream.Close()
	for _, body := range []string{`{"skip_verify":true}`, `{"auth":null}`} {
		code, response := registryRequestAt(t, "PUT", path, body)
		if code != 200 {
			t.Fatalf("non-auth update contacted offline registry: %d %s", code, response)
		}
	}
}

func TestRegistryAuthBearer(t *testing.T) {
	for _, tc := range []struct {
		name      string
		password  string
		tokenCode int
		tokenBody string
		probeCode int
		want      int
	}{
		{"token", "secret-token", 200, `{"token":"temporary-token"}`, 200, 200},
		{"access token", "secret-token", 200, `{"access_token":"temporary-token"}`, 200, 200},
		{"invalid credentials", "wrong", 401, `secret-token`, 200, 400},
		{"denied", "secret-token", 403, `secret-token`, 200, 400},
		{"auth service failure", "secret-token", 500, `secret-token`, 200, 502},
		{"missing token", "secret-token", 200, `{}`, 200, 502},
		{"malformed response", "secret-token", 200, `secret-token`, 200, 502},
		{"registry rejects token", "secret-token", 200, `{"token":"temporary-token"}`, 401, 400},
		{"auth redirect", "secret-token", 302, ``, 200, 502},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var tokenCalls, authenticatedProbes atomic.Int32
			var serverURL string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v2/":
					if r.Header.Get("Authorization") == "Bearer temporary-token" {
						authenticatedProbes.Add(1)
						w.WriteHeader(tc.probeCode)
						return
					}
					if r.Header.Get("Authorization") != "" {
						t.Error("credentials sent before challenge")
					}
					w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s/token",service="registry"`, serverURL))
					w.WriteHeader(401)
				case "/token":
					tokenCalls.Add(1)
					username, password, ok := r.BasicAuth()
					if !ok || username != "alice" || password != tc.password || r.URL.Query().Get("service") != "registry" || r.URL.Query().Has("scope") {
						t.Error("token request did not use submitted credentials and unscoped registry service")
					}
					w.Header().Set("Location", serverURL+"/must-not-follow")
					w.WriteHeader(tc.tokenCode)
					fmt.Fprint(w, tc.tokenBody)
				default:
					t.Error("followed authentication redirect")
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			serverURL = server.URL
			registry, tree := authTestTree(t, server.URL)
			err := validateRegistryAuth(context.Background(), registry, tree, &RegistryAuth{"alice", tc.password})
			requireAuthError(t, err, tc.want)
			if tokenCalls.Load() != 1 || (tc.want == 200 && authenticatedProbes.Load() != 1) {
				t.Fatal("token exchange or authenticated registry probe was skipped")
			}
		})
	}
}

func TestRegistryAuthUnverifiableAndUnexpectedResponses(t *testing.T) {
	for _, tc := range []struct {
		name      string
		code      int
		challenge string
		want      int
	}{
		{"anonymous", 200, "", 422},
		{"missing challenge", 401, "", 422},
		{"unsupported challenge", 401, `Digest realm="registry"`, 422},
		{"malformed realm", 401, `Bearer realm=":invalid"`, 502},
		{"relative realm", 401, `Bearer realm="/token"`, 422},
		{"realm userinfo", 401, `Bearer realm="https://user@auth.example.com/token"`, 422},
		{"unsupported scheme", 401, `Bearer realm="file:///tmp/token"`, 422},
		{"registry redirect", 302, "", 502},
		{"not registry", 404, "", 502},
		{"unavailable", 503, "", 502},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registryTestDirs(t)
			verifyRegistryAuth = validateRegistryAuth
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "" {
					t.Error("sent credentials without a supported challenge")
				}
				w.Header().Set("WWW-Authenticate", tc.challenge)
				w.Header().Set("Location", "https://untrusted.invalid/")
				w.WriteHeader(tc.code)
			}))
			defer server.Close()
			registry, _ := authTestTree(t, server.URL)
			body, _ := json.Marshal(map[string]interface{}{"server": server.URL, "auth": RegistryAuth{"alice", "secret-token"}})
			code, response := registryRequestAt(t, "PUT", "/"+registry, string(body))
			if code != tc.want || strings.Contains(response, "secret-token") {
				t.Fatalf("got %d %s, want %d", code, response, tc.want)
			}
			if _, err := os.Stat(registryDir(registry)); !os.IsNotExist(err) {
				t.Fatal("unverified credentials were saved")
			}
		})
	}
}

func TestRegistryAuthTLSAndTimeout(t *testing.T) {
	registryTestDirs(t)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := r.BasicAuth(); ok {
			w.WriteHeader(200)
			return
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
		w.WriteHeader(401)
	}))
	defer server.Close()
	registry, tree := authTestTree(t, server.URL)
	credentials := &RegistryAuth{"alice", "secret-token"}
	requireAuthError(t, validateRegistryAuth(context.Background(), registry, tree, credentials), 502)
	tree.Set("skip_verify", true)
	requireAuthError(t, validateRegistryAuth(context.Background(), registry, tree, credentials), 200)
	tree.Set("skip_verify", false)
	if err := os.MkdirAll(registryDir(registry), 0755); err != nil {
		t.Fatal(err)
	}
	caPath := filepath.Join(registryDir(registry), "ca.pem")
	if err := os.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	for _, hostTable := range []bool{false, true} {
		if hostTable {
			tree.Delete("ca")
			tree.SetPath([]string{"host", server.URL, "ca"}, "ca.pem")
		} else {
			tree.Set("ca", "ca.pem")
		}
		requireAuthError(t, validateRegistryAuth(context.Background(), registry, tree, credentials), 200)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	<-ctx.Done()
	requireAuthError(t, validateRegistryAuth(ctx, registry, tree, credentials), 504)
}

func TestRegistryAuthCrossOriginBearer(t *testing.T) {
	for _, tc := range []struct {
		name     string
		tls      bool
		password string
		want     int
	}{
		{"HTTP", false, "secret-token", 200},
		{"HTTPS", true, "secret-token", 200},
		{"invalid credentials", true, "wrong", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registryTestDirs(t)
			verifyRegistryAuth = validateRegistryAuth
			var tokenCalls, authenticatedProbes atomic.Int32
			authServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tokenCalls.Add(1)
				username, password, ok := r.BasicAuth()
				if r.URL.Path != "/custom/auth" || r.URL.Query().Get("tenant") != "olares" ||
					r.URL.Query().Get("service") != "registry" || r.URL.Query().Has("scope") ||
					!ok || username != "alice" || password != tc.password {
					t.Error("authentication service did not receive the challenge URL and submitted credentials")
					w.WriteHeader(400)
					return
				}
				if password != "secret-token" {
					w.WriteHeader(401)
					return
				}
				fmt.Fprint(w, `{"token":"temporary-token"}`)
			}))
			if tc.tls {
				authServer.StartTLS()
			} else {
				authServer.Start()
			}
			defer authServer.Close()
			// Use both a different hostname and port from the registry.
			realm := strings.Replace(authServer.URL, "127.0.0.1", "localhost", 1) + "/custom/auth?tenant=olares"
			upstream := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/" {
					t.Error("unexpected registry probe path")
				}
				if r.Header.Get("Authorization") == "Bearer temporary-token" {
					authenticatedProbes.Add(1)
					w.WriteHeader(200)
					return
				}
				if r.Header.Get("Authorization") != "" {
					t.Error("password was sent to the registry instead of its authentication service")
				}
				w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s",service="registry"`, realm))
				w.WriteHeader(401)
			}))
			if tc.tls {
				upstream.StartTLS()
			} else {
				upstream.Start()
			}
			defer upstream.Close()
			registry, _ := authTestTree(t, upstream.URL)
			body, _ := json.Marshal(map[string]interface{}{
				"server": upstream.URL, "skip_verify": tc.tls,
				"auth": RegistryAuth{"alice", tc.password},
			})
			code, response := registryRequestAt(t, "PUT", "/"+registry, string(body))
			if code != tc.want || strings.Contains(response, "secret-token") || strings.Contains(response, "temporary-token") {
				t.Fatalf("cross-origin login: %d %s, want %d", code, response, tc.want)
			}
			if tokenCalls.Load() != 1 {
				t.Fatal("cross-origin authentication service was not called exactly once")
			}
			if tc.want == 200 {
				if authenticatedProbes.Load() != 1 {
					t.Fatal("registry did not receive the exchanged token")
				}
				if _, err := os.Stat(registryHostsPath(registry)); err != nil {
					t.Fatalf("verified configuration was not saved: %v", err)
				}
			} else if _, err := os.Stat(registryDir(registry)); !os.IsNotExist(err) {
				t.Fatal("rejected cross-origin credentials were saved")
			}
		})
	}
}

func TestRegistryAuthRejectsHTTPSDowngrade(t *testing.T) {
	var tokenCalls atomic.Int32
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls.Add(1)
		w.WriteHeader(500)
	}))
	defer authServer.Close()
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s/token",service="registry"`, authServer.URL))
		w.WriteHeader(401)
	}))
	defer upstream.Close()
	registry, tree := authTestTree(t, upstream.URL)
	tree.Set("skip_verify", true)
	err := validateRegistryAuth(context.Background(), registry, tree, &RegistryAuth{"alice", "secret-token"})
	requireAuthError(t, err, 422)
	if tokenCalls.Load() != 0 {
		t.Fatal("credentials were sent to a downgraded HTTP authentication service")
	}
}

func TestRegistryAuthRealmURLValidation(t *testing.T) {
	for _, tc := range []struct {
		server string
		realm  string
		want   bool
	}{
		{"https://ghcr.io", "https://ghcr.io/token", true},
		{"https://registry-1.docker.io", "https://auth.docker.io/token", true},
		{"https://registry.gitlab.com", "https://gitlab.com/jwt/auth", true},
		{"http://localhost:5000", "http://localhost:5000/token", true},
		{"https://ghcr.io", "http://ghcr.io/token", false},
		{"https://ghcr.io", "https://ghcr.io:444/token", true},
		{"https://ghcr.io", "https://user@ghcr.io/token", false},
		{"https://registry.example.com:5050", "https://gitlab.example.com/jwt/auth", true},
		{"https://registry.example.com", "https://auth.other.example/custom/token?service=registry", true},
		{"http://registry.example.com:5000", "http://auth.example.com:8080/token", true},
		{"http://registry.example.com:5000", "https://auth.example.com/token", true},
		{"https://registry.example.com", "/token", false},
		{"https://registry.example.com", "//auth.example.com/token", false},
		{"https://registry.example.com", "https://auth.example.com/token#fragment", false},
		{"https://registry.example.com", "file:///tmp/token", false},
		{"https://registry.example.com", "https://:443/token", false},
	} {
		server, _ := url.Parse(tc.server)
		realm, _ := url.Parse(tc.realm)
		if got := validRegistryAuthRealm(server, realm); got != tc.want {
			t.Errorf("realm URL validation %s -> %s: got %t", tc.server, tc.realm, got)
		}
	}
}
