package utils

import (
	syscall "golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

func GetDiskSize() (uint64, error) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs("/", &fs)
	if err != nil {
		klog.Error("get disk space size error, ", err)
		return 0, err
	}

	size := fs.Blocks * uint64(fs.Bsize)
	return size, nil
}
