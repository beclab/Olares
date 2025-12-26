package upgrade

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/containerd"
	"github.com/dustin/go-humanize"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/utils/strings/slices"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type preCheck struct {
	commands.Operation
}

var _ commands.Interface = &preCheck{}

func NewPreCheck() commands.Interface {
	return &preCheck{
		Operation: commands.Operation{
			Name: commands.UpgradePreCheck,
		},
	}
}

func (i *preCheck) Execute(ctx context.Context, p any) (res any, err error) {
	klog.Info("Starting upgrade pre check")

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

		if podStatus != "Running" && podStatus != "Completed" && podStatus != "Succeeded" {
			klog.Warningf("Pod %s/%s is not healthy before upgrade: %s", pod.Namespace, pod.Name, podStatus)
			continue
		}

		if !utils.IsPodReady(&pod) && pod.Status.Phase == corev1.PodRunning {
			klog.Warningf("Pod %s/%s is running but not ready before upgrade", pod.Namespace, pod.Name)
			continue
		}
	}

	// if any user is in the progress of activation
	// upgrade cannot start
	dynamicClient, err := utils.GetDynamicClient()
	if err != nil {
		err = fmt.Errorf("failed to get dynamic client: %v", err)
		klog.Error(err.Error())
		return nil, err
	}

	users, err := utils.ListUsers(ctx, dynamicClient)
	if err != nil {
		err = fmt.Errorf("failed to list users: %v", err)
		klog.Error(err.Error())
		return nil, err
	}

	var activatingUsers []string
	for _, user := range users {
		status, ok := user.GetAnnotations()["bytetrade.io/wizard-status"]
		if !ok || slices.Contains([]string{"", "wait_activate_vault", "completed"}, status) {
			continue
		}
		activatingUsers = append(activatingUsers, user.GetName())
	}
	if len(activatingUsers) > 0 {
		return nil, fmt.Errorf("waiting for user to finish activation: %s", strings.Join(activatingUsers, ", "))
	}

	// the new MongoDB version has a different implementation from the old version.
	// if an old MongoDB instance exists, it must be uninstalled before upgrading.
	{
		gvr := schema.GroupVersionResource{Group: "app.bytetrade.io", Version: "v1alpha1", Resource: "applicationmanagers"}
		am, err := dynamicClient.Resource(gvr).Get(ctx, "os-platform-mongodb", metav1.GetOptions{})
		if err != nil {
			if !apierrors.IsNotFound(err) {
				return nil, fmt.Errorf("failed to check mongodb application manager: %v", err)
			}
		} else if am != nil {
			state, _, _ := unstructured.NestedString(am.Object, "status", "state")
			switch strings.ToLower(state) {
			case "installing", "running":
				return nil, fmt.Errorf("mongodb is %s, please remove it before upgrade. if mongodb is installing, you can cancel it in market. if running, execute 'kubectl delete appmgr os-platform-mongodb' in control-hub -> olares shell", state)
			}
		}
	}

	// in v1.12.3, argo has been moved to os-platform.
	// to avoid CRD resource conflicts, if wise is installed and includes argo crd, you must uninstall wise first, including sharedserver.
	{
		isKnowledgeSharedNsExist := false
		if _, err := client.CoreV1().Namespaces().Get(ctx, "knowledge-shared", metav1.GetOptions{}); err == nil {
			isKnowledgeSharedNsExist = true
		} else if !apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to check namespace 'knowledge-shared': %v", err)
		}
		xclient, err := utils.GetApixClient()
		if err != nil {
			err = fmt.Errorf("failed to get apix client: %v", err)
			klog.Error(err.Error())
			return nil, err
		}
		isKnowledgeArgoCRDExist := false
		crds, err := xclient.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
		if err != nil {
			klog.Errorf("failed to list crds %v", err)
			return nil, err
		}
		for _, crd := range crds.Items {
			if crd.Spec.Group == "argoproj.io" && crd.Annotations["meta.helm.sh/release-name"] == "knowledge" {
				isKnowledgeArgoCRDExist = true
				break
			}
		}
		if isKnowledgeSharedNsExist && isKnowledgeArgoCRDExist {
			return nil, fmt.Errorf("namespace 'knowledge-shared' exists (wise); please uninstall Wise and Shared Server before upgrade")
		}
	}
	klog.Info("pre checks passed for upgrade")

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
