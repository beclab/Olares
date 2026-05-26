package authz

import (
	"errors"
	"strings"
)

const (
	headerL5dClientID = "l5d-client-id"
)

var (
	errEmptyL5d       = errors.New("empty l5d-client-id")
	errMalformedL5d   = errors.New("malformed l5d-client-id")
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
