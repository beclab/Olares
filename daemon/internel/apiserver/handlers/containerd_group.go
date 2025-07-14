package handlers

import (
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"k8s.io/klog/v2"
)

func init() {
	s := server.API
	containerd := s.App.Group("containerd")
	containerd.Get("/registries", handlers.RequireSignature(handlers.ListRegistries))

	registry := containerd.Group("registry")
	mirrors := registry.Group("mirrors")

	mirrors.Get("/", handlers.RequireSignature(handlers.GetRegistryMirrors))
	mirrors.Get("/:registry", handlers.RequireSignature(handlers.GetRegistryMirror))
	mirrors.Put("/:registry", handlers.RequireSignature(handlers.UpdateRegistryMirror))
	mirrors.Delete("/:registry", handlers.RequireSignature(handlers.DeleteRegistryMirror))

	image := containerd.Group("images")

	image.Get("/", handlers.RequireSignature(handlers.ListImages))
	image.Delete("/:image", handlers.RequireSignature(handlers.DeleteImage))
	image.Post("/prune", handlers.RequireSignature(handlers.PruneImages))

	klog.Info("containerd handlers initialized")
}
