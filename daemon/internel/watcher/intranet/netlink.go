//go:build !linux
// +build !linux

package intranet

import "errors"

func getPodNeighborInfo(podIp string) (mac, iface string, err error) {
	return "", "", errors.New("not implemented")
}
