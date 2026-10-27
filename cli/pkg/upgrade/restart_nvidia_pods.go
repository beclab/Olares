package upgrade

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type restartNvidiaApplicationPods struct {
	common.KubeAction
}

func (a *restartNvidiaApplicationPods) Execute(runtime connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return restartNvidiaPods(ctx, kube)
}

// Existing GPU containers must be recreated to pick up the upgraded NVIDIA
// runtime. Match resource consumers, including HAMi GPU memory and init containers.
func podUsesNvidiaResources(pod *corev1.Pod) bool {
	usesNvidia := func(containers []corev1.Container) bool {
		for _, container := range containers {
			for _, resources := range []corev1.ResourceList{container.Resources.Requests, container.Resources.Limits} {
				for name, quantity := range resources {
					if strings.HasPrefix(string(name), "nvidia.com/") && quantity.Sign() > 0 {
						return true
					}
				}
			}
		}
		return false
	}
	return usesNvidia(pod.Spec.Containers) || usesNvidia(pod.Spec.InitContainers)
}

func restartNvidiaPods(ctx context.Context, kube kubernetes.Interface) error {
	pods, err := kube.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list NVIDIA application pods: %w", err)
	}
	for i := range pods.Items {
		pod := &pods.Items[i]
		if pod.DeletionTimestamp != nil || pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || !podUsesNvidiaResources(pod) {
			continue
		}
		// Guard against deleting a replacement with the same name (StatefulSets).
		err := kube.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{UID: &pod.UID},
		})
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("restart NVIDIA application pod %s/%s: %w", pod.Namespace, pod.Name, err)
		}
		logger.Infof("deleted NVIDIA application pod %s/%s for recreation", pod.Namespace, pod.Name)
	}
	return nil
}
