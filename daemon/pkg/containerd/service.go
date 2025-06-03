package containerd

import (
	"context"
	"k8s.io/klog/v2"
	"os/exec"
	"time"
)

// todo: wait for containerd to be connectable again?
// we can't easily determine whether it's initializing or is stuck in some permanent issues
// to avoid block here forever, just sleep 0.5s for now
func restartContainerd(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "systemctl", "restart", "containerd")
	err := cmd.Run()
	if err == nil {
		klog.Info("successfully restarted containerd")
		time.Sleep(500 * time.Millisecond)
	} else {
		klog.Error("failed to restart containerd")
	}
	return err
}
