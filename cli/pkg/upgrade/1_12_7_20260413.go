package upgrade

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type upgrader_1_12_7_20260413 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_7_20260413) Version() *semver.Version {
	return semver.MustParse("1.12.7-20260413")
}

func (u upgrader_1_12_7_20260413) UpgradeSystemComponents() []task.Interface {
	tasks := append([]task.Interface{
		&task.LocalTask{
			Name:   "PatchNodeAffinityToRequired",
			Action: new(patchNodeAffinityToRequired),
			Retry:  3,
			Delay:  5 * time.Second,
		},
	}, u.upgraderBase.UpgradeSystemComponents()...)
	return tasks
}

func init() {
	registerDailyUpgrader(upgrader_1_12_7_20260413{})
}

type patchNodeAffinityToRequired struct {
	common.KubeAction
}

func (a *patchNodeAffinityToRequired) Execute(_ connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %v", err)
	}

	c, err := ctrlclient.New(config, ctrlclient.Options{})
	if err != nil {
		return fmt.Errorf("failed to create controller-runtime client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var citusList appsv1.StatefulSetList
	if err := c.List(ctx, &citusList, ctrlclient.MatchingLabels{"managed-by": "citus-operator"}); err != nil {
		return fmt.Errorf("failed to list citus statefulsets: %v", err)
	}
	for i := range citusList.Items {
		sts := &citusList.Items[i]
		if convertAffinityPreferredToRequired(&sts.Spec.Template.Spec) {
			if err := c.Update(ctx, sts); err != nil {
				return fmt.Errorf("failed to update citus statefulset %s/%s: %v", sts.Namespace, sts.Name, err)
			}
			logger.Infof("patched citus statefulset %s/%s node affinity to required", sts.Namespace, sts.Name)
		}
	}

	var kvrocksList appsv1.StatefulSetList
	if err := c.List(ctx, &kvrocksList, ctrlclient.MatchingLabels{"managed-by": "kvrocks-operator"}); err != nil {
		return fmt.Errorf("failed to list kvrocks statefulsets: %v", err)
	}
	for i := range kvrocksList.Items {
		sts := &kvrocksList.Items[i]
		if convertAffinityPreferredToRequired(&sts.Spec.Template.Spec) {
			if err := c.Update(ctx, sts); err != nil {
				return fmt.Errorf("failed to update kvrocks statefulset %s/%s: %v", sts.Namespace, sts.Name, err)
			}
			logger.Infof("patched kvrocks statefulset %s/%s node affinity to required", sts.Namespace, sts.Name)
		}
	}

	var l4 appsv1.Deployment
	l4Key := ctrlclient.ObjectKey{Namespace: "os-network", Name: "l4-bfl-proxy"}
	if err := c.Get(ctx, l4Key, &l4); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Infof("l4-bfl-proxy deployment not found, skipping")
			return nil
		}
		return fmt.Errorf("failed to get l4-bfl-proxy deployment: %v", err)
	}
	if convertAffinityPreferredToRequired(&l4.Spec.Template.Spec) {
		if err := c.Update(ctx, &l4); err != nil {
			return fmt.Errorf("failed to update l4-bfl-proxy deployment: %v", err)
		}
		logger.Infof("patched l4-bfl-proxy deployment node affinity to required")
	}

	return nil
}

func convertAffinityPreferredToRequired(podSpec *corev1.PodSpec) bool {
	if podSpec.Affinity == nil || podSpec.Affinity.NodeAffinity == nil {
		return false
	}
	na := podSpec.Affinity.NodeAffinity
	if len(na.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
		return false
	}
	if na.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		return false
	}

	terms := make([]corev1.NodeSelectorTerm, 0, len(na.PreferredDuringSchedulingIgnoredDuringExecution))
	for _, p := range na.PreferredDuringSchedulingIgnoredDuringExecution {
		for i, expr := range p.Preference.MatchExpressions {
			if expr.Key == "node-role.kubernetes.io/master" {
				p.Preference.MatchExpressions[i].Key = "node-role.kubernetes.io/control-plane"
			}
		}
		terms = append(terms, p.Preference)
	}

	na.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
		NodeSelectorTerms: terms,
	}
	na.PreferredDuringSchedulingIgnoredDuringExecution = nil
	return true
}
