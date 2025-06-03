package utils

import (
	"io/ioutil"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"k8s.io/klog/v2"
)

const osxCmd = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"
const osxArgs = "-I"
const linuxCmd = "iwgetid"
const linuxArgs = "--raw"

func WifiName() *string {
	platform := runtime.GOOS
	if platform == "darwin" {
		return forOSX()
	} else if platform == "win32" {
		// TODO for Windows
		return nil
	} else {
		// TODO for Linux
		return forLinux()
	}
}

func forLinux() *string {
	cmd := exec.Command(linuxCmd, linuxArgs)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		klog.Error("get wifi ssid error, ", err)
		return nil
	}

	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		klog.Error("get wifi ssid error, ", err)
		return nil
	}
	defer cmd.Wait()

	var str string

	if b, err := ioutil.ReadAll(stdout); err == nil {
		str += (string(b) + "\n")
	}

	name := strings.Replace(str, "\n", "", -1)
	return &name
}

func forOSX() *string {

	cmd := exec.Command(osxCmd, osxArgs)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		klog.Error("get wifi ssid error, ", err)
		return nil
	}

	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		klog.Error("get wifi ssid error, ", err)
		return nil
	}
	defer cmd.Wait()
	var str string

	if b, err := ioutil.ReadAll(stdout); err == nil {
		str += (string(b) + "\n")
	}

	r := regexp.MustCompile(`s*SSID: (.+)s*`)

	name := r.FindAllStringSubmatch(str, -1)

	if len(name) <= 1 {
		return nil
	} else {
		return &name[1][1]
	}
}
