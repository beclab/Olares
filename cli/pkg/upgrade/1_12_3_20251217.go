package upgrade

import (
	"context"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"

	"github.com/Masterminds/semver/v3"
	apixclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type upgrader_1_12_3_20251217 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_3_20251217) Version() *semver.Version {
	return semver.MustParse("1.12.3-20251217")
}

func (u upgrader_1_12_3_20251217) PrepareForUpgrade() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks,
		&task.LocalTask{
			Name:   "DeleteArgoProjV1alpha1CRDs",
			Action: new(deleteArgoProjV1alpha1CRDs),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	)
	tasks = append(tasks, u.upgraderBase.PrepareForUpgrade()...)
	return tasks
}

func (u upgrader_1_12_3_20251217) NeedRestart() bool {
	return true
}

func (u upgrader_1_12_3_20251217) UpdateOlaresVersion() []task.Interface {
	var tasks []task.Interface
	tasks = append(tasks,
		&task.LocalTask{
			Name:   "UpgradeGPUDriver",
			Action: new(upgradeGPUDriverIfNeeded),
		},
	)
	tasks = append(tasks, u.upgraderBase.UpdateOlaresVersion()...)
	tasks = append(tasks,
		&task.LocalTask{
			Name:   "RebootIfNeeded",
			Action: new(rebootIfNeeded),
		},
	)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_3_20251217{})
}

type deleteArgoProjV1alpha1CRDs struct {
	common.KubeAction
}

func (a *deleteArgoProjV1alpha1CRDs) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %s", err)
	}
	client, err := apixclientset.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create crd client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	crds, err := client.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list CRDs: %v", err)
	}

	for _, crd := range crds.Items {
		if crd.Spec.Group != "argoproj.io" {
			continue
		}
		if crd.Annotations["meta.helm.sh/release-name"] != "knowledge" {
			continue
		}
		if err := client.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, crd.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete CRD %s: %v", crd.Name, err)
		}
	}

	return nil
}
