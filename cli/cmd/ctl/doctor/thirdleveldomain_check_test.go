package doctor

import (
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func sharedApp(name, owner, settingsBlob string, userSettings map[string]map[string]string) appv1alpha1.Application {
	return appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{appv1alpha1.AppSharedLabel: appv1alpha1.AppSharedTrue},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:  name,
			Owner: owner,
			Settings: map[string]string{
				"customDomain": settingsBlob,
			},
			UserSettings: userSettings,
		},
	}
}

func perUserApp(name, owner, settingsBlob string) appv1alpha1.Application {
	return appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1alpha1.ApplicationSpec{
			Name:  name,
			Owner: owner,
			Settings: map[string]string{
				"customDomain": settingsBlob,
			},
		},
	}
}

func blob(entrance, thirdLevel string) string {
	return `{"` + entrance + `":{"third_level_domain":"` + thirdLevel + `","third_party_domain":""}}`
}

func TestFindThirdLevelDomainIssues_clean(t *testing.T) {
	apps := []appv1alpha1.Application{
		perUserApp("files", "alice", blob("file", "files")),
		perUserApp("vault", "alice", blob("main", "vault")),
	}
	if got := FindThirdLevelDomainIssues(apps, []string{"alice"}); len(got) != 0 {
		t.Fatalf("want no issues, got %#v", got)
	}
}

func TestFindThirdLevelDomainIssues_duplicateInZone(t *testing.T) {
	apps := []appv1alpha1.Application{
		perUserApp("files", "alice", blob("file", "same")),
		perUserApp("vault", "alice", blob("main", "Same")), // case-insensitive
		perUserApp("notes", "bob", blob("main", "same")),  // other zone — ok
	}
	got := FindThirdLevelDomainIssues(apps, []string{"alice", "bob"})
	if len(got) != 2 {
		t.Fatalf("want 2 duplicate issues for alice, got %#v", got)
	}
	for _, iss := range got {
		if iss.User != "alice" || iss.Issue != "duplicate" {
			t.Fatalf("unexpected issue %#v", iss)
		}
	}
}

func TestFindThirdLevelDomainIssues_reserved(t *testing.T) {
	apps := []appv1alpha1.Application{
		perUserApp("evil", "alice", blob("web", "Auth")),
	}
	got := FindThirdLevelDomainIssues(apps, []string{"alice"})
	if len(got) != 1 || got[0].Issue != "reserved" || got[0].ThirdLevel != "Auth" {
		t.Fatalf("want one reserved issue, got %#v", got)
	}
}

func TestFindThirdLevelDomainIssues_sharedOverlayScopedPerUser(t *testing.T) {
	apps := []appv1alpha1.Application{
		sharedApp("chat", "admin", blob("web", "chat"), map[string]map[string]string{
			"alice": {"customDomain": blob("web", "dup")},
			"bob":   {"customDomain": blob("web", "dup")},
		}),
		perUserApp("files", "alice", blob("file", "dup")),
	}
	got := FindThirdLevelDomainIssues(apps, []string{"alice", "bob"})
	// alice: chat overlay + files both "dup" → 2 duplicate rows
	// bob: only one "dup" → clean
	alice := 0
	for _, iss := range got {
		if iss.User == "alice" && iss.Issue == "duplicate" {
			alice++
		}
		if iss.User == "bob" {
			t.Fatalf("bob should have no issues, got %#v", got)
		}
	}
	if alice != 2 {
		t.Fatalf("want 2 alice duplicates, got %#v", got)
	}
}

func TestFindThirdLevelDomainIssues_perUserOtherOwnerIgnored(t *testing.T) {
	apps := []appv1alpha1.Application{
		perUserApp("alice-app", "alice", blob("a", "x")),
		perUserApp("bob-app", "bob", blob("b", "x")),
	}
	got := FindThirdLevelDomainIssues(apps, []string{"alice"})
	if len(got) != 0 {
		t.Fatalf("bob's app must not collide in alice zone, got %#v", got)
	}
}

func TestCollectZoneUsers(t *testing.T) {
	apps := []appv1alpha1.Application{
		sharedApp("chat", "admin", "", map[string]map[string]string{
			"alice": {"customDomain": blob("web", "a")},
		}),
		perUserApp("files", "bob", blob("f", "b")),
	}
	got := CollectZoneUsers(apps, []string{"carol"})
	want := []string{"admin", "alice", "bob", "carol"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestPlanForceDedupeClears_keepsFirstClearsRest(t *testing.T) {
	apps := []appv1alpha1.Application{
		perUserApp("files", "alice", blob("file", "same")),
		perUserApp("vault", "alice", blob("main", "Same")),
	}
	ops := PlanForceDedupeClears(apps, []string{"alice"})
	if len(ops) != 1 {
		t.Fatalf("want 1 clear op, got %#v", ops)
	}
	if ops[0].App != "vault" || ops[0].Entrance != "main" || ops[0].Reason != "duplicate" {
		t.Fatalf("should clear vault/main (files kept first), got %#v", ops[0])
	}
	if ops[0].KeepApp != "files" || ops[0].KeepEntrance != "file" {
		t.Fatalf("keep target %#v", ops[0])
	}
}

func TestPlanForceDedupeClears_reserved(t *testing.T) {
	apps := []appv1alpha1.Application{
		perUserApp("evil", "alice", blob("web", "Auth")),
		perUserApp("ok", "alice", blob("main", "myapp")),
	}
	ops := PlanForceDedupeClears(apps, []string{"alice"})
	if len(ops) != 1 || ops[0].Reason != "reserved" || ops[0].App != "evil" {
		t.Fatalf("want one reserved clear on evil, got %#v", ops)
	}
	if err := ApplyClearOpsToApp(&apps[0], ops); err != nil {
		t.Fatal(err)
	}
	if parseCustomDomain(apps[0].Spec.Settings["customDomain"])["web"].thirdLevel != "" {
		t.Fatalf("reserved third_level should be cleared")
	}
	if remaining := FindThirdLevelDomainIssues(apps, []string{"alice"}); len(remaining) != 0 {
		t.Fatalf("want clean after reserved clear, got %#v", remaining)
	}
}

func TestApplyClearOpsToApp_perUserAndShared(t *testing.T) {
	apps := []appv1alpha1.Application{
		sharedApp("chat", "admin", blob("web", "chat"), map[string]map[string]string{
			"alice": {"customDomain": blob("web", "dup")},
		}),
		perUserApp("files", "alice", blob("file", "dup")),
	}
	ops := PlanForceDedupeClears(apps, []string{"alice"})
	if len(ops) != 1 {
		t.Fatalf("want 1 op, got %#v", ops)
	}
	// keep chat (lexicographically before files), clear files
	if ops[0].App != "files" {
		t.Fatalf("want clear files, got %#v", ops[0])
	}
	if err := ApplyClearOpsToApp(&apps[1], ops); err != nil {
		t.Fatal(err)
	}
	cfg := parseCustomDomain(apps[1].Spec.Settings["customDomain"])["file"]
	if cfg.thirdLevel != "" {
		t.Fatalf("files third_level should be cleared, got %q", cfg.thirdLevel)
	}
	if remaining := FindThirdLevelDomainIssues(apps, []string{"alice"}); len(remaining) != 0 {
		t.Fatalf("want no issues after clear, got %#v", remaining)
	}
}

func TestApplyClearOpsToApp_sharedExistingOverlayLoserCleared(t *testing.T) {
	// Shared app whose per-user overlay already holds the duplicate and is
	// the loser (aaa < zshared → keep aaa, clear the shared overlay). The
	// clear must come off the effective/overlay blob, not diverge from it.
	apps := []appv1alpha1.Application{
		perUserApp("aaa", "alice", blob("file", "dup")),
		sharedApp("zshared", "admin", blob("web", "other"), map[string]map[string]string{
			"alice": {"customDomain": blob("web", "dup")},
		}),
	}
	ops := PlanForceDedupeClears(apps, []string{"alice"})
	if len(ops) != 1 || !ops[0].Shared || ops[0].App != "zshared" || ops[0].Reason != "duplicate" {
		t.Fatalf("want clear shared zshared overlay, got %#v", ops)
	}
	if err := ApplyClearOpsToApp(&apps[1], ops); err != nil {
		t.Fatal(err)
	}
	if parseCustomDomain(apps[1].Spec.UserSettings["alice"]["customDomain"])["web"].thirdLevel != "" {
		t.Fatalf("alice overlay web third_level should be cleared, got %#v", apps[1].Spec.UserSettings)
	}
	// global Settings (admin's default) must be untouched
	if parseCustomDomain(apps[1].Spec.Settings["customDomain"])["web"].thirdLevel != "other" {
		t.Fatalf("global Settings must stay intact, got %#v", apps[1].Spec.Settings)
	}
	if remaining := FindThirdLevelDomainIssues(apps, []string{"alice"}); len(remaining) != 0 {
		t.Fatalf("want clean after force-dedupe, got %#v", remaining)
	}
}

func TestApplyClearOpsToApp_sharedInheritedWritesOverlay(t *testing.T) {
	// aaa sorts before chat → keep per-user, clear shared inherited value via overlay
	apps := []appv1alpha1.Application{
		perUserApp("aaa", "alice", blob("file", "dup")),
		sharedApp("chat", "admin", blob("web", "dup"), nil),
	}
	ops := PlanForceDedupeClears(apps, []string{"alice"})
	if len(ops) != 1 || !ops[0].Shared || ops[0].App != "chat" {
		t.Fatalf("want clear shared chat, got %#v", ops)
	}
	if err := ApplyClearOpsToApp(&apps[1], ops); err != nil {
		t.Fatal(err)
	}
	if parseCustomDomain(apps[1].Spec.Settings["customDomain"])["web"].thirdLevel != "dup" {
		t.Fatalf("global Settings must stay intact, got %#v", apps[1].Spec.Settings)
	}
	if parseCustomDomain(apps[1].Spec.UserSettings["alice"]["customDomain"])["web"].thirdLevel != "" {
		t.Fatalf("alice overlay should clear third_level, got %#v", apps[1].Spec.UserSettings)
	}
	if remaining := FindThirdLevelDomainIssues(apps, []string{"alice"}); len(remaining) != 0 {
		t.Fatalf("want clean after force-dedupe, got %#v", remaining)
	}
}
