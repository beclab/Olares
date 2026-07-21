package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

func (h *Handlers) DisableOverlayGateway(ctx *fiber.Ctx, cmd commands.Interface) error {
	operateOverlayGatewayMutex.Lock()
	defer operateOverlayGatewayMutex.Unlock()

	s, err := h.getOverlayGatewayStatus(ctx.Context())
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	if s.Disable {
		return h.ErrJSON(ctx, http.StatusBadRequest, s.DisableReason)
	}

	// check if the lock file exists
	if _, err := os.Stat(OverlayGatewayDisableLockFile); err == nil {
		s.Status = OverlayGatewayDeactivating
		return h.OkJSON(ctx, "success", s)
	}

	if _, err := os.Stat(OverlayGatewayEnableLockFile); err == nil {
		s.Status = OverlayGatewayActivating
		return h.OkJSON(ctx, "success", s)
	}

	if s.Status != OverlayGatewayOn {
		return h.ErrJSON(ctx, http.StatusBadRequest, "overlay gateway is already disabled")
	}

	// create the lock file synchronously while holding the mutex so that a
	// concurrent disable request observes it and returns "deactivating"
	// instead of starting a second teardown (single-flight).
	f, err := os.Create(OverlayGatewayDisableLockFile)
	if err != nil {
		klog.Errorf("overlay gateway disable: create lock file failed: %v", err)
		return h.ErrJSON(ctx, http.StatusInternalServerError, "failed to create overlay gateway disable lock")
	}
	_ = f.Close()
	disableOverlayGatewayError = ""
	enableOverlayGatewayError = ""

	go func() {
		defer os.Remove(OverlayGatewayDisableLockFile)
		bgCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, err := cmd.Execute(bgCtx, nil)
		if err != nil {
			disableOverlayGatewayError = err.Error()
			return
		}

		t := time.NewTicker(2 * time.Second)
		timeout := time.NewTimer(60 * time.Second)
		defer t.Stop()
		defer timeout.Stop()
		for {
			select {
			case <-t.C:
				s, err := h.getOverlayGatewayStatus(bgCtx)
				if err != nil {
					disableOverlayGatewayError = err.Error()
					return
				}
				if s.Status == OverlayGatewayOff {
					klog.Info("overlay gateway is disabled successfully")
					return
				}
			case <-timeout.C:
				klog.Error("overlay gateway disable timeout")
				disableOverlayGatewayError = "overlay gateway disable timeout"
				return
			}
		}
	}()
	return h.ErrJSON(ctx, http.StatusOK, "success")
}
