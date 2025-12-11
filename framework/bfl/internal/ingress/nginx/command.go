package nginx

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

const (
	DefNgxBinary               = "/usr/local/openresty/bin/openresty"
	DefNgxCfgPath              = "/etc/nginx/nginx.conf"
	DefNgxSSLCertificationPath = "/etc/nginx/ssl"
)

// NginxCommand stores context around a given nginx executable path
type NginxCommand struct {
	binary   string
	confPath string
	errorLog *errorLog
}

type errorLog struct {
	latestLogs map[string]struct{}
}

func (e *errorLog) load() error {
	// read the last 100 lines of error log
	//  tail -n 100 /var/log/nginx/error.log
	out, err := exec.Command("tail", "-n", "100", "/var/log/nginx/error.log").CombinedOutput()
	if err != nil {
		return err
	}

	logs := strings.Split(string(out), "\n")
	for _, log := range logs {
		e.latestLogs[strings.TrimSpace(log)] = struct{}{}
	}
	return nil
}

func newErrlog() (*errorLog, error) {
	errlog := &errorLog{
		latestLogs: make(map[string]struct{}),
	}
	err := errlog.load()
	return errlog, err
}

func hasNewError(newErrlog, oldErrorlog *errorLog) error {
	if len(newErrlog.latestLogs) == 0 {
		return nil
	}

	hasError := func(log string) bool {
		return strings.Contains(log, "[emerg]") || strings.Contains(log, "[crit]")
	}

	if len(oldErrorlog.latestLogs) == 0 {
		for log := range newErrlog.latestLogs {
			if hasError(log) {
				return errors.New("nginx error log detected: " + log)
			}
		}
		return nil
	}

	// compare old error log with new error log
	for log := range newErrlog.latestLogs {
		if !hasError(log) {
			continue
		}

		if _, exists := oldErrorlog.latestLogs[log]; !exists {
			return errors.New("nginx error log detected: " + log)
		}
	}

	return nil
}

// NewNginxCommand returns a new NginxCommand from which path
// has been detected from environment variable NGINX_BINARY or default
func NewNginxCommand() *NginxCommand {
	command := NginxCommand{
		binary:   DefNgxBinary,
		confPath: DefNgxCfgPath,
	}
	binary := os.Getenv("NGINX_BINARY")
	if binary != "" {
		command.binary = binary
	}

	ngxCfgPath := os.Getenv("NGINX_CONF_PATH")
	if ngxCfgPath != "" {
		command.confPath = ngxCfgPath
	}

	var err error
	command.errorLog, err = newErrlog()
	if err != nil {
		klog.V(2).ErrorS(err, "load error log")
	}
	return &command
}

func (n *NginxCommand) StartCmd(args ...string) *exec.Cmd {
	var cmdArgs []string

	cmdArgs = append(cmdArgs, "-c", n.confPath)
	cmdArgs = append(cmdArgs, args...)
	return exec.Command(n.binary, cmdArgs...)
}

func (n *NginxCommand) output(args ...string) ([]byte, error) {
	return exec.Command(n.binary, args...).CombinedOutput()
}

// Test checks if config file is a syntax valid nginx configuration
func (n *NginxCommand) Test(cfg string) ([]byte, error) {
	var confPath = n.confPath
	if cfg != "" {
		confPath = cfg
	}
	return n.output("-c", confPath, "-t")
}

func (n *NginxCommand) Reload() ([]byte, error) {
	out, err := n.output("-s", "reload")
	if err != nil {
		return out, err
	}

	newErrlog, err := newErrlog()
	if err != nil {
		klog.V(2).ErrorS(err, "load error log")
		return out, err
	}

	defer func() {
		n.errorLog = newErrlog
	}()

	if err := hasNewError(newErrlog, n.errorLog); err != nil {
		return out, err
	}

	return out, nil
}

func (n *NginxCommand) Quit() ([]byte, error) {
	return n.output("-s", "quit")
}

func (n *NginxCommand) Version() ([]byte, error) {
	return n.output("-v")
}

func (n *NginxCommand) VersionAndOption() ([]byte, error) {
	return n.output("-V")
}

func IsRunning() bool {
	// processes, _ := ps.Processes()
	// for _, p := range processes {
	// 	if p.Executable() == "nginx" {
	// 		return true
	// 	}
	// }

	// return false
	out, err := os.ReadFile(PID)
	if err != nil {
		klog.V(2).ErrorS(err, "read ", PID)
		return false
	}

	pid := string(out)
	if pid != "" {
		pid = strings.TrimSpace(pid)
		_, err = strconv.ParseUint(pid, 10, 32)
		if err == nil {
			return true
		}
	}
	return false

}
