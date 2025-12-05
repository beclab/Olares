package utils

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
	XClientIP     = "x-client-ip"
)

// RemoteIP extracts the remote IP address from the provided http.Request object.
func RemoteIP(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get(XClientIP); ip != "" {
		remoteAddr = ip
	} else if ip := req.Header.Get(XRealIP); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get(XForwardedFor); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

// GetMyExternalIPAddr get my network outgoing ip address
func GetMyExternalIPAddr() string {
	sites := map[string]string{
		"httpbin":    "https://httpbin.org/ip",
		"ifconfigme": "https://ifconfig.me/all.json",
		"externalip": "https://myexternalip.com/json",
	}

	type httpBin struct {
		Origin string `json:"origin"`
	}

	type ifconfigMe struct {
		IPAddr     string `json:"ip_addr"`
		RemoteHost string `json:"remote_host,omitempty"`
		UserAgent  string `json:"user_agent,omitempty"`
		Port       int    `json:"port,omitempty"`
		Method     string `json:"method,omitempty"`
		Encoding   string `json:"encoding,omitempty"`
		Via        string `json:"via,omitempty"`
		Forwarded  string `json:"forwarded,omitempty"`
	}

	type externalIP struct {
		IP string `json:"ip"`
	}

	var unmarshalFuncs = map[string]func(v []byte) string{
		"httpbin": func(v []byte) string {
			var hb httpBin
			if err := json.Unmarshal(v, &hb); err == nil && hb.Origin != "" {
				return hb.Origin
			}
			return ""
		},
		"ifconfigme": func(v []byte) string {
			var ifMe ifconfigMe
			if err := json.Unmarshal(v, &ifMe); err == nil && ifMe.IPAddr != "" {
				return ifMe.IPAddr
			}
			return ""
		},
		"externalip": func(v []byte) string {
			var extip externalIP
			if err := json.Unmarshal(v, &extip); err == nil && extip.IP != "" {
				return extip.IP
			}
			return ""
		},
	}

	ch := make(chan any)

	for name := range sites {
		go func(name string) {
			c := http.Client{Timeout: 5 * time.Second}
			resp, err := c.Get(sites[name])
			if err != nil {
				ch <- err
				return
			}
			defer resp.Body.Close()
			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ch <- err
				return
			}

			ip := unmarshalFuncs[name](respBytes)
			//println(name, site, ip)
			ch <- ip
		}(name)
	}

	for r := range ch {
		if v, ok := r.(string); ok {
			ip := net.ParseIP(v)
			if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
				return v
			}
		}
	}

	return ""
}

func SubnetSplit(n int) map[string]*net.IPNet {
	subnetMap := make(map[string]*net.IPNet)
	log2n := int(math.Ceil(math.Log2(float64(n))))
	alignedN := 1 << log2n
	_, ipNet, _ := net.ParseCIDR("100.64.0.0/10")

	baseIP := ipNet.IP.To4()
	originalMaskLen, _ := ipNet.Mask.Size()

	newMaskLen := originalMaskLen + log2n
	ipsPerSubnet := 1 << (32 - newMaskLen)

	for i := 0; i < alignedN; i++ {
		offset := uint32(i * ipsPerSubnet)
		subnetIP := make(net.IP, 4)
		copy(subnetIP, baseIP)
		for j := 3; j >= 0 && offset > 0; j-- {
			subnetIP[j] += byte(offset & 0xFF)
			offset >>= 8
		}
		firstUsableIP := make(net.IP, 4)
		copy(firstUsableIP, subnetIP)
		firstUsableIP[3]++

		subnet := &net.IPNet{
			IP:   subnetIP,
			Mask: net.CIDRMask(newMaskLen, 32),
		}
		index := strconv.FormatInt(int64(i), 10)
		subnetMap[index] = subnet
	}

	return subnetMap
}
