package upgrade

import (
	"context"
	"fmt"
	"time"

	v1alpha1 "bytetrade.io/web3os/app-service/api/sys.bytetrade.io/v1alpha1"
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	corev1 "k8s.io/api/core/v1"
	apixclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type upgrader_1_12_2_20251020 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_2_20251020) Version() *semver.Version {
	return semver.MustParse("1.12.2-20251020")
}

// Ensure we remove legacy UserEnv objects before applying fresh ones
func (u upgrader_1_12_2_20251020) UpgradeSystemComponents() []task.Interface {
	pre := []task.Interface{
		&task.LocalTask{
			Name:   "DeleteUserEnvConfigMapIfExists",
			Action: new(deleteUserEnvConfigMapIfExists),
			Retry:  3,
			Delay:  5 * time.Second,
		},
		&task.LocalTask{
			Name:   "DeleteOldUserEnvsIfExists",
			Action: new(deleteUserEnvsIfExists),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	}
	return append(pre, u.upgraderBase.UpgradeSystemComponents()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_2_20251020{})
}

type deleteUserEnvConfigMapIfExists struct {
	common.KubeAction
}

func (d *deleteUserEnvConfigMapIfExists) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %s", err)
	}
	scheme := kruntime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add corev1 scheme: %v", err)
	}
	c, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var cm corev1.ConfigMap
	key := ctrlclient.ObjectKey{Namespace: common.NamespaceOsFramework, Name: "user-env"}
	if err := c.Get(ctx, key, &cm); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Debugf("user-env ConfigMap not found, skip deletion")
			return nil
		}
		return fmt.Errorf("failed to get user-env ConfigMap: %v", err)
	}
	if err := c.Delete(ctx, &cm); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete user-env ConfigMap: %v", err)
	}
	logger.Debugf("deleted user-env ConfigMap in namespace %s", common.NamespaceOsFramework)
	return nil
}

type deleteUserEnvsIfExists struct {
	common.KubeAction
}

func (d *deleteUserEnvsIfExists) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	apix, err := apixclientset.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create crd client: %v", err)
	}
	_, err = apix.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, "userenvs.sys.bytetrade.io", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Debugf("UserEnv CRD not found, skipping deletion of UserEnvs")
			return nil
		}
		return fmt.Errorf("failed to get UserEnv CRD: %v", err)
	}

	scheme := kruntime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add userenv scheme: %v", err)
	}
	c, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	var ueList v1alpha1.UserEnvList
	if err := c.List(ctx, &ueList); err != nil {
		return fmt.Errorf("failed to list UserEnvs: %v", err)
	}
	if len(ueList.Items) == 0 {
		logger.Debugf("no UserEnvs found to delete")
		return nil
	}
	for i := range ueList.Items {
		ue := &ueList.Items[i]
		if err := c.Delete(ctx, ue); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete UserEnv %s: %v", ue.Name, err)
		}
		logger.Debugf("deleted UserEnv %s", ue.Name)
	}
	return nil
}

func init() {
	registerDailyUpgrader(upgrader_1_12_2_20251020{})
}
