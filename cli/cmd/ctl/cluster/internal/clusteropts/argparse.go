package clusteropts

import (
	"fmt"
	"strings"
)

// SplitNsName resolves the canonical "<namespace>/<name>" or bare-name
// argument grammar used across the cluster tree's get/yaml/pods/etc.
// verbs. Matches the SPA's resource picker: either pass an explicit
// --namespace and a bare name, or embed the namespace in the argument
// as <ns>/<name>; conflicts between the two are rejected.
//
// Returns (namespace, name, error). Errors are user-actionable and
// already include the offending argument verbatim.
func SplitNsName(nsFlag, arg string) (string, string, error) {
	if strings.Contains(arg, "/") {
		parts := strings.SplitN(arg, "/", 2)
		if parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("invalid <namespace>/<name>: %q", arg)
		}
		if nsFlag != "" && nsFlag != parts[0] {
			return "", "", fmt.Errorf("argument namespace %q conflicts with --namespace %q", parts[0], nsFlag)
		}
		return parts[0], parts[1], nil
	}
	if nsFlag == "" {
		return "", "", fmt.Errorf("namespace required: pass --namespace or use <namespace>/<name>")
	}
	return nsFlag, arg, nil
}
