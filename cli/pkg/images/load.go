package images

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/containerd/containerd/pkg/cri/labels"
	"github.com/containerd/containerd/reference/docker"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/cache"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/cavaliergopher/grab/v3"
)

const MAX_IMPORT_RETRY int = 5

type CheckImageManifest struct {
	common.KubePrepare
}

func (p *CheckImageManifest) PreCheck(runtime connector.Runtime) (bool, error) {
	// var imageManifest = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.ManifestDir, cc.ManifestImage)
	var imageManifest = path.Join(runtime.GetBaseDir(), cc.ManifestDir, cc.ManifestImage)

	if utils.IsExist(imageManifest) {
		return true, nil
	}
	return false, fmt.Errorf("image manifest not exist")
}

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
				if HasSuffixI(imgFileName, ".tar.gz", ".tgz") {
					loadCmd = fmt.Sprintf("gunzip -c %s | ctr -n k8s.io images import %s -", imageFileName, loadParm)
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
		parsedRef, err := docker.ParseNormalizedNamed(ref)
		if err != nil {
			logger.Warnf("parse image name %s error: %v, skip pinning", ref, err)
			continue
		}
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("ctr -n k8s.io i label %s %s=%s", parsedRef.String(), labels.PinnedImageLabelKey, labels.PinnedImageLabelValue), false, false); err != nil {
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

func downloadImageFile(arch, imageRepoTag, imageFilePath string) error {
	var err error
	if arch == common.Amd64 {
		arch = ""
	} else {
		arch = arch + "/"
	}

	var imageFileName = path.Base(imageFilePath)

	var url = fmt.Sprintf("%s/%s%s", cc.DownloadUrl, arch, imageFileName)
	for i := 5; i > 0; i-- {
		totalSize, _ := getImageFileSize(url)
		if totalSize > 0 {
			logger.Infof("get image %s size: %s", imageRepoTag, utils.FormatBytes(totalSize))
		}

		client := grab.NewClient()
		req, _ := grab.NewRequest(imageFilePath, url)
		req.HTTPRequest = req.HTTPRequest.WithContext(context.Background())
		ctx, cancel := context.WithTimeout(req.HTTPRequest.Context(), 5*time.Minute)
		defer cancel()

		req.HTTPRequest = req.HTTPRequest.WithContext(ctx)
		resp := client.Do(req)

		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()

		var downloaded int64
	Loop:
		for {
			select {
			case <-t.C:
				downloaded = resp.BytesComplete()
				var progressInfo string
				if totalSize != 0 {
					result := float64(downloaded) / float64(totalSize)
					progressInfo = fmt.Sprintf("transferred %s %s / %s (%.2f%%) / speed: %s", imageFileName, utils.FormatBytes(resp.BytesComplete()), utils.FormatBytes(totalSize), math.Round(result*10000)/100, utils.FormatBytes(int64(resp.BytesPerSecond())))
					logger.Info(progressInfo)
				} else {
					progressInfo = fmt.Sprintf("transferred %s %s / speed: %s\n", imageFileName, utils.FormatBytes(resp.BytesComplete()), utils.FormatBytes(int64(resp.BytesPerSecond())))
					logger.Infof(progressInfo)
				}
			case <-resp.Done:
				break Loop
			}
		}

		if err = resp.Err(); err != nil {
			logger.Infof("download %s error %v", imageFileName, err)
			time.Sleep(2 * time.Second)
			continue
		}
	}

	return err
}

func pullImage(runner *connector.Runner, containerManager, imageRepoTag, imageHashTag, dst string) error {
	var pullCmd string = "docker"
	var inspectCmd string = "docker"
	var exportCmd string = "docker"
	switch containerManager {
	case "crio": // not implement
		pullCmd = "ctr"
		inspectCmd = "ctr"
		exportCmd = "ctr"
	case "containerd":
		pullCmd = "crictl pull %s"
		inspectCmd = "crictl inspecti -q %s"
		exportCmd = "ctr -n k8s.io image export %s %s"
	case "isula": // not implement
		pullCmd = "isula"
		inspectCmd = "isula"
		exportCmd = "isula"
	default:
		pullCmd = "docker pull %s"
		exportCmd = "docker save -o %s %s"
	}

	var cmd = fmt.Sprintf(pullCmd, imageRepoTag)
	if _, err := runner.Host.SudoCmd(cmd, false, false); err != nil {
		return fmt.Errorf("pull %s error %v", imageRepoTag, err)
	}

	var repoTag = imageRepoTag
	if containerManager == "containerd" {
		cmd = fmt.Sprintf(inspectCmd, imageRepoTag)
		stdout, err := runner.Host.SudoCmd(cmd, false, false)
		if err != nil {
			return fmt.Errorf("inspect %s error %v", imageRepoTag, err)
		}
		var ii ImageInspect
		if err = json.Unmarshal([]byte(stdout), &ii); err != nil {
			return fmt.Errorf("unmarshal %s error %v", imageRepoTag, err)
		}
		repoTag = ii.Status.RepoTags[0]
	}

	var dstFile = path.Join(dst, fmt.Sprintf("%s.tar", imageHashTag))
	cmd = fmt.Sprintf(exportCmd, dstFile, repoTag)
	if _, err := runner.Host.SudoCmd(cmd, false, false); err != nil {
		return fmt.Errorf("export %s error: %v", imageRepoTag, err)
	}
	if _, err := runner.Host.SudoCmd(fmt.Sprintf("gzip %s", dstFile), false, false); err != nil {
		return fmt.Errorf("gzip %s error: %v", dstFile, err)
	}

	return nil
}

func getImageFileSize(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("bad status: %s", resp.Status)
	}

	size, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return -1, fmt.Errorf("failed to parse content length: %v, header: %s", err, resp.Header.Get("Content-Length"))
	}
	return size, nil
}

type RateLimiter struct {
	r, n int
}

func NewLimiter(r int) grab.RateLimiter {
	return &RateLimiter{r: r}
}

func (c *RateLimiter) WaitN(ctx context.Context, n int) (err error) {
	c.n += n
	time.Sleep(
		time.Duration(1.00 / float64(c.r) * float64(n) * float64(time.Second)))
	return
}
