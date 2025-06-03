package utils

import "strings"

type storageDevice struct {
	DevPath       string
	Vender        string
	IDSerial      string
	IDSerialShort string
	PartitionUUID string
}

type mountedPath struct {
	Path          string     `json:"path"`
	Type          DeviceType `json:"type"`
	Invalid       bool       `json:"invalid"`
	IDSerial      string
	IDSerialShort string
	PartitionUUID string
	Device        string
	ReadOnly      bool
}

func SubpathOfMountedPath(path string) func(mountedPath) bool {
	return func(mp mountedPath) bool {
		if mp.Type != SMB {
			return false
		}

		pathToken := strings.Split(path, "/")
		deviceToken := strings.Split(mp.Device, "/")

		if len(pathToken) < len(deviceToken) {
			return false
		}

		for i, p := range deviceToken {
			if p != pathToken[i] {
				return false
			}
		}

		return true
	}
}

type DeviceType string

const (
	USB DeviceType = "usb"
	HDD DeviceType = "hdd"
	SMB DeviceType = "smb"
)
