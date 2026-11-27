package gateway

import (
	"reflect"
	"testing"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSystemPlaneHostPatternsOwner(t *testing.T) {
	auth := SystemPlaneServices[0]
	got := SystemPlaneHostPatterns(auth, "alice.olares.com", "alice", false)
	want := []string{"auth.alice.olares.com", "auth.alice.olares.local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("auth hosts=%v want %v", got, want)
	}

	profile := SystemPlaneServices[3]
	got = SystemPlaneHostPatterns(profile, "alice.olares.com", "alice", false)
	want = []string{"alice.olares.com", "alice.olares.local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("profile hosts=%v want %v", got, want)
	}
}

func TestReadyWizardHostsRequiresCurrentTenantRoute(t *testing.T) {
	ctx := t.Context()
	scheme := runtime.NewScheme()
	if err := srrv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&srrv1alpha1.SharedRouteRegistry{}).Build()
	const zone = "owner.olares.com"
	for _, user := range []string{"alice", "bob"} {
		if err := EnsureSystemPlaneRoutes(ctx, c, user, zone, "user-space-"+user, true); err != nil {
			t.Fatal(err)
		}
	}
	setReady := func(user, reason string, current bool) {
		t.Helper()
		var r srrv1alpha1.SharedRouteRegistry
		if err := c.Get(ctx, client.ObjectKey{Namespace: "user-space-" + user, Name: "sys-wizard"}, &r); err != nil {
			t.Fatal(err)
		}
		generation := r.Generation
		if !current {
			generation++
		}
		r.Status.HTTPRouteName = r.Name
		r.Status.Conditions = []metav1.Condition{{Type: "Ready", Status: metav1.ConditionTrue, Reason: reason, ObservedGeneration: generation}}
		if err := c.Status().Update(ctx, &r); err != nil {
			t.Fatal(err)
		}
	}
	check := func(user, ownerZone string, want []string) {
		t.Helper()
		hosts, err := ReadyWizardHosts(ctx, c, user, ownerZone)
		if err != nil || !reflect.DeepEqual(hosts, want) {
			t.Fatalf("activation hosts for %s/%s = %v, %v; want %v", user, ownerZone, hosts, err, want)
		}
	}
	check("alice", zone, nil)
	setReady("alice", WizardEntryReadyReason, true)
	check("alice", zone, []string{"wizard-alice.owner.olares.com", "wizard-alice.owner.olares.local"})
	check("alice", "different.olares.com", nil)
	setReady("bob", "Reconciled", true)
	check("bob", zone, nil)
	setReady("bob", WizardEntryReadyReason, false)
	check("bob", zone, nil)
	setReady("bob", WizardEntryReadyReason, true)
	check("bob", zone, []string{"wizard-bob.owner.olares.com", "wizard-bob.owner.olares.local"})
}

func TestSystemPlaneHostPatternsEphemeralWizardOnly(t *testing.T) {
	wizard := SystemPlaneServices[2]
	got := SystemPlaneHostPatterns(wizard, "alice.olares.com", "guest1", true)
	want := []string{"wizard-guest1.alice.olares.com", "wizard-guest1.alice.olares.local"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ephemeral wizard hosts=%v want %v", got, want)
	}
	if h := SystemPlaneHostPatterns(SystemPlaneServices[0], "alice.olares.com", "guest1", true); len(h) != 0 {
		t.Fatalf("ephemeral auth should be empty, got %v", h)
	}
	if h := SystemPlaneHostPatterns(SystemPlaneServices[3], "alice.olares.com", "guest1", true); len(h) != 0 {
		t.Fatalf("ephemeral profile should be empty, got %v", h)
	}
}
