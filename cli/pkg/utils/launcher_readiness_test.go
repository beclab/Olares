package utils

import (
	"context"
	"errors"
	"testing"
	"time"

	"helm.sh/helm/v3/pkg/release"
	helmtime "helm.sh/helm/v3/pkg/time"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestLauncherRouteRequiresLinkerdAndCorrectParent(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]any{}}
	obj.SetNamespace("user-space-alice")
	obj.SetGeneration(2)
	for _, tc := range []struct {
		name, controller, parent, ns, accepted string
		generation                             int64
		want                                   bool
	}{
		{"accepted", "linkerd.io/policy-controller", "bfl-api", "user-space-alice", "True", 2, true},
		{"another controller", "another/controller", "bfl-api", "user-space-alice", "True", 2, false},
		{"wrong server", "linkerd.io/policy-controller", "systemserver-api", "user-space-alice", "True", 2, false},
		{"another tenant", "linkerd.io/policy-controller", "bfl-api", "user-space-bob", "True", 2, false},
		{"rejected", "linkerd.io/policy-controller", "bfl-api", "user-space-alice", "False", 2, false},
		{"stale generation", "linkerd.io/policy-controller", "bfl-api", "user-space-alice", "True", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj.Object["status"] = map[string]any{"parents": []any{map[string]any{
				"controllerName": tc.controller,
				"parentRef":      map[string]any{"group": "policy.linkerd.io", "kind": "Server", "name": tc.parent, "namespace": tc.ns},
				"conditions":     []any{map[string]any{"type": "Accepted", "status": tc.accepted, "observedGeneration": tc.generation}},
			}}}
			if got := launcherRouteAccepted(obj, "bfl-api"); got != tc.want {
				t.Fatalf("accepted=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestOnlyEquivalentAbandonedLauncherUpgradeCanRecover(t *testing.T) {
	now := time.Now()
	deployed := &release.Release{Version: 3, Manifest: "existing", Config: map[string]any{"bfl": "existing"}}
	for _, tc := range []struct {
		name, manifest string
		age            time.Duration
		status         release.Status
		want           bool
	}{
		{"stale identical", "existing", 30 * time.Minute, release.StatusPendingUpgrade, true},
		{"active identical", "existing", time.Minute, release.StatusPendingUpgrade, false},
		{"changed manifest", "candidate", 30 * time.Minute, release.StatusPendingUpgrade, false},
		{"abandoned install", "existing", 30 * time.Minute, release.StatusPendingInstall, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			latest := &release.Release{Version: 4, Manifest: tc.manifest, Config: deployed.Config,
				Info: &release.Info{Status: tc.status, LastDeployed: helmtime.Time{Time: now.Add(-tc.age)}}}
			if got := recoverableLauncherUpgrade(latest, deployed, now); got != tc.want {
				t.Fatalf("recoverable=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestLauncherWaitsForEachTenantRoute(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1", "kind": "HTTPRoute",
		"status": map[string]any{"parents": []any{map[string]any{
			"controllerName": "linkerd.io/policy-controller",
			"parentRef":      map[string]any{"group": "policy.linkerd.io", "kind": "Server", "name": "bfl-api"},
			"conditions":     []any{map[string]any{"type": "Accepted", "status": "True"}},
		}}},
	}}
	obj.SetName("bfl-api-frontend-apps")
	obj.SetNamespace("user-space-alice")
	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), obj)
	alice := launcherRoute{"user-space-alice", "bfl-api-frontend-apps", "bfl-api", ""}
	if err := waitForLauncherRoutes(t.Context(), client, []launcherRoute{alice}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	bob := launcherRoute{"user-space-bob", alice.name, alice.server, ""}
	if err := waitForLauncherRoutes(ctx, client, []launcherRoute{alice, bob}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("another tenant's accepted route passed readiness: %v", err)
	}
}

func TestLauncherWaitsForResetAuthorizationAndAuthentication(t *testing.T) {
	const ns = "user-system-alice"
	const policyAPI = "policy.linkerd.io/v1alpha1"
	const routeAPI = "gateway.networking.k8s.io/v1"
	const name = "activation-password"
	route := launcherRoute{ns, name, "systemserver-existing-28080", name}
	metadata := map[string]any{"namespace": ns, "name": name}
	httpRoute := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": routeAPI, "kind": "HTTPRoute", "metadata": metadata,
		"status": map[string]any{"parents": []any{map[string]any{
			"controllerName": "linkerd.io/policy-controller",
			"parentRef":      map[string]any{"group": "policy.linkerd.io", "kind": "Server", "name": route.server},
			"conditions":     []any{map[string]any{"type": "Accepted", "status": "True"}},
		}}},
	}}
	policy := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": policyAPI, "kind": "AuthorizationPolicy", "metadata": metadata,
		"spec": map[string]any{
			"targetRef":                  map[string]any{"group": "gateway.networking.k8s.io", "kind": "HTTPRoute", "name": name},
			"requiredAuthenticationRefs": []any{map[string]any{"group": "policy.linkerd.io", "kind": "MeshTLSAuthentication", "name": name}},
		},
	}}
	auth := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": policyAPI, "kind": "MeshTLSAuthentication", "metadata": metadata,
		"spec": map[string]any{"identityRefs": []any{map[string]any{"kind": "ServiceAccount", "name": "bytetrade-controller", "namespace": "user-space-alice"}}},
	}}
	empty := auth.DeepCopy()
	delete(empty.Object, "spec")
	foreign := auth.DeepCopy()
	foreign.SetNamespace("user-system-bob")
	wrong := policy.DeepCopy()
	wrong.Object["spec"].(map[string]any)["targetRef"].(map[string]any)["name"] = "activation-provider"
	for _, tc := range []struct {
		name      string
		resources []runtime.Object
		ready     bool
	}{
		{"accepted route only", nil, false},
		{"missing mTLS binding", []runtime.Object{policy}, false},
		{"empty mTLS binding", []runtime.Object{policy, empty}, false},
		{"wrong namespace", []runtime.Object{policy, foreign}, false},
		{"wrong target", []runtime.Object{wrong, auth}, false},
		{"complete binding", []runtime.Object{policy, auth}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resources := append([]runtime.Object{httpRoute}, tc.resources...)
			client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), resources...)
			ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
			defer cancel()
			err := waitForLauncherRoutes(ctx, client, []launcherRoute{route})
			if tc.ready && err != nil || !tc.ready && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("readiness=%v, want ready=%v", err, tc.ready)
			}
		})
	}
}
