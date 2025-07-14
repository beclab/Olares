package handlers

import (
	"net/http"

	"github.com/beclab/Olares/daemon/pkg/containerd"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

func (h *Handlers) ListRegistries(ctx *fiber.Ctx) error {
	images, err := containerd.ListRegistries(ctx)
	if err != nil {
		klog.Error("list registries error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}
	return h.OkJSON(ctx, "success", images)
}

func (h *Handlers) GetRegistryMirrors(ctx *fiber.Ctx) error {
	mirrors, err := containerd.GetRegistryMirrors(ctx)
	if err != nil {
		klog.Error("get registry mirrors error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success", mirrors)
}

func (h *Handlers) GetRegistryMirror(ctx *fiber.Ctx) error {
	mirror, err := containerd.GetRegistryMirror(ctx)
	if err != nil {
		klog.Error("get registry mirror error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success", mirror)
}

func (h *Handlers) UpdateRegistryMirror(ctx *fiber.Ctx) error {
	mirror, err := containerd.UpdateRegistryMirror(ctx)
	if err != nil {
		klog.Error("update registry mirror error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success", mirror)
}

func (h *Handlers) DeleteRegistryMirror(ctx *fiber.Ctx) error {
	if err := containerd.DeleteRegistryMirror(ctx); err != nil {
		klog.Error("delete registry mirror error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success")
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
