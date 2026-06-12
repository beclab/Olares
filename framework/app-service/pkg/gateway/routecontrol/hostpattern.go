// Package routecontrol implements app-service shared ingress route control:
// for each SharedRouteRegistry (SRR) it maintains the Gateway API HTTPRoute
// and the companion NetworkPolicy that allow app-gateway to reach the
// upstream Service (in the SRR namespace or upstream.serviceNamespace).
// No service mesh.
package routecontrol

import (
	"fmt"
	"regexp"
	"strings"
)

// LogicalPattern is a parsed <hash8>.*.<platformDomain> host pattern.
type LogicalPattern struct {
	Hash8          string
	PlatformDomain string
}

// IsLogicalPattern reports whether s has the <hash8>.*.<domain> shape.
func IsLogicalPattern(s string) bool {
	labels := strings.Split(s, ".")
	return len(labels) >= 3 && labels[1] == "*"
}

// ParseLogicalPattern returns the parsed parts of a logical pattern. The
// second return value is false when s is not a logical pattern.
func ParseLogicalPattern(s string) (LogicalPattern, bool) {
	if !IsLogicalPattern(s) {
		return LogicalPattern{}, false
	}
	labels := strings.Split(s, ".")
	hash := strings.ToLower(strings.TrimSpace(labels[0]))
	if !validHash8(hash) {
		return LogicalPattern{}, false
	}
	dom := strings.ToLower(strings.TrimSpace(strings.Join(labels[2:], ".")))
	if dom == "" || strings.Contains(dom, "*") {
		return LogicalPattern{}, false
	}
	if !validPlatformDomain(dom) {
		return LogicalPattern{}, false
	}
	return LogicalPattern{Hash8: hash, PlatformDomain: dom}, true
}

func validHash8(s string) bool {
	if len(s) != 8 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		default:
			return false
		}
	}
	return true
}

func validPlatformDomain(s string) bool {
	if s == "" || len(s) > 253 {
		return false
	}
	if s[0] == '-' || s[0] == '.' || s[len(s)-1] == '-' || s[len(s)-1] == '.' {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '.':
		default:
			return false
		}
	}
	return true
}

// HostRegexValue returns an anchored RE2 regex for (hash8, platformDomain)
// with any RFC-1123 single-label viewer between them.
func HostRegexValue(p LogicalPattern) string {
	return fmt.Sprintf(`^%s\.[a-z0-9]([-a-z0-9]*[a-z0-9])?\.%s$`,
		regexp.QuoteMeta(p.Hash8), regexp.QuoteMeta(p.PlatformDomain))
}

// HostHeaderMatch returns the Gateway API Host header match for a logical
// pattern. Exact hosts do not need a header match.
func HostHeaderMatch(pattern string) (map[string]any, bool) {
	p, ok := ParseLogicalPattern(pattern)
	if !ok {
		return nil, false
	}
	return map[string]any{
		"name":  "Host",
		"type":  "RegularExpression",
		"value": HostRegexValue(p),
	}, true
}

// HTTPRouteHostnames builds spec.hostnames for an HTTPRoute from SRR
// hostPatterns. Logical patterns become *.<platformDomain>; exact hosts are
// copied verbatim. The result is deduplicated and order-stable.
func HTTPRouteHostnames(patterns []string) []any {
	seen := make(map[string]struct{}, len(patterns))
	out := make([]any, 0, len(patterns))
	for _, p := range patterns {
		var key string
		if lp, ok := ParseLogicalPattern(p); ok {
			key = "*." + lp.PlatformDomain
		} else {
			key = strings.ToLower(strings.TrimSpace(p))
		}
		if key == "" {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

// HTTPRouteHostHeaderMatches builds Host RegularExpression header matches,
// one per logical pattern. Exact hosts contribute nothing.
func HTTPRouteHostHeaderMatches(patterns []string) []map[string]any {
	out := make([]map[string]any, 0, len(patterns))
	seen := make(map[string]struct{}, len(patterns))
	for _, p := range patterns {
		m, ok := HostHeaderMatch(p)
		if !ok {
			continue
		}
		key, _ := m["value"].(string)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, m)
	}
	return out
}
