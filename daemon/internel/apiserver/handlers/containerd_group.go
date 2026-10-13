package handlers

import (
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"k8s.io/klog/v2"
)

func init() {
	s := server.API
	containerd := s.App.Group("containerd")

	// Node-level registry config (containerd v3 certs.d/hosts.toml).
	containerd.Get("/registries", handlers.RequireLocal(handlers.ListRegistries))

	registry := containerd.Group("registry")
	configs := registry.Group("configs")
	configs.Get("/", handlers.RequireAuthorization(handlers.RequireLocal(handlers.GetRegistryConfigs)))
	// The greedy fallback also accepts an unescaped scheme containing slashes.
	for _, path := range []string{"/:registry", "/+"} {
		configs.Get(path, handlers.RequireAuthorization(handlers.RequireLocal(handlers.GetRegistryConfig)))
		configs.Put(path, handlers.RequireAuthorization(handlers.RequireLocal(handlers.UpdateRegistryConfig)))
		configs.Delete(path, handlers.RequireAuthorization(handlers.RequireLocal(handlers.DeleteRegistryConfig)))
	}

	image := containerd.Group("images")
	image.Get("/", handlers.RequireLocal(handlers.ListImages))
	image.Delete("/:image", handlers.RequireLocal(handlers.DeleteImage))
	image.Post("/prune", handlers.RequireLocal(handlers.PruneImages))

	klog.V(8).Info("containerd handlers initialized")
}
