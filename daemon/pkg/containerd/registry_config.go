package containerd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/pelletier/go-toml"
)

var registryConfigMu sync.Mutex
var certsDDir = CertsDDir

// RegistryConfig describes a registry upstream and its ordered pull-through mirrors.
// Password is write-only. Omitted auth preserves credentials; auth:null removes them.
type RegistryConfig struct {
	Server     string   `json:"server"`
	SkipVerify bool     `json:"skip_verify"`
	HasAuth    bool     `json:"has_auth"`
	Endpoints  []string `json:"endpoint"`
}

type RegistryAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func validateRegistryName(registry string) error {
	u, err := url.Parse("https://" + registry)
	if err != nil || u == nil || u.Host != registry || u.Hostname() == "" || u.User != nil ||
		u.Path != "" || u.RawQuery != "" || u.Fragment != "" || strings.ContainsAny(registry, "/\\ \t\r\n") || registry == "." || registry == ".." || registry == "_default" {
		return fiber.NewError(fiber.StatusBadRequest, "registry must be a hostname with an optional port")
	}
	return nil
}

// normalizeRegistryName accepts URL-style input but keeps the registry namespace
// as host[:port]. Scheme is only stripped; it does not change the upstream URL.
func normalizeRegistryName(registry string) (string, error) {
	if _, remainder, found := strings.Cut(registry, "://"); found {
		parsed, err := url.Parse(registry)
		if err != nil || parsed.Scheme == "" {
			return "", fiber.NewError(fiber.StatusBadRequest, "invalid registry address")
		}
		registry = remainder
	}
	if err := validateRegistryName(registry); err != nil {
		return "", err
	}
	return registry, nil
}

func requestRegistryName(ctx *fiber.Ctx) (string, error) {
	registry := ctx.Params(ParamRegistryName, ctx.Params("+"))
	registry, err := url.PathUnescape(registry)
	if err != nil {
		return "", fiber.NewError(fiber.StatusBadRequest, "invalid registry address encoding")
	}
	return normalizeRegistryName(registry)
}

func readRegistryTree(registry string) (*toml.Tree, error) {
	data, err := os.ReadFile(registryHostsPath(registry))
	if os.IsNotExist(err) {
		return toml.Load("")
	}
	if err != nil {
		return nil, err
	}
	return toml.LoadBytes(data)
}

// Preserve table order: marshaling the whole tree would alphabetize mirror hosts.
func saveRegistryTree(registry string, tree *toml.Tree, endpoints []string) error {
	hostTree, _ := tree.Get("host").(*toml.Tree)
	root, err := toml.Load("")
	if err != nil {
		return err
	}
	for _, key := range tree.Keys() {
		if key != "host" {
			root.SetPath([]string{key}, tree.GetPath([]string{key}))
		}
	}
	data, err := root.ToTomlString()
	if err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString(data)
	b.WriteString("\n[host]\n")
	for _, endpoint := range endpoints {
		var host *toml.Tree
		if hostTree != nil {
			host, _ = hostTree.GetPath([]string{endpoint}).(*toml.Tree)
		}
		if host == nil {
			host, _ = toml.Load("")
			host.Set("capabilities", []string{"pull", "resolve"})
		}
		entry, _ := toml.Load("")
		entry.SetPath([]string{"host", endpoint}, host)
		s, err := entry.ToTomlString()
		if err != nil {
			return err
		}
		b.WriteString("\n")
		b.WriteString(strings.Replace(s, "[host]\n", "", 1))
	}
	if err := os.MkdirAll(registryDir(registry), 0755); err != nil {
		return err
	}
	// Rename makes credential/config changes atomic for CRI and image-service.
	file, err := os.CreateTemp(registryDir(registry), ".hosts-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.WriteString(b.String()); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), registryHostsPath(registry))
}

func registryConfigView(registry string, tree *toml.Tree, endpoints []string) *RegistryConfig {
	server, _ := tree.Get("server").(string)
	if server == "" {
		server = canonicalServer(registry)
	}
	skipVerify, _ := tree.Get("skip_verify").(bool)
	hasAuth := tree.HasPath([]string{"header", "Authorization"})
	if host, ok := tree.GetPath([]string{"host", server}).(*toml.Tree); ok {
		hasAuth = hasAuth || host.HasPath([]string{"header", "Authorization"})
		if value, ok := host.Get("skip_verify").(bool); ok {
			skipVerify = value
		}
	}
	if endpoints == nil {
		endpoints = []string{}
	}
	return &RegistryConfig{Server: server, SkipVerify: skipVerify, HasAuth: hasAuth, Endpoints: endpoints}
}

func readRegistryConfig(registry string) (*RegistryConfig, error) {
	tree, err := readRegistryTree(registry)
	if err != nil {
		return nil, err
	}
	endpoints, err := readRegistryMirror(registry)
	if err != nil {
		return nil, err
	}
	return registryConfigView(registry, tree, endpoints), nil
}

// GetRegistryConfigs returns saved registry configuration without exposing credentials.
func GetRegistryConfigs(ctx *fiber.Ctx) (map[string]*RegistryConfig, error) {
	registryConfigMu.Lock()
	defer registryConfigMu.Unlock()
	configs := make(map[string]*RegistryConfig)
	entries, err := os.ReadDir(certsDDir)
	if os.IsNotExist(err) {
		return configs, nil
	}
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() || validateRegistryName(entry.Name()) != nil {
			continue
		}
		if _, err := os.Stat(registryHostsPath(entry.Name())); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		config, err := readRegistryConfig(entry.Name())
		if err != nil {
			return nil, err
		}
		configs[entry.Name()] = config
	}
	return configs, nil
}

func GetRegistryConfig(ctx *fiber.Ctx) (*RegistryConfig, error) {
	registry, err := requestRegistryName(ctx)
	if err != nil {
		return nil, err
	}
	registryConfigMu.Lock()
	defer registryConfigMu.Unlock()
	return readRegistryConfig(registry)
}

// UpdateRegistryConfig replaces supplied entries and preserves omitted entries.
// Auth is a complete credential pair or null; endpoint replaces the ordered list.
func UpdateRegistryConfig(ctx *fiber.Ctx) (*RegistryConfig, error) {
	registry, err := requestRegistryName(ctx)
	if err != nil {
		return nil, err
	}
	var request struct {
		Server     *string       `json:"server"`
		SkipVerify *bool         `json:"skip_verify"`
		Auth       *RegistryAuth `json:"auth"`
		Endpoints  []string      `json:"endpoint"`
	}
	var fields map[string]json.RawMessage
	if err := ctx.BodyParser(&request); err != nil {
		return nil, fiber.NewError(400, "invalid request body")
	}
	if err := ctx.BodyParser(&fields); err != nil || fields == nil {
		return nil, fiber.NewError(400, "request body must be an object")
	}
	if _, supplied := fields["server"]; supplied && request.Server == nil {
		return nil, fiber.NewError(400, "server must be a string")
	}
	if _, supplied := fields["skip_verify"]; supplied && request.SkipVerify == nil {
		return nil, fiber.NewError(400, "skip_verify must be a boolean")
	}
	expected, _ := url.Parse(canonicalServer(registry))
	var server *url.URL
	if request.Server != nil {
		value := *request.Server
		if value == "" {
			value = canonicalServer(registry)
		}
		normalized, err := normalizeEndpoint(value)
		if err != nil {
			return nil, fiber.NewError(400, "invalid registry server")
		}
		server, _ = url.Parse(normalized)
		// Credentials belong to this registry only, never to an unrelated upstream.
		if server.Host != expected.Host || (server.Path != "" && server.Path != "/") || server.User != nil || server.RawQuery != "" || server.Fragment != "" {
			return nil, fiber.NewError(400, "server must match the registry hostname")
		}
	}
	if request.Auth != nil && (request.Auth.Username == "" || request.Auth.Password == "" || strings.Contains(request.Auth.Username, ":")) {
		return nil, fiber.NewError(400, "auth requires a username and password/token")
	}
	var replacements []string
	if _, supplied := fields["endpoint"]; supplied {
		replacements = make([]string, 0, len(request.Endpoints))
		seen := make(map[string]bool)
		for _, endpoint := range request.Endpoints {
			normalized, err := normalizeEndpoint(endpoint)
			if err != nil {
				return nil, fiber.NewError(400, err.Error())
			}
			if seen[normalized] {
				return nil, fiber.NewError(400, "duplicate mirror endpoint")
			}
			seen[normalized] = true
			replacements = append(replacements, normalized)
		}
	}
	registryConfigMu.Lock()
	defer registryConfigMu.Unlock()
	tree, err := readRegistryTree(registry)
	if err != nil {
		return nil, err
	}
	endpoints, err := readRegistryMirror(registry)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return registryConfigView(registry, tree, endpoints), nil
	}
	if replacements != nil {
		// Imported hosts files may store upstream credentials/TLS only in a
		// host table. Clearing mirrors must not clear those omitted entries.
		upstream := registryConfigView(registry, tree, endpoints).Server
		retained := false
		for _, endpoint := range replacements {
			if endpoint == upstream {
				retained = true
				break
			}
		}
		if host, ok := tree.GetPath([]string{"host", upstream}).(*toml.Tree); ok && !retained {
			for _, key := range []string{"skip_verify", "ca", "client", "dial_timeout"} {
				if host.Has(key) {
					tree.Set(key, host.Get(key))
				}
			}
			if headers, ok := host.Get("header").(*toml.Tree); ok {
				for _, key := range headers.Keys() {
					tree.SetPath([]string{"header", key}, headers.GetPath([]string{key}))
				}
			}
		}
	}
	if server != nil {
		tree.Set("server", server.String())
	} else if !tree.Has("server") {
		tree.Set("server", canonicalServer(registry))
	}
	if request.SkipVerify != nil {
		tree.Set("skip_verify", *request.SkipVerify)
	}
	if _, supplied := fields["auth"]; supplied {
		if request.Auth == nil {
			tree.DeletePath([]string{"header", "Authorization"})
		} else {
			value := "Basic " + base64.StdEncoding.EncodeToString([]byte(request.Auth.Username+":"+request.Auth.Password))
			tree.SetPath([]string{"header", "Authorization"}, []string{value})
		}
	}
	// Existing files may keep upstream settings in a host table. Apply only
	// supplied upstream entries; unrelated mirror settings remain unchanged.
	for _, endpoint := range endpoints {
		ep, err := url.Parse(endpoint)
		if err != nil || ep.Host != expected.Host {
			continue
		}
		if host, ok := tree.GetPath([]string{"host", endpoint}).(*toml.Tree); ok {
			if request.SkipVerify != nil {
				host.Set("skip_verify", *request.SkipVerify)
			}
			if _, supplied := fields["auth"]; supplied {
				if request.Auth == nil {
					host.DeletePath([]string{"header", "Authorization"})
				} else {
					host.SetPath([]string{"header", "Authorization"}, tree.GetPath([]string{"header", "Authorization"}))
				}
			}
		}
	}
	if replacements != nil {
		endpoints = replacements
	}
	if request.Auth != nil {
		if err := verifyRegistryAuth(ctx.UserContext(), registry, tree, request.Auth); err != nil {
			return nil, err
		}
	}
	if err := saveRegistryTree(registry, tree, endpoints); err != nil {
		return nil, fmt.Errorf("save registry config: %w", err)
	}
	return readRegistryConfig(registry)
}

func DeleteRegistryConfig(ctx *fiber.Ctx) error {
	registry, err := requestRegistryName(ctx)
	if err != nil {
		return err
	}
	registryConfigMu.Lock()
	defer registryConfigMu.Unlock()
	return os.RemoveAll(filepath.Join(certsDDir, registry))
}
