package binaries

import (
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"github.com/pkg/errors"
)

func GetSocat(basePath string, manifestMap manifest.InstallationManifest) (string, string, error) {
	socat, err := manifestMap.Get("socat")
	if err != nil {
		return "", "", err
	}

	path := socat.FilePath(basePath)
	if !util.IsExist(path) {
		return "", "", errors.Errorf("socat not found in %s", path)
	}
	return basePath, socat.Filename, nil
}

func GetFlex(basePath string, manifestMap manifest.InstallationManifest) (string, string, error) {
	flex, err := manifestMap.Get("flex")
	if err != nil {
		return "", "", err
	}

	path := flex.FilePath(basePath)
	if !util.IsExist(path) {
		return "", "", errors.Errorf("flex not found in %s", path)
	}
	return basePath, flex.Filename, nil
}

func GetConntrack(basePath string, manifestMap manifest.InstallationManifest) (string, string, error) {
	conntrack, err := manifestMap.Get("conntrack")
	if err != nil {
		return "", "", err
	}

	path := conntrack.FilePath(basePath)
	if !util.IsExist(path) {
		return "", "", errors.Errorf("conntrack not found in %s", path)
	}
	return basePath, conntrack.Filename, nil
}

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
