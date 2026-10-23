package apiserver

import (
	"context"
	"fmt"
	"time"

	computevalidation "github.com/beclab/Olares/framework/app-service/pkg/compute/validation"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/api/pkg/generated/clientset/versioned"
	"github.com/beclab/api/pkg/generated/informers/externalversions"
	lister_v1alpha1 "github.com/beclab/api/pkg/generated/listers/app.bytetrade.io/v1alpha1"

	// upgrade removed from direct usage in handlers
	lifecycleapi "github.com/beclab/Olares/framework/app-service/pkg/applifecycle/apiserver"
	"github.com/beclab/Olares/framework/app-service/pkg/users/userspace/v1"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"github.com/beclab/Olares/framework/app-service/pkg/webhook"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Handler include several fields that used for managing interactions with associated services.
type Handler struct {
	kubeHost           string
	serviceCtx         context.Context
	userspaceManager   *userspace.Manager
	kubeConfig         *rest.Config // helm's kubeConfig. TODO: insecure
	sidecarWebhook     *webhook.Webhook
	ctrlClient         client.Client
	informer           externalversions.SharedInformerFactory
	appLister          lister_v1alpha1.ApplicationLister
	appmgrLister       lister_v1alpha1.ApplicationManagerLister
	appSynced          cache.InformerSynced
	appmgrSynced       cache.InformerSynced
	opController       *OpController
	preflightCollector preflightSnapshotCollector
	preflightIsAdmin   func(context.Context, string) (bool, error)

	// opDispatcher is the seam through which state-machine-aware
	// (v2) HTTP handlers deliver a synthetic TriggerOp to the
	// applifecycle Runner. Set to a non-nil implementation
	// (usually *lifecycleapi.RunnerDispatcher) once main.go wires
	// up the Runner; nil in tests and in the legacy hot path
	// (which reaches past the state machine to write Status
	// directly). See pkg/applifecycle/apiserver/doc.go for
	// the reasoning.
	//
	// v2 handlers MUST guard on opDispatcher != nil and fall
	// through to the legacy path when it is nil, so wiring can
	// be flipped incrementally without a big-bang cutover.
	opDispatcher lifecycleapi.OpDispatcher
}

type handlerBuilder struct {
	ctx          context.Context
	ksHost       string
	kubeConfig   *rest.Config
	ctrlClient   client.Client
	informer     externalversions.SharedInformerFactory
	opDispatcher lifecycleapi.OpDispatcher
	err          error
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

// WithOpDispatcher registers the applifecycle Runner adapter that
// v2 HTTP handlers use to deliver TriggerOp actions. Optional: when
// nil (the default), v2 handlers fall through to the legacy Status-
// write path so wiring can be flipped one binary at a time.
//
// Wiring invariant (enforced by main.go): the dispatcher's
// underlying Requeuer MUST be the same instance the Runner uses
// internally for CallWorkload callbacks, so a single workqueue
// drives every reconcile.
func (b *handlerBuilder) WithOpDispatcher(d lifecycleapi.OpDispatcher) *handlerBuilder {
	b.opDispatcher = d
	return b
}

func (b *handlerBuilder) WithAppInformer() *handlerBuilder {
	if b.err != nil {
		return b
	}
	appClient, err := versioned.NewForConfig(b.kubeConfig)
	if err != nil {
		b.err = fmt.Errorf("create app clientset for informer: %w", err)
		return b
	}

	b.informer = externalversions.NewSharedInformerFactory(appClient, 10*time.Minute)
	return b
}

func (b *handlerBuilder) Build() (*Handler, error) {
	if b.err != nil {
		return nil, b.err
	}
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
	err = wh.CreateOrUpdateTLSReplicaMountValidatingWebhook()
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
	err = wh.CreateOrUpdateCliCredentialMutatingWebhook()
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
	err = wh.DeleteAppManagerMutatingWebhook()
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
	err = wh.CreateOrUpdateMacvlanInitMutatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdateMacvlanAnnotationValidatingWebhook()
	if err != nil {
		return nil, err
	}
	err = wh.CreateOrUpdatePodArchNodeSelectorMutatingWebhook()
	if err != nil {
		return nil, err
	}

	return &Handler{
		kubeHost:           b.ksHost,
		serviceCtx:         b.ctx,
		kubeConfig:         b.kubeConfig,
		userspaceManager:   userspace.NewManager(b.ctx),
		sidecarWebhook:     wh,
		ctrlClient:         b.ctrlClient,
		informer:           b.informer,
		appLister:          b.informer.App().V1alpha1().Applications().Lister(),
		appmgrLister:       b.informer.App().V1alpha1().ApplicationManagers().Lister(),
		appSynced:          b.informer.App().V1alpha1().Applications().Informer().HasSynced,
		appmgrSynced:       b.informer.App().V1alpha1().ApplicationManagers().Informer().HasSynced,
		opController:       NewQueue(b.ctx),
		preflightCollector: computevalidation.NewPreflightSnapshotCollector(b.ctrlClient),
		preflightIsAdmin: func(ctx context.Context, user string) (bool, error) {
			return kubesphere.IsAdmin(ctx, b.kubeConfig, user)
		},
		opDispatcher: b.opDispatcher,
	}, nil

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
