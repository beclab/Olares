//go:build !(linux && amd64)
// +build !linux !amd64

package intranet

import (
	"errors"
	"net"
)

type DSRProxy struct {
}

func NewDSRProxy() *DSRProxy {
	return &DSRProxy{}
}

func (d *DSRProxy) WithVIP(vip string, intf string) error {
	return nil
}

func (d *DSRProxy) WithBackend(backendIP string, backendMAC string) error {
	return nil
}

func (d *DSRProxy) WithCalicoInterface(intf string) error {
	return nil
}

func (d *DSRProxy) Close() {}

func (d *DSRProxy) Stop() error {
	return nil
}

func (d *DSRProxy) start() error { return nil }

func (d *DSRProxy) Start() error {
	return nil
}

// handleResponse processes response packets from backend, rewriting source IP back to VIP
func (d *DSRProxy) handleResponse(data []byte, conn net.PacketConn) {}

func (d *DSRProxy) regonfigure() error {
	return errors.New("unsupported operation")
}
