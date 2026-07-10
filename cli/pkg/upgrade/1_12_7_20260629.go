package upgrade

import (
	"context"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"

	"github.com/Masterminds/semver/v3"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const systemClusterCriticalPriorityClassName = "system-cluster-critical"

// systemNamespacesForPriorityPatch lists the well-known system namespaces in
// which Deployments / DaemonSets / StatefulSets without an explicit
// priorityClassName should be patched to system-cluster-critical.
var systemNamespacesForPriorityPatch = []string{
	"kube-system",
	"kubesphere-monitoring-system",
	"kubesphere-system",
	"os-framework",
	"os-gateway",
	"os-mesh",
	"os-gpu",
	"os-network",
	"os-protected",
}

type upgrader_1_12_7_20260629 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260629) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260629")
}

func (u upgrader_1_12_7_20260629) UpgradeSystemComponents() []task.Interface {
	tasks := append([]task.Interface{
		&task.LocalTask{
			Name:   "PatchWorkloadPriorityClassName",
			Action: new(patchWorkloadPriorityClassName),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	})
	tasks = append(u.upgraderBase.UpgradeSystemComponents(), tasks...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260629{})
}

// patchWorkloadPriorityClassName iterates over a fixed list of system
// namespaces plus the per-user user-space-<name> / user-system-<name>
// namespaces and, for every Deployment / DaemonSet / StatefulSet whose pod
// template does NOT already set a priorityClassName, patches it to
// system-cluster-critical via a strategic merge patch. Workloads that already
// have any priorityClassName configured (e.g. system-node-critical) are left
// untouched, as are Helm-managed workloads (so the patch is not reverted on
// the next `helm upgrade`).
type patchWorkloadPriorityClassName struct {
	common.KubeAction
}

// isHelmManagedWorkload reports whether the given object is owned by Helm.
// We treat both the standard managed-by label and the meta.helm.sh release
// annotations as evidence, because some charts only set one or the other.
func isHelmManagedWorkload(obj metav1.Object) bool {
	if v, ok := obj.GetLabels()["app.kubernetes.io/managed-by"]; ok && v == "Helm" {
		return true
	}
	annotations := obj.GetAnnotations()
	if _, ok := annotations["meta.helm.sh/release-name"]; ok {
		return true
	}
	if _, ok := annotations["meta.helm.sh/release-namespace"]; ok {
		return true
	}
	return false
}

func (a *patchWorkloadPriorityClassName) Execute(_ connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %v", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	namespaces := append([]string{}, systemNamespacesForPriorityPatch...)
	userNamespaces, err := listUserScopedNamespacesForPriorityPatch(config)
	if err != nil {
		return fmt.Errorf("failed to list user-scoped namespaces: %v", err)
	}
	namespaces = append(namespaces, userNamespaces...)

	for _, ns := range namespaces {
		if err := patchNamespaceWorkloadsMissingPriorityClassName(client, ns); err != nil {
			return fmt.Errorf("failed to patch workloads in namespace %s: %v", ns, err)
		}
	}
	return nil
}

func listUserScopedNamespacesForPriorityPatch(config *rest.Config) ([]string, error) {
	scheme := kruntime.NewScheme()
	if err := iamv1alpha2.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add iam scheme: %v", err)
	}
	userClient, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create user client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	var userList iamv1alpha2.UserList
	if err := userClient.List(ctx, &userList); err != nil {
		return nil, fmt.Errorf("failed to list users: %v", err)
	}
	namespaces := make([]string, 0, len(userList.Items)*2)
	for _, user := range userList.Items {
		if user.DeletionTimestamp != nil {
			continue
		}
		namespaces = append(namespaces,
			fmt.Sprintf("user-space-%s", user.Name),
			fmt.Sprintf("user-system-%s", user.Name),
		)
	}
	return namespaces, nil
}

func patchNamespaceWorkloadsMissingPriorityClassName(client *kubernetes.Clientset, namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if _, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("namespace %s not found, skipping priorityClassName patch", namespace)
			return nil
		}
		return fmt.Errorf("failed to get namespace %s: %v", namespace, err)
	}

	patch := []byte(fmt.Sprintf(
		`{"spec":{"template":{"spec":{"priorityClassName":%q}}}}`,
		systemClusterCriticalPriorityClassName,
	))

	deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments in %s: %v", namespace, err)
	}
	for i := range deployments.Items {
		d := &deployments.Items[i]
		if d.Spec.Template.Spec.PriorityClassName != "" {
			continue
		}
		if isHelmManagedWorkload(&d.ObjectMeta) {
			logger.Infof("skipping Helm-managed deployment %s/%s priorityClassName patch", namespace, d.Name)
			continue
		}
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			_, patchErr := client.AppsV1().Deployments(namespace).Patch(
				ctx, d.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{},
			)
			return patchErr
		}); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("failed to patch deployment %s/%s priorityClassName: %v", namespace, d.Name, err)
		}
		logger.Infof("patched deployment %s/%s priorityClassName to %s", namespace, d.Name, systemClusterCriticalPriorityClassName)
	}

	daemonSets, err := client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list daemonsets in %s: %v", namespace, err)
	}
	for i := range daemonSets.Items {
		ds := &daemonSets.Items[i]
		if ds.Spec.Template.Spec.PriorityClassName != "" {
			continue
		}
		if isHelmManagedWorkload(&ds.ObjectMeta) {
			logger.Infof("skipping Helm-managed daemonset %s/%s priorityClassName patch", namespace, ds.Name)
			continue
		}
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			_, patchErr := client.AppsV1().DaemonSets(namespace).Patch(
				ctx, ds.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{},
			)
			return patchErr
		}); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("failed to patch daemonset %s/%s priorityClassName: %v", namespace, ds.Name, err)
		}
		logger.Infof("patched daemonset %s/%s priorityClassName to %s", namespace, ds.Name, systemClusterCriticalPriorityClassName)
	}

	statefulSets, err := client.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list statefulsets in %s: %v", namespace, err)
	}
	for i := range statefulSets.Items {
		sts := &statefulSets.Items[i]
		if sts.Spec.Template.Spec.PriorityClassName != "" {
			continue
		}
		if isHelmManagedWorkload(&sts.ObjectMeta) {
			logger.Infof("skipping Helm-managed statefulset %s/%s priorityClassName patch", namespace, sts.Name)
			continue
		}
		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			_, patchErr := client.AppsV1().StatefulSets(namespace).Patch(
				ctx, sts.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{},
			)
			return patchErr
		}); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return fmt.Errorf("failed to patch statefulset %s/%s priorityClassName: %v", namespace, sts.Name, err)
		}
		logger.Infof("patched statefulset %s/%s priorityClassName to %s", namespace, sts.Name, systemClusterCriticalPriorityClassName)
	}

	return nil
}
