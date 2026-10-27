package registry

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/remotes/docker/config"
)

const HostsDir = "/etc/containerd/certs.d"

// NewResolver reads the same per-registry hosts.toml files as CRI on each pull.
// Keep the original image name: the resolver handles mirrors, TLS and fallback.
func NewResolver(ctx context.Context) remotes.Resolver {
	return docker.NewResolver(docker.ResolverOptions{Hosts: registryHosts(ctx, HostsDir)})
}

func registryHosts(ctx context.Context, root string) docker.RegistryHosts {
	hosts := config.ConfigureHosts(ctx, config.HostOptions{HostDir: config.HostDirFromRoot(root)})
	return func(namespace string) ([]docker.RegistryHost, error) {
		configured, err := hosts(namespace)
		if err != nil {
			return nil, err
		}
		for i := range configured {
			h := &configured[i]
			username, password, found, err := basicCredentials(h.Header)
			if err != nil {
				return nil, err
			}
			if found {
				// Basic credentials must also participate in Bearer token challenges
				// (GitLab/GHCR); sending only a static header cannot do this.
				hostname := h.Host
				h.Authorizer = docker.NewDockerAuthorizer(docker.WithAuthClient(h.Client), docker.WithAuthCreds(func(host string) (string, string, error) {
					if host != hostname {
						return "", "", nil
					}
					return username, password, nil
				}))
				h.Header.Del("Authorization")
			}
		}
		return configured, nil
	}
}

func basicCredentials(header http.Header) (string, string, bool, error) {
	value := header.Get("Authorization")
	scheme, encoded, ok := strings.Cut(value, " ")
	if !ok || !strings.EqualFold(scheme, "Basic") {
		return "", "", false, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", false, fmt.Errorf("invalid registry Basic authorization")
	}
	username, password, ok := strings.Cut(string(decoded), ":")
	if !ok {
		return "", "", false, fmt.Errorf("invalid registry Basic authorization")
	}
	return username, password, true, nil
}
