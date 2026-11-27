package utils

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// AppServiceAPITenant is the inventory taken from the installer owner or the
// authoritative User API, with identity bound to user-system-<Name>.
type AppServiceAPITenant struct {
	Name string
	Role string
}

// AppServiceAPIPolicyChartExists keeps older installation bundles usable.
func AppServiceAPIPolicyChartExists(frameworkChart string) (bool, error) {
	_, err := os.Stat(filepath.Join(frameworkChart, "templates", "deploy", "appservice_api_mesh.yaml"))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// PrepareAppServiceAPIValues prepares the 6755 policy without changing release
// enforcement flags. Repeat installs and upgrades preserve the deployed posture.
func PrepareAppServiceAPIValues(cfg *action.Configuration, releaseName string, values map[string]any, tenants []AppServiceAPITenant) error {
	var previous map[string]any
	rel, err := cfg.Releases.Deployed(releaseName)
	if err == nil {
		previous = rel.Config
	} else if !errors.Is(err, driver.ErrNoDeployedReleases) && !errors.Is(err, driver.ErrReleaseNotFound) {
		return fmt.Errorf("read deployed framework mesh configuration: %w", err)
	}
	return prepareAppServiceAPIValues(values, previous, tenants)
}

func prepareAppServiceAPIValues(values, previous map[string]any, tenants []AppServiceAPITenant) error {
	roles := make(map[string]string, len(tenants))
	for _, tenant := range tenants {
		if tenant.Name == "" || len(tenant.Name) > 51 || len(validation.IsDNS1123Label(tenant.Name)) != 0 {
			return fmt.Errorf("invalid app-service API tenant name %q", tenant.Name)
		}
		if _, exists := roles[tenant.Name]; exists {
			return fmt.Errorf("duplicate app-service API tenant %q", tenant.Name)
		}
		roles[tenant.Name] = tenant.Role
	}
	names := slices.Sorted(maps.Keys(roles))
	if len(names) == 0 {
		return fmt.Errorf("no eligible app-service API tenants")
	}
	configured, _ := previous["appServiceAPIMesh"].(map[string]any)
	// The owner is enrolled on a fresh installation. Existing explicitly
	// delegated provisioners survive upgrades only while still administrators.
	provisioners := map[string]bool{}
	for name, role := range roles {
		if role == "owner" {
			provisioners[name] = true
		}
	}
	if len(provisioners) != 1 {
		return fmt.Errorf("expected exactly one owner for app-service API policies")
	}
	priorDelegates, _ := configured["userProvisioners"].([]any)
	for _, entry := range priorDelegates {
		if name, ok := entry.(string); ok && roles[name] == "admin" {
			provisioners[name] = true
		}
	}
	next := map[string]any{"prepare": true, "tenants": names, "userProvisioners": slices.Sorted(maps.Keys(provisioners))}
	for _, flag := range []string{"activate", "marketLifecycle", "systemEnvWrites"} {
		if configured[flag] == true {
			next[flag] = true
		}
	}
	// Existing opt-ins apply to the cohort the operator enrolled. Discovery of
	// a new tenant must not expand system configuration or Market write access.
	for _, feature := range []struct{ flag, scope string }{
		{"systemEnvWrites", "systemEnvWriterTenants"},
		{"marketLifecycle", "marketLifecycleTenants"},
	} {
		if next[feature.flag] != true {
			continue
		}
		configuredTenants, exists := configured[feature.scope]
		if !exists {
			configuredTenants = configured["tenants"]
		}
		scoped := []string{}
		if previousTenants, ok := configuredTenants.([]any); ok {
			for _, entry := range previousTenants {
				if name, ok := entry.(string); ok && roles[name] != "" {
					scoped = append(scoped, name)
				}
			}
		}
		slices.Sort(scoped)
		next[feature.scope] = slices.Compact(scoped)
	}
	values["appServiceAPIMesh"] = next
	if market, ok := previous["marketBackendMTLS"].(map[string]any); ok && market["prepare"] == true {
		marketValues := maps.Clone(market)
		marketValues["tenants"] = names
		values["marketBackendMTLS"] = marketValues
		if previous["marketLinkerd"] == true {
			values["marketLinkerd"] = true
		}
	}
	return nil
}

// WaitForPreparedAppServiceAPI checks a separate policy stage and is safe
// when the listener's strict selector is still inactive.
func WaitForPreparedAppServiceAPI(ctx context.Context, config *rest.Config, tenants []AppServiceAPITenant) error {
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	return waitForPreparedAppServiceAPI(ctx, client, tenants)
}

func waitForPreparedAppServiceAPI(ctx context.Context, client dynamic.Interface, tenants []AppServiceAPITenant) error {
	routes := []launcherRoute{{"os-framework", "market-image-info", "app-service-api"}}
	for _, tenant := range tenants {
		routes = append(routes,
			launcherRoute{"os-framework", "users-" + tenant.Name, "app-service-api"},
			launcherRoute{"os-framework", "user-apps-" + tenant.Name, "app-service-api"},
		)
		if tenant.Role == "owner" {
			routes = append(routes, launcherRoute{"os-framework", "settings-user-create-" + tenant.Name, "app-service-api"})
		}
	}
	server := client.Resource(schema.GroupVersionResource{Group: "policy.linkerd.io", Version: "v1beta3", Resource: "servers"}).Namespace("os-framework")
	obj, err := server.Get(ctx, "app-service-api", metav1.GetOptions{})
	if err != nil {
		return err
	}
	port, _, _ := unstructured.NestedInt64(obj.Object, "spec", "port")
	policy, _, _ := unstructured.NestedString(obj.Object, "spec", "accessPolicy")
	selector, _, _ := unstructured.NestedStringMap(obj.Object, "spec", "podSelector", "matchLabels")
	if obj.GetDeletionTimestamp() != nil || port != 6755 || policy != "deny" ||
		selector["tier"] != "app-service" || selector["traffic.olares.io/app-service-api-mtls"] != "required" {
		return fmt.Errorf("app-service-api Server does not protect the expected listener")
	}
	gvr := schema.GroupVersionResource{Group: "policy.linkerd.io", Version: "v1alpha1"}
	var missing string
	err = wait.PollUntilContextTimeout(ctx, time.Second, 90*time.Second, true, func(ctx context.Context) (bool, error) {
		for _, kind := range []string{"authorizationpolicies", "meshtlsauthentications"} {
			gvr.Resource = kind
			for _, route := range routes {
				missing = kind + "/" + route.name
				obj, err := client.Resource(gvr).Namespace(route.namespace).Get(ctx, route.name, metav1.GetOptions{})
				if apierrors.IsNotFound(err) {
					return false, nil
				}
				if err != nil {
					return false, err
				}
				if obj.GetDeletionTimestamp() != nil {
					return false, nil
				}
			}
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for app-service API authorization %s: %w", missing, err)
	}
	return waitForLauncherRoutes(ctx, client, routes)
}
