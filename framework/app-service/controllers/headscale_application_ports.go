package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/go-resty/resty/v2"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

const (
	headscaleApplicationPortsURL  = "http://headscale-server-svc.os-network:8001/internal/policy/application-ports"
	applicationPortsReconcileName = "application-ports"
)

// applicationPort is one user's dynamic tcp/udp ports on the shared headscale.
// PUT replaces the whole set, so a user omitted here loses their dynamic ports.
type applicationPort struct {
	User string   `json:"user"`
	TCP  []string `json:"tcp"`
	UDP  []string `json:"udp"`
}

type applicationPortsAPI interface {
	Get(ctx context.Context) ([]applicationPort, error)
	Put(ctx context.Context, ports []applicationPort) error
}

type headscalePolicyClient struct {
	url  string
	http *resty.Client
}

func newHeadscalePolicyClient() *headscalePolicyClient {
	return &headscalePolicyClient{
		url:  headscaleApplicationPortsURL,
		http: resty.New().SetTimeout(10 * time.Second),
	}
}

type headscaleEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *headscalePolicyClient) do(ctx context.Context, method string, body any, out any) error {
	req := c.http.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json")
	if body != nil {
		req.SetBody(body)
	}
	resp, err := req.Execute(method, c.url)
	if err != nil {
		klog.Errorf("headscale application ports %s %s failed: %v", method, c.url, err)
		return fmt.Errorf("headscale application ports: %w", err)
	}
	klog.Infof("headscale application ports %s %s -> %d %s", method, c.url, resp.StatusCode(), truncateForLog(resp.String()))
	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return fmt.Errorf("headscale application ports: %s %d: %s", method, resp.StatusCode(), resp.String())
	}
	var envelope headscaleEnvelope
	if err := json.Unmarshal(resp.Body(), &envelope); err != nil {
		return fmt.Errorf("headscale application ports: decode response: %w", err)
	}
	if envelope.Code != 0 {
		return fmt.Errorf("headscale application ports: code %d: %s", envelope.Code, envelope.Message)
	}
	if out != nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("headscale application ports: decode data: %w", err)
		}
	}
	return nil
}

func (c *headscalePolicyClient) Get(ctx context.Context) ([]applicationPort, error) {
	var data struct {
		ApplicationPorts []applicationPort `json:"applicationPorts"`
	}
	if err := c.do(ctx, http.MethodGet, nil, &data); err != nil {
		return nil, err
	}
	if data.ApplicationPorts == nil {
		return []applicationPort{}, nil
	}
	return data.ApplicationPorts, nil
}

func (c *headscalePolicyClient) Put(ctx context.Context, ports []applicationPort) error {
	if ports == nil {
		ports = []applicationPort{}
	}
	body := struct {
		ApplicationPorts []applicationPort `json:"applicationPorts"`
	}{ApplicationPorts: ports}
	return c.do(ctx, http.MethodPut, body, nil)
}

type userPorts struct {
	tcp sets.String
	udp sets.String
}

// applicationPortsFromApps builds the full dynamic port set from every
// Application. Default shared ports stay in headscale's static policy and
// are not added here.
func applicationPortsFromApps(apps []v1alpha1.Application) []applicationPort {
	byUser := map[string]userPorts{}
	for i := range apps {
		user := policyUser(apps[i].Spec.Owner)
		acls := make([]v1alpha1.ACL, 0, len(apps[i].Spec.TailScale.ACLs)+len(apps[i].Spec.TailScaleACLs))
		acls = append(acls, apps[i].Spec.TailScale.ACLs...)
		acls = append(acls, apps[i].Spec.TailScaleACLs...)
		if len(acls) == 0 {
			continue
		}
		klog.Infof("headscale application ports app=%s/%s specName=%s owner=%q aclCount=%d",
			apps[i].Namespace, apps[i].Name, apps[i].Spec.Name, apps[i].Spec.Owner, len(acls))
		if user == "" {
			klog.Infof("skip tailscale acls for app=%s/%s: empty owner", apps[i].Namespace, apps[i].Name)
			continue
		}
		for _, acl := range acls {
			// An empty proto allows TCP and UDP, matching Tailscale ACL behavior.
			// The application-ports API has no other protocol fields.
			protos := protosForACL(acl.Proto)
			if len(protos) == 0 {
				if strings.TrimSpace(acl.Proto) != "" || len(acl.Dst) > 0 {
					klog.Infof("skip tailscale acl proto %q for user %s", acl.Proto, user)
				}
				continue
			}
			klog.Infof("headscale application ports app=%s/%s owner=%s proto=%q dst=%v -> %v",
				apps[i].Namespace, apps[i].Name, user, acl.Proto, acl.Dst, protos)
			for _, dest := range acl.Dst {
				ports := portsFromDestination(dest)
				klog.Infof("headscale application ports app=%s/%s owner=%s dst=%q -> ports=%v",
					apps[i].Namespace, apps[i].Name, user, dest, ports)
				for _, proto := range protos {
					addUserPorts(byUser, user, proto, ports)
				}
			}
		}
	}

	users := make([]string, 0, len(byUser))
	for user := range byUser {
		users = append(users, user)
	}
	sort.Strings(users)

	out := make([]applicationPort, 0, len(users))
	for _, user := range users {
		ports := byUser[user]
		tcp := sortedPortSet(ports.tcp)
		udp := sortedPortSet(ports.udp)
		if len(tcp) == 0 && len(udp) == 0 {
			continue
		}
		out = append(out, applicationPort{User: user, TCP: tcp, UDP: udp})
	}
	if len(out) == 0 {
		klog.Infof("headscale application ports: no dynamic ports from %d applications", len(apps))
	}
	return out
}

func truncateForLog(s string) string {
	const limit = 2048
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "...(truncated)"
}

func protosForACL(proto string) []string {
	switch strings.ToLower(strings.TrimSpace(proto)) {
	case "":
		return []string{"tcp", "udp"}
	case "tcp":
		return []string{"tcp"}
	case "udp":
		return []string{"udp"}
	default:
		return nil
	}
}

func addUserPorts(byUser map[string]userPorts, user, proto string, ports []string) {
	if len(ports) == 0 {
		return
	}
	entry, ok := byUser[user]
	if !ok {
		entry.tcp = sets.NewString()
		entry.udp = sets.NewString()
	}
	switch proto {
	case "tcp":
		entry.tcp.Insert(ports...)
	case "udp":
		entry.udp.Insert(ports...)
	}
	byUser[user] = entry
}

func portsFromDestination(dest string) []string {
	portToken, err := app.ACLDestinationPort(dest)
	if err != nil {
		klog.Warningf("skip tailscale acl destination %q: %v", dest, err)
		return nil
	}
	var out []string
	for _, part := range strings.Split(portToken, ",") {
		spec, ok := canonicalPortSpec(part)
		if !ok {
			klog.Warningf("skip tailscale acl port %q from destination %q", strings.TrimSpace(part), dest)
			continue
		}
		out = append(out, spec)
	}
	return out
}

func canonicalPortSpec(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if s == "" || strings.Contains(s, "*") {
		return "", false
	}
	if strings.Contains(s, "-") {
		bounds := strings.Split(s, "-")
		if len(bounds) != 2 {
			return "", false
		}
		lo, okLo := parsePortNumber(bounds[0])
		hi, okHi := parsePortNumber(bounds[1])
		if !okLo || !okHi || lo > hi {
			return "", false
		}
		return strconv.Itoa(lo) + "-" + strconv.Itoa(hi), true
	}
	n, ok := parsePortNumber(s)
	if !ok {
		return "", false
	}
	return strconv.Itoa(n), true
}

func parsePortNumber(s string) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 1 || n > 65535 {
		return 0, false
	}
	return n, true
}

func policyUser(owner string) string {
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(owner), "@"))
}

func sortedPortSet(s sets.String) []string {
	if s == nil || s.Len() == 0 {
		return []string{}
	}
	return s.List()
}

func applicationPortsEqual(a, b []applicationPort) bool {
	an := normalizeApplicationPorts(a)
	bn := normalizeApplicationPorts(b)
	if len(an) != len(bn) {
		return false
	}
	for i := range an {
		if an[i].User != bn[i].User {
			return false
		}
		if !stringSliceEqual(an[i].TCP, bn[i].TCP) || !stringSliceEqual(an[i].UDP, bn[i].UDP) {
			return false
		}
	}
	return true
}

func normalizeApplicationPorts(in []applicationPort) []applicationPort {
	out := make([]applicationPort, 0, len(in))
	for _, p := range in {
		user := policyUser(p.User)
		if user == "" {
			continue
		}
		tcp := dedupeSorted(p.TCP)
		udp := dedupeSorted(p.UDP)
		if len(tcp) == 0 && len(udp) == 0 {
			continue
		}
		out = append(out, applicationPort{User: user, TCP: tcp, UDP: udp})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].User < out[j].User })
	return out
}

func dedupeSorted(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	s := sets.NewString()
	for _, p := range in {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		s.Insert(p)
	}
	if s.Len() == 0 {
		return []string{}
	}
	return s.List()
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
