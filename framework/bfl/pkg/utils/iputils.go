package utils

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"bytetrade.io/web3os/bfl/internal/log"
	"bytetrade.io/web3os/bfl/pkg/constants"
)

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
	XClientIP     = "x-client-ip"
)

func RemoteIp(req *http.Request) string {
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

	type siteConfig struct {
		url           string
		unmarshalFunc func(v []byte) string
	}

	sites := []siteConfig{
		{
			url: constants.APIMyExternalIP,
			unmarshalFunc: func(v []byte) string {
				return strings.TrimSpace(string(v))
			},
		},
		{
			url: "https://httpbin.org/ip",
			unmarshalFunc: func(v []byte) string {
				var hb httpBin
				if err := json.Unmarshal(v, &hb); err == nil && hb.Origin != "" {
					return hb.Origin
				}
				return ""
			},
		},
		{
			url: "https://ifconfig.me/all.json",
			unmarshalFunc: func(v []byte) string {
				var ifMe ifconfigMe
				if err := json.Unmarshal(v, &ifMe); err == nil && ifMe.IPAddr != "" {
					return ifMe.IPAddr
				}
				return ""
			},
		},
		{
			url: "https://myexternalip.com/json",
			unmarshalFunc: func(v []byte) string {
				var extip externalIP
				if err := json.Unmarshal(v, &extip); err == nil && extip.IP != "" {
					return extip.IP
				}
				return ""
			},
		},
	}

	client := http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	for _, site := range sites {
		resp, err := client.Get(site.url)
		if err != nil {
			log.Warnf("failed to get external ip from %s, %v", site.url, err)
			continue
		}

		respBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			log.Warnf("failed to read response from %s, %v", site.url, readErr)
			continue
		}

		ipStr := site.unmarshalFunc(respBytes)
		ip := net.ParseIP(ipStr)
		if ip != nil && ip.To4() != nil && !ip.IsLoopback() && !ip.IsMulticast() {
			return ipStr
		}
	}

	return ""
}
