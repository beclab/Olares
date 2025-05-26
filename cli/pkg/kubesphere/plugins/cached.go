package plugins

import (
	"path"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type CachedBuilder struct {
	common.KubeAction
}

func (t *CachedBuilder) Execute(runtime connector.Runtime) error {
	// cachedDir := path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.ManifestDir)
	cachedDir := path.Join(runtime.GetBaseDir(), cc.ManifestDir)
	if !util.IsExist(cachedDir) {
		util.Mkdir(cachedDir)
	}

	// cachedImageDir := path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.ImageCacheDir)
	cachedImageDir := path.Join(runtime.GetBaseDir(), cc.ImageCacheDir)
	if !util.IsExist(cachedImageDir) {
		util.Mkdir(cachedImageDir)
	}

	// cachedPkgDir := path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir)
	cachedPkgDir := path.Join(runtime.GetBaseDir(), cc.PackageCacheDir)
	if !util.IsExist(cachedPkgDir) {
		util.Mkdir(cachedPkgDir)
	}

	// cachedBuildFilesDir := path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.BuildFilesCacheDir)
	cachedBuildFilesDir := path.Join(runtime.GetBaseDir(), cc.BuildFilesCacheDir)
	if !util.IsExist(cachedBuildFilesDir) {
		util.Mkdir(cachedBuildFilesDir)
	}

	return nil
}
