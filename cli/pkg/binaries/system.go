package binaries

import (
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/pkg/errors"
)

func GetUbutun24AppArmor(basePath string, manifestMap manifest.InstallationManifest) (string, error) {
	apparmor, err := manifestMap.Get("apparmor")
	if err != nil {
		return "", err
	}

	path := apparmor.FilePath(basePath)
	if !util.IsExist(path) {
		return "", errors.Errorf("apparmor not found in %s", path)
	}

	return path, nil

}
