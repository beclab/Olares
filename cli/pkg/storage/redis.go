package storage

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/logger"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/utils"

	redisTemplates "bytetrade.io/web3os/installer/pkg/storage/templates"
	"github.com/pkg/errors"
)

type CheckRedisServiceState struct {
	common.KubeAction
}

func (t *CheckRedisServiceState) Execute(runtime connector.Runtime) error {
	var systemInfo = runtime.GetSystemInfo()
	var localIp = systemInfo.GetLocalIp()
	var rpwd, _ = t.PipelineCache.GetMustString(common.CacheHostRedisPassword)
	var cmd = fmt.Sprintf("%s -h %s -a %s ping", RedisCliFile, localIp, rpwd)
	if pong, _ := runtime.GetRunner().SudoCmd(cmd, false, false); !strings.Contains(pong, "PONG") {
		return fmt.Errorf("failed to connect redis server: %s:6379", localIp)
	}

	return nil
}

type EnableRedisService struct {
	common.KubeAction
}

func (t *EnableRedisService) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("sysctl -w vm.overcommit_memory=1 net.core.somaxconn=10240", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl restart redis-server", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl enable redis-server", false, false); err != nil {
		return err
	}

	var cmd = "( sleep 10 && systemctl --no-pager status redis-server ) || ( systemctl restart redis-server && sleep 3 && systemctl --no-pager status redis-server ) || ( systemctl restart redis-server && sleep 3 && systemctl --no-pager status redis-server )"
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return err
	}

	return nil
}

type GetOrSetRedisConfig struct {
	common.KubeAction
}

func (t *GetOrSetRedisConfig) Execute(runtime connector.Runtime) (err error) {
	var redisPassword string
	redisAddress := runtime.RemoteHost().GetInternalAddress()
	defer func() {
		if err == nil {
			t.PipelineCache.Set(common.CacheHostRedisPassword, redisPassword)
			t.PipelineCache.Set(common.CacheHostRedisAddress, redisAddress)
		}
	}()
	exist, err := runtime.GetRunner().FileExist(RedisConfigFile)
	if err != nil {
		return err
	}
	if !exist {
		redisPassword, _ = utils.GeneratePassword(16)
		return
	}
	redisPassword, err = getRedisPwdFromConfigFile(runtime)
	if err != nil {
		return
	}
	if redisPassword != "" {
		logger.Debugf("using existing Redis password: %s found in %s", redisPassword, RedisConfigFile)
		return
	}
	logger.Warnf("found Redis config file %s but password is not set, generating a new one", RedisConfigFile)
	redisPassword, _ = utils.GeneratePassword(16)
	return
}

func getRedisPwdFromConfigFile(runtime connector.Runtime) (string, error) {
	var cmd = fmt.Sprintf("cat %s 2>&1 |grep requirepass |cut -d' ' -f2 |tr -d '\n'", RedisConfigFile)
	if res, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return "", errors.Wrap(err, "failed to get redis password")
	} else {
		return res, nil
	}
}

type GenerateRedisService struct {
	common.KubeAction
}

func (t *GenerateRedisService) Execute(runtime connector.Runtime) error {
	redisPassword, ok := t.PipelineCache.GetMustString(common.CacheHostRedisPassword)
	if !ok || redisPassword == "" {
		return errors.New("redis password not set")
	}
	var systemInfo = runtime.GetSystemInfo()
	var localIp = systemInfo.GetLocalIp()
	if !utils.IsExist(RedisRootDir) {
		utils.Mkdir(RedisConfigDir)
		utils.Mkdir(RedisDataDir)
		utils.Mkdir(RedisLogDir)
		utils.Mkdir(RedisRunDir)
	}

	var data = util.Data{
		"LocalIP":  localIp,
		"RootPath": RedisRootDir,
		"Password": redisPassword,
	}
	redisConfStr, err := util.Render(redisTemplates.RedisConf, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render redis conf template failed")
	}
	if err := util.WriteFile(RedisConfigFile, []byte(redisConfStr), cc.FileMode0640); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write redis conf %s failed", RedisConfigFile))
	}

	data = util.Data{
		"RedisBinPath":  RedisServerFile,
		"RootPath":      RedisRootDir,
		"RedisConfPath": RedisConfigFile,
	}
	redisServiceStr, err := util.Render(redisTemplates.RedisService, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render redis service template failed")
	}
	if err := util.WriteFile(RedisServiceFile, []byte(redisServiceStr), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write redis service %s failed", RedisServiceFile))
	}

	return nil
}

type CheckRedisExists struct {
	common.KubePrepare
}

func (p *CheckRedisExists) PreCheck(runtime connector.Runtime) (bool, error) {
	if !utils.IsExist(RedisServerInstalledFile) {
		return true, nil
	}

	if !utils.IsExist(RedisServerFile) {
		return true, nil
	}

	return false, nil
}

type InstallRedis struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *InstallRedis) Execute(runtime connector.Runtime) error {
	compName := "redis"
	version, err := t.getGlibcVersion(runtime)
	if err != nil {
		logger.Warnf("get glibc version error, %v", err)
	} else {
		if version.lessThan("2.32") {
			compName = "redis-231"
		}
	}

	redis, err := t.Manifest.Get(compName)
	if err != nil {
		return err
	}

	path := redis.FilePath(t.BaseDir)

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf /tmp/redis-* && cp -f %s /tmp/%s && cd /tmp && tar xf ./%s", path, redis.Filename, redis.Filename), false, false); err != nil {
		return errors.Wrapf(errors.WithStack(err), "untar redis failed")
	}

	unpackPath := strings.TrimSuffix(redis.Filename, ".tar.gz")
	var cmd = fmt.Sprintf("cd /tmp/%s && cp ./* /usr/local/bin/ && rm -rf ./%s",
		unpackPath, unpackPath)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return err
	}
	// if _, err := runtime.GetRunner().SudoCmd("[[ ! -f /usr/local/bin/redis-sentinel ]] && /usr/local/bin/redis-server /usr/local/bin/redis-sentinel || true", false, true); err != nil {
	// 	return err
	// }
	if exist, err := runtime.GetRunner().FileExist(RedisServerFile); err != nil {
		return err
	} else if !exist {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("ln -s %s %s", RedisServerInstalledFile, RedisServerFile), false, true); err != nil {
			return err
		}
	}
	if exist, err := runtime.GetRunner().FileExist(RedisCliFile); err != nil {
		return err
	} else if !exist {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("ln -s %s %s", RedisCliInstalledFile, RedisCliFile), false, true); err != nil {
			return err
		}
	}

	return nil
}

func (t *InstallRedis) getGlibcVersion(runtime connector.Runtime) (GlibcVersion, error) {
	output, err := runtime.GetRunner().SudoCmd("ldd --version", false, false)
	if err != nil {
		return "", err
	}

	// Convert the output to string
	outputStr := string(output)

	// Find the line containing the glibc version
	lines := strings.Split(outputStr, "\n")

	for _, line := range lines {
		if strings.Contains(line, "GLIBC") {
			lineToken := strings.Split(strings.TrimSpace(line), " ")

			version := lineToken[len(lineToken)-1]

			return GlibcVersion(version), nil
		}
	}

	return "", errors.New("glibc version not found")
}

type GlibcVersion string

func (v GlibcVersion) lessThan(v2 GlibcVersion) bool {
	return compareGLibC(string(v), string(v2)) < 0
}

func compareGLibC(versionA, versionB string) int {
	partsA := strings.Split(versionA, ".")
	partsB := strings.Split(versionB, ".")

	// Convert major and minor versions to integers
	majorA, _ := strconv.Atoi(partsA[0])
	minorA, _ := strconv.Atoi(partsA[1])

	majorB, _ := strconv.Atoi(partsB[0])
	minorB, _ := strconv.Atoi(partsB[1])

	if majorA < majorB {
		return -1
	} else if majorA > majorB {
		return 1
	} else {
		if minorA < minorB {
			return -1
		} else if minorA > minorB {
			return 1
		}
	}

	return 0 // versions are equal
}

type InstallRedisModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool
}

func (m *InstallRedisModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallRedisModule) Init() {
	m.Name = "InstallRedis"

	installRedis := &task.RemoteTask{
		Name:    "Install",
		Hosts:   m.Runtime.GetAllHosts(),
		Prepare: &CheckRedisExists{},
		Action: &InstallRedis{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Parallel: false,
		Retry:    0,
	}

	getOrSetRedisPassword := &task.LocalTask{
		Name:   "GetOrSetRedisConfig",
		Action: new(GetOrSetRedisConfig),
		Retry:  1,
	}

	configRedis := &task.RemoteTask{
		Name:     "Config",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(GenerateRedisService),
		Parallel: false,
		Retry:    0,
	}

	enableRedisService := &task.RemoteTask{
		Name:     "EnableRedisService",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(EnableRedisService),
		Parallel: false,
		Retry:    0,
	}

	checkRedisServiceState := &task.RemoteTask{
		Name:     "CheckState",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(CheckRedisServiceState),
		Parallel: false,
		Retry:    3,
		Delay:    3 * time.Second,
	}

	m.Tasks = []task.Interface{
		installRedis,
		getOrSetRedisPassword,
		configRedis,
		enableRedisService,
		checkRedisServiceState,
	}
}
