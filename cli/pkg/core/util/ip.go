/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package util

import (
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"encoding/binary"
	"fmt"
	"github.com/libp2p/go-netroute"
	"github.com/pkg/errors"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

func ParseIp(ip string) []string {
	var availableIPs []string
	// if ip is "1.1.1.1/",trim /
	ip = strings.TrimRight(ip, "/")
	if strings.Contains(ip, "/") == true {
		if strings.Contains(ip, "/32") == true {
			aip := strings.Replace(ip, "/32", "", -1)
			availableIPs = append(availableIPs, aip)
		} else {
			availableIPs = GetAvailableIP(ip)
		}
	} else if strings.Contains(ip, "-") == true {
		ipRange := strings.SplitN(ip, "-", 2)
		availableIPs = GetAvailableIPRange(ipRange[0], ipRange[1])
	} else {
		availableIPs = append(availableIPs, ip)
	}
	return availableIPs
}

func GetAvailableIPRange(ipStart, ipEnd string) []string {
	var availableIPs []string

	firstIP := net.ParseIP(ipStart)
	endIP := net.ParseIP(ipEnd)
	if firstIP.To4() == nil || endIP.To4() == nil {
		return availableIPs
	}
	firstIPNum := ipToInt(firstIP.To4())
	EndIPNum := ipToInt(endIP.To4())
	pos := int32(1)

	newNum := firstIPNum

	for newNum <= EndIPNum {
		availableIPs = append(availableIPs, intToIP(newNum).String())
		newNum = newNum + pos
	}
	return availableIPs
}

func GetAvailableIP(ipAndMask string) []string {
	var availableIPs []string

	ipAndMask = strings.TrimSpace(ipAndMask)
	ipAndMask = IPAddressToCIDR(ipAndMask)
	_, ipnet, _ := net.ParseCIDR(ipAndMask)

	firstIP, _ := networkRange(ipnet)
	ipNum := ipToInt(firstIP)
	size := networkSize(ipnet.Mask)
	pos := int32(1)
	max := size - 2 // -1 for the broadcast address, -1 for the gateway address

	var newNum int32
	for attempt := int32(0); attempt < max; attempt++ {
		newNum = ipNum + pos
		pos = pos%max + 1
		availableIPs = append(availableIPs, intToIP(newNum).String())
	}
	return availableIPs
}

func IPAddressToCIDR(ipAddress string) string {
	if strings.Contains(ipAddress, "/") == true {
		ipAndMask := strings.Split(ipAddress, "/")
		ip := ipAndMask[0]
		mask := ipAndMask[1]
		if strings.Contains(mask, ".") == true {
			mask = IPMaskStringToCIDR(mask)
		}
		return ip + "/" + mask
	} else {
		return ipAddress
	}
}

func IPMaskStringToCIDR(netmask string) string {
	netmaskList := strings.Split(netmask, ".")
	var mint []int
	for _, v := range netmaskList {
		strv, _ := strconv.Atoi(v)
		mint = append(mint, strv)
	}
	myIPMask := net.IPv4Mask(byte(mint[0]), byte(mint[1]), byte(mint[2]), byte(mint[3]))
	ones, _ := myIPMask.Size()
	return strconv.Itoa(ones)
}

func networkRange(network *net.IPNet) (net.IP, net.IP) {
	netIP := network.IP.To4()
	firstIP := netIP.Mask(network.Mask)
	lastIP := net.IPv4(0, 0, 0, 0).To4()
	for i := 0; i < len(lastIP); i++ {
		lastIP[i] = netIP[i] | ^network.Mask[i]
	}
	return firstIP, lastIP
}

func networkSize(mask net.IPMask) int32 {
	m := net.IPv4Mask(0, 0, 0, 0)
	for i := 0; i < net.IPv4len; i++ {
		m[i] = ^mask[i]
	}
	return int32(binary.BigEndian.Uint32(m)) + 1
}

func ipToInt(ip net.IP) int32 {
	return int32(binary.BigEndian.Uint32(ip.To4()))
}

func intToIP(n int32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return net.IP(b)
}

func IsValidIPv4Addr(ip net.IP) bool {
	if ip == nil {
		return false
	}
	ip4 := ip.To4()
	return ip4 != nil && ip4.IsGlobalUnicast()
}

func GetValidIPv4AddrsFromOS() ([]net.IP, error) {
	var ifAddrs []net.Addr
	infs, err := net.Interfaces()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get network interfaces")
	}
	for _, inf := range infs {
		if inf.Flags&net.FlagPointToPoint != 0 {
			continue
		}
		addrs, err := inf.Addrs()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get addresses of network interface: %s", inf.Name)
		}
		ifAddrs = append(ifAddrs, addrs...)
	}
	var validIfIPs []net.IP
	for _, ifAddr := range ifAddrs {
		if ipNet, ok := ifAddr.(*net.IPNet); ok && IsValidIPv4Addr(ipNet.IP) {
			validIfIPs = append(validIfIPs, ipNet.IP.To4())
		}
	}
	return validIfIPs, nil
}

// GetLocalIP gets the local ip
// either by getting the "OS_LOCALIP" environment variable
// or by resolving the local hostname to IP address if the env is not set
// in both cases, the IP address is matched against the valid addresses of the local interfaces
// to check if it is an actually bindable address
// by valid it means the address is a non-loopback IPv4 address
// the explicitly specified env var takes precedence than the hostname
// if a valid IP is not found in both cases, a default IP is selected as a fallback
func GetLocalIP() (net.IP, error) {
	validIfIPs, err := GetValidIPv4AddrsFromOS()
	if err != nil {
		return nil, err
	}
	if len(validIfIPs) == 0 {
		return nil, errors.New("no valid IP can be found from local network interfaces")
	}

	if envIPStr := os.Getenv("OS_LOCALIP"); envIPStr != "" {
		envIP := net.ParseIP(envIPStr)
		for _, validIP := range validIfIPs {
			if validIP.Equal(envIP) {
				return validIP, nil
			}
		}
	}

	localHostname, err := os.Hostname()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get local hostname")
	}
	hostIPs, err := net.LookupIP(localHostname)

	// if the hostname can not be looked up
	// it means the user has not set the /etc/hosts
	// in this case we just choose a default IP
	// otherwise it's an operation error we should return
	// because we don't want to select a different IP than what the user specifies
	if err != nil && !strings.Contains(err.Error(), "no such host") {
		return nil, errors.Wrap(err, "failed to resolve local hostname")
	}

	// get the IP address of the interface connected to the default gateway
	// by checking the route table
	r, err := netroute.New()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the default route")
	}
	_, _, defaultRouteSrcIP, err := r.Route(net.IPv4(0, 0, 0, 0))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the default route")
	}
	defaultRouteSrcIP = defaultRouteSrcIP.To4()

	// the IP address of the default route has the highest priority
	sort.Slice(validIfIPs, func(i, j int) bool {
		if defaultRouteSrcIP.Equal(validIfIPs[i]) {
			return true
		}
		return false
	})

	for _, validIP := range validIfIPs {
		for _, hostIP := range hostIPs {
			if validIP.Equal(hostIP) {
				return validIP, nil
			}
		}
	}

	return validIfIPs[0], nil
}

// GetPublicIPsFromOS gets a list of public ips by looking at the local network interfaces
// if any
func GetPublicIPsFromOS() ([]net.IP, error) {
	var validIfPublicIPs []net.IP
	validIfIPs, err := GetValidIPv4AddrsFromOS()
	if err != nil {
		return nil, err
	}
	for _, ip := range validIfIPs {
		if !ip.IsPrivate() {
			validIfPublicIPs = append(validIfPublicIPs, ip)
		}
	}
	return validIfPublicIPs, nil
}

func GetPublicIPFromAWSIMDS() (net.IP, error) {
	token, err := GetTokenFROMAWSIMDS()
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS IMDS token: %v", err)
	}
	url := "http://169.254.169.254/latest/meta-data/public-ipv4"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build http request: %v", err)
	}
	req.Header.Set("X-aws-ec2-metadata-token", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to reach AWS metadata service")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response from AWS metadata service")
	}
	logger.Debugf("retrieved public IP info from AWS metadata service: %s", string(body))
	return net.ParseIP(strings.TrimSpace(string(body))), nil
}

func GetTokenFROMAWSIMDS() (string, error) {
	url := "http://169.254.169.254/latest/api/token"
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build http request: %v", err)
	}
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "600")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to reach AWS metadata service")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response from AWS metadata service")
	}
	return strings.TrimSpace(string(body)), nil
}
