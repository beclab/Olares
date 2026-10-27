package network

import (
	"fmt"
	"net"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/constants"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	Namespace             = "os-network"
	ConfigName            = "default-network"
	CertificateName       = "default-domain-tls"
	DomainCredentialsName = "default-domain-credentials"
	ProxyCredentialsName  = "default-proxy-credentials"
	FRPCConfigName        = "default-frpc-config"
	AgentName             = "default-frpc"
	DomainUserAnnotation  = "bytetrade.io/domain-user"
	DomainKey             = "domainName"
	ConfigKey             = "proxy"
	StatusKey             = "status"
	NATGatewayIPKey       = "nat-gateway-ip"
	LocalIPKey            = "local-ip"
	LocalDNSIPKey         = "local-dns-ip"
	L4ReadyKey            = "l4-ready"
	PublicIPKey           = "public-ip"
	PublicIPErrorKey      = "public-ip-error"
	ModeNone              = "none"
	ModePublicIP          = "public-ip"
	ModeFRP               = "frp"
)

type DomainBinding struct {
	Domain   string `json:"domain"`
	Username string `json:"username,omitempty"`
}

// The first binding is machine-wide. Additional user bindings will have their
// own configuration and credentials instead of changing this default binding.
type ProxyConfig struct {
	Mode          string `json:"mode"`
	FRPServer     string `json:"frp_server,omitempty"`
	FRPPort       int    `json:"frp_port,omitempty"`
	FRPAuthMethod string `json:"frp_auth_method,omitempty"`
	FRPAuthToken  string `json:"frp_auth_token,omitempty"`
	DNSTarget     string `json:"dns_target,omitempty"`
}

type ProxySpec struct {
	Config   ProxyConfig `json:"config"`
	Revision int64       `json:"revision"`
	Username string      `json:"username,omitempty"`
}

type ProxyStatus struct {
	Phase            string `json:"phase"`
	Message          string `json:"message,omitempty"`
	DNSTarget        string `json:"dns_target,omitempty"`
	ObservedRevision int64  `json:"observed_revision"`
}

type DomainStatus struct {
	Phase                string `json:"phase"`
	Message              string `json:"message,omitempty"`
	CertificateReady     bool   `json:"certificate_ready"`
	CertificateExpiresAt string `json:"certificate_expires_at,omitempty"`
	DNSConfigured        bool   `json:"dns_configured"`
	DNSTarget            string `json:"dns_target,omitempty"`
}

type Status struct {
	Domain    DomainStatus `json:"domain"`
	Proxy     ProxyStatus  `json:"proxy"`
	UpdatedAt string       `json:"updated_at,omitempty"`
}

type ProxyView struct {
	Config       ProxyConfig `json:"config"`
	Username     string      `json:"username,omitempty"`
	HasAuthToken bool        `json:"has_auth_token"`
	Status       ProxyStatus `json:"status"`
	Entrance     Entrance    `json:"entrance"`
}

type Entrance struct {
	LocalIP       string `json:"local_ip,omitempty"`
	LocalDNSIP    string `json:"local_dns_ip,omitempty"`
	NATGatewayIP  string `json:"nat_gateway_ip,omitempty"`
	L4Ready       bool   `json:"l4_ready"`
	PublicIP      string `json:"public_ip,omitempty"`
	PublicIPError string `json:"public_ip_error,omitempty"`
}

func entranceFromData(data map[string]string) Entrance {
	return Entrance{
		LocalIP: data[LocalIPKey], LocalDNSIP: data[LocalDNSIPKey],
		NATGatewayIP: data[NATGatewayIPKey], L4Ready: data[L4ReadyKey] == "true",
		PublicIP: data[PublicIPKey], PublicIPError: data[PublicIPErrorKey],
	}
}

type DomainView struct {
	DomainBinding
	Status DomainStatus `json:"status"`
}

func NormalizeDomain(value string) (string, constants.TerminusName, error) {
	domain := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
	for _, root := range []string{"olares.com", "olares.cn"} {
		prefix := strings.TrimSuffix(domain, "."+root)
		if prefix != domain && len(validation.IsDNS1123Label(prefix)) == 0 {
			return domain, constants.NewTerminusName(prefix, root), nil
		}
	}
	return "", "", fmt.Errorf("domain must be an Olares-managed domain such as name.olares.com or name.olares.cn")
}

func (c *ProxyConfig) Normalize() error {
	c.Mode = strings.ToLower(strings.TrimSpace(c.Mode))
	switch c.Mode {
	case ModeNone:
		*c = ProxyConfig{Mode: ModeNone}
	case ModePublicIP:
		*c = ProxyConfig{Mode: ModePublicIP}
	case ModeFRP:
		c.FRPServer = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(c.FRPServer)), ".")
		if !validTarget(c.FRPServer) {
			return fmt.Errorf("frp_server must be a hostname or IP address")
		}
		if c.FRPPort == 0 {
			c.FRPPort = 7000
		}
		if c.FRPPort < 1 || c.FRPPort > 65535 {
			return fmt.Errorf("invalid frp_port")
		}
		if c.FRPAuthMethod == "" {
			c.FRPAuthMethod = "jws"
		}
		if c.FRPAuthMethod != "jws" && c.FRPAuthMethod != "token" {
			return fmt.Errorf("frp_auth_method must be jws or token")
		}
		c.DNSTarget = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(c.DNSTarget)), ".")
		if c.DNSTarget == "" {
			c.DNSTarget = c.FRPServer
		}
		if !validTarget(c.DNSTarget) {
			return fmt.Errorf("dns_target must be a hostname or IP address")
		}
	default:
		return fmt.Errorf("mode must be none, public-ip or frp; Cloudflare Tunnel is not supported yet")
	}
	return nil
}

func validTarget(value string) bool {
	if value == "" {
		return false
	}
	if ip := net.ParseIP(value); ip != nil {
		return ip.IsGlobalUnicast()
	}
	return len(validation.IsDNS1123Subdomain(value)) == 0
}
