package download

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"

	cc "bytetrade.io/web3os/installer/pkg/core/common"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/utils"
)

type PackageDownload struct {
	common.KubeAction
	Manifest       string
	BaseDir        string
	DownloadCdnUrl string
	existingItems  []*manifest.ManifestItem
	missingItems   []*manifest.ManifestItem
}

type CheckDownload struct {
	PackageDownload
}

func (d *PackageDownload) Execute(runtime connector.Runtime) error {
	baseDir := d.BaseDir
	if runtime.GetSystemInfo().IsWsl() {
		var wslPackageDir = d.KubeConf.Arg.GetWslUserPath()
		if wslPackageDir != "" {
			baseDir = path.Join(wslPackageDir, cc.DefaultBaseDir)
		}
	}
	logger.Info("checking local cache ...")
	err := d.CheckLocalCache(runtime)
	if err != nil {
		return errors.Wrap(err, "failed to check local cache")
	}
	if len(d.missingItems) == 0 {
		logger.Info("all files are already downloaded and is the expected version")
	} else {
		logger.Infof("%d out of %d files need to be downloaded", len(d.missingItems), len(d.missingItems)+len(d.existingItems))
	}
	for i := range d.missingItems {
		err = d.downloadItem(runtime, baseDir, i)
		if err != nil {
			logger.Fatal(err)
		}
	}

	return nil
}

// CheckLocalCache compares the items in the manifest file
// against the local files by MD5 checksum
// and filters out the existing and missing items
func (d *PackageDownload) CheckLocalCache(runtime connector.Runtime) error {
	if d.Manifest == "" {
		return errors.New("manifest path is empty")
	}
	baseDir := d.BaseDir

	if runtime.GetSystemInfo().IsWsl() {
		var wslPackageDir = d.KubeConf.Arg.GetWslUserPath()
		if wslPackageDir != "" {
			baseDir = path.Join(wslPackageDir, cc.DefaultBaseDir)
		}
	}

	f, err := os.Open(d.Manifest)
	if err != nil {
		return errors.Wrap(err, "unable to open manifest")
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		item := must(manifest.ReadItem(line))
		if must(isRealExists(runtime, item, baseDir)) {
			d.existingItems = append(d.existingItems, item)
		} else {
			d.missingItems = append(d.missingItems, item)
		}
	}
	return nil
}

func (d *CheckDownload) Execute(runtime connector.Runtime) error {
	if err := d.CheckLocalCache(runtime); err != nil {
		return err
	}

	if len(d.missingItems) > 0 {
		logger.Fatalf("found %d missing items", len(d.missingItems))
	}

	logger.Info("suceess to check download")
	return nil

}

// if the file exists and the checksum passed
func isRealExists(runtime connector.Runtime, item *manifest.ManifestItem, baseDir string) (bool, error) {
	arch := runtime.GetSystemInfo().GetOsArch()
	targetPath := getDownloadTargetPath(item, baseDir)
	exists, err := runtime.GetRunner().FileExist(targetPath)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	checksum := utils.LocalMd5Sum(targetPath)
	// FIXME: run in remote
	return checksum == item.GetItemUrlForHost(arch).Checksum, nil
}

func (d *PackageDownload) downloadItem(runtime connector.Runtime, baseDir string, index int) error {
	arch := runtime.GetSystemInfo().GetOsArch()
	os := runtime.GetSystemInfo().GetOsType()
	item := d.missingItems[index]
	url := item.GetItemUrlForHost(arch)

	component := new(files.KubeBinary)
	component.ID = item.Filename
	component.Arch = runtime.GetSystemInfo().GetOsArch()
	component.BaseDir = getDownloadTargetBasePath(item, baseDir)
	component.Url = fmt.Sprintf("%s/%s", d.DownloadCdnUrl, strings.TrimPrefix(url.Url, "/"))
	component.FileName = item.Filename
	component.CheckMd5Sum = true
	component.Md5sum = url.Checksum
	component.Os = os

	downloadPath := component.Path()
	if utils.IsExist(downloadPath) {
		_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", downloadPath), false, false)
	}

	if !utils.IsExist(component.BaseDir) {
		if err := component.CreateBaseDir(); err != nil {
			return err
		}
	}

	logger.Infof("(%d/%d) downloading %s %s, file: %s",
		index+1, len(d.missingItems),
		friendlyItemType(item.Type),
		item.FileID,
		item.Filename)

	if err := component.Download(); err != nil {
		return fmt.Errorf("Failed to download %s binary: %s error: %w ", component.ID, component.Url, err)
	}

	return nil
}

// friendlyItemType is a simple translation from
// path-based item type to a friendly type that's more readable by user
// "image.*" is simplified to "image"  and all other types to "package"
// to shadow the details.
// should Only be used for output to console
func friendlyItemType(itemType string) string {
	if strings.Contains(itemType, "image") {
		return "image"
	}
	return "package"
}

func getDownloadTargetPath(item *manifest.ManifestItem, baseDir string) string {
	return fmt.Sprintf("%s/%s/%s", baseDir, item.Path, item.Filename)
}

func getDownloadTargetBasePath(item *manifest.ManifestItem, baseDir string) string {
	return fmt.Sprintf("%s/%s", baseDir, item.Path)
}

func must[T any](r T, e error) T {
	if e != nil {
		logger.Fatal(e)
	}

	return r
}
