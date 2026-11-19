package ble

import (
	"context"
	"encoding/json"

	"github.com/beclab/Olares/daemon/internel/wifi"
	"k8s.io/klog/v2"
	"tinygo.org/x/bluetooth"
)

type service struct {
	ctx    context.Context
	cancel context.CancelFunc

	wifiConnectState string
	apPages          [bleAPPageCount]string
	terminusInfo     string

	update             func()
	notifyConnectState func()

	updateApListCB func([]AccessPoint)
}

var OlaresServiceUUID = bluetooth.New16BitUUID(0x2345)
var ConnectWifiUUID = bluetooth.New16BitUUID(0x1201)
var ListAPUUID = bluetooth.New16BitUUID(0x1200)
var TerminusInfoUUID = bluetooth.New16BitUUID(0x1199)
var ListAPPage2UUID = bluetooth.New16BitUUID(0x1202)
var ListAPPage3UUID = bluetooth.New16BitUUID(0x1203)
var ListAPPage4UUID = bluetooth.New16BitUUID(0x1204)
var ListAPPage5UUID = bluetooth.New16BitUUID(0x1205)

const (
	// bleAPPageCount is the number of ListAP characteristics; each carries one page.
	bleAPPageCount = 5
	// bleAPPageSize is the maximum serialized size of a single page.
	bleAPPageSize = 512
)

// ListAPPageUUIDs maps page index to its characteristic. Page 1 is the
// original ListAPUUID for compatibility with existing clients; pages 2-5 use
// ListAPPage2UUID through ListAPPage5UUID.
var ListAPPageUUIDs = [bleAPPageCount]bluetooth.UUID{
	ListAPUUID, ListAPPage2UUID, ListAPPage3UUID, ListAPPage4UUID, ListAPPage5UUID,
}

func emptyAPPages() [bleAPPageCount]string {
	var pages [bleAPPageCount]string
	for i := range pages {
		pages[i] = "[]"
	}
	return pages
}

type ExecState int

const (
	Connecting ExecState = -1
	OK         ExecState = iota
	Fail
)

type ConnectState struct {
	State  *ExecState `json:"state,omitempty"`
	ErrMsg *string    `json:"errMsg,omitempty"`
}

type ConnectContext = wifi.ConnectRequest

type AccessPoint struct {
	wifi.AccessPoint `json:",inline"`
	Connected        bool `json:"-"` // Retained for sorting and internal callbacks only.
}

// advertisedAP contains the network name, signal strength and authentication types.
type advertisedAP struct {
	Strength      byte            `json:"st"`
	SSID          string          `json:"ss"`
	SecurityTypes []wifi.Security `json:"se"`
}

func (c ConnectState) String() string {
	encode, err := json.Marshal(c)
	if err != nil {
		klog.Error("marshal error, ", err)
		return ""
	}

	return string(encode)
}
