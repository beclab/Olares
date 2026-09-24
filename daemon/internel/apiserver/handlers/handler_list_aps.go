package handlers

import (
	"context"
	"net/http"

	"github.com/beclab/Olares/daemon/internel/wifi"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

var listWifiAccessPoints = wifi.ListAccessPoints

// apiAccessPoint adds connection status only to the HTTP response.
type apiAccessPoint struct {
	wifi.AccessPoint
	Connected bool `json:"connected"`
}

func (h *Handlers) GetListAPs(ctx *fiber.Ctx) error {
	scanCtx, cancel := context.WithCancel(ctx.UserContext())
	defer cancel()
	if h.mainCtx != nil {
		stop := context.AfterFunc(h.mainCtx, cancel)
		defer stop()
	}
	aps, err := listWifiAccessPoints(scanCtx)
	if err != nil {
		klog.Error("Failed to list wifi access points")
		return h.ErrJSON(ctx, http.StatusServiceUnavailable, "failed to list wifi access points")
	}
	// HTTP includes connection status and the full list of access points.
	result := make([]apiAccessPoint, 0, len(aps))
	for _, ap := range aps {
		result = append(result, apiAccessPoint{AccessPoint: ap, Connected: ap.Connected})
	}
	return h.OkJSON(ctx, "", result)
}
