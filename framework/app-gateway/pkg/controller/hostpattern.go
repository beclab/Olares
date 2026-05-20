// Package controller — host pattern translation helpers introduced in PR-7.
//
// The SharedRouteRegistry CRD allows two hostPattern forms:
//
//	exact:    abc.shared.example.com
//	logical:  <hash8>.*.<platformDomain>
//
// app-service-routecontrol materialises every SRR into a Gateway-API HTTPRoute. For
// exact hosts the host is copied verbatim into `spec.hostnames` (Phase-A
// behaviour, kept as a fallback). For logical patterns the hostnames list
// uses `*.<platformDomain>` (which Envoy Gateway implements as a single-label
// wildcard) and the rule grows a `headers` match of type RegularExpression on
// the `Host` header that pins the hash8 prefix.
//
// References:
//   - archdoc/方案/shared应用/Shared外部访问主流程打通方案-2026-05-20-明确方案.md §5
//   - archdoc/方案/shared应用/Shared外部访问v2评审决议-2026-05-20.md (R-V2-2, R-V2-10, §4 POC)
package controller

import (
	"fmt"
	"regexp"
	"strings"
)

// LogicalPattern represents a parsed <hash8>.*.<platformDomain>.
type LogicalPattern struct {
	Hash8          string // lower hex, exactly 8 chars
	PlatformDomain string // lower-case DNS, no leading dot
}

// IsLogicalPattern reports whether s looks like <hash8>.*.<domain>.
func IsLogicalPattern(s string) bool {
	labels := strings.Split(s, ".")
	return len(labels) >= 3 && labels[1] == "*"
}

// ParseLogicalPattern returns the parsed parts of a logical pattern. The
// second return value is false when s is not a logical pattern; in that case
// callers should fall back to exact-host materialisation.
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

// HostnamesForPlatformDomain returns the wildcard hostnames list used by the
// HTTPRoute when at least one logical pattern targets a given platformDomain.
// Envoy Gateway interprets "*.example.com" as a single-label match — the host
// regex match then constrains it to the hash8 prefix.
func HostnamesForPlatformDomain(dom string) []string {
	return []string{"*." + strings.ToLower(dom)}
}

// HostRegexValue returns a Gateway-API headers.value RegularExpression that
// pins (hash8, platformDomain) while allowing any RFC-1123 single-label
// viewer between them. The regex is anchored.
//
// Example for hash8=01234567, dom=olares.com:
//
//	^01234567\.[a-z0-9]([-a-z0-9]*[a-z0-9])?\.olares\.com$
func HostRegexValue(p LogicalPattern) string {
	// Envoy / RE2 syntax. Single-label viewer per RFC-1123 (lowercase + digits + hyphen).
	return fmt.Sprintf(`^%s\.[a-z0-9]([-a-z0-9]*[a-z0-9])?\.%s$`,
		regexp.QuoteMeta(p.Hash8), regexp.QuoteMeta(p.PlatformDomain))
}

// HostHeaderMatch returns the Gateway-API headers element that pins the
// Host header to the logical pattern. Returns (nil, false) when the pattern
// is exact (no host header match needed).
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

// MaterializeHostnames takes the SRR hostPatterns and returns the
// `spec.hostnames` slice for the materialised HTTPRoute. Logical patterns
// contribute their wildcard form ("*.domain"); exact hosts contribute the
// host verbatim. The slice is de-duplicated and ordered for determinism.
func MaterializeHostnames(patterns []string) []any {
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

// MaterializeHostHeaders returns the Host RegularExpression matches that the
// HTTPRoute rule must apply, one per logical pattern. Exact hosts contribute
// nothing here because spec.hostnames already pins them.
func MaterializeHostHeaders(patterns []string) []map[string]any {
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
