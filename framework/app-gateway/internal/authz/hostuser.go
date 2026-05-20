// Package authz contains the pure-function deciders that app-service-ext-authz
// composes into a Check decision. Each decider returns one of three actions
// (Allow, Deny, Pass) plus a stable error code so the call site can log a
// single line per request.
//
// Phase-A v2 introduces a single decider — HostUser — that enforces the
// invariant
//
//	Host[1] == X-BFL-USER
//
// where Host is the lower-cased :authority header (typically of the form
// <hash8>.<viewer>.<platformDomain>). The check is intentionally
// independent of the platform PDP: any mismatch returns INVALID_HOST_USER
// and never escalates to system-server.
//
// References:
//   - archdoc/方案/shared应用/Shared外部访问主流程打通方案-2026-05-20-明确方案.md §6
//   - archdoc/方案/shared应用/Shared外部访问v2评审决议-2026-05-20.md      R-V2-9
//   - archdoc/方案/shared应用/统一认证组件设计与开发方案-2026-05-20-明确方案.md
package authz

import (
	"strings"
)

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
	Code     string // stable machine code (e.g. INVALID_HOST_USER), "" for Allow/Pass
	Message  string // human-friendly reason; safe to put into Denied.body
	Viewer   string // extracted viewer (Host label[1]) for logging
	Username string // value of X-BFL-USER (lower-case) for logging
}

// HostUserConfig controls the HostUser decider.
type HostUserConfig struct {
	// Enabled toggles the entire decider. When false the decider always
	// returns ActionPass (PR-8 §6.3 H10).
	Enabled bool
	// SkipPrefixes lets local operators bypass the check for specific
	// host prefixes. Each entry is matched as a DNS label suffix against
	// Host[1] (the viewer label). Empty by default.
	SkipPrefixes []string
}

// DefaultHostUserConfig returns a config that enables host-user enforcement.
func DefaultHostUserConfig() HostUserConfig {
	return HostUserConfig{Enabled: true}
}

// HostUser returns a Decision for the given request inputs. It never panics.
//
// Inputs:
//
//	authority — :authority pseudo-header (typically Host); already lower-cased
//	            by Envoy; the decider re-lowercases defensively.
//	headers   — map of lower-case headers (Envoy ext_authz contract).
//	cfg       — decider configuration.
//
// Behaviour:
//
//	cfg.Enabled=false                   → Pass
//	host has <3 labels                  → Deny INVALID_HOST_USER
//	header X-BFL-USER missing/empty     → Deny INVALID_HOST_USER
//	host[1] != x-bfl-user (lc compare)  → Deny INVALID_HOST_USER
//	otherwise                           → Allow
//
// Hash label (host[0]) must be 8-char lowercase hex; otherwise Deny
// INVALID_HOST_PREFIX (defence-in-depth: the SRR uniqueness check already
// rejects collisions on the control plane).
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
		return Decision{
			Action:   ActionDeny,
			Code:     "INVALID_HOST_PREFIX",
			Message:  "host prefix is not an 8-char lowercase hex hash",
			Viewer:   labels[1],
			Username: lower(headers["x-bfl-user"]),
		}
	}
	viewer := labels[1]
	user := lower(headers["x-bfl-user"])
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
		// Only strip when the suffix is purely numeric (IPv4 with port).
		// IPv6 literals are not expected on this code path.
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
