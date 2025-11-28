package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mackerelio/go-osstat/uptime"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

func GetSystemPendingShutdowm() (mode string, shuttingdown bool, err error) {
	path := "/run/systemd/shutdown/scheduled"
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			return
		}

		klog.Error("read system pending shutdown error, ", err)
		return
	}

	envs, err := godotenv.Read(path)
	if err != nil {
		klog.Error("read pending shudown file error, ", err)
		return
	}

	mode, ok := envs["MODE"]
	if !ok {
		mode = "shutdown"
	}

	return
}

func GetDeviceName() *string {
	data, err := os.ReadFile("/etc/machine.info")
	if err != nil {
		if os.IsNotExist(err) {
			// default device name
			return pointer.String("Selfhosted")
		}

		klog.Error("read machine info err, ", err)
	} else {
		return pointer.String(strings.TrimSpace(string(data)))
	}

	return nil
}

func IsEmptyDir(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// read in ONLY one file
	_, err = f.Readdir(1)

	// and if the file is EOF... well, the dir is empty.
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func SystemStartLessThan(minute time.Duration) (bool, error) {
	sysUptime, err := uptime.Get()
	if err != nil {
		klog.Error("get system uptime error, ", err)
		return false, err
	}

	return sysUptime <= minute, nil
}

func MoveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}

	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}

	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}

	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}

	return nil
}

func GetDataFromReleaseFile() (map[string]string, error) {
	data, err := godotenv.Read("/etc/olares/release")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("olars release file not found")
		}
		return nil, fmt.Errorf("read olars release file error: %w", err)
	}

	return data, nil
}

func GetOlaresNameFromReleaseFile() (string, error) {
	data, err := GetDataFromReleaseFile()
	if err != nil {
		return "", err
	}

	name := data["OLARES_NAME"]
	return name, nil
}

func GetBaseDirFromReleaseFile() (string, error) {
	data, err := GetDataFromReleaseFile()
	if err != nil {
		return "", err
	}

	baseDir := data["OLARES_BASE_DIR"]
	return baseDir, nil
}
