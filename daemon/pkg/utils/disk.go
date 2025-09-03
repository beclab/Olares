package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	syscall "golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

func GetDiskSize() (string, error) {
	// Get the actual disk size (not just partition size)
	size, diskType, err := getDiskSizeAndType()
	if err != nil {
		klog.Error("get disk size and type error, ", err)
		return "", err
	}

	// Format size in human readable format
	formattedSize := formatBytes(size)

	return fmt.Sprintf("%s %s", formattedSize, diskType), nil
}

// formatBytes converts bytes to human readable format (TB, GB, MB, etc.)
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	result := float64(bytes) / float64(div)
	if result >= 100 {
		return fmt.Sprintf("%.0f%s", result, units[exp])
	} else if result >= 10 {
		return fmt.Sprintf("%.1f%s", result, units[exp])
	}
	return fmt.Sprintf("%.2f%s", result, units[exp])
}

// getDiskSizeAndType gets the total disk size and type by reading from /sys/block
func getDiskSizeAndType() (uint64, string, error) {
	// Find the root device first
	rootDevice, err := getRootBlockDevice()
	if err != nil {
		return 0, "", err
	}

	if rootDevice == "" {
		return 0, "", fmt.Errorf("could not find root block device")
	}

	// Read disk size from /sys/block/{device}/size
	sizePath := fmt.Sprintf("/sys/block/%s/size", rootDevice)
	sizeData, err := os.ReadFile(sizePath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read disk size from %s: %v", sizePath, err)
	}

	// Size is in 512-byte sectors
	sectors, err := strconv.ParseUint(strings.TrimSpace(string(sizeData)), 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse disk size: %v", err)
	}

	// Convert sectors to bytes (512 bytes per sector)
	totalSize := sectors * 512

	// Get disk type
	diskType := getDiskTypeByDevice(rootDevice)

	return totalSize, diskType, nil
}

// getRootBlockDevice finds the block device that contains the root filesystem
func getRootBlockDevice() (string, error) {
	// Read /proc/mounts to find which device contains the root filesystem
	mountsData, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(mountsData), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "/" {
			// Found root mount point
			devicePath := fields[0]

			// Extract device name from device path
			if strings.HasPrefix(devicePath, "/dev/") {
				mountedDevice := strings.TrimPrefix(devicePath, "/dev/")

				// Handle partition numbers to get the base device
				// (e.g., sda1 -> sda, nvme0n1p1 -> nvme0n1)
				baseDevice := getBaseDeviceName(mountedDevice)
				return baseDevice, nil
			}
		}
	}

	return "", fmt.Errorf("root filesystem not found in /proc/mounts")
}

// getBaseDeviceName extracts the base device name from a partition name
func getBaseDeviceName(deviceName string) string {
	if strings.HasPrefix(deviceName, "nvme") {
		// For NVMe devices, remove partition suffix (p1, p2, etc.)
		if idx := strings.LastIndex(deviceName, "p"); idx != -1 {
			if _, err := strconv.Atoi(deviceName[idx+1:]); err == nil {
				return deviceName[:idx]
			}
		}
	} else {
		// For SATA/SCSI devices, remove numeric suffix
		for i := len(deviceName) - 1; i >= 0; i-- {
			if deviceName[i] < '0' || deviceName[i] > '9' {
				return deviceName[:i+1]
			}
		}
	}
	return deviceName
}

// getDiskTypeByDevice determines the disk type for a specific device
func getDiskTypeByDevice(deviceName string) string {
	rotationalFile := fmt.Sprintf("/sys/block/%s/queue/rotational", deviceName)

	// Read rotational flag
	rotationalData, err := os.ReadFile(rotationalFile)
	if err != nil {
		klog.V(4).Infof("Failed to read rotational file for %s: %v", deviceName, err)
		return "Unknown"
	}

	rotational := strings.TrimSpace(string(rotationalData))

	if rotational == "0" {
		// It's an SSD, check if it's NVMe
		if strings.HasPrefix(deviceName, "nvme") {
			return "NVMe SSD"
		}
		return "SSD"
	} else if rotational == "1" {
		return "HDD"
	}

	return "Unknown"
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
