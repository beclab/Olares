//go:build linux
// +build linux

package ble

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func NewBleService(ctx context.Context) (*service, error) {
	err := adapter.Enable()

	if err != nil {
		klog.Error("enable ble adapter error, ", err)
		return nil, err
	}

	adapter.SetConnectHandler(func(device bluetooth.Device, connected bool) {
		// If this is a new device, log info
		if connected {
			klog.Info("new device connected, ", device)
		}
	})

	localName, err := os.Hostname()
	if err != nil {
		klog.Warning("get host name error, ", err)
		localName = "Olare BLE"
	}

	adv := adapter.DefaultAdvertisement()
	adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    localName,
		ServiceUUIDs: []bluetooth.UUID{OlaersServiceUUID},
	})

	klog.Info("starting adapter")
	adv.Start()

	// create a service
	c, cancel := context.WithCancel(ctx)
	s := &service{ctx: c, cancel: cancel}

	var connectWifi bluetooth.Characteristic
	var listAp bluetooth.Characteristic
	var terminusInfo bluetooth.Characteristic

	adapter.AddService(&bluetooth.Service{
		UUID: OlaersServiceUUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				Handle: &connectWifi,
				UUID:   ConnectWifiUUID,
				Value:  []byte(s.wifiConnectState),
				Flags: bluetooth.CharacteristicReadPermission |
					bluetooth.CharacteristicWritePermission |
					bluetooth.CharacteristicWriteWithoutResponsePermission |
					bluetooth.CharacteristicNotifyPermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					if offset != 0 || len(value) == 0 {
						klog.Info("ignore invalid write, ", offset, ", ", string(value))
						return
					}

					var state ConnectState
					if err := json.Unmarshal(value, &state); err == nil && state.State != nil {
						// write state value
						return
					}

					writeData := string(value)

					klog.Info("start to connect, ", writeData)

					// connect
					if err := json.Unmarshal([]byte(s.wifiConnectState), &state); err == nil {
						if *state.State == Connecting {
							// ignore redo connecting
							return
						}
					}

					state = ConnectState{
						State:  ptr.To(Connecting),
						ErrMsg: ptr.To(fmt.Sprintf("start to connect wifi")),
					}

					s.wifiConnectState = state.String()
					connectWifi.Write([]byte(s.wifiConnectState))

					// connect wifi
					go s.connectWifi(value)
				},
			},
			{
				Handle: &listAp,
				UUID:   ListAPUUID,
				Value:  []byte(s.apList),
				Flags: bluetooth.CharacteristicReadPermission |
					bluetooth.CharacteristicNotifyPermission,
			},
			{
				Handle: &terminusInfo,
				UUID:   TerminusInfoUUID,
				Value:  []byte(s.wifiConnectState),
				Flags: bluetooth.CharacteristicReadPermission |
					bluetooth.CharacteristicNotifyPermission,
			},
		},
	})

	s.update = func() {
		listAp.Write([]byte(s.apList))
		terminusInfo.Write([]byte(s.terminusInfo))
		connectWifi.Write([]byte(s.wifiConnectState))
	}

	// wait for signal to close the advertisement
	go func() {
		<-s.ctx.Done()
		adv.Stop()
	}()
	return s, nil
}
