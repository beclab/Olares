package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/joho/godotenv"
	cpu "github.com/klauspost/cpuid/v2"
	"github.com/mackerelio/go-osstat/uptime"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

const (
	// shutdownModeReboot indicates the system is going down for a reboot/kexec.
	shutdownModeReboot = "reboot"
	// shutdownModePoweroff indicates the system is powering off/halting.
	shutdownModePoweroff = "shutdown"
)

// systemdShutdownUnits maps the systemd target units that drive a shutdown
// transition to the mode they represent. When one of these targets has a
// pending job, the system is going down.
//
// NOTE: we key off the *pending job*, not the unit's ActiveState. An immediate
// `reboot`/`poweroff` enqueues a start job for the corresponding target via
// logind (StartUnit(..., "replace-irreversibly")), but the target only becomes
// "active" after its systemd-{reboot,poweroff}.service runs, and that service
// never returns (it reboots/powers off the kernel). So in practice the target
// stays "inactive (dead)" for the whole shutdown window while its start job is
// queued. The job is therefore the only reliable pollable signal.
var systemdShutdownUnits = map[string]string{
	"reboot.target":   shutdownModeReboot,
	"kexec.target":    shutdownModeReboot,
	"poweroff.target": shutdownModePoweroff,
	"halt.target":     shutdownModePoweroff,
}

// systemdJob mirrors the struct returned by
// org.freedesktop.systemd1.Manager.ListJobs (a(usssoo)), i.e. the same job list
// shown by `systemctl list-jobs`.
type systemdJob struct {
	ID       uint32
	Unit     string
	JobType  string
	State    string
	JobPath  dbus.ObjectPath
	UnitPath dbus.ObjectPath
}

// GetSystemPendingShutdowm reports whether the system is about to shut down or
// reboot, and in which mode ("reboot" or "shutdown").
//
// Relying on /run/systemd/shutdown/scheduled alone is not enough: that file is
// only created by systemd-logind for *delayed* shutdowns (e.g. `shutdown +5`).
// Immediate commands such as `reboot`, `reboot now`, `poweroff` or
// `shutdown now` go straight through logind and never create it. Instead we
// query systemd over D-Bus, combining:
//  1. logind's ScheduledShutdown property (timed shutdowns), and
//  2. pending systemd jobs for the shutdown target units (immediate and
//     in-progress shutdowns).
//
// The legacy file-based check is kept as a fallback when D-Bus is unavailable.
func GetSystemPendingShutdowm() (mode string, shuttingdown bool, err error) {
	if !IsLinux() {
		return "", false, nil
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		klog.Warningf("connect system dbus failed, fallback to file based shutdown check: %v", err)
		return getPendingShutdownFromFile()
	}

	// 1. delayed/timed shutdown registered in logind
	if m, ok := getScheduledShutdownFromLogind(conn); ok {
		return m, true, nil
	}

	// 2. immediate or in-progress shutdown reflected by pending systemd jobs
	m, ok, e := getActiveShutdownFromSystemd(conn)
	if e != nil {
		klog.Warningf("query systemd shutdown units failed, fallback to file based shutdown check: %v", e)
		return getPendingShutdownFromFile()
	}
	if ok {
		return m, true, nil
	}

	return "", false, nil
}

// getScheduledShutdownFromLogind reads logind's ScheduledShutdown property,
// which is set for delayed shutdowns (the same source that backs
// /run/systemd/shutdown/scheduled).
func getScheduledShutdownFromLogind(conn *dbus.Conn) (mode string, scheduled bool) {
	obj := conn.Object("org.freedesktop.login1", dbus.ObjectPath("/org/freedesktop/login1"))
	v, err := obj.GetProperty("org.freedesktop.login1.Manager.ScheduledShutdown")
	if err != nil {
		klog.Warningf("read logind ScheduledShutdown property failed: %v", err)
		return "", false
	}

	// the property is a struct (s: action, t: usec)
	fields, ok := v.Value().([]interface{})
	if !ok || len(fields) == 0 {
		return "", false
	}

	action, _ := fields[0].(string)
	action = strings.TrimSpace(action)
	// empty action means nothing is scheduled; dry-run actions only send a wall
	// message and do not actually power the machine down.
	if action == "" || strings.HasPrefix(action, "dry-") {
		return "", false
	}

	return normalizeShutdownMode(action), true
}

// getActiveShutdownFromSystemd checks whether systemd has a pending job for any
// of the shutdown target units. This is the same list shown by
// `systemctl list-jobs`: an immediate `reboot`/`poweroff`/`halt` enqueues a
// start job for the corresponding target that stays queued for the whole
// shutdown window (the target never reaches ActiveState=active before the
// machine goes down), so the job is the reliable pollable signal.
func getActiveShutdownFromSystemd(conn *dbus.Conn) (mode string, shuttingdown bool, err error) {
	obj := conn.Object("org.freedesktop.systemd1", dbus.ObjectPath("/org/freedesktop/systemd1"))
	var jobs []systemdJob
	if err = obj.Call("org.freedesktop.systemd1.Manager.ListJobs", 0).Store(&jobs); err != nil {
		return "", false, err
	}

	for _, j := range jobs {
		if m, tracked := systemdShutdownUnits[j.Unit]; tracked {
			return m, true, nil
		}
	}

	return "", false, nil
}

// normalizeShutdownMode maps a logind/systemd action string to the coarse mode
// used by callers ("reboot" or "shutdown").
func normalizeShutdownMode(action string) string {
	if strings.Contains(action, shutdownModeReboot) || strings.Contains(action, "kexec") {
		return shutdownModeReboot
	}
	return shutdownModePoweroff
}

// getPendingShutdownFromFile is the legacy detection based on
// /run/systemd/shutdown/scheduled, used only as a fallback when D-Bus cannot be
// reached. It only detects delayed shutdowns.
func getPendingShutdownFromFile() (mode string, shuttingdown bool, err error) {
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

	shuttingdown = true
	mode, ok := envs["MODE"]
	if !ok {
		mode = shutdownModePoweroff
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

var (
	cpuNameOnce  sync.Once
	cpuNameValue string
)

// GetCPUName returns the CPU brand name. The result is static for the lifetime
// of the process, so it is computed once and cached: the fallbacks below shell
// out to lscpu/dmidecode, which is wasteful to repeat every status tick.
func GetCPUName() string {
	cpuNameOnce.Do(func() {
		cpuNameValue = computeCPUName()
	})
	return cpuNameValue
}

func computeCPUName() string {
	brandName := cpu.CPU.BrandName
	if brandName == "" {
		// cannot read info from /proc/cpuinfo, try to get from lscpu command
		cmd := exec.Command("sh", "-c", "lscpu | awk -F: '/BIOS Model name/ {print $2}' | head -1 | sed 's/^[ \t]*//'")
		output, err := cmd.Output()
		if err != nil {
			klog.Error("get CPU name error, ", err)
			return ""
		}
		brandName = strings.TrimSpace(string(output))
	}

	// try to get AIBOOK M1000 model name
	if brandName == "" {
		cmd := exec.Command("sh", "-c", "command -v dmidecode 1>/dev/null && dmidecode -s processor-version")
		output, err := cmd.Output()
		if err != nil {
			klog.Error("get CPU name error, ", err)
			return ""
		}
		brandName = strings.TrimSpace(string(output))
	}

	// try to get rockchip model name for rockchip devices which cannot get cpu info from /proc/cpuinfo
	if brandName == "" {
		cmd := exec.Command("sh", "-c", "test -f /proc/device-tree/model && cat /proc/device-tree/model || true")
		output, err := cmd.Output()
		if err != nil {
			klog.Error("get CPU name error, ", err)
			return ""
		}
		brandName = strings.TrimSpace(string(output))
	}

	return brandName
}

func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsWSL() bool {
	if !IsLinux() {
		return false
	}

	// get kernal name from /proc/sys/kernel/osrelease
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	kernelName := strings.TrimSpace(string(data))
	if strings.Contains(kernelName, "-WSL") {
		return true
	}
	return false
}
