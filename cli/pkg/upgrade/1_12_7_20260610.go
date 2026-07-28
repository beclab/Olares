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

type upgrader_1_12_7_20260610 struct {
	upgrader_1_12_7_20260605
}

const reverseProxyAgentImage = "beclab/reverse-proxy:v0.1.11"

func (u upgrader_1_12_7_20260610) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260610")
}

func (u upgrader_1_12_7_20260610) UpgradeSystemComponents() []task.Interface {
	pre := []task.Interface{
		&task.LocalTask{
			Name:   "PatchCoreDNSControlPlaneAffinity",
			Action: new(patchCoreDNSControlPlaneAffinity),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	}
	return append(pre, u.upgrader_1_12_7_20260605.UpgradeSystemComponents()...)
}

func (u upgrader_1_12_7_20260610) PrepareForUpgrade() []task.Interface {
	pre := []task.Interface{
		&task.LocalTask{
			Name:   "UpgradeUserReverseProxyAgent",
			Action: new(upgradeUserReverseProxyAgent),
			Retry:  5,
			Delay:  10 * time.Second,
		},
	}
	return append(pre, u.upgraderBase.PrepareForUpgrade()...)
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260610{})
}

// patchCoreDNSControlPlaneAffinity patches the existing coredns deployment
// in kube-system so that it can only be scheduled onto control-plane (master)
// nodes, matching the affinity now baked into the install template.
// It uses the new "node-role.kubernetes.io/control-plane" label with the
// Exists operator as a hard scheduling requirement.
type patchCoreDNSControlPlaneAffinity struct {
	common.KubeAction
}

func (a *patchCoreDNSControlPlaneAffinity) Execute(_ connector.Runtime) error {
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
					"affinity": {
						"nodeAffinity": {
							"requiredDuringSchedulingIgnoredDuringExecution": {
								"nodeSelectorTerms": [
									{
										"matchExpressions": [
											{
												"key": "node-role.kubernetes.io/control-plane",
												"operator": "Exists"
											}
										]
									}
								]
							}
						}
					}
				}
			}
		}
	}`)

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := client.AppsV1().Deployments("kube-system").Patch(
			ctx,
			"coredns",
			types.StrategicMergePatchType,
			patch,
			metav1.PatchOptions{},
		)
		return err
	})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("coredns deployment not found in kube-system, skipping control-plane affinity patch")
			return nil
		}
		return fmt.Errorf("failed to patch coredns deployment with control-plane affinity: %v", err)
	}

	logger.Infof("patched coredns deployment with control-plane node affinity")
	return nil
}
