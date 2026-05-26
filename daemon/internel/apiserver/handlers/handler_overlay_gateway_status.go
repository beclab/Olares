package handlers

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

const (
	OverlayGatewayOn              = "on"
	OverlayGatewayOff             = "off"
	OverlayGatewayActivating      = "activating"
	OverlayGatewayDeactivating    = "deactivating"
	OverlayGatewayDisableLockFile = "/var/run/overlay_gateway_disable.lock"
	OverlayGatewayEnableLockFile  = "/var/run/overlay_gateway_enable.lock"
)

var disableOverlayGatewayError string = ""
var enableOverlayGatewayError string = ""
var operateOverlayGatewayMutex sync.Mutex

type OverlayGatewaySupportedApp struct {
	AppName   string `json:"app_name"`
	Enabled   bool   `json:"enabled"`
	SharedApp bool   `json:"shared_app"`
	AppID     string `json:"app_id"`
}

type OverlayGatewayStatus struct {
	Status        string                       `json:"status"`
	Disable       bool                         `json:"disable"`
	DisableReason string                       `json:"disable_reason"`
	SupportedApps []OverlayGatewaySupportedApp `json:"supported_apps"`
	ErrorMessage  string                       `json:"error_message"`
}

func (h *Handlers) GetOverlayGatewayStatus(ctx *fiber.Ctx) error {
	user := ctx.Params("user")
	if user == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "user is required")
	}

	s, err := h.getOverlayGatewayStatus(ctx.Context())
	if err != nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, err.Error())
	}

	if s == nil {
		return h.ErrJSON(ctx, http.StatusInternalServerError, "failed to get overlay gateway status")
	}

	if enableOverlayGatewayError != "" {
		s.ErrorMessage = enableOverlayGatewayError
	}
	if disableOverlayGatewayError != "" {
		s.ErrorMessage = disableOverlayGatewayError
	}

	if _, err := os.Stat(OverlayGatewayEnableLockFile); err == nil {
		s.Status = OverlayGatewayActivating
		return h.OkJSON(ctx, "success", s)
	}
	if _, err := os.Stat(OverlayGatewayDisableLockFile); err == nil {
		s.Status = OverlayGatewayDeactivating
		return h.OkJSON(ctx, "success", s)
	}

	if s.Status == OverlayGatewayOn {
		// get the supported apps
		supportedApps, err := h.getOverlayGatewaySupportedApps(ctx.Context(), user)
		if err != nil {
			klog.Error("get overlay gateway supported apps error, ", err)
		}
		s.SupportedApps = supportedApps
	}

	return h.OkJSON(ctx, "success", s)
}

func (h *Handlers) getOverlayGatewaySupportedApps(ctx context.Context, user string) ([]OverlayGatewaySupportedApp, error) {
	supportedApps, err := utils.GetOverlayGatewaySupportedApps(ctx, user)
	if err != nil {
		return nil, err
	}

	var apps []OverlayGatewaySupportedApp
	for _, app := range supportedApps {
		apps = append(apps, OverlayGatewaySupportedApp{
			AppName:   app.AppName,
			Enabled:   app.Enabled,
			SharedApp: app.SharedApp,
			AppID:     app.AppID,
		})
	}

	return apps, nil
}

func (h *Handlers) getOverlayGatewayStatus(ctx context.Context) (*OverlayGatewayStatus, error) {
	s := &OverlayGatewayStatus{
		Status: OverlayGatewayOff,
	}

	c, err := utils.FindBridgeConnection(ctx)
	if err != nil {
		return nil, err
	}

	if c == nil {
		s.Disable, s.DisableReason = h.isUnsupported(ctx)
		return s, nil
	}

	if c.Active {
		s.Status = OverlayGatewayOn
		return s, nil
	}

	s.Disable, s.DisableReason = h.isUnsupported(ctx)

	return s, nil
}

func (h *Handlers) isUnsupported(ctx context.Context) (unsupported bool, reason string) {
	isEthernetConnected := func(ctx context.Context) bool {
		iface, _, _, err := utils.GetEthernetConnection(ctx)
		if err != nil {
			return false
		}
		return iface != ""
	}

	switch {
	case utils.IsWSL():
		return true, "WSL is not supported"
	case utils.IsDarwin():
		return true, "MacOS is not supported"
	case !isEthernetConnected(ctx):
		return true, "Ethernet connection is not active"
	}

	return false, ""
}
