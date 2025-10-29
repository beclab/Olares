package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	unix "golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

// FilesystemStat represents capacity and inode stats for a mounted filesystem.
type FilesystemStat struct {
	Device     string
	MountPoint string
	FSType     string
	SizeBytes  uint64
	FreeBytes  uint64
	AvailBytes uint64
	Files      uint64
	FilesFree  uint64
	ReadOnly   bool
}

const (
	ignoredMountPointsPattern = "^/(dev|proc|sys|var/lib/docker/.+|var/lib/containerd/.+|var/lib/kubelet/.+)($|/)"
	ignoredFSTypesPattern     = "^(autofs|binfmt_misc|cgroup|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|mqueue|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|sysfs|tracefs)$"
)

func GetNodeFilesystemTotalSize() (uint64, error) {
	stats, err := GetFilesystemStats()
	if err != nil {
		return 0, fmt.Errorf("unable to get filesystem stats: %w", err)
	}

	var totalSize uint64
	filteredStats := make(map[string]FilesystemStat)
	for _, stat := range stats {
		if !strings.HasPrefix(stat.Device, "/dev") || strings.HasPrefix(stat.Device, "/dev/loop") {
			continue
		}
		if _, ok := filteredStats[stat.Device]; ok {
			continue
		}
		filteredStats[stat.Device] = stat
	}

	for _, stat := range filteredStats {
		totalSize += stat.SizeBytes
	}

	return totalSize, nil
}

// GetFilesystemStats returns filesystem stats for all mounts, filtered by built-in regex patterns.
func GetFilesystemStats() ([]FilesystemStat, error) {
	mounts, err := readMountInfo()
	if err != nil {
		return nil, fmt.Errorf("unable to read mount info: %w", err)
	}

	mpFilter, err := regexp.Compile(ignoredMountPointsPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid built-in mount points regex: %w", err)
	}
	fsFilter, err := regexp.Compile(ignoredFSTypesPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid built-in fs types regex: %w", err)
	}

	var out []FilesystemStat
	for _, m := range mounts {
		if mpFilter.MatchString(m.mountPoint) {
			continue
		}
		if fsFilter.MatchString(m.fsType) {
			continue
		}

		var readOnly bool
		for _, opt := range strings.Split(m.options, ",") {
			if opt == "ro" {
				readOnly = true
				break
			}
		}

		buf := new(unix.Statfs_t)
		if err := unix.Statfs(m.mountPoint, buf); err != nil {
			klog.Warningf("unable to statfs mount point %q: %v", m.mountPoint, err)
			continue
		}

		bsize := uint64(buf.Bsize)
		out = append(out, FilesystemStat{
			Device:     m.device,
			MountPoint: m.mountPoint,
			FSType:     m.fsType,
			SizeBytes:  uint64(buf.Blocks) * bsize,
			FreeBytes:  uint64(buf.Bfree) * bsize,
			AvailBytes: uint64(buf.Bavail) * bsize,
			Files:      uint64(buf.Files),
			FilesFree:  uint64(buf.Ffree),
			ReadOnly:   readOnly,
		})
	}
	return out, nil
}

type mountInfo struct {
	device     string
	mountPoint string
	fsType     string
	options    string
}

// readMountInfo parses /proc/1/mountinfo (or falls back to self) and returns essential fields.
func readMountInfo() ([]mountInfo, error) {
	file, err := os.Open("/proc/1/mountinfo")
	if errors.Is(err, os.ErrNotExist) {
		file, err = os.Open("/proc/self/mountinfo")
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var mounts []mountInfo
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 10 {
			return nil, fmt.Errorf("malformed mount point information: %q", scanner.Text())
		}

		m := 5
		for parts[m+1] != "-" {
			m++
		}

		// Unescape as per fstab: \040 (space), \011 (tab).
		mountPoint := strings.ReplaceAll(parts[4], "\\040", " ")
		mountPoint = strings.ReplaceAll(mountPoint, "\\011", "\t")

		mounts = append(mounts, mountInfo{
			device:     parts[m+3],
			mountPoint: mountPoint,
			fsType:     parts[m+2],
			options:    parts[5],
		})
	}
	return mounts, scanner.Err()
}
