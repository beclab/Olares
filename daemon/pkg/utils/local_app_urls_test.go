package utils

import (
	"reflect"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLocalAppEntrancesWithoutOwnerOrPublicZone(t *testing.T) {
	app := &appv1.Application{
		ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{appv1.AppApiVersionLabel: "v3", appv1.AppSharedLabel: "true"}},
		Spec: appv1.ApplicationSpec{
			Name: "olares-app", Appid: "olares-app", Owner: "olares",
			Entrances: []appv1.Entrance{{Name: "settings"}, {Name: "files"}},
			Settings:  map[string]string{"defaultThirdLevelDomainConfig": `[{"appName":"olares-app","entranceName":"settings","thirdLevelDomain":"settings"},{"appName":"olares-app","entranceName":"files","thirdLevelDomain":"files"}]`},
		},
	}
	before := app.DeepCopy()
	got, err := localAppEntrances(app, "guotestasd", "guotestasd.olares.local")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].URL != "settings.guotestasd.olares.local" || got[1].URL != "files.guotestasd.olares.local" {
		t.Fatalf("local URLs: %+v", got)
	}
	if !reflect.DeepEqual(app, before) {
		t.Fatal("mutated informer object")
	}
	// An entrance prefix override also applies when there is only one entrance.
	single := app.DeepCopy()
	single.Spec.Entrances = single.Spec.Entrances[:1]
	got, err = localAppEntrances(single, "guotestasd", "guotestasd.olares.local")
	if err != nil || got[0].URL != "settings.guotestasd.olares.local" {
		t.Fatalf("single entrance: %+v, %v", got, err)
	}
}

func TestLocalAppEntrancesLegacyIDsAndInvalidConfig(t *testing.T) {
	app := &appv1.Application{Spec: appv1.ApplicationSpec{Appid: "abc123", Owner: "alice", Entrances: []appv1.Entrance{{Name: "web"}, {Name: "api"}}}}
	got, err := localAppEntrances(app, "alice", "alice.olares.local")
	if err != nil || got[0].URL != "abc1230.alice.olares.local" || got[1].URL != "abc1231.alice.olares.local" {
		t.Fatalf("legacy URLs: %+v, %v", got, err)
	}
	app.Spec.Settings = map[string]string{"defaultThirdLevelDomainConfig": "invalid"}
	if _, err := localAppEntrances(app, "alice", "alice.olares.local"); err == nil {
		t.Fatal("invalid config must be retried, not cached as success")
	}
}
