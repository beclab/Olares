package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/pkg/containerd"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

// olaresd manages THIS node's local containerd images and registry
// config (containerd v3 config_path -> /etc/containerd/certs.d/<registry>/hosts.toml).
// Registry config changes need no restart, since hosts.toml is read per-pull.

func (h *Handlers) ListRegistries(ctx *fiber.Ctx) error {
	registries, err := containerd.ListRegistries(ctx)
	if err != nil {
		klog.Error("list registries error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}
	return h.OkJSON(ctx, "success", registries)
}

func (h *Handlers) ListImages(ctx *fiber.Ctx) error {
	registry := ctx.Query("registry")
	images, err := containerd.ListImages(ctx, registry)
	if err != nil {
		klog.Error("list images error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}
	return h.OkJSON(ctx, "success", images)
}

func (h *Handlers) DeleteImage(ctx *fiber.Ctx) error {
	if err := containerd.DeleteImage(ctx); err != nil {
		klog.Error("delete image error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}
	return h.OkJSON(ctx, "success")
}

func (h *Handlers) PruneImages(ctx *fiber.Ctx) error {
	res, err := containerd.PruneImages(ctx)
	if err != nil {
		klog.Error("prune images error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}
	return h.OkJSON(ctx, "success", res)
}

func (h *Handlers) GetRegistryConfigs(ctx *fiber.Ctx) error {
	configs, err := containerd.GetRegistryConfigs(ctx)
	if err != nil {
		return registryConfigError(h, ctx, err)
	}
	return h.OkJSON(ctx, "success", configs)
}

func (h *Handlers) GetRegistryConfig(ctx *fiber.Ctx) error {
	config, err := containerd.GetRegistryConfig(ctx)
	if err != nil {
		return registryConfigError(h, ctx, err)
	}
	return h.OkJSON(ctx, "success", config)
}

func (h *Handlers) UpdateRegistryConfig(ctx *fiber.Ctx) error {
	config, err := containerd.UpdateRegistryConfig(ctx)
	if err != nil {
		return registryConfigError(h, ctx, err)
	}
	return h.OkJSON(ctx, "success", config)
}

func (h *Handlers) DeleteRegistryConfig(ctx *fiber.Ctx) error {
	if err := containerd.DeleteRegistryConfig(ctx); err != nil {
		return registryConfigError(h, ctx, err)
	}
	return h.OkJSON(ctx, "success")
}

func registryConfigError(h *Handlers, ctx *fiber.Ctx, err error) error {
	code := http.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return h.ErrJSON(ctx, code, err.Error())
}
