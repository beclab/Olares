package dashboard

import (
	"context"
	"net/http"
	"net/url"
	"sort"
)

// ----------------------------------------------------------------------------
// FetchSystemIFS — overview network (capi /system/ifs)
// ----------------------------------------------------------------------------

// SystemIFSItem mirrors the dashboard SystemIFSItem type: the union of
// fields the SPA's overview network page reads.
type SystemIFSItem struct {
	Iface             string `json:"iface"`
	IsHostIp          bool   `json:"isHostIp,omitempty"`
	Hostname          string `json:"hostname,omitempty"`
	Method            string `json:"method,omitempty"`
	MTU               any    `json:"mtu,omitempty"`
	IP                string `json:"ip,omitempty"`
	IPv4Mask          string `json:"ipv4Mask,omitempty"`
	IPv4Gateway       string `json:"ipv4Gateway,omitempty"`
	IPv4DNS           string `json:"ipv4DNS,omitempty"`
	IPv6Address       string `json:"ipv6Address,omitempty"`
	IPv6Gateway       string `json:"ipv6Gateway,omitempty"`
	IPv6DNS           string `json:"ipv6DNS,omitempty"`
	InternetConnected bool   `json:"internetConnected,omitempty"`
	IPv6Connectivity  bool   `json:"ipv6Connectivity,omitempty"`
	TxRate            any    `json:"txRate,omitempty"`
	RxRate            any    `json:"rxRate,omitempty"`
}

// FetchSystemIFS queries `/capi/system/ifs?testConnectivity=...`. The SPA's
// initial fetch passes testConnectivity=true so the server probes outgoing
// connectivity per-iface.
func FetchSystemIFS(ctx context.Context, c *Client, testConnectivity bool) ([]SystemIFSItem, error) {
	q := url.Values{}
	if testConnectivity {
		q.Set("testConnectivity", "true")
	} else {
		q.Set("testConnectivity", "false")
	}
	var raw []SystemIFSItem
	if err := c.DoJSON(ctx, http.MethodGet, "/capi/system/ifs", q, nil, &raw); err != nil {
		return nil, err
	}
	// SPA sorts isHostIp first.
	sort.SliceStable(raw, func(i, j int) bool {
		if raw[i].IsHostIp != raw[j].IsHostIp {
			return raw[i].IsHostIp
		}
		return false
	})
	return raw, nil
}

// ----------------------------------------------------------------------------
// FetchSystemFan — overview fan live (capi user-service)
// ----------------------------------------------------------------------------

// SystemFanData mirrors the SPA's getSystemFan response payload
// (.data field of /user-service/api/mdns/olares-one/cpu-gpu).
type SystemFanData struct {
	GPUFanSpeed    float64 `json:"gpu_fan_speed"`
	GPUTemperature float64 `json:"gpu_temperature"`
	CPUFanSpeed    float64 `json:"cpu_fan_speed"`
	CPUTemperature float64 `json:"cpu_temperature"`
}

// FetchSystemFan queries `/user-service/api/mdns/olares-one/cpu-gpu` and
// returns the live RPM + temperature pair the overview fan live leaf
// renders.
func FetchSystemFan(ctx context.Context, c *Client) (*SystemFanData, error) {
	var raw struct {
		Data SystemFanData `json:"data"`
	}
	if err := c.DoJSON(ctx, http.MethodGet, "/user-service/api/mdns/olares-one/cpu-gpu", nil, nil, &raw); err != nil {
		return nil, err
	}
	return &raw.Data, nil
}

// FanCurveTable is the hardcoded fan-curve specification — 1:1 with the
// SPA's `apps/packages/app/src/apps/dashboard/pages/Overview2/Fan/config.ts`
// `tableData` constant. NEVER drift from upstream without updating both
// sides; the iteration red-line in SKILL.md pins this.
var FanCurveTable = []FanCurveRow{
	{Step: 1, CPUFanRPM: 0, GPUFanRPM: 0, CPUTempRange: "0 - 54", GPUTempRange: "0 - 48"},
	{Step: 2, CPUFanRPM: 1100, GPUFanRPM: 1300, CPUTempRange: "47 - 64", GPUTempRange: "39 - 58"},
	{Step: 3, CPUFanRPM: 1300, GPUFanRPM: 1500, CPUTempRange: "54 - 71", GPUTempRange: "48 - 65"},
	{Step: 4, CPUFanRPM: 1500, GPUFanRPM: 1700, CPUTempRange: "64 - 74", GPUTempRange: "58 - 68"},
	{Step: 5, CPUFanRPM: 1800, GPUFanRPM: 2000, CPUTempRange: "71 - 77", GPUTempRange: "65 - 71"},
	{Step: 6, CPUFanRPM: 2100, GPUFanRPM: 2300, CPUTempRange: "74 - 80", GPUTempRange: "68 - 74"},
	{Step: 7, CPUFanRPM: 2300, GPUFanRPM: 2500, CPUTempRange: "77 - 83", GPUTempRange: "71 - 77"},
	{Step: 8, CPUFanRPM: 2300, GPUFanRPM: 2700, CPUTempRange: "80 - 86", GPUTempRange: "75 - 80"},
	{Step: 9, CPUFanRPM: 2700, GPUFanRPM: 2900, CPUTempRange: "83 - 88", GPUTempRange: "77 - 83"},
	{Step: 10, CPUFanRPM: 2900, GPUFanRPM: 3100, CPUTempRange: "86 - 96", GPUTempRange: "80 - 86"},
}

// FanSpeedMaxCPU / FanSpeedMaxGPU mirror the same constants in
// Fan/config.ts. Used by overview fan live to expose the "max RPM"
// column alongside the live RPM reading.
const (
	FanSpeedMaxCPU = 2900
	FanSpeedMaxGPU = 3100
)

// FanCurveRow is one row of FanCurveTable.
type FanCurveRow struct {
	Step         int
	CPUFanRPM    int
	GPUFanRPM    int
	CPUTempRange string
	GPUTempRange string
}
