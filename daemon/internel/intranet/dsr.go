package intranet

import "encoding/binary"

func ipv4Checksum(hdr []byte) uint16 {
	var sum uint32
	// header length is multiple of 2
	for i := 0; i < len(hdr); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(hdr[i : i+2]))
	}
	for (sum >> 16) != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}

// fragmentIPv4 attempts to split an Ethernet frame carrying an IPv4 packet
// into multiple Ethernet frames where each IP fragment fits within the
// given interface MTU. mtu is the interface MTU (i.e., maximum IP packet
// size including IP header). Returns a slice of full ethernet frames ready
// to send. If the frame is not IPv4 or can't be fragmented (DF bit set)
// an error is returned.
func fragmentIPv4(frame []byte, mtu int) ([][]byte, error) {
	// Need at least Ethernet + minimum IP header
	if len(frame) < 14+20 {
		return nil, fmtError("frame too short for IPv4")
	}
	ethType := binary.BigEndian.Uint16(frame[12:14])
	const etherTypeIPv4 = 0x0800
	if ethType != etherTypeIPv4 {
		return nil, fmtError("not an IPv4 ethernet frame")
	}

	ipStart := 14
	verIhl := frame[ipStart]
	if verIhl>>4 != 4 {
		return nil, fmtError("not IPv4")
	}
	ihl := int(verIhl & 0x0f)
	ipHeaderLen := ihl * 4
	if ipHeaderLen < 20 || len(frame) < ipStart+ipHeaderLen {
		return nil, fmtError("invalid ip header length")
	}

	// Read total length from IP header
	totalLen := int(binary.BigEndian.Uint16(frame[ipStart+2 : ipStart+4]))
	if totalLen < ipHeaderLen {
		return nil, fmtError("invalid total length")
	}
	payloadLen := totalLen - ipHeaderLen
	if len(frame) < ipStart+ipHeaderLen+payloadLen {
		// allow pcap frames with extra trailing bytes (FCS); but ensure payload present
		if len(frame) < ipStart+ipHeaderLen {
			return nil, fmtError("frame shorter than ip header")
		}
		// adjust payloadLen to available bytes
		available := len(frame) - (ipStart + ipHeaderLen)
		if available <= 0 {
			return nil, fmtError("no ip payload available")
		}
		payloadLen = available
		totalLen = ipHeaderLen + payloadLen
	}

	// Check DF (Don't Fragment)
	flagsFrag := binary.BigEndian.Uint16(frame[ipStart+6 : ipStart+8])
	const dfMask = 0x4000
	if flagsFrag&dfMask != 0 {
		return nil, fmtError("DF set; cannot fragment")
	}

	// Compute per-fragment payload size: mtu - ipHeaderLen. Must be multiple of 8.
	if mtu <= ipHeaderLen {
		return nil, fmtError("mtu too small for ip header")
	}
	maxPayload := mtu - ipHeaderLen
	// Round down to multiple of 8
	maxPayload = maxPayload &^ 7
	if maxPayload <= 0 {
		return nil, fmtError("mtu too small for fragmentation unit")
	}

	ipHeader := make([]byte, ipHeaderLen)
	copy(ipHeader, frame[ipStart:ipStart+ipHeaderLen])
	payload := make([]byte, payloadLen)
	copy(payload, frame[ipStart+ipHeaderLen:ipStart+ipHeaderLen+payloadLen])

	// Iterate and build fragments
	var frags [][]byte
	offset := 0
	for offset < payloadLen {
		chunk := maxPayload
		if remaining := payloadLen - offset; remaining <= maxPayload {
			chunk = remaining
		}

		// Create new IP header for fragment
		newIP := make([]byte, ipHeaderLen)
		copy(newIP, ipHeader)

		// Set total length
		binary.BigEndian.PutUint16(newIP[2:4], uint16(ipHeaderLen+chunk))

		// Set flags+offset: preserve DF, set MF for non-last
		origFlags := binary.BigEndian.Uint16(ipHeader[6:8])
		df := origFlags & dfMask
		var mf uint16
		if offset+chunk < payloadLen {
			mf = 0x2000
		}
		fragOffset := uint16(offset / 8)
		combined := df | mf | (fragOffset & 0x1fff)
		binary.BigEndian.PutUint16(newIP[6:8], combined)

		// Zero checksum and compute
		newIP[10] = 0
		newIP[11] = 0
		csum := ipv4Checksum(newIP)
		binary.BigEndian.PutUint16(newIP[10:12], csum)

		// Build ethernet frame: copy original ethernet header, but use the modified IP header + fragment payload
		eth := make([]byte, 14)
		copy(eth, frame[:14])
		fragFrame := make([]byte, 14+ipHeaderLen+chunk)
		copy(fragFrame[:14], eth)
		copy(fragFrame[14:14+ipHeaderLen], newIP)
		copy(fragFrame[14+ipHeaderLen:], payload[offset:offset+chunk])

		frags = append(frags, fragFrame)
		offset += chunk
	}

	return frags, nil
}

// fmtError is a tiny helper to produce errors without importing fmt across file
func fmtError(s string) error { return &simpleErr{s} }

type simpleErr struct{ s string }

func (e *simpleErr) Error() string { return e.s }
