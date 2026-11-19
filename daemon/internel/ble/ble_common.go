package ble

import (
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/beclab/Olares/daemon/internel/wifi"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	connectwifi "github.com/beclab/Olares/daemon/pkg/commands/connect_wifi"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

const (
	// bleScanIntervalActive is used during onboarding (the node is not online
	// yet), when the LarePass app needs a fresh WiFi AP list to connect.
	bleScanIntervalActive = 2 * time.Second
	// bleScanIntervalIdle is used once the node is online, when the AP list
	// rarely matters. Backing off here removes the bulk of the periodic D-Bus
	// WiFi scan CPU cost.
	bleScanIntervalIdle = 30 * time.Second
)

// scanInterval picks the WiFi scan cadence based on network connectivity:
// frequent while onboarding (no network), relaxed once the node is online.
func (s *service) scanInterval() time.Duration {
	if state.CurrentState.WiredConnected || state.CurrentState.WifiConnected {
		return bleScanIntervalIdle
	}
	return bleScanIntervalActive
}

func (s *service) Start() {
	go func() {
		defer s.cancel()
		s.scanOnce()
		lastScan := time.Now()
		// Poll at the active cadence so a connectivity drop is noticed within a
		// couple of seconds, but only run the expensive D-Bus scan when due:
		// every tick while offline, every bleScanIntervalIdle once online.
		for {
			timer := time.NewTimer(bleScanIntervalActive)
			select {
			case <-s.ctx.Done():
				timer.Stop()
				return

			case <-timer.C:
				if time.Since(lastScan) >= s.scanInterval() {
					s.scanOnce()
					lastScan = time.Now()
				}
			}
		}
	}()
}

func (s *service) scanOnce() {
	s.getAPList()
	s.getTerminusInfo()
	if s.update != nil {
		s.update()
	}
}

func (s *service) Stop() {
	s.cancel()
}

func (s *service) SetUpdateApListCB(f func([]AccessPoint)) {
	s.updateApListCB = f
}

func (s *service) connectWifi(value []byte) {
	cctx, err := s.parseConnectConext(value)
	if err != nil {
		state := ConnectState{
			State:  ptr.To(Fail),
			ErrMsg: ptr.To(err.Error()),
		}

		s.wifiConnectState = state.String()
		return
	}

	param := *cctx

	cmd := connectwifi.New()
	_, err = cmd.Execute(s.ctx, &param)
	if err != nil {
		state := ConnectState{
			State:  ptr.To(Fail),
			ErrMsg: ptr.To(err.Error()),
		}

		s.wifiConnectState = state.String()
		return
	}

	state := ConnectState{
		State: ptr.To(OK),
	}

	s.wifiConnectState = state.String()
}

func (s *service) parseConnectConext(value []byte) (*ConnectContext, error) {
	var cctx ConnectContext
	if len(value) > 512 {
		return nil, errors.New("wifi connection request exceeds BLE limit of 512 bytes")
	}

	err := json.Unmarshal(value, &cctx)
	if err != nil {
		klog.Error("parse wifi connect context error, ", err)
		return nil, err
	}

	return &cctx, nil
}

func (s *service) getTerminusInfo() {
	/*
		terminusName
		terminusVersion
		installedTime
		terminusState
		os_type
		hostIp
		device_name
	*/
	res := map[string]interface{}{
		"terminusName":    state.CurrentState.TerminusName,
		"terminusVersion": state.CurrentState.TerminusVersion,
		"installedTime":   state.CurrentState.InstalledTime,
		"terminusState":   state.CurrentState.TerminusState,
		"os_type":         state.CurrentState.OsType,
		"hostIp":          state.CurrentState.HostIP,
		"device_name":     state.CurrentState.DeviceName,
	}
	info, err := json.Marshal(res)
	if err != nil {
		klog.Error("marshal current state error, ", err)
		return
	}

	s.terminusInfo = string(info)
}

func (s *service) getAPList() {
	wm, err := wifi.NewManager()
	if err != nil {
		klog.Error("create wifi manager error, ", err)
		return
	}

	defer wm.Close()
	devices, err := wm.GetWifiDevices()
	if err != nil {
		klog.Errorf("Failed to list wifi devices: %s", err)
		return
	}

	var list []AccessPoint
	for _, device := range devices {
		aps, err := wm.GetAccessPoints(device.Path)
		if err != nil {
			klog.Errorf("Error getting access points list for %s: %s", device.Interface, err)
			continue
		}
		for _, a := range aps {
			list = append(list, AccessPoint{AccessPoint: a, Connected: a.Connected})
		}
	}

	slices.SortFunc(list, func(o1, o2 AccessPoint) int {
		if o2.Connected && !o1.Connected {
			return 1
		}

		if o1.Connected && !o2.Connected {
			return -1
		}

		return int(o2.Strength) - int(o1.Strength)
	})

	s.publishAPList(list)
}

func (s *service) publishAPList(list []AccessPoint) {
	// HTTP receives the complete scan; only the BLE characteristic is bounded.
	if s.updateApListCB != nil {
		s.updateApListCB(list)
	}
	// Merge equivalent networks across radios, keeping the strongest signal.
	// Retain the first occurrence's connection/signal priority in the input list.
	type apKey struct{ ssid, security string }
	positions := make(map[apKey]int, len(list))
	compact := make([]advertisedAP, 0, len(list))
	for _, ap := range list {
		security := append([]wifi.Security(nil), ap.SecurityTypes...)
		slices.Sort(security)
		encodedSecurity, _ := json.Marshal(security)
		key := apKey{ap.SSID, string(encodedSecurity)}
		if index, ok := positions[key]; ok {
			if ap.Strength > compact[index].Strength {
				compact[index].Strength = ap.Strength
			}
			continue
		}
		positions[key] = len(compact)
		compact = append(compact, advertisedAP{SSID: ap.SSID, Strength: ap.Strength, SecurityTypes: security})
	}
	// Measure serialized bytes, including JSON punctuation and escaping. Skip
	// entries that cannot fit so shorter later entries can fill the capacity.
	data := []byte{'['}
	for _, ap := range compact {
		item, err := json.Marshal(ap)
		if err != nil {
			klog.Error("Failed to marshal wifi access point")
			continue
		}
		separator := 0
		if len(data) > 1 {
			separator = 1
		}
		if len(data)+separator+len(item)+1 > 512 {
			continue
		}
		if separator != 0 {
			data = append(data, ',')
		}
		data = append(data, item...)
	}
	s.apList = string(append(data, ']'))
}
