package terminus

import (
	"context"
	"fmt"
	"time"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const egDataPlaneGatewayLabel = "gateway.envoyproxy.io/owning-gateway-name"

// finalizeAppGatewayMesh strips namespace-level inject and rolls EG data-plane so EnvoyProxy pod annotations apply.
func finalizeAppGatewayMesh(ctx context.Context, c client.Client, ns string, def agwconfig.Defaults) error {
	if err := stripNamespaceLinkerdInject(ctx, c, ns); err != nil {
		return err
	}
	if !def.MeshLinkerdEnabled() || !def.EnvoyProxy.Enabled {
		logger.InfoInstallationProgress("app-gateway mesh: EG data-plane linkerd injection disabled in defaults.yaml")
		return nil
	}
	gwName := def.Gateway.Name
	if gwName == "" {
		gwName = "app-gateway"
	}
	logger.InfoInstallationProgress("app-gateway mesh: rolling EG data-plane for linkerd-proxy injection ...")
	if err := rolloutRestartEGDataPlane(ctx, c, ns, gwName); err != nil {
		return err
	}
	// Mesh readiness is enforced by WaitAppGatewayDataPlaneMeshed (upgrade/install pipeline, Retry: 30).
	return finalizeDemoMeshDebug(ctx, c, ns, def)
}

func stripNamespaceLinkerdInject(ctx context.Context, c client.Client, ns string) error {
	var existing corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: ns}, &existing); err != nil {
		return err
	}
	if existing.Annotations == nil {
		return nil
	}
	if _, ok := existing.Annotations["linkerd.io/inject"]; !ok {
		return nil
	}
	patch := client.MergeFrom(existing.DeepCopy())
	delete(existing.Annotations, "linkerd.io/inject")
	logger.InfoInstallationProgress(fmt.Sprintf("Removed %s annotation linkerd.io/inject (EG mesh uses EnvoyProxy pod template only)", ns))
	return c.Patch(ctx, &existing, patch)
}

func rolloutRestartEGDataPlane(ctx context.Context, c client.Client, ns, gatewayName string) error {
	var list appsv1.DeploymentList
	if err := c.List(ctx, &list, client.InNamespace(ns), client.MatchingLabels{
		egDataPlaneGatewayLabel: gatewayName,
	}); err != nil {
		return err
	}
	if len(list.Items) == 0 {
		return fmt.Errorf("no EG data-plane Deployment for gateway %q in %s (Gateway/EnvoyProxy still reconciling)", gatewayName, ns)
	}
	now := time.Now().Format(time.RFC3339)
	for i := range list.Items {
		dep := &list.Items[i]
		patch := client.MergeFrom(dep.DeepCopy())
		if dep.Spec.Template.Annotations == nil {
			dep.Spec.Template.Annotations = map[string]string{}
		}
		dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = now
		if err := c.Patch(ctx, dep, patch); err != nil {
			return err
		}
	}
	return nil
}

func waitEGDataPlaneMeshed(ctx context.Context, c client.Client, ns, gatewayName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastReason string
	for time.Now().Before(deadline) {
		ok, reason, err := egDataPlaneMeshReady(ctx, c, ns, gatewayName)
		if err != nil {
			return err
		}
		if ok {
			logger.InfoInstallationProgress("app-gateway mesh: EG data-plane pods have linkerd-proxy and are Ready")
			return nil
		}
		lastReason = reason
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	return fmt.Errorf("timeout waiting for meshed EG data-plane in %s (gateway=%s, last=%s)", ns, gatewayName, lastReason)
}

func egDataPlaneMeshReady(ctx context.Context, c client.Client, ns, gatewayName string) (bool, string, error) {
	var pods corev1.PodList
	if err := c.List(ctx, &pods, client.InNamespace(ns), client.MatchingLabels{
		egDataPlaneGatewayLabel: gatewayName,
	}); err != nil {
		return false, "", err
	}
	if len(pods.Items) == 0 {
		return false, "no data-plane pods", nil
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodFailed {
			return false, pod.Name + " failed", nil
		}
		hasProxy := false
		hasEnvoy := false
		allReady := true
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Name == "linkerd-proxy" {
				hasProxy = true
				if !cs.Ready {
					allReady = false
				}
			}
			if cs.Name == "envoy" {
				hasEnvoy = true
				if !cs.Ready {
					allReady = false
				}
			}
		}
		if !hasEnvoy {
			return false, pod.Name + " missing envoy container", nil
		}
		if !hasProxy {
			return false, pod.Name + " missing linkerd-proxy", nil
		}
		if !allReady {
			return false, pod.Name + " not ready", nil
		}
	}
	return true, "", nil
}

// WaitAppGatewayDataPlaneMeshed waits until EG data-plane pods include a ready linkerd-proxy when mesh is enabled.
type WaitAppGatewayDataPlaneMeshed struct {
	common.KubeAction
}

func (t *WaitAppGatewayDataPlaneMeshed) Execute(runtime connector.Runtime) error {
	if !appGatewayStackEnabled() {
		return nil
	}
	def, err := agwconfig.Load()
	if err != nil {
		return err
	}
	if !def.MeshLinkerdEnabled() || !def.EnvoyProxy.Enabled {
		return nil
	}
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	c, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	ns := resolveAppGatewayNamespace()
	gwName := def.Gateway.Name
	if gwName == "" {
		gwName = "app-gateway"
	}
	return waitEGDataPlaneMeshed(ctx, c, ns, gwName, 5*time.Minute)
}
