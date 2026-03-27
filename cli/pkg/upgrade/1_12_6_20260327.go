package upgrade

import (
	"context"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"

	"github.com/Masterminds/semver/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
)

type upgrader_1_12_6_20260327 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_6_20260327) Version() *semver.Version {
	return semver.MustParse("1.12.6-20260327")
}

func (u upgrader_1_12_6_20260327) UpgradeSystemComponents() []task.Interface {
	pre := []task.Interface{
		&task.LocalTask{
			Name:   "UpdateL4DeploymentSpec",
			Action: new(updateL4DeploymentSpec),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	}
	return append(pre, u.upgraderBase.UpgradeSystemComponents()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_6_20260327{})
}

type updateL4DeploymentSpec struct {
	common.KubeAction
}

func (a *updateL4DeploymentSpec) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %s", err)
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
				"metadata": {
					"labels": {
						"tier": "bfl"
					}
				},
				"spec": {
					"containers": [
						{
							"name": "proxy",
							"command": [
								"/l4-bfl-proxy",
								"-v",
								"4",
								"-xds-tcp-idle-timeout",
								"1h",
								"-xds-http-stream-idle-timeout",
								"1h",
								"-xds-connect-timeout",
								"5s"
							],
							"livenessProbe": {
								"tcpSocket": null,
								"failureThreshold": 8,
								"httpGet": {
									"path": "/healthz",
									"port": 8081,
									"scheme": "HTTP"
								},
								"initialDelaySeconds": 3,
								"periodSeconds": 5,
								"successThreshold": 1,
								"timeoutSeconds": 10
							},
							"readinessProbe": {
								"tcpSocket": null,
								"failureThreshold": 5,
								"httpGet": {
									"path": "/readyz",
									"port": 8081,
									"scheme": "HTTP"
								},
								"periodSeconds": 3,
								"successThreshold": 1,
								"timeoutSeconds": 10
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
		return fmt.Errorf("failed to patch l4-bfl-proxy deployment: %v", err)
	}

	return nil
}
