package security

import (
	"strings"

	"github.com/thoas/go-funk"
)

var (
	// UnderLayerNamespaces indicates the namespaces that created by kubernetes and kubesphere.
	UnderLayerNamespaces = []string{
		"default",
		"kube-node-lease",
		"kube-public",
		"kube-system",
		"kubekey-system",
		"kubesphere-controls-system",
		"kubesphere-monitoring-federated",
		"kubesphere-monitoring-system",
		"kubesphere-system",
	}

	// OSSystemNamespaces indicates the namespaces that used by terminus of system applications.
	OSSystemNamespaces = []string{
		"os-framework",
		"os-platform",
	}
	// GPUSystemNamespaces indicates the namespaces that for gpu system.
	GPUSystemNamespaces = []string{
		"os-gpu",
	}

	OSNetworkNamespaces = []string{
		"os-network",
	}

	OSProtectedNamespaces = []string{
		"os-protected",
	}
)

// IsUserInternalNamespaces returns true if namespace is user level namespace for a user application, false otherwise.
func IsUserInternalNamespaces(ns string) (bool, string) {
	for _, prefix := range []string{"user-space-", "user-system-"} {
		if strings.HasPrefix(ns, prefix) {
			strToken := strings.Split(ns, "-")
			if len(strToken) >= 3 {
				return true, ns[len(prefix):]
			}
		}
	}

	return false, ""
}

// IsUserSpaceNamespaces return true if namespace is user space namespace, false otherwise.
func IsUserSpaceNamespaces(ns string) bool {
	if !strings.HasPrefix(ns, "user-space-") {
		return false
	}

	strToken := strings.Split(ns, "-")
	return len(strToken) >= 3
}

// IsUserSystemNamespaces return true if namespace is user system namespace, false otherwise.
func IsUserSystemNamespaces(ns string) bool {
	if !strings.HasPrefix(ns, "user-system-") {
		return false
	}

	strToken := strings.Split(ns, "-")
	return len(strToken) >= 3
}

// IsUnderLayerNamespace returns true if ns is contained by UnderLayerNamespaces, false otherwise.
func IsUnderLayerNamespace(ns string) bool {
	return funk.Contains(UnderLayerNamespaces, ns)
}

// IsOSSystemNamespace returns true if ns is contained by OSSystemNamespaces, false otherwise.
func IsOSSystemNamespace(ns string) bool {
	return funk.Contains(OSSystemNamespaces, ns)
}

func IsOSNetworkNamespace(ns string) bool {
	return funk.Contains(OSNetworkNamespaces, ns)
}

func IsOSGpuNamespace(ns string) bool {
	return funk.Contains(GPUSystemNamespaces, ns)
}

func IsPublicNamespace(ns string) bool {
	return funk.Contains(GPUSystemNamespaces, ns)
}

func IsOSProtectedNamespace(ns string) bool {
	return funk.Contains(OSProtectedNamespaces, ns)
}
