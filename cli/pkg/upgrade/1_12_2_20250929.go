package upgrade

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type upgrader_1_12_2_20250929 struct {
	breakingUpgraderBase
}

func (u upgrader_1_12_2_20250929) Version() *semver.Version {
	return semver.MustParse("1.12.2-20250929")
}

func (u upgrader_1_12_2_20250929) preToPrepareForUpgrade() []task.Interface {
	return []task.Interface{
		&task.LocalTask{
			Name:   "takeOverKubeAppliedKbCRDByHelm",
			Action: &takeOverKbCRDByHelm{},
		},
	}
}

func (u upgrader_1_12_2_20250929) PrepareForUpgrade() []task.Interface {
	preTasks := u.preToPrepareForUpgrade()
	return append(preTasks, u.upgraderBase.PrepareForUpgrade()...)
}

type takeOverKbCRDByHelm struct {
	common.KubeAction
}

func (c *takeOverKbCRDByHelm) Execute(runtime connector.Runtime) error {
	releaseName := "settings"
	releaseNamespace := "default"

	cfg, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	apiExtClient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// We only operate on apiextensions.k8s.io/v1 CRDs
	crdList, err := apiExtClient.ApiextensionsV1().CustomResourceDefinitions().List(ctx, meta.ListOptions{LabelSelector: "app.kubernetes.io/name=kubeblocks"})
	if err != nil {
		return fmt.Errorf("list CRDs failed: %w", err)
	}

	for i := range crdList.Items {
		crd := crdList.Items[i].DeepCopy()
		if crd.Annotations == nil {
			crd.Annotations = map[string]string{}
		}
		if crd.Labels == nil {
			crd.Labels = map[string]string{}
		}
		changed := false
		if crd.Annotations["meta.helm.sh/release-name"] == "" {
			crd.Annotations["meta.helm.sh/release-name"] = releaseName
			changed = true
		}
		if crd.Annotations["meta.helm.sh/release-namespace"] == "" {
			crd.Annotations["meta.helm.sh/release-namespace"] = releaseNamespace
			changed = true
		}
		if crd.Labels["app.kubernetes.io/managed-by"] == "" {
			crd.Labels["app.kubernetes.io/managed-by"] = "Helm"
			changed = true
		}

		if !changed {
			continue
		}
		_, err = apiExtClient.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, crd, meta.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update CRD %s, %w", crd.Name, err)
		}
	}

	return nil
}

func init() {
	registerDailyUpgrader(upgrader_1_12_2_20250929{})
}
