//go:build !linux
// +build !linux

package utils

import (
	"context"
	"errors"

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

func ManagedAllDevices(ctx context.Context) (map[string]Device, error) {
	klog.Warning("not implement")
	return nil, nil
}

func ManagedDeviceStatus(ctx context.Context) (map[string]Device, error) {
	klog.Warning("not implement")
	return nil, nil
}

func UpdateNetworkTraffic(ctx context.Context) {
	klog.Warning("not implement")
}

func GetInterfaceTraffic(iface string) (rxBytes, txBytes float64, err error) {
	return 0, 0, nil
}

func GetEthernetConnection(ctx context.Context) (iface, ifUUID, connection string, err error) {
	return "", "", "", errors.New("not implemented")
}

func FindBridgeConnection(ctx context.Context) (*BridgeConnection, error) {
	return nil, errors.New("not implemented")
}

func CreateBridgeConnection(ctx context.Context) error {
	return errors.New("not implemented")
}

func ResetBridgeConnection(ctx context.Context) error {
	return errors.New("not implemented")
}

func CheckOverlayGatewayStatus(ctx context.Context) error {
	return errors.New("not implemented")
}
