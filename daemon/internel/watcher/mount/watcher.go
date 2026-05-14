package mount

import (
	"context"
	"slices"

	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

var _ watcher.Watcher = &mountWatcher{}

type mountWatcher struct{}

func NewMountWatcher() *mountWatcher {
	return &mountWatcher{}
}

func (w *mountWatcher) Watch(ctx context.Context) {
	if state.CurrentState.TerminusState != state.TerminusRunning {
		return
	}

	records, err := utils.LoadMountRecords(commands.MOUNT_RECORDS_FILE)
	if err != nil {
		klog.Warning("load mount records error, ", err)
		return
	}

	if len(records) == 0 {
		return
	}

	mountedPoints, err := utils.GetMountedPoints(ctx)
	if err != nil {
		klog.Warning("list mounted path error in mount watcher, ", err)
		return
	}

	for _, record := range records {
		if slices.Contains(mountedPoints, record.MountPoint) {
			continue
		}

		klog.Infof("attempting to remount %s (%s)", record.MountPoint, record.Type)

		switch record.Type {
		case utils.SMB:
			if err := utils.MountSambaDriver(ctx, commands.MOUNT_BASE_DIR, record.SmbPath, record.User, record.Password); err != nil {
				klog.Warningf("remount smb %s failed: %v", record.MountPoint, err)
			} else {
				klog.Infof("remount smb %s success", record.MountPoint)
			}
		case utils.NFS:
			if err := utils.MountNfsDriver(ctx, commands.MOUNT_BASE_DIR, record.MountName, record.Server, record.NfsPath); err != nil {
				klog.Warningf("remount nfs %s failed: %v", record.MountPoint, err)
			} else {
				klog.Infof("remount nfs %s success", record.MountPoint)
			}
		}
	}
}
