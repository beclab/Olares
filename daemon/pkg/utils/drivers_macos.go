//go:build (linux && !cgo) || darwin
// +build linux,!cgo darwin

package utils

import (
	"context"

	"k8s.io/klog/v2"
)

func DetectdUsbDevices(ctx context.Context) (usbDevs []storageDevice, err error) {
	klog.Warning("not implement")
	return
}

func DetectdHddDevices(ctx context.Context) (usbDevs []storageDevice, err error) {
	klog.Warning("not implement")
	return
}

func MonitorUsbDevice(ctx context.Context, cb func(action, id string) error) error {
	klog.Warning("not implement")
	return nil
}

func MountedUsbPath(ctx context.Context) ([]string, error) {
	klog.Warning("not implement")
	return nil, nil
}

func MountUsbDevice(ctx context.Context, mountBaseDir string, dev []storageDevice) (mountedPath []string, err error) {
	klog.Warning("not implement")
	return nil, nil
}

func MountedHddPath(ctx context.Context) ([]string, error) {
	klog.Warning("not implement")
	return nil, nil
}

func UmountUsbDevice(ctx context.Context, path string) error {
	klog.Warning("not implement")
	return nil
}

func UmountBrokenMount(ctx context.Context, baseDir string) error {
	klog.Warning("not implement")
	return nil
}

func MountSambaDriver(ctx context.Context, mountBaseDir string, smbPath string, user, pwd string) error {
	klog.Warning("not implement")
	return nil
}

func UmountSambaDriver(ctx context.Context, mountDir string) error {
	klog.Warning("not implement")
	return nil
}

func MountedSambaPath(ctx context.Context) ([]mountedPath, error) {
	klog.Warning("not implement")
	return nil, nil
}

func ForceMountHdd(ctx context.Context) {
	klog.Warning("not implement")
}

func MountedPath(ctx context.Context) ([]mountedPath, error) {
	klog.Warning("not implement")
	return nil, nil
}

func FilterBySerial(serial string) func(dev storageDevice) bool {
	return func(dev storageDevice) bool {
		return dev.IDSerial == serial || dev.IDSerialShort == serial
	}
}
