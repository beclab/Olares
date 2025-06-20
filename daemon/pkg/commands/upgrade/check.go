package upgrade

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
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
	version, ok := p.(string)
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
	versionHintFile := filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+version, "version.hint")
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

func (i *healthCheck) Execute(ctx context.Context, _ any) (res any, err error) {
	klog.Info("Starting upgrade health check")

	const minAvailableSpace = 100 * 1024 * 1024 * 1024 // 100GB in bytes
	availableSpace, err := utils.GetDiskAvailableSpace("/")
	if err != nil {
		return nil, fmt.Errorf("error checking disk space: %s", err)
	}
	klog.Infof("Root partition available space: %.2fGB", float64(availableSpace)/(1024*1024*1024))
	if availableSpace < minAvailableSpace {
		return nil, fmt.Errorf("insufficient disk space: %.2fGB available, minimum 100GB required",
			float64(availableSpace)/(1024*1024*1024))
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
		roles := sets.NewString()
		for k, v := range node.Labels {
			switch {
			case strings.HasPrefix(k, "node-role.kubernetes.io/"):
				if role := strings.TrimPrefix(k, "node-role.kubernetes.io/"); len(role) > 0 {
					roles.Insert(role)
				}

			case k == "kubernetes.io/role" && v != "":
				roles.Insert(v)
			}
		}
		if !roles.HasAny("control-plane", "master") {
			continue
		}
		if node.Spec.Unschedulable {
			return nil, fmt.Errorf("node %s is unschedulable", node.Name)
		}
		var readyConditionExists bool
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				readyConditionExists = true
				if condition.Status != corev1.ConditionTrue {
					return nil, fmt.Errorf("node %s is not ready", node.Name)
				}
			}
		}
		if !readyConditionExists {
			return nil, fmt.Errorf("node %s's condition is unknown", node.Name)
		}
	}

	criticalNamespaces := []string{"os-platform", "os-framework"}
	for _, namespace := range criticalNamespaces {
		pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("error listing pods in namespace %s: %s", namespace, err)
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodSucceeded {
				continue
			}

			podStatus := utils.GetPodStatus(&pod)

			if podStatus != "Running" && podStatus != "Completed" {
				klog.Errorf("Pod %s/%s is not healthy: %s", namespace, pod.Name, podStatus)
				return nil, fmt.Errorf("pod %s/%s is not healthy: %s", namespace, pod.Name, podStatus)
			}

			if !utils.IsPodReady(&pod) && pod.Status.Phase == corev1.PodRunning {
				klog.Warningf("Pod %s/%s is running but not ready", namespace, pod.Name)
				return nil, fmt.Errorf("pod %s/%s is running but not ready", namespace, pod.Name)
			}
		}
	}

	klog.Info("health checks passed for upgrade")

	return newExecutionRes(true, nil), nil
}
