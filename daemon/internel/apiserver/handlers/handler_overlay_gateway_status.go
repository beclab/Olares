package handlers

import (
	"context"
	"net/http"
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

// Node facts the status derivation depends on, as variables so the derivation
// can be unit-tested without netlink, systemd or NetworkManager.
var (
	overlayGatewayDesired = utils.OverlayGatewayDesired
	overlayParentLinkUp   = utils.OverlayParentLinkUp
	isCniDhcpActive       = utils.IsCniDhcpActive
	kernelSupportsAltname = utils.KernelSupportsAltname
	isWSL                 = utils.IsWSL
	isDarwin              = utils.IsDarwin
	isEthernetConnected   = func(ctx context.Context) bool {
		iface, _, _, err := utils.GetEthernetConnection(ctx)
		if err != nil {
			return false
		}
		return iface != ""
	}
)

var disableOverlayGatewayError string = ""
var enableOverlayGatewayError string = ""
var operateOverlayGatewayMutex sync.Mutex

type OverlayGatewaySupportedApp struct {
	AppName          string                  `json:"app_name"`
	Enabled          bool                    `json:"enabled"`
	SharedApp        bool                    `json:"shared_app"`
	AppID            string                  `json:"app_id"`
	UnderlayNetworks []utils.UnderlayNetwork `json:"underlay_networks"`
}

type OverlayGatewayStatus struct {
	Status        string                       `json:"status"`
	Disable       bool                         `json:"disable"`
	DisableReason string                       `json:"disable_reason"`
	SupportedApps []OverlayGatewaySupportedApp `json:"supported_apps"`
	ErrorMessage  string                       `json:"error_message"`
	// CniDhcpActive reports the CNI DHCP daemon health. It is infrastructure
	// that runs regardless of the switch, so it is never folded into Status.
	CniDhcpActive bool `json:"cni_dhcp_active"`
}

func (h *Handlers) GetOverlayGatewayStatus(ctx *fiber.Ctx) error {
	user := ctx.Params("user")
	if user == "" {
		return h.ErrJSON(ctx, http.StatusBadRequest, "user is required")
	}

	if err := h.itsMe(ctx, user); err != nil {
		return h.ErrJSON(ctx, http.StatusForbidden, err.Error())
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

	if inFlight, stale := consumeOverlayGatewayOpLock(OverlayGatewayEnableLockFile); inFlight {
		s.Status = OverlayGatewayActivating
		return h.OkJSON(ctx, "success", s)
	} else if stale != "" {
		s.ErrorMessage = stale
	}
	if inFlight, stale := consumeOverlayGatewayOpLock(OverlayGatewayDisableLockFile); inFlight {
		s.Status = OverlayGatewayDeactivating
		return h.OkJSON(ctx, "success", s)
	} else if stale != "" {
		s.ErrorMessage = stale
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
			AppName:          app.AppName,
			Enabled:          app.Enabled,
			SharedApp:        app.SharedApp,
			AppID:            app.AppID,
			UnderlayNetworks: app.UnderlayNetworks,
		})
	}

	return apps, nil
}

// getOverlayGatewayStatus derives the switch state from the desired-state file
// and the presence of the overlay parent alternative name. When the state is
// desired but the name cannot be resolved, the switch reads "off" with an
// error message: enabling it again re-adds the name, which is the repair path.
func (h *Handlers) getOverlayGatewayStatus(ctx context.Context) (*OverlayGatewayStatus, error) {
	s := &OverlayGatewayStatus{
		Status:        OverlayGatewayOff,
		CniDhcpActive: isCniDhcpActive(ctx),
	}

	if !overlayGatewayDesired() {
		s.Disable, s.DisableReason = h.isUnsupported(ctx)
		return s, nil
	}

	dev, up, err := overlayParentLinkUp(ctx)
	if err != nil {
		klog.Errorf("overlay gateway status: %s is not present although the gateway is enabled: %v", utils.OverlayParentAltname, err)
		s.ErrorMessage = "overlay parent interface " + utils.OverlayParentAltname + " is missing; enable the overlay gateway again to restore it"
		return s, nil
	}
	if !up {
		klog.Warningf("overlay gateway status: overlay parent %s (%s) is down", dev, utils.OverlayParentAltname)
		s.ErrorMessage = "overlay parent interface " + dev + " is down"
	}
	s.Status = OverlayGatewayOn
	return s, nil
}

func (h *Handlers) isUnsupported(ctx context.Context) (unsupported bool, reason string) {
	switch {
	case isWSL():
		return true, "WSL is not supported"
	case isDarwin():
		return true, "MacOS is not supported"
	case !kernelSupportsAltname():
		return true, "Kernel is too old for alternative interface names (5.5 or later is required)"
	case !isEthernetConnected(ctx):
		return true, "Ethernet connection is not active"
	}

	return false, ""
}
