package terminus

import (
	"context"
	"time"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	getConfigForPreparePKI     = ctrl.GetConfig
	newClientForPreparePKI     = func(cfg *rest.Config) (client.Client, error) { return client.New(cfg, client.Options{}) }
	loadOrCreateLinkerdPKIFunc = loadOrCreateLinkerdPKI
)

// PrepareLinkerdPKI prepares namespace ownership and PKI secret before Helm install.
type PrepareLinkerdPKI struct {
	common.KubeAction
}

func (t *PrepareLinkerdPKI) Execute(runtime connector.Runtime) error {
	if !appGatewayStackEnabled() {
		return nil
	}

	config, err := getConfigForPreparePKI()
	if err != nil {
		return err
	}
	c, err := newClientForPreparePKI(config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	return prepareLinkerdPKIWithClient(ctx, c, resolveInstallerDir(runtime))
}

func prepareLinkerdPKIWithClient(ctx context.Context, c client.Client, installerDir string) error {
	linkerdNS := agwconfigLinkerdNamespace()
	releaseNS := resolveAppGatewayNamespace()

	if err := ensureHelmOwnedNamespace(ctx, c, linkerdNS, appGatewaySystemReleaseName, releaseNS); err != nil {
		return err
	}
	_, err := loadOrCreateLinkerdPKIFunc(ctx, c, linkerdNS, appGatewaySystemPath(installerDir))
	if err != nil {
		return errors.Wrap(err, "prepare linkerd identity certificates")
	}
	return nil
}

func agwconfigLinkerdNamespace() string {
	return agwconfig.LinkerdNamespace()
}

func ensureHelmOwnedNamespace(ctx context.Context, c client.Client, namespace, releaseName, releaseNamespace string) error {
	var ns corev1.Namespace
	err := c.Get(ctx, types.NamespacedName{Name: namespace}, &ns)
	if apierrors.IsNotFound(err) {
		ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		if err := c.Create(ctx, &ns); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	patch := client.MergeFrom(ns.DeepCopy())
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}
	if ns.Annotations == nil {
		ns.Annotations = map[string]string{}
	}

	ns.Labels["app.kubernetes.io/managed-by"] = "Helm"
	ns.Annotations["meta.helm.sh/release-name"] = releaseName
	ns.Annotations["meta.helm.sh/release-namespace"] = releaseNamespace

	return c.Patch(ctx, &ns, patch)
}
