package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"k8s.io/klog/v2"
)

const (
	TERMINUS_CLI = "/usr/local/bin/olares-cli"
)

type terminusCli struct {
}

func NewTerminusCli() (*terminusCli, error) {
	if _, err := os.Stat(TERMINUS_CLI); err != nil {
		klog.Error(err, ", ", TERMINUS_CLI)
		return nil, err
	}

	return &terminusCli{}, nil
}

func (c *terminusCli) InitTerminus(ip string) (context.CancelFunc, error) {
	// run in background
	ctx, cancel := context.WithCancel(context.Background())
	// TODO: kube type
	cmd := exec.CommandContext(ctx, TERMINUS_CLI, "init", "--kube", "k3s")
	go func() {
		cmd.Env = append(os.Environ(), fmt.Sprintf("OS_LOCALIP=%s", ip))
		if err := cmd.Run(); err != nil {
			klog.Error("running command 'olares init' error, ", err)
			return
		}

		klog.Info("finished running command 'olares init'")
	}()

	return func() {
		cmd.Cancel()
		cancel()
	}, nil
}
