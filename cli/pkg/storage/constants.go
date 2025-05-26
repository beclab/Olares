package storage

import (
	"bytetrade.io/web3os/installer/pkg/storage/templates"
	"path"

	"bytetrade.io/web3os/installer/pkg/core/common"
)

var (
	Root                 = path.Join("/")
	StorageDataDir       = path.Join(Root, "osdata")
	StorageDataOlaresDir = path.Join(StorageDataDir, common.OlaresDir)
	OlaresRootDir        = path.Join(Root, common.OlaresDir)
	OlaresSharedLibDir   = path.Join(OlaresRootDir, "share")
	OlaresUserDataDir    = path.Join(OlaresRootDir, "userdata")

	RedisRootDir             = path.Join(OlaresRootDir, "data", "redis")
	RedisConfigDir           = path.Join(RedisRootDir, "etc")
	RedisDataDir             = path.Join(RedisRootDir, "data")
	RedisLogDir              = path.Join(RedisRootDir, "log")
	RedisRunDir              = path.Join(RedisRootDir, "run")
	RedisConfigFile          = path.Join(RedisConfigDir, "redis.conf")
	RedisServiceFile         = path.Join(Root, "etc", "systemd", "system", "redis-server.service")
	RedisServerFile          = path.Join(Root, "usr", "bin", "redis-server")
	RedisCliFile             = path.Join(Root, "usr", "bin", "redis-cli")
	RedisServerInstalledFile = path.Join(Root, "usr", "local", "bin", "redis-server")
	RedisCliInstalledFile    = path.Join(Root, "usr", "local", "bin", "redis-cli")

	JuiceFsFile          = path.Join(Root, "usr", "local", "bin", "juicefs")
	JuiceFsDataDir       = path.Join(OlaresRootDir, "data", "juicefs")
	JuiceFsCacheDir      = path.Join(OlaresRootDir, "jfscache")
	OlaresJuiceFSRootDir = path.Join(OlaresRootDir, "rootfs")
	JuiceFsServiceFile   = path.Join(Root, "etc", "systemd", "system", templates.JuicefsService.Name())

	MinioRootUser    = "minioadmin"
	MinioDataDir     = path.Join(OlaresRootDir, "data", "minio", "vol1")
	MinioFile        = path.Join(Root, "usr", "local", "bin", "minio")
	MinioServiceFile = path.Join(Root, "etc", "systemd", "system", "minio.service")
	MinioConfigFile  = path.Join(Root, "etc", "default", "minio")

	MinioOperatorFile = path.Join(Root, "usr", "local", "bin", "minio-operator")
)
