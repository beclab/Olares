package access

import (
	"fmt"
	"strings"
)

const (
	RouteModeDirect        = "Direct"
	RouteModeViaAppGateway = "ViaAppGateway"
	DesiredActive          = "Active"
	DesiredRevoked         = "Revoked"
)

// PolicySpec is the compiler input projected from AppAccessPolicy.
type PolicySpec struct {
	Name              string
	Namespace         string // target application namespace
	DesiredState      string
	RouteMode         string
	ServerPort        int32
	SourceSANamespace string
	SourceSAName      string
	GatewaySAName     string // required when ViaAppGateway
	Rules             []HTTPRule
}

// CompiledObjects are deterministic names for Linkerd/Gateway/NP projections.
type CompiledObjects struct {
	ServerName               string
	HTTPRouteName            string
	AuthorizationPolicyName  string
	MeshTLSAuthName          string
	NetworkPolicyName        string
	IdentityPrincipal        string
	SkipCompile              bool
	Reason                   string
}

// CompilePolicy maps an access policy into managed object names and principals.
// It does not talk to the API server; controllers apply the returned plan.
func CompilePolicy(p PolicySpec, compilationEnabled bool) (CompiledObjects, error) {
	out := CompiledObjects{}
	if !compilationEnabled {
		out.SkipCompile = true
		out.Reason = "policyCompilationEnabled=false"
		return out, nil
	}
	if p.Name == "" || p.Namespace == "" {
		return out, fmt.Errorf("policy name and namespace are required")
	}
	if p.DesiredState == DesiredRevoked {
		out.SkipCompile = true
		out.Reason = "desiredState=Revoked"
		return out, nil
	}
	if p.DesiredState != "" && p.DesiredState != DesiredActive {
		return out, fmt.Errorf("unsupported desiredState %q", p.DesiredState)
	}
	mode := p.RouteMode
	if mode == "" {
		mode = RouteModeDirect
	}
	if mode != RouteModeDirect && mode != RouteModeViaAppGateway {
		return out, fmt.Errorf("unsupported routeMode %q", mode)
	}
	if err := ValidateHTTPRules(p.Rules); err != nil {
		return out, err
	}
	if p.ServerPort <= 0 {
		return out, fmt.Errorf("serverPort must be positive")
	}
	base := sanitizeDNSLabel(p.Name)
	out.ServerName = base + "-server"
	out.HTTPRouteName = base + "-route"
	out.AuthorizationPolicyName = base + "-authz"
	out.MeshTLSAuthName = base + "-mtls"
	out.NetworkPolicyName = base + "-np"
	switch mode {
	case RouteModeDirect:
		if p.SourceSANamespace == "" || p.SourceSAName == "" {
			return out, fmt.Errorf("Direct mode requires source ServiceAccount")
		}
		out.IdentityPrincipal = fmt.Sprintf("%s.%s.serviceaccount.identity.linkerd.cluster.local", p.SourceSAName, p.SourceSANamespace)
	case RouteModeViaAppGateway:
		if p.GatewaySAName == "" {
			return out, fmt.Errorf("ViaAppGateway requires gateway ServiceAccount")
		}
		// Target must authorize the gateway data-plane identity only (no Direct bypass).
		out.IdentityPrincipal = fmt.Sprintf("%s.os-gateway.serviceaccount.identity.linkerd.cluster.local", p.GatewaySAName)
	}
	// Estimate managed objects: Server + HTTPRoute + Authz + MeshTLSAuth + NP.
	if err := ValidateManagedObjectBudget(5); err != nil {
		return out, err
	}
	return out, nil
}

func sanitizeDNSLabel(in string) string {
	in = strings.ToLower(in)
	var b strings.Builder
	for _, r := range in {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "policy"
	}
	if len(out) > 50 {
		out = out[:50]
	}
	return out
}
