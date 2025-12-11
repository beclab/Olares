package prodiverregistry

import (
	"context"
	"errors"

	sysv1alpha1 "bytetrade.io/web3os/system-server/pkg/apis/sys/v1alpha1"
	"bytetrade.io/web3os/system-server/pkg/constants"
	clientset "bytetrade.io/web3os/system-server/pkg/generated/clientset/versioned"
	v1alpha1 "bytetrade.io/web3os/system-server/pkg/generated/listers/sys/v1alpha1"
	"bytetrade.io/web3os/system-server/pkg/utils"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
)

var ErrProviderNotFound = errors.New("provider not found")

type Registry struct {
	registryClientset clientset.Interface
	registryLister    v1alpha1.ProviderRegistryLister
	namespace         string
}

func NewRegistry(clientset clientset.Interface, lister v1alpha1.ProviderRegistryLister) *Registry {
	registry := &Registry{
		registryClientset: clientset,
		registryLister:    lister,
		namespace:         constants.MyNamespace,
	}

	return registry
}

func (r *Registry) GetProvider(_ context.Context, dataType, group, version string) (*sysv1alpha1.ProviderRegistry, error) {
	providerRegistries, err := r.registryLister.
		ProviderRegistries(r.namespace).
		List(labels.Everything())
	if err != nil {
		return nil, err
	}

	if len(providerRegistries) > 0 {

		for _, pr := range providerRegistries {
			if pr.Status.State == sysv1alpha1.Active {
				if pr.Spec.DataType == dataType &&
					pr.Spec.Group == group &&
					pr.Spec.Version == version &&
					pr.Spec.Kind == sysv1alpha1.Provider {
					return pr, nil
				}
			}
		}

	}

	return nil, ErrProviderNotFound
}

func (r *Registry) GetWatchers(ctx context.Context, dataType, group, version string) ([]*sysv1alpha1.ProviderRegistry, error) {
	providerRegistries, err := r.registryLister.
		ProviderRegistries(r.namespace).
		List(labels.Everything())
	if err != nil {
		return nil, err
	}

	prs := make([]*sysv1alpha1.ProviderRegistry, 0)
	if len(providerRegistries) > 0 {

		for _, pr := range providerRegistries {
			if pr.Status.State == sysv1alpha1.Active {
				if pr.Spec.DataType == dataType &&
					pr.Spec.Group == group &&
					pr.Spec.Version == version &&
					pr.Spec.Kind == sysv1alpha1.Watcher {
					klog.Info("watcher callbacks, ", utils.PrettyJSON(pr))

					prs = append(prs, pr.DeepCopy())
				}
			}
		}

	}

	return prs, nil
}
