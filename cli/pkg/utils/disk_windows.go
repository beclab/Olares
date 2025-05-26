package utils

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func GetDrives() []string {
	var drives []string
	for i := 'A'; i <= 'Z'; i++ {
		drive := fmt.Sprintf("%c:\\", i)
		if _, err := os.Stat(drive); err == nil {
			drives = append(drives, drive)
		}
	}
	return drives
}

func GetDiskSpace(path string) (total uint64, free uint64, err error) {
	lpDirectoryName, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, err
	}

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64

	err = windows.GetDiskFreeSpaceEx(
		lpDirectoryName,
		&freeBytesAvailable,
		&totalNumberOfBytes,
		&totalNumberOfFreeBytes,
	)
	if err != nil {
		return 0, 0, err
	}

	return totalNumberOfBytes, totalNumberOfFreeBytes, nil
}
