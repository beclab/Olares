package gateway

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"sort"
	"strings"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Exact-host ownership annotation on SharedRouteRegistry (JSON map host→viewer).
const AnnotationExactHostOwners = "gateway.olares.io/exact-host-owners"

const (
	customDomainCertLabel      = "app.bytetrade.io/custom-domain-cert"
	customDomainCertLabelValue = "true"

	// indexApplicationThirdPartyDomain indexes Application objects by each
	// normalized customDomain.third_party_domain so cert ConfigMap events can
	// enqueue only referring apps (O(refs) not O(all apps)).
	indexApplicationThirdPartyDomain = "gateway.olares.io/third-party-domain"
)

// Deny reasons for EligibleCustomHost (stable; safe for metrics labels).
const (
	DenyInvalidDNS       = "invalid_dns"
	DenyReservedSuffix   = "reserved_suffix"
	DenyNoCert           = "no_cert"
	DenyCertNameMismatch = "cert_name_mismatch"
	DenyCNAMENotActive   = "cname_not_active"
)

// ExactHostOwner is one Eligible host owned by a viewer (or app owner).
type ExactHostOwner struct {
	Host  string
	Owner string
}

// CertMaterializer supplies cert/key for a FQDN when settings omit them (e.g. CM).
type CertMaterializer func(host string) (cert, key string, ok bool)

// customDomainEntranceFields is the per-entrance JSON shape under customDomain.
type customDomainEntranceFields struct {
	ThirdPartyDomain  string `json:"third_party_domain"`
	ThirdLevelDomain  string `json:"third_level_domain"`
	Cert              string `json:"cert"`
	Key               string `json:"key"`
	CnameTargetStatus string `json:"cname_target_status"`
	CnameStatus       string `json:"cname_status"`
}

// EligibleCustomHost reports whether a third-party custom domain may be
// projected into SRR / CoreDNS / mesh-in allowlists. materializer is optional
// and used only when settings omit cert/key (ARCH: settings or CM).
func EligibleCustomHost(cfg customDomainEntranceFields, platformDomain string, materializer CertMaterializer) (bool, string) {
	raw := strings.TrimSpace(cfg.ThirdPartyDomain)
	if raw == "" {
		return false, DenyInvalidDNS
	}
	host, err := NormalizeHostPattern(raw)
	if err != nil || strings.Contains(host, "*") {
		return false, DenyInvalidDNS
	}
	if reason := reservedExactHostReason(host, platformDomain); reason != "" {
		return false, reason
	}
	cert := strings.TrimSpace(cfg.Cert)
	key := strings.TrimSpace(cfg.Key)
	if cert == "" || key == "" {
		if materializer != nil {
			if c, k, ok := materializer(host); ok {
				cert, key = c, k
			}
		}
	}
	if cert == "" || key == "" {
		return false, DenyNoCert
	}
	if !certCoversHost(cert, key, host) {
		return false, DenyCertNameMismatch
	}
	if !strings.EqualFold(strings.TrimSpace(cfg.CnameTargetStatus), "set") ||
		!strings.EqualFold(strings.TrimSpace(cfg.CnameStatus), "active") {
		return false, DenyCNAMENotActive
	}
	return true, ""
}

func reservedExactHostReason(host, platformDomain string) string {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return DenyInvalidDNS
	}
	if net.ParseIP(h) != nil {
		return DenyReservedSuffix
	}
	for _, suf := range []string{"localhost", "cluster.local"} {
		if h == suf || strings.HasSuffix(h, "."+suf) {
			return DenyReservedSuffix
		}
	}
	if h == "svc" || strings.HasSuffix(h, ".svc") || strings.Contains(h, ".svc.") {
		return DenyReservedSuffix
	}
	dom := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(platformDomain, ".")))
	if dom != "" {
		if h == dom || strings.HasSuffix(h, "."+dom) {
			return DenyReservedSuffix
		}
	}
	// Reserved system prefixes only when the remainder is under platformDomain
	// (align with reservedThirdLevelDomains), not arbitrary public FQDNs.
	first, rest, ok := strings.Cut(h, ".")
	if ok {
		switch first {
		case "auth", "desktop", "wizard":
			if dom != "" && (rest == dom || strings.HasSuffix(rest, "."+dom)) {
				return DenyReservedSuffix
			}
		}
	}
	return ""
}

func certCoversHost(certPEM, keyPEM, host string) bool {
	if _, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM)); err != nil {
		return false
	}
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil || block.Type != "CERTIFICATE" {
		return false
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}
	if err := cert.VerifyHostname(host); err != nil {
		return false
	}
	return true
}

// CollectEligibleExactHosts returns Eligible third_party_domain hosts for one
// entrance. Duplicate FQDNs keep the lexicographically smallest owner and log.
func CollectEligibleExactHosts(app *appv1alpha1.Application, entranceName, platformDomain string, materializer CertMaterializer) []ExactHostOwner {
	if app == nil || strings.TrimSpace(entranceName) == "" {
		return nil
	}
	byHost := map[string]string{}

	consider := func(owner, blob string) {
		owner = strings.ToLower(strings.TrimSpace(owner))
		if owner == "" || blob == "" {
			return
		}
		cfg, ok := parseEntranceCustomDomain(blob, entranceName)
		if !ok {
			return
		}
		okElig, reason := EligibleCustomHost(cfg, platformDomain, materializer)
		if !okElig {
			klog.V(4).Infof("gateway-tpd: eligible deny app=%s entrance=%s owner=%s reason=%s",
				app.Spec.Name, entranceName, owner, reason)
			return
		}
		host, err := NormalizeHostPattern(cfg.ThirdPartyDomain)
		if err != nil {
			return
		}
		if prev, dup := byHost[host]; dup {
			if owner < prev {
				klog.Warningf("gateway-tpd: duplicate exact host app=%s entrance=%s host_hash=%s keep_owner=%s drop_owner=%s",
					app.Spec.Name, entranceName, shortHostHash(host), owner, prev)
				byHost[host] = owner
			} else if owner != prev {
				klog.Warningf("gateway-tpd: duplicate exact host app=%s entrance=%s host_hash=%s keep_owner=%s drop_owner=%s",
					app.Spec.Name, entranceName, shortHostHash(host), prev, owner)
			}
			return
		}
		byHost[host] = owner
	}

	if blob := app.Spec.Settings["customDomain"]; blob != "" {
		owner := strings.ToLower(strings.TrimSpace(app.Spec.Owner))
		if owner == "" {
			klog.Warningf("gateway-tpd: app=%s has customDomain in settings but empty spec.owner; skipping settings blob",
				app.Spec.Name)
		} else {
			consider(owner, blob)
		}
	}
	users := make([]string, 0, len(app.Spec.UserSettings))
	for user := range app.Spec.UserSettings {
		users = append(users, user)
	}
	sort.Strings(users)
	for _, user := range users {
		us := app.Spec.UserSettings[user]
		if us == nil {
			continue
		}
		consider(user, us["customDomain"])
	}
	return sortedExactHostOwners(byHost)
}

// CollectUserThirdLevelExactHosts projects per-owner third_level_domain as
// exact Hosts <prefix>.<owner>.<platformDomain> so Shared overlays do not
// broadcast other users' prefixes via logical prefix.*.
func CollectUserThirdLevelExactHosts(app *appv1alpha1.Application, entranceName, platformDomain string) []ExactHostOwner {
	if app == nil || strings.TrimSpace(entranceName) == "" {
		return nil
	}
	platformDomain = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(platformDomain, ".")))
	if platformDomain == "" {
		return nil
	}
	byHost := map[string]string{}

	consider := func(owner, blob string) {
		owner = strings.ToLower(strings.TrimSpace(owner))
		if owner == "" || blob == "" {
			return
		}
		cfg, ok := parseEntranceCustomDomain(blob, entranceName)
		if !ok {
			return
		}
		prefix := strings.ToLower(strings.TrimSpace(cfg.ThirdLevelDomain))
		if prefix == "" || strings.Contains(prefix, ".") || strings.Contains(prefix, "*") {
			return
		}
		switch prefix {
		case "auth", "desktop", "wizard":
			return
		}
		host := fmt.Sprintf("%s.%s.%s", prefix, owner, platformDomain)
		norm, err := NormalizeHostPattern(host)
		if err != nil {
			klog.Warningf("gateway-tpd: third_level exact host normalize failed entrance=%s: %v", entranceName, err)
			return
		}
		if prev, dup := byHost[norm]; dup {
			if owner < prev {
				byHost[norm] = owner
			}
			return
		}
		byHost[norm] = owner
	}

	if blob := app.Spec.Settings["customDomain"]; blob != "" {
		owner := strings.ToLower(strings.TrimSpace(app.Spec.Owner))
		if owner != "" {
			consider(owner, blob)
		}
	}
	users := make([]string, 0, len(app.Spec.UserSettings))
	for user := range app.Spec.UserSettings {
		users = append(users, user)
	}
	sort.Strings(users)
	for _, user := range users {
		us := app.Spec.UserSettings[user]
		if us == nil {
			continue
		}
		consider(user, us["customDomain"])
	}
	return sortedExactHostOwners(byHost)
}

// MergeExactHostOwners unions owner maps; on conflict keeps lexicographically
// smaller owner.
func MergeExactHostOwners(parts ...[]ExactHostOwner) []ExactHostOwner {
	byHost := map[string]string{}
	for _, part := range parts {
		for _, o := range part {
			host := strings.ToLower(strings.TrimSpace(o.Host))
			owner := strings.ToLower(strings.TrimSpace(o.Owner))
			if host == "" || owner == "" {
				continue
			}
			if prev, ok := byHost[host]; ok {
				if owner < prev {
					byHost[host] = owner
				}
				continue
			}
			byHost[host] = owner
		}
	}
	return sortedExactHostOwners(byHost)
}

func sortedExactHostOwners(byHost map[string]string) []ExactHostOwner {
	if len(byHost) == 0 {
		return nil
	}
	hosts := make([]string, 0, len(byHost))
	for h := range byHost {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)
	out := make([]ExactHostOwner, 0, len(hosts))
	for _, h := range hosts {
		out = append(out, ExactHostOwner{Host: h, Owner: byHost[h]})
	}
	return out
}

func shortHostHash(host string) string {
	h := strings.ToLower(strings.TrimSpace(host))
	if len(h) <= 12 {
		return h
	}
	return h[:8] + "…"
}

func parseEntranceCustomDomain(blob, entranceName string) (customDomainEntranceFields, bool) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(blob), &raw); err != nil {
		klog.Warningf("gateway-tpd: customDomain unmarshal: %v", err)
		return customDomainEntranceFields{}, false
	}
	entry, ok := raw[entranceName]
	if !ok || len(entry) == 0 {
		return customDomainEntranceFields{}, false
	}
	var cfg customDomainEntranceFields
	if err := json.Unmarshal(entry, &cfg); err != nil {
		klog.Warningf("gateway-tpd: customDomain entrance unmarshal: %v", err)
		return customDomainEntranceFields{}, false
	}
	return cfg, true
}

// EncodeExactHostOwnersJSON serializes host→owner for the SRR annotation.
func EncodeExactHostOwnersJSON(owners []ExactHostOwner) string {
	if len(owners) == 0 {
		return ""
	}
	m := make(map[string]string, len(owners))
	for _, o := range owners {
		if o.Host == "" || o.Owner == "" {
			continue
		}
		m[o.Host] = o.Owner
	}
	if len(m) == 0 {
		return ""
	}
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}

// ParseExactHostOwnersJSON decodes the SRR ownership annotation.
func ParseExactHostOwnersJSON(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.ToLower(strings.TrimSpace(v))
		if k == "" || v == "" {
			continue
		}
		out[k] = v
	}
	return out
}

// CollectApplicationThirdPartyDomains returns every normalized
// customDomain.third_party_domain across Spec.Settings and UserSettings.
// Used for field indexing so cert ConfigMap changes enqueue only referring apps.
func CollectApplicationThirdPartyDomains(app *appv1alpha1.Application) []string {
	if app == nil {
		return nil
	}
	seen := map[string]struct{}{}
	addBlob := func(blob string) {
		if blob == "" {
			return
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal([]byte(blob), &raw); err != nil {
			return
		}
		for _, entry := range raw {
			if len(entry) == 0 {
				continue
			}
			var cfg customDomainEntranceFields
			if err := json.Unmarshal(entry, &cfg); err != nil {
				continue
			}
			host, err := NormalizeHostPattern(cfg.ThirdPartyDomain)
			if err != nil || host == "" {
				continue
			}
			seen[host] = struct{}{}
		}
	}
	addBlob(app.Spec.Settings["customDomain"])
	for _, us := range app.Spec.UserSettings {
		if us == nil {
			continue
		}
		addBlob(us["customDomain"])
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for h := range seen {
		out = append(out, h)
	}
	sort.Strings(out)
	return out
}

// NewConfigMapCertMaterializer returns a CertMaterializer that reads labeled
// custom-domain-cert ConfigMaps whose Data["zone"] matches the host.
func NewConfigMapCertMaterializer(ctx context.Context, c client.Client) CertMaterializer {
	if c == nil {
		return nil
	}
	type pair struct{ cert, key string }
	cache := map[string]pair{}
	loaded := false
	load := func() {
		if loaded {
			return
		}
		list := &corev1.ConfigMapList{}
		if err := c.List(ctx, list, client.MatchingLabels{customDomainCertLabel: customDomainCertLabelValue}); err != nil {
			// Do not set loaded: retry on the next materializer call in this reconcile.
			klog.Errorf("gateway-tpd: list custom-domain-cert ConfigMaps failed: %v", err)
			return
		}
		for i := range list.Items {
			cm := &list.Items[i]
			zone := strings.ToLower(strings.TrimSpace(cm.Data["zone"]))
			cert := strings.TrimSpace(cm.Data["cert"])
			key := strings.TrimSpace(cm.Data["key"])
			if zone == "" || cert == "" || key == "" {
				continue
			}
			cache[zone] = pair{cert: cert, key: key}
		}
		loaded = true
	}
	return func(host string) (string, string, bool) {
		load()
		host = strings.ToLower(strings.TrimSpace(host))
		p, ok := cache[host]
		if !ok {
			return "", "", false
		}
		return p.cert, p.key, true
	}
}
