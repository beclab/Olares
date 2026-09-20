package v2alpha1

import (
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type fakeAppLister struct {
	apps []*appv1alpha1.Application
	err  error
}

func (f *fakeAppLister) List(selector labels.Selector) ([]*appv1alpha1.Application, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.apps, nil
}

func TestResolveProviderRefFromHost(t *testing.T) {
	nonSharedApp := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "demo"},
		Spec: appv1alpha1.ApplicationSpec{
			Appid: "appidabc",
			Name:  "demo",
			Owner: "alice",
			Entrances: []appv1alpha1.Entrance{
				{Name: "web"},
				{Name: "admin"},
			},
			Settings: map[string]string{
				"defaultThirdLevelDomainConfig": `[{"appName":"demo","entranceName":"admin","thirdLevelDomain":"custom-admin"}]`,
				"customDomain": `{
					"web":{"third_level_domain":"myweb","third_party_domain":""},
					"admin":{"third_level_domain":"","third_party_domain":"admin.example.com"}
				}`,
			},
		},
	}

	sharedApp := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "sharedemo",
			Labels: map[string]string{
				appv1alpha1.AppSharedLabel: appv1alpha1.AppSharedTrue,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Appid: "sharexyz",
			Name:  "sharedemo",
			Owner: "admin",
			Entrances: []appv1alpha1.Entrance{
				{Name: "web"},
			},
			Settings: map[string]string{},
			UserSettings: map[string]map[string]string{
				"bob": {
					"customDomain": `{"web":{"third_level_domain":"bobalias","third_party_domain":"bob.custom.io"}}`,
				},
			},
		},
	}

	lister := &fakeAppLister{apps: []*appv1alpha1.Application{nonSharedApp, sharedApp}}

	tests := []struct {
		name   string
		host   string
		lister ApplicationLister
		want   string
	}{
		{
			name:   "no alias falls back to ProviderRefFromHost",
			host:   "appidabc0.alice.olares.com",
			lister: lister,
			want:   ProviderRefFromHost("appidabc0.alice.olares.com"),
		},
		{
			name:   "third_level_domain maps to appid index",
			host:   "myweb.alice.olares.com",
			lister: lister,
			want:   "alice/appidabc0",
		},
		{
			name:   "third_party_domain maps to defaultThirdLevelDomain",
			host:   "admin.example.com",
			lister: lister,
			want:   "alice/custom-admin",
		},
		{
			name:   "third_level zone mismatch does not match",
			host:   "myweb.bob.olares.com",
			lister: lister,
			want:   ProviderRefFromHost("myweb.bob.olares.com"),
		},
		{
			name:   "shared user third_level maps to owner canonical",
			host:   "bobalias.bob.olares.com",
			lister: lister,
			want:   "admin/sharexyz",
		},
		{
			name:   "shared user third_party maps to owner canonical",
			host:   "bob.custom.io",
			lister: lister,
			want:   "admin/sharexyz",
		},
		{
			name:   "host with port is normalized",
			host:   "admin.example.com:443",
			lister: lister,
			want:   "alice/custom-admin",
		},
		{
			name:   "nil lister falls back",
			host:   "myweb.alice.olares.com",
			lister: nil,
			want:   ProviderRefFromHost("myweb.alice.olares.com"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveProviderRefFromHost(tt.host, tt.lister)
			if got != tt.want {
				t.Fatalf("ResolveProviderRefFromHost(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}

var _ ApplicationLister = (*fakeAppLister)(nil)
