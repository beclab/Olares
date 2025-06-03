package ble

import (
	"encoding/json"
	"slices"
	"time"

	"bytetrade.io/web3os/terminusd/internel/wifi"
	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"bytetrade.io/web3os/terminusd/pkg/utils"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

func (s *service) Start() {
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				s.cancel()
				ticker.Stop()

			case <-ticker.C:
				s.getAPList()
				s.getTerminusInfo()
				if s.update != nil {
					s.update()
				}
			}
		}
	}()
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

	err = utils.ConnectWifi(s.ctx, cctx.SSID, cctx.Password)
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
			connected := false
			if state.CurrentState.WifiSSID != nil && *state.CurrentState.WifiSSID == a.SSID {
				connected = true
			}
			list = append(list, AccessPoint{a, connected})
		}
	}

	slices.SortFunc(list, func(o1, o2 AccessPoint) int {
		if o2.Connected {
			return 1
		}

		if o1.Connected {
			return -1
		}

		return int(o2.Strength) - int(o1.Strength)
	})

	// marshal list, limit top 5
	var (
		listData []byte
		retList  []AccessPoint
	)
	sublen := len(list)
	for {
		var err error
		retList = list[:sublen]
		listData, err = json.Marshal(retList)
		if err != nil {
			klog.Errorf("Failed to marshal list wifi ap: %s", err)
			return
		}

		// According to the Bluetooth specification, the maximum size of any attribute is 512 bytes.
		if len(listData) <= 512 {
			break
		}

		sublen--
	}

	s.apList = string(listData)

	// update ap list callback
	if s.updateApListCB != nil {
		s.updateApListCB(retList)
	}
}
