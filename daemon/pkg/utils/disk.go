package utils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	syscall "golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

func GetDiskSize() (uint64, error) {
	return GetDiskTotalBytesForPath("/")
}

func GetDiskAvailableSpace(path string) (uint64, error) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		klog.Error("get disk available space error, ", err)
		return 0, err
	}

	available := fs.Bavail * uint64(fs.Bsize)
	return available, nil
}

// Find the mount device for a given path (from /proc/mounts), choose the longest matching mount point
func deviceForPath(path string) (string, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return "", err
	}
	defer f.Close()

	var bestDevice, bestMount string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// /proc/mounts: device mountpoint fs ...  (space-separated, mountpoint may have \040 etc. escapes)
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		device := fields[0]
		mount := fields[1]
		// Handle escaped spaces (simple processing)
		mount = strings.ReplaceAll(mount, "\\040", " ")
		// Choose the longest matching mount prefix (to prevent nested mounts)
		if strings.HasPrefix(path, mount) {
			if len(mount) > len(bestMount) {
				bestMount = mount
				bestDevice = device
			}
		}
	}
	if bestDevice == "" {
		return "", fmt.Errorf("no device found for path %s", path)
	}
	return bestDevice, nil
}

// Given a device path (e.g. /dev/sda1), find the top-level block device name (e.g. sda)
func topBlockDeviceName(devPath string) (string, error) {
	name := filepath.Base(devPath) // e.g. sda1, nvme0n1p1, dm-0, mapper/xxx -> basename
	sysPath := filepath.Join("/sys/class/block", name)
	real, err := filepath.EvalSymlinks(sysPath)
	if err != nil {
		// Sometimes device paths may not be /dev/* (e.g. UUID paths), return error when lookup fails with basename directly
		return "", err
	}
	// real might be .../block/sda/sda1, taking parent directory name gives us the top-level device sda
	parent := filepath.Base(filepath.Dir(real))
	// If parent equals name (no parent), then parent is itself
	if parent == "" {
		parent = name
	}
	return parent, nil
}

// Read /sys/class/block/<dev>/size (in sectors), multiply by 512 to get bytes
func diskSizeBySysfs(topDev string) (uint64, error) {
	sizePath := filepath.Join("/sys/class/block", topDev, "size")
	b, err := ioutil.ReadFile(sizePath)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(b))
	sectors, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	const sectorSize = 512
	return sectors * sectorSize, nil
}

// Comprehensive: given a path (mount point or path), return the total bytes of the associated physical device
func GetDiskTotalBytesForPath(path string) (uint64, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return 0, err
	}
	device, err := deviceForPath(abs)
	if err != nil {
		return 0, err
	}
	topDev, err := topBlockDeviceName(device)
	if err != nil {
		return 0, err
	}
	size, err := diskSizeBySysfs(topDev)
	if err != nil {
		return 0, err
	}
	return size, nil
}
