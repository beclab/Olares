package webhook

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	appfake "github.com/beclab/api/pkg/generated/clientset/versioned/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// INV-MAC-1 (OG-DES-002 §3.1.1): every macvlan interface on br-olares shares
// one L2 broadcast domain, so two tenants running the same app must never be
// handed the same MAC. A duplicate would collapse their DHCP identity as well,
// because the client identifier is derived from the MAC.

// tenant describes one user's installation of an app.
//
// namespace and owner must agree: for a "user-space*" namespace, AppNamespace
// (pkg/utils/app_utils.go:65-73) resolves the Application through the OWNER
// label rather than the raw namespace, returning "user-space-<owner>". A pod
// carrying only the app-name label resolves to "user-space-" and is not found.
type tenant struct {
	namespace string
	owner     string
	appName   string
	uid       string
}

// multiTenantWebhook builds a Webhook whose Application store contains one
// Application per tenant. Application names follow FmtAppMgrName, i.e.
// "<namespace>-<app>", which is what the allocator resolves a pod to.
func multiTenantWebhook(tenants ...tenant) *Webhook {
	objects := make([]runtime.Object, 0, len(tenants))
	for _, tn := range tenants {
		objects = append(objects, &appv1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name: tn.namespace + "-" + tn.appName,
				UID:  types.UID(tn.uid),
			},
			Spec: appv1alpha1.ApplicationSpec{
				Name:      tn.appName,
				Namespace: tn.namespace,
				Settings:  map[string]string{"enableOverlayGateway": "true"},
			},
		})
	}
	// LIST on the ledger needs its list kind registered with the fake dynamic
	// client, otherwise client-go panics rather than returning an error.
	listKinds := map[schema.GroupVersionResource]string{
		overlayMACAllocationGVR: "OverlayMACAllocationList",
	}
	return &Webhook{
		kubeClient:    k8sfake.NewSimpleClientset(),
		dynamicClient: appfake.NewSimpleClientset(objects...),
		allocationClient: dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(), listKinds),
	}
}

// tenantPod returns a pod belonging to the given tenant.
func tenantPod(tn tenant, podName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: tn.namespace,
			Labels: map[string]string{
				constants.ApplicationNameLabel:  tn.appName,
				constants.ApplicationOwnerLabel: tn.owner,
			},
		},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: tn.appName}}},
	}
}

// injectedMAC runs the mutating patch and returns the MAC written into the
// Multus network selection annotation.
func injectedMAC(t *testing.T, wh *Webhook, pod *corev1.Pod) string {
	t.Helper()
	if _, err := wh.CreateMacvlanInitPatch(macvlanBypassAdmissionRequest(t, pod), pod); err != nil {
		t.Fatalf("CreateMacvlanInitPatch(%s/%s): %v", pod.Namespace, pod.Name, err)
	}
	var selections []map[string]interface{}
	raw := pod.Annotations["k8s.v1.cni.cncf.io/networks"]
	if err := json.Unmarshal([]byte(raw), &selections); err != nil {
		t.Fatalf("decode networks annotation %q: %v", raw, err)
	}
	if len(selections) != 1 {
		t.Fatalf("unexpected network selections: %#v", selections)
	}
	mac, _ := selections[0]["mac"].(string)
	if err := validateOverlayMAC(mac); err != nil {
		t.Fatalf("validateOverlayMAC(%q): %v", mac, err)
	}
	return mac
}

// TestOverlayMACDiffersAcrossTenantsForSameApp is the regression guard for
// INV-MAC-1: the same app name installed by two tenants must not collide.
func TestOverlayMACDiffersAcrossTenantsForSameApp(t *testing.T) {
	alice := tenant{namespace: "user-space-alice", owner: "alice", appName: "jellyfin", uid: "uid-alice"}
	bob := tenant{namespace: "user-space-bob", owner: "bob", appName: "jellyfin", uid: "uid-bob"}
	wh := multiTenantWebhook(alice, bob)

	aliceMAC := injectedMAC(t, wh, tenantPod(alice, "jellyfin-0"))
	bobMAC := injectedMAC(t, wh, tenantPod(bob, "jellyfin-0"))

	if aliceMAC == bobMAC {
		t.Fatalf("INV-MAC-1 violated: both tenants got MAC %q for app %q", aliceMAC, alice.appName)
	}
}

// TestOverlayMACPerTenantLedgerEntriesAreIndependent asserts the ledger records
// one claim per tenant, each owned by that tenant's Application. Distinct MACs
// alone are not enough: the claims must not alias onto one another.
func TestOverlayMACPerTenantLedgerEntriesAreIndependent(t *testing.T) {
	alice := tenant{namespace: "user-space-alice", owner: "alice", appName: "jellyfin", uid: "uid-alice"}
	bob := tenant{namespace: "user-space-bob", owner: "bob", appName: "jellyfin", uid: "uid-bob"}
	wh := multiTenantWebhook(alice, bob)

	aliceMAC := injectedMAC(t, wh, tenantPod(alice, "jellyfin-0"))
	bobMAC := injectedMAC(t, wh, tenantPod(bob, "jellyfin-0"))

	list, err := wh.allocationClient.Resource(overlayMACAllocationGVR).
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list allocations: %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("allocation count = %d, want 2 (one per tenant)", len(list.Items))
	}

	byInstanceKey := map[string]string{}
	byUID := map[string]string{}
	for _, item := range list.Items {
		spec, _ := item.Object["spec"].(map[string]interface{})
		mac, _ := spec["mac"].(string)
		instanceKey, _ := spec["instanceKey"].(string)
		uid, _ := spec["applicationUID"].(string)
		byInstanceKey[instanceKey] = mac
		byUID[uid] = mac
	}

	wantAliceKey := alice.namespace + "/" + alice.appName
	wantBobKey := bob.namespace + "/" + bob.appName
	if got := byInstanceKey[wantAliceKey]; got != aliceMAC {
		t.Fatalf("ledger instanceKey %q -> %q, want %q", wantAliceKey, got, aliceMAC)
	}
	if got := byInstanceKey[wantBobKey]; got != bobMAC {
		t.Fatalf("ledger instanceKey %q -> %q, want %q", wantBobKey, got, bobMAC)
	}
	if byUID[alice.uid] == byUID[bob.uid] {
		t.Fatalf("both tenants' claims resolved to the same MAC %q", byUID[alice.uid])
	}
}

// TestOverlayMACRemainsStablePerTenantAcrossRecreate asserts that isolation and
// stability hold together: each tenant keeps its own MAC across pod recreation,
// and the two never converge.
func TestOverlayMACRemainsStablePerTenantAcrossRecreate(t *testing.T) {
	alice := tenant{namespace: "user-space-alice", owner: "alice", appName: "jellyfin", uid: "uid-alice"}
	bob := tenant{namespace: "user-space-bob", owner: "bob", appName: "jellyfin", uid: "uid-bob"}
	wh := multiTenantWebhook(alice, bob)

	aliceFirst := injectedMAC(t, wh, tenantPod(alice, "jellyfin-0"))
	bobFirst := injectedMAC(t, wh, tenantPod(bob, "jellyfin-0"))

	aliceSecond := injectedMAC(t, wh, tenantPod(alice, "jellyfin-recreated"))
	bobSecond := injectedMAC(t, wh, tenantPod(bob, "jellyfin-recreated"))

	if aliceSecond != aliceFirst {
		t.Fatalf("alice MAC changed across recreate: %q -> %q", aliceFirst, aliceSecond)
	}
	if bobSecond != bobFirst {
		t.Fatalf("bob MAC changed across recreate: %q -> %q", bobFirst, bobSecond)
	}
	if aliceSecond == bobSecond {
		t.Fatalf("INV-MAC-1 violated after recreate: both tenants on %q", aliceSecond)
	}
}

// TestOverlayMACDiffersAcrossTenantsForDistinctApps covers the wider case: two
// tenants, two different apps, four claims, all MACs distinct.
func TestOverlayMACDiffersAcrossTenantsForDistinctApps(t *testing.T) {
	tenants := []tenant{
		{namespace: "user-space-alice", owner: "alice", appName: "jellyfin", uid: "uid-alice-jellyfin"},
		{namespace: "user-space-alice", owner: "alice", appName: "nextcloud", uid: "uid-alice-nextcloud"},
		{namespace: "user-space-bob", owner: "bob", appName: "jellyfin", uid: "uid-bob-jellyfin"},
		{namespace: "user-space-bob", owner: "bob", appName: "nextcloud", uid: "uid-bob-nextcloud"},
	}
	wh := multiTenantWebhook(tenants...)

	seen := map[string]string{}
	for _, tn := range tenants {
		mac := injectedMAC(t, wh, tenantPod(tn, tn.appName+"-0"))
		key := tn.namespace + "/" + tn.appName
		if prev, dup := seen[mac]; dup {
			t.Fatalf("INV-MAC-1 violated: %q and %q share MAC %q", prev, key, mac)
		}
		seen[mac] = key
	}
	if len(seen) != len(tenants) {
		t.Fatalf("distinct MACs = %d, want %d", len(seen), len(tenants))
	}
}
