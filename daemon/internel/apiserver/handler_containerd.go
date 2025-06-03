package apiserver

import (
	"net/http"

	"bytetrade.io/web3os/terminusd/pkg/containerd"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

func (h *handlers) ListRegistries(ctx *fiber.Ctx) error {
	images, err := containerd.ListRegistries(ctx)
	if err != nil {
		klog.Error("list registries error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}
	return h.OkJSON(ctx, "success", images)
}

func (h *handlers) GetRegistryMirrors(ctx *fiber.Ctx) error {
	mirrors, err := containerd.GetRegistryMirrors(ctx)
	if err != nil {
		klog.Error("get registry mirrors error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success", mirrors)
}

func (h *handlers) GetRegistryMirror(ctx *fiber.Ctx) error {
	mirror, err := containerd.GetRegistryMirror(ctx)
	if err != nil {
		klog.Error("get registry mirror error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success", mirror)
}

func (h *handlers) UpdateRegistryMirror(ctx *fiber.Ctx) error {
	mirror, err := containerd.UpdateRegistryMirror(ctx)
	if err != nil {
		klog.Error("update registry mirror error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success", mirror)
}

func (h *handlers) DeleteRegistryMirror(ctx *fiber.Ctx) error {
	if err := containerd.DeleteRegistryMirror(ctx); err != nil {
		klog.Error("delete registry mirror error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	return h.OkJSON(ctx, "success")
}

func (h *handlers) ListImages(ctx *fiber.Ctx) error {
	registry := ctx.Query("registry")
	images, err := containerd.ListImages(ctx, registry)
	if err != nil {
		klog.Error("list images error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}
	return h.OkJSON(ctx, "success", images)
}

func (h *handlers) DeleteImage(ctx *fiber.Ctx) error {
	if err := containerd.DeleteImage(ctx); err != nil {
		klog.Error("delete image error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}
	return h.OkJSON(ctx, "success")
}

func (h *handlers) PruneImages(ctx *fiber.Ctx) error {
	res, err := containerd.PruneImages(ctx)
	if err != nil {
		klog.Error("prune images error, ", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}
	return h.OkJSON(ctx, "success", res)
}
