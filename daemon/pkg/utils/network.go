//go:build !linux
// +build !linux

package utils

import (
	"context"

	"k8s.io/klog/v2"
)

func ConnectWifi(ctx context.Context, ssid, password string) error {
	klog.Warning("not implement")
	return nil
}

func EnableWifi(ctx context.Context) error {
	klog.Warning("not implement")
	return nil
}

func GetWifiDevice(ctx context.Context) (map[string]Device, error) {
	klog.Warning("not implement")
	return nil, nil
}

func GetAllDevice(ctx context.Context) (map[string]Device, error) {
	klog.Warning("not implement")
	return nil, nil
}
