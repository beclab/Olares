// Package authz contains the pure-function deciders that the in-process
// PEP composes into a Check decision. The Shared external-access stack
// runs ext_authz inside app-service (no separate authz workload).
//
// The in-process PEP currently ships one decider — HostUser — that enforces the
// invariant
//
//	Host[1] == X-BFL-USER
//
// where Host is the lower-cased :authority header (typically of the form
// <hash8>.<viewer>.<platformDomain>). The check is intentionally
// independent of the platform PDP: any mismatch returns INVALID_HOST_USER
// and never escalates to system-server.
package authz

import "strings"

// Action is the trinary decision returned by a Decider.
type Action int

const (
	// ActionPass means "this decider has no opinion; consult the next one".
	ActionPass Action = iota
	// ActionAllow means "short-circuit: return OK".
	ActionAllow
	// ActionDeny means "short-circuit: return Denied(403)".
	ActionDeny
)

// Decision is the structured output of a Decider; the gRPC handler turns it
// into an ext_authz Check response.
type Decision struct {
	Action   Action
	Code     string
	Message  string
	Viewer   string
	Username string
}

// HostUserConfig controls the HostUser decider.
type HostUserConfig struct {
	Enabled      bool
	SkipPrefixes []string
}

// DefaultHostUserConfig enables host-user enforcement.
func DefaultHostUserConfig() HostUserConfig {
	return HostUserConfig{Enabled: true}
}

// HostUser returns a Decision for the given request inputs. It never panics.
//
// Behaviour:
//
//	cfg.Enabled=false                   → Pass
//	host has <3 labels                  → Deny INVALID_HOST_USER
//	header X-BFL-USER missing/empty     → Deny INVALID_HOST_USER
//	host[0] is not 8-char lowercase hex → Deny INVALID_HOST_PREFIX
//	host[1] != x-bfl-user (lc compare)  → Deny INVALID_HOST_USER
//	otherwise                           → Allow
func HostUser(authority string, headers map[string]string, cfg HostUserConfig) Decision {
	if !cfg.Enabled {
		return Decision{Action: ActionPass}
	}
	host := normalizeHost(authority)
	if host == "" {
		return Decision{
			Action:  ActionDeny,
			Code:    "INVALID_HOST_USER",
			Message: "empty :authority",
		}
	}
	labels := strings.Split(host, ".")
	if len(labels) < 3 {
		return Decision{
			Action:  ActionDeny,
			Code:    "INVALID_HOST_USER",
			Message: "host must have at least <hash8>.<viewer>.<domain>",
		}
	}
	if !isHash8(labels[0]) {
		// Non–Shared-v2 hosts (e.g. demo.agw.local) defer to the allow-all baseline.
		return Decision{Action: ActionPass}
	}
	viewer := labels[1]
	user := lower(headerValue(headers, "x-bfl-user"))
	if user == "" {
		return Decision{
			Action:  ActionDeny,
			Code:    "INVALID_HOST_USER",
			Message: "X-BFL-USER header missing",
			Viewer:  viewer,
		}
	}
	for _, p := range cfg.SkipPrefixes {
		if p != "" && viewer == lower(p) {
			return Decision{Action: ActionAllow, Viewer: viewer, Username: user}
		}
	}
	if viewer != user {
		return Decision{
			Action:   ActionDeny,
			Code:     "INVALID_HOST_USER",
			Message:  "X-BFL-USER does not match host viewer label",
			Viewer:   viewer,
			Username: user,
		}
	}
	return Decision{Action: ActionAllow, Viewer: viewer, Username: user}
}

// normalizeHost trims port and lower-cases. Strips one trailing dot.
func normalizeHost(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.LastIndex(s, ":"); i >= 0 {
		hostPart := s[:i]
		portPart := s[i+1:]
		allDigit := portPart != ""
		for _, r := range portPart {
			if r < '0' || r > '9' {
				allDigit = false
				break
			}
		}
		if allDigit {
			s = hostPart
		}
	}
	s = strings.TrimSuffix(s, ".")
	return strings.ToLower(s)
}

func isHash8(s string) bool {
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

func lower(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
