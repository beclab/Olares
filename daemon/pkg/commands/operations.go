package commands

import (
	"context"
	"fmt"
	"os"
	osexec "os/exec"
	"regexp"
	"strings"
	"sync"

	"k8s.io/klog/v2"
	"k8s.io/utils/exec"
)

const commandErrorTailLimit = 4096

var (
	commandBearer = regexp.MustCompile(`(?i)(\bbearer\s+)([^\s,;]+)`)
	commandSecret = regexp.MustCompile(`(?i)((?:[a-z0-9_-]*(?:password|passwd|token|secret|authorization|api[_-]?key))\s*[=:]\s*|--(?:password|passwd|token|secret|api-key)\s+)([^\s,;]+)`)
)

// commandErrorTail keeps only the end of stderr. CLI prints its final error
// there, while its full progress and diagnostic output remains in log files.
type commandErrorTail struct {
	mu        sync.Mutex
	data      []byte
	truncated bool
}

func (t *commandErrorTail) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(p) >= commandErrorTailLimit {
		t.truncated = true
		t.data = append(t.data[:0], p[len(p)-commandErrorTailLimit:]...)
		return len(p), nil
	}
	t.data = append(t.data, p...)
	if len(t.data) > commandErrorTailLimit {
		t.truncated = true
		t.data = append([]byte(nil), t.data[len(t.data)-commandErrorTailLimit:]...)
	}
	return len(p), nil
}

func (t *commandErrorTail) summary() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	data := t.data
	if t.truncated {
		// The first retained line may start in the middle of a secret.
		// Discard it instead of exposing an unrecognizable value fragment.
		end := strings.IndexByte(string(data), '\n')
		if end < 0 {
			return ""
		}
		data = data[end+1:]
	}
	message := strings.TrimSpace(string(data))
	message = commandBearer.ReplaceAllString(message, "${1}[redacted]")
	message = commandSecret.ReplaceAllString(message, "${1}[redacted]")
	if len(message) > 1024 {
		message = message[len(message)-1024:]
	}
	return strings.TrimSpace(message)
}

type Operations string

const (
	Install                  Operations = "install"
	Initialize               Operations = "initialize"
	ChangeIp                 Operations = "changeIp"
	Uninstall                Operations = "uninstall"
	CreateUpgradeTarget      Operations = "createUpgradeTarget"
	RemoveUpgradeTarget      Operations = "removeUpgradeTarget"
	DownloadCLI              Operations = "downloadCLI"
	DownloadWizard           Operations = "downloadWizard"
	UpgradePreCheck          Operations = "PreCheck"
	DownloadSpaceCheck       Operations = "downloadSpaceCheck"
	DownloadComponent        Operations = "downloadComponent"
	ImportImages             Operations = "importImages"
	InstallOlaresd           Operations = "installOlaresd"
	Upgrade                  Operations = "upgrade"
	InstallCLI               Operations = "installCLI"
	Reboot                   Operations = "reboot"
	Shutdown                 Operations = "shutdown"
	ConnectWifi              Operations = "connectWifi"
	ChangeHost               Operations = "changeHost"
	UmountUsb                Operations = "umountUsb"
	CollectLogs              Operations = "collectLogs"
	MountSmb                 Operations = "mountSmb"
	UmountSmb                Operations = "umountSmb"
	SetSSHPassword           Operations = "setSSHPassword"
	MountNfs                 Operations = "mountNfs"
	UmountNfs                Operations = "umountNfs"
	EnableOverlayGateway     Operations = "enableOverlayGateway"
	DisableOverlayGateway    Operations = "disableOverlayGateway"
	EnableAppOverlayGateway  Operations = "enableAppOverlayGateway"
	DisableAppOverlayGateway Operations = "disableAppOverlayGateway"
)

func (p Operations) Stirng() string {
	return string(p)
}

type BaseCommand struct {
	executor exec.Interface
	dir      string
	slient   bool
	pidFile  string
	envs     map[string]string
	watchDog func(ctx context.Context)
}

func NewBaseCommand() *BaseCommand {
	return &BaseCommand{executor: exec.New(), envs: make(map[string]string)}
}

func (c *BaseCommand) Run_(ctx context.Context, cmdStr string, args ...string) (string, error) {
	cmd := c.executor.CommandContext(ctx, cmdStr, args...)
	if c.dir != "" {
		cmd.SetDir(c.dir)
	}

	var (
		err    error
		output []byte
	)

	if c.slient {
		err = cmd.Run()
	} else {
		output, err = cmd.CombinedOutput()
		klog.Info("command output: \n", string(output))
	}

	if err != nil {
		klog.Error("run command error, ", err, ", ", cmdStr)
	}
	return string(output), err
}

func (c *BaseCommand) WithDir_(dir string) *BaseCommand {
	c.dir = dir
	return c
}

func (c *BaseCommand) WithPid_(path string) *BaseCommand {
	c.pidFile = path
	return c
}

func (c *BaseCommand) Silent_() *BaseCommand {
	c.slient = true
	return c
}

func (c *BaseCommand) AddEnv_(key, value string) *BaseCommand {
	c.envs[key] = value
	return c
}

func (c *BaseCommand) WithWatchDog_(fn func(ctx context.Context)) *BaseCommand {
	c.watchDog = fn
	return c
}

func (c *BaseCommand) RunAsync_(ctx context.Context, cmdStr string, args ...string) error {
	_, err := c.runAsync_(ctx, false, cmdStr, args...)
	return err
}

// RunAsyncWithResult_ reports the final process result to callers that must
// distinguish an unsuccessful CLI run from a missing progress log marker.
func (c *BaseCommand) RunAsyncWithResult_(ctx context.Context, cmdStr string, args ...string) (<-chan error, error) {
	return c.runAsync_(ctx, true, cmdStr, args...)
}

func (c *BaseCommand) runAsync_(ctx context.Context, reportResult bool, cmdStr string, args ...string) (<-chan error, error) {
	cmd := osexec.CommandContext(ctx, cmdStr, args...)
	if c.dir != "" {
		cmd.Dir = c.dir
	}

	cmd.Env = append(os.Environ(), cmd.Env...)
	for k, v := range c.envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	var stderr commandErrorTail
	if reportResult {
		cmd.Stderr = &stderr
	}

	err := cmd.Start()
	if err != nil {
		klog.Error("run command error, ", err, ", ", cmdStr, " ", args)
		return nil, err
	}

	if c.pidFile != "" {
		pid := cmd.Process.Pid
		err = os.WriteFile(c.pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
		if err != nil {
			klog.Error("write pid file error, ", err)
			return nil, err
		}
	}
	var result chan error
	if reportResult {
		result = make(chan error, 1)
	}

	go func() {
		var (
			cancel   context.CancelFunc
			watchCtx context.Context
		)

		if c.watchDog != nil {
			watchCtx, cancel = context.WithCancel(ctx)
			c.watchDog(watchCtx)
		}

		defer func() {
			if c.pidFile != "" {
				err = os.Remove(c.pidFile)
				if err != nil {
					klog.Warning("remove pid error, ", err)
				}

			}

			if cancel != nil {
				cancel()
			}
		}()

		waitErr := cmd.Wait()
		if waitErr != nil {
			klog.Errorf("Command finished with error: %v, %s %v", waitErr, cmdStr, args)
			if detail := stderr.summary(); detail != "" {
				waitErr = fmt.Errorf("%w: %s", waitErr, detail)
			}
		}
		if result != nil {
			result <- waitErr
			close(result)
		}
		if waitErr != nil {
			return
		}

		klog.Info("run command completed, ", cmdStr, " ", args)
	}()

	return result, nil
}
