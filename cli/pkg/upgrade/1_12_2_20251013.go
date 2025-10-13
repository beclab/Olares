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
	apixclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type upgrader_1_12_2_20251013 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_2_20251013) Version() *semver.Version {
	return semver.MustParse("1.12.2-20251013")
}

func (u upgrader_1_12_2_20251013) UpgradeSystemComponents() []task.Interface {
	pre := []task.Interface{
		&task.LocalTask{
			Name:   "DeleteOldSystemEnvsIfExists",
			Action: new(deleteSystemEnvsIfExists),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	}
	return append(pre, u.upgraderBase.UpgradeSystemComponents()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_2_20251013{})
}

type deleteSystemEnvsIfExists struct {
	common.KubeAction
}

func (d *deleteSystemEnvsIfExists) Execute(runtime connector.Runtime) error {
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
	_, err = apix.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, "systemenvs.sys.bytetrade.io", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Debugf("SystemEnv CRD not found, skipping deletion of SystemEnvs")
			return nil
		}
		return fmt.Errorf("failed to get SystemEnv CRD: %v", err)
	}

	scheme := kruntime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add systemenv scheme: %v", err)
	}
	c, err := ctrlclient.New(config, ctrlclient.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	var seList v1alpha1.SystemEnvList
	if err := c.List(ctx, &seList); err != nil {
		return fmt.Errorf("failed to list SystemEnvs: %v", err)
	}
	if len(seList.Items) == 0 {
		logger.Debugf("no SystemEnvs found to delete")
		return nil
	}
	for i := range seList.Items {
		se := &seList.Items[i]
		if err := c.Delete(ctx, se); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete SystemEnv %s: %v", se.Name, err)
		}
		logger.Debugf("deleted SystemEnv %s", se.Name)
	}
	return nil
}
