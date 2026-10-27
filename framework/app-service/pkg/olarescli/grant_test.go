package olarescli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testApp      = "lares"
	testOwner    = "alice"
	testOlaresID = "alice@olares.cn"
	testNS       = "lares-alice"
)

// fakeLLDAP stands in for the derive endpoints and records what was asked of
// them, so tests can assert on the calls the store makes rather than only on
// the Secrets it leaves behind.
type fakeLLDAP struct {
	server    *httptest.Server
	minted    int
	revoked   []string
	deriveErr bool
}

func newFakeLLDAP(t *testing.T) *fakeLLDAP {
	t.Helper()
	f := &fakeLLDAP{}
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/token/derive", func(w http.ResponseWriter, r *http.Request) {
		if f.deriveErr {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		var body struct {
			Username string `json:"username"`
			TTLDays  int    `json:"ttl_days"`
			Label    string `json:"label"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.minted++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Grant{
			RefreshToken: "token-plaintext+" + body.Username,
			Username:     body.Username,
			ExpiresAt:    "2036-01-01T00:00:00Z",
		})
	})
	mux.HandleFunc("/auth/token/derive/revoke", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.revoked = append(f.revoked, body.RefreshToken)
	})
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func newStore(t *testing.T, lldap *fakeLLDAP, objects ...runtime.Object) (*Store, *fake.Clientset) {
	t.Helper()
	client := fake.NewSimpleClientset(objects...)
	cli := NewClient("sa-token")
	cli.baseURL = lldap.server.URL
	return NewStore(client, cli), client
}

func credentialSecret(t *testing.T, client *fake.Clientset) *corev1.Secret {
	t.Helper()
	secret, err := client.CoreV1().Secrets(testNS).
		Get(context.Background(), CredentialSecretName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("credential secret: %v", err)
	}
	return secret
}

func TestEnsureCredentialMintsOnce(t *testing.T) {
	lldap := newFakeLLDAP(t)
	store, client := newStore(t, lldap)
	ctx := context.Background()

	first, err := store.EnsureCredential(ctx, testApp, testOwner, testOlaresID, testNS)
	if err != nil {
		t.Fatalf("EnsureCredential: %v", err)
	}

	// Install retries land here again; a second token would leave the first
	// one valid for a decade with nothing pointing at it.
	second, err := store.EnsureCredential(ctx, testApp, testOwner, testOlaresID, testNS)
	if err != nil {
		t.Fatalf("EnsureCredential (retry): %v", err)
	}
	if lldap.minted != 1 {
		t.Errorf("minted %d grants, want 1", lldap.minted)
	}
	if first.RefreshToken != second.RefreshToken {
		t.Errorf("retry returned a different token: %q vs %q", first.RefreshToken, second.RefreshToken)
	}

	secret := credentialSecret(t, client)
	got := decodeCredentialFile(t, secret)
	if got.RefreshToken != "token-plaintext+alice" {
		t.Errorf("refresh_token = %q", got.RefreshToken)
	}
	if got.OlaresID != testOlaresID {
		t.Errorf("olaresId = %q, want %q", got.OlaresID, testOlaresID)
	}
	if got.AppName != testApp {
		t.Errorf("appName = %q, want %q", got.AppName, testApp)
	}
	if len(secret.Data) != 1 {
		t.Errorf("secret keys = %v, want only %s", keys(secret.Data), KeyCredentialFile)
	}
}

func TestEnsureCredentialReplacesIncompleteSecret(t *testing.T) {
	lldap := newFakeLLDAP(t)
	store, client := newStore(t, lldap, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CredentialSecretName,
			Namespace: testNS,
		},
		Data: map[string][]byte{"username": []byte(testOwner)},
	})

	if _, err := store.EnsureCredential(context.Background(), testApp, testOwner, testOlaresID, testNS); err != nil {
		t.Fatalf("EnsureCredential: %v", err)
	}
	if lldap.minted != 1 {
		t.Fatalf("minted %d grants, want 1", lldap.minted)
	}
	if got := decodeCredentialFile(t, credentialSecret(t, client)).RefreshToken; got == "" {
		t.Error("the half-written secret was left without a token")
	}
}

// Rerunning EnsureCredential has to keep the existing grant when the Secret
// is complete — an upgrade must not mint a second decade-long token.
func TestEnsureCredentialKeepsCompleteSecret(t *testing.T) {
	lldap := newFakeLLDAP(t)
	store, client := newStore(t, lldap, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: CredentialSecretName, Namespace: testNS},
		Data: map[string][]byte{
			KeyCredentialFile: []byte(`{"refreshToken":"kept-token","olaresId":"alice@olares.cn","appName":"lares"}`),
		},
	})

	grant, err := store.EnsureCredential(context.Background(), testApp, testOwner, testOlaresID, testNS)
	if err != nil {
		t.Fatalf("EnsureCredential: %v", err)
	}
	if lldap.minted != 0 {
		t.Errorf("minted %d grants for an already-complete secret", lldap.minted)
	}
	if grant.RefreshToken != "kept-token" {
		t.Errorf("refresh_token = %q, want the existing one", grant.RefreshToken)
	}
	got := decodeCredentialFile(t, credentialSecret(t, client))
	if got.RefreshToken != "kept-token" {
		t.Errorf("secret refresh_token = %q", got.RefreshToken)
	}
	if got.OlaresID != testOlaresID {
		t.Errorf("olaresId = %q", got.OlaresID)
	}
	if got.AppName != testApp {
		t.Errorf("appName = %q", got.AppName)
	}
}

func TestReleaseRevokesAndDeletes(t *testing.T) {
	lldap := newFakeLLDAP(t)
	store, client := newStore(t, lldap)
	ctx := context.Background()

	if _, err := store.EnsureCredential(ctx, testApp, testOwner, testOlaresID, testNS); err != nil {
		t.Fatalf("EnsureCredential: %v", err)
	}

	if err := store.Release(ctx, testApp, testOwner, testNS); err != nil {
		t.Fatalf("Release: %v", err)
	}

	if len(lldap.revoked) != 1 || lldap.revoked[0] != "token-plaintext+alice" {
		t.Errorf("revoked %v, want the refresh token itself", lldap.revoked)
	}
	if _, err := client.CoreV1().Secrets(testNS).Get(ctx, CredentialSecretName, metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Errorf("credential secret survived uninstall: %v", err)
	}
}

// Uninstall runs for apps that never had a credential, and for apps whose
// teardown is being retried after it already succeeded.
func TestReleaseWithoutGrantIsQuiet(t *testing.T) {
	lldap := newFakeLLDAP(t)
	store, _ := newStore(t, lldap)

	if err := store.Release(context.Background(), testApp, testOwner, testNS); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if len(lldap.revoked) != 0 {
		t.Errorf("revoked %v for an app that has no grant", lldap.revoked)
	}
}

func TestEnsureCredentialSurfacesDeriveFailure(t *testing.T) {
	lldap := newFakeLLDAP(t)
	lldap.deriveErr = true
	store, client := newStore(t, lldap)

	if _, err := store.EnsureCredential(context.Background(), testApp, testOwner, testOlaresID, testNS); err == nil {
		t.Fatal("expected the derive failure to be reported")
	}
	_, err := client.CoreV1().Secrets(testNS).
		Get(context.Background(), CredentialSecretName, metav1.GetOptions{})
	if !apierrors.IsNotFound(err) {
		t.Errorf("a secret was written despite the failed derive: %v", err)
	}
}

func TestEnsureCredentialFillsMissingAppName(t *testing.T) {
	lldap := newFakeLLDAP(t)
	store, client := newStore(t, lldap, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: CredentialSecretName, Namespace: testNS},
		Data: map[string][]byte{
			KeyCredentialFile: []byte(`{"refreshToken":"kept-token","olaresId":"alice@olares.cn"}`),
		},
	})

	grant, err := store.EnsureCredential(context.Background(), testApp, testOwner, testOlaresID, testNS)
	if err != nil {
		t.Fatalf("EnsureCredential: %v", err)
	}
	if lldap.minted != 0 {
		t.Errorf("minted %d grants while filling appName", lldap.minted)
	}
	if grant.RefreshToken != "kept-token" || grant.AppName != testApp {
		t.Errorf("grant = %+v", grant)
	}
	got := decodeCredentialFile(t, credentialSecret(t, client))
	if got.RefreshToken != "kept-token" || got.OlaresID != testOlaresID || got.AppName != testApp {
		t.Errorf("credential.json = %+v", got)
	}
}

func TestEnsureCredentialReplacesEmptyToken(t *testing.T) {
	lldap := newFakeLLDAP(t)
	store, client := newStore(t, lldap, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: CredentialSecretName, Namespace: testNS},
		Data: map[string][]byte{
			KeyCredentialFile: []byte(`{"refreshToken":"","olaresId":"alice@olares.cn"}`),
		},
	})

	grant, err := store.EnsureCredential(context.Background(), testApp, testOwner, testOlaresID, testNS)
	if err != nil {
		t.Fatalf("EnsureCredential: %v", err)
	}
	if lldap.minted != 1 {
		t.Fatalf("minted %d grants, want 1 for an empty refresh_token", lldap.minted)
	}
	if grant.RefreshToken == "" {
		t.Error("re-derive left the refresh_token empty")
	}
	if got := decodeCredentialFile(t, credentialSecret(t, client)); got.RefreshToken == "" {
		t.Error("the secret still has an empty refresh_token")
	}
}

func TestEnsureCredentialIgnoresLegacyTokenKey(t *testing.T) {
	lldap := newFakeLLDAP(t)
	store, client := newStore(t, lldap, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: CredentialSecretName, Namespace: testNS},
		Data: map[string][]byte{
			KeyCredentialFile: []byte(`{"refresh_token":"old-token","olaresId":"alice@olares.cn"}`),
		},
	})

	grant, err := store.EnsureCredential(context.Background(), testApp, testOwner, testOlaresID, testNS)
	if err != nil {
		t.Fatalf("EnsureCredential: %v", err)
	}
	if lldap.minted != 1 {
		t.Fatalf("minted %d grants, want 1 when the file only has refresh_token", lldap.minted)
	}
	if grant.RefreshToken == "old-token" {
		t.Error("kept the snake_case refresh_token instead of deriving a new grant")
	}
	got := decodeCredentialFile(t, credentialSecret(t, client))
	if got.RefreshToken == "old-token" || got.RefreshToken == "" {
		t.Errorf("credential.json = %+v", got)
	}
}

func decodeCredentialFile(t *testing.T, secret *corev1.Secret) struct {
	RefreshToken string `json:"refreshToken"`
	OlaresID     string `json:"olaresId"`
	AppName      string `json:"appName"`
} {
	t.Helper()
	var file struct {
		RefreshToken string `json:"refreshToken"`
		OlaresID     string `json:"olaresId"`
		AppName      string `json:"appName"`
	}
	raw := secret.Data[KeyCredentialFile]
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("credential.json: %v (%q)", err, raw)
	}
	return file
}

func keys(data map[string][]byte) []string {
	out := make([]string, 0, len(data))
	for k := range data {
		out = append(out, k)
	}
	return out
}
