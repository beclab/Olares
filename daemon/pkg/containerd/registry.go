package containerd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/containerd/containerd/reference"
	"github.com/gofiber/fiber/v2"
	"github.com/pelletier/go-toml"
	"k8s.io/klog/v2"
)

const (
	// CertsDDir is containerd's registry config_path (config v3). Each subdir is a
	// registry namespace (e.g. docker.io) holding a hosts.toml with its mirror
	// hosts. This replaces the deprecated inline registry.mirrors in config.toml,
	// which containerd 2.x ignores once config_path is set.
	CertsDDir = "/etc/containerd/certs.d"

	ParamRegistryName   = "registry"
	DefaultRegistryName = "docker.io"
)

func registryDir(registry string) string { return filepath.Join(CertsDDir, registry) }

func registryHostsPath(registry string) string {
	return filepath.Join(registryDir(registry), "hosts.toml")
}

// canonicalServer is the upstream server URL for a registry namespace; it is the
// hosts.toml `server` fallback tried after the mirror hosts. docker.io is special:
// its registry API is served by registry-1.docker.io (docker.io itself does not
// serve /v2). containerd only auto-maps this when `server` is empty, so we spell
// it out because we always write `server`.
func canonicalServer(registry string) string {
	if registry == DefaultRegistryName {
		return "https://registry-1.docker.io"
	}
	return "https://" + registry
}

// readRegistryMirror returns the ordered mirror host endpoints for a registry
// from certs.d/<registry>/hosts.toml. A missing file yields nil (no error).
func readRegistryMirror(registry string) ([]string, error) {
	data, err := os.ReadFile(registryHostsPath(registry))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return parseHostsTOMLMirrors(data)
}

// parseHostsTOMLMirrors returns the [host."<url>"] mirror endpoints from a
// containerd certs.d hosts.toml, ordered by their position in the file (highest
// priority first). TOML tables are unordered as a data model, so — like
// containerd's own resolver — we recover precedence from each host table's source
// line number rather than map iteration order.
func parseHostsTOMLMirrors(data []byte) ([]string, error) {
	tree, err := toml.LoadBytes(data)
	if err != nil {
		return nil, err
	}
	hostTree, ok := tree.Get("host").(*toml.Tree)
	if !ok {
		return nil, nil
	}
	hosts := hostTree.Keys()
	// Note: use GetPath (single path element), not Get — Get treats the key as a
	// dot-separated path, which would mangle host URLs that contain dots/colons.
	sort.SliceStable(hosts, func(i, j int) bool {
		ti, _ := hostTree.GetPath([]string{hosts[i]}).(*toml.Tree)
		tj, _ := hostTree.GetPath([]string{hosts[j]}).(*toml.Tree)
		if ti == nil || tj == nil {
			return false
		}
		return ti.Position().Line < tj.Position().Line
	})
	return hosts, nil
}

// writeRegistryMirror writes certs.d/<registry>/hosts.toml with the given ordered
// endpoints as [host] entries (highest priority first) and the canonical upstream
// as `server`. Empty endpoints removes the hosts.toml so default resolution
// applies. hosts.toml is consulted per-pull, so no containerd restart is needed.
func writeRegistryMirror(registry string, endpoints []string) error {
	if len(endpoints) == 0 {
		if err := os.Remove(registryHostsPath(registry)); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(registryDir(registry), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "server = %q\n", canonicalServer(registry))
	for _, ep := range endpoints {
		fmt.Fprintf(&b, "\n[host.%q]\n  capabilities = [\"pull\", \"resolve\"]\n", ep)
	}
	return os.WriteFile(registryHostsPath(registry), []byte(b.String()), 0o644)
}

func normalizeEndpoint(endpoint string) (string, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return "", fmt.Errorf("endpoint is required")
	}
	u, err := url.ParseRequestURI(endpoint)
	if err != nil || u == nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("invalid mirror endpoint: %s", endpoint)
	}
	return u.String(), nil
}

// GetRegistryMirrors returns all configured registry mirrors keyed by registry
// namespace (from every certs.d/<registry>/hosts.toml).
func GetRegistryMirrors(ctx *fiber.Ctx) (map[string]Mirror, error) {
	entries, err := os.ReadDir(CertsDDir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]Mirror{}, nil
		}
		return nil, err
	}
	mirrors := make(map[string]Mirror)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		eps, err := readRegistryMirror(e.Name())
		if err != nil {
			klog.Errorf("read hosts.toml for %s failed: %v", e.Name(), err)
			continue
		}
		if len(eps) == 0 {
			continue
		}
		mirrors[e.Name()] = Mirror{Endpoints: eps}
	}
	return mirrors, nil
}

// GetRegistryMirror returns the mirror config for a single registry (defaulting
// to docker.io).
func GetRegistryMirror(ctx *fiber.Ctx) (*Mirror, error) {
	registry := ctx.Params(ParamRegistryName)
	if registry == "" {
		registry = DefaultRegistryName
	}
	eps, err := readRegistryMirror(registry)
	if err != nil {
		return nil, err
	}
	return &Mirror{Endpoints: eps}, nil
}

// UpdateRegistryMirror replaces the mirror host list for a registry. The request
// body is a Mirror ({"endpoint": [...]}), matching the previous interface.
func UpdateRegistryMirror(ctx *fiber.Ctx) (*Mirror, error) {
	registry := ctx.Params(ParamRegistryName)
	if registry == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "registry name is required")
	}
	var mirror Mirror
	if err := ctx.BodyParser(&mirror); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	normalized := make([]string, 0, len(mirror.Endpoints))
	for _, ep := range mirror.Endpoints {
		n, err := normalizeEndpoint(ep)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		normalized = append(normalized, n)
	}
	if err := writeRegistryMirror(registry, normalized); err != nil {
		return nil, err
	}
	return &Mirror{Endpoints: normalized}, nil
}

// DeleteRegistryMirror removes a registry's mirror config (its certs.d subdir).
func DeleteRegistryMirror(ctx *fiber.Ctx) error {
	registry := ctx.Params(ParamRegistryName)
	if registry == "" {
		return fiber.NewError(fiber.StatusBadRequest, "registry name is required")
	}
	return os.RemoveAll(registryDir(registry))
}

// EnsureRegistryMirror ensures endpoint is configured as a mirror host for
// registry: if it is already present (in any position) the list is left
// unchanged; otherwise it is prepended as the highest-priority host. Returns
// updated=true if it changed.
func EnsureRegistryMirror(ctx context.Context, registry, endpoint string) (bool, error) {
	return ReconcileMirrorEndpoint(ctx, registry, endpoint, true)
}

// ReconcileMirrorEndpoint reconciles a single mirror host for registry:
//   - present=true: if endpoint is already in the list (any position) the list is
//     left as-is; if absent it is prepended as the highest-priority host.
//   - present=false: endpoint is removed, preserving the order of the rest.
//
// Returns updated=true if the hosts.toml changed. Useful for runtime reconcilers
// that add/withdraw a single mirror endpoint.
func ReconcileMirrorEndpoint(ctx context.Context, registry, endpoint string, present bool) (bool, error) {
	if registry == "" {
		registry = DefaultRegistryName
	}
	endpoint, err := normalizeEndpoint(endpoint)
	if err != nil {
		return false, err
	}
	existing, err := readRegistryMirror(registry)
	if err != nil {
		return false, err
	}
	found := false
	for _, ep := range existing {
		if ep == endpoint {
			found = true
			break
		}
	}

	var next []string
	if present {
		if found {
			// already configured (in any position): keep the list untouched.
			return false, nil
		}
		// absent: prepend as the highest-priority host.
		next = append([]string{endpoint}, existing...)
	} else {
		if !found {
			return false, nil
		}
		next = make([]string, 0, len(existing))
		for _, ep := range existing {
			if ep != endpoint {
				next = append(next, ep)
			}
		}
	}
	return true, writeRegistryMirror(registry, next)
}

// ListRegistries merges configured mirrors with registries discovered from
// locally cached images.
func ListRegistries(ctx *fiber.Ctx) ([]*Registry, error) {
	nameToRegistries := make(map[string]*Registry)
	mirrors, err := GetRegistryMirrors(ctx)
	if err != nil {
		return nil, err
	}
	mirrorEndpointHosts := make(map[string]struct{})
	for registryName, mirror := range mirrors {
		nameToRegistries[registryName] = &Registry{Name: registryName, Endpoints: mirror.Endpoints}
		for _, ep := range mirror.Endpoints {
			u, perr := url.Parse(ep)
			if perr != nil || u == nil {
				klog.Errorf("failed to parse mirror endpoint %q for registry %q: %v", ep, registryName, perr)
				continue
			}
			if hn := u.Hostname(); hn != "" {
				mirrorEndpointHosts[hn] = struct{}{}
			}
			if h := u.Host; h != "" {
				mirrorEndpointHosts[h] = struct{}{}
			}
		}
	}
	for host := range mirrorEndpointHosts {
		delete(nameToRegistries, host)
	}
	images, err := ListImages(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			refspec, err := reference.Parse(tag)
			if err != nil {
				klog.Errorf("failed to parse image tag %s: %v", tag, err)
				continue
			}
			host := refspec.Hostname()
			if _, isMirrorEndpoint := mirrorEndpointHosts[host]; isMirrorEndpoint {
				continue
			}
			if host == "" {
				klog.Errorf("failed to parse image tag %s: empty host", tag)
				continue
			}
			if registry, ok := nameToRegistries[host]; !ok || registry == nil {
				nameToRegistries[host] = &Registry{Name: host}
			}
			nameToRegistries[host].ImageCount += 1
			nameToRegistries[host].ImageSize += image.Size
		}
	}
	var registries []*Registry
	for _, registry := range nameToRegistries {
		registries = append(registries, registry)
	}
	return registries, nil
}
