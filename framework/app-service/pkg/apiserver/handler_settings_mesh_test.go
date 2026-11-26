package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	"github.com/beclab/Olares/framework/app-service/pkg/users"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	applisters "github.com/beclab/api/pkg/generated/listers/app.bytetrade.io/v1alpha1"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestSettingsMeshReadsRespectCurrentTenant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/apis/iam.kubesphere.io/v1alpha2/users/alice", "/apis/iam.kubesphere.io/v1alpha2/users/bob":
			_, _ = w.Write([]byte(`{"apiVersion":"iam.kubesphere.io/v1alpha2","kind":"User"}`))
		case "/apis/app.bytetrade.io/v1alpha1/applicationmanagers":
			_, _ = w.Write([]byte(`{"apiVersion":"app.bytetrade.io/v1alpha1","kind":"ApplicationManagerList","items":[]}`))
		default:
			t.Errorf("unexpected API lookup %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := clientcmdapi.Config{
		CurrentContext: "test",
		Contexts:       map[string]*clientcmdapi.Context{"test": {Cluster: "test"}},
		Clusters:       map[string]*clientcmdapi.Cluster{"test": {Server: server.URL}},
	}
	configPath := filepath.Join(t.TempDir(), "kubeconfig")
	if err := clientcmd.WriteToFile(config, configPath); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KUBECONFIG", configPath)
	t.Setenv("SYS_APPS", "olares-app")

	scheme := testutil.NewScheme()
	if err := iamv1alpha2.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	appIndex := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	managerIndex := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	builder := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			if key.Name == "env-store-error-alice" {
				return errors.New("environment store unavailable")
			}
			return c.Get(ctx, key, obj, opts...)
		},
	})
	for name, role := range map[string]string{"alice": "normal", "bob": "normal", "operator": "admin", "invalid-role": ""} {
		builder.WithObjects(&iamv1alpha2.User{ObjectMeta: metav1.ObjectMeta{
			Name: name, Annotations: map[string]string{users.UserAnnotationOwnerRole: role},
		}})
		if role != "normal" {
			continue
		}
		if err := appIndex.Add(&appv1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{Name: "olares-app-" + name},
			Spec:       appv1alpha1.ApplicationSpec{Owner: name, Name: "olares-app", IsSysApp: true, Namespace: "user-space-" + name},
		}); err != nil {
			t.Fatal(err)
		}
		if err := managerIndex.Add(&appv1alpha1.ApplicationManager{
			ObjectMeta: metav1.ObjectMeta{Name: "example-app-" + name},
			Spec: appv1alpha1.ApplicationManagerSpec{
				AppName: "example-app", AppOwner: name, AppNamespace: "example-app-" + name, Type: appv1alpha1.App, Config: "{}",
			},
			Status: appv1alpha1.ApplicationManagerStatus{State: appv1alpha1.Running},
		}); err != nil {
			t.Fatal(err)
		}

		builder.WithObjects(&sysv1alpha1.AppEnv{
			ObjectMeta: metav1.ObjectMeta{Name: "olares-app-" + name, Namespace: "user-space-" + name},
			AppName:    "olares-app", AppOwner: name,
			Envs: []sysv1alpha1.AppEnvVar{{EnvVarSpec: sysv1alpha1.EnvVarSpec{EnvName: "FIXTURE", Value: name}}},
		})
	}

	h := &Handler{
		ctrlClient: builder.Build(), appLister: applisters.NewApplicationLister(appIndex),
		appmgrLister: applisters.NewApplicationManagerLister(managerIndex),
	}
	ws := new(restful.WebService).Path("/app-service/v1").Produces(restful.MIME_JSON)
	ws.Filter(h.authenticate)
	ws.Route(ws.GET("/control-hub/apps").To(h.controlHubApps))
	ws.Route(ws.GET("/apps/{" + ParamAppName + "}/env").To(h.getAppEnv))
	container := restful.NewContainer().Add(ws)

	request := func(path, caller string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/app-service/v1"+path, nil)
		if caller != "" {
			req.Header.Set(constants.BflUserKey, caller)
		}
		resp := httptest.NewRecorder()
		container.ServeHTTP(resp, req)
		return resp
	}
	for _, tc := range []struct {
		caller string
		owners []string
	}{
		{"alice", []string{"alice", "alice"}},
		{"bob", []string{"bob", "bob"}},
		{"operator", []string{"alice", "alice", "bob", "bob"}},
	} {
		t.Run(tc.caller, func(t *testing.T) {
			resp := request("/control-hub/apps", tc.caller)
			var apps []appv1alpha1.Application
			if resp.Code != http.StatusOK || json.Unmarshal(resp.Body.Bytes(), &apps) != nil {
				t.Fatalf("inventory failed with HTTP %d", resp.Code)
			}
			var owners []string
			for _, app := range apps {
				owners = append(owners, app.Spec.Owner)
			}
			slices.Sort(owners)
			if !slices.Equal(owners, tc.owners) {
				t.Fatalf("visible owners %v; want %v", owners, tc.owners)
			}
			if tc.caller != "operator" {
				resp := request("/apps/olares-app/env", tc.caller)
				var envs []sysv1alpha1.AppEnvVar
				if resp.Code != http.StatusOK || json.Unmarshal(resp.Body.Bytes(), &envs) != nil || len(envs) != 1 || envs[0].Value != tc.caller {
					t.Fatal("application environment read did not use the caller's tenant")
				}
			}
		})
	}
	var mismatched sysv1alpha1.AppEnv
	if err := h.ctrlClient.Get(t.Context(), client.ObjectKey{Namespace: "user-space-bob", Name: "olares-app-bob"}, &mismatched); err != nil {
		t.Fatal(err)
	}
	mismatched.AppOwner = "alice"
	if err := h.ctrlClient.Update(t.Context(), &mismatched); err != nil {
		t.Fatal(err)
	}
	if resp := request("/apps/olares-app/env", "bob"); resp.Code != http.StatusForbidden {
		t.Fatalf("environment record with a different owner returned %d", resp.Code)
	}

	if err := h.ctrlClient.Delete(t.Context(), &mismatched); err != nil {
		t.Fatal(err)
	}
	if resp := request("/apps/olares-app/env", "bob"); resp.Code != http.StatusOK || strings.TrimSpace(resp.Body.String()) != "[]" {
		t.Fatalf("missing environment record returned %d, %q", resp.Code, resp.Body.String())
	}
	mismatched = sysv1alpha1.AppEnv{
		ObjectMeta: metav1.ObjectMeta{Namespace: "user-space-bob", Name: "olares-app-bob"},
		AppName:    "olares-app", AppOwner: "bob",
	}
	if err := h.ctrlClient.Create(t.Context(), &mismatched); err != nil {
		t.Fatal(err)
	}
	if resp := request("/apps/olares-app/env", "bob"); resp.Code != http.StatusOK || strings.TrimSpace(resp.Body.String()) != "[]" {
		t.Fatalf("empty environment record returned %d, %q", resp.Code, resp.Body.String())
	}
	if resp := request("/control-hub/apps", "not-a-user"); resp.Code != http.StatusUnauthorized {
		t.Fatalf("unknown caller received Control Hub inventory: %d", resp.Code)
	}
	if resp := request("/control-hub/apps", "invalid-role"); resp.Code != http.StatusForbidden {
		t.Fatalf("caller without a recognized role received Control Hub inventory: %d", resp.Code)
	}

	if resp := request("/apps/env-store-error/env", "alice"); resp.Code != http.StatusInternalServerError || strings.Contains(resp.Body.String(), "store unavailable") {
		t.Fatalf("environment read failure returned %d or exposed the backend error", resp.Code)
	}

	for _, path := range []string{"/control-hub/apps", "/apps/olares-app/env"} {
		if resp := request(path, ""); resp.Code != http.StatusUnauthorized {
			t.Fatalf("anonymous request for %s returned %d", path, resp.Code)
		}
	}
}
