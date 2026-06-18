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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
)

type upgrader_1_12_7_20260617 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260617) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260617")
}

func (u upgrader_1_12_7_20260617) UpgradeSystemComponents() []task.Interface {
	tasks := make([]task.Interface, 0)
	tasks = append(tasks, upgradeNodeExporter()...)
	tasks = append(tasks, retagLegacyAMDGPUImage()...)

	tasks = append(tasks, u.upgraderBase.UpgradeSystemComponents()...)
	tasks = append(tasks, &task.LocalTask{
		Name:   "PatchL4BFLProxyProbePort",
		Action: new(patchL4BFLProxyProbePort),
		Retry:  3,
		Delay:  5 * time.Second,
	})
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260617{})
}

// patchL4BFLProxyProbePort updates the livenessProbe and readinessProbe
// httpGet port of the l4-bfl-proxy deployment in the os-network namespace
// from the legacy 8081 to 18081 to match the probe address now exposed
// by the new l4-bfl-proxy binary.
type patchL4BFLProxyProbePort struct {
	common.KubeAction
}

func (a *patchL4BFLProxyProbePort) Execute(_ connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %v", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	patch := []byte(`{
		"spec": {
			"template": {
				"spec": {
					"containers": [
						{
							"name": "proxy",
							"livenessProbe": {
								"httpGet": {
									"port": 18081
								}
							},
							"readinessProbe": {
								"httpGet": {
									"port": 18081
								}
							}
						}
					]
				}
			}
		}
	}`)

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := client.AppsV1().Deployments("os-network").Patch(
			ctx,
			"l4-bfl-proxy",
			types.StrategicMergePatchType,
			patch,
			metav1.PatchOptions{},
		)
		return err
	})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("l4-bfl-proxy deployment not found in os-network, skipping probe port patch")
			return nil
		}
		return fmt.Errorf("failed to patch l4-bfl-proxy deployment probe port: %v", err)
	}

	logger.Infof("patched l4-bfl-proxy deployment livenessProbe/readinessProbe port to 18081")
	return nil
}
