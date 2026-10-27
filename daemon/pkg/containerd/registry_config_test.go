package containerd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/pelletier/go-toml"
)

func registryTestDirs(t *testing.T) {
	t.Helper()
	oldHosts := certsDDir
	oldVerify := verifyRegistryAuth
	certsDDir = filepath.Join(t.TempDir(), "certs.d")
	// Configuration-only tests do not contact external registries. The real
	// login flow and pre-save behavior are exercised in registry_auth_test.go.
	verifyRegistryAuth = func(context.Context, string, *toml.Tree, *RegistryAuth) error { return nil }
	t.Cleanup(func() {
		certsDDir = oldHosts
		verifyRegistryAuth = oldVerify
	})
}

func registryRequest(t *testing.T, method, body string) (int, string) {
	t.Helper()
	return registryRequestAt(t, method, "/gitlab.example.com", body)
}

func registryRequestAt(t *testing.T, method, path, body string) (int, string) {
	t.Helper()
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		v, err := GetRegistryConfigs(c)
		if err != nil {
			return err
		}
		return c.JSON(v)
	})
	for _, path := range []string{"/:registry", "/+"} {
		app.Get(path, func(c *fiber.Ctx) error {
			v, err := GetRegistryConfig(c)
			if err != nil {
				return err
			}
			return c.JSON(v)
		})
		app.Put(path, func(c *fiber.Ctx) error {
			v, err := UpdateRegistryConfig(c)
			if err != nil {
				return err
			}
			return c.JSON(v)
		})
		app.Delete(path, DeleteRegistryConfig)
	}

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(data)
}

func TestRegistryConfigCredentialsAndMirrors(t *testing.T) {
	registryTestDirs(t)
	code, body := registryRequest(t, "PUT", `{"auth":{"username":"alice","password":"test-token"}}`)
	if code != 200 || strings.Contains(body, "test-token") || !strings.Contains(body, `"has_auth":true`) {
		t.Fatalf("response: %d %s", code, body)
	}
	if err := writeRegistryMirror("gitlab.example.com", []string{"https://z.example.com", "https://a.example.com"}); err != nil {
		t.Fatal(err)
	}
	tree, err := readRegistryTree("gitlab.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !tree.HasPath([]string{"header", "Authorization"}) {
		t.Fatal("mirror update removed credentials")
	}
	mirrors, err := readRegistryMirror("gitlab.example.com")
	if err != nil || len(mirrors) != 2 || mirrors[0] != "https://z.example.com" {
		t.Fatalf("mirror order: %v %v", mirrors, err)
	}
	code, _ = registryRequest(t, "PUT", `{"skip_verify":true}`)
	if code != 200 {
		t.Fatalf("TLS-only update: %d", code)
	}
	if err := writeRegistryMirror("gitlab.example.com", nil); err != nil {
		t.Fatal(err)
	}
	tree, err = readRegistryTree("gitlab.example.com")
	if err != nil || !tree.HasPath([]string{"header", "Authorization"}) {
		t.Fatal("clearing mirrors removed registry auth")
	}
	info, err := os.Stat(registryHostsPath("gitlab.example.com"))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("private config permissions: %v %v", info, err)
	}
	code, body = registryRequest(t, "GET", "")
	if code != 200 || strings.Contains(body, "test-token") || !strings.Contains(body, `"has_auth":true`) {
		t.Fatalf("GET leaked credentials or lost auth: %d %s", code, body)
	}
	code, _ = registryRequest(t, "PUT", `{"auth":{"username":"alice","password":"replacement-token"}}`)
	if code != 200 {
		t.Fatalf("replace auth: %d", code)
	}
	tree, err = readRegistryTree("gitlab.example.com")
	if err != nil {
		t.Fatal(err)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:replacement-token"))
	auth, ok := tree.GetPath([]string{"header", "Authorization"}).([]interface{})
	if !ok || len(auth) != 1 || auth[0] != wantAuth {
		t.Fatal("upstream credential was not replaced")
	}
	code, body = registryRequest(t, "PUT", `{"auth":null}`)
	if code != 200 || !strings.Contains(body, `"has_auth":false`) {
		t.Fatalf("clear auth: %d %s", code, body)
	}
	tree, err = readRegistryTree("gitlab.example.com")
	if err != nil || tree.HasPath([]string{"header", "Authorization"}) {
		t.Fatalf("upstream credentials were not removed: %v", err)
	}
	code, _ = registryRequest(t, "DELETE", "")
	if code != 200 {
		t.Fatalf("delete: %d", code)
	}
	if _, err := os.Stat(registryHostsPath("gitlab.example.com")); !os.IsNotExist(err) {
		t.Fatal("hosts not deleted")
	}
}

func TestRegistryValidation(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../docker.io", "https://ghcr.io", "user@ghcr.io", "ghcr.io/path", "ghcr.io?x", "ghcr.io#x", "a\\b"} {
		if validateRegistryName(name) == nil {
			t.Errorf("accepted %q", name)
		}
	}
	for _, name := range []string{"docker.io", "gitlab.example.com", "localhost:5000", "[::1]:5000"} {
		if err := validateRegistryName(name); err != nil {
			t.Errorf("rejected %q: %v", name, err)
		}
	}
	registryTestDirs(t)
	for _, body := range []string{`{"server":"https://other.example.com"}`, `{"auth":{"username":"alice"}}`, `{"auth":{"username":"a:b","password":"token"}}`} {
		code, _ := registryRequest(t, "PUT", body)
		if code != 400 {
			t.Errorf("accepted invalid body %s: %d", body, code)
		}
	}
}

func TestMirrorUpdatePreservesHostSettings(t *testing.T) {
	registryTestDirs(t)
	os.MkdirAll(registryDir("docker.io"), 0755)
	original := `server = "https://registry-1.docker.io"
[host."https://z.example.com"]
 capabilities = ["pull", "resolve"]
 skip_verify = true
 [host."https://z.example.com".header]
  Authorization = ["Basic dTpw"]
[host."https://a.example.com"]
 capabilities = ["pull"]
 override_path = true
`
	os.WriteFile(registryHostsPath("docker.io"), []byte(original), 0600)
	if err := writeRegistryMirror("docker.io", []string{"https://z.example.com", "https://a.example.com"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(registryHostsPath("docker.io"))
	if err != nil {
		t.Fatal(err)
	}
	tree, err := toml.LoadBytes(data)
	if err != nil {
		t.Fatalf("invalid TOML: %v\n%s", err, data)
	}
	if tree.GetPath([]string{"host", "https://z.example.com", "skip_verify"}) != true || !tree.HasPath([]string{"host", "https://z.example.com", "header", "Authorization"}) {
		t.Fatal("host TLS/auth settings lost")
	}
}

func TestRegistryConfigUpdatesUpstreamEntriesPreservingMirrorList(t *testing.T) {
	registryTestDirs(t)
	os.MkdirAll(registryDir("gitlab.example.com"), 0755)
	original := `server = "https://gitlab.example.com"
[host."https://gitlab.example.com"]
 capabilities = ["pull", "resolve"]
 [host."https://gitlab.example.com".header]
  Authorization = ["Basic b2xkOm9sZA=="]
`
	os.WriteFile(registryHostsPath("gitlab.example.com"), []byte(original), 0600)
	code, _ := registryRequest(t, "PUT", `{"server":"http://gitlab.example.com","auth":{"username":"alice","password":"new-token"}}`)
	if code != 200 {
		t.Fatal(code)
	}
	tree, err := readRegistryTree("gitlab.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if tree.Get("server") != "http://gitlab.example.com" {
		t.Fatal("upstream server not updated")
	}
	endpoints, err := readRegistryMirror("gitlab.example.com")
	if err != nil || !reflect.DeepEqual(endpoints, []string{"https://gitlab.example.com"}) {
		t.Fatalf("omitted endpoints changed: %v %v", endpoints, err)
	}
	if !tree.HasPath([]string{"host", "https://gitlab.example.com", "header", "Authorization"}) {
		t.Fatal("upstream host credentials not updated")
	}
	code, _ = registryRequest(t, "PUT", `{"auth":null}`)
	if code != 200 {
		t.Fatal(code)
	}
	tree, _ = readRegistryTree("gitlab.example.com")
	if tree.HasPath([]string{"host", "https://gitlab.example.com", "header", "Authorization"}) {
		t.Fatal("upstream host still has credentials")
	}
	if tree.Get("server") != "http://gitlab.example.com" {
		t.Fatal("omitted server changed while removing credentials")
	}
}

func TestRegistryConfigPartialUpdates(t *testing.T) {
	for _, tc := range []struct {
		name       string
		body       string
		server     string
		skipVerify bool
		hasAuth    bool
		endpoints  []string
	}{
		{name: "empty object", body: `{}`, server: "http://gitlab.example.com", skipVerify: true, hasAuth: true, endpoints: []string{"https://z.example.com", "https://a.example.com"}},
		{name: "mirrors only", body: `{"endpoint":["https://new.example.com"]}`, server: "http://gitlab.example.com", skipVerify: true, hasAuth: true, endpoints: []string{"https://new.example.com"}},
		{name: "explicit false", body: `{"skip_verify":false}`, server: "http://gitlab.example.com", hasAuth: true, endpoints: []string{"https://z.example.com", "https://a.example.com"}},
		{name: "auth removal", body: `{"auth":null}`, server: "http://gitlab.example.com", skipVerify: true, endpoints: []string{"https://z.example.com", "https://a.example.com"}},
		{name: "auth replacement", body: `{"auth":{"username":"bob","password":"new-token"}}`, server: "http://gitlab.example.com", skipVerify: true, hasAuth: true, endpoints: []string{"https://z.example.com", "https://a.example.com"}},
		{name: "server only", body: `{"server":"https://gitlab.example.com"}`, server: "https://gitlab.example.com", skipVerify: true, hasAuth: true, endpoints: []string{"https://z.example.com", "https://a.example.com"}},
		{name: "empty mirror list", body: `{"endpoint":[]}`, server: "http://gitlab.example.com", skipVerify: true, hasAuth: true, endpoints: []string{}},
		{name: "null mirror list", body: `{"endpoint":null}`, server: "http://gitlab.example.com", skipVerify: true, hasAuth: true, endpoints: []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registryTestDirs(t)
			code, _ := registryRequest(t, "PUT", `{"server":"http://gitlab.example.com","skip_verify":true,"auth":{"username":"alice","password":"old-token"},"endpoint":["https://z.example.com","https://a.example.com"]}`)
			if code != 200 {
				t.Fatal(code)
			}
			before, err := os.ReadFile(registryHostsPath("gitlab.example.com"))
			if err != nil {
				t.Fatal(err)
			}
			code, body := registryRequest(t, "PUT", tc.body)
			if code != 200 {
				t.Fatalf("update: %d %s", code, body)
			}
			var config RegistryConfig
			if err := json.Unmarshal([]byte(body), &config); err != nil {
				t.Fatal(err)
			}
			if config.Server != tc.server || config.SkipVerify != tc.skipVerify || config.HasAuth != tc.hasAuth || !reflect.DeepEqual(config.Endpoints, tc.endpoints) {
				t.Fatalf("unexpected merged config: %+v", config)
			}
			code, got := registryRequest(t, "GET", "")
			if code != 200 || got != body {
				t.Fatalf("GET differs from persisted PUT: %d %s, want %s", code, got, body)
			}
			if tc.body == `{}` {
				after, err := os.ReadFile(registryHostsPath("gitlab.example.com"))
				if err != nil || string(after) != string(before) {
					t.Fatal("empty update rewrote hosts file")
				}
			}
		})
	}
}

func TestRegistryConfigInvalidUpdatePreservesFile(t *testing.T) {
	registryTestDirs(t)
	code, _ := registryRequest(t, "PUT", `{"auth":{"username":"alice","password":"old-token"},"endpoint":["https://mirror.example.com"]}`)
	if code != 200 {
		t.Fatal(code)
	}
	before, err := os.ReadFile(registryHostsPath("gitlab.example.com"))
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{
		`null`, `[]`, `{"server":null}`, `{"skip_verify":null}`, `{"skip_verify":"false"}`,
		`{"endpoint":"https://mirror.example.com"}`, `{"auth":{"username":"bob"}}`,
		`{"auth":null,"endpoint":["https://valid.example.com","invalid"]}`,
		`{"endpoint":["https://duplicate.example.com","https://duplicate.example.com"]}`,
	} {
		code, response := registryRequest(t, "PUT", body)
		if code != 400 {
			t.Fatalf("invalid update accepted: %s, response %d %s", body, code, response)
		}
		after, err := os.ReadFile(registryHostsPath("gitlab.example.com"))
		if err != nil || string(before) != string(after) {
			t.Fatalf("invalid update modified hosts: %s", body)
		}
	}
}

func TestRegistryConfigMirrorReplacementPreservesHostSettings(t *testing.T) {
	registryTestDirs(t)
	if err := os.MkdirAll(registryDir("gitlab.example.com"), 0755); err != nil {
		t.Fatal(err)
	}
	original := `server = "http://gitlab.example.com"
skip_verify = true
[header]
 Authorization = ["Basic YWxpY2U6dG9rZW4="]
[host."https://z.example.com"]
 capabilities = ["pull"]
 skip_verify = true
 override_path = true
 [host."https://z.example.com".header]
  X-Mirror = ["retained"]
[host."https://old.example.com"]
 capabilities = ["pull", "resolve"]
`
	if err := os.WriteFile(registryHostsPath("gitlab.example.com"), []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	code, body := registryRequest(t, "PUT", `{"endpoint":["https://z.example.com","https://new.example.com"]}`)
	if code != 200 {
		t.Fatalf("update: %d %s", code, body)
	}
	tree, err := readRegistryTree("gitlab.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if tree.Get("server") != "http://gitlab.example.com" || tree.Get("skip_verify") != true || !tree.HasPath([]string{"header", "Authorization"}) {
		t.Fatal("mirror update lost upstream entries")
	}
	if tree.GetPath([]string{"host", "https://z.example.com", "skip_verify"}) != true || tree.GetPath([]string{"host", "https://z.example.com", "override_path"}) != true || !tree.HasPath([]string{"host", "https://z.example.com", "header", "X-Mirror"}) {
		t.Fatal("retained mirror settings lost")
	}
	if tree.HasPath([]string{"host", "https://old.example.com"}) {
		t.Fatal("removed mirror retained")
	}
	endpoints, err := readRegistryMirror("gitlab.example.com")
	if err != nil || !reflect.DeepEqual(endpoints, []string{"https://z.example.com", "https://new.example.com"}) {
		t.Fatalf("mirror order: %v %v", endpoints, err)
	}
}

func TestRegistryConfigsListAndDefaults(t *testing.T) {
	registryTestDirs(t)
	code, body := registryRequestAt(t, "GET", "/", "")
	if code != 200 || body != `{}` {
		t.Fatalf("empty list: %d %s", code, body)
	}
	code, body = registryRequestAt(t, "GET", "/docker.io", "")
	if code != 200 || !strings.Contains(body, `"server":"https://registry-1.docker.io"`) || !strings.Contains(body, `"endpoint":[]`) {
		t.Fatalf("missing config defaults: %d %s", code, body)
	}
	for _, name := range []string{"gitlab.example.com", "localhost:5000"} {
		code, body = registryRequestAt(t, "PUT", "/"+name, `{"auth":{"username":"alice","password":"secret-token"}}`)
		if code != 200 {
			t.Fatalf("create: %d %s", code, body)
		}
	}
	if err := os.MkdirAll(registryDir("no-config.example.com"), 0755); err != nil {
		t.Fatal(err)
	}
	code, body = registryRequestAt(t, "GET", "/", "")
	var configs map[string]RegistryConfig
	if err := json.Unmarshal([]byte(body), &configs); err != nil {
		t.Fatal(err)
	}
	if code != 200 || len(configs) != 2 || strings.Contains(body, "secret-token") || strings.Contains(body, `"auth":`) || !configs["gitlab.example.com"].HasAuth || !configs["localhost:5000"].HasAuth {
		t.Fatalf("list response: %d %s", code, body)
	}
}

func TestRegistryConfigClearingMirrorsPreservesImportedUpstream(t *testing.T) {
	registryTestDirs(t)
	if err := os.MkdirAll(registryDir("gitlab.example.com"), 0755); err != nil {
		t.Fatal(err)
	}
	original := `server = "http://gitlab.example.com"
[host."http://gitlab.example.com"]
 capabilities = ["pull", "resolve"]
 skip_verify = true
 ca = "/etc/certs/gitlab.pem"
 [host."http://gitlab.example.com".header]
  Authorization = ["Basic YWxpY2U6dG9rZW4="]
`
	if err := os.WriteFile(registryHostsPath("gitlab.example.com"), []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	code, body := registryRequest(t, "PUT", `{"endpoint":[]}`)
	if code != 200 {
		t.Fatalf("clear mirrors: %d %s", code, body)
	}
	var config RegistryConfig
	if err := json.Unmarshal([]byte(body), &config); err != nil {
		t.Fatal(err)
	}
	if !config.HasAuth || !config.SkipVerify || config.Server != "http://gitlab.example.com" || len(config.Endpoints) != 0 {
		t.Fatalf("upstream entries lost: %+v", config)
	}
	tree, err := readRegistryTree("gitlab.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !tree.HasPath([]string{"header", "Authorization"}) || tree.Get("ca") != "/etc/certs/gitlab.pem" {
		t.Fatal("imported upstream credentials/TLS not preserved")
	}
	code, got := registryRequest(t, "GET", "")
	if code != 200 || got != body {
		t.Fatalf("persisted config mismatch: %d %s", code, got)
	}
}

func TestNormalizeRegistryName(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"gitlab.example.com", "gitlab.example.com"},
		{"https://gitlab.example.com", "gitlab.example.com"},
		{"http://gitlab.example.com:5050", "gitlab.example.com:5050"},
		{"HTTPS://ghcr.io", "ghcr.io"},
		{"oci://localhost:5000", "localhost:5000"},
		{"https://[::1]:5000", "[::1]:5000"},
	} {
		got, err := normalizeRegistryName(tc.input)
		if err != nil || got != tc.want {
			t.Errorf("normalize %q: got %q %v, want %q", tc.input, got, err, tc.want)
		}
	}
	for _, input := range []string{"https://", "://ghcr.io", "1bad://ghcr.io", "https://user@ghcr.io", "https://ghcr.io/path", "http://../docker.io", "https://ghcr.io?x", "https://ghcr.io#x", "http://localhost:bad"} {
		if got, err := normalizeRegistryName(input); err == nil {
			t.Errorf("accepted invalid registry %q as %q", input, got)
		}
	}
}

func TestRegistryConfigSchemeCompatibility(t *testing.T) {
	for _, encoded := range []bool{false, true} {
		for _, address := range []string{"https://gitlab.example.com", "http://gitlab.example.com:5050", "oci://localhost:5000", "https://[::1]:5000"} {
			t.Run(fmt.Sprintf("%s/encoded=%t", address, encoded), func(t *testing.T) {
				registryTestDirs(t)
				registry, err := normalizeRegistryName(address)
				if err != nil {
					t.Fatal(err)
				}
				path := "/" + address
				if encoded {
					path = "/" + url.PathEscape(address)
				}
				code, body := registryRequestAt(t, "PUT", path, `{"skip_verify":true,"auth":{"username":"alice","password":"test-token"}}`)
				if code != 200 {
					t.Fatalf("PUT %s: %d %s", path, code, body)
				}
				if _, err := os.Stat(registryHostsPath(registry)); err != nil {
					t.Fatalf("normalized hosts file missing: %v", err)
				}
				code, plain := registryRequestAt(t, "GET", "/"+registry, "")
				if code != 200 || plain != body {
					t.Fatalf("bare-host GET differs: %d %s", code, plain)
				}
				code, got := registryRequestAt(t, "GET", path, "")
				if code != 200 || got != body {
					t.Fatalf("scheme GET differs: %d %s", code, got)
				}
				var config RegistryConfig
				if err := json.Unmarshal([]byte(body), &config); err != nil {
					t.Fatal(err)
				}
				if config.Server != canonicalServer(registry) {
					t.Fatal("path scheme changed the upstream server")
				}
				code, body = registryRequestAt(t, "GET", "/", "")
				var configs map[string]RegistryConfig
				if err := json.Unmarshal([]byte(body), &configs); err != nil {
					t.Fatal(err)
				}
				if code != 200 || len(configs) != 1 || !configs[registry].HasAuth {
					t.Fatalf("non-canonical list keys: %d %s", code, body)
				}
				code, body = registryRequestAt(t, "DELETE", path, "")
				if code != 200 {
					t.Fatalf("DELETE: %d %s", code, body)
				}
				if _, err := os.Stat(registryHostsPath(registry)); !os.IsNotExist(err) {
					t.Fatalf("normalized hosts not deleted: %v", err)
				}
			})
		}
	}
}

func TestRegistryConfigRejectsInvalidSchemeAddress(t *testing.T) {
	registryTestDirs(t)
	for _, address := range []string{"https://user@ghcr.io", "https://ghcr.io/path", "http://../docker.io", "https://ghcr.io?x", "https://ghcr.io#x"} {
		for _, method := range []string{"GET", "PUT", "DELETE"} {
			code, body := registryRequestAt(t, method, "/"+url.PathEscape(address), `{}`)
			if code != 400 {
				t.Fatalf("invalid address %s %s: %d %s", method, address, code, body)
			}
		}
	}
	if _, err := os.Stat(certsDDir); !os.IsNotExist(err) {
		t.Fatal("invalid registry created configuration files")
	}
}
