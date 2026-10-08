package commands

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// NewCommandContext gives an upgrade CLI its own process group. Cancelling a
// stage must stop the CLI's helm/download children before another attempt starts.
func NewCommandContext(ctx context.Context, binary string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = 5 * time.Second
	return cmd
}
