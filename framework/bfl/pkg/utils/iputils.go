package utils

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"bytetrade.io/web3os/bfl/internal/log"
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
		"joinolares": "https://myip.joinolares.cn/ip",
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
		"joinolares": func(v []byte) string {
			return strings.TrimSpace(string(v))
		},
	}

	var mu sync.Mutex
	ch := make(chan any, len(sites))
	chSyncOp := func(f func()) {
		mu.Lock()
		defer mu.Unlock()
		if ch != nil {
			f()
		}
	}

	for site := range sites {
		go func(name string) {
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			c := http.Client{Timeout: 5 * time.Second}
			resp, err := c.Get(sites[name])
			if err != nil {
				chSyncOp(func() { ch <- err })
				return
			}
			defer resp.Body.Close()
			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				chSyncOp(func() { ch <- err })
				return
			}

			ip := unmarshalFuncs[name](respBytes)
			//println(name, site, ip)
			chSyncOp(func() { ch <- ip })

		}(site)
	}

	tr := time.NewTimer(time.Duration(5*len(sites)+3) * time.Second)
	defer func() {
		tr.Stop()
		chSyncOp(func() {
			close(ch)
			ch = nil
		})
	}()

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
				log.Warnf("got an error, %v", v)
			}
		case <-tr.C:
			tr.Stop()
			log.Warnf("timed out")
			break LOOP
		}
	}

	return ""
}
