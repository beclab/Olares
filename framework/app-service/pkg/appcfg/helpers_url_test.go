package appcfg

import (
	"context"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func withMockZone(zone string) func() {
	orig := getUserZone
	getUserZone = func(context.Context, string) (string, error) {
		return zone, nil
	}
	return func() {
		getUserZone = orig
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
	got, err := GenSharedEntranceURL(context.Background(), app)
	if err != nil {
		t.Fatal(err)
	}
	want := appv1alpha1.SharedEntranceIDV2("demo", 0, 1) + ".shared.owner.org.example.cn:8080"
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
	got, err := GenSharedEntranceURL(context.Background(), app)
	if err != nil {
		t.Fatal(err)
	}
	for i := range got {
		want := appv1alpha1.SharedEntranceID("demo", i, len(got)) + ".shared.org.example.com"
		if i == 1 {
			want += ":9000"
		}
		if got[i].URL != want {
			t.Fatalf("entrance %d URL=%q, want %q", i, got[i].URL, want)
		}
	}
}
