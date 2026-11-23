package entryreadiness

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/beclab/l4-bfl-proxy/internal/message"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var systemRouteRegistryGVK = schema.GroupVersionKind{
	Group: "gateway.olares.io", Version: "v1alpha1", Kind: "SharedRouteRegistry",
}

// Checker reports app-service's observed gateway and Linkerd readiness for
// the login and Desktop authorities that this proxy actually serves.
func Checker(resources *message.ProviderResources, reader client.Reader, gateway Gateway) func(*http.Request) error {
	return checker(resources, reader, gateway.client)
}

func checker(resources *message.ProviderResources, reader client.Reader, newClient entryClient) func(*http.Request) error {
	return func(req *http.Request) error {
		ctx, cancel := context.WithTimeout(req.Context(), 4*time.Second)
		defer cancel()
		snapshots := resources.LoadAll()
		if len(snapshots) == 0 {
			return fmt.Errorf("L4 user snapshot not published")
		}
		for _, snapshot := range snapshots {
			if snapshot == nil {
				return fmt.Errorf("L4 user snapshot not published")
			}
			if err := checkSnapshot(ctx, reader, snapshot, newClient); err != nil {
				return err
			}
		}
		return nil
	}
}

func checkSnapshot(ctx context.Context, reader client.Reader, snapshot *message.Resources, newClient entryClient) error {
	var verify verifyEntry
	for _, user := range snapshot.Users {
		if user == nil || user.IsEphemeral || user.SSL == nil || user.SSL.Zone == "" {
			continue
		}
		if !snapshot.EntryMTLS.Ready() {
			return fmt.Errorf("Entry mTLS client material is unavailable")
		}
		zone := strings.ToLower(strings.TrimSpace(user.SSL.Zone))
		for _, entrance := range []string{"auth", "desktop"} {
			name := "sys-" + entrance
			if err := checkSystemRoute(ctx, reader, user.Namespace, user.Name, name,
				entrance+"."+zone); err != nil {
				return err
			}
		}
		if verify == nil {
			clientProbe, closeClient, err := newClient(snapshot.EntryMTLS)
			if err != nil {
				return err
			}
			defer closeClient()
			verify = clientProbe
		}
		if err := verify(ctx, zone); err != nil {
			return err
		}
	}
	return nil
}

func checkSystemRoute(ctx context.Context, reader client.Reader, namespace, owner, name, host string) error {
	ref := client.ObjectKey{Namespace: namespace, Name: name}
	if reader == nil {
		return fmt.Errorf("API reader unavailable for %s", ref)
	}
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(systemRouteRegistryGVK)
	if err := reader.Get(ctx, ref, route); err != nil {
		return fmt.Errorf("get Entry route status for %s: %w", ref, err)
	}
	if !systemRouteReady(route, owner, host) {
		return fmt.Errorf("Entry route %s is not ready for %s", ref, host)
	}
	return nil
}

func systemRouteReady(route *unstructured.Unstructured, owner, host string) bool {
	if !route.GetDeletionTimestamp().IsZero() || route.GetGeneration() <= 0 {
		return false
	}
	labels, annotations := route.GetLabels(), route.GetAnnotations()
	if route.GetNamespace() != "user-space-"+owner || annotations["gateway.olares.io/extauth-user"] != owner ||
		labels["gateway.olares.io/system-plane"] != "true" ||
		labels["app.kubernetes.io/managed-by"] != "app-service" {
		return false
	}
	mode, _, _ := unstructured.NestedString(route.Object, "spec", "routeMode")
	class, _, _ := unstructured.NestedString(route.Object, "spec", "entranceClass")
	hosts, _, _ := unstructured.NestedStringSlice(route.Object, "spec", "hostPatterns")
	bound := false
	for _, name := range hosts {
		if name == host {
			bound = true
			break
		}
	}
	observed, _, _ := unstructured.NestedInt64(route.Object, "status", "observedGeneration")
	httpRoute, _, _ := unstructured.NestedString(route.Object, "status", "httpRouteName")
	if mode != "gateway" || class != "application" || !bound || observed != route.GetGeneration() || httpRoute != route.GetName() {
		return false
	}
	conditions, _, _ := unstructured.NestedSlice(route.Object, "status", "conditions")
	for _, item := range conditions {
		condition, ok := item.(map[string]any)
		if !ok || condition["type"] != "Ready" {
			continue
		}
		generation, _, _ := unstructured.NestedInt64(condition, "observedGeneration")
		return generation == route.GetGeneration() && condition["status"] == "True" && condition["reason"] == "Reconciled"
	}
	return false
}
