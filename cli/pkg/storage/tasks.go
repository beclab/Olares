package storage

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

type MkStorageDir struct {
	common.KubeAction
}

func (t *MkStorageDir) Execute(runtime connector.Runtime) error {
	if utils.IsExist(StorageDataDir) {
		if utils.IsExist(cc.OlaresDir) {
			_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", cc.OlaresDir), false, false)
		}

		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", StorageDataOlaresDir), false, false); err != nil {
			return err
		}
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("ln -s %s %s", StorageDataOlaresDir, cc.OlaresDir), false, false); err != nil {
			return err
		}
	}

	return nil
}

type DownloadStorageCli struct {
	common.KubeAction
}

func (t *DownloadStorageCli) Execute(runtime connector.Runtime) error {
	if _, err := util.GetCommand(common.CommandUnzip); err != nil {
		runtime.GetRunner().SudoCmd("apt install -y unzip", false, true)
	}

	var storageType = t.KubeConf.Arg.Storage.StorageType
	var osType = runtime.GetSystemInfo().GetOsType()
	var osArch = runtime.GetSystemInfo().GetOsArch()
	var osVersion = runtime.GetSystemInfo().GetOsVersion()
	var osPlatformFamily = runtime.GetSystemInfo().GetOsPlatformFamily()
	var arch = fmt.Sprintf("%s-%s", osType, osArch)

	// var prePath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir)
	var prePath = path.Join(runtime.GetBaseDir(), cc.PackageCacheDir)
	var binary *files.KubeBinary
	switch storageType {
	case "s3":
		binary = files.NewKubeBinary("awscli", arch, osType, osVersion, osPlatformFamily, "", prePath, "")
	case "oss":
		binary = files.NewKubeBinary("ossutil", arch, osType, osVersion, osPlatformFamily, kubekeyapiv1alpha2.DefaultOssUtilVersion, prePath, "")
	case "cos":
		binary = files.NewKubeBinary("cosutil", arch, osType, osVersion, osPlatformFamily, kubekeyapiv1alpha2.DefaultCosUtilVersion, prePath, "")
	default:
		return nil
	}

	binaries := []*files.KubeBinary{binary}
	binariesMap := make(map[string]*files.KubeBinary)
	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		binariesMap[binary.ID] = binary
		var exists = util.IsExist(binary.Path())
		if exists {
			p := binary.Path()
			if err := binary.SHA256Check(); err != nil {
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
			} else {
				continue
			}
		}

		if !exists || binary.OverWrite {
			logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, binary.ID, binary.Version)
			if err := binary.Download(); err != nil {
				return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.Url, err)
			}
		}
	}

	t.PipelineCache.Set(common.KubeBinaries+"-"+arch, binariesMap)

	return nil
}

type UnMountS3 struct {
	common.KubeAction
}

func (t *UnMountS3) Execute(runtime connector.Runtime) error {
	// exp https://terminus-os-us-west-1.s3.us-west-1.amazonaws.com
	// s3  s3://terminus-os-us-west-1

	storageBucket := t.KubeConf.Arg.Storage.StorageBucket
	storageAccessKey, _ := t.PipelineCache.GetMustString(common.CacheAccessKey)
	storageSecretKey, _ := t.PipelineCache.GetMustString(common.CacheSecretKey)
	storageToken, _ := t.PipelineCache.GetMustString(common.CacheToken)
	storageClusterId, _ := t.PipelineCache.GetMustString(common.CacheClusterId)

	if storageAccessKey == "" || storageSecretKey == "" {
		return nil
	}

	_, a, f := strings.Cut(storageBucket, "://")
	if !f {
		logger.Errorf("get s3 bucket failed %s", storageBucket)
		return nil
	}
	sa := strings.Split(a, ".")
	if len(sa) < 2 {
		logger.Errorf("get s3 bucket failed %s", storageBucket)
		return nil
	}
	endpoint := fmt.Sprintf("s3://%s", sa[0])
	var cmd = fmt.Sprintf("AWS_ACCESS_KEY_ID=%s AWS_SECRET_ACCESS_KEY=%s AWS_SESSION_TOKEN=%s /usr/local/bin/aws s3 rm %s/%s --recursive",
		storageAccessKey, storageSecretKey, storageToken, endpoint, storageClusterId,
	)

	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		logger.Errorf("failed to unmount s3 bucket %s: %v", storageBucket, err)
	}

	return nil
}

type UnMountOSS struct {
	common.KubeAction
}

func (t *UnMountOSS) Execute(runtime connector.Runtime) error {
	storageBucket := t.KubeConf.Arg.Storage.StorageBucket
	storageAccessKey, _ := t.PipelineCache.GetMustString(common.CacheAccessKey)
	storageSecretKey, _ := t.PipelineCache.GetMustString(common.CacheSecretKey)
	storageToken, _ := t.PipelineCache.GetMustString(common.CacheToken)
	storageClusterId, _ := t.PipelineCache.GetMustString(common.CacheClusterId)

	if storageAccessKey == "" || storageSecretKey == "" {
		return nil
	}

	// exp: https://name.area.aliyuncs.com
	// oss  oss://name
	// endpoint: https://area.aliyuncs.com

	b, a, f := strings.Cut(storageBucket, "://")
	if !f {
		logger.Errorf("get oss bucket failed %s", storageBucket)
		return nil
	}

	s := strings.Split(a, ".")
	if len(s) != 4 {
		logger.Errorf("get oss bucket failed %s", storageBucket)
		return nil
	}
	ossName := fmt.Sprintf("oss://%s", s[0])
	ossEndpoint := fmt.Sprintf("%s://%s.%s.%s", b, s[1], s[2], s[3])

	var cmd = fmt.Sprintf("/usr/local/sbin/ossutil64 rm %s/%s/ --endpoint=%s --access-key-id=%s --access-key-secret=%s --sts-token=%s -r -f", ossName, storageClusterId, ossEndpoint, storageAccessKey, storageSecretKey, storageToken)

	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		logger.Errorf("failed to unmount oss bucket %s: %v", storageBucket, err)
	}

	return nil
}

type UnMountCOS struct {
	common.KubeAction
}

func (t *UnMountCOS) Execute(runtime connector.Runtime) error {
	storageBucket := t.KubeConf.Arg.Storage.StorageBucket
	storageAccessKey, _ := t.PipelineCache.GetMustString(common.CacheAccessKey)
	storageSecretKey, _ := t.PipelineCache.GetMustString(common.CacheSecretKey)
	storageToken, _ := t.PipelineCache.GetMustString(common.CacheToken)
	storageClusterId, _ := t.PipelineCache.GetMustString(common.CacheClusterId)

	if storageAccessKey == "" || storageSecretKey == "" {
		return nil
	}

	_, a, f := strings.Cut(storageBucket, "://")
	if !f {
		logger.Errorf("get cos bucket failed %s", storageBucket)
		return nil
	}

	s := strings.Split(a, ".")
	if len(s) != 5 {
		logger.Errorf("get cos bucket failed %s", storageBucket)
		return nil
	}
	cosName := fmt.Sprintf("cos://%s", s[0])
	cosEndpoint := fmt.Sprintf("%s.%s.%s.%s", s[1], s[2], s[3], s[4])
	var cmd = fmt.Sprintf("/usr/local/bin/cosutil rm %s/%s/ --endpoint %s --secret-id %s --secret-key %s --token %s --init-skip -r -f", cosName, storageClusterId, cosEndpoint, storageAccessKey, storageSecretKey, storageToken)

	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		logger.Errorf("failed to unmount cos bucket %s: %v", storageBucket, err)
	}

	return nil
}

type StopJuiceFS struct {
	common.KubeAction
}

func (t *StopJuiceFS) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmd("systemctl stop juicefs; systemctl disable juicefs", false, false)

	_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf /var/jfsCache %s", JuiceFsCacheDir), false, false)

	_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("umount %s", OlaresJuiceFSRootDir), false, false)

	_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", OlaresJuiceFSRootDir), false, false)

	return nil
}

type StopMinio struct {
	common.KubeAction
}

func (t *StopMinio) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmd("systemctl stop minio; systemctl disable minio", false, false)
	return nil
}

type StopMinioOperator struct {
	common.KubeAction
}

func (t *StopMinioOperator) Execute(runtime connector.Runtime) error {
	var cmd = "systemctl stop minio-operator; systemctl disable minio-operator"
	_, _ = runtime.GetRunner().SudoCmd(cmd, false, false)
	return nil
}

type StopRedis struct {
	common.KubeAction
}

func (t *StopRedis) Execute(runtime connector.Runtime) error {
	var cmd = "systemctl stop redis-server; systemctl disable redis-server"
	_, _ = runtime.GetRunner().SudoCmd(cmd, false, false)
	_, _ = runtime.GetRunner().SudoCmd("killall -9 redis-server", false, false)
	_, _ = runtime.GetRunner().SudoCmd("unlink /usr/bin/redis-server; unlink /usr/bin/redis-cli", false, false)

	return nil
}

type RemoveJuiceFSFiles struct {
	common.KubeAction
}

func (t *RemoveJuiceFSFiles) Execute(runtime connector.Runtime) error {
	var files = []string{
		"/usr/local/bin/redis-*",
		"/usr/bin/redis-*",
		"/sbin/mount.juicefs",
		"/etc/init.d/redis-server",
		"/usr/local/bin/juicefs",
		"/etc/systemd/system/redis-server.service",
		"/etc/systemd/system/juicefs.service",
	}

	for _, f := range files {
		runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", f), false, true)
	}

	return nil
}

type RemoveTerminusFiles struct {
	common.KubeAction
}

func (t *RemoveTerminusFiles) Execute(runtime connector.Runtime) error {
	var files = []string{
		"/usr/local/bin/minio",
		"/usr/local/bin/velero",
		"/etc/systemd/system/minio.service",
		"/etc/systemd/system/minio-operator.service",
	}

	for _, f := range files {
		runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", f), false, true)
	}

	return nil
}

type DeleteTmp struct {
	common.KubeAction
}

func (t *DeleteTmp) Execute(runtime connector.Runtime) error {
	var tmpPath = path.Join(common.RootDir, "tmp", "install_log")
	if util.IsExist(tmpPath) {
		util.RemoveDir(tmpPath)
	}
	return nil
}

type DeletePhaseFlagFile struct {
	common.KubeAction
	PhaseFile string
	BaseDir   string
}

func (t *DeletePhaseFlagFile) Execute(runtime connector.Runtime) error {
	phaseFileName := path.Join(t.BaseDir, t.PhaseFile)

	if util.IsExist(phaseFileName) {
		util.RemoveFile(phaseFileName)
	}

	return nil
}

type DeleteCaches struct {
	common.KubeAction
	BaseDir string
}

func (t *DeleteCaches) Execute(runtime connector.Runtime) error {
	var cachesDirs []string

	filepath.WalkDir(t.BaseDir, func(path string, d fs.DirEntry, err error) error {
		if path != t.BaseDir {
			if d.IsDir() {
				cachesDirs = append(cachesDirs, path)
				return filepath.SkipDir
			}
		}
		return nil
	},
	)

	if cachesDirs != nil && len(cachesDirs) > 0 {
		for _, cachesDir := range cachesDirs {
			if util.IsExist(cachesDir) {
				if err := util.RemoveDir(cachesDir); err != nil {
					logger.Errorf("remove %s failed %v", cachesDir, err)
				}
			}
		}
	}

	return nil
}

type DeleteTerminusUserData struct {
	common.KubeAction
}

func (t *DeleteTerminusUserData) Execute(runtime connector.Runtime) error {
	userdataDirs := []string{
		OlaresUserDataDir,
		JuiceFsCacheDir,
	}

	if util.IsExist(RedisServiceFile) {
		userdataDirs = append(userdataDirs, OlaresJuiceFSRootDir)
	}

	for _, d := range userdataDirs {
		if util.IsExist(d) {
			if err := util.RemoveDir(d); err != nil {
				logger.Errorf("remove %s failed %v", d, err)
			}
		}
	}

	return nil
}

type DeleteTerminusData struct {
	common.KubeAction
}

func (t *DeleteTerminusData) Execute(runtime connector.Runtime) error {
	var dirs []string
	var shareExists bool
	filepath.WalkDir(OlaresRootDir, func(path string, d fs.DirEntry, err error) error {
		if path != OlaresRootDir {
			if !d.IsDir() {
				return nil
			}

			if strings.HasPrefix(path, OlaresSharedLibDir) {
				shareExists = true
			} else {
				dirs = append(dirs, path)
				return filepath.SkipDir
			}
		}

		return nil
	},
	)

	for _, dir := range dirs {
		if err := util.RemoveDir(dir); err != nil {
			logger.Errorf("remove %s failed %v", dir, err)
		}
	}

	if !shareExists {
		if err := util.RemoveDir(OlaresRootDir); err != nil {
			logger.Errorf("remove %s failed %v", OlaresRootDir, err)
		}
	}

	if util.IsExist(StorageDataDir) {
		runtime.GetRunner().SudoCmd(fmt.Sprintf("umount %s", StorageDataDir), false, true)
		if err := util.RemoveDir(StorageDataDir); err != nil {
			logger.Errorf("remove %s failed %v", StorageDataDir, err)
		}
	}

	return nil
}
