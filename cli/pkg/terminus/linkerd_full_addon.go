package terminus

import (
	"context"
	"fmt"
	"strings"
	"time"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const linkerdPKIGuardianDeployment = "linkerd-pki-guardian"

var linkerdControlPlaneDeployments = []string{
	"linkerd-destination",
	"linkerd-identity",
	"linkerd-proxy-injector",
}

// linkerdControlPlaneReadyTimeout covers identity restart after PKI sync plus
// guardian cold start on resource-constrained user hardware.
const linkerdControlPlaneReadyTimeout = 10 * time.Minute

const linkerdControlPlanePollInterval = 5 * time.Second

// SyncLinkerdPKIAndIdentity patches Linkerd identity secrets after os-framework apply.
type SyncLinkerdPKIAndIdentity struct {
	common.KubeAction
}

func (t *SyncLinkerdPKIAndIdentity) Execute(_ connector.Runtime) error {
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
	linkerdNS := agwconfig.LinkerdNamespace()
	mat, err := prepareLinkerdPKI(ctx, k8sClient, linkerdNS)
	if err != nil {
		return errors.Wrap(err, "prepare linkerd pki")
	}
	changed, err := syncLinkerdIdentitySecrets(ctx, k8sClient, linkerdNS, mat)
	if err != nil {
		return errors.Wrap(err, "sync linkerd identity secrets")
	}
	if changed {
		if err := markLinkerdControlPlaneRestartPending(ctx, k8sClient, linkerdNS); err != nil {
			return errors.Wrap(err, "mark linkerd control plane restart pending")
		}
	}
	if err := ensureLinkerdControlPlaneRestartedAfterPKISync(ctx, k8sClient, linkerdNS); err != nil {
		return err
	}
	return nil
}

func ensureLinkerdControlPlaneRestartedAfterPKISync(ctx context.Context, c client.Client, ns string) error {
	required, syncGeneration, err := linkerdControlPlaneRestartRequired(ctx, c, ns)
	if err != nil {
		return errors.Wrap(err, "check linkerd control plane restart required")
	}
	if !required {
		return nil
	}
	if err := restartLinkerdControlPlaneAfterPKISync(ctx, c, ns, syncGeneration); err != nil {
		return errors.Wrap(err, "restart linkerd control plane after pki sync")
	}
	if err := clearLinkerdControlPlaneRestartPending(ctx, c, ns); err != nil {
		return errors.Wrap(err, "clear linkerd control plane restart pending")
	}
	return nil
}

// WaitLinkerdControlPlaneReady waits for core Linkerd deployments.
type WaitLinkerdControlPlaneReady struct {
	common.KubeAction
}

func (t *WaitLinkerdControlPlaneReady) Execute(_ connector.Runtime) error {
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
	return waitLinkerdControlPlaneReadyWithPoll(
		ctx, k8sClient, agwconfig.LinkerdNamespace(), linkerdControlPlaneReadyTimeout, linkerdControlPlanePollInterval,
	)
}

func waitLinkerdControlPlaneReady(ctx context.Context, c client.Client, ns string, timeout time.Duration) error {
	return waitLinkerdControlPlaneReadyWithPoll(ctx, c, ns, timeout, linkerdControlPlanePollInterval)
}

func waitLinkerdControlPlaneReadyWithPoll(
	ctx context.Context, c client.Client, ns string, timeout, pollInterval time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pending, err := linkerdControlPlaneNotReady(ctx, c, ns)
		if err != nil {
			return err
		}
		if len(pending) == 0 {
			logger.InfoInstallationProgress("Linkerd control plane and PKI guardian are ready")
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
	pending, err := linkerdControlPlaneNotReady(ctx, c, ns)
	if err != nil {
		return err
	}
	return fmt.Errorf(
		"WaitLinkerdControlPlaneReady: timed out after %s waiting for Linkerd control plane deployments in namespace %s; not ready: %s",
		timeout, ns, strings.Join(pending, ", "),
	)
}

func linkerdControlPlaneNotReady(ctx context.Context, c client.Client, ns string) ([]string, error) {
	var pending []string
	for _, name := range linkerdControlPlaneDeployments {
		var dep appsv1.Deployment
		if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &dep); err != nil {
			if apierrors.IsNotFound(err) {
				pending = append(pending, name+" (not found)")
				continue
			}
			return nil, err
		}
		if dep.Status.ReadyReplicas < 1 {
			pending = append(pending, name)
		}
	}
	var guardian appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: linkerdPKIGuardianDeployment}, &guardian); err != nil {
		if apierrors.IsNotFound(err) {
			pending = append(pending, linkerdPKIGuardianDeployment+" (not found)")
		} else {
			return nil, err
		}
	} else if guardian.Status.ReadyReplicas < 1 {
		pending = append(pending, linkerdPKIGuardianDeployment)
	}
	return pending, nil
}
