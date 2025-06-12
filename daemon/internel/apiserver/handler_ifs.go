package apiserver

import (
	"net/http"
	"os"

	"github.com/beclab/Olares/daemon/internel/ble"
	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

type NetIf struct {
	Iface             string  `json:"iface"`
	IP                string  `json:"ip"`
	IsHostIp          bool    `json:"isHostIp"`
	IsWifi            bool    `json:"isWifi"`
	SSID              *string `json:"ssid,omitempty"`
	Strength          *int    `json:"strength,omitempty"`
	MTU               int     `json:"mtu,omitempty"`
	InternetConnected *bool   `json:"internetConnected,omitempty"`

	Ipv4Gateway      *string  `json:"ipv4Gateway,omitempty"`
	Ipv6Gateway      *string  `json:"ipv6Gateway,omitempty"`
	Ipv4DNS          *string  `json:"ipv4DNS,omitempty"`
	Ipv6DNS          *string  `json:"ipv6DNS,omitempty"`
	Ipv6Address      *string  `json:"ipv6Address,omitempty"`
	Ipv4Mask         *string  `json:"ipv4Mask,omitempty"`
	Method           *string  `json:"method,omitempty"` // dhcp, auto, manual, etc.
	Ipv6Connectivity *bool    `json:"ipv6Connectivity,omitempty"`
	RxRate           *float64 `json:"rxRate,omitempty"` // in bytes per second
	TxRate           *float64 `json:"txRate,omitempty"` // in bytes per second
}

func (h *handlers) GetNetIfs(ctx *fiber.Ctx) error {
	test := ctx.Query("testConnectivity", "false")

	ifaces, err := nets.GetInternalIpv4Addr(test != "true")
	if err != nil {
		return h.ErrJSON(ctx, http.StatusServiceUnavailable, err.Error())
	}

	host, err := os.Hostname()
	if err != nil {
		return h.ErrJSON(ctx, http.StatusServiceUnavailable, err.Error())
	}

	hostip, err := nets.GetHostIpFromHostsFile(host)
	if err != nil {
		return h.ErrJSON(ctx, http.StatusServiceUnavailable, err.Error())
	}

	wifiDevs, err := utils.GetWifiDevice(ctx.Context())
	if err != nil {
		klog.Error("get wifi device info error, ", err)
	}

	var res []NetIf
	ifMap := make(map[string]string)
	for _, i := range ifaces {
		r := NetIf{
			Iface:    i.Iface.Name,
			IP:       i.IP,
			IsHostIp: i.IP == hostip,
			MTU:      i.Iface.MTU,
		}

		if wifiDevs != nil {
			if wd, ok := wifiDevs[r.Iface]; ok {
				r.IsWifi = true
				r.SSID = &wd.Connection

				if r.SSID != nil && *r.SSID != "" {
					if ap := h.findAp(*r.SSID); ap != nil {
						r.Strength = ptr.To(int(ap.Strength))
					}
				}
			}
		}

		if test == "true" {
			r.InternetConnected = ptr.To(utils.CheckInterfaceIPv4Connectivity(ctx.Context(), i.Iface.Name))
			devices, err := utils.GetAllDevice(ctx.Context())
			if err != nil {
				klog.Error("get all devices error, ", err)
				return h.ErrJSON(ctx, http.StatusServiceUnavailable, err.Error())
			}

			if d, ok := devices[r.Iface]; ok {
				r.Ipv4Gateway = &d.Ipv4Gateway
				r.Ipv6Gateway = &d.Ipv6Gateway
				r.Ipv4DNS = &d.Ipv4DNS
				r.Ipv6DNS = &d.Ipv6DNS
				r.Ipv6Address = &d.Ipv6Address
				r.Ipv4Mask = &d.Ipv4Mask
				r.Method = &d.Method
			}

			if r.Ipv6Address != nil && *r.Ipv6Address != "" {
				// check ipv6 connectivity
				connected := utils.CheckInterfaceIPv6Connectivity(ctx.Context(), i.Iface.Name)
				r.Ipv6Connectivity = &connected
			}

			if rx, tx, err := utils.GetInterfaceTraffic(r.Iface); err == nil {
				r.RxRate = ptr.To(rx)
				r.TxRate = ptr.To(tx)
			} else {
				klog.Error("get interface rx/tx rate error, ", err)
			}
		}

		res = append(res, r)
		ifMap[r.Iface] = r.Iface
	}

	// append not-connected wifi
	for _, d := range wifiDevs {
		if _, ok := ifMap[d.Name]; !ok {
			r := NetIf{
				Iface:    d.Name,
				IP:       "",
				IsHostIp: false,
				IsWifi:   true,
			}

			res = append(res, r)
		}
	}

	return h.OkJSON(ctx, "", res)
}

func (h *handlers) findAp(ssid string) *ble.AccessPoint {
	for _, ap := range h.apList {
		if ap.SSID == ssid {
			return &ap
		}
	}

	return nil
}
