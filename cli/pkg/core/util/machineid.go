package util

import (
	"os"
	"strings"
)

// machineIDFiles and systemUUIDFiles mirror the sources that the kubelet
// (through cAdvisor) uses to populate node.status.nodeInfo.machineID and
// node.status.nodeInfo.systemUUID respectively, so that values read here can be
// compared against the ones reported on a Kubernetes Node object.
//
// machineID:  /etc/machine-id, falling back to /var/lib/dbus/machine-id
// systemUUID: DMI product_uuid, PowerPC device-tree UUIDs, then the s390
// /etc/machine-id fallback
var (
	machineIDFiles  = []string{"/etc/machine-id", "/var/lib/dbus/machine-id"}
	systemUUIDFiles = []string{
		"/sys/class/dmi/id/product_uuid",
		"/proc/device-tree/system-id",
		"/proc/device-tree/vm,uuid",
		"/etc/machine-id",
	}
)

// GetMachineID returns the host machine-id, or an empty string if none of the
// known files exist or are readable. It does not error out: an empty result is
// a valid "unknown" answer, matching cAdvisor's behavior.
func GetMachineID() string {
	return firstNonEmptyFileContent(machineIDFiles)
}

// GetSystemUUID follows cAdvisor's Linux lookup order and returns an empty
// string when none of the platform-specific identifiers can be read.
func GetSystemUUID() string {
	return firstNonEmptyFileContent(systemUUIDFiles)
}

func firstNonEmptyFileContent(paths []string) string {
	for _, p := range paths {
		content, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		id := strings.TrimSpace(string(content))
		id = strings.TrimSpace(strings.TrimRight(id, "\x00"))
		if id != "" {
			return id
		}
	}
	return ""
}
