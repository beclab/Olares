//go:build !linux
// +build !linux

package utils

import "k8s.io/klog/v2"

func GetGpuInfo() (*string, error) {
	klog.Warning("not implement")

	return nil, nil
}
