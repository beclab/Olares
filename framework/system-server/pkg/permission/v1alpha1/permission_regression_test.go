package permission

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	sys "bytetrade.io/web3os/system-server/pkg/apis/sys/v1alpha1"
	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/constants"
	fake "bytetrade.io/web3os/system-server/pkg/generated/clientset/versioned/fake"
	listers "bytetrade.io/web3os/system-server/pkg/generated/listers/sys/v1alpha1"
	"github.com/emicklei/go-restful/v3"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/rest"
	kt "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

func permissionFixture(t *testing.T) (*PermissionControlSet, *fake.Clientset, cache.Indexer) {
	t.Helper()
	old := constants.MyNamespace
	constants.MyNamespace = "user-system-alice"
	t.Cleanup(func() { constants.MyNamespace = old })
	index := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	client := fake.NewSimpleClientset()
	return &PermissionControlSet{Ctrl: NewPermissionControl(client, listers.NewApplicationPermissionLister(index)), Mgr: NewAccessManager()}, client, index
}
func samplePermission() sys.PermissionRequire {
	return sys.PermissionRequire{Group: "notes", DataType: "document", Version: "v1", Ops: []string{"read"}}
}
func sampleApp() *sys.ApplicationPermission {
	return &sys.ApplicationPermission{ObjectMeta: metav1.ObjectMeta{Name: "notes", Namespace: constants.MyNamespace}, Spec: sys.ApplicationPermissionSpec{App: "notes", Key: "fixture-key", Secret: "fixture-secret", Permission: []sys.PermissionRequire{samplePermission()}}}
}

func TestPermissionRegistrationLifecycle(t *testing.T) {
	set, client, index := permissionFixture(t)
	reg := &PermissionRegister{App: "notes", AppID: "notes-id", Perm: []sys.PermissionRequire{samplePermission()}}
	created, err := set.Ctrl.applyPermission(t.Context(), reg)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := client.SysV1alpha1().ApplicationPermissions(constants.MyNamespace).Get(t.Context(), reg.App, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if created.AppKey == "" || len(created.AppSecret) != 16 || stored.Spec.Appid != reg.AppID || stored.Status.State != sys.Active {
		t.Fatal("registration failed to persist credentials and active permission")
	}
	if err = index.Add(stored); err != nil {
		t.Fatal(err)
	}
	reg.Perm[0].Ops = []string{"write"}
	updated, err := set.Ctrl.applyPermission(t.Context(), reg)
	if err != nil {
		t.Fatal(err)
	}
	if *updated != *created {
		t.Fatal("permission update rotated credentials")
	}
	stored, err = client.SysV1alpha1().ApplicationPermissions(constants.MyNamespace).Get(t.Context(), reg.App, metav1.GetOptions{})
	if err != nil || stored.Spec.Permission[0].Ops[0] != "write" {
		t.Fatal("permission update not persisted")
	}
	if err = set.Ctrl.deletePermission(t.Context(), reg.App); err != nil {
		t.Fatal(err)
	}
	if _, err = client.SysV1alpha1().ApplicationPermissions(constants.MyNamespace).Get(t.Context(), reg.App, metav1.GetOptions{}); !apierrors.IsNotFound(err) {
		t.Fatal("permission not deleted")
	}
	if err = set.Ctrl.deletePermission(t.Context(), "missing"); err != nil {
		t.Fatal("missing delete not idempotent")
	}
}

func TestPermissionWritesPropagateAPIFailures(t *testing.T) {
	for _, verb := range []string{"create", "update", "delete"} {
		t.Run(verb, func(t *testing.T) {
			set, client, index := permissionFixture(t)
			if verb != "create" {
				if err := index.Add(sampleApp()); err != nil {
					t.Fatal(err)
				}
			}
			client.PrependReactor(verb, "applicationpermissions", func(kt.Action) (bool, runtime.Object, error) { return true, nil, errors.New("fixture API unavailable") })
			var err error
			if verb == "delete" {
				err = set.Ctrl.deletePermission(t.Context(), "notes")
			} else {
				_, err = set.Ctrl.applyPermission(t.Context(), &PermissionRegister{App: "notes", Perm: []sys.PermissionRequire{samplePermission()}})
			}
			if err == nil || !strings.Contains(err.Error(), "fixture API unavailable") {
				t.Fatal("API failure hidden")
			}
		})
	}
}

func TestAccessTokenScopeAndExpiry(t *testing.T) {
	set, _, index := permissionFixture(t)
	if _, err := set.Ctrl.getAppPermissionFromAppKey(t.Context(), "unknown"); err == nil {
		t.Fatal("empty permission set authorized")
	}
	app := sampleApp()
	if err := index.Add(app); err != nil {
		t.Fatal(err)
	}
	foreign := app.DeepCopy()
	foreign.Namespace = "user-system-bob"
	foreign.Spec.Key = "bob-key"
	if err := index.Add(foreign); err != nil {
		t.Fatal(err)
	}
	if _, err := set.Ctrl.getAppPermissionFromAppKey(t.Context(), "bob-key"); err == nil {
		t.Fatal("cross-namespace key resolved")
	}
	resolved, err := set.Ctrl.getAppPermissionFromAppKey(t.Context(), app.Spec.Key)
	if err != nil || resolved.Name != app.Name {
		t.Fatal("local key not resolved")
	}
	now := time.Now().Unix()
	hash, err := bcrypt.GenerateFromPassword([]byte(app.Spec.Key+strconv.FormatInt(now, 10)+app.Spec.Secret), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	req := &AccessTokenRequest{AppKey: app.Spec.Key, Timestamp: now, Token: string(hash), Perm: samplePermission()}
	token, err := set.Mgr.getAccessToken(req, app)
	if err != nil {
		t.Fatal(err)
	}
	if token != base64.StdEncoding.EncodeToString(hash) {
		t.Fatal("access token encoding changed")
	}
	perm := samplePermission()
	perm.AppKey = app.Spec.Key
	set.Mgr.cacheAccessToken(token, &perm)
	key, err := ValidateAccessToken(token, "read", "document", "v1", "notes", set)
	if err != nil || key != app.Spec.Key {
		t.Fatal("authorized read rejected")
	}
	for _, tc := range []struct{ op, kind, version, group string }{{"write", "document", "v1", "notes"}, {"read", "document", "v1", "other"}, {"read", "other", "v1", "notes"}, {"read", "document", "v2", "notes"}} {
		if _, err := ValidateAccessToken(token, tc.op, tc.kind, tc.version, tc.group, set); err == nil {
			t.Fatal("out-of-scope access accepted")
		}
	}
	wire := httptest.NewRequest("GET", "/", nil)
	r := restful.NewRequest(wire)
	r.PathParameters()[api.ParamDataType] = "document"
	r.PathParameters()[api.ParamVersion] = "v1"
	r.PathParameters()[api.ParamGroup] = "notes"
	if _, err := ValidateAccessTokenWithRequest(token, "read", r, set); err != nil {
		t.Fatal(err)
	}
	set.Mgr.cache.Delete(token)
	if _, err := ValidateAccessToken(token, "read", "document", "v1", "notes", set); err == nil {
		t.Fatal("removed token accepted")
	}
	for _, stamp := range []int64{now - 60, now + 60} {
		req.Timestamp = stamp
		if _, err := set.Mgr.getAccessToken(req, app); err == nil {
			t.Fatal("stale/future token accepted")
		}
	}
	req.Timestamp = now
	req.Token = "invalid-fixture-hash"
	if _, err := set.Mgr.getAccessToken(req, app); err == nil {
		t.Fatal("forged hash accepted")
	}
	if !set.Ctrl.verifyPermission(app, &perm) {
		t.Fatal("allowed permission denied")
	}
	perm.Ops = []string{"write"}
	if set.Ctrl.verifyPermission(app, &perm) {
		t.Fatal("write escalated")
	}
}

func localSecretAPI(t *testing.T) (*rest.Config, *int) {
	t.Helper()
	mode := new(int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/namespaces/os-platform/secrets/lldap-credentials" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if *mode == 1 {
			w.WriteHeader(403)
			json.NewEncoder(w).Encode(&metav1.Status{Status: "Failure", Reason: metav1.StatusReasonForbidden, Code: 403})
			return
		}
		secret := core.Secret{TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"}, Data: map[string][]byte{}}
		if *mode != 2 {
			secret.Data["lldap-jwt-secret"] = []byte("local-fixture-signing-key")
		}
		json.NewEncoder(w).Encode(secret)
	}))
	t.Cleanup(server.Close)
	return &rest.Config{Host: server.URL}, mode
}
func signedUser(t *testing.T, user string, expiry int64) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{Username: user, StandardClaims: jwt.StandardClaims{ExpiresAt: expiry}})
	s, err := token.SignedString([]byte("local-fixture-signing-key"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func TestPermissionJWTValidation(t *testing.T) {
	cfg, mode := localSecretAPI(t)
	valid := signedUser(t, "alice", time.Now().Add(time.Minute).Unix())
	if user, err := validateToken(t.Context(), cfg, valid); err != nil || user != "alice" {
		t.Fatal("valid JWT rejected")
	}
	for _, token := range []string{"garbage", signedUser(t, "alice", time.Now().Add(-time.Minute).Unix()), valid + "forged"} {
		if _, err := validateToken(t.Context(), cfg, token); err == nil {
			t.Fatal("invalid JWT accepted")
		}
	}
	none := jwt.NewWithClaims(jwt.SigningMethodNone, &Claims{Username: "alice"})
	unsigned, err := none.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = validateToken(t.Context(), cfg, unsigned); err == nil {
		t.Fatal("unsigned identity accepted")
	}
	for _, m := range []int{1, 2} {
		*mode = m
		if _, err := validateToken(t.Context(), cfg, valid); err == nil {
			t.Fatal("missing/denied signing key accepted")
		}
	}
	if _, err := getLLdapJwtKey(context.Background(), &rest.Config{Host: "://invalid"}); err == nil {
		t.Fatal("invalid API config accepted")
	}
}

func TestLegacyRegistrationTenantBoundary(t *testing.T) {
	set, _, _ := permissionFixture(t)
	cfg, _ := localSecretAPI(t)
	h := newHandler(set, cfg)
	for _, method := range []struct {
		name string
		call func(*restful.Request, *restful.Response)
	}{{"register", h.register}, {"unregister", h.unregister}} {
		for _, tc := range []struct {
			name, user, body string
			want             int
		}{{"own", "alice", `{"app":"notes","perm":[]}`, 200}, {"cross-tenant", "bob", `{"app":"notes"}`, 401}, {"anonymous", "", `{}`, 401}, {"malformed", "alice", `{`, 500}} {
			t.Run(method.name+"/"+tc.name, func(t *testing.T) {
				wire := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
				wire.Header.Set("Content-Type", "application/json")
				wire.Header.Set("Remote-User", "alice")
				wire.Header.Set("X-BFL-USER", "alice")
				if tc.user != "" {
					wire.Header.Set(api.AuthorizationTokenHeader, signedUser(t, tc.user, time.Now().Add(time.Minute).Unix()))
				}
				res := httptest.NewRecorder()
				response := restful.NewResponse(res)
				response.SetRequestAccepts(restful.MIME_JSON)
				method.call(restful.NewRequest(wire), response)
				if res.Code != tc.want {
					t.Fatalf("status=%d want=%d", res.Code, tc.want)
				}
			})
		}
	}
}

func TestAppKeySignatureAndProviderScope(t *testing.T) {
	permissionFixture(t)
	app := sampleApp()
	provider := sys.ProviderRegistry{Spec: sys.ProviderRegistrySpec{Kind: sys.Provider, Group: "notes", Version: "v1", DataType: "document", OpApis: []sys.OpApisItem{{Name: "read", URI: "/read"}}}, Status: sys.ProviderRegistryStatus{State: sys.Active}}
	mode := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if mode == 4 {
			http.Error(w, "fixture error", 500)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/applicationpermissions") {
			items := []sys.ApplicationPermission{*app}
			if mode == 1 {
				items = nil
			}
			json.NewEncoder(w).Encode(sys.ApplicationPermissionList{TypeMeta: metav1.TypeMeta{Kind: "ApplicationPermissionList", APIVersion: "sys.bytetrade.io/v1alpha1"}, Items: items})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/providerregistries") {
			items := []sys.ProviderRegistry{provider}
			if mode == 2 {
				items = nil
			}
			if mode == 3 {
				items[0].Status.State = "inactive"
			}
			json.NewEncoder(w).Encode(sys.ProviderRegistryList{TypeMeta: metav1.TypeMeta{Kind: "ProviderRegistryList", APIVersion: "sys.bytetrade.io/v1alpha1"}, Items: items})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()
	config := filepath.Join(t.TempDir(), "kubeconfig")
	if err := os.WriteFile(config, []byte(fmt.Sprintf("apiVersion: v1\nkind: Config\nclusters:\n- name: fixture\n  cluster:\n    server: %s\ncontexts:\n- name: fixture\n  context:\n    cluster: fixture\n    user: fixture\ncurrent-context: fixture\nusers:\n- name: fixture\n  user: {}\n", server.URL)), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KUBECONFIG", config)
	signature := func() string {
		sum := sha256.Sum256([]byte(app.Spec.Key + app.Spec.Secret + strconv.FormatInt(time.Now().Truncate(time.Minute).Unix(), 10)))
		return hex.EncodeToString(sum[:])
	}
	if err := ValidateAppKey(app.Spec.Key, "read/item", "document", "v1", "notes", signature()); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAppKey(app.Spec.Key, "/write", "document", "v1", "notes", signature()); err == nil {
		t.Fatal("unauthorized URI accepted")
	}
	if err := ValidateAppKey(app.Spec.Key, "/read", "document", "v1", "notes", "forged"); err == nil {
		t.Fatal("forged signature accepted")
	}
	r := restful.NewRequest(httptest.NewRequest("GET", "/", nil))
	r.Request.Header.Set("X-Auth-Signature", "forged")
	if err := ValidateAppKeyWithRequest(app.Spec.Key, r); err == nil {
		t.Fatal("request without matching scope accepted")
	}
	for _, m := range []int{1, 2, 3, 4} {
		mode = m
		if err := ValidateAppKey(app.Spec.Key, "/read", "document", "v1", "notes", signature()); err == nil {
			t.Fatalf("missing permission/provider or API failure accepted: %d", m)
		}
	}
}

func TestLegacyAccessGrantRequiresProofAndScope(t *testing.T) {
	set, _, index := permissionFixture(t)
	app := sampleApp()
	if err := index.Add(app); err != nil {
		t.Fatal(err)
	}
	h := newHandler(set, nil)
	for _, tc := range []string{"allowed", "malformed", "unknown-key", "stale", "denied-scope"} {
		t.Run(tc, func(t *testing.T) {
			stamp := time.Now().Unix()
			key := app.Spec.Key
			permission := samplePermission()
			if tc == "stale" {
				stamp -= 60
			}
			if tc == "unknown-key" {
				key = "other-key"
			}
			if tc == "denied-scope" {
				permission.Ops = []string{"write"}
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(key+strconv.FormatInt(stamp, 10)+app.Spec.Secret), bcrypt.MinCost)
			if err != nil {
				t.Fatal(err)
			}
			entity, err := json.Marshal(AccessTokenRequest{AppKey: key, Timestamp: stamp, Token: string(hash), Perm: permission})
			if err != nil {
				t.Fatal(err)
			}
			if tc == "malformed" {
				entity = []byte("{")
			}
			wire := httptest.NewRequest("POST", "/", strings.NewReader(string(entity)))
			wire.Header.Set("Content-Type", restful.MIME_JSON)
			recorder := httptest.NewRecorder()
			response := restful.NewResponse(recorder)
			response.SetRequestAccepts(restful.MIME_JSON)
			h.auth(restful.NewRequest(wire), response)
			var result struct {
				Code int                 `json:"code"`
				Data AccessTokenResponse `json:"data"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if tc == "allowed" {
				if result.Code != 0 || result.Data.AccessToken == "" || !result.Data.ExpiredAt.After(time.Now()) {
					t.Fatal("valid proof did not create a grant")
				}
				granted, err := ValidateAccessToken(result.Data.AccessToken, "read", "document", "v1", "notes", set)
				if err != nil || granted != app.Spec.Key {
					t.Fatal("issued grant could not authorize its scope")
				}
			} else if result.Code == 0 || result.Data.AccessToken != "" {
				t.Fatal("invalid proof or excess scope issued a grant")
			}
		})
	}
}
