package images

import (
	"fmt"
	"github.com/distribution/reference"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/cache"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/manifest"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/containerd/containerd/pkg/cri/labels"
)

const MAX_IMPORT_RETRY int = 5

type LoadImages struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *LoadImages) Execute(runtime connector.Runtime) (reserr error) {
	var minikubepath = getMinikubePath(t.PipelineCache)
	var minikubeprofile = t.KubeConf.Arg.MinikubeProfile
	var containerManager = t.KubeConf.Cluster.Kubernetes.ContainerManager
	var host = runtime.RemoteHost()

	imageManifests, manifests := t.Manifest.GetImageList()

	retry := func(f func() error, times int) (err error) {
		for i := 0; i < times; i++ {
			err = f()
			if err == nil {
				return nil
			}
			var dur = 5 + (i+1)*10
			// fmt.Printf("import %s failed, wait for %d seconds(%d times)\n", err, dur, i+1)
			logger.Errorf("import error %v, wait for %d seconds(%d times)", err, dur, i+1)
			if (i + 1) < times {
				time.Sleep(time.Duration(dur) * time.Second)
			}
		}
		return
	}

	var mf = filterMinikubeImages(runtime.GetRunner(), host.GetOs(), minikubepath, manifests, minikubeprofile)
	var missingImages []string
	for _, imageRepoTag := range mf {
		if imageRepoTag == "" {
			continue
		}
		reserr = nil
		if inspectImage(runtime.GetRunner(), containerManager, imageRepoTag) == nil {
			logger.Infof("%s already exists", imageRepoTag)
			continue
		}
		missingImages = append(missingImages, imageRepoTag)
	}
	for index, imageRepoTag := range missingImages {
		var start = time.Now()
		var imageHashTag = utils.MD5(imageRepoTag)
		var imageFileName string

		imagesDir := filepath.Join(t.BaseDir, imageManifests[index].Path)

		var found = false
		filepath.Walk(imagesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			if !strings.HasPrefix(info.Name(), imageHashTag) ||
				!HasSuffixI(info.Name(), ".tar.gz", ".tgz", ".tar") {
				return nil
			}

			if strings.HasPrefix(info.Name(), imageHashTag) {
				found = true
				imageFileName = path
				return filepath.SkipDir
			}

			return nil
		})

		if !found {
			return fmt.Errorf("image %s not found in %s", imageRepoTag, imagesDir)
		}

		var imgFileName = filepath.Base(imageFileName)
		var loadCmd string
		var loadParm string

		// unused
		// if runtime.GetSystemInfo().GetFsType() == "zfs" {
		// 	loadParm = "--snapshotter=zfs"
		// }

		if t.KubeConf.Arg.IsOlaresInContainer {
			loadParm = "--no-unpack"
		}

		if runtime.RemoteHost().GetOs() == common.Darwin {
			if HasSuffixI(imgFileName, ".tar.gz", ".tgz") {
				loadCmd = fmt.Sprintf("gunzip -c %s | %s -p %s image load -", imageFileName, minikubepath, minikubeprofile)
			} else {
				loadCmd = fmt.Sprintf("%s -p %s image load %s", minikubepath, minikubeprofile, imageFileName)
			}
		} else {
			switch containerManager {
			case "crio":
				loadCmd = "ctr" // not implement
			case "containerd":
				parsedRef, err := reference.ParseNormalizedNamed(imageRepoTag)
				if err != nil {
					logger.Warnf("parse image name %s error: %v, skip importing", imageRepoTag, err)
					continue
				}
				if HasSuffixI(imgFileName, ".tar.gz", ".tgz") {
					loadCmd = fmt.Sprintf("gunzip -c %s | ctr -n k8s.io images import --index-name %s %s -", imageFileName, parsedRef, loadParm)
				} else {
					loadCmd = fmt.Sprintf("ctr -n k8s.io images import %s %s", imageFileName, loadParm)
				}
			case "isula":
				loadCmd = "isula" // not implement
			default:
			}
		}

		if err := retry(func() error {
			if _, err := runtime.GetRunner().SudoCmd(loadCmd, false, false); err != nil {
				return fmt.Errorf("%s(%s) error: %v", imageRepoTag, imgFileName, err)
			} else {
				logger.Infof("(%d/%d) imported image: %s, time: %s", index+1, len(missingImages), imageRepoTag, time.Since(start))
			}
			return nil
		}, MAX_IMPORT_RETRY); err != nil {
			reserr = fmt.Errorf("%s(%s) error: %v", imageRepoTag, imgFileName, err)
			break
		}
	}
	return
}

type PinImages struct {
	common.KubeAction
	manifest.ManifestAction
}

func (a *PinImages) Execute(runtime connector.Runtime) error {
	_, manifests := a.Manifest.GetImageList()
	if !runtime.GetSystemInfo().IsLinux() {
		return nil
	}
	for _, ref := range manifests {
		parsedRef, err := reference.ParseNormalizedNamed(ref)
		if err != nil {
			logger.Warnf("parse image name %s error: %v, skip pinning", ref, err)
			continue
		}
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("ctr -n k8s.io i label %s %s=%s", parsedRef.String(), labels.PinnedImageLabelKey, labels.PinnedImageLabelValue), false, false); err != nil {
			if strings.Contains(err.Error(), "DEPRECATION") {
				continue
			}
			// tolerate cases where some images are not found
			// e.g., like in the cloud environment and some images are not in the ami
			logger.Warnf("pin image %s error: %v", parsedRef.String(), err)
		}
	}
	return nil
}

func filterMinikubeImages(runner *connector.Runner, osType string, minikubepath string, imagesManifest []string, minikubeProfile string) []string {
	if !strings.EqualFold(osType, common.Darwin) {
		return imagesManifest
	}

	stdout, err := runner.Host.SudoCmd(fmt.Sprintf("%s -p %s image ls", minikubepath, minikubeProfile), false, false)
	if err != nil {
		return imagesManifest
	}

	injectedImages := strings.Split(stdout, "\n")
	if injectedImages == nil || len(injectedImages) == 0 {
		return imagesManifest
	}

	injectedImagesMap := make(map[string]string)
	for _, injected := range injectedImages {
		injectedImagesMap[injected] = injected
	}

	var mf []string
	for _, im := range imagesManifest {
		if _, ok := injectedImagesMap[im]; ok {
			continue
		}
		mf = append(mf, im)
	}

	return mf
}

func getMinikubePath(pipelineCache *cache.Cache) string {
	minikubepath, _ := pipelineCache.GetMustString(common.CacheCommandMinikubePath)
	if minikubepath == "" {
		minikubepath = common.CommandMinikube
	}
	return minikubepath
}

func inspectImage(runner *connector.Runner, containerManager, imageRepoTag string) error {
	if runner.Host.GetOs() == common.Darwin {
		return fmt.Errorf("skip inspect")
	}

	var inspectCmd string = "docker image inspect %s"
	if runner.Host.GetOs() != common.Darwin {
		switch containerManager {
		case "crio": //  not implement
			inspectCmd = "ctr"
		case "containerd":
			inspectCmd = "crictl inspecti -q %s"
		case "isula": // not implement
			inspectCmd = "isula"
		default:
		}
	}

	var cmd = fmt.Sprintf(inspectCmd, imageRepoTag)
	if _, err := runner.Host.SudoCmd(cmd, false, false); err != nil {
		return fmt.Errorf("inspect %s error %v", imageRepoTag, err)
	}

	return nil
}
