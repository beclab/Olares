package apiserver

import (
	"net/http"
	"os"

	"bytetrade.io/web3os/terminusd/internel/ble"
	"bytetrade.io/web3os/terminusd/pkg/nets"
	"bytetrade.io/web3os/terminusd/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

type NetIf struct {
	Iface    string  `json:"iface"`
	IP       string  `json:"ip"`
	IsHostIp bool    `json:"isHostIp"`
	IsWifi   bool    `json:"isWifi"`
	SSID     *string `json:"ssid,omitempty"`
	Strength *int    `json:"strength,omitempty"`
}

func (h *handlers) GetNetIfs(ctx *fiber.Ctx) error {
	ifaces, err := nets.GetInternalIpv4Addr()
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
		}

		if wifiDevs != nil {
			if wd, ok := wifiDevs[r.Iface]; ok {
				r.IsWifi = true
				r.SSID = &wd.Connection

				if ap := h.findAp(*r.SSID); ap != nil {
					r.Strength = ptr.To(int(ap.Strength))
				}
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
