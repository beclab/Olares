package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/util"
	juicefsTemplates "github.com/beclab/Olares/cli/pkg/storage/templates"

	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
)

// - InstallJuiceFsModule
type InstallJuiceFsModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool
}

func (m *InstallJuiceFsModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallJuiceFsModule) Init() {
	m.Name = "InstallJuiceFs"

	getRedisConfig := &task.RemoteTask{
		Name:   "GetRedisConfig",
		Hosts:  m.Runtime.GetHostsByRole(common.Master),
		Action: new(GetOrSetRedisConfig),
		Retry:  1,
	}

	installJuiceFs := &task.LocalTask{
		Name:    "InstallJuiceFs",
		Prepare: &CheckJuiceFsExists{},
		Action: &InstallJuiceFs{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Retry: 1,
	}

	configJuiceFsMetaDB := &task.LocalTask{
		Name:   "ConfigJuiceFSMetaDB",
		Action: new(ConfigJuiceFsMetaDB),
		Retry:  1,
	}

	enableJuiceFsService := &task.LocalTask{
		Name:   "EnableJuiceFsService",
		Action: new(EnableJuiceFsService),
		Retry:  1,
	}

	checkJuiceFsState := &task.LocalTask{
		Name:   "CheckJuiceFsState",
		Action: new(CheckJuiceFsState),
		Retry:  5,
		Delay:  5 * time.Second,
	}

	m.Tasks = []task.Interface{
		getRedisConfig,
		installJuiceFs,
		configJuiceFsMetaDB,
		enableJuiceFsService,
		checkJuiceFsState,
	}
}

type CheckJuiceFsExists struct {
	common.KubePrepare
}

func (p *CheckJuiceFsExists) PreCheck(runtime connector.Runtime) (bool, error) {
	if utils.IsExist(JuiceFsFile) {
		return false, nil
	}

	return true, nil
}

type InstallJuiceFs struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *InstallJuiceFs) Execute(runtime connector.Runtime) error {
	if !utils.IsExist(JuiceFsDataDir) {
		err := utils.Mkdir(JuiceFsDataDir)
		if err != nil {
			return err
		}
	}

	juicefs, err := t.Manifest.Get("juicefs")
	if err != nil {
		return err
	}

	path := juicefs.FilePath(t.BaseDir)

	var cmd = fmt.Sprintf("rm -rf /tmp/juicefs* && cp -f %s /tmp/%s && cd /tmp && tar -zxf ./%s && chmod +x juicefs && install juicefs /usr/local/bin && install juicefs /sbin/mount.juicefs && rm -rf ./LICENSE ./README.md ./README_CN.md ./juicefs*", path, juicefs.Filename, juicefs.Filename)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return err
	}
	return nil
}

type EnableJuiceFsService struct {
	common.KubeAction
}

func (t *EnableJuiceFsService) Execute(runtime connector.Runtime) error {
	var redisAddress, _ = t.PipelineCache.GetMustString(common.CacheHostRedisAddress)
	var redisPassword, _ = t.PipelineCache.GetMustString(common.CacheHostRedisPassword)
	if redisAddress == "" || redisPassword == "" {
		exist, err := runtime.GetRunner().FileExist(JuiceFsServiceFile)
		if err != nil {
			return err
		}
		if !exist {
			return errors.New("no redis config is available")
		}
	} else {
		redisService := fmt.Sprintf("redis://:%s@%s:6379/1", redisPassword, redisAddress)
		data := util.Data{
			"JuiceFsBinPath":    JuiceFsFile,
			"JuiceFsCachePath":  JuiceFsCacheDir,
			"JuiceFsMetaDb":     redisService,
			"JuiceFsMountPoint": OlaresJuiceFSRootDir,
		}

		juiceFsServiceStr, err := util.Render(juicefsTemplates.JuicefsService, data)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "render juicefs service template failed")
		}
		if err := util.WriteFile(JuiceFsServiceFile, []byte(juiceFsServiceStr), cc.FileMode0644); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write juicefs service %s failed", JuiceFsServiceFile))
		}
	}

	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("systemctl restart juicefs", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("systemctl enable juicefs", false, false); err != nil {
		return err
	}

	if err := advanceMigratePhase(phaseEnabled); err != nil {
		return errors.Wrap(err, "failed to record migration phase after enabling the juicefs service")
	}

	return nil
}

type ConfigJuiceFsMetaDB struct {
	common.KubeAction
}

func (t *ConfigJuiceFsMetaDB) Execute(runtime connector.Runtime) error {
	exist, err := runtime.GetRunner().FileExist(RedisServiceFile)
	if err != nil {
		return err
	}
	// on a worker node, no need to config
	if !exist {
		return nil
	}
	var systemInfo = runtime.GetSystemInfo()
	var localIp = systemInfo.GetLocalIp()
	var redisAddress, _ = t.PipelineCache.GetMustString(common.CacheHostRedisAddress)
	var redisPassword, _ = t.PipelineCache.GetMustString(common.CacheHostRedisPassword)
	var redisService = fmt.Sprintf("redis://:%s@%s:6379/1", redisPassword, redisAddress)
	if redisPassword == "" {
		return fmt.Errorf("redis password not found")
	}

	storageFlags, err := getStorageFlags(t.KubeConf.Arg.Storage, localIp)
	if err != nil {
		return err
	}
	var cmd = fmt.Sprintf("%s format %s", JuiceFsFile, redisService)
	cmd = cmd + storageFlags

	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return err
	}
	return nil
}

// CleanupStaleJuiceFSMount detaches a dead JuiceFS FUSE mount left behind at
// the rootfs mount point, so that the next mount attempt can succeed.
//
// On shutdown the juicefs client gives its own umount 30 seconds and then
// force-exits ("umount does not finish in 30 seconds, force exit"), leaving the
// fuse.juicefs entry in /proc/self/mounts with no process serving it. systemd
// still reports the unit as cleanly deactivated. Every later access to the path
// then fails with ENOTCONN, and a fresh `juicefs mount` cannot recover on its
// own: fusermount3 has to stat the mount point first and gets ENOTCONN too, so
// juicefs.service crash-loops indefinitely. k3s never starts either, because
// its ExecStartPre runs `juicefs summary` on the same path.
//
// A lazy umount detaches the dead mount immediately while letting any remaining
// references drain, which is exactly what is needed before remounting. This is
// a no-op when the mount point is absent, or when juicefs.service is running
// (so an already-started Olares is never disturbed).
type CleanupStaleJuiceFSMount struct {
	common.KubeAction
}

func (t *CleanupStaleJuiceFSMount) Execute(runtime connector.Runtime) error {
	mounted, err := isJuiceFSMounted(runtime, OlaresJuiceFSRootDir)
	if err != nil {
		return errors.Wrap(err, "failed to read the mount table")
	}
	if !mounted {
		return nil
	}
	// Liveness is decided by the unit state, not by touching the path: the
	// client mounts with --attr-cache 300, so the kernel keeps answering stat()
	// on the mount root from its attribute cache for minutes after the serving
	// process is gone. A mount whose unit is not active has nothing behind it.
	if out, _ := runtime.GetRunner().SudoCmd("systemctl is-active juicefs", false, false); strings.TrimSpace(out) == "active" {
		return nil
	}
	logger.Infof("detaching stale JuiceFS mount at %s (juicefs.service is not active)", OlaresJuiceFSRootDir)
	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("umount -l %s", OlaresJuiceFSRootDir), false, false); err != nil {
		return errors.Wrapf(err, "failed to detach the stale JuiceFS mount at %s", OlaresJuiceFSRootDir)
	}
	return nil
}

type CheckJuiceFsState struct {
	common.KubeAction
}

func (t *CheckJuiceFsState) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl --no-pager -n 0 status juicefs", false, false); err != nil {
		return fmt.Errorf("JuiceFs Pending")
	}

	time.Sleep(5 * time.Second)

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s summary %s", JuiceFsFile, OlaresJuiceFSRootDir), false, false); err != nil {
		return err
	}

	return nil
}

func getStorageFlags(storage *common.Storage, localIp string) (string, error) {
	var storageFlags string
	var fsName string
	var err error

	switch storage.StorageType {
	case common.ManagedMinIO:
		storageFlags, err = getManagedMinIOAccessFlags(localIp)
		if err != nil {
			return "", err
		}
	default:
		storageFlags = getExternalStorageAccessFlags(storage)
	}

	if storage.StorageVendor == "true" && storage.StorageClusterId != "" {
		fsName = storage.StorageClusterId
	} else {
		fsName = "rootfs"
	}

	storageFlags = storageFlags + fmt.Sprintf(" %s --trash-days 0", fsName)

	return storageFlags, nil
}

func getExternalStorageAccessFlags(storage *common.Storage) string {
	var params = fmt.Sprintf(" --storage %s --bucket %s", storage.StorageType, storage.StorageBucket)
	if storage.StorageVendor == "true" {
		if storage.StorageToken != "" {
			params = params + fmt.Sprintf(" --session-token %s", storage.StorageToken)
		}
	}
	if storage.StorageAccessKey != "" {
		params = params + fmt.Sprintf(" --access-key %s", storage.StorageAccessKey)
	}
	if storage.StorageSecretKey != "" {
		params = params + fmt.Sprintf(" --secret-key %s", storage.StorageSecretKey)
	}

	return params
}

func getManagedMinIOAccessFlags(localIp string) (string, error) {
	minioPassword, err := getMinioPwdFromConfigFile()
	if err != nil {
		return "", errors.Wrap(err, "failed to get password of managed MinIO")
	}
	return fmt.Sprintf(" --storage minio --bucket http://%s:9000/%s --access-key %s --secret-key %s",
		localIp, cc.OlaresDir, MinioRootUser, minioPassword), nil
}

func GetRootFSType() string {
	if util.IsExist(JuiceFsServiceFile) {
		return "jfs"
	}
	return "fs"
}

func init() {
	common.SetSystemEnv("OLARES_SYSTEM_ROOTFS_TYPE", GetRootFSType())
}
