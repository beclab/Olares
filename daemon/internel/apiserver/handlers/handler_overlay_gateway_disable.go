package handlers

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

var disableOverlayGatewayMutex sync.Mutex

func (h *Handlers) DisableOverlayGateway(ctx *fiber.Ctx, cmd commands.Interface) error {
	disableOverlayGatewayMutex.Lock()
	defer disableOverlayGatewayMutex.Unlock()

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

	go func() {
		os.Create(OverlayGatewayDisableLockFile)
		defer os.Remove(OverlayGatewayDisableLockFile)
		disableOverlayGatewayError = ""
		enableOverlayGatewayError = ""
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, err = cmd.Execute(ctx, nil)
		if err != nil {
			disableOverlayGatewayError = err.Error()
		}

		t := time.NewTicker(2 * time.Second)
		timeout := time.NewTimer(10 * time.Second)
		defer t.Stop()
		defer timeout.Stop()
		for {
			select {
			case <-t.C:
				s, err = h.getOverlayGatewayStatus(ctx)
				if err != nil {
					return
				}
				if s.Status == OverlayGatewayOff {
					klog.Info("overlay gateway is disabled successfully")
					return
				}
			case <-timeout.C:
				klog.Error("overlay gateway disable timeout")
				return
			}
		}
	}()
	return h.ErrJSON(ctx, http.StatusOK, "success")
}
