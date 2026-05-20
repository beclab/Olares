package terminus

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// finalizeDemoMeshDebug rolls demo Deployments so linkerd-proxy is present (dev/debug only).
// Release installs keep defaults.yaml demo.meshDebug=false; override with APP_GATEWAY_DEMO_MESH_DEBUG=1 locally.
func finalizeDemoMeshDebug(ctx context.Context, c client.Client, ns string, def agwconfig.Defaults) error {
	if !demoMeshDebugEnabled(def) {
		return nil
	}
	logger.InfoInstallationProgress("app-gateway mesh (debug): rolling demo workloads (curl-client, http-echo) for linkerd-proxy ...")
	if err := rolloutRestartDemoDeployments(ctx, c, ns); err != nil {
		return err
	}
	return waitDemoMeshed(ctx, c, ns, 3*time.Minute)
}

func demoMeshDebugEnabled(def agwconfig.Defaults) bool {
	if v := os.Getenv("APP_GATEWAY_DEMO_MESH_DEBUG"); v == "1" || strings.EqualFold(v, "true") {
		return def.Demo.Enabled
	}
	return def.DemoMeshDebugEnabled()
}

var demoMeshDeploymentNames = []string{"curl-client", "http-echo"}

func rolloutRestartDemoDeployments(ctx context.Context, c client.Client, ns string) error {
	now := time.Now().Format(time.RFC3339)
	for _, name := range demoMeshDeploymentNames {
		var dep appsv1.Deployment
		if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &dep); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return err
		}
		patch := client.MergeFrom(dep.DeepCopy())
		if dep.Spec.Template.Annotations == nil {
			dep.Spec.Template.Annotations = map[string]string{}
		}
		dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = now
		if err := c.Patch(ctx, &dep, patch); err != nil {
			return err
		}
	}
	return nil
}

func waitDemoMeshed(ctx context.Context, c client.Client, ns string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastReason string
	for time.Now().Before(deadline) {
		ok, reason, err := demoMeshReady(ctx, c, ns)
		if err != nil {
			return err
		}
		if ok {
			logger.InfoInstallationProgress("app-gateway mesh (debug): demo pods have linkerd-proxy and are Ready")
			return nil
		}
		lastReason = reason
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	return fmt.Errorf("timeout waiting for meshed demo pods in %s (last=%s)", ns, lastReason)
}

func demoMeshReady(ctx context.Context, c client.Client, ns string) (bool, string, error) {
	found := 0
	for _, name := range demoMeshDeploymentNames {
		var dep appsv1.Deployment
		if err := c.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, &dep); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return false, "", err
		}
		found++
		var pods corev1.PodList
		if err := c.List(ctx, &pods, client.InNamespace(ns), client.MatchingLabels(dep.Spec.Selector.MatchLabels)); err != nil {
			return false, "", err
		}
		if len(pods.Items) == 0 {
			return false, name + " has no pods", nil
		}
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodFailed {
				return false, pod.Name + " failed", nil
			}
			hasProxy := false
			allReady := true
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.Name == "linkerd-proxy" {
					hasProxy = true
					if !cs.Ready {
						allReady = false
					}
				}
			}
			if !hasProxy {
				return false, pod.Name + " missing linkerd-proxy", nil
			}
			if !allReady {
				return false, pod.Name + " not ready", nil
			}
		}
	}
	if found == 0 {
		return false, "no demo deployments (curl-client/http-echo)", nil
	}
	return true, "", nil
}
