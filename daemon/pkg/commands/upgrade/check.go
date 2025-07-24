package upgrade

import (
	"context"
	"errors"
	"fmt"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/containerd"
	"github.com/dustin/go-humanize"
	v1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
)

type versionCompatibilityCheck struct {
	commands.Operation
}

var _ commands.Interface = &versionCompatibilityCheck{}

func NewVersionCompatibilityCheck() commands.Interface {
	return &versionCompatibilityCheck{
		Operation: commands.Operation{
			Name: commands.VersionCompatibilityCheck,
		},
	}
}

func (i *versionCompatibilityCheck) Execute(ctx context.Context, p any) (res any, err error) {
	target, ok := p.(state.UpgradeTarget)
	if !ok {
		err = errors.New("invalid param")
		return
	}
	dynamicClient, err := utils.GetDynamicClient()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes client: %v", err)
	}
	olaresVersion, err := utils.GetTerminusVersion(ctx, dynamicClient)
	if err != nil {
		return nil, fmt.Errorf("error getting olares version: %v", err)
	}
	currentVersion, err := semver.NewVersion(*olaresVersion)
	if err != nil {
		return nil, fmt.Errorf("error parsing current olares version %s: %v", *olaresVersion, err)
	}
	versionHintFile := filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+target.Version.Original(), "version.hint")
	content, err := os.ReadFile(versionHintFile)
	if err != nil {
		return nil, fmt.Errorf("error reading version hint file %s: %v", versionHintFile, err)
	}
	versionHint := make(map[string]map[string]string)
	if err := yaml.Unmarshal(content, &versionHint); err != nil {
		return nil, fmt.Errorf("error parsing version hint file %s: %v", versionHintFile, err)
	}
	upgradeField, ok := versionHint["upgrade"]
	if !ok {
		return nil, fmt.Errorf("no upgrade field found in version hint file %s, content: %s", versionHintFile, content)
	}
	minVersionStr, ok := upgradeField["minVersion"]
	if !ok || minVersionStr == "" {
		return nil, fmt.Errorf("no minVersion field found in version hint file %s, content: %s", versionHintFile, content)
	}
	minVersion, err := semver.NewVersion(minVersionStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing minVersion: %s in version hint file %s: %v", minVersionStr, versionHintFile, err)
	}
	if currentVersion.LessThan(minVersion) {
		return nil, fmt.Errorf("minVersion %s is greater than current version %s", minVersionStr, *olaresVersion)
	}
	return newExecutionRes(true, nil), nil
}

type healthCheck struct {
	commands.Operation
}

var _ commands.Interface = &healthCheck{}

func NewHealthCheck() commands.Interface {
	return &healthCheck{
		Operation: commands.Operation{
			Name: commands.UpgradeHealthCheck,
		},
	}
}

func (i *healthCheck) Execute(ctx context.Context, p any) (res any, err error) {
	klog.Info("Starting upgrade health check")

	target, ok := p.(state.UpgradeTarget)
	if !ok {
		return nil, errors.New("invalid param")
	}
	arch := "amd64"
	if runtime.GOARCH == "arm" {
		arch = "arm64"
	}
	componentManifestFilePath := filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+target.Version.Original(), "images", "installation.manifest."+arch)
	components, err := unmarshalComponentManifestFile(componentManifestFilePath)
	if err != nil {
		return nil, fmt.Errorf("error parsing component manifest file %s: %v", componentManifestFilePath, err)
	}
	criImageService, err := containerd.NewCRIImageService()
	if err != nil {
		return nil, fmt.Errorf("error creating cri image service: %v", err)
	}
	images, err := criImageService.ListImages(ctx, &v1.ImageFilter{})
	if err != nil {
		return nil, fmt.Errorf("error listing images: %v", err)
	}
	var requiredSpace uint64
	for _, component := range components {
		if component.Type != "image" {
			continue
		}
		var imageExists bool
		for _, image := range images {
			for _, repoTag := range image.RepoTags {
				if strings.Contains(repoTag, component.FileID) {
					imageExists = true
					break
				}
			}
			if imageExists {
				break
			}
		}
		if !imageExists {
			// for now, the compressed layer and sha256 content hash in the manifest
			// can not be used for us to compare and calculate a precise space requirement
			// because the "docker save" command exports image in uncompressed format
			// and dockerhub stores & distributes the image in compressed format
			// so we just make a rough number based on the compressed image archive file
			requiredSpace += component.Size * 3
		}
	}
	klog.Infof("Required space for image import: %s", humanize.Bytes(requiredSpace))
	if err := tryToUseDiskSpace(containerd.DefaultContainerdRootPath, requiredSpace); err != nil {
		return nil, err
	}

	client, err := utils.GetKubeClient()
	if err != nil {
		return nil, fmt.Errorf("error getting kubernetes client: %s", err)
	}
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing nodes: %s", err)
	}
	for _, node := range nodes.Items {
		//roles := sets.NewString()
		//for k, v := range node.Labels {
		//	switch {
		//	case strings.HasPrefix(k, "node-role.kubernetes.io/"):
		//		if role := strings.TrimPrefix(k, "node-role.kubernetes.io/"); len(role) > 0 {
		//			roles.Insert(role)
		//		}
		//
		//	case k == "kubernetes.io/role" && v != "":
		//		roles.Insert(v)
		//	}
		//}
		//if !roles.HasAny("control-plane", "master") {
		//	continue
		//}
		if node.Spec.Unschedulable {
			return nil, fmt.Errorf("node %s: unschedulable", node.Name)
		}
		var readyConditionExists bool
		for _, condition := range node.Status.Conditions {
			switch condition.Type {
			case corev1.NodeReady:
				readyConditionExists = true
				if condition.Status != corev1.ConditionTrue {
					return nil, fmt.Errorf("node %s: not ready", node.Name)
				}
			case corev1.NodeMemoryPressure, corev1.NodeDiskPressure,
				corev1.NodePIDPressure, corev1.NodeNetworkUnavailable:
				if condition.Status == corev1.ConditionTrue {
					return nil, fmt.Errorf("node %s: %s", node.Name, condition.Type)
				}
			}
		}
		if !readyConditionExists {
			return nil, fmt.Errorf("node %s: condition unknown", node.Name)
		}
	}

	pods, err := client.CoreV1().Pods(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	for _, pod := range pods.Items {
		if !strings.HasPrefix(pod.Namespace, "os-") {
			continue
		}
		if pod.Status.Phase == corev1.PodSucceeded {
			continue
		}

		podStatus := utils.GetPodStatus(&pod)

		if podStatus != "Running" && podStatus != "Completed" {
			klog.Errorf("Pod %s/%s is not healthy: %s", pod.Namespace, pod.Name, podStatus)
			return nil, fmt.Errorf("pod %s/%s is not healthy: %s", pod.Namespace, pod.Name, podStatus)
		}

		if !utils.IsPodReady(&pod) && pod.Status.Phase == corev1.PodRunning {
			klog.Warningf("Pod %s/%s is running but not ready", pod.Namespace, pod.Name)
			return nil, fmt.Errorf("pod %s/%s is running but not ready", pod.Namespace, pod.Name)
		}
	}

	klog.Info("health checks passed for upgrade")

	return newExecutionRes(true, nil), nil
}

type downloadSpaceCheck struct {
	commands.Operation
}

var _ commands.Interface = &downloadSpaceCheck{}

func NewDownloadSpaceCheck() commands.Interface {
	return &downloadSpaceCheck{
		Operation: commands.Operation{
			Name: commands.DownloadSpaceCheck,
		},
	}
}

func (i *downloadSpaceCheck) Execute(ctx context.Context, p any) (res any, err error) {
	target, ok := p.(state.UpgradeTarget)
	if !ok {
		return nil, errors.New("invalid param")
	}
	klog.Info("Starting download space check")
	arch := "amd64"
	if runtime.GOARCH == "arm" {
		arch = "arm64"
	}
	componentManifestFilePath := filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+target.Version.Original(), "images", "installation.manifest."+arch)
	components, err := unmarshalComponentManifestFile(componentManifestFilePath)
	if err != nil {
		return nil, fmt.Errorf("error parsing component manifest file %s: %v", componentManifestFilePath, err)
	}
	var requiredSpace uint64
	for name, component := range components {
		path := filepath.Join(commands.TERMINUS_BASE_DIR, component.Path, name)
		_, err := os.Stat(path)
		if err == nil {
			continue
		}
		if os.IsNotExist(err) {
			requiredSpace += component.Size
			continue
		}
		return nil, fmt.Errorf("failed to check existence of file %s: %v", path, err)
	}
	klog.Infof("Required space for download: %s", humanize.Bytes(requiredSpace))

	if err := tryToUseDiskSpace(commands.TERMINUS_BASE_DIR, requiredSpace); err != nil {
		return nil, err
	}

	klog.Info("Space check passed for download")

	return newExecutionRes(true, nil), nil
}
