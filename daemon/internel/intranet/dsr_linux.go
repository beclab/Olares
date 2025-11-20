//go:build linux && amd64
// +build linux,amd64

package intranet

import (
	"fmt"
	"syscall"

	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/mdlayher/raw"
	"k8s.io/klog/v2"
)

type DSRProxy struct {
	vip             net.IP
	vipInterface    *net.Interface
	backendIP       net.IP
	backendMAC      net.HardwareAddr
	calicoInterface *net.Interface // calico interface for backend IP

	configChanged bool

	pcapHandle   *pcap.Handle
	responseConn *raw.Conn
	backendConn  *raw.Conn

	closed bool
	mu     sync.Mutex
	stopCh chan struct{}

	requestPortMap *sync.Map // map[uint16]uint16
}

func NewDSRProxy() *DSRProxy {
	return &DSRProxy{
		stopCh:         make(chan struct{}),
		requestPortMap: new(sync.Map),
	}
}

func (d *DSRProxy) WithVIP(vip string, intf string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var err error
	if d.vip != nil && d.vip.String() == vip &&
		d.vipInterface != nil && d.vipInterface.Name == intf {
		return nil
	}

	d.configChanged = true
	d.vip = net.ParseIP(vip)
	d.vipInterface, err = net.InterfaceByName(intf)
	if err != nil {
		klog.Error("parse VIP interface failed:", err)
		return err
	}
	return nil
}

func (d *DSRProxy) WithBackend(backendIP string, backendMAC string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var err error
	if d.backendIP != nil && d.backendIP.String() == backendIP &&
		d.backendMAC != nil && d.backendMAC.String() == backendMAC {
		return nil
	}

	d.configChanged = true
	d.backendIP = net.ParseIP(backendIP)
	d.backendMAC, err = net.ParseMAC(backendMAC)
	if err != nil {
		klog.Error("parse backend MAC failed:", err)
		return err
	}
	return nil
}

func (d *DSRProxy) WithCalicoInterface(intf string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var err error
	if d.calicoInterface != nil && d.calicoInterface.Name == intf {
		return nil
	}

	d.configChanged = true
	d.calicoInterface, err = net.InterfaceByName(intf)
	if err != nil {
		klog.Error("parse calico interface failed:", err)
		return err
	}
	return nil
}

func (d *DSRProxy) Close() {

	if d.pcapHandle != nil {
		d.pcapHandle.Close()
		d.pcapHandle = nil
	}
	if d.responseConn != nil {
		d.responseConn.Close()
		d.responseConn = nil
	}
	if d.backendConn != nil {
		d.backendConn.Close()
		d.backendConn = nil
	}

	d.closed = true
}

func (d *DSRProxy) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.closed {
		d.Close()
	}

	close(d.stopCh)
	return nil
}

func (d *DSRProxy) start() error {
	if err := func() error {
		d.mu.Lock()
		defer d.mu.Unlock()
		if d.pcapHandle == nil || d.responseConn == nil || d.backendConn == nil {
			return errors.New("dsr proxy not configured")
		}

		return nil
	}(); err != nil {
		return err
	}

	log.Printf("Will send requests via: %s, responses via: %s", d.calicoInterface.Name, d.vipInterface.Name)

	packetSource := gopacket.NewPacketSource(d.pcapHandle, d.pcapHandle.LinkType())
	packets := packetSource.Packets()

	log.Println("start dsr proxy on", d.vipInterface.Name, "vip", d.vip)

	for {
		select {
		case p, ok := <-packets:
			if !ok {
				klog.Error("read packets failed")
				return errors.New("read packets error")
			}

			// raw packet bytes
			data := p.Data()
			// safety
			if len(data) < 14 {
				continue
			}

			// Determine if this is a request (to VIP) or response (from backend)
			isResponse := false
			if len(data) >= 14+20 {
				ethType := binary.BigEndian.Uint16(data[12:14])
				if ethType == 0x0800 { // IPv4
					ipStart := 14
					srcIP := net.IP(data[ipStart+12 : ipStart+16])
					dstIP := net.IP(data[ipStart+16 : ipStart+20])
					protocol := data[ipStart+9]

					// Check if this is a response from backend (direct or NAT'd)
					// Case 1: Direct response from backend IP
					if srcIP.Equal(d.backendIP) {
						isResponse = true
						log.Printf("=== RESPONSE PACKET from backend %s (direct) ===", d.backendIP)
						break
					}

					// Case 2: NAT'd response from VIP with wrong source port
					// This is UDP from VIP but destination port is not 53
					if !isResponse && srcIP.Equal(d.vip) && protocol == 17 {
						// Check UDP header
						verIhl := data[ipStart]
						ihl := int(verIhl & 0x0f)
						ipHeaderLen := ihl * 4
						if len(data) >= ipStart+ipHeaderLen+8 {
							udpStart := ipStart + ipHeaderLen
							srcPort := binary.BigEndian.Uint16(data[udpStart : udpStart+2])
							dstPort := binary.BigEndian.Uint16(data[udpStart+2 : udpStart+4])

							// If source is VIP, it's UDP, and source port is NOT 53
							// but destination port suggests this is a DNS response (>1024)
							// This is likely a NAT'd DNS response that we need to fix
							if srcPort != 53 && dstPort > 1024 {
								if _, ok := d.requestPortMap.Load(dstPort); !ok {
									continue
								}
								d.requestPortMap.Delete(dstPort)
								isResponse = true
								log.Printf("=== RESPONSE PACKET from VIP (NAT'd, fixing port %d->53) ===", srcPort)
							}
						}
					}

					if !isResponse {
						if dstIP.Equal(d.vip) {
							log.Printf("=== REQUEST PACKET to VIP %s ===", d.vip)
						} else {
							continue
						}
					}
				}
			}

			// Handle response packets (from backend to client)
			if isResponse {
				d.handleResponse(data, d.responseConn)
				continue
			}

			// Skip packets that are already destined to backend MAC
			// (these are packets we've already modified and re-injected)
			// This prevents forwarding loops
			if bytes.Equal(data[0:6], d.backendMAC) {
				log.Printf("Skipping: packet already forwarded to backend MAC")
				continue
			}

			log.Printf("Intercepted packet: src=%s, dst=%s, len=%d",
				net.HardwareAddr(data[6:12]), net.HardwareAddr(data[0:6]), len(data))

			// Debug: Print original packet details
			if klog.V(8).Enabled() {
				if len(data) >= 14+20 {
					ethType := binary.BigEndian.Uint16(data[12:14])
					if ethType == 0x0800 {
						ipStart := 14
						srcIP := net.IP(data[ipStart+12 : ipStart+16])
						dstIP := net.IP(data[ipStart+16 : ipStart+20])
						oldChecksum := binary.BigEndian.Uint16(data[ipStart+10 : ipStart+12])
						log.Printf("BEFORE: src_ip=%s, dst_ip=%s, ip_checksum=0x%04x", srcIP, dstIP, oldChecksum)
						// Print first 20 bytes of IP header in hex
						log.Printf("BEFORE IP header (hex): % x", data[ipStart:ipStart+20])
					}
				}
			}

			// Rewrite ethernet header: set destination MAC to backend container MAC
			// Source MAC will be the send interface MAC (Calico veth host side)
			copy(data[0:6], d.backendMAC)                    // dst = container MAC
			copy(data[6:12], d.calicoInterface.HardwareAddr) // src = Calico veth host side MAC

			// rewrite IP destination address (critical for backend to accept the packet)
			if len(data) >= 14+20 {
				ethType := binary.BigEndian.Uint16(data[12:14])
				if ethType == 0x0800 { // IPv4
					ipStart := 14
					verIhl := data[ipStart]
					ihl := int(verIhl & 0x0f)
					ipHeaderLen := ihl * 4

					if ipHeaderLen >= 20 && len(data) >= ipStart+ipHeaderLen {
						// Get protocol
						protocol := data[ipStart+9]

						// Replace destination IP with backend IP
						oldDstIP := make([]byte, 4)
						copy(oldDstIP, data[ipStart+16:ipStart+20])
						srcIP := net.IP(data[ipStart+12 : ipStart+16])
						copy(data[ipStart+16:ipStart+20], d.backendIP.To4())

						log.Printf("Rewriting IP: src=%s, dst=%s->%s, proto=%d",
							srcIP, net.IP(oldDstIP), d.backendIP, protocol)

						// Recalculate IP checksum
						data[ipStart+10] = 0
						data[ipStart+11] = 0
						csum := ipv4Checksum(data[ipStart : ipStart+ipHeaderLen])
						binary.BigEndian.PutUint16(data[ipStart+10:ipStart+12], csum)

						log.Printf("New IP checksum: 0x%04x", csum)

						// For UDP (protocol 17), recalculate UDP checksum
						if protocol == 17 && len(data) >= ipStart+ipHeaderLen+8 {
							udpStart := ipStart + ipHeaderLen
							// UDP checksum is optional for IPv4, can be set to 0
							// But if present, we need to update it
							oldChecksum := binary.BigEndian.Uint16(data[udpStart+6 : udpStart+8])
							if oldChecksum != 0 {
								// For simplicity, set UDP checksum to 0 (valid for IPv4)
								data[udpStart+6] = 0
								data[udpStart+7] = 0
								log.Printf("UDP checksum set to 0 (was 0x%04x)", oldChecksum)
							}
						}
					}
				}
			}

			// Debug: Print modified packet details
			if klog.V(8).Enabled() {
				if len(data) >= 14+20 {
					ethType := binary.BigEndian.Uint16(data[12:14])
					if ethType == 0x0800 {
						ipStart := 14
						srcIP := net.IP(data[ipStart+12 : ipStart+16])
						dstIP := net.IP(data[ipStart+16 : ipStart+20])
						newChecksum := binary.BigEndian.Uint16(data[ipStart+10 : ipStart+12])
						log.Printf("AFTER: src_ip=%s, dst_ip=%s, ip_checksum=0x%04x", srcIP, dstIP, newChecksum)
						// Print first 20 bytes of IP header in hex
						log.Printf("AFTER IP header (hex): % x", data[ipStart:ipStart+20])
					}
				}
			}

			// Extract UDP source port for tracking
			if len(data) >= 14+20 {
				ethType := binary.BigEndian.Uint16(data[12:14])
				if ethType == 0x0800 { // IPv4
					ipStart := 14
					verIhl := data[ipStart]
					ihl := int(verIhl & 0x0f)
					ipHeaderLen := ihl * 4
					protocol := data[ipStart+9]

					// For UDP (protocol 17), extract source port
					if protocol == 17 && len(data) >= ipStart+ipHeaderLen+2 {
						udpStart := ipStart + ipHeaderLen
						srcPort := binary.BigEndian.Uint16(data[udpStart : udpStart+2])
						d.requestPortMap.Store(srcPort, 1)
					}
				}
			}

			// send modified frame
			log.Printf("Forwarding to backend: MAC=%s, IP=%s", d.backendMAC, d.backendIP)
			// If the frame is larger than the interface MTU + ethernet header,
			// attempt IPv4 fragmentation and send fragments. For non-IPv4
			// frames we can't fragment at L2, so skip them.
			maxFrame := d.vipInterface.MTU + 14 // interface MTU (IP payload + IP header must fit in MTU) + ethernet header
			if len(data) > maxFrame {
				frags, err := fragmentIPv4(data, d.vipInterface.MTU)
				if err != nil {
					log.Printf("fragment error: %v, skipping frame (len=%d, max=%d)", err, len(data), maxFrame)
					continue
				}

				addr := &raw.Addr{HardwareAddr: d.backendMAC}
				for _, f := range frags {
					if _, err := d.backendConn.WriteTo(f, addr); err != nil {
						log.Printf("writeto err: %v", err)
					}
				}
				continue
			}

			addr := &raw.Addr{HardwareAddr: d.backendMAC}
			if _, err := d.backendConn.WriteTo(data, addr); err != nil {
				log.Printf("writeto err: %v", err)
			}
		case <-d.stopCh:
			log.Println("stopping")
			return nil
		}
	}
}

func (d *DSRProxy) Start() error {
	go func() {
		var done bool
		for !done {
			if err := d.start(); err != nil {
				time.Sleep(10 * time.Second)
			} else {
				done = true
			}
		}
	}()

	return nil
}

// handleResponse processes response packets from backend, rewriting source IP back to VIP
func (d *DSRProxy) handleResponse(data []byte, conn net.PacketConn) {
	if len(data) < 14+20 {
		return
	}

	ethType := binary.BigEndian.Uint16(data[12:14])
	if ethType != 0x0800 { // Only handle IPv4
		return
	}

	ipStart := 14
	verIhl := data[ipStart]
	ihl := int(verIhl & 0x0f)
	ipHeaderLen := ihl * 4

	if ipHeaderLen < 20 || len(data) < ipStart+ipHeaderLen {
		return
	}

	srcIP := net.IP(data[ipStart+12 : ipStart+16])
	dstIP := net.IP(data[ipStart+16 : ipStart+20])
	protocol := data[ipStart+9]

	log.Printf("Response BEFORE: src_ip=%s, dst_ip=%s, proto=%d", srcIP, dstIP, protocol)

	// Rewrite source IP from backend IP to VIP (if needed)
	if !srcIP.Equal(d.vip) {
		copy(data[ipStart+12:ipStart+16], d.vip.To4())
		log.Printf("Response: Rewriting src_ip %s -> %s", srcIP, d.vip)
	}

	// Fix UDP source port if it's not 53
	if protocol == 17 && len(data) >= ipStart+ipHeaderLen+8 {
		udpStart := ipStart + ipHeaderLen
		srcPort := binary.BigEndian.Uint16(data[udpStart : udpStart+2])

		if srcPort != 53 {
			log.Printf("Response: Fixing UDP src_port %d -> 53", srcPort)
			binary.BigEndian.PutUint16(data[udpStart:udpStart+2], 53)
		}

		// Set UDP checksum to 0 (optional for IPv4)
		data[udpStart+6] = 0
		data[udpStart+7] = 0
	}

	log.Printf("Response AFTER: src_ip=%s, dst_ip=%s", d.vip, dstIP)

	// Recalculate IP checksum
	data[ipStart+10] = 0
	data[ipStart+11] = 0
	csum := ipv4Checksum(data[ipStart : ipStart+ipHeaderLen])
	binary.BigEndian.PutUint16(data[ipStart+10:ipStart+12], csum)

	// Get destination MAC from original packet (client's MAC)
	// The packet is already set up correctly for L2 routing back to client
	// Just send it via the main interface

	// Send back via main interface
	addr := &raw.Addr{HardwareAddr: net.HardwareAddr(data[0:6])}
	if _, err := conn.WriteTo(data, addr); err != nil {
		log.Printf("response writeto err: %v", err)
	} else {
		log.Printf("Response sent back to client MAC=%s", net.HardwareAddr(data[0:6]))
	}
}

func (d *DSRProxy) regonfigure() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.configChanged {
		return nil
	}

	if !d.closed {
		d.Close()
	}

	klog.Info("reconfigure DSR proxy")
	klog.Infof("VIP: %s on interface %s", d.vip.String(), d.vipInterface.Name)
	klog.Infof("Backend: %s with MAC %s", d.backendIP.String(), d.backendMAC.String())
	klog.Infof("Calico interface: %s", d.calicoInterface.Name)

	var err error
	d.pcapHandle, err = pcap.OpenLive(d.vipInterface.Name, 65536, false, pcap.BlockForever)
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
