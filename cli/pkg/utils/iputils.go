package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/logger"
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

	ch := make(chan any, len(sites))

	for site := range sites {
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
		}(site)
	}

	tr := time.NewTimer(time.Duration(5*len(sites)+3) * time.Second)

LOOP:
	for i := 0; i < len(sites); i++ {
		select {
		case r, ok := <-ch:
			if !ok {
				continue
			}

			switch v := r.(type) {
			case string:
				ip := net.ParseIP(v)
				if ip != nil && ip.To4() != nil && !ip.IsLoopback() && !ip.IsMulticast() {
					return v
				}
			case error:
				logger.Warnf("got an error, %v", v)
			}
		case <-tr.C:
			tr.Stop()
			logger.Warnf("timed out")
			break LOOP
		}
	}

	return ""
}

func ExtractIP(host string) ([]string, error) {
	var ips []string
	re := regexp.MustCompile(`\(([\d\.]+)\)|from ([\d\.]+):`)
	matches := re.FindStringSubmatch(host)
	if len(matches) > 1 {
		if matches[1] != "" {
			ips = append(ips, matches[1])
		} else if len(matches) > 2 && matches[2] != "" {
			ips = append(ips, matches[2])
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("failed to extract ip from %s", host)
	}

	return ips, nil
}

func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func ExtractIPAddress(addr string) string {
	var ip string
	scanner := bufio.NewScanner(strings.NewReader(addr))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "inet ") {
			continue
		}
		ip = line
		break
	}

	ip = strings.TrimSpace(ip)
	fields := strings.Split(ip, " ")
	ips := strings.Split(fields[1], "/")
	return ips[0]
}
