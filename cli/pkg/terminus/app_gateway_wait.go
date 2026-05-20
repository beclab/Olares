package terminus

import (
	"context"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const egControlPlaneDeployName = "envoy-gateway"

// envoyGatewayCRDsPresent reports whether Envoy Gateway / Gateway API CRDs are already registered.
func envoyGatewayCRDsPresent(cfg *rest.Config) bool {
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return false
	}
	for _, gv := range []string{
		"gateway.envoyproxy.io/v1alpha1",
		"gateway.networking.k8s.io/v1",
	} {
		if _, err := dc.ServerResourcesForGroupVersion(gv); err != nil {
			return false
		}
	}
	return true
}

func ensureAppGatewayNamespace(ctx context.Context, c client.Client, ns string) error {
	var existing corev1.Namespace
	err := c.Get(ctx, types.NamespacedName{Name: ns}, &existing)
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}
	logger.InfoInstallationProgress(fmt.Sprintf("Creating namespace %s for Envoy Gateway ...", ns))
	return c.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
			Labels: map[string]string{
				"app.kubernetes.io/name": "app-gateway",
			},
			Annotations: map[string]string{
				"bytetrade.io/ns-type": "platform",
			},
		},
	})
}

func waitEnvoyGatewayControlPlaneReady(ctx context.Context, c client.Client, ns string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastMsg string
	attempt := 0

	for time.Now().Before(deadline) {
		attempt++
		reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		var eg appsv1.Deployment
		err := c.Get(reqCtx, types.NamespacedName{Namespace: ns, Name: egControlPlaneDeployName}, &eg)
		cancel()

		if err == nil {
			if eg.Status.ReadyReplicas >= 1 {
				logger.InfoInstallationProgress("Envoy Gateway control plane is ready")
				return nil
			}
			lastMsg = fmt.Sprintf("deployment/%s exists but ReadyReplicas=%d (want >=1)",
				egControlPlaneDeployName, eg.Status.ReadyReplicas)
		} else if apierrors.IsNotFound(err) {
			lastMsg = fmt.Sprintf("deployment/%s not found in namespace %s (helm release eg may not have created it yet)",
				egControlPlaneDeployName, ns)
			if attempt == 1 || attempt%6 == 0 {
				listCtx, listCancel := context.WithTimeout(ctx, 15*time.Second)
				hint, _ := listAppGatewayDeployments(listCtx, c, ns)
				listCancel()
				if hint != "" {
					lastMsg += "; deployments in ns: " + hint
				}
			}
		} else if apierrors.IsTimeout(err) || reqCtx.Err() == context.DeadlineExceeded {
			lastMsg = fmt.Sprintf("kubernetes API timeout talking to apiserver while getting deployment/%s: %v",
				egControlPlaneDeployName, err)
		} else {
			lastMsg = fmt.Sprintf("get deployment/%s: %v", egControlPlaneDeployName, err)
		}

		if attempt == 1 || attempt%6 == 0 {
			logger.InfoInstallationProgress(fmt.Sprintf("Waiting for Envoy Gateway control plane (%s)", lastMsg))
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout after %s waiting for %s/%s: %s", timeout, ns, egControlPlaneDeployName, lastMsg)
		case <-time.After(5 * time.Second):
		}
	}
	return fmt.Errorf("timeout after %s waiting for %s/%s: %s", timeout, ns, egControlPlaneDeployName, lastMsg)
}

func listAppGatewayDeployments(ctx context.Context, c client.Client, ns string) (string, error) {
	var list appsv1.DeploymentList
	if err := c.List(ctx, &list, client.InNamespace(ns)); err != nil {
		return "", err
	}
	if len(list.Items) == 0 {
		return "(none)", nil
	}
	names := make([]string, 0, len(list.Items))
	for _, d := range list.Items {
		names = append(names, d.Name)
	}
	return fmt.Sprintf("%v", names), nil
}
