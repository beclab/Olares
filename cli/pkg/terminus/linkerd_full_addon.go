package terminus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var linkerdControlPlaneDeployments = []string{
	"linkerd-destination",
	"linkerd-identity",
	"linkerd-proxy-injector",
}

// linkerdControlPlaneReadyTimeout covers identity restart after PKI sync plus
// guardian cold start on resource-constrained user hardware.
const linkerdControlPlaneReadyTimeout = 10 * time.Minute

// ValidateLinkerdDeployAssets ensures os-framework deploy includes Linkerd control plane YAML.
type ValidateLinkerdDeployAssets struct {
	common.KubeAction
}

func (t *ValidateLinkerdDeployAssets) Execute(runtime connector.Runtime) error {
	if !linkerdPostInstallEnabled() {
		return nil
	}
	return validateLinkerdControlPlaneDeploy(resolveInstallerDir(runtime))
}

func validateLinkerdControlPlaneDeploy(installerDir string) error {
	path := filepath.Join(osFrameworkDeployPath(installerDir), "linkerd-control-plane.yaml")
	st, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("installer incomplete for Linkerd control plane: missing %s", path)
	}
	if st.IsDir() {
		return fmt.Errorf("installer incomplete for Linkerd control plane: %s is a directory", path)
	}
	return nil
}

// SyncLinkerdPKIAndIdentity patches Linkerd identity secrets after os-framework apply.
type SyncLinkerdPKIAndIdentity struct {
	common.KubeAction
}

func (t *SyncLinkerdPKIAndIdentity) Execute(runtime connector.Runtime) error {
	if !linkerdPostInstallEnabled() {
		return nil
	}
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	linkerdNS := agwconfigLinkerdNamespace()
	mat, err := prepareLinkerdPKI(ctx, k8sClient, linkerdNS)
	if err != nil {
		return errors.Wrap(err, "prepare linkerd pki")
	}
	if err := syncLinkerdIdentitySecrets(ctx, k8sClient, linkerdNS, mat); err != nil {
		return errors.Wrap(err, "sync linkerd identity secrets")
	}
	if err := restartLinkerdIdentityIfNeeded(ctx, k8sClient, linkerdNS); err != nil {
		return errors.Wrap(err, "restart linkerd-identity")
	}
	return nil
}

// ApplyLinkerdMeshBootstrapNP applies bootstrap NP when app-service has not reconciled yet.
type ApplyLinkerdMeshBootstrapNP struct {
	common.KubeAction
}

func (t *ApplyLinkerdMeshBootstrapNP) Execute(runtime connector.Runtime) error {
	if !linkerdPostInstallEnabled() {
		return nil
	}
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	_, settings, err := utils.InitConfigForAppGateway(config, agwconfigLinkerdNamespace())
	if err != nil {
		return err
	}
	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	deployDir := osFrameworkDeployPath(resolveInstallerDir(runtime))
	return applyLinkerdMeshBootstrapNP(ctx, k8sClient, settings, deployDir)
}

// WaitLinkerdControlPlaneReady waits for core Linkerd deployments.
type WaitLinkerdControlPlaneReady struct {
	common.KubeAction
}

func (t *WaitLinkerdControlPlaneReady) Execute(runtime connector.Runtime) error {
	if !linkerdPostInstallEnabled() {
		return nil
	}
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), linkerdControlPlaneReadyTimeout)
	defer cancel()
	return waitLinkerdControlPlaneReady(ctx, k8sClient, agwconfigLinkerdNamespace(), linkerdControlPlaneReadyTimeout)
}

func waitLinkerdControlPlaneReady(ctx context.Context, c client.Client, ns string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ready := true
		for _, name := range linkerdControlPlaneDeployments {
			var dep appsv1.Deployment
			if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &dep); err != nil {
				if apierrors.IsNotFound(err) {
					ready = false
					break
				}
				return err
			}
			if dep.Status.ReadyReplicas < 1 {
				ready = false
				break
			}
		}
		if ready {
			var guardian appsv1.Deployment
			if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: "linkerd-pki-guardian"}, &guardian); err == nil {
				if guardian.Status.ReadyReplicas < 1 {
					ready = false
				}
			}
		}
		if ready {
			logger.InfoInstallationProgress("Linkerd control plane and PKI guardian are ready")
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	return fmt.Errorf("timeout waiting for Linkerd control plane in namespace %s", ns)
}
