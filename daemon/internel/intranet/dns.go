package intranet

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/eball/zeroconf"
	"k8s.io/klog/v2"
)

type DNSConfig struct {
	Domain string
}

type mDNSServer struct {
	mu sync.Mutex

	// domains holds the names to answer for, in the dotted form SetHosts is
	// given, e.g. "settings.alice.olares".
	domains map[string]bool

	// A single zeroconf instance serves every domain through host aliases.
	// One instance per domain would mean one socket pair and one stream of
	// multicast announcements per domain, and a deployment easily has a dozen.
	server  *zeroconf.Server
	primary string
	aliases map[string]bool
}

func NewMDNSServer() (*mDNSServer, error) {
	s := &mDNSServer{
		domains: make(map[string]bool),
	}
	return s, nil
}

func (s *mDNSServer) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.shutdown()
}

func (s *mDNSServer) shutdown() {
	if s.server == nil {
		return
	}

	s.server.Shutdown()
	s.server = nil
	s.primary = ""
	s.aliases = nil
	klog.Info("Intranet mDNS server closed")
}

func (s *mDNSServer) StartAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.domains) == 0 {
		return nil
	}

	iface, err := s.findIntranetInterface()
	if err != nil {
		klog.Error("find intranet interface error, ", err)
		return err
	}

	// The service entry fixes its hostname when it is registered, so a change
	// of primary is the one case that forces a rebuild.
	primary := pickPrimaryDomain(s.domains)
	if s.server != nil && s.primary != primary {
		klog.Infof("mDNS primary host changed from %s to %s, restarting", s.primary, primary)
		s.shutdown()
	}

	if s.server == nil {
		disableWirelessPowerSave(iface)

		klog.Infof("Registering mDNS service on host: %s", primary)
		server, err := zeroconf.RegisterAll("olares", "_http._tcp", "local.", primary, 80, []string{"txtv=0", "lo=1", "la=0", "path=/"}, []net.Interface{*iface}, true, false, true)
		if err != nil {
			klog.Errorf("Failed to register mDNS service for host %s: %v", primary, err)
			return err
		}

		s.server = server
		s.primary = primary
		s.aliases = make(map[string]bool)
	}

	s.syncAliases()
	klog.V(8).Info("Intranet mDNS server started")
	return nil
}

// syncAliases makes the running server answer for exactly the current domains.
func (s *mDNSServer) syncAliases() {
	want := make(map[string]bool)
	for domain := range s.domains {
		for _, name := range hostNamesFor(domain) {
			want[name] = true
		}
	}

	for name := range want {
		if s.aliases[name] {
			continue
		}
		if err := s.server.AddHostAlias(name); err != nil {
			klog.Errorf("add host alias %s error, %v", name, err)
			continue
		}
		s.aliases[name] = true
		klog.Info("add host alias, ", name)
	}

	for name := range s.aliases {
		if want[name] {
			continue
		}
		s.server.RemoveHostAlias(name)
		delete(s.aliases, name)
		klog.Info("remove host alias, ", name)
	}
}

// hostNamesFor returns the mDNS names a domain answers to: the dotted form and
// the hyphenated one, for clients that only accept a single label under .local.
func hostNamesFor(domain string) []string {
	return []string{
		domain + ".local.",
		strings.ReplaceAll(domain, ".", "-") + ".local.",
	}
}

// pickPrimaryDomain returns the name to register the service under. The domain
// with the fewest labels is the root one, which outlives the per-app
// subdomains, so choosing it keeps apps coming and going from restarting the
// server.
func pickPrimaryDomain(domains map[string]bool) string {
	var primary string
	for domain := range domains {
		if primary == "" {
			primary = domain
			continue
		}
		labels, best := strings.Count(domain, "."), strings.Count(primary, ".")
		if labels < best || (labels == best && domain < primary) {
			primary = domain
		}
	}
	return primary
}

// disableWirelessPowerSave keeps the radio awake between beacons. With power
// save on, frames are held back until the next DTIM beacon, which delays mDNS
// queries and makes the access point more likely to drop them: multicast is
// never acknowledged and never retransmitted.
func disableWirelessPowerSave(iface *net.Interface) {
	if _, err := os.Stat(filepath.Join("/sys/class/net", iface.Name, "wireless")); err != nil {
		return
	}

	out, err := exec.Command("iw", "dev", iface.Name, "set", "power_save", "off").CombinedOutput()
	if err != nil {
		klog.Warningf("cannot disable wireless power save on %s, %v, %s", iface.Name, err, strings.TrimSpace(string(out)))
		return
	}

	klog.Info("wireless power save disabled on ", iface.Name)
}

// SetHosts sets the hosts for the mDNS server
// if reset is true, it will restart the server before serving them
func (s *mDNSServer) SetHosts(hosts []DNSConfig, reset bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	domains := make(map[string]bool, len(hosts))
	for _, host := range hosts {
		if host.Domain == "" {
			continue
		}
		domains[host.Domain] = true
	}
	s.domains = domains

	if reset {
		s.shutdown()
	}
}

func (s *mDNSServer) findIntranetInterface() (*net.Interface, error) {
	ips, err := nets.GetInternalIpv4Addr()
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, errors.New("cannot get any ip on server")
	}

	hostIp, err := nets.GetHostIp()
	if err != nil {
		klog.Error("get host ip error, ", err)
	}

	// host ip in priority, next is the ethernet ip-
	var (
		iface *net.Interface
	)

	for _, i := range ips {
		if i.IP == hostIp {
			iface = i.Iface
			break
		}
	}

	if iface == nil {
		iface = ips[0].Iface
	}

	return iface, nil
}
