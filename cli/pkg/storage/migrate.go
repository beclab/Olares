package storage

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	v1alpha1 "github.com/beclab/Olares/framework/app-service/api/sys.bytetrade.io/v1alpha1"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/pkg/errors"
	apixclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// MigrationTempMountPoint is a temporary mount point used to populate the newly
// formatted JuiceFS filesystem with the data from the local rootfs before the
// real rootfs directory is swapped out. It is intentionally a sibling of the
// real rootfs (not a child of it) so the two never overlap.
var MigrationTempMountPoint = path.Join(OlaresRootDir, ".rootfs-jfs-migrate")

// RootFSLocalBackupDir is where the original local rootfs directory is moved to
// after its content has been synced into JuiceFS. It is kept around (not
// deleted) so that the migration can be rolled back manually if needed.
var RootFSLocalBackupDir = path.Join(OlaresRootDir, "rootfs.local.bak")

// IsJuiceFSEnabled reports whether this node has already been switched over to a
// JuiceFS-backed rootfs.
//
// We deliberately key this off the existence of the juicefs systemd unit file
// rather than whether /olares/rootfs is currently a JuiceFS mount. The unit
// file is written by EnableJuiceFsService, which only runs AFTER the original
// local rootfs has been backed up and swapped out. So its presence guarantees
// the data migration already completed. Keying off "is currently mounted"
// instead would be unsafe: a migrated-but-stopped node would look un-migrated
// and could trigger a re-sync of an (empty) underlying directory into the live
// JuiceFS with --delete, destroying all data.
func IsJuiceFSEnabled() bool {
	return util.IsExist(JuiceFsServiceFile)
}

// isPathMounted returns whether the given absolute path is currently a mount
// point (of any filesystem type) by consulting /proc/self/mounts on the node.
func isPathMounted(runtime connector.Runtime, target string) (bool, error) {
	out, err := runtime.GetRunner().SudoCmd("cat /proc/self/mounts", false, false)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == target {
			return true, nil
		}
	}
	return false, nil
}

// CheckMigrationPrecheck validates that the node is in a state where the
// local rootfs can be migrated onto JuiceFS, and that there is enough free
// space on the data disk to hold a second copy of the data (only relevant for
// the bundled managed-MinIO backend, where the object data lives on the same
// local disk).
type CheckMigrationPrecheck struct {
	common.KubeAction
}

func (t *CheckMigrationPrecheck) Execute(runtime connector.Runtime) error {
	if !util.IsExist(OlaresJuiceFSRootDir) {
		return fmt.Errorf("rootfs directory %s does not exist, is Olares installed on this node?", OlaresJuiceFSRootDir)
	}

	// rsync is required for the two-phase data sync. Olares only supports
	// apt-based distros, so just install it if missing.
	if _, err := runtime.GetRunner().SudoCmd("command -v rsync", false, false); err != nil {
		logger.Info("rsync not found, installing it via apt")
		if _, err := runtime.GetRunner().SudoCmd("apt-get update && apt-get install -y rsync", false, true); err != nil {
			return errors.Wrap(err, "failed to install rsync, which is required for the rootfs data migration")
		}
	}

	// For external object storage (S3/OSS/COS/external MinIO) the data is not
	// written to the local disk, so the 2x space requirement does not apply.
	if t.KubeConf.Arg.Storage == nil || t.KubeConf.Arg.Storage.StorageType != common.ManagedMinIO {
		logger.Info("storage backend is not managed MinIO, skipping local disk space precheck")
		return nil
	}

	used, err := dirUsageBytes(runtime, OlaresJuiceFSRootDir)
	if err != nil {
		return errors.Wrap(err, "failed to determine the size of the current rootfs")
	}
	avail, err := availBytes(runtime, OlaresRootDir)
	if err != nil {
		return errors.Wrap(err, "failed to determine the available space on the data disk")
	}

	// require a 10% margin on top of the current usage
	required := used + used/10
	if avail < required {
		return fmt.Errorf("not enough free space to migrate rootfs onto the bundled MinIO: rootfs uses ~%d bytes, but only %d bytes are available on the disk holding %s (need ~%d). "+
			"Free up space, attach a larger disk, or use an external S3-compatible object store instead",
			used, avail, OlaresRootDir, required)
	}
	logger.Infof("disk space precheck passed: rootfs uses ~%d bytes, %d bytes available", used, avail)
	return nil
}

func dirUsageBytes(runtime connector.Runtime, dir string) (int64, error) {
	out, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("du -sb %s | awk '{print $1}'", dir), false, false)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(out), 10, 64)
}

func availBytes(runtime connector.Runtime, dir string) (int64, error) {
	out, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("df -B1 --output=avail %s | tail -n 1", dir), false, false)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(out), 10, 64)
}

// CleanupStaleJuiceFsBinaryPath removes a stale /usr/local/bin/juicefs that was
// auto-created as an (empty) directory by the osnode-init daemonset, which
// hostPath-mounts /usr/local/bin and bind-mounts the juicefs binary into its
// container via `subPath: juicefs`. On a single-node install that never had
// JuiceFS, the binary doesn't exist, so kubelet's subPath handling MkdirAll's
// the path into an empty directory.
//
// Left in place, CheckJuiceFsExists would treat that directory as an installed
// binary and skip installation, and `juicefs format` would then fail with
// "/usr/local/bin/juicefs: Is a directory". The path is only a bind-mount
// source (not a host mount point), so removing it is safe; the running
// osnode-init pod keeps its stale mount until it is restarted later in the
// migration, after which subPath correctly mounts the real binary file.
type CleanupStaleJuiceFsBinaryPath struct {
	common.KubeAction
}

func (t *CleanupStaleJuiceFsBinaryPath) Execute(runtime connector.Runtime) error {
	// Only remove it when it is a directory; never touch a real binary file.
	cmd := fmt.Sprintf("if [ -d %s ]; then rm -rf %s; fi", JuiceFsFile, JuiceFsFile)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return errors.Wrap(err, "failed to remove stale juicefs binary directory")
	}
	return nil
}

// MountJuiceFSForMigration mounts the freshly formatted JuiceFS filesystem on a
// temporary mount point so its content can be populated from the local rootfs.
// It is idempotent: if the temp mount point is already mounted it does nothing.
type MountJuiceFSForMigration struct {
	common.KubeAction
}

func (t *MountJuiceFSForMigration) Execute(runtime connector.Runtime) error {
	mounted, err := isPathMounted(runtime, MigrationTempMountPoint)
	if err != nil {
		return err
	}
	if mounted {
		logger.Infof("%s is already mounted, skipping", MigrationTempMountPoint)
		return nil
	}

	redisAddress, _ := t.PipelineCache.GetMustString(common.CacheHostRedisAddress)
	redisPassword, _ := t.PipelineCache.GetMustString(common.CacheHostRedisPassword)
	if redisAddress == "" || redisPassword == "" {
		return errors.New("redis config is not available, cannot mount JuiceFS for migration")
	}
	metaURL := fmt.Sprintf("redis://:%s@%s:6379/1", redisPassword, redisAddress)

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", MigrationTempMountPoint), false, false); err != nil {
		return err
	}
	cmd := fmt.Sprintf("%s mount --background --cache-dir %s %s %s", JuiceFsFile, JuiceFsCacheDir, metaURL, MigrationTempMountPoint)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return errors.Wrap(err, "failed to mount JuiceFS for migration")
	}
	return nil
}

// SyncRootFSData copies the content of the local rootfs into the JuiceFS
// filesystem mounted at the migration temp mount point.
//
// It is meant to be run twice:
//   - first (Delete=false) while Olares is still running, to copy the bulk of
//     the data without downtime;
//   - then (Delete=true) after the workloads have been stopped, to copy the
//     incremental changes and remove anything deleted in the meantime.
type SyncRootFSData struct {
	common.KubeAction
	Delete bool
}

func (t *SyncRootFSData) Execute(runtime connector.Runtime) error {
	// Safety: the source must be the real local rootfs, never a JuiceFS mount.
	// Otherwise a --delete sync from an (empty) JuiceFS into JuiceFS could wipe data.
	srcMounted, err := isJuiceFSMounted(runtime, OlaresJuiceFSRootDir)
	if err != nil {
		return err
	}
	if srcMounted {
		return fmt.Errorf("refusing to sync: source %s is already a JuiceFS mount", OlaresJuiceFSRootDir)
	}
	dstMounted, err := isPathMounted(runtime, MigrationTempMountPoint)
	if err != nil {
		return err
	}
	if !dstMounted {
		return fmt.Errorf("refusing to sync: destination %s is not mounted", MigrationTempMountPoint)
	}

	flags := "-aHAX --numeric-ids"
	// JuiceFS exposes internal control files at the mount root (.accesslog,
	// .config, .stats, .trash, .control). They exist only on the destination
	// (the JuiceFS mount), never in the local source, so a --delete pass would
	// try to remove them and fail with "Operation not permitted". Exclude them
	// (an excluded file is also protected from --delete by default).
	flags += " --exclude=/.accesslog --exclude=/.config --exclude=/.stats --exclude=/.trash --exclude=/.control"
	if t.Delete {
		flags += " --delete"
	}
	// trailing slashes: copy the contents of the source into the destination
	cmd := fmt.Sprintf("rsync %s %s/ %s/", flags, OlaresJuiceFSRootDir, MigrationTempMountPoint)
	if _, err := runtime.GetRunner().SudoCmd(cmd, true, true); err != nil {
		return errors.Wrap(err, "failed to sync rootfs data into JuiceFS")
	}
	return nil
}

// UnmountJuiceFSMigration unmounts and removes the temporary migration mount
// point. It is idempotent.
type UnmountJuiceFSMigration struct {
	common.KubeAction
}

func (t *UnmountJuiceFSMigration) Execute(runtime connector.Runtime) error {
	mounted, err := isPathMounted(runtime, MigrationTempMountPoint)
	if err != nil {
		return err
	}
	if mounted {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("umount %s", MigrationTempMountPoint), false, true); err != nil {
			return errors.Wrap(err, "failed to unmount the migration mount point")
		}
	}
	// best-effort cleanup of the now-empty temp dir
	_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rmdir %s", MigrationTempMountPoint), false, false)
	return nil
}

// BackupAndSwapRootFS moves the original local rootfs aside and recreates an
// empty rootfs directory that EnableJuiceFsService will then mount JuiceFS on.
// It is idempotent: if the rootfs is already a JuiceFS mount it does nothing.
type BackupAndSwapRootFS struct {
	common.KubeAction
}

func (t *BackupAndSwapRootFS) Execute(runtime connector.Runtime) error {
	if IsJuiceFSEnabled() {
		logger.Info("JuiceFS is already enabled, skipping rootfs backup and swap")
		return nil
	}
	mounted, err := isJuiceFSMounted(runtime, OlaresJuiceFSRootDir)
	if err != nil {
		return err
	}
	if mounted {
		logger.Infof("%s is already a JuiceFS mount, skipping backup and swap", OlaresJuiceFSRootDir)
		return nil
	}

	backup := RootFSLocalBackupDir
	if util.IsExist(backup) {
		backup = fmt.Sprintf("%s.%d", RootFSLocalBackupDir, time.Now().Unix())
	}
	logger.Infof("moving original rootfs %s to %s", OlaresJuiceFSRootDir, backup)
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mv %s %s", OlaresJuiceFSRootDir, backup), false, true); err != nil {
		return errors.Wrap(err, "failed to back up the original rootfs")
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", OlaresJuiceFSRootDir), false, false); err != nil {
		return errors.Wrap(err, "failed to recreate the rootfs mount point")
	}
	return nil
}

func isJuiceFSMounted(runtime connector.Runtime, target string) (bool, error) {
	out, err := runtime.GetRunner().SudoCmd("cat /proc/self/mounts", false, false)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == target && strings.HasPrefix(fields[2], "fuse.juicefs") {
			return true, nil
		}
	}
	return false, nil
}

// UpdateRootFSTypeSystemEnv flips the OLARES_SYSTEM_ROOTFS_TYPE system env from
// "fs" to "jfs" so that subsequently installed apps are rendered for a shared
// (JuiceFS) rootfs. Existing apps keep working regardless because their PVs are
// hostPath PVs whose paths are unchanged.
type UpdateRootFSTypeSystemEnv struct {
	common.KubeAction
}

func (t *UpdateRootFSTypeSystemEnv) Execute(runtime connector.Runtime) error {
	const envName = "OLARES_SYSTEM_ROOTFS_TYPE"
	const value = "jfs"

	common.SetSystemEnv(envName, value)

	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %w", err)
	}

	apix, err := apixclientset.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create crd client: %w", err)
	}

	ctx := context.Background()
	_, err = apix.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, "systemenvs.sys.bytetrade.io", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Debugf("SystemEnv CRD not found, skipping rootfs type update")
			return nil
		}
		return fmt.Errorf("failed to get SystemEnv CRD: %w", err)
	}

	scheme := kruntime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add systemenv scheme: %w", err)
	}

	c, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	resourceName, err := apputils.EnvNameToResourceName(envName)
	if err != nil {
		return fmt.Errorf("invalid system env name: %s", envName)
	}

	var existingSystemEnv v1alpha1.SystemEnv
	err = c.Get(ctx, types.NamespacedName{Name: resourceName}, &existingSystemEnv)
	if err == nil {
		if existingSystemEnv.Default != value {
			existingSystemEnv.Default = value
			if err := c.Update(ctx, &existingSystemEnv); err != nil {
				return fmt.Errorf("failed to update SystemEnv %s: %w", resourceName, err)
			}
			logger.Infof("Updated SystemEnv %s default to %s", resourceName, value)
		}
		return nil
	}

	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get SystemEnv %s: %w", resourceName, err)
	}

	systemEnv := &v1alpha1.SystemEnv{
		ObjectMeta: metav1.ObjectMeta{
			Name: resourceName,
		},
		EnvVarSpec: v1alpha1.EnvVarSpec{
			EnvName: envName,
			Default: value,
		},
	}
	if err := c.Create(ctx, systemEnv); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create SystemEnv %s: %w", resourceName, err)
	}
	logger.Infof("Created SystemEnv: %s with default %s", envName, value)
	return nil
}
