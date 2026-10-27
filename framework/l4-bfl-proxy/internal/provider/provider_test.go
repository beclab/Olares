package provider

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func mustTestKeyPair(t *testing.T) (certPEM, keyPEM string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test.example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}))
	return certPEM, keyPEM
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
	certPEM, keyPEM := mustTestKeyPair(t)

	t.Run("shared app with domain+cert+key in user overlay yields one cert", func(t *testing.T) {
		app := sharedApp("ollamaclient", "ollama123", map[string]map[string]string{
			"myxzxcvb": {
				settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
					"ollamaclient": {
						settingsCustomDomainThirdPartyDomain: "www.app-test-mayuxing.cn",
						settingsCustomDomainCert:             certPEM,
						settingsCustomDomainKey:              keyPEM,
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
		if certs[0].CertData != certPEM || certs[0].KeyData != keyPEM {
			t.Errorf("cert/key mismatch")
		}
	})

	t.Run("shared app not configured by this user yields nothing", func(t *testing.T) {
		app := sharedApp("ollamaclient", "ollama123", map[string]map[string]string{
			"myxzxcvb": {
				settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
					"ollamaclient": {
						settingsCustomDomainThirdPartyDomain: "www.app-test-mayuxing.cn",
						settingsCustomDomainCert:             certPEM,
						settingsCustomDomainKey:              keyPEM,
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

	t.Run("placeholder cert/key like test/test is skipped", func(t *testing.T) {
		app := nonSharedApp("qbittorrent", "qbittorrent", "olaresid", map[string]string{
			settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
				"qbittorrent-svc": {
					settingsCustomDomainThirdPartyDomain: "com.sd",
					settingsCustomDomainCert:             "test",
					settingsCustomDomainKey:              "test",
				},
			}),
		})

		if certs := customDomainCertsForUser("olaresid", []*appv1alpha1.Application{app}); len(certs) != 0 {
			t.Fatalf("want 0 certs for invalid PEM placeholder, got %d", len(certs))
		}
	})

	t.Run("non-shared app reads customDomain from Spec.Settings", func(t *testing.T) {
		app := nonSharedApp("vault", "vault123", "alice", map[string]string{
			settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
				"vault-frontend": {
					settingsCustomDomainThirdPartyDomain: "vault.example.com",
					settingsCustomDomainCert:             certPEM,
					settingsCustomDomainKey:              keyPEM,
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
						settingsCustomDomainCert:             certPEM,
						settingsCustomDomainKey:              keyPEM,
					},
					"e2": {
						settingsCustomDomainThirdPartyDomain: "dup.example.com",
						settingsCustomDomainCert:             certPEM,
						settingsCustomDomainKey:              keyPEM,
					},
				}),
			},
		})
		app2 := sharedApp("app2", "app2id", map[string]map[string]string{
			"alice": {
				settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
					"e1": {
						settingsCustomDomainThirdPartyDomain: "dup.example.com",
						settingsCustomDomainCert:             certPEM,
						settingsCustomDomainKey:              keyPEM,
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
						settingsCustomDomainCert:             certPEM,
						settingsCustomDomainKey:              keyPEM,
					},
					"e2": {
						settingsCustomDomainThirdPartyDomain: "a.example.com",
						settingsCustomDomainCert:             certPEM,
						settingsCustomDomainKey:              keyPEM,
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

	t.Run("valid domain kept when sibling entrance has invalid PEM", func(t *testing.T) {
		app := nonSharedApp("multi", "multi", "alice", map[string]string{
			settingsCustomDomain: customDomainJSON(t, map[string]map[string]string{
				"bad": {
					settingsCustomDomainThirdPartyDomain: "bad.example.com",
					settingsCustomDomainCert:             "test",
					settingsCustomDomainKey:              "test",
				},
				"good": {
					settingsCustomDomainThirdPartyDomain: "good.example.com",
					settingsCustomDomainCert:             certPEM,
					settingsCustomDomainKey:              keyPEM,
				},
			}),
		})

		certs := customDomainCertsForUser("alice", []*appv1alpha1.Application{app})
		if len(certs) != 1 {
			t.Fatalf("want 1 cert, got %d", len(certs))
		}
		if certs[0].Domain != "good.example.com" {
			t.Errorf("domain = %q, want good.example.com", certs[0].Domain)
		}
	})
}

func TestListApplicationDetails_NoImplicitHeadscalePublicApp(t *testing.T) {
	p := &Provider{}
	publicApps, _ := p.listApplicationDetails("alice", nil)
	if len(publicApps) != 0 {
		t.Fatalf("publicApps = %v, want empty", publicApps)
	}
}

// Only the cluster data used by listUsers is needed for this policy test.
type userPolicyCache struct {
	ctrlcache.Cache
}

func (c *userPolicyCache) List(_ context.Context, list client.ObjectList, _ ...client.ListOption) error {
	switch dst := list.(type) {
	case *corev1.PodList:
		*dst = corev1.PodList{}
	case *corev1.NodeList:
		*dst = corev1.NodeList{Items: []corev1.Node{{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "control-plane",
				Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""},
			},
		}}}
	default:
		return fmt.Errorf("unexpected list %T", list)
	}
	return nil
}

func TestListUsersKeepsEachUsersIdentityAndPolicy(t *testing.T) {
	creator := iamv1alpha2.User{ObjectMeta: metav1.ObjectMeta{
		Name: "alice",
		Annotations: map[string]string{
			userAnnotationDid:       "did:example:alice",
			userLauncherAccessLevel: "1",
			userLauncherAllowCIDR:   "192.168.1.0/24",
			userDenyAllPolicy:       "0",
		},
	}}
	child := iamv1alpha2.User{ObjectMeta: metav1.ObjectMeta{
		Name: "bob",
		Annotations: map[string]string{
			"bytetrade.io/creator":      "alice",
			"bytetrade.io/is-ephemeral": "true",
			userAnnotationDid:           "did:example:bob",
			userLauncherAccessLevel:     "3",
			userLauncherAllowCIDR:       "10.0.0.0/24, 172.16.0.0/16",
			userDenyAllPolicy:           "1",
		},
	}}
	for _, tc := range []struct {
		name  string
		users []iamv1alpha2.User
	}{
		{name: "creator exists", users: []iamv1alpha2.User{creator, child}},
		{name: "creator deleted", users: []iamv1alpha2.User{child}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := New(&userPolicyCache{}, nil, &Config{UserNamespacePrefix: "user-space"})
			users, err := p.listUsers(context.Background(), tc.users, nil, nil, "machine.olares.cn", "machine.olares.local")
			require.NoError(t, err)
			require.Len(t, users, len(tc.users), "a user must not disappear with its creator")
			for i, user := range users {
				assert.Equal(t, "machine.olares.cn", user.Zone)
				assert.Equal(t, tc.users[i].Annotations[userAnnotationDid], user.Did)
				switch user.Name {
				case "alice":
					assert.Equal(t, uint64(1), user.AccessLevel)
					assert.Equal(t, []string{"192.168.1.0/24"}, user.AllowCIDRs)
					assert.False(t, user.DenyAll)
				case "bob":
					assert.Equal(t, uint64(3), user.AccessLevel)
					assert.Equal(t, []string{"10.0.0.0/24", "172.16.0.0/16"}, user.AllowCIDRs)
					assert.True(t, user.DenyAll)
				default:
					t.Fatalf("unexpected user %q", user.Name)
				}
			}
		})
	}
}
