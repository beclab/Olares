package apiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	appfake "github.com/beclab/api/pkg/generated/clientset/versioned/fake"
	"github.com/beclab/api/pkg/generated/informers/externalversions"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func amObj(name, owner, appName string, state appv1alpha1.ApplicationManagerState, v3 bool) *appv1alpha1.ApplicationManager {
	labels := map[string]string{}
	if v3 {
		labels[appv1alpha1.AppApiVersionLabel] = appv1alpha1.AppVersionV3
	}
	return &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName:      appName,
			AppNamespace: appName + "-" + owner,
			AppOwner:     owner,
			Source:       "market",
			Type:         appv1alpha1.App,
		},
		Status: appv1alpha1.ApplicationManagerStatus{State: state},
	}
}

func newHandlerWithAMs(t *testing.T, ams ...*appv1alpha1.ApplicationManager) *Handler {
	t.Helper()
	objs := make([]runtime.Object, 0, len(ams))
	for _, am := range ams {
		objs = append(objs, am)
	}
	client := appfake.NewSimpleClientset(objs...)
	factory := externalversions.NewSharedInformerFactory(client, 0)
	lister := factory.App().V1alpha1().ApplicationManagers().Lister()

	stop := make(chan struct{})
	t.Cleanup(func() { close(stop) })
	factory.Start(stop)
	synced := factory.WaitForCacheSync(stop)
	for typ, ok := range synced {
		if !ok {
			t.Fatalf("informer cache for %v failed to sync", typ)
		}
	}
	return &Handler{appmgrLister: lister}
}

type appsStatusResult struct {
	Result []struct {
		Name string `json:"name"`
	} `json:"result"`
}

func callAppsStatus(t *testing.T, h *Handler, owner, query string) appsStatusResult {
	t.Helper()
	httpReq := httptest.NewRequest(http.MethodGet, "/apps?"+query, nil)
	req := restful.NewRequest(httpReq)
	req.SetAttribute(constants.UserContextAttribute, owner)
	rec := httptest.NewRecorder()
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)

	h.appsStatus(req, resp)

	if rec.Code != http.StatusOK {
		t.Fatalf("appsStatus status=%d body=%s", rec.Code, rec.Body.String())
	}
	var out appsStatusResult
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rec.Body.String())
	}
	return out
}

func names(r appsStatusResult) map[string]bool {
	m := map[string]bool{}
	for _, a := range r.Result {
		m[a.Name] = true
	}
	return m
}

func TestAppsStatusFiltersByOwnerAndV3(t *testing.T) {
	h := newHandlerWithAMs(t,
		amObj("a1", "alice", "nginx", appv1alpha1.Running, false),
		amObj("a2", "bob", "mysql", appv1alpha1.Running, false),
		amObj("a3", "bob", "shared", appv1alpha1.Running, true), // v3, visible to all
	)
	// wait until lister sees all three before asserting (List is eventually consistent).
	deadline := time.Now().Add(3 * time.Second)
	for {
		got := names(callAppsStatus(t, h, "alice", ""))
		if got["nginx"] && got["shared"] || time.Now().After(deadline) {
			if got["mysql"] {
				t.Errorf("alice should not see bob's non-v3 app, got %v", got)
			}
			if !got["nginx"] {
				t.Errorf("alice should see her own app, got %v", got)
			}
			if !got["shared"] {
				t.Errorf("alice should see the v3 app, got %v", got)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestAppsStatusFiltersByState(t *testing.T) {
	h := newHandlerWithAMs(t,
		amObj("a1", "alice", "nginx", appv1alpha1.Running, false),
		amObj("a2", "alice", "redis", appv1alpha1.Stopped, false),
	)
	got := names(callAppsStatus(t, h, "alice", "state="+appv1alpha1.Running.String()))
	if !got["nginx"] {
		t.Errorf("running app nginx should be present, got %v", got)
	}
	if got["redis"] {
		t.Errorf("stopped app redis should be filtered out, got %v", got)
	}
}

func TestAppsStatusFiltersBySysApp(t *testing.T) {
	t.Setenv("SYS_APPS", "settings")
	h := newHandlerWithAMs(t,
		amObj("a1", "alice", "settings", appv1alpha1.Running, false),
		amObj("a2", "alice", "nginx", appv1alpha1.Running, false),
	)
	got := names(callAppsStatus(t, h, "alice", "issysapp=true"))
	if !got["settings"] {
		t.Errorf("system app settings should be present, got %v", got)
	}
	if got["nginx"] {
		t.Errorf("non-system app nginx should be filtered out, got %v", got)
	}
}
