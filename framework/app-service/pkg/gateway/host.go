// Package gateway contains helpers app-service uses to project Application
// state into SharedRouteRegistry (gateway.olares.io/v1alpha1) objects.
package gateway

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidHostPattern is returned when a host string cannot be normalized
// into a Phase A hostPattern (lowercase, no port, no scheme — F-3).
var ErrInvalidHostPattern = errors.New("invalid host pattern")

// NormalizeHostPattern returns a Phase-A-compliant hostPattern derived from
// the (possibly noisy) string the caller produced via GenSharedEntranceURL:
//
//   - whitespace is trimmed
//   - a single trailing ":<port>" is removed
//   - the host is lower-cased
//   - URLs (anything containing "://") are rejected
//   - empty strings, paths, query strings, port-only inputs are rejected
//
// The output matches the openAPIV3 pattern declared on the CRD
// (^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$). Wildcards are not accepted in Phase A.
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
// The first invalid pattern aborts and is returned as the error.
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

// isValidHostLabel checks the regex used by the CRD schema:
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
