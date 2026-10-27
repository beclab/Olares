package containerd

import (
	criruntimev1 "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// DefaultContainerdRootPath is containerd's default data root, used by the
// upgrade disk-space precheck.
const DefaultContainerdRootPath = "/var/lib/containerd"

// Mirror is the internal ordered endpoint view used by registry listings and
// node mirror reconciliation. Public configuration uses RegistryConfig.
type Mirror struct {
	Endpoints []string `json:"endpoint" toml:"endpoint"`
}

// Registry is a registry view merged from configured mirrors and locally cached
// images (used by the ListRegistries endpoint).
type Registry struct {
	Name       string   `json:"name"`
	Endpoints  []string `json:"endpoints"`
	ImageCount int      `json:"image_count"`
	ImageSize  uint64   `json:"image_size"`
}

type PruneImageResult struct {
	Images []*criruntimev1.Image `json:"images"`
	Count  int                   `json:"count"`
	Size   uint64                `json:"size"`
}
