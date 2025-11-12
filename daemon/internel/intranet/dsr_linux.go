//go:build linux
// +build linux

package intranet

import (
	"fmt"
	"syscall"

	"github.com/google/gopacket/pcap"
	"github.com/mdlayher/raw"
	"k8s.io/klog/v2"
)

func (d *DSRProxy) regonfigure() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.closed {
		d.Close()
	}

	if !d.configChanged {
		return nil
	}

	klog.Info("reconfigure DSR proxy")

	var err error
	d.pcapHandle, err = pcap.OpenLive(d.vipInterface.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		klog.Error("pcap openlive failed:", err)
		return err
	}

	bpf := fmt.Sprintf("(dst host %s and dst port 53) or (src host %s and udp)",
		d.vip.String(), d.vip.String())
	if err := d.pcapHandle.SetBPFFilter(bpf); err != nil {
		klog.Errorf("error: set bpf failed: %v", err)
		return err
	}

	d.backendConn, err = raw.ListenPacket(d.calicoInterface, syscall.ETH_P_ALL, nil)
	if err != nil {
		klog.Errorf("raw listen on send interface: %v", err)
		return err
	}

	d.responseConn, err = raw.ListenPacket(d.vipInterface, syscall.ETH_P_ALL, nil)
	if err != nil {
		klog.Errorf("raw listen on response interface: %v", err)
		return err
	}

	d.closed = false
	d.configChanged = false

	return nil
}
