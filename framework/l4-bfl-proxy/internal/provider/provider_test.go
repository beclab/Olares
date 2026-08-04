package provider

import (
	"encoding/json"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// customDomainJSON marshals a per-entrance customDomain settings blob.
func customDomainJSON(t *testing.T, m map[string]map[string]string) string {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal customDomain: %v", err)
	}
	return string(b)
}

func sharedApp(name, appid string, userSettings map[string]map[string]string) *appv1alpha1.Application {
	return &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{appv1alpha1.AppSharedLabel: appv1alpha1.AppSharedTrue},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:         name,
			Appid:        appid,
			UserSettings: userSettings,
		},
	}
}

func nonSharedApp(name, appid, owner string, settings map[string]string) *appv1alpha1.Application {
	return &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1alpha1.ApplicationSpec{
			Name:     name,
			Appid:    appid,
			Owner:    owner,
			Settings: settings,
		},
	}
}

func TestCustomDomainCertsForUser(t *testing.T) {
	t.Run("shared app with domain+cert+key in user overlay yields one cert", func(t *testing.T) {
		app := sharedApp("ollamaclient", "ollama123", map[string]map[string]string{
			"myxzxcvb": {
				settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
					"ollamaclient": {
						settingsCustomDomainThirdPartyDomain: "www.app-test-mayuxing.cn",
						settingsCustomDomainCert:             "cert-pem",
						settingsCustomDomainKey:              "key-pem",
					},
				}),
			},
		})

		certs := customDomainCertsForUser("myxzxcvb", []*appv1alpha1.Application{app})
		if len(certs) != 1 {
			t.Fatalf("want 1 cert, got %d", len(certs))
		}
		if certs[0].Domain != "www.app-test-mayuxing.cn" {
			t.Errorf("domain = %q, want www.app-test-mayuxing.cn", certs[0].Domain)
		}
		if certs[0].CertData != "cert-pem" || certs[0].KeyData != "key-pem" {
			t.Errorf("cert/key = %q/%q, want cert-pem/key-pem", certs[0].CertData, certs[0].KeyData)
		}
	})

	t.Run("shared app not configured by this user yields nothing", func(t *testing.T) {
		app := sharedApp("ollamaclient", "ollama123", map[string]map[string]string{
			"myxzxcvb": {
				settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
					"ollamaclient": {
						settingsCustomDomainThirdPartyDomain: "www.app-test-mayuxing.cn",
						settingsCustomDomainCert:             "cert-pem",
						settingsCustomDomainKey:              "key-pem",
					},
				}),
			},
		})

		// A different user viewing the same fanned-out shared app has no overlay.
		if certs := customDomainCertsForUser("myx05202", []*appv1alpha1.Application{app}); len(certs) != 0 {
			t.Fatalf("want 0 certs for user without overlay, got %d", len(certs))
		}
	})

	t.Run("empty third_party_domain yields nothing (orphaned-cert case)", func(t *testing.T) {
		app := sharedApp("ollamaclient", "ollama123", map[string]map[string]string{
			"myx05202": {
				settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
					"ollamaclient": {
						settingsCustomDomainThirdPartyDomain: "",
						settingsCustomDomainCert:             "",
						settingsCustomDomainKey:              "",
					},
				}),
			},
		})

		if certs := customDomainCertsForUser("myx05202", []*appv1alpha1.Application{app}); len(certs) != 0 {
			t.Fatalf("want 0 certs, got %d", len(certs))
		}
	})

	t.Run("domain set but cert not yet issued yields nothing", func(t *testing.T) {
		app := sharedApp("ollamaclient", "ollama123", map[string]map[string]string{
			"myxzxcvb": {
				settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
					"ollamaclient": {
						settingsCustomDomainThirdPartyDomain: "www.app-test-mayuxing.cn",
						settingsCustomDomainCert:             "",
						settingsCustomDomainKey:              "",
					},
				}),
			},
		})

		if certs := customDomainCertsForUser("myxzxcvb", []*appv1alpha1.Application{app}); len(certs) != 0 {
			t.Fatalf("want 0 certs, got %d", len(certs))
		}
	})

	t.Run("non-shared app reads customDomain from Spec.Settings", func(t *testing.T) {
		app := nonSharedApp("vault", "vault123", "alice", map[string]string{
			settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
				"vault-frontend": {
					settingsCustomDomainThirdPartyDomain: "vault.example.com",
					settingsCustomDomainCert:             "cert-pem",
					settingsCustomDomainKey:              "key-pem",
				},
			}),
		})

		certs := customDomainCertsForUser("alice", []*appv1alpha1.Application{app})
		if len(certs) != 1 {
			t.Fatalf("want 1 cert, got %d", len(certs))
		}
		if certs[0].Domain != "vault.example.com" {
			t.Errorf("domain = %q, want vault.example.com", certs[0].Domain)
		}
	})

	t.Run("duplicate domain across entrances and apps is de-duped", func(t *testing.T) {
		app1 := sharedApp("app1", "app1id", map[string]map[string]string{
			"alice": {
				settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
					"e1": {
						settingsCustomDomainThirdPartyDomain: "dup.example.com",
						settingsCustomDomainCert:             "cert-pem",
						settingsCustomDomainKey:              "key-pem",
					},
					"e2": {
						settingsCustomDomainThirdPartyDomain: "dup.example.com",
						settingsCustomDomainCert:             "cert-pem",
						settingsCustomDomainKey:              "key-pem",
					},
				}),
			},
		})
		app2 := sharedApp("app2", "app2id", map[string]map[string]string{
			"alice": {
				settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
					"e1": {
						settingsCustomDomainThirdPartyDomain: "dup.example.com",
						settingsCustomDomainCert:             "cert-pem",
						settingsCustomDomainKey:              "key-pem",
					},
				}),
			},
		})

		certs := customDomainCertsForUser("alice", []*appv1alpha1.Application{app1, app2})
		if len(certs) != 1 {
			t.Fatalf("want 1 de-duped cert, got %d", len(certs))
		}
		if certs[0].Domain != "dup.example.com" {
			t.Errorf("domain = %q, want dup.example.com", certs[0].Domain)
		}
	})

	t.Run("multiple distinct domains are sorted by domain", func(t *testing.T) {
		app := sharedApp("multi", "multiid", map[string]map[string]string{
			"alice": {
				settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
					"e1": {
						settingsCustomDomainThirdPartyDomain: "b.example.com",
						settingsCustomDomainCert:             "cert-pem",
						settingsCustomDomainKey:              "key-pem",
					},
					"e2": {
						settingsCustomDomainThirdPartyDomain: "a.example.com",
						settingsCustomDomainCert:             "cert-pem",
						settingsCustomDomainKey:              "key-pem",
					},
				}),
			},
		})

		certs := customDomainCertsForUser("alice", []*appv1alpha1.Application{app})
		if len(certs) != 2 {
			t.Fatalf("want 2 certs, got %d", len(certs))
		}
		if certs[0].Domain != "a.example.com" || certs[1].Domain != "b.example.com" {
			t.Errorf("domains = %q,%q, want a.example.com,b.example.com", certs[0].Domain, certs[1].Domain)
		}
	})
}
