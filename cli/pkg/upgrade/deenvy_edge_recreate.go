package upgrade

import (
	"context"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

// oesRecreateBatchLimit caps deletes per attempt to avoid a thundering herd.
const oesRecreateBatchLimit = 50

const (
	l4ProxyNamespace      = "os-network"
	l4ProxyDeploymentName = "l4-bfl-proxy"
	l4ReadyWaitTimeout    = 3 * time.Minute
	l4ReadyPollInterval   = 5 * time.Second
)

func kubeClientFromRuntime() (kubernetes.Interface, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func l4ProxyDeploymentReady(ctx context.Context, kube kubernetes.Interface) bool {
	if kube == nil {
		return false
	}
	dep, err := kube.AppsV1().Deployments(l4ProxyNamespace).Get(ctx, l4ProxyDeploymentName, metav1.GetOptions{})
	if err != nil {
		return false
	}
	return dep.Status.ReadyReplicas >= 1
}

// waitL4ProxyReady polls until l4-bfl-proxy has a Ready replica or timeout.
func waitL4ProxyReady(ctx context.Context, kube kubernetes.Interface, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if l4ProxyDeploymentReady(ctx, kube) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		select {
		case <-ctx.Done():
			return l4ProxyDeploymentReady(ctx, kube)
		case <-time.After(l4ReadyPollInterval):
		}
	}
}

func listPodsWithBusinessOES(ctx context.Context, kube kubernetes.Interface) ([]corev1.Pod, error) {
	pods, err := kube.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var out []corev1.Pod
	for i := range pods.Items {
		p := pods.Items[i]
		for _, c := range append(p.Spec.Containers, p.Spec.InitContainers...) {
			if isBusinessOESContainer(p, c) {
				out = append(out, p)
				break
			}
		}
	}
	return out, nil
}

// recreatePodsWithBusinessOES deletes pods still carrying business oes so
// controllers reschedule them; admission skips oes once l4 is Ready.
func recreatePodsWithBusinessOES(ctx context.Context, kube kubernetes.Interface) (int, error) {
	offenders, err := listPodsWithBusinessOES(ctx, kube)
	if err != nil {
		return 0, err
	}
	deleted := 0
	for i := range offenders {
		if deleted >= oesRecreateBatchLimit {
			break
		}
		p := offenders[i]
		err := kube.CoreV1().Pods(p.Namespace).Delete(ctx, p.Name, metav1.DeleteOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return deleted, fmt.Errorf("delete oes pod %s/%s: %w", p.Namespace, p.Name, err)
		}
		deleted++
		logger.Infof("deenvy: recreate (delete) oes pod %s/%s", p.Namespace, p.Name)
	}
	return deleted, nil
}

// deenvyEdgeBestEffortRecreate runs only on the upgrade path (daily upgrader).
// It never blocks UpdateOlaresVersion: failures and l4-not-Ready are logged only.
type deenvyEdgeBestEffortRecreate struct {
	common.KubeAction
}

func (a *deenvyEdgeBestEffortRecreate) Execute(runtime connector.Runtime) error {
	kube, err := kubeClientFromRuntime()
	if err != nil {
		logger.Errorf("deenvy: skip oes recreate; kube client: %v", err)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), l4ReadyWaitTimeout+2*time.Minute)
	defer cancel()

	if !waitL4ProxyReady(ctx, kube, l4ReadyWaitTimeout) {
		logger.Warnf("deenvy: l4-bfl-proxy not Ready within %s; skip oes recreate (upgrade continues)", l4ReadyWaitTimeout)
		return nil
	}
	logger.Infof("deenvy: l4-bfl-proxy Ready; best-effort recreate of business oes pods")

	n, err := recreatePodsWithBusinessOES(ctx, kube)
	if err != nil {
		logger.Errorf("deenvy: oes recreate failed (upgrade continues): %v", err)
		return nil
	}
	if n > 0 {
		logger.Infof("deenvy: deleted %d oes pods for recreate (batch cap %d)", n, oesRecreateBatchLimit)
	} else {
		logger.Infof("deenvy: no business oes pods to recreate")
	}
	return nil
}

// deenvyEdgeUpgradeTasks returns upgrade-only best-effort recreate (never gates version write).
func deenvyEdgeUpgradeTasks() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "DeenvyEdgeBestEffortRecreate",
			Action: &deenvyEdgeBestEffortRecreate{},
			Retry:  1,
		},
	}
}
