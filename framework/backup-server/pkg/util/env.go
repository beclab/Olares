package util

import (
	"os"
	"strings"
	// "k8s.io/klog"
	// "k8s.io/utils/ptr"
)

func EnvOrDefault(name, def string) string {
	v := os.Getenv(name)

	if v == "" && def != "" {
		return def
	}
	v = strings.TrimRight(v, "/")
	return v
}
