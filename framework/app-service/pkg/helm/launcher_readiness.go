package helm

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var ErrLauncherPending = errors.New("launcher provisioning is still pending")

type launcherRoute struct{ namespace, name, server string }

// WaitForTenantActivationPolicies checks the port-specific routes before the
// installer or tenant controller reports provisioning complete.
func WaitForTenantActivationPolicies(ctx context.Context, config *rest.Config, space string) error {
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	system := "user-system-" + strings.TrimPrefix(space, "user-space-")
	return waitForLauncherRoutes(ctx, client, []launcherRoute{
		{system, "activation-password", "systemserver-existing-28080"},
		{system, "activation-password-deny", "systemserver-existing-28080"},
		{system, "activation-provider", "systemserver-existing-28080"},
		{system, "activation-login", "systemserver-existing-9091"},
		{system, "activation-vault", "systemserver-existing-3000"},
	})
}

// WaitForLauncherPolicies waits for the Linkerd controller to attach the routes
// before the launcher can select the strict BFL Server.
func WaitForLauncherPolicies(ctx context.Context, config *rest.Config, space string, tenantAPI bool) error {
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	tenant := strings.TrimPrefix(space, "user-space-")
	routes := []launcherRoute{
		{space, "bfl-api-frontend-apps", "bfl-api"},
		{space, "bfl-api-login-password", "bfl-api"},
		{"os-framework", "vault-public-rpc", "vault-public-rpc"},
	}
	// Full app-service API enforcement has its own enrollment prerequisites.
	// Keep compatibility for installers that have not enabled that profile.
	server := schema.GroupVersionResource{Group: "policy.linkerd.io", Version: "v1beta3", Resource: "servers"}
	if _, err := client.Resource(server).Namespace("os-framework").Get(ctx, "app-service-api", metav1.GetOptions{}); err == nil {
		routes = append(routes, launcherRoute{"os-framework", "bfl-list-" + tenant, "app-service-api"})
		if tenantAPI {
			routes = append(routes, launcherRoute{"os-framework", "launcher-api-" + tenant, "app-service-api"})
		}
	} else if !apierrors.IsNotFound(err) {
		return err
	}
	return waitForLauncherRoutes(ctx, client, routes)
}

func waitForLauncherRoutes(ctx context.Context, client dynamic.Interface, routes []launcherRoute) error {
	resource := schema.GroupVersionResource{Group: "gateway.networking.k8s.io", Version: "v1", Resource: "httproutes"}
	var pending string
	err := wait.PollUntilContextTimeout(ctx, time.Second, 90*time.Second, true, func(ctx context.Context) (bool, error) {
		for _, route := range routes {
			pending = route.namespace + "/" + route.name
			obj, err := client.Resource(resource).Namespace(route.namespace).Get(ctx, route.name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			if !launcherRouteAccepted(obj, route.server) {
				return false, nil
			}
		}
		return true, nil
	})
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: waiting for Linkerd route %s: %w", ErrLauncherPending, pending, err)
	}
	if err != nil {
		return fmt.Errorf("waiting for Linkerd route %s: %w", pending, err)
	}
	return nil
}

func launcherRouteAccepted(obj *unstructured.Unstructured, server string) bool {
	if obj.GetDeletionTimestamp() != nil {
		return false
	}
	parents, _, _ := unstructured.NestedSlice(obj.Object, "status", "parents")
	for _, raw := range parents {
		parent, ok := raw.(map[string]any)
		if !ok || parent["controllerName"] != "linkerd.io/policy-controller" {
			continue
		}
		ref, _, _ := unstructured.NestedMap(parent, "parentRef")
		if ref["group"] != "policy.linkerd.io" || ref["kind"] != "Server" || ref["name"] != server ||
			(ref["namespace"] != nil && ref["namespace"] != obj.GetNamespace()) {
			continue
		}
		conditions, _, _ := unstructured.NestedSlice(parent, "conditions")
		for _, raw := range conditions {
			condition, ok := raw.(map[string]any)
			if ok && condition["type"] == "Accepted" && condition["status"] == "True" &&
				(condition["observedGeneration"] == nil || condition["observedGeneration"] == obj.GetGeneration()) {
				return true
			}
		}
	}
	return false
}

// RecoverLauncherUpgrade reconciles an abandoned, equivalent pending upgrade.
// Never undo manifests or values from a different revision automatically.
func RecoverLauncherUpgrade(ctx context.Context, cfg *action.Configuration, name string) error {
	latest, err := cfg.Releases.Last(name)
	if errors.Is(err, driver.ErrReleaseNotFound) {
		return nil
	}
	if err != nil || !latest.Info.Status.IsPending() {
		return err
	}
	deployed, err := cfg.Releases.Deployed(name)
	if err != nil && !errors.Is(err, driver.ErrNoDeployedReleases) {
		return err
	}
	if !recoverableLauncherUpgrade(latest, deployed, time.Now()) {
		return fmt.Errorf("%w: launcher %s revision %d is %s", ErrLauncherPending, name, latest.Version, latest.Info.Status)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// The recovered revision is equivalent to the deployed revision; rollback
	// creates a new Helm revision without removing live resources or storage.
	rollback := action.NewRollback(cfg)
	rollback.Version = deployed.Version
	rollback.DisableHooks = true
	rollback.Wait = true
	rollback.Timeout = time.Minute
	if err := rollback.Run(name); err != nil {
		return fmt.Errorf("recover launcher %s revision %d: %w", name, latest.Version, err)
	}
	return nil
}

func recoverableLauncherUpgrade(latest, deployed *release.Release, now time.Time) bool {
	return deployed != nil && latest.Info.Status == release.StatusPendingUpgrade &&
		!latest.Info.LastDeployed.IsZero() && now.Sub(latest.Info.LastDeployed.Time) >= 15*time.Minute &&
		latest.Version == deployed.Version+1 && latest.Manifest == deployed.Manifest &&
		reflect.DeepEqual(latest.Config, deployed.Config)
}
