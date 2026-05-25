package upgrade

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apixclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"
)

// applyGPUBindingCRD upserts the gpu.bytetrade.io/v1alpha1 GPUBinding CRD into
// the cluster from the bundled HAMi chart on disk. This is necessary because
// `helm upgrade` does NOT update objects under a chart's `crds/` directory
// (only `helm install` does), so any schema change shipped via the chart
// won't propagate to an already-installed cluster without an explicit apply.
//
// The new schema adds spec.namespace / spec.owner, which the app-service
// migration step and the new HAMi scheduler both rely on; without applying
// the updated CRD first, the apiserver would prune those fields and the
// migration would silently produce empty Owner values on legacy bindings.
type applyGPUBindingCRD struct {
	common.KubeAction
}

const gpuBindingCRDRelPath = "wizard/config/gpu/hami/crds/gpu.bytetrade.io_gpubindings.yaml"

func (a *applyGPUBindingCRD) Execute(runtime connector.Runtime) error {
	crdPath := filepath.Join(runtime.GetInstallerDir(), gpuBindingCRDRelPath)
	data, err := os.ReadFile(crdPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Infof("GPUBinding CRD manifest not found at %s, skipping CRD upgrade", crdPath)
			return nil
		}
		return fmt.Errorf("read GPUBinding CRD manifest %s: %w", crdPath, err)
	}

	desired := &apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.Unmarshal(data, desired); err != nil {
		return fmt.Errorf("decode GPUBinding CRD manifest: %w", err)
	}
	if desired.Name == "" {
		return fmt.Errorf("GPUBinding CRD manifest at %s has no metadata.name", crdPath)
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("get rest config: %w", err)
	}
	client, err := apixclientset.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("create apiextensions client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	existing, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, desired.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// CRD not yet present (e.g. fresh GPU-less cluster being upgraded). It
		// will get installed by `helm install` later if/when GPU is enabled,
		// so just no-op here.
		logger.Infof("GPUBinding CRD %s not found in cluster, skipping CRD upgrade", desired.Name)
		return nil
	}
	if err != nil {
		return fmt.Errorf("get existing GPUBinding CRD: %w", err)
	}

	// Preserve resource version / status, replace spec + labels + annotations
	// with the desired manifest. We keep the cluster's status block to avoid
	// fighting with the apiextensions controller.
	updated := existing.DeepCopy()
	updated.Spec = desired.Spec
	if desired.Labels != nil {
		if updated.Labels == nil {
			updated.Labels = map[string]string{}
		}
		for k, v := range desired.Labels {
			updated.Labels[k] = v
		}
	}
	if desired.Annotations != nil {
		if updated.Annotations == nil {
			updated.Annotations = map[string]string{}
		}
		for k, v := range desired.Annotations {
			updated.Annotations[k] = v
		}
	}

	if _, err := client.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, updated, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("update GPUBinding CRD: %w", err)
	}
	logger.Infof("GPUBinding CRD %s updated to match bundled manifest", desired.Name)
	return nil
}
