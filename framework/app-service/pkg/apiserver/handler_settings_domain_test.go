package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned"
	appfake "github.com/beclab/api/pkg/generated/clientset/versioned/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

// dcApp builds an Application fixture for the domain-duplicate tests. The CR
// metadata.Name is derived as "<specName>-<owner>" so per-user installs of the
// same app name can coexist in the fake client, while the conflict logic
// matches on Spec.Name. Shared apps also carry the v3 api-version label so the
// EffectiveSettings overlay (used by callerCustomDomainBlob for shared apps)
// actually applies the per-user Spec.UserSettings.
func dcApp(specName, owner, appid string, shared bool, entrances ...string) *v1alpha1.Application {
	labels := map[string]string{}
	if shared {
		labels[constants.AppSharedLabel] = constants.AppSharedTrue
		labels[constants.AppApiVersionLabel] = constants.AppVersionV3
	}
	ents := make([]v1alpha1.Entrance, 0, len(entrances))
	for _, e := range entrances {
		ents = append(ents, v1alpha1.Entrance{Name: e})
	}
	return &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:   specName + "-" + owner,
			Labels: labels,
		},
		Spec: v1alpha1.ApplicationSpec{
			Name:      specName,
			Owner:     owner,
			Appid:     appid,
			Entrances: ents,
		},
	}
}

func customDomainJSON(t *testing.T, entries map[string]map[string]string) string {
	t.Helper()
	b, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal customDomain blob: %v", err)
	}
	return string(b)
}

// setCustomDomain writes the global Spec.Settings customDomain entry for an
// entrance. This is the real store only for a per-user (non-shared) app owned
// by the caller; shared apps keep each user's customDomain in a per-user
// Spec.UserSettings overlay instead (see setUserCustomDomain).
func setCustomDomain(t *testing.T, a *v1alpha1.Application, entrance, thirdLevel, thirdParty string) *v1alpha1.Application {
	t.Helper()
	if a.Spec.Settings == nil {
		a.Spec.Settings = map[string]string{}
	}
	a.Spec.Settings["customDomain"] = customDomainJSON(t, map[string]map[string]string{
		entrance: {"third_level_domain": thirdLevel, "third_party_domain": thirdParty},
	})
	return a
}

// setUserCustomDomain writes a per-user Spec.UserSettings overlay customDomain
// entry (the value effective in that specific user's zone for a shared app).
func setUserCustomDomain(t *testing.T, a *v1alpha1.Application, user, entrance, thirdLevel, thirdParty string) *v1alpha1.Application {
	t.Helper()
	if a.Spec.UserSettings == nil {
		a.Spec.UserSettings = map[string]map[string]string{}
	}
	if a.Spec.UserSettings[user] == nil {
		a.Spec.UserSettings[user] = map[string]string{}
	}
	a.Spec.UserSettings[user]["customDomain"] = customDomainJSON(t, map[string]map[string]string{
		entrance: {"third_level_domain": thirdLevel, "third_party_domain": thirdParty},
	})
	return a
}

// setDefaultThirdLevelConfig sets the defaultThirdLevelDomainConfig blob that
// overrides the positional "<appid><i>" default for a specific entrance of a
// multi-entrance app.
func setDefaultThirdLevelConfig(t *testing.T, a *v1alpha1.Application, appName, entrance, thirdLevel string) *v1alpha1.Application {
	t.Helper()
	if a.Spec.Settings == nil {
		a.Spec.Settings = map[string]string{}
	}
	cfg := []map[string]string{{
		"appName":          appName,
		"entranceName":     entrance,
		"thirdLevelDomain": thirdLevel,
	}}
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal defaultThirdLevelDomainConfig: %v", err)
	}
	a.Spec.Settings["defaultThirdLevelDomainConfig"] = string(b)
	return a
}

func newDomainClient(objs ...*v1alpha1.Application) versioned.Interface {
	ro := make([]runtime.Object, 0, len(objs))
	for _, o := range objs {
		ro = append(ro, o)
	}
	return appfake.NewSimpleClientset(ro...)
}

// Both requested values empty (or whitespace-only) short-circuit before any
// listing: the check returns nil without touching the API server. The list
// reactor is wired to fail so the test also proves listing never happens.
func TestCheckEntranceDomainDuplicate_EmptyInputsSkipListing(t *testing.T) {
	client := appfake.NewSimpleClientset()
	client.PrependReactor("list", "applications", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("list must not be called for empty inputs")
	})

	cases := []struct{ tl, tp string }{
		{"", ""},
		{"   ", "\t"},
		{" ", ""},
	}
	for _, c := range cases {
		if err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", c.tl, c.tp); err != nil {
			t.Fatalf("empty/whitespace inputs must skip listing and return nil, got %v", err)
		}
	}
}

// A failure listing applications is surfaced verbatim to the caller.
func TestCheckEntranceDomainDuplicate_ListError(t *testing.T) {
	client := appfake.NewSimpleClientset()
	client.PrependReactor("list", "applications", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("list boom")
	})

	err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", "wanted", "")
	if err == nil || !strings.Contains(err.Error(), "list boom") {
		t.Fatalf("expected list error to propagate, got %v", err)
	}
}

// Reserved system subdomains can never be claimed as a third_level_domain,
// regardless of which apps exist, and the match is case-insensitive.
func TestCheckEntranceDomainDuplicate_ReservedThirdLevel(t *testing.T) {
	client := newDomainClient()
	for _, name := range []string{"auth", "desktop", "wizard", "AUTH", "Desktop", "WIZARD"} {
		err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", name, "")
		if err == nil || !strings.Contains(err.Error(), "reserved") {
			t.Fatalf("third_level_domain %q must be rejected as reserved, got %v", name, err)
		}
	}
}

// The reserved-subdomain guard only applies to third_level_domain; the same
// string is a perfectly valid third_party_domain when nothing else claims it.
func TestCheckEntranceDomainDuplicate_ReservedDoesNotBlockThirdParty(t *testing.T) {
	client := newDomainClient()
	if err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", "", "auth"); err != nil {
		t.Fatalf("reserved names must not block a third_party_domain, got %v", err)
	}
}

// third_party_domain uniqueness is global (across every user and zone): a value
// already claimed by a different app anywhere blocks the request.
func TestCheckEntranceDomainDuplicate_ThirdPartyConflictOtherApp(t *testing.T) {
	other := setCustomDomain(t, dcApp("other", "bob", "otherid", false, "web"), "web", "", "example.com")
	client := newDomainClient(other)

	err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", "", "example.com")
	if err == nil || !strings.Contains(err.Error(), "third_party_domain") {
		t.Fatalf("expected third_party_domain conflict, got %v", err)
	}
}

// third_party match is case-insensitive.
func TestCheckEntranceDomainDuplicate_ThirdPartyConflictCaseInsensitive(t *testing.T) {
	other := setCustomDomain(t, dcApp("other", "bob", "otherid", false, "web"), "web", "", "Example.COM")
	client := newDomainClient(other)

	err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", "", "example.com")
	if err == nil || !strings.Contains(err.Error(), "third_party_domain") {
		t.Fatalf("expected case-insensitive third_party_domain conflict, got %v", err)
	}
}

// On a shared app each user sets their own third_party overlay per entrance.
// Another user's overlay on the same app+entrance still counts as a conflict
// for the caller, because third_party is compared globally and only the
// caller's OWN overlay is treated as self.
func TestCheckEntranceDomainDuplicate_ThirdPartyConflictOtherUsersOverlay(t *testing.T) {
	app := dcApp("cloud", "alice", "cloudid", true, "web")
	setUserCustomDomain(t, app, "bob", "web", "", "bob.example.com")
	client := newDomainClient(app)

	// alice edits the very same shared entrance and asks for the value bob
	// already claimed in his zone -> conflict.
	err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "cloud", "web", "", "bob.example.com")
	if err == nil || !strings.Contains(err.Error(), "third_party_domain") {
		t.Fatalf("expected conflict against another user's overlay, got %v", err)
	}
}

// Re-saving the caller's own existing third_party value on the entry being
// edited is never a conflict (isSelf excludes it via owner==caller).
func TestCheckEntranceDomainDuplicate_ThirdPartySelfIsNotConflict(t *testing.T) {
	app := dcApp("cloud", "alice", "cloudid", true, "web")
	setUserCustomDomain(t, app, "bob", "web", "", "bob.example.com")
	client := newDomainClient(app)

	err := checkEntranceDomainDuplicate(context.Background(), client, "bob", "cloud", "web", "", "bob.example.com")
	if err != nil {
		t.Fatalf("re-saving own third_party value must not conflict, got %v", err)
	}
}

// third_level_domain uniqueness is scoped to the caller's zone. A value already
// used by another entrance's custom domain that resolves in the caller's zone
// is a conflict, reported as "already used by entrance" (non-default).
func TestCheckEntranceDomainDuplicate_ThirdLevelConflictCustomDomain(t *testing.T) {
	// per-user app owned by the caller: its Spec.Settings customDomain
	// resolves in the caller's own zone.
	a := setCustomDomain(t, dcApp("one", "alice", "oneid", false, "web"), "web", "taken", "")
	client := newDomainClient(a)

	err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "two", "x", "taken", "")
	if err == nil {
		t.Fatalf("expected third_level conflict, got nil")
	}
	if !strings.Contains(err.Error(), "already used by entrance") {
		t.Fatalf("custom-domain collision should report 'already used by entrance', got %q", err.Error())
	}
}

// Leading/trailing whitespace on the requested value is trimmed before the
// comparison, so a padded request still collides.
func TestCheckEntranceDomainDuplicate_ThirdLevelTrimmedBeforeCompare(t *testing.T) {
	a := setCustomDomain(t, dcApp("one", "alice", "oneid", false, "web"), "web", "taken", "")
	client := newDomainClient(a)

	err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "two", "x", "   taken  ", "")
	if err == nil || !strings.Contains(err.Error(), "already used by entrance") {
		t.Fatalf("whitespace-padded third_level must still collide, got %v", err)
	}
}

// A third_level_domain colliding with the DEFAULT subdomain of a shared app
// (which renders in every user's zone) is a conflict reported as "conflicts
// with the default domain". A single-entrance app uses the bare "<appid>".
func TestCheckEntranceDomainDuplicate_ThirdLevelConflictSharedDefault(t *testing.T) {
	shared := dcApp("cloud", "admin", "cloudid", true, "web")
	client := newDomainClient(shared)

	err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", "cloudid", "")
	if err == nil {
		t.Fatalf("expected default-domain conflict, got nil")
	}
	if !strings.Contains(err.Error(), "default domain") {
		t.Fatalf("shared default collision should mention the default domain, got %q", err.Error())
	}
}

// The default subdomain of a per-user app the caller OWNS also resolves in the
// caller's zone and therefore collides.
func TestCheckEntranceDomainDuplicate_ThirdLevelConflictOwnDefault(t *testing.T) {
	mine := dcApp("mine", "alice", "myid", false, "web")
	client := newDomainClient(mine)

	err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", "myid", "")
	if err == nil || !strings.Contains(err.Error(), "default domain") {
		t.Fatalf("expected own default-domain conflict, got %v", err)
	}
}

// Another user's per-user app renders under THEIR zone, so neither its default
// subdomain nor its custom domain can collide with the caller's request.
func TestCheckEntranceDomainDuplicate_ThirdLevelOtherUsersPerUserAppNoConflict(t *testing.T) {
	hers := setCustomDomain(t, dcApp("hers", "bob", "herid", false, "web"), "web", "taken", "")
	client := newDomainClient(hers)

	// Neither the default "herid" nor the custom "taken" is in alice's zone.
	if err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", "herid", ""); err != nil {
		t.Fatalf("other user's per-user default must not collide, got %v", err)
	}

	if err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", "taken", ""); err != nil {
		t.Fatalf("other user's per-user custom domain must not collide, got %v", err)
	}
}

// Editing an entrance never conflicts with that same entrance's own default
// subdomain (isSelf excludes the entry being edited).
func TestCheckEntranceDomainDuplicate_ThirdLevelSelfDefaultSkipped(t *testing.T) {
	shared := dcApp("cloud", "admin", "cloudid", true, "web")
	client := newDomainClient(shared)

	err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "cloud", "web", "cloudid", "")
	if err != nil {
		t.Fatalf("an entrance's own default domain must not conflict with itself, got %v", err)
	}
}

// On a shared app each user's customDomain lives in their own per-user overlay
// (Spec.UserSettings[user]) and resolves in that user's zone. A third_level the
// caller already claimed on the shared app therefore collides in the caller's
// own zone.
func TestCheckEntranceDomainDuplicate_ThirdLevelSharedOwnOverlayConflict(t *testing.T) {
	shared := dcApp("svc", "admin", "svcid", true, "web")
	setUserCustomDomain(t, shared, "bob", "web", "bobdom", "")
	client := newDomainClient(shared)

	err := checkEntranceDomainDuplicate(context.Background(), client, "bob", "bar", "x", "bobdom", "")
	if err == nil || !strings.Contains(err.Error(), "already used by entrance") {
		t.Fatalf("caller's own shared-app overlay domain must collide in their zone, got %v", err)
	}
}

// A third_level_domain set only in ANOTHER user's per-user overlay of a shared
// app is invisible to the caller: it resolves in that other user's zone, not
// the caller's, so it must not collide. This is the counterpart to
// TestCheckEntranceDomainDuplicate_ThirdLevelSharedOwnOverlayConflict — a
// per-user overlay is scoped to that one user's zone.
func TestCheckEntranceDomainDuplicate_ThirdLevelOtherUsersOverlayNoConflict(t *testing.T) {
	shared := dcApp("svc", "admin", "svcid", true, "web")
	setUserCustomDomain(t, shared, "admin", "web", "adminonly", "")
	client := newDomainClient(shared)

	err := checkEntranceDomainDuplicate(context.Background(), client, "bob", "bar", "x", "adminonly", "")
	if err != nil {
		t.Fatalf("another user's third_level overlay must not collide in the caller's zone, got %v", err)
	}
}

// A multi-entrance app derives positional defaults "<appid><index>"; requesting
// any of them collides as a default-domain conflict.
func TestCheckEntranceDomainDuplicate_ThirdLevelMultiEntranceDefault(t *testing.T) {
	shared := dcApp("multi", "admin", "mid", true, "web", "admin")
	client := newDomainClient(shared)

	for _, want := range []string{"mid0", "mid1"} {
		err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", want, "")
		if err == nil || !strings.Contains(err.Error(), "default domain") {
			t.Fatalf("positional default %q must collide, got %v", want, err)
		}
	}
}

// defaultThirdLevelDomainConfig overrides the positional default for a specific
// entrance of a multi-entrance app; the overridden value is what collides.
func TestCheckEntranceDomainDuplicate_ThirdLevelDefaultConfigOverride(t *testing.T) {
	shared := dcApp("multi", "admin", "mid", true, "web", "admin")
	setDefaultThirdLevelConfig(t, shared, "multi", "web", "webcustom")
	client := newDomainClient(shared)

	// The configured override collides...
	if err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", "webcustom", ""); err == nil ||
		!strings.Contains(err.Error(), "default domain") {
		t.Fatalf("configured default override must collide, got %v", err)
	}
	// ...and the positional value it replaced no longer does.
	if err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", "mid0", ""); err != nil {
		t.Fatalf("replaced positional default must not collide, got %v", err)
	}
}

// A value used by another entrance of the SAME app the caller is editing still
// conflicts (only the exact edited entrance is excluded).
func TestCheckEntranceDomainDuplicate_ThirdLevelSiblingEntranceConflict(t *testing.T) {
	// alice claimed "webdom" on the shared app's "web" entrance (her own
	// per-user overlay); reusing it on the sibling "admin" entrance collides.
	shared := dcApp("cloud", "admin", "cloudid", true, "web", "admin")
	setUserCustomDomain(t, shared, "alice", "web", "webdom", "")
	client := newDomainClient(shared)

	err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "cloud", "admin", "webdom", "")
	if err == nil || !strings.Contains(err.Error(), "already used by entrance") {
		t.Fatalf("sibling entrance's domain must collide, got %v", err)
	}
}

// A fresh, unused third_level and third_party value on a clean cluster is
// accepted.
func TestCheckEntranceDomainDuplicate_NoConflict(t *testing.T) {
	client := newDomainClient(dcApp("cloud", "admin", "cloudid", true, "web"))
	if err := checkEntranceDomainDuplicate(context.Background(), client, "alice", "bar", "x", "freshdom", "fresh.example.com"); err != nil {
		t.Fatalf("unused values must be accepted, got %v", err)
	}
}
