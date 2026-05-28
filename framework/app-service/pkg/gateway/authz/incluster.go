package authz

import (
	"errors"
	"strings"
)

const (
	headerL5dClientID = "l5d-client-id"

	nsOwnerLabel = "bytetrade.io/ns-owner"
)

const (
	SourceNone             = "none"
	SourceNsLabel          = "ns_label"
	SourcePrefixUserSpace  = "prefix_user_space"
	SourcePrefixUserSystem = "prefix_user_system"
	SourceAppUserFallback  = "app_user_fallback"
)

var (
	errEmptyL5d     = errors.New("empty l5d-client-id")
	errMalformedL5d = errors.New("malformed l5d-client-id")
)

// ParseL5dClientID parses Linkerd source identity:
// "<sa>.<ns>.serviceaccount.identity.linkerd.cluster.local".
func ParseL5dClientID(header string) (sa, ns string, err error) {
	v := strings.TrimSpace(header)
	if v == "" {
		return "", "", errEmptyL5d
	}
	const suffix = ".serviceaccount.identity.linkerd.cluster.local"
	if !strings.HasSuffix(v, suffix) {
		return "", "", errMalformedL5d
	}
	core := strings.TrimSuffix(v, suffix)
	parts := strings.Split(core, ".")
	if len(parts) < 2 {
		return "", "", errMalformedL5d
	}
	ns = parts[len(parts)-1]
	sa = strings.Join(parts[:len(parts)-1], ".")
	if sa == "" || ns == "" {
		return "", "", errMalformedL5d
	}
	return sa, ns, nil
}

// DeriveViewer maps a caller workload namespace to an Olares viewer name.
func DeriveViewer(callerNS string) (viewer string, ok bool) {
	ns := strings.ToLower(strings.TrimSpace(callerNS))
	switch {
	case strings.HasPrefix(ns, "user-space-"):
		return strings.TrimPrefix(ns, "user-space-"), true
	case strings.HasPrefix(ns, "user-system-"):
		return strings.TrimPrefix(ns, "user-system-"), true
	default:
		return "", false
	}
}

// DeriveViewerWithMeta resolves the Olares viewer from a caller namespace,
// optional caller-namespace labels and an optional known-users set.
//
// requirement: WI-27 closes the LiteLLM G-B gap by accepting <app>-<user>
// caller namespaces (e.g. litellm-brucedai) when <user> is a known Olares
// user; the legacy DeriveViewer only handled user-space-/user-system- prefixes.
// behavior: priority paths (first match wins) —
//
//	1) nsLabels["bytetrade.io/ns-owner"] non-empty -> ns_label
//	2) HasPrefix "user-space-"                      -> prefix_user_space
//	3) HasPrefix "user-system-"                     -> prefix_user_system
//	4) <app>-<user> where <user> in knownUsers      -> app_user_fallback
//
// Returns ("", "none", false) when no path matches. Pure function.
// test: TC-031~036 in incluster_test.go.
func DeriveViewerWithMeta(callerNS string, nsLabels map[string]string, knownUsers map[string]struct{}) (viewer string, source string, ok bool) {
	if owner := strings.ToLower(strings.TrimSpace(nsLabels[nsOwnerLabel])); owner != "" {
		return owner, SourceNsLabel, true
	}
	ns := strings.ToLower(strings.TrimSpace(callerNS))
	if ns == "" {
		return "", SourceNone, false
	}
	switch {
	case strings.HasPrefix(ns, "user-space-"):
		return strings.TrimPrefix(ns, "user-space-"), SourcePrefixUserSpace, true
	case strings.HasPrefix(ns, "user-system-"):
		return strings.TrimPrefix(ns, "user-system-"), SourcePrefixUserSystem, true
	}
	if i := strings.LastIndex(ns, "-"); i > 0 && i < len(ns)-1 {
		candidate := ns[i+1:]
		if _, hit := knownUsers[candidate]; hit {
			return candidate, SourceAppUserFallback, true
		}
	}
	return "", SourceNone, false
}

// IsSharedInclusterHost reports whether authority looks like a v2 Shared entrance host.
func IsSharedInclusterHost(authority string) bool {
	host := normalizeHost(authority)
	if host == "" {
		return false
	}
	labels := strings.Split(host, ".")
	if len(labels) < 3 {
		return false
	}
	return isHash8(labels[0])
}

// HostViewerLabel returns the viewer segment from a Shared in-cluster host.
func HostViewerLabel(authority string) string {
	host := normalizeHost(authority)
	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return ""
	}
	return labels[1]
}
