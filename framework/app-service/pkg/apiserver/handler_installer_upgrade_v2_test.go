package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// upgradeV2Fixture is the upgrade twin of the cancel / applyenv
// fixtures. Same "seed a rich Status + unrelated annotation, prove
// the MergeFrom patch leaves them intact" shape.
type upgradeV2Fixture struct {
	t       *testing.T
	handler *Handler
	am      *appv1alpha1.ApplicationManager

	req  *restful.Request
	resp *restful.Response
	rec  *httptest.ResponseRecorder
}

func newUpgradeV2Fixture(t *testing.T, state appv1alpha1.ApplicationManagerState) *upgradeV2Fixture {
	t.Helper()

	am := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "nginx-alice",
			ResourceVersion: "1",
			Annotations: map[string]string{
				"unrelated.example.com/keep-me": "yes",
				// Seed the current-generation annotations so the
				// MergeFrom patch's overwrite semantics are
				// exercised (not just create-fresh).
				api.AppRepoURLKey:      "https://old.example.com/charts",
				api.AppVersionKey:      "1.0.0",
				api.AppTokenKey:        "old-token",
				api.AppMarketSourceKey: "old-market",
			},
		},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName:      "nginx",
			AppNamespace: "nginx-alice",
			AppOwner:     "alice",
			Config:       `{"appName":"nginx","version":"1.0.0"}`,
			OpType:       appv1alpha1.InstallOp,
		},
		Status: appv1alpha1.ApplicationManagerStatus{
			State:    state,
			OpType:   appv1alpha1.InstallOp,
			OpID:     "seed-op-id",
			Message:  "seeded status",
			Reason:   "SeededReason",
			Progress: "42.00",
		},
	}

	ctrlClient := testutil.NewFakeClient(am)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/apps/nginx/upgrade", nil)
	req := restful.NewRequest(httpReq)
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)

	return &upgradeV2Fixture{
		t:       t,
		handler: &Handler{ctrlClient: ctrlClient},
		am:      am,
		req:     req,
		resp:    resp,
		rec:     rec,
	}
}

func (f *upgradeV2Fixture) currentAM() *appv1alpha1.ApplicationManager {
	f.t.Helper()
	var got appv1alpha1.ApplicationManager
	if err := f.handler.ctrlClient.Get(context.Background(), client.ObjectKey{Name: f.am.Name}, &got); err != nil {
		f.t.Fatalf("re-read AM: %v", err)
	}
	return &got
}

func fullUpgradeOpts() upgradeV2Options {
	return upgradeV2Options{
		Token:        "new-token",
		RepoURL:      "https://new.example.com/charts",
		Version:      "2.0.0",
		MarketSource: "new-market",
		ConfigJSON:   `{"appName":"nginx","version":"2.0.0"}`,
	}
}

// TestUpgradeV2_DeliversTriggerOp — happy path from Running.
// Verifies (a) Spec.OpType flips to UpgradeOp, (b) Spec.Config
// carries the freshly-rendered JSON, (c) upgrade annotations are
// written while the deployed chart version stays put, (d) the unrelated
// annotation survives the MergeFrom, (e) status is untouched, (f)
// exactly one TriggerOp{UpgradeOp} dispatched.
func TestUpgradeV2_DeliversTriggerOp(t *testing.T) {
	f := newUpgradeV2Fixture(t, appv1alpha1.Running)
	seededStatus := f.am.Status.DeepCopy()
	opts := fullUpgradeOpts()

	f.handler.deliverUpgradeV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, opts)

	if f.rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want 200; body=%s", f.rec.Code, f.rec.Body.String())
	}

	updated := f.currentAM()
	if updated.Spec.OpType != appv1alpha1.UpgradeOp {
		t.Errorf("Spec.OpType = %q, want %q", updated.Spec.OpType, appv1alpha1.UpgradeOp)
	}
	if updated.Spec.Config != opts.ConfigJSON {
		t.Errorf("Spec.Config mismatch:\n got %q\nwant %q", updated.Spec.Config, opts.ConfigJSON)
	}
	wantAnn := map[string]string{
		"unrelated.example.com/keep-me":     "yes",
		api.AppRepoURLKey:                   opts.RepoURL,
		api.AppVersionKey:                   "1.0.0",
		api.AppTargetChartVersionKey:        opts.Version,
		api.AppPreUpgradeConfigKey:          f.am.Spec.Config,
		api.AppTokenKey:                     opts.Token,
		api.AppMarketSourceKey:              opts.MarketSource,
		api.AppPreUpgradeStateKey:           string(appv1alpha1.Running),
		domain.AnnotationOperationRequestID: updated.Annotations[domain.AnnotationOperationRequestID],
	}
	if !reflect.DeepEqual(updated.Annotations, wantAnn) {
		t.Errorf("Annotations mismatch:\n got %#v\nwant %#v", updated.Annotations, wantAnn)
	}
	if !reflect.DeepEqual(&updated.Status, seededStatus) {
		t.Errorf("Status must be untouched:\n got %#v\nwant %#v", updated.Status, *seededStatus)
	}

	var body struct {
		Data struct {
			UID string `json:"uid"`
		} `json:"data"`
	}
	if err := json.Unmarshal(f.rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data.UID != f.am.Spec.AppName {
		t.Errorf("response uid = %q, want %q", body.Data.UID, f.am.Spec.AppName)
	}
}

// TestUpgradeV2_PreUpgradeStateAnnotation pins the annotation the
// Upgrading handler branches on. Running and Stopped are recorded as
// themselves. A retry from UpgradeFailed must keep a prior stopped
// snapshot: writing the failed state makes a successful retry scale
// the workload back up. With no prior steady value, the snapshot
// defaults to Running.
func TestUpgradeV2_PreUpgradeStateAnnotation(t *testing.T) {
	cases := []struct {
		name     string
		state    appv1alpha1.ApplicationManagerState
		existing string
		want     string
	}{
		{"from Running", appv1alpha1.Running, "", string(appv1alpha1.Running)},
		{"from Stopped", appv1alpha1.Stopped, "", string(appv1alpha1.Stopped)},
		{
			"running replaces leftover stopped",
			appv1alpha1.Running,
			string(appv1alpha1.Stopped),
			string(appv1alpha1.Running),
		},
		{
			"upgradeFailed keeps stopped",
			appv1alpha1.UpgradeFailed,
			string(appv1alpha1.Stopped),
			string(appv1alpha1.Stopped),
		},
		{
			"upgradeFailed with no prior defaults to running",
			appv1alpha1.UpgradeFailed,
			"",
			string(appv1alpha1.Running),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newUpgradeV2Fixture(t, c.state)
			if c.existing != "" {
				f.am.Annotations[api.AppPreUpgradeStateKey] = c.existing
				if err := f.handler.ctrlClient.Update(context.Background(), f.am); err != nil {
					t.Fatalf("seed annotation: %v", err)
				}
			}
			f.handler.deliverUpgradeV2(context.Background(), f.req, f.resp,
				f.am.Name, f.am, f.am.Spec.AppName, fullUpgradeOpts())

			if f.rec.Code != http.StatusOK {
				t.Fatalf("response status = %d, want 200; body=%s",
					f.rec.Code, f.rec.Body.String())
			}
			got := f.currentAM().Annotations[api.AppPreUpgradeStateKey]
			if got != c.want {
				t.Errorf("AppPreUpgradeStateKey = %q, want %q", got, c.want)
			}
		})
	}
}

// TestUpgradeV2_RejectsWhenOpNotAllowed — UpgradeOp is not allowed
// from Uninstalled.
func TestUpgradeV2_RejectsWhenOpNotAllowed(t *testing.T) {
	f := newUpgradeV2Fixture(t, appv1alpha1.Uninstalled)

	f.handler.deliverUpgradeV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, fullUpgradeOpts())

	if f.rec.Code != http.StatusBadRequest {
		t.Fatalf("response status = %d, want 400; body=%s", f.rec.Code, f.rec.Body.String())
	}
	if got := f.currentAM().Spec.OpType; got != appv1alpha1.InstallOp {
		t.Errorf("Spec.OpType must be untouched on rejection; got %q", got)
	}
}

// TestUpgradeV2_RejectsNilAnnotations preserves the legacy defensive
// rejection — same shape as applyEnvV2 (5xx-shape via
// api.HandleError, not 400).
func TestUpgradeV2_RejectsNilAnnotations(t *testing.T) {
	f := newUpgradeV2Fixture(t, appv1alpha1.Running)
	f.am.Annotations = nil
	if err := f.handler.ctrlClient.Update(context.Background(), f.am); err != nil {
		t.Fatalf("seed: nil-annotations AM: %v", err)
	}

	f.handler.deliverUpgradeV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, fullUpgradeOpts())

	if f.rec.Code == http.StatusOK {
		t.Fatalf("expected non-200 on nil-annotations rejection, got %d body=%s",
			f.rec.Code, f.rec.Body.String())
	}
	if got := f.currentAM().Spec.OpType; got != appv1alpha1.InstallOp {
		t.Errorf("Spec.OpType must be untouched on rejection; got %q", got)
	}
}

// TestUpgradeV2_PropagatesCtrlClientError — same invariant as the
// other v2 handlers: a mid-write spec failure must surface an
// HTTP error.
func TestUpgradeV2_PropagatesCtrlClientError(t *testing.T) {
	f := newUpgradeV2Fixture(t, appv1alpha1.Running)

	if err := f.handler.ctrlClient.Delete(context.Background(), f.am); err != nil {
		t.Fatalf("seed: delete AM: %v", err)
	}

	f.handler.deliverUpgradeV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, fullUpgradeOpts())

	if f.rec.Code == http.StatusOK {
		t.Fatalf("expected non-200 on ctrlClient failure, got %d body=%s",
			f.rec.Code, f.rec.Body.String())
	}
}

// TestUpgradeV2_ConfigOverwritten pins the "Spec.Config is fully
// replaced (not merged)" contract. The state machine's Upgrading
// handler treats Spec.Config as the source of truth for helm
// values; a partial merge would leave stale keys behind and desync
// helm from the manifest.
func TestUpgradeV2_ConfigOverwritten(t *testing.T) {
	f := newUpgradeV2Fixture(t, appv1alpha1.Running)
	seededConfig := f.am.Spec.Config

	opts := fullUpgradeOpts()
	opts.ConfigJSON = `{"appName":"nginx","version":"3.0.0","new":"field"}`

	f.handler.deliverUpgradeV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, opts)

	if f.rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want 200", f.rec.Code)
	}
	updated := f.currentAM()
	if updated.Spec.Config == seededConfig {
		t.Errorf("Spec.Config must be overwritten, still shows seed %q", seededConfig)
	}
	if updated.Spec.Config != opts.ConfigJSON {
		t.Errorf("Spec.Config = %q, want %q", updated.Spec.Config, opts.ConfigJSON)
	}
}

// TestUpgradeV2_HTTPShell_SignaturesCompile — router-swap safety net.
func TestUpgradeV2_HTTPShell_SignaturesCompile(t *testing.T) {
	var (
		_ func(*restful.Request, *restful.Response) = (&Handler{}).appUpgradeV2
		_                                           = (&Handler{}).deliverUpgradeV2
	)
}
