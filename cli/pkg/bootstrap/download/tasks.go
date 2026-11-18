package download

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/distribution/reference"
	"github.com/pkg/errors"

	cc "github.com/beclab/Olares/cli/pkg/core/common"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/files"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/utils"
)

type PackageDownload struct {
	common.KubeAction
	Manifest      string
	BaseDir       string
	CDNService    string
	existingItems []*manifest.ManifestItem
	missingItems  []*manifest.ManifestItem
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
		logger.Info("all required packages and images are already available")
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
// or, for images, the tags already available in containerd,
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
	defer f.Close()

	var containerdImages map[string]struct{}
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
			if item.ImageName != "" {
				if containerdImages == nil {
					containerdImages = listContainerdImages(runtime)
				}
				if _, ok := containerdImages[normalizeImageRef(item.ImageName)]; ok {
					logger.Infof("image %s already exists in containerd, skipping download", item.ImageName)
					d.existingItems = append(d.existingItems, item)
					continue
				}
			}
			d.missingItems = append(d.missingItems, item)
		}
	}
	return scanner.Err()
}

// Query once, only when an image archive is missing or invalid. Downloads also
// run before containerd is installed, so an unavailable runtime is a cache miss.
func listContainerdImages(runtime connector.Runtime) map[string]struct{} {
	images := make(map[string]struct{})
	if !runtime.GetSystemInfo().IsLinux() {
		return images
	}
	output, err := runtime.GetRunner().SudoCmd(
		"command -v ctr >/dev/null 2>&1 && test -S /run/containerd/containerd.sock && ctr --connect-timeout 5s --timeout 5s -n k8s.io images list -q",
		false, false)
	if err != nil {
		logger.Debugf("unable to check containerd image cache, falling back to archive download: %v", err)
		return images
	}
	for _, name := range strings.Fields(output) {
		if ref := normalizeImageRef(name); ref != "" {
			images[ref] = struct{}{}
		}
	}
	return images
}

func normalizeImageRef(name string) string {
	ref, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return ""
	}
	return reference.TagNameOnly(ref).String()
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
	component.Url = fmt.Sprintf("%s/%s", d.CDNService, strings.TrimPrefix(url.Url, "/"))
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
