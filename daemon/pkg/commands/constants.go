package commands

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
)

var (
	INSTALLED_VERSION = ""
	KUBE_TYPE         = "k3s"
	COMMAND_BASE_DIR  = "" // deprecated shell command base dir
	CDN_URL           = "https://dc3p1870nn3cj.cloudfront.net"

	OS_ROOT_DIR               = "/olares"
	INSTALLING_PID_FILE       = "installing.pid"
	UNINSTALLING_PID_FILE     = "uninstalling.pid"
	CHANGINGIP_PID_FILE       = "changingip.pid"
	UPGRADE_TARGET_FILE       = "upgrade.target"
	UPGRADE_DOWNLOADONLY_FILE = "upgrade.downloadonly"
	UPGRADE_DOWNLOADED_FILE   = "upgrade.downloaded"
	PREV_IP_TO_CHANGE_FILE    = ".prev_ip"
	PREV_IP_CHANGE_FAILED     = ".ip_change_failed"
	INSTALL_LOCK              = ".installed"
	LOG_FILE                  = "install.log"
	TERMINUS_BASE_DIR         = ""
	MOUNT_BASE_DIR            = path.Join(OS_ROOT_DIR, "share")
	PREPARE_LOCK              = ".prepared"
	REDIS_CONF                = OS_ROOT_DIR + "/data/redis/etc/redis.conf"
	EXPORT_POD_LOGS_DIR       = "Home/pod_logs"

	ProgressNumFinished = 100
)

func Init() {
	baseDir := mustEnv("BASE_DIR")
	INSTALLED_VERSION = mustEnv("INSTALLED_VERSION")
	KUBE_TYPE = os.Getenv("KUBE_TYPE")
	CDN_URL = os.Getenv("DOWNLOAD_CDN_URL")

	TERMINUS_BASE_DIR = baseDir
	INSTALLING_PID_FILE = filepath.Join(baseDir, INSTALLING_PID_FILE)
	UNINSTALLING_PID_FILE = filepath.Join(baseDir, UNINSTALLING_PID_FILE)
	CHANGINGIP_PID_FILE = filepath.Join(baseDir, CHANGINGIP_PID_FILE)
	UPGRADE_TARGET_FILE = filepath.Join(baseDir, UPGRADE_TARGET_FILE)
	UPGRADE_DOWNLOADONLY_FILE = filepath.Join(baseDir, UPGRADE_DOWNLOADONLY_FILE)
	UPGRADE_DOWNLOADED_FILE = filepath.Join(baseDir, UPGRADE_DOWNLOADED_FILE)
	INSTALL_LOCK = filepath.Join(baseDir, INSTALL_LOCK)
	PREPARE_LOCK = filepath.Join(baseDir, PREPARE_LOCK)
	PREV_IP_TO_CHANGE_FILE = filepath.Join(baseDir, PREV_IP_TO_CHANGE_FILE)
	PREV_IP_CHANGE_FAILED = filepath.Join(baseDir, PREV_IP_CHANGE_FAILED)

	COMMAND_BASE_DIR = filepath.Join(baseDir, "versions", "v"+INSTALLED_VERSION)
	LOG_FILE = filepath.Join(COMMAND_BASE_DIR, "logs", LOG_FILE)

	klog.Info("var INSTALLING_PID_FILE, ", INSTALLING_PID_FILE)
	klog.Info("var UNINSTALLING_PID_FILE, ", UNINSTALLING_PID_FILE)
	klog.Info("var CHANGINGIP_PID_FILE, ", CHANGINGIP_PID_FILE)
	klog.Info("var INSTALL_LOCK, ", INSTALL_LOCK)
	klog.Info("var PREPARE_LOCK, ", PREPARE_LOCK)
	klog.Info("var COMMAND_BASE_DIR, ", COMMAND_BASE_DIR)
	klog.Info("var LOG_FILE, ", LOG_FILE)
	klog.Info("var MOUNT_BASE_DIR, ", MOUNT_BASE_DIR)
}

func mustEnv(env string) string {
	e := strings.TrimSpace(os.Getenv(env))
	if e == "" {
		panic(fmt.Errorf("env [%s] value is empty", env))
	}

	return e
}
