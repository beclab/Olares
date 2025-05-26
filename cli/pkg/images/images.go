/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package images

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"github.com/pkg/errors"
)

const (
	cnRegistry          = "registry.cn-beijing.aliyuncs.com"
	cnNamespaceOverride = "kubesphereio"
)

// Image defines image's info.
type Image struct {
	RepoAddr          string
	Namespace         string
	NamespaceOverride string
	Repo              string
	Tag               string
	Group             string
	Enable            bool
}

// Images contains a list of Image
type Images struct {
	Images []Image
}

// ImageName is used to generate image's full name.
func (image Image) ImageName() string {
	return fmt.Sprintf("%s:%s", image.ImageRepo(), image.Tag)
}

// ImageRepo is used to generate image's repo address.
func (image Image) ImageRepo() string {
	var prefix string

	if os.Getenv("KKZONE") == "cn" {
		if image.RepoAddr == "" || image.RepoAddr == cnRegistry {
			image.RepoAddr = cnRegistry
			image.NamespaceOverride = cnNamespaceOverride
		}
	}

	if image.RepoAddr == "" {
		if image.Namespace == "" {
			prefix = ""
		} else {
			prefix = fmt.Sprintf("%s/", image.Namespace)
		}
	} else {
		if image.NamespaceOverride == "" {
			if image.Namespace == "" {
				if strings.Contains(image.RepoAddr, "docker.io") {
					prefix = fmt.Sprintf("%s/library/", image.RepoAddr)
				} else {
					prefix = fmt.Sprintf("%s/", image.RepoAddr)
				}
			} else {
				prefix = fmt.Sprintf("%s/%s/", image.RepoAddr, image.Namespace)
			}
		} else {
			prefix = fmt.Sprintf("%s/%s/", image.RepoAddr, image.NamespaceOverride)
		}
	}

	return fmt.Sprintf("%s%s", prefix, image.Repo)
}

// PullImages is used to pull images in the list of Image.
func (images *Images) PullImages(runtime connector.Runtime, kubeConf *common.KubeConf) error {
	pullCmd := "docker"
	switch kubeConf.Cluster.Kubernetes.ContainerManager {
	case "crio":
		pullCmd = "crictl"
	case "containerd":
		pullCmd = "crictl"
	case "isula":
		pullCmd = "isula"
	default:
		pullCmd = "docker"
	}

	host := runtime.RemoteHost()
	for _, image := range images.Images {
		switch {
		case host.IsRole(common.Master) && image.Group == kubekeyapiv1alpha2.Master && image.Enable,
			host.IsRole(common.Worker) && image.Group == kubekeyapiv1alpha2.Worker && image.Enable,
			(host.IsRole(common.Master) || host.IsRole(common.Worker)) && image.Group == kubekeyapiv1alpha2.K8s && image.Enable,
			host.IsRole(common.ETCD) && image.Group == kubekeyapiv1alpha2.Etcd && image.Enable:

			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s inspecti -q %s", pullCmd, image.ImageName()), false, false); err == nil {
				logger.Infof("%s pull image %s exists", pullCmd, image.ImageName())
				continue
			}

			// fmt.Printf("%s downloading image %s\n", pullCmd, image.ImageName())
			logger.Debugf("%s pull image: %s - %s", host.GetName(), image.ImageName(), runtime.RemoteHost().GetName())
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s pull %s", pullCmd, image.ImageName()), false, false); err != nil {
				return errors.Wrap(err, "pull image failed")
			}
		default:
			continue
		}

	}
	return nil
}

// type LocalImage struct {
// 	Filename string
// }

// type LocalImages []LocalImage

// func (i LocalImages) LoadImages(runtime connector.Runtime, kubeConf *common.KubeConf) error {
// 	loadCmd := "docker"
// 	host := runtime.RemoteHost()
// 	retry := func(f func() error, times int) (err error) {
// 		for i := 0; i < times; i++ {
// 			err = f()
// 			if err == nil {
// 				return nil
// 			}
// 			var dur = 5 + (i+1)*10
// 			logger.Warnf("load image %s failed, wait for %d seconds(%d times)", err, dur, i+1)
// 			if (i + 1) < times {
// 				time.Sleep(time.Duration(dur) * time.Second)
// 			}
// 		}

// 		return
// 	}

// 	for _, image := range i {
// 		switch {
// 		case host.IsRole(common.Master):
// 			// logger.Debugf("%s preloading image: %s", host.GetName(), image.Filename)
// 			start := time.Now()
// 			fileName := filepath.Base(image.Filename)
// 			// fileName = strings.ReplaceAll(fileName, ".gz", "")
// 			// fmt.Println(">>> ", fileName, HasSuffixI(image.Filename, ".tar.gz", ".tgz"))
// 			if HasSuffixI(image.Filename, ".tar.gz", ".tgz") {
// 				switch kubeConf.Cluster.Kubernetes.ContainerManager {
// 				case "crio":
// 					loadCmd = "ctr" // BUG
// 				case "containerd":
// 					loadCmd = "ctr -n k8s.io images import -"
// 				case "isula":
// 					loadCmd = "isula"
// 				default:
// 					loadCmd = "docker load"
// 				}

// 				// continue if load image error
// 				if err := retry(func() error {
// 					logger.Infof("preloading image: %s", fileName)
// 					if stdout, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("env PATH=$PATH gunzip -c %s | %s", image.Filename, loadCmd), false, false); err != nil {
// 						return fmt.Errorf("%s", fileName)
// 					} else {
// 						logger.Infof("%s in %s\n", formatLoadImageRes(stdout, fileName), time.Since(start))
// 						// fmt.Printf("%s in %s\n", formatLoadImageRes(stdout, fileName), time.Since(start))
// 					}
// 					return nil
// 				}, 5); err != nil {
// 					return fmt.Errorf("%s", fileName)
// 				}
// 			} else if HasSuffixI(image.Filename, ".tar") {
// 				switch kubeConf.Cluster.Kubernetes.ContainerManager {
// 				case "crio":
// 					loadCmd = "ctr" // BUG
// 				case "containerd":
// 					loadCmd = "ctr -n k8s.io images import"
// 				case "isula":
// 					loadCmd = "isula"
// 				default:
// 					loadCmd = "docker load -i"
// 				}

// 				if err := retry(func() error {
// 					logger.Infof("preloading image: %s", fileName)
// 					if stdout, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("env PATH=$PATH %s %s", loadCmd, image.Filename), false, false); err != nil {
// 						return fmt.Errorf("%s", fileName)
// 					} else {
// 						logger.Infof("%s in %s\n", formatLoadImageRes(stdout, fileName), time.Since(start))
// 						// fmt.Printf("%s in %s\n", formatLoadImageRes(stdout, fileName), time.Since(start))
// 					}

// 					return nil
// 				}, 5); err != nil {
// 					return fmt.Errorf("%s", fileName)
// 				}
// 			} else {
// 				logger.Warnf("invalid image file name %s, skip ...", image.Filename)
// 				return nil
// 			}
// 		default:
// 			continue
// 		}

// 	}
// 	return nil

// }

func formatLoadImageRes(str string, fileName string) string {
	if strings.Contains(str, "(sha256:") {
		str = strings.Split(str, "(sha256:")[0]
	} else {
		return fmt.Sprintf("%s %s", str, fileName)
	}
	return fmt.Sprintf("%s (%s)...done ", str, fileName)
}

func HasSuffixI(s string, suffixes ...string) bool {
	s = strings.ToLower(s)
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, strings.ToLower(suffix)) {
			return true
		}
	}
	return false
}

func readImageManifest(mfPath string) ([]string, error) {
	if !util.IsExist(mfPath) {
		return nil, nil
	}

	file, err := os.Open(mfPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var res []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		res = append(res, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
