package util

import (
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

func SysctlSomaxconn() int {
	maxConns, err := getSysctl("net/core/somaxconn")
	if err != nil || maxConns < 512 {
		klog.V(3).InfoS("Using default net.core.somaxconn", "value", maxConns)
		return 511
	}

	return maxConns
}

// getSysctl returns the value for the specified sysctl setting
func getSysctl(sysctl string) (int, error) {
	data, err := ioutil.ReadFile(path.Join("/proc/sys", sysctl))
	if err != nil {
		return -1, err
	}

	val, err := strconv.Atoi(strings.Trim(string(data), " \n"))
	if err != nil {
		return -1, err
	}

	return val, nil
}
