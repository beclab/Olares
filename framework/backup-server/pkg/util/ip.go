package util

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"olares.com/backup-server/pkg/util/log"
)

var ipExtraHeaders = []string{"X-Forwarded-For", "X-Real-IP", "X-Client-IP"}

// RealClientIP returns a string is the http request real client ip address
func RealClientIP(req *http.Request) string {
	remoteAddr := req.RemoteAddr

	for _, header := range ipExtraHeaders {
		if ip := req.Header.Get(header); ip != "" {
			remoteAddr = ip
		}
	}

	if strings.Contains(remoteAddr, ":") {
		ip, _, err := net.SplitHostPort(remoteAddr)
		if err == nil && ip != "" {
			remoteAddr = ip
		}
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

// MyExternalIPAddr get my network outgoing ip address
func MyExternalIPAddr() string {
	sites := map[string]string{
		"httpbin":    "https://httpbin.org/ip",
		"ifconfigme": "https://ifconfig.me/all.json",
		"externalip": "https://myexternalip.com/json",
		"dyndns":     "http://checkip.dyndns.org",
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
		"dyndns": func(v []byte) string {
			reg := regexp.MustCompile("[0-9.]+")

			if r := reg.Find(v); r != nil {
				if ip := string(r); ip != "" {
					return ip
				}
			}
			return ""
		},
	}

	tr := time.NewTimer(time.Duration(5*len(sites)+3) * time.Second)
	ch := make(chan string)

	for site := range sites {
		s := site
		go func(name string) {
			c := http.Client{Timeout: 5 * time.Second}
			resp, err := c.Get(sites[name])
			if err != nil {
				log.Errorf("%s err: %v", name, err)
				return
			}
			defer resp.Body.Close()
			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("%s err: %v", name, err)
				return
			}

			ip := unmarshalFuncs[name](respBytes)
			//println(name, s, ip)
			if ip != "" {
				ch <- ip
			}
		}(s)
	}

	select {
	case r, ok := <-ch:
		if ok && IsValidIP(r) {
			return r
		}
	case <-tr.C:
		tr.Stop()
		log.Warnf("timed out")
	}

	return ""
}

func IsValidIP(s string) bool {
	ip := net.ParseIP(s)

	return ip != nil && ip.IsGlobalUnicast()
}

func IsPrivateIPv4Addr(ip net.IP) bool {
	if ip == nil {
		return false
	}

	ip4 := ip.To4()

	return ip4 != nil && ip4.IsGlobalUnicast() && (ip4[0] == 10 ||
		(ip4[0] == 172 && ip4[1]&0xf0 == 16) ||
		(ip4[0] == 192 && ip4[1] == 168))
}

func GetLocalIPv4Addrs() ([]net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	ipAddrs := make([]net.IP, 0)

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet != nil {
			if ipnet.IP != nil && IsPrivateIPv4Addr(ipnet.IP) {
				ipAddrs = append(ipAddrs, ipnet.IP)
			}
		}
	}
	if len(ipAddrs) == 0 {
		return nil, errors.New("no valid private ipv4 address")
	}

	return ipAddrs, nil
}
