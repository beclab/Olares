package nets

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-netroute"
	pkg_errors "github.com/pkg/errors"
	"github.com/txn2/txeh"
	"k8s.io/klog/v2"
)

type NetInterface struct {
	Iface *net.Interface
	IP    string
}

func GetInternalIpv4Addr() (internalAddrs []*NetInterface, err error) {
	var (
		iefs  []net.Interface
		addrs []net.Addr
	)

	if iefs, err = net.Interfaces(); err != nil {
		klog.Error("list network interfaces error, ", err)
		return
	}

	// get the IP address of the interface connected to the default gateway
	// by checking the route table
	r, err := netroute.New()
	if err != nil {
		return nil, pkg_errors.Wrap(err, "failed to get the default route")
	}

	gatewayInf, _, _, err := r.Route(net.IPv4(0, 0, 0, 0))
	if err != nil {
		return nil, pkg_errors.Wrap(err, "failed to get the default route")
	}

	for _, ief := range iefs {
		switch {
		case strings.HasPrefix(ief.Name, "eth"):
		case strings.HasPrefix(ief.Name, "en"):
		case strings.HasPrefix(ief.Name, "wl"):
		case ief.Name == gatewayInf.Name:
		default:
			continue
		}

		if (ief.Flags & net.FlagUp) == 0 {
			// inactive
			continue
		}

		if addrs, err = ief.Addrs(); err != nil {
			klog.Error("get interface address error, ", err, ", ", ief.Name)
			return
		}

		var ipv4Addr net.IP
		for _, addr := range addrs { // get ipv4 address
			if ipv4Addr = addr.(*net.IPNet).IP.To4(); len(ipv4Addr) > 0 {
				break
			}
		}

		if len(ipv4Addr) == 0 {
			klog.V(8).Infof("interface %s don't have an ipv4 address\n", ief.Name)
			continue
		}

		if !ipv4Addr.IsGlobalUnicast() {
			klog.V(8).Infof("interface %s don't have a valid ipv4 address\n", ief.Name)
			continue
		}

		// ethernet in priority
		if strings.HasPrefix(ief.Name, "eth") {
			internalAddrs = append([]*NetInterface{{&ief, ipv4Addr.String()}}, internalAddrs...)
		} else {
			internalAddrs = append(internalAddrs, &NetInterface{&ief, ipv4Addr.String()})
		}
	}

	return
}

func GetHostIp() (addr string, err error) {
	hostname, err := os.Hostname()
	if err != nil {
		klog.Error("get hostname error, ", err)
		return
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		klog.Error("get host ip error, ", err, ", ", hostname)
		return
	}

	for _, ip := range ips {
		ipv4 := ip.To4()
		if ipv4 != nil {
			addr = ipv4.String()
			return
		}
	}

	err = errors.New("host ip not found")
	return
}

func GetHostIpInterface() (*net.Interface, error) {
	hostIp, err := GetHostIp()
	if err != nil {
		return nil, err
	}

	return GetInterfaceByIp(hostIp)
}

func GetInterfaceByIp(ip string) (*net.Interface, error) {
	var (
		iefs     []net.Interface
		addrs    []net.Addr
		ipv4Addr net.IP
		err      error
	)

	if iefs, err = net.Interfaces(); err != nil {
		klog.Error("list network interfaces error, ", err)
		return nil, err
	}

	for _, ief := range iefs {
		if addrs, err = ief.Addrs(); err != nil {
			klog.Error("get interface address error, ", err, ", ", ief.Name)
			return nil, err
		}

		for _, addr := range addrs { // get ipv4 address
			if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil && ipv4Addr.String() == ip {
				return &ief, nil
			}
		}
	}

	return nil, errors.New("interface is not found")
}

func WriteIpToHostsFile(ip, domain string) error {
	return WriteToHostsFile([]*HostsItem{
		{
			IP:   ip,
			Host: domain,
		},
	})
}

func WriteToHostsFile(items []*HostsItem) error {
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		klog.Error("read hosts file error, ", err)
		return err
	}

	for _, i := range items {
		if _, err := netip.ParseAddr(i.IP); err != nil {
			klog.Error("invalid ip address, ", err, ", ", i.IP)
			return err
		}

		// force update domain
		hosts.RemoveHost(i.Host)

		hosts.AddHost(i.IP, i.Host)
	}

	err = hosts.Save()
	if err != nil {
		klog.Error("save hosts file error, ", err)
	}

	return err
}

func ConflictDomainIpInHostsFile(domain string) (bool, error) {
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		klog.Error("read hosts file error, ", err)
		return false, err
	}

	found := make(map[string]string)
	for _, h := range *hosts.GetHostFileLines() {
		for _, n := range h.Hostnames {
			if n != domain { // don't care
				continue
			}

			if ip, ok := found[n]; ok {
				if ip != h.Address {
					return true, nil
				}
			} else {
				found[n] = h.Address
			}
		}
	}

	// olny all addresses are same of domain should return false
	return false, nil
}

func GetHostIpFromHostsFile(domain string) (string, error) {
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		klog.Error("read hosts file error, ", err)
		return "", err
	}

	_, ip, _ := hosts.HostAddressLookup(domain, txeh.IPFamilyV4)
	return ip, nil
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

	ch := make(chan any, len(sites))

	for site := range sites {
		go func(name string) {
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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
				klog.Warningf("got an error, %v", v)
			}
		case <-tr.C:
			tr.Stop()
			klog.Warning("timed out")
			break LOOP
		}
	}

	return ""
}

func FixHostIP(ip string) error {
	hostname, err := os.Hostname()
	if err != nil {
		klog.Error("get hostname error, ", err)
		return err
	}
	klog.Info("fix host ip: ", ip, ", ", hostname)

	return WriteIpToHostsFile(ip, hostname)
}

func GetHostsFile() ([]HostsItem, error) {
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		klog.Error("read hosts file error, ", err)
		return nil, err
	}

	var found []HostsItem
	var filters filterChain = filterChain{
		filterIpv6,
		filterHostname,
		filterBlacklist,
	}
	for _, h := range *hosts.GetHostFileLines() {
		for _, n := range h.Hostnames {
			item := HostsItem{
				IP:   h.Address,
				Host: n,
			}

			if filters.filter(&item) {
				found = append(found, item)
			}
		}
	}

	return found, nil
}

func ForceUpdateHostsFile(items []*HostsItem) error {
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		klog.Error("read hosts file error, ", err)
		return err
	}

	var filters filterChain = filterChain{
		filterIpv6,
		filterHostname,
		filterBlacklist,
	}

	// delete prev items
	for _, h := range *hosts.GetHostFileLines() {
		for _, n := range h.Hostnames {
			item := HostsItem{
				IP:   h.Address,
				Host: n,
			}

			if filters.filter(&item) {
				hosts.RemoveHost(n)
			}
		}
	}

	// add current items
	for _, i := range items {
		hosts.AddHost(i.IP, i.Host)
	}

	err = hosts.Save()
	if err != nil {
		klog.Error("save hosts file error, ", err)
	}

	return err
}

type filterChain []func(i *HostsItem) bool

func (fc filterChain) filter(i *HostsItem) bool {
	for _, f := range fc {
		if !f(i) {
			return false
		}
	}

	return true
}

func filterIpv6(i *HostsItem) bool {
	ip := net.ParseIP(i.IP)
	return ip.To4() != nil // ipv4 is valid
}

func filterHostname(i *HostsItem) bool {
	hostname, err := os.Hostname()
	if err != nil {
		klog.Error("get hostname error, ", err)
		return false
	}

	return i.Host != hostname
}

func filterBlacklist(i *HostsItem) bool {
	for _, b := range internalHostsItem {
		if strings.Contains(i.Host, b) {
			return false
		}
	}

	return true
}
