// Package gateway contains helpers app-service uses to project Application
// state into SharedRouteRegistry (gateway.olares.io/v1alpha1) objects.
package gateway

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidHostPattern is returned when a host string cannot be normalized
// into a strict hostPattern (lowercase, no port, no scheme).
var ErrInvalidHostPattern = errors.New("invalid host pattern")

// NormalizeHostPattern returns a strict hostPattern: whitespace trimmed, a
// single trailing ":<port>" removed, lower-cased; URLs, paths and wildcards
// are rejected. The output matches the CRD pattern
// (^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$).
func NormalizeHostPattern(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("%w: empty host", ErrInvalidHostPattern)
	}
	if strings.Contains(s, "://") {
		return "", fmt.Errorf("%w: scheme not allowed (%q)", ErrInvalidHostPattern, raw)
	}
	if strings.ContainsAny(s, "/?#") {
		return "", fmt.Errorf("%w: path/query not allowed (%q)", ErrInvalidHostPattern, raw)
	}

	if idx := strings.LastIndex(s, ":"); idx >= 0 {
		host := s[:idx]
		port := s[idx+1:]
		if host == "" || port == "" {
			return "", fmt.Errorf("%w: malformed host:port (%q)", ErrInvalidHostPattern, raw)
		}
		if !isAllDigits(port) {
			return "", fmt.Errorf("%w: port must be numeric (%q)", ErrInvalidHostPattern, raw)
		}
		s = host
	}

	s = strings.ToLower(s)

	if !isValidHostLabel(s) {
		return "", fmt.Errorf("%w: %q does not match host pattern", ErrInvalidHostPattern, raw)
	}
	return s, nil
}

// NormalizeHostPatterns runs NormalizeHostPattern over a slice and de-duplicates
// while preserving first-seen order. An empty input slice returns nil, nil.
func NormalizeHostPatterns(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		h, err := NormalizeHostPattern(r)
		if err != nil {
			return nil, err
		}
		if _, dup := seen[h]; dup {
			continue
		}
		seen[h] = struct{}{}
		out = append(out, h)
	}
	return out, nil
}

// NormalizeHostOrLogicalPattern accepts exact hosts (via NormalizeHostPattern)
// and per-viewer logical patterns <hash8>.*.<platformDomain>. The wildcard
// label MUST appear as a single literal "*" exactly once and MUST be the
// second label from the left. The output is the lowercase canonical form.
func NormalizeHostOrLogicalPattern(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("%w: empty host", ErrInvalidHostPattern)
	}
	if strings.Contains(s, "://") {
		return "", fmt.Errorf("%w: scheme not allowed (%q)", ErrInvalidHostPattern, raw)
	}
	if strings.ContainsAny(s, "/?#") {
		return "", fmt.Errorf("%w: path/query not allowed (%q)", ErrInvalidHostPattern, raw)
	}
	if idx := strings.LastIndex(s, ":"); idx >= 0 {
		host := s[:idx]
		port := s[idx+1:]
		if host == "" || port == "" {
			return "", fmt.Errorf("%w: malformed host:port (%q)", ErrInvalidHostPattern, raw)
		}
		if !isAllDigits(port) {
			return "", fmt.Errorf("%w: port must be numeric (%q)", ErrInvalidHostPattern, raw)
		}
		s = host
	}
	s = strings.ToLower(s)

	if !strings.Contains(s, "*") {
		if !isValidHostLabel(s) {
			return "", fmt.Errorf("%w: %q does not match host pattern", ErrInvalidHostPattern, raw)
		}
		return s, nil
	}

	labels := strings.Split(s, ".")
	if len(labels) < 3 {
		return "", fmt.Errorf("%w: logical pattern needs <hash8>.*.<domain> (got %q)", ErrInvalidHostPattern, raw)
	}
	if labels[1] != "*" {
		return "", fmt.Errorf("%w: wildcard must be the 2nd label (got %q)", ErrInvalidHostPattern, raw)
	}
	for i, lab := range labels {
		if i == 1 {
			continue
		}
		if lab == "" || strings.Contains(lab, "*") {
			return "", fmt.Errorf("%w: only one literal '*' label allowed (got %q)", ErrInvalidHostPattern, raw)
		}
		if !isValidDNSLabel(lab) {
			return "", fmt.Errorf("%w: label %q invalid in %q", ErrInvalidHostPattern, lab, raw)
		}
	}
	return s, nil
}

// IsLogicalHostPattern reports whether s is a logical hostPattern
// (<hash8>.*.<domain>) rather than an exact hostname.
func IsLogicalHostPattern(s string) bool {
	labels := strings.Split(s, ".")
	return len(labels) >= 3 && labels[1] == "*"
}

func isValidDNSLabel(s string) bool {
	if s == "" || len(s) > 63 {
		return false
	}
	if !isAlphaNum(rune(s[0])) {
		return false
	}
	if len(s) > 1 && !isAlphaNum(rune(s[len(s)-1])) {
		return false
	}
	for _, r := range s {
		if !(isAlphaNum(r) || r == '-') {
			return false
		}
	}
	return true
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isValidHostLabel checks the CRD schema regex
// ^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$. The host must already be lowercase.
func isValidHostLabel(s string) bool {
	if s == "" {
		return false
	}
	if !isAlphaNum(rune(s[0])) {
		return false
	}
	if len(s) > 1 && !isAlphaNum(rune(s[len(s)-1])) {
		return false
	}
	for _, r := range s {
		if !(isAlphaNum(r) || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}
