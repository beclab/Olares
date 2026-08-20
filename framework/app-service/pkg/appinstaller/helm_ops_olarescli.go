package appinstaller

import (
	"context"

	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/olarescli"
	"github.com/beclab/Olares/framework/app-service/pkg/users"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// EnsureOlaresCLICredential mints (or reuses) the owner's long-lived Olares
// credential for an app that declares permission.loginOlaresCLI, and writes
// it into the app's namespace for the pod webhook to pick up.
//
// It runs before the chart is installed so the Secret exists by the time the
// first pod is admitted; a pod that starts without it would come up logged
// out and stay that way until it restarted.
//
// Failures are fatal to the install: an app that asked for a credential and
// silently came up without one is worse than an install that says why it
// stopped.
func (h *HelmOps) EnsureOlaresCLICredential() error {
	if !h.app.LoginOlaresCLI {
		return nil
	}

	store, err := h.olaresCLIStore()
	if err != nil {
		return err
	}
	client, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		return err
	}

	if err = ensureNamespace(h.ctx, client, h.app.Namespace); err != nil {
		return err
	}
	zone, err := kubesphere.GetUserZone(h.ctx, h.app.OwnerName)
	if err != nil {
		return err
	}
	olaresID := string(users.NewOlaresName(h.app.OwnerName, zone))
	_, err = store.EnsureCredential(h.ctx, h.app.AppName, h.app.OwnerName, olaresID, h.app.Namespace)
	return err
}

// ReleaseOlaresCLICredential revokes the app's grant and removes the Secret.
//
// Uninstall must not fail on this: the release is already gone by the time it
// runs, and refusing to finish teardown because lldap was briefly unreachable
// would leave the app stuck. A grant that outlives its app is a real (if
// bounded) exposure, so the failure is logged loudly rather than swallowed.
func (h *HelmOps) ReleaseOlaresCLICredential() {
	if !h.app.LoginOlaresCLI {
		return
	}
	store, err := h.olaresCLIStore()
	if err != nil {
		klog.Errorf("Failed to build olares-cli client while uninstalling %s: %v", h.app.AppName, err)
		return
	}
	if err = store.Release(h.ctx, h.app.AppName, h.app.OwnerName, h.app.Namespace); err != nil {
		klog.Errorf("Failed to revoke olares-cli credential of app %s (user %s); "+
			"the grant may outlive the app and has to be revoked by hand: %v",
			h.app.AppName, h.app.OwnerName, err)
	}
}

func (h *HelmOps) olaresCLIStore() (*olarescli.Store, error) {
	client, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		return nil, err
	}
	// app-service's own ServiceAccount token, not h.token: lldap authorizes
	// the derive endpoints by TokenReview against a list of platform service
	// accounts, and h.token belongs to the user's namespace.
	saToken, err := utils.GetServerServiceAccountToken()
	if err != nil {
		return nil, err
	}
	return olarescli.NewStore(client, olarescli.NewClient(saToken)), nil
}

// ensureNamespace creates the app namespace when it does not exist yet. On the
// v1 install path helm is what normally creates it, which is too late for
// anything that has to be in place before the first pod is admitted.
func ensureNamespace(ctx context.Context, client kubernetes.Interface, namespace string) error {
	_, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}
	_, err = client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespace,
			Labels: map[string]string{"name": namespace},
		},
	}, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
