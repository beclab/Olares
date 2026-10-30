package appcfg

import (
	"context"
	sys "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func withMockZone(zone string) func() {
	old := cluster.NewDomainReader
	scheme := runtime.NewScheme()
	_ = sys.AddToScheme(scheme)
	reader := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&sys.Terminus{ObjectMeta: metav1.ObjectMeta{Name: "terminus"}, Spec: sys.TerminusSpec{Settings: map[string]string{"domainName": zone}}}).Build()
	cluster.NewDomainReader = func() (client.Reader, error) { return reader, nil }
	return func() { cluster.NewDomainReader = old }
}
func persistURLFixture(a *Application) {
	for i := range a.Spec.Entrances {
		a.Spec.Entrances[i].ID = appv1alpha1.EntranceID(a.Spec.Appid, i, len(a.Spec.Entrances))
	}
	for i := range a.Spec.SharedEntrances {
		if IsShared(a) {
			a.Spec.SharedEntrances[i].ID = appv1alpha1.SharedEntranceID(a.Spec.Appid, i, len(a.Spec.SharedEntrances))
		} else {
			a.Spec.SharedEntrances[i].ID = appv1alpha1.SharedEntranceIDV2(a.Spec.Appid, i, len(a.Spec.SharedEntrances))
		}
	}
}

func TestGenSharedEntranceURLSingleEntranceUsesApiSharedID(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	restore := withMockZone("alice.olares.com")
	defer restore()

	app := &Application{
		Spec: ApplicationSpec{
			Appid: "app1abcd",
			Owner: "alice",
			SharedEntrances: []Entrance{
				{Name: "api", Port: 8080},
			},
		},
	}

	persistURLFixture(app)
	got, err := GenSharedEntranceURL(context.Background(), app)
	if err != nil {
		t.Fatalf("GenSharedEntranceURL: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 shared entrance, got %d", len(got))
	}

	sharedID := appv1alpha1.SharedEntranceIDV2(app.Spec.Appid, 0, 1)
	want := sharedID + ".shared.olares.com:8080"
	if got[0].URL != want {
		t.Fatalf("shared URL = %q, want %q", got[0].URL, want)
	}
	if strings.Contains(got[0].URL, sharedID+"0.") {
		t.Fatalf("single shared entrance URL must not contain index suffix: %q", got[0].URL)
	}
}

func TestGenEntranceURLPreservesCustomThirdLevelDomain(t *testing.T) {
	restore := withMockZone("alice.olares.com")
	defer restore()

	app := &Application{
		Spec: ApplicationSpec{
			Appid: "customid",
			Name:  "demo",
			Owner: "alice",
			Settings: map[string]string{
				"defaultThirdLevelDomainConfig": `[{"appName":"demo","entranceName":"admin","thirdLevelDomain":"custom-admin"}]`,
			},
			Entrances: []Entrance{
				{Name: "web"},
				{Name: "admin"},
			},
		},
	}

	persistURLFixture(app)
	got, err := GenEntranceURL(context.Background(), app)
	if err != nil {
		t.Fatalf("GenEntranceURL: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entrances, got %d", len(got))
	}
	if got[0].URL != "customid0.alice.olares.com" {
		t.Fatalf("entrance[0].URL = %q, want customid0.alice.olares.com", got[0].URL)
	}
	if got[1].URL != "custom-admin.alice.olares.com" {
		t.Fatalf("entrance[1].URL = %q, want custom-admin.alice.olares.com", got[1].URL)
	}
}

func TestApplicationConfigGenURLPassesAppID(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	restore := withMockZone("alice.olares.com")
	defer restore()

	cfg := &ApplicationConfig{
		AppID:     "appidfromcfg",
		AppName:   "name-not-used-for-id",
		OwnerName: "alice",
		Entrances: []Entrance{
			{Name: "web"},
		},
		SharedEntrances: []Entrance{
			{Name: "api"},
		},
	}

	cfg.Entrances[0].ID = cfg.AppID
	cfg.SharedEntrances[0].ID = appv1alpha1.SharedEntranceIDV2(cfg.AppID, 0, 1)
	entrances, err := cfg.GenEntranceURL(context.Background())
	if err != nil {
		t.Fatalf("ApplicationConfig.GenEntranceURL: %v", err)
	}
	if len(entrances) != 1 || entrances[0].URL != "appidfromcfg.alice.olares.com" {
		t.Fatalf("entrances URL = %+v, want appidfromcfg.alice.olares.com", entrances)
	}

	sharedEntrances, err := cfg.GenSharedEntranceURL(context.Background())
	if err != nil {
		t.Fatalf("ApplicationConfig.GenSharedEntranceURL: %v", err)
	}
	if len(sharedEntrances) != 1 {
		t.Fatalf("expected 1 shared entrance, got %d", len(sharedEntrances))
	}
	sharedID := appv1alpha1.SharedEntranceIDV2(cfg.AppID, 0, len(sharedEntrances))
	wantSharedURL := sharedID + ".shared.olares.com"
	if sharedEntrances[0].URL != wantSharedURL {
		t.Fatalf("shared URL = %q, want %q", sharedEntrances[0].URL, wantSharedURL)
	}
}

func TestGenSharedEntranceURLUsesPlatformDomain(t *testing.T) {
	restore := withMockZone("alice.unrelated.example")
	defer restore()
	cluster.PrimePlatformDomainForTest("owner.org.example.cn")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	app := &Application{Spec: ApplicationSpec{Appid: "demo", Owner: "alice", SharedEntrances: []Entrance{{Name: "web", Port: 8080}}}}
	persistURLFixture(app)
	got, err := GenSharedEntranceURL(context.Background(), app)
	if err != nil {
		t.Fatal(err)
	}
	want := appv1alpha1.SharedEntranceIDV2("demo", 0, 1) + ".shared.olares.com:8080"
	if got[0].URL != want {
		t.Fatalf("got %q, want %q", got[0].URL, want)
	}
}

func TestGenSharedEntranceURLMissingPlatformDomain(t *testing.T) {
	cluster.PrimePlatformDomainForTest("")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	app := &Application{Spec: ApplicationSpec{SharedEntrances: []Entrance{{Name: "web"}}}}
	if _, err := GenSharedEntranceURL(context.Background(), app); err == nil {
		t.Fatal("expected platform domain error")
	}
	if _, err := GenSharedEntranceURL(context.Background(), &Application{}); err != nil {
		t.Fatalf("empty shared entrances: %v", err)
	}
}

func TestGenSharedEntranceURLSharedAppMultipleEntrances(t *testing.T) {
	cluster.PrimePlatformDomainForTest("org.example.com")
	t.Cleanup(cluster.ResetPlatformDomainForTest)
	app := &Application{Spec: ApplicationSpec{Appid: "demo", SharedEntrances: []Entrance{{Name: "web"}, {Name: "api", Port: 9000}}}}
	app.Labels = map[string]string{appv1alpha1.AppSharedLabel: appv1alpha1.AppSharedTrue}
	persistURLFixture(app)
	got, err := GenSharedEntranceURL(context.Background(), app)
	if err != nil {
		t.Fatal(err)
	}
	for i := range got {
		want := appv1alpha1.SharedEntranceID("demo", i, len(got)) + ".shared.olares.com"
		if i == 1 {
			want += ":9000"
		}
		if got[i].URL != want {
			t.Fatalf("entrance %d URL=%q, want %q", i, got[i].URL, want)
		}
	}
}

func TestRuntimeURLsUseSavedIDs(t *testing.T) {
	restore := withMockZone("machine.example.com")
	defer restore()
	a := &Application{Spec: ApplicationSpec{Appid: "abcd1234", Owner: "bob", Entrances: []Entrance{{Name: "web", ID: "saved-web"}}, SharedEntrances: []Entrance{{Name: "shared", ID: "saved-share"}}}}
	normal, err := GenEntranceURL(context.Background(), a)
	if err != nil || normal[0].URL != "saved-web.machine.example.com" {
		t.Fatal(normal, err)
	}
	shared, err := GenSharedEntranceURL(context.Background(), a)
	if err != nil || shared[0].URL != "saved-share.shared.olares.com" {
		t.Fatal(shared, err)
	}
	a.Spec.Entrances[0].ID = ""
	if _, err := GenEntranceURL(context.Background(), a); err == nil {
		t.Fatal("runtime recomputed missing ID")
	}
	a.Spec.SharedEntrances[0].ID = ""
	if _, err := GenSharedEntranceURL(context.Background(), a); err == nil {
		t.Fatal("runtime recomputed missing shared ID")
	}
}
