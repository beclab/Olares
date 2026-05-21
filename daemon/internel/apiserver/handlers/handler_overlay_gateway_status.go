package handlers

import (
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

const (
	OverlayGatewayOn           = "on"
	OverlayGatewayOff          = "off"
	OverlayGatewayActivating   = "activating"
	OverlayGatewayDeactivating = "deactivating"
)

type OverlayGatewaySupportedApp struct {
	AppName string `json:"app_name"`
	Enabled bool   `json:"enabled"`
}

type OverlayGatewayStatus struct {
	Status        string                       `json:"status"`
	Disable       bool                         `json:"disable"`
	DisableReason string                       `json:"disable_reason"`
	SupportedApps []OverlayGatewaySupportedApp `json:"supported_apps"`
}

func (h *Handlers) GetOverlayGatewayStatus(ctx *fiber.Ctx) error {
	s := OverlayGatewayStatus{
		Status: OverlayGatewayOff,
	}
	return h.OkJSON(ctx, "success", s)
}

func (h *Handlers) getOverlayGatewaySupportedApps(ctx *fiber.Ctx) ([]OverlayGatewaySupportedApp, error) {
	return []OverlayGatewaySupportedApp{
		{
			AppName: "Olares",
			Enabled: true,
		},
		{
			AppName: "Olares",
			Enabled: true,
		},
	}, nil
}

func (h *Handlers) getOverlayGatewayStatus(ctx *fiber.Ctx) (*OverlayGatewayStatus, error) {
	c, err := utils.FindBridgeConnection(ctx.Context())
	if err != nil {
		return nil, err
	}

	return &OverlayGatewayStatus{
		Status: OverlayGatewayOn,
	}, nil
}
