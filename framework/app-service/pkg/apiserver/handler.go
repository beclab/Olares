package apiserver

import (
	"context"
	"fmt"
	"time"

	"bytetrade.io/web3os/app-service/pkg/generated/clientset/versioned"
	"bytetrade.io/web3os/app-service/pkg/generated/informers/externalversions"
	lister_v1alpha1 "bytetrade.io/web3os/app-service/pkg/generated/listers/app.bytetrade.io/v1alpha1"

	// upgrade removed from direct usage in handlers
	"bytetrade.io/web3os/app-service/pkg/users/userspace/v1"
	"bytetrade.io/web3os/app-service/pkg/utils"
	"bytetrade.io/web3os/app-service/pkg/webhook"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Handler include several fields that used for managing interactions with associated services.
type Handler struct {
	kubeHost         string
	serviceCtx       context.Context
	userspaceManager *userspace.Manager
	kubeConfig       *rest.Config // helm's kubeConfig. TODO: insecure
	sidecarWebhook   *webhook.Webhook
	ctrlClient       client.Client
	informer         externalversions.SharedInformerFactory
	appLister        lister_v1alpha1.ApplicationLister
	appmgrLister     lister_v1alpha1.ApplicationManagerLister
	appSynced        cache.InformerSynced
	appmgrSynced     cache.InformerSynced
	opController     *OpController
}

type handlerBuilder struct {
	ctx        context.Context
	ksHost     string
	kubeConfig *rest.Config
	ctrlClient client.Client
	informer   externalversions.SharedInformerFactory
}

func (b *handlerBuilder) WithKubesphereConfig(ksHost string) *handlerBuilder {
	b.ksHost = ksHost
	return b
}

func (b *handlerBuilder) WithContext(ctx context.Context) *handlerBuilder {
	b.ctx = ctx
	return b
}

func (b *handlerBuilder) WithKubernetesConfig(config *rest.Config) *handlerBuilder {
	b.kubeConfig = config
	return b
}

func (b *handlerBuilder) WithCtrlClient(client client.Client) *handlerBuilder {
	b.ctrlClient = client
	return b
}

func (b *handlerBuilder) WithAppInformer() *handlerBuilder {
	appClient, err := versioned.NewForConfig(b.kubeConfig)
	if err != nil {
		return nil
	}

	informer := externalversions.NewSharedInformerFactory(appClient, 10*time.Minute)
	b.informer = informer
	return b
}

func (b *handlerBuilder) Build() (*Handler, error) {
	wh, err := webhook.New(b.kubeConfig)
	if err != nil {
		return nil, err
	}

	err = wh.CreateOrUpdateSandboxMutatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdateAppNamespaceValidatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdateGpuLimitMutatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdateProviderRegistryValidatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.DeleteKubeletEvictionValidatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdateCronWorkflowMutatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdateRunAsUserMutatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdateAppLabelMutatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdateUserValidatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdateApplicationManagerMutatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdateApplicationManagerValidatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdateArgoResourceValidatingWebhook()
	if err != nil {
		return nil, err
	}

	return &Handler{
		kubeHost:         b.ksHost,
		serviceCtx:       b.ctx,
		kubeConfig:       b.kubeConfig,
		userspaceManager: userspace.NewManager(b.ctx),
		sidecarWebhook:   wh,
		ctrlClient:       b.ctrlClient,
		informer:         b.informer,
		appLister:        b.informer.App().V1alpha1().Applications().Lister(),
		appmgrLister:     b.informer.App().V1alpha1().ApplicationManagers().Lister(),
		appSynced:        b.informer.App().V1alpha1().Applications().Informer().HasSynced,
		appmgrSynced:     b.informer.App().V1alpha1().ApplicationManagers().Informer().HasSynced,
		opController:     NewQueue(b.ctx),
	}, err

}

func (h *Handler) Run(stopCh <-chan struct{}) error {
	h.informer.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, h.appSynced, h.appmgrSynced) {
		return fmt.Errorf("failed to wait for application caches to sync")
	}
	return nil
}

func (h *Handler) GetServerServiceAccountToken() string {
	return h.kubeConfig.BearerToken
}

func (h *Handler) GetUserServiceAccountToken(ctx context.Context, user string) (string, error) {
	kubeClient, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		klog.Errorf("Failed to create kube client: %v", err)
		return "", err
	}
	return utils.GetUserServiceAccountToken(ctx, kubeClient, user)
}
