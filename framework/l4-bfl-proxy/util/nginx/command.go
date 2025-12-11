package nginx

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

var (
	DefNgxBinary  = "/usr/local/openresty/bin/openresty"
	DefNgxCfgPath = "/etc/nginx/nginx.conf"
)

// Command stores context around a given nginx executable path
type Command struct {
	binary   string
	confPath string
}

// NewCommand returns a new Command from which path
// has been detected from environment variable NGINX_BINARY or default
func NewCommand() *Command {
	command := Command{
		binary:   DefNgxBinary,
		confPath: DefNgxCfgPath,
	}
	binary := os.Getenv("NGINX_BINARY")
	if binary != "" {
		command.binary = binary
		DefNgxBinary = binary
	}

	ngxCfgPath := os.Getenv("NGINX_CONF_PATH")
	if ngxCfgPath != "" {
		command.confPath = ngxCfgPath
		DefNgxCfgPath = ngxCfgPath
	}
	return &command
}

func (n *Command) output(args ...string) ([]byte, error) {
	return exec.Command(n.binary, args...).CombinedOutput()
}

func (n *Command) Start() ([]byte, error) {
	return n.output("-c", n.confPath)
}

func (n *Command) StartCmd(args ...string) *exec.Cmd {
	var cmdArgs []string

	cmdArgs = append(cmdArgs, "-c", n.confPath)
	cmdArgs = append(cmdArgs, args...)
	return exec.Command(n.binary, cmdArgs...)
}

// Test checks if config file is a syntax valid nginx configuration
func (n *Command) Test(cfg string) ([]byte, error) {
	var confPath = n.confPath
	if cfg != "" {
		confPath = cfg
	}
	return n.output("-c", confPath, "-t")
}

func (n *Command) Reload() ([]byte, error) {
	return n.output("-s", "reload")
}

func (n *Command) Quit() ([]byte, error) {
	return n.output("-s", "quit")
}

func (n *Command) Version() ([]byte, error) {
	return n.output("-v")
}

func (n *Command) VersionAndOption() ([]byte, error) {
	return n.output("-V")
}

func IsRunning() bool {

	// out, err := exec.Command("/usr/bin/pgrep", "/sbin/nginx -c /etc/nginx/nginx.conf").Output()
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
