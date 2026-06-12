package appcfg

import (
	"context"
	"strings"
	"testing"

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

	sharedID := appv1alpha1.SharedEntranceID(app.Spec.Appid, 0, 1)
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
	sharedID := appv1alpha1.SharedEntranceID(cfg.AppID, 0, 1)
	wantSharedURL := sharedID + ".shared.olares.com"
	if sharedEntrances[0].URL != wantSharedURL {
		t.Fatalf("shared URL = %q, want %q", sharedEntrances[0].URL, wantSharedURL)
	}
}
