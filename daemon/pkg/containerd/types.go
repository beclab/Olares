package containerd

import (
	"github.com/containerd/containerd/pkg/cri/config"
	criruntimev1 "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// Mirror is an alias for the containerd CRI plugin Mirror type for convenience
type Mirror = config.Mirror

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
