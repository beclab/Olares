package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	lifecycleapi "github.com/beclab/Olares/framework/app-service/pkg/applifecycle/apiserver"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// installV2Fixture is the install twin of the upgrade / cancel /
// applyenv fixtures. Where those handlers always mutate an existing
// AM, install has TWO paths (create-fresh vs. patch-reinstallable),
// so the fixture supports both by exposing the pre-seed hook.
type installV2Fixture struct {
	t          *testing.T
	handler    *Handler
	dispatcher *lifecycleapi.FakeDispatcher

	req  *restful.Request
	resp *restful.Response
	rec  *httptest.ResponseRecorder
}

func newInstallV2Fixture(t *testing.T, seed ...client.Object) *installV2Fixture {
	t.Helper()

	ctrlClient := testutil.NewFakeClient(seed...)
	dispatcher := &lifecycleapi.FakeDispatcher{}

	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/apps/nginx/install", nil)
	req := restful.NewRequest(httpReq)
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)

	return &installV2Fixture{
		t:          t,
		handler:    &Handler{ctrlClient: ctrlClient, opDispatcher: dispatcher},
		dispatcher: dispatcher,
		req:        req,
		resp:       resp,
		rec:        rec,
	}
}

func (f *installV2Fixture) currentAM(name string) *appv1alpha1.ApplicationManager {
	f.t.Helper()
	var got appv1alpha1.ApplicationManager
	if err := f.handler.ctrlClient.Get(context.Background(), client.ObjectKey{Name: name}, &got); err != nil {
		f.t.Fatalf("re-read AM %q: %v", name, err)
	}
	return &got
}

func baseInstallAppCfg() *appcfg.ApplicationConfig {
	return &appcfg.ApplicationConfig{
		AppName:   "nginx",
		Namespace: "nginx-alice",
		OwnerName: "alice",
		Version:   "1.2.3",
		Title:     "Nginx",
		Type:      "app",
	}
}

func baseInstallReq() *api.InstallRequest {
	return &api.InstallRequest{
		RepoURL: "https://example.com/charts",
		Source:  api.Market,
	}
}

func baseInstallOpts() installV2Options {
	return installV2Options{
		Token:        "sa-token",
		MarketSource: "market-a",
		ChartOwner:   "alice",
		ExtraLabels:  nil,
		App:          "nginx",
		RawAppName:   "nginx",
	}
}

// TestInstallV2_CreatesFreshAM covers the first-time install path:
// no AM at the target name → v2 CREATES one carrying all
// spec/labels/annotations legacy would have written, then delivers
// exactly one TriggerOp{InstallOp}. Also pins that Status stays at
// its zero value (the Runner will drive the "" → Pending bootstrap
// on its first Reconcile).
func TestInstallV2_CreatesFreshAM(t *testing.T) {
	f := newInstallV2Fixture(t)
	appCfg := baseInstallAppCfg()
	insReq := baseInstallReq()
	insReq.Images = []api.Image{{Name: "docker.io/library/nginx:1.25"}}
	opts := baseInstallOpts()

	f.handler.deliverInstallV2(context.Background(), f.req, f.resp,
		"nginx-alice-nginx", appCfg, insReq, opts)

	if f.rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want 200; body=%s", f.rec.Code, f.rec.Body.String())
	}

	got := f.currentAM("nginx-alice-nginx")

	if got.Spec.AppName != "nginx" ||
		got.Spec.RawAppName != "nginx" ||
		got.Spec.AppNamespace != "nginx-alice" ||
		got.Spec.AppOwner != "alice" ||
		got.Spec.Source != api.Market.String() ||
		got.Spec.Type != appv1alpha1.Type("app") ||
		got.Spec.OpType != appv1alpha1.InstallOp {
		t.Errorf("Spec mismatch: %+v", got.Spec)
	}

	var cfg appcfg.ApplicationConfig
	if err := json.Unmarshal([]byte(got.Spec.Config), &cfg); err != nil {
		t.Fatalf("Spec.Config not valid JSON: %v", err)
	}
	if cfg.AppName != "nginx" || cfg.Version != "1.2.3" || cfg.ChartOwner != "alice" {
		t.Errorf("Spec.Config decoded mismatch: %+v", cfg)
	}
	// A market install with no RawAppName has an empty ClonedFrom.
	if cfg.ClonedFrom != "" {
		t.Errorf("ClonedFrom = %q, want empty on fresh non-clone install", cfg.ClonedFrom)
	}

	wantAnn := map[string]string{
		api.AppTokenKey:                     "sa-token",
		api.AppRepoURLKey:                   "https://example.com/charts",
		api.AppVersionKey:                   "1.2.3",
		api.AppMarketSourceKey:              "market-a",
		api.AppInstallSourceKey:             "app-service",
		constants.ApplicationTitleLabel:     "Nginx",
		constants.ApplicationImageLabel:     `[{"name":"docker.io/library/nginx:1.25","size":0}]`,
		domain.AnnotationOperationRequestID: got.Annotations[domain.AnnotationOperationRequestID],
	}
	if !reflect.DeepEqual(got.Annotations, wantAnn) {
		t.Errorf("Annotations mismatch:\n got %#v\nwant %#v", got.Annotations, wantAnn)
	}

	// Labels: even on non-clone installs the two "always-stamped"
	// keys must be present so downstream label selectors work.
	if v, ok := got.Labels[constants.AppClonedFromKey]; !ok || v != "" {
		t.Errorf("Labels[%q] = %q, want empty-but-present", constants.AppClonedFromKey, v)
	}
	if v, ok := got.Labels[constants.AppChartOwnerKey]; !ok || v != "alice" {
		t.Errorf("Labels[%q] = %q, want %q", constants.AppChartOwnerKey, v, "alice")
	}

	// Status must be the zero value — Runner drives the bootstrap.
	if got.Status.State != appv1alpha1.ApplicationManagerState("") {
		t.Errorf("Status.State = %q, want zero value (Runner owns bootstrap)", got.Status.State)
	}
	if got.Status.OpID != "" {
		t.Errorf("Status.OpID = %q, want empty (Runner has not consumed the request)", got.Status.OpID)
	}

	calls := f.dispatcher.Calls()
	want := []lifecycleapi.DispatchCall{
		{Name: "nginx-alice-nginx", Action: domain.TriggerOp{Op: appv1alpha1.InstallOp}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("dispatcher calls mismatch:\n got %#v\nwant %#v", calls, want)
	}

	var body struct {
		Data struct {
			UID  string `json:"uid"`
			OpID string `json:"opID"`
		} `json:"data"`
	}
	if err := json.Unmarshal(f.rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data.UID != "nginx" {
		t.Errorf("response uid = %q, want %q", body.Data.UID, "nginx")
	}
	// The response reports the durable request ID as opID. Market
	// rejects an accepted install whose data carries no opID, so the
	// value must be present before the Runner has written any Status
	// and must match what the Runner will later copy into Status.OpID.
	if body.Data.OpID != got.Annotations[domain.AnnotationOperationRequestID] {
		t.Errorf("response opID = %q, want durable request ID %q",
			body.Data.OpID, got.Annotations[domain.AnnotationOperationRequestID])
	}
	if body.Data.OpID == "" {
		t.Error("response opID is empty; market rejects the operation")
	}
}

// TestInstallV2_PatchesReinstallableAM covers the reinstall path:
// an existing AM in a terminal-reinstallable state (Uninstalled) is
// re-targeted with a fresh spec, labels, annotations. Critically:
// Status must NOT be touched (it carries the prior lifecycle's
// terminal OpRecords the UI still surfaces).
func TestInstallV2_PatchesReinstallableAM(t *testing.T) {
	const name = "nginx-alice-nginx"
	prior := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			ResourceVersion: "1",
			Labels: map[string]string{
				"unrelated.example.com/keep-me": "yes",
			},
			Annotations: map[string]string{
				"unrelated.example.com/keep-me": "yes",
				api.AppTokenKey:                 "old-token",
			},
		},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName:      "nginx",
			AppNamespace: "nginx-alice",
			AppOwner:     "alice",
			OpType:       appv1alpha1.UninstallOp,
			Source:       api.Market.String(),
		},
		Status: appv1alpha1.ApplicationManagerStatus{
			State:   appv1alpha1.Uninstalled,
			OpType:  appv1alpha1.UninstallOp,
			OpID:    "prior-op",
			Message: "seeded",
			OpRecords: []appv1alpha1.OpRecord{
				{OpType: appv1alpha1.UninstallOp, OpID: "prior-op"},
			},
		},
	}
	f := newInstallV2Fixture(t, prior)
	priorStatus := prior.Status.DeepCopy()

	f.handler.deliverInstallV2(context.Background(), f.req, f.resp,
		name, baseInstallAppCfg(), baseInstallReq(), baseInstallOpts())

	if f.rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want 200; body=%s", f.rec.Code, f.rec.Body.String())
	}

	got := f.currentAM(name)

	if got.Spec.OpType != appv1alpha1.InstallOp {
		t.Errorf("Spec.OpType = %q, want InstallOp", got.Spec.OpType)
	}
	if !reflect.DeepEqual(&got.Status, priorStatus) {
		t.Errorf("Status must be untouched on reinstall Patch:\n got %#v\nwant %#v",
			got.Status, *priorStatus)
	}

	// Unrelated label + annotation from the prior generation must
	// survive the MergeFrom patch.
	if v := got.Labels["unrelated.example.com/keep-me"]; v != "yes" {
		t.Errorf("unrelated label lost: got %q", v)
	}
	if v := got.Annotations["unrelated.example.com/keep-me"]; v != "yes" {
		t.Errorf("unrelated annotation lost: got %q", v)
	}
	// AppTokenKey annotation was overwritten (not appended) with the
	// new value.
	if v := got.Annotations[api.AppTokenKey]; v != "sa-token" {
		t.Errorf("Annotations[%q] = %q, want %q", api.AppTokenKey, v, "sa-token")
	}
	requireDurableRequestID(t, got.Annotations)

	// Exactly one TriggerOp{InstallOp} — same trigger for both the
	// create-fresh and patch-reinstallable paths.
	calls := f.dispatcher.Calls()
	want := []lifecycleapi.DispatchCall{
		{Name: name, Action: domain.TriggerOp{Op: appv1alpha1.InstallOp}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("dispatcher calls mismatch:\n got %#v\nwant %#v", calls, want)
	}
}

// TestInstallV2_RejectsWhenOpNotAllowed proves the reinstall gate.
// An AM in Running is NOT a terminal-reinstallable state; installV2
// must reject with 400 and NOT touch the AM or fire the dispatcher.
func TestInstallV2_RejectsWhenOpNotAllowed(t *testing.T) {
	const name = "nginx-alice-nginx"
	prior := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name, ResourceVersion: "1"},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName: "nginx", AppOwner: "alice", OpType: appv1alpha1.InstallOp,
		},
		Status: appv1alpha1.ApplicationManagerStatus{
			State:  appv1alpha1.Running,
			OpType: appv1alpha1.InstallOp,
			OpID:   "running-op",
		},
	}
	f := newInstallV2Fixture(t, prior)
	priorSpec := prior.Spec.DeepCopy()

	f.handler.deliverInstallV2(context.Background(), f.req, f.resp,
		name, baseInstallAppCfg(), baseInstallReq(), baseInstallOpts())

	if f.rec.Code != http.StatusBadRequest {
		t.Fatalf("response status = %d, want 400; body=%s", f.rec.Code, f.rec.Body.String())
	}
	got := f.currentAM(name)
	if !reflect.DeepEqual(&got.Spec, priorSpec) {
		t.Errorf("Spec must be untouched on rejection")
	}
	if got := f.dispatcher.Calls(); len(got) != 0 {
		t.Errorf("dispatcher must not be triggered on rejection; got %#v", got)
	}
}

// TestInstallV2_V3SharedLabels — extraLabels from the V3 helper
// (api-version=v3, shared=true) must be stamped alongside the two
// always-present keys.
func TestInstallV2_V3SharedLabels(t *testing.T) {
	f := newInstallV2Fixture(t)
	appCfg := baseInstallAppCfg()
	appCfg.Shared = true
	opts := baseInstallOpts()
	opts.ExtraLabels = map[string]string{
		constants.AppApiVersionLabel: constants.AppVersionV3,
		constants.AppSharedLabel:     "true",
	}

	f.handler.deliverInstallV2(context.Background(), f.req, f.resp,
		"nginx-alice-nginx", appCfg, baseInstallReq(), opts)

	if f.rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want 200; body=%s", f.rec.Code, f.rec.Body.String())
	}
	got := f.currentAM("nginx-alice-nginx")
	wantLabels := map[string]string{
		constants.AppApiVersionLabel: constants.AppVersionV3,
		constants.AppSharedLabel:     "true",
		constants.AppClonedFromKey:   "",
		constants.AppChartOwnerKey:   "alice",
	}
	if !reflect.DeepEqual(got.Labels, wantLabels) {
		t.Errorf("Labels mismatch:\n got %#v\nwant %#v", got.Labels, wantLabels)
	}
}

// TestInstallV2_CloneStampsClonedFrom — non-empty RawAppName +
// TemplateClone should stamp the "template" ClonedFrom marker; a
// clone-of-app stamps "app". Both go through the label AND the
// embedded Spec.Config so downstream (deployment / Application
// label / market analytics) can see the origin.
func TestInstallV2_CloneStampsClonedFrom(t *testing.T) {
	cases := []struct {
		name          string
		insReq        *api.InstallRequest
		wantClonedVal string
	}{
		{
			name: "template clone",
			insReq: &api.InstallRequest{
				RepoURL: "u", Source: api.Market,
				RawAppName: "nginx", TemplateClone: true,
			},
			wantClonedVal: constants.AppClonedFromTemplate,
		},
		{
			name: "app clone",
			insReq: &api.InstallRequest{
				RepoURL: "u", Source: api.Market,
				RawAppName: "nginx", TemplateClone: false,
			},
			wantClonedVal: constants.AppClonedFromApp,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newInstallV2Fixture(t)
			opts := baseInstallOpts()
			opts.App = "nginx-clone"
			opts.RawAppName = "nginx"

			f.handler.deliverInstallV2(context.Background(), f.req, f.resp,
				"nginx-clone-alice-nginx-clone", baseInstallAppCfg(), c.insReq, opts)

			if f.rec.Code != http.StatusOK {
				t.Fatalf("status = %d body=%s", f.rec.Code, f.rec.Body.String())
			}
			got := f.currentAM("nginx-clone-alice-nginx-clone")
			if v := got.Labels[constants.AppClonedFromKey]; v != c.wantClonedVal {
				t.Errorf("Labels[%q] = %q, want %q", constants.AppClonedFromKey, v, c.wantClonedVal)
			}
			var cfg appcfg.ApplicationConfig
			if err := json.Unmarshal([]byte(got.Spec.Config), &cfg); err != nil {
				t.Fatalf("decode Spec.Config: %v", err)
			}
			if cfg.ClonedFrom != c.wantClonedVal {
				t.Errorf("Spec.Config.ClonedFrom = %q, want %q", cfg.ClonedFrom, c.wantClonedVal)
			}
		})
	}
}

// TestInstallV2_HTTPShell_SignaturesCompile — router-swap safety
// net. Same shape as every other v2 handler; guards against a
// silent signature change breaking the eventual router wiring.
func TestInstallV2_HTTPShell_SignaturesCompile(t *testing.T) {
	var (
		_ func(*restful.Request, *restful.Response) = (&Handler{}).installV2
		_                                           = (&Handler{}).deliverInstallV2
	)
}
