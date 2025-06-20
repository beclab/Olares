package commands

import (
	"context"
	"fmt"
	"os"
	osexec "os/exec"

	"k8s.io/klog/v2"
	"k8s.io/utils/exec"
)

type Operations string

const (
	Install                   Operations = "install"
	Initialize                Operations = "initialize"
	ChangeIp                  Operations = "changeIp"
	Uninstall                 Operations = "uninstall"
	CreateUpgradeTarget       Operations = "createUpgradeTarget"
	RemoveUpgradeTarget       Operations = "removeUpgradeTarget"
	DownloadCLI               Operations = "downloadCLI"
	DownloadWizard            Operations = "downloadWizard"
	VersionCompatibilityCheck Operations = "versionCompatibilityCheck"
	UpgradeHealthCheck        Operations = "upgradeHealthCheck"
	DownloadComponent         Operations = "downloadComponent"
	ImportImages              Operations = "importImages"
	InstallOlaresd            Operations = "installOlaresd"
	Upgrade                   Operations = "upgrade"
	InstallCLI                Operations = "installCLI"
	Reboot                    Operations = "reboot"
	Shutdown                  Operations = "shutdown"
	ConnectWifi               Operations = "connectWifi"
	ChangeHost                Operations = "changeHost"
	UmountUsb                 Operations = "umountUsb"
	CollectLogs               Operations = "collectLogs"
	MountSmb                  Operations = "mountSmb"
	UmountSmb                 Operations = "umountSmb"
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
	cmd := osexec.CommandContext(ctx, cmdStr, args...)
	if c.dir != "" {
		cmd.Dir = c.dir
	}

	cmd.Env = append(os.Environ(), cmd.Env...)
	for k, v := range c.envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	err := cmd.Start()
	if err != nil {
		klog.Error("run command error, ", err, ", ", cmdStr, " ", args)
		return err
	}

	if c.pidFile != "" {
		pid := cmd.Process.Pid
		err = os.WriteFile(c.pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
		if err != nil {
			klog.Error("write pid file error, ", err)
			return err
		}
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

		err = cmd.Wait()
		if err != nil {
			klog.Errorf("Command finished with error: %v, %s %v", err, cmdStr, args)
			return
		}

		klog.Info("run command completed, ", cmdStr, " ", args)
	}()

	return nil
}
