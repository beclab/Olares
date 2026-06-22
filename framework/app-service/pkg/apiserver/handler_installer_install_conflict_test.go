package apiserver

import (
	"context"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	appfake "github.com/beclab/api/pkg/generated/clientset/versioned/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// conflictAM is a thin AM fixture for checkAppNameConflict tests; the
// upstream amObj() helper in apps_status_test.go always stamps the v3
// api-version label which is irrelevant to the conflict check (the check
// keys on the AppSharedLabel only).
func conflictAM(name, appName, owner string, state appv1alpha1.ApplicationManagerState, shared bool) *appv1alpha1.ApplicationManager {
	labels := map[string]string{}
	if shared {
		labels[constants.AppSharedLabel] = constants.AppSharedTrue
	} else {
		labels[constants.AppSharedLabel] = "false"
	}
	return &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName:      appName,
			AppNamespace: name,
			AppOwner:     owner,
		},
		Status: appv1alpha1.ApplicationManagerStatus{State: state},
	}
}

func newConflictHelper(t *testing.T, app string, ams ...*appv1alpha1.ApplicationManager) *installHandlerHelper {
	t.Helper()
	objs := make([]runtime.Object, 0, len(ams))
	for _, am := range ams {
		objs = append(objs, am)
	}
	client := appfake.NewSimpleClientset(objs...)
	return &installHandlerHelper{
		app:    app,
		owner:  "alice",
		client: client,
	}
}

// Different app names → no conflict regardless of shared flag.
func TestCheckAppNameConflict_DifferentAppName(t *testing.T) {
	other := conflictAM("other-bob-other", "other", "bob", appv1alpha1.Running, true)
	h := newConflictHelper(t, "myapp", other)

	if err := h.checkAppNameConflict(context.Background(), false); err != nil {
		t.Fatalf("unrelated app should not conflict, got %v", err)
	}
}

// Same app name + same type (both shared OR both per-user) is delegated to
// applyAppMgr's name-based Get/Create path, so the conflict check must
// stay silent — even on active states.
func TestCheckAppNameConflict_SameTypeFallThrough(t *testing.T) {
	cases := []struct {
		name          string
		existingShare bool
		newShared     bool
	}{
		{"both shared", true, true},
		{"both per-user", false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			am := conflictAM("myapp-foo", "myapp", "foo", appv1alpha1.Running, c.existingShare)
			h := newConflictHelper(t, "myapp", am)
			if err := h.checkAppNameConflict(context.Background(), c.newShared); err != nil {
				t.Fatalf("same-type collision must fall through to applyAppMgr, got %v", err)
			}
		})
	}
}

// Cross-type conflict on an ACTIVE state → return a non-nil error describing
// the existing install. The function returns a plain fmt.Errorf, not an
// apierrors.StatusError, and the install handler surfaces it via
// api.HandleBadRequest; assert only on the error message so this test stays
// honest about what the function actually produces.
func TestCheckAppNameConflict_CrossTypeActiveReturnsConflict(t *testing.T) {
	cases := []struct {
		name           string
		existingShared bool
		newShared      bool
	}{
		{"shared blocks per-user", true, false},
		{"per-user blocks shared", false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			am := conflictAM("myapp-amgr", "myapp", "bob", appv1alpha1.Running, c.existingShared)
			h := newConflictHelper(t, "myapp", am)

			err := h.checkAppNameConflict(context.Background(), c.newShared)
			if err == nil {
				t.Fatalf("expected conflict error, got nil")
			}
			if !strings.Contains(err.Error(), "is already installed as") {
				t.Fatalf("error %q must mention the conflicting install", err.Error())
			}

			// AM in an active state must NOT be removed by the check.
			_, getErr := h.client.AppV1alpha1().ApplicationManagers().Get(
				context.Background(), am.Name, metav1.GetOptions{})
			if getErr != nil {
				t.Fatalf("AM in active state must be preserved on conflict, got %v", getErr)
			}
		})
	}
}

// Cross-type collisions where the existing AM sits in an
// IsTerminalReinstallable state are SKIPPED by the conflict check: no error
// is surfaced and the install proceeds, but the check itself does NOT delete
// the stale AM. Reclaiming the AM CR is the appmgr GC controller's job
// (after AppMgrTerminalRetention) — keeping the responsibility in one place
// avoids racing two deleters on the same object.
//
// Iterate every reinstallable state to guard against future drift in the
// state-set predicate.
func TestCheckAppNameConflict_CrossTypeReinstallableSkipped(t *testing.T) {
	reinstallableStates := []appv1alpha1.ApplicationManagerState{
		appv1alpha1.Uninstalled,
		appv1alpha1.InstallingCanceled,
		appv1alpha1.InstallFailed,
		appv1alpha1.PendingCanceled,
		appv1alpha1.DownloadingCanceled,
		appv1alpha1.DownloadFailed,
	}
	for _, state := range reinstallableStates {
		t.Run(string(state), func(t *testing.T) {
			am := conflictAM("myapp-bob-myapp", "myapp", "bob", state, false)
			h := newConflictHelper(t, "myapp", am)

			// newShared=true (installer asking to install shared variant)
			// vs existing per-user → cross-type. Reinstallable existing AM
			// must let the install proceed.
			if err := h.checkAppNameConflict(context.Background(), true); err != nil {
				t.Fatalf("reinstallable state %s should not surface conflict, got %v", state, err)
			}
			// The stale AM must remain — checkAppNameConflict does not own
			// reclamation; the appmgr GC controller does.
			if _, getErr := h.client.AppV1alpha1().ApplicationManagers().Get(
				context.Background(), am.Name, metav1.GetOptions{}); getErr != nil {
				t.Fatalf("AM in %s must be preserved by the conflict check, got err=%v", state, getErr)
			}
		})
	}
}

// Multiple conflicting AMs: one in a reinstallable terminal state and one in
// an active state. The active one must win and block the install. Neither AM
// is removed by the check itself — the stale one is left for the appmgr GC
// controller to reclaim later, and the live one is preserved because there
// is nothing safe to do with it from a conflict-check call site.
//
// (The current implementation walks the list in order and short-circuits on
// the first active conflict; this test simply documents that "an active
// conflict anywhere in the list blocks the install".)
func TestCheckAppNameConflict_MixedAMsActiveWins(t *testing.T) {
	stale := conflictAM("myapp-old", "myapp", "carol", appv1alpha1.Uninstalled, false)
	live := conflictAM("myapp-shared-myapp", "myapp", "alice", appv1alpha1.Running, true)
	h := newConflictHelper(t, "myapp", stale, live)

	err := h.checkAppNameConflict(context.Background(), false)
	if err == nil {
		t.Fatalf("expected conflict error from live AM, got nil")
	}
	if !strings.Contains(err.Error(), "is already installed as") {
		t.Fatalf("error %q must mention the conflicting install", err.Error())
	}

	if _, getErr := h.client.AppV1alpha1().ApplicationManagers().Get(
		context.Background(), live.Name, metav1.GetOptions{}); getErr != nil {
		t.Errorf("live AM must be preserved, got %v", getErr)
	}
	// The stale (reinstallable) AM is also untouched: checkAppNameConflict
	// does not reclaim AMs, and the active conflict short-circuits the loop
	// before this entry could even be considered.
	if _, getErr := h.client.AppV1alpha1().ApplicationManagers().Get(
		context.Background(), stale.Name, metav1.GetOptions{}); getErr != nil {
		t.Errorf("stale AM must be preserved (reclamation belongs to GC), got %v", getErr)
	}
}
