package containerd

import (
	"github.com/containerd/containerd"
	cdefaults "github.com/containerd/containerd/defaults"
)

const (
	DefaultNamespace = "k8s.io"
)

func NewClient() (*containerd.Client, error) {
	return containerd.New(
		cdefaults.DefaultAddress,
		containerd.WithDefaultRuntime(cdefaults.DefaultRuntime),
		containerd.WithDefaultNamespace(DefaultNamespace),
	)
}
