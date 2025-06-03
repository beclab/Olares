package ble

import (
	"context"
	"encoding/json"

	"bytetrade.io/web3os/terminusd/internel/wifi"
	"k8s.io/klog/v2"
	"tinygo.org/x/bluetooth"
)

type service struct {
	ctx    context.Context
	cancel context.CancelFunc

	wifiConnectState string
	apList           string
	terminusInfo     string

	update func()

	updateApListCB func([]AccessPoint)
}

var OlaersServiceUUID = bluetooth.New16BitUUID(0x2345)
var ConnectWifiUUID = bluetooth.New16BitUUID(0x1201)
var ListAPUUID = bluetooth.New16BitUUID(0x1200)
var TerminusInfoUUID = bluetooth.New16BitUUID(0x1199)

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

type ConnectContext struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
}

type AccessPoint struct {
	wifi.AccessPoint `json:",inline"`
	Connected        bool `json:"co"`
}

func (c ConnectState) String() string {
	encode, err := json.Marshal(c)
	if err != nil {
		klog.Error("marshal error, ", err)
		return ""
	}

	return string(encode)
}
