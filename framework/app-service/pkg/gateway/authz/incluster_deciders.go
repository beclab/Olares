package authz

import "strings"

// InClusterIdentity validates Linkerd source identity for Shared in-cluster hosts.
//
// requirement: gateway-path in-cluster requests use l5d-client-id; missing identity
// falls through to HostUser (weak path). Malformed identity is denied.
// behavior: Pass for non-Shared hosts; Pass when l5d absent; Deny on parse error;
// Deny when derived viewer disagrees with host viewer label; Allow with Viewer set.
func InClusterIdentity(authority string, headers map[string]string) Decision {
	if !IsSharedInclusterHost(authority) {
		return Decision{Action: ActionPass}
	}
	l5d := headerValue(headers, headerL5dClientID)
	if l5d == "" {
		return Decision{Action: ActionPass}
	}
	_, callerNS, err := ParseL5dClientID(l5d)
	if err != nil {
		return Decision{
			Action:  ActionDeny,
			Code:    CodeInvalidCallerIdentity,
			Message: err.Error(),
		}
	}
	derived, ok := DeriveViewer(callerNS)
	if !ok {
		return Decision{
			Action:  ActionDeny,
			Code:    CodeInvalidCallerIdentity,
			Message: "caller namespace not mapped to a viewer",
		}
	}
	hostViewer := HostViewerLabel(authority)
	if hostViewer != "" && hostViewer != derived {
		return Decision{
			Action:   ActionDeny,
			Code:     CodeInvalidHostUser,
			Message:  "host viewer does not match linkerd identity",
			Viewer:   hostViewer,
			Username: derived,
		}
	}
	return Decision{Action: ActionAllow, Viewer: derived, Username: derived}
}

// InClusterSharedAllow implements Phase A default-allow for Shared entrance hosts.
func InClusterSharedAllow(authority string) Decision {
	if IsSharedInclusterHost(authority) {
		return Decision{Action: ActionAllow}
	}
	return Decision{Action: ActionPass}
}

func headerValue(headers map[string]string, key string) string {
	if headers == nil {
		return ""
	}
	if v, ok := headers[key]; ok && v != "" {
		return v
	}
	return headers[strings.ToLower(key)]
}

// HeadersWithDerivedUser injects X-BFL-USER from the identity decider when absent.
func HeadersWithDerivedUser(headers map[string]string, id Decision) map[string]string {
	if id.Viewer == "" {
		return headers
	}
	if headerValue(headers, "x-bfl-user") != "" {
		return headers
	}
	out := make(map[string]string, len(headers)+1)
	for k, v := range headers {
		out[k] = v
	}
	out["x-bfl-user"] = id.Viewer
	return out
}
