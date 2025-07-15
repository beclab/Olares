package handlers

import (
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"k8s.io/klog/v2"
)

func init() {
	s := server.API
	containerd := s.App.Group("containerd")
	containerd.Get("/registries", handlers.RequireLocal(handlers.ListRegistries))

	registry := containerd.Group("registry")
	mirrors := registry.Group("mirrors")

	mirrors.Get("/", handlers.RequireLocal(handlers.GetRegistryMirrors))
	mirrors.Get("/:registry", handlers.RequireLocal(handlers.GetRegistryMirror))
	mirrors.Put("/:registry", handlers.RequireLocal(handlers.UpdateRegistryMirror))
	mirrors.Delete("/:registry", handlers.RequireLocal(handlers.DeleteRegistryMirror))

	image := containerd.Group("images")

	image.Get("/", handlers.RequireLocal(handlers.ListImages))
	image.Delete("/:image", handlers.RequireLocal(handlers.DeleteImage))
	image.Post("/prune", handlers.RequireLocal(handlers.PruneImages))

	klog.V(8).Info("containerd handlers initialized")
}
