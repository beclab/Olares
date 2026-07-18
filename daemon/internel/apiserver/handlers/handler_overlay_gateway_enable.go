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

func (h *Handlers) EnableOverlayGateway(ctx *fiber.Ctx, cmd commands.Interface) error {
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
	if _, err := os.Stat(OverlayGatewayEnableLockFile); err == nil {
		s.Status = OverlayGatewayActivating
		return h.OkJSON(ctx, "success", s)
	}

	if _, err := os.Stat(OverlayGatewayDisableLockFile); err == nil {
		s.Status = OverlayGatewayDeactivating
		return h.OkJSON(ctx, "success", s)
	}

	if s.Status != OverlayGatewayOff {
		return h.ErrJSON(ctx, http.StatusBadRequest, "overlay gateway is already enabled")
	}

	// create the lock file synchronously while holding the mutex so that a
	// concurrent enable request observes it and returns "activating" instead of
	// starting a second switch (single-flight).
	os.Create(OverlayGatewayEnableLockFile)
	enableOverlayGatewayError = ""
	disableOverlayGatewayError = ""

	go func() {
		defer os.Remove(OverlayGatewayEnableLockFile)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// CreateBridgeConnection is atomic: a nil error means the bridge is up
		// with an IPv4 address, a non-nil error means it has been rolled back to
		// the original network. Surface the failure so the UI stops spinning.
		_, err := cmd.Execute(ctx, nil)
		if err != nil {
			enableOverlayGatewayError = err.Error()
			return
		}

		// confirm the reported status reaches "on"
		t := time.NewTicker(2 * time.Second)
		timeout := time.NewTimer(60 * time.Second)
		defer t.Stop()
		defer timeout.Stop()
		for {
			select {
			case <-t.C:
				s, err := h.getOverlayGatewayStatus(ctx)
				if err != nil {
					enableOverlayGatewayError = err.Error()
					return
				}
				if s.Status == OverlayGatewayOn {
					klog.Info("overlay gateway is enabled successfully")
					return
				}
			case <-timeout.C:
				klog.Error("overlay gateway enable timeout")
				enableOverlayGatewayError = "overlay gateway enable timeout"
				return
			}
		}
	}()
	return h.ErrJSON(ctx, http.StatusOK, "success")
}
