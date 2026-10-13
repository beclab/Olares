package registry

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/remotes/docker"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func writeTestHosts(t *testing.T, root, namespace, data string) {
	t.Helper()
	if !strings.Contains(data, "[host") {
		data += "\n[host]\n"
	}
	dir := filepath.Join(root, namespace)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "hosts.toml"), []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestPrivateRegistryBasicAndBearer(t *testing.T) {
	for _, bearer := range []bool{false, true} {
		t.Run(fmt.Sprint("bearer=", bearer), func(t *testing.T) {
			var tokenRequests atomic.Int32
			config := `{"architecture":"amd64","os":"linux","config":{"Env":["A=B"]},"rootfs":{"type":"layers","diff_ids":[]}}`
			configDigest := digest.FromString(config)
			manifest := fmt.Sprintf(`{"schemaVersion":2,"mediaType":%q,"config":{"mediaType":%q,"digest":%q,"size":%d},"layers":[]}`, ocispec.MediaTypeImageManifest, ocispec.MediaTypeImageConfig, configDigest, len(config))
			manifestDigest := digest.FromString(manifest)
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/token" {
					u, p, ok := r.BasicAuth()
					if r.Method == "POST" {
						r.ParseForm()
						u, p, ok = r.Form.Get("username"), r.Form.Get("password"), true
					}
					if !ok || u != "alice" || p != "test-token" {
						t.Error("token endpoint missing credentials")
						w.WriteHeader(401)
						return
					}
					tokenRequests.Add(1)
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprint(w, `{"token":"access-token","access_token":"access-token","expires_in":300}`)
					return
				}
				authorized := false
				if bearer {
					authorized = r.Header.Get("Authorization") == "Bearer access-token"
				} else {
					u, p, ok := r.BasicAuth()
					authorized = ok && u == "alice" && p == "test-token"
				}
				if !authorized {
					if bearer {
						w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s/token",service="registry",scope="repository:team/app:pull"`, server.URL))
					} else {
						w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
					}
					w.WriteHeader(401)
					return
				}
				switch {
				case strings.Contains(r.URL.Path, "/manifests/"):
					w.Header().Set("Content-Type", ocispec.MediaTypeImageManifest)
					w.Header().Set("Docker-Content-Digest", manifestDigest.String())
					w.Header().Set("Content-Length", fmt.Sprint(len(manifest)))
					if r.Method != "HEAD" {
						fmt.Fprint(w, manifest)
					}
				case strings.HasSuffix(r.URL.Path, configDigest.String()):
					fmt.Fprint(w, config)
				default:
					w.WriteHeader(404)
				}
			}))
			defer server.Close()
			root := t.TempDir()
			host, _ := url.Parse(server.URL)
			basic := "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:test-token"))
			writeTestHosts(t, root, host.Host, fmt.Sprintf("server = %q\n[header]\nAuthorization = [%q]\n", server.URL, basic))
			resolver := docker.NewResolver(docker.ResolverOptions{Hosts: registryHosts(context.Background(), root)})
			info, layers, err := inspectWithResolver(context.Background(), host.Host+"/team/app:v1", resolver)
			if err != nil {
				t.Fatal(err)
			}
			if info.Architecture != "amd64" || len(layers) != 0 || len(info.Config.Env) != 1 {
				t.Fatalf("unexpected metadata: %+v", info)
			}
			if bearer && tokenRequests.Load() == 0 {
				t.Fatal("Bearer token endpoint was not used")
			}
		})
	}
}

func TestRegistryHostsReloadAndIsolation(t *testing.T) {
	root := t.TempDir()
	hosts := registryHosts(context.Background(), root)
	writeTestHosts(t, root, "gitlab.example.com", `server = "https://gitlab.example.com"
[header]
 Authorization = ["Basic YWxpY2U6dGVzdA=="]
[host."https://mirror.example.com"]
 capabilities = ["pull", "resolve"]
`)
	configured, err := hosts("gitlab.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(configured) != 2 || configured[0].Host != "mirror.example.com" || configured[0].Header.Get("Authorization") != "" {
		t.Fatal("upstream credentials leaked to mirror")
	}
	unrelated, err := hosts("ghcr.io")
	if err != nil || unrelated[0].Header.Get("Authorization") != "" {
		t.Fatal("credentials leaked across registries")
	}
	writeTestHosts(t, root, "gitlab.example.com", `server = "http://gitlab.example.com"`)
	configured, err = hosts("gitlab.example.com")
	if err != nil || len(configured) != 1 || configured[0].Scheme != "http" {
		t.Fatal("config update not reloaded")
	}
	os.RemoveAll(filepath.Join(root, "gitlab.example.com"))
	configured, err = hosts("gitlab.example.com")
	if err != nil || configured[0].Scheme != "https" {
		t.Fatal("removed config remained cached")
	}
}

func TestMirrorPriorityPathTLSAndFallback(t *testing.T) {
	root := t.TempDir()
	writeTestHosts(t, root, "ghcr.io", `server = "https://ghcr.io"
[host."https://z.example.com/cache/v2"]
 capabilities = ["pull", "resolve"]
 override_path = true
 skip_verify = true
[host."http://a.example.com"]
 capabilities = ["pull"]
`)
	hosts, err := registryHosts(context.Background(), root)("ghcr.io")
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 3 || hosts[0].Host != "z.example.com" || hosts[0].Path != "/cache/v2" || hosts[1].Scheme != "http" || hosts[2].Host != "ghcr.io" {
		t.Fatalf("unexpected hosts: %+v", hosts)
	}
	transport := hosts[0].Client.Transport.(*http.Transport)
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("skip_verify not applied")
	}
	if hosts[1].Capabilities&docker.HostCapabilityResolve != 0 {
		t.Fatal("pull-only capability ignored")
	}
}

func TestInspectMultiArch(t *testing.T) {
	platform := platforms.DefaultSpec()
	config := fmt.Sprintf(`{"architecture":%q,"os":%q,"config":{},"rootfs":{"type":"layers","diff_ids":[]}}`, platform.Architecture, platform.OS)
	configDigest := digest.FromString(config)
	manifest := fmt.Sprintf(`{"schemaVersion":2,"mediaType":%q,"config":{"mediaType":%q,"digest":%q,"size":%d},"layers":[]}`, ocispec.MediaTypeImageManifest, ocispec.MediaTypeImageConfig, configDigest, len(config))
	manifestDigest := digest.FromString(manifest)
	index := fmt.Sprintf(`{"schemaVersion":2,"mediaType":%q,"manifests":[{"mediaType":%q,"digest":%q,"size":%d,"platform":{"architecture":"invalid","os":"invalid"}},{"mediaType":%q,"digest":%q,"size":%d,"platform":{"architecture":%q,"os":%q}}]}`, ocispec.MediaTypeImageIndex, ocispec.MediaTypeImageManifest, digest.FromString("wrong"), 5, ocispec.MediaTypeImageManifest, manifestDigest, len(manifest), platform.Architecture, platform.OS)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, media := index, ocispec.MediaTypeImageIndex
		if strings.HasSuffix(r.URL.Path, manifestDigest.String()) {
			data, media = manifest, ocispec.MediaTypeImageManifest
		}
		if strings.HasSuffix(r.URL.Path, configDigest.String()) {
			data, media = config, ocispec.MediaTypeImageConfig
		}
		w.Header().Set("Content-Type", media)
		w.Header().Set("Content-Length", fmt.Sprint(len(data)))
		w.Header().Set("Docker-Content-Digest", digest.FromString(data).String())
		if r.Method != "HEAD" {
			fmt.Fprint(w, data)
		}
	}))
	defer server.Close()
	host, _ := url.Parse(server.URL)
	root := t.TempDir()
	writeTestHosts(t, root, host.Host, fmt.Sprintf("server = %q\n", server.URL))
	resolver := docker.NewResolver(docker.ResolverOptions{Hosts: registryHosts(context.Background(), root)})
	image, _, err := inspectWithResolver(context.Background(), host.Host+"/app:v1", resolver)
	if err != nil {
		t.Fatal(err)
	}
	if image.OS != platform.OS || image.Architecture != platform.Architecture {
		t.Fatalf("wrong platform: %+v", image)
	}
}

func TestResolveUsesRegistryNamespace(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	namespaces := []string{"docker.io", "ghcr.io", "gitlab.example.com:5000", "gitea.example.com"}
	manifest := `{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json","config":{"mediaType":"application/vnd.oci.image.config.v1+json","digest":"sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","size":2},"layers":[]}`
	for _, namespace := range namespaces {
		token := "token-for-" + namespace
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok || username != "alice" || password != token {
				w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
				w.WriteHeader(401)
				return
			}
			if r.URL.Query().Get("ns") != namespace {
				t.Errorf("registry hint: %q, want %q", r.URL.Query().Get("ns"), namespace)
			}
			if r.URL.Path != "/v2/team/app/manifests/v1" {
				t.Errorf("repository path: %q", r.URL.Path)
			}
			w.Header().Set("Content-Type", ocispec.MediaTypeImageManifest)
			w.Header().Set("Content-Length", fmt.Sprint(len(manifest)))
			w.Header().Set("Docker-Content-Digest", digest.FromString(manifest).String())
			if r.Method != "HEAD" {
				fmt.Fprint(w, manifest)
			}
		}))
		t.Cleanup(server.Close)
		basic := "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:"+token))
		writeTestHosts(t, root, namespace, fmt.Sprintf("server = %q\n[header]\nAuthorization = [%q]\n", server.URL, basic))
	}
	resolver := docker.NewResolver(docker.ResolverOptions{Hosts: registryHosts(ctx, root)})
	for _, namespace := range namespaces {
		ref := namespace + "/team/app:v1"
		name, _, err := resolver.Resolve(ctx, ref)
		if err != nil {
			t.Errorf("resolve %s: %v", ref, err)
			continue
		}
		if name != ref {
			t.Errorf("registry identity changed: got %q, want %q", name, ref)
		}
	}
}
