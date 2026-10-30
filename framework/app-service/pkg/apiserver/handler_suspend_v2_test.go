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

// suspendResumeV2Fixture is shared by the suspend and resume tests
// because both handlers land under the same handler_suspend.go /
// _v2.go pair and follow the same spec-patch shape.
//
// Spec.Config seeds an empty JSON object so
// compute.DeleteAllocationsForComputeTarget (invoked by
// deliverSuspendV2) unmarshals to zero-value appcfg.ApplicationConfig,
// which is neither V2 nor cluster-shared, so it resolves to an
// empty-target DeleteAllocationsForApp that no-ops against the
// fake client's empty allocation ConfigMap + empty HAMI binding
// list.
type suspendResumeV2Fixture struct {
	t       *testing.T
	handler *Handler
	am      *appv1alpha1.ApplicationManager

	req  *restful.Request
	resp *restful.Response
	rec  *httptest.ResponseRecorder
}

func newSuspendResumeV2Fixture(t *testing.T, state appv1alpha1.ApplicationManagerState) *suspendResumeV2Fixture {
	t.Helper()

	am := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "nginx-alice",
			ResourceVersion: "1",
			Annotations: map[string]string{
				"unrelated.example.com/keep-me": "yes",
			},
		},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName:      "nginx",
			AppNamespace: "nginx-alice",
			AppOwner:     "alice",
			Config:       "{}",
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
	httpReq := httptest.NewRequest(http.MethodPost, "/apps/nginx/suspend", nil)
	req := restful.NewRequest(httpReq)
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)

	return &suspendResumeV2Fixture{
		t:       t,
		handler: &Handler{ctrlClient: ctrlClient},
		am:      am,
		req:     req,
		resp:    resp,
		rec:     rec,
	}
}

func (f *suspendResumeV2Fixture) currentAM() *appv1alpha1.ApplicationManager {
	f.t.Helper()
	var got appv1alpha1.ApplicationManager
	if err := f.handler.ctrlClient.Get(context.Background(), client.ObjectKey{Name: f.am.Name}, &got); err != nil {
		f.t.Fatalf("re-read AM: %v", err)
	}
	return &got
}

// -----------------------------------------------------------------
// suspendV2
// -----------------------------------------------------------------

// TestSuspendV2_DeliversTriggerOp — happy path from Running.
func TestSuspendV2_DeliversTriggerOp(t *testing.T) {
	f := newSuspendResumeV2Fixture(t, appv1alpha1.Running)
	seededStatus := f.am.Status.DeepCopy()

	f.handler.deliverSuspendV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, true /* stopAll */)

	if f.rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want 200; body=%s", f.rec.Code, f.rec.Body.String())
	}

	updated := f.currentAM()
	if updated.Spec.OpType != appv1alpha1.StopOp {
		t.Errorf("Spec.OpType = %q, want %q", updated.Spec.OpType, appv1alpha1.StopOp)
	}
	wantAnn := map[string]string{
		"unrelated.example.com/keep-me":     "yes",
		api.AppStopAllKey:                   "true",
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

// TestSuspendV2_StopAllAnnotationEncoding pins the "%t" formatter
// (same shape as the uninstall bool-annotation table test).
func TestSuspendV2_StopAllAnnotationEncoding(t *testing.T) {
	cases := []struct {
		name    string
		stopAll bool
		want    string
	}{
		{"stopAll true", true, "true"},
		{"stopAll false", false, "false"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newSuspendResumeV2Fixture(t, appv1alpha1.Running)

			f.handler.deliverSuspendV2(context.Background(), f.req, f.resp,
				f.am.Name, f.am, f.am.Spec.AppName, c.stopAll)

			if f.rec.Code != http.StatusOK {
				t.Fatalf("response status = %d, want 200", f.rec.Code)
			}
			if got := f.currentAM().Annotations[api.AppStopAllKey]; got != c.want {
				t.Errorf("AppStopAllKey = %q, want %q", got, c.want)
			}
		})
	}
}

// TestSuspendV2_RejectsWhenOpNotAllowed — StopOp is only allowed
// from Running / StopFailed. Uninstalled rejects it.
func TestSuspendV2_RejectsWhenOpNotAllowed(t *testing.T) {
	f := newSuspendResumeV2Fixture(t, appv1alpha1.Uninstalled)

	f.handler.deliverSuspendV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, false)

	if f.rec.Code != http.StatusBadRequest {
		t.Fatalf("response status = %d, want 400; body=%s", f.rec.Code, f.rec.Body.String())
	}
	if got := f.currentAM().Spec.OpType; got != appv1alpha1.InstallOp {
		t.Errorf("Spec.OpType must be untouched on rejection; got %q", got)
	}
}

// TestSuspendV2_RejectsMalformedConfig pins the "AM.Spec.Config must
// unmarshal" invariant. A hand-corrupted Spec.Config (invalid JSON)
// short-circuits with an error before a new request id is stamped.
// The Runner must not start Stop against an AM whose compute
// couldn't be resolved.
func TestSuspendV2_RejectsMalformedConfig(t *testing.T) {
	f := newSuspendResumeV2Fixture(t, appv1alpha1.Running)
	f.am.Spec.Config = "{not json"
	if err := f.handler.ctrlClient.Update(context.Background(), f.am); err != nil {
		t.Fatalf("seed: corrupt Spec.Config: %v", err)
	}

	f.handler.deliverSuspendV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, false)

	if f.rec.Code == http.StatusOK {
		t.Fatalf("expected non-200 on malformed Spec.Config, got %d body=%s",
			f.rec.Code, f.rec.Body.String())
	}
}

// TestSuspendV2_PropagatesCtrlClientError — spec write fails → no
// TriggerOp. Same invariant as cancel/uninstall/applyenv.
func TestSuspendV2_PropagatesCtrlClientError(t *testing.T) {
	f := newSuspendResumeV2Fixture(t, appv1alpha1.Running)

	if err := f.handler.ctrlClient.Delete(context.Background(), f.am); err != nil {
		t.Fatalf("seed: delete AM: %v", err)
	}

	f.handler.deliverSuspendV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, false)

	if f.rec.Code == http.StatusOK {
		t.Fatalf("expected non-200 on ctrlClient failure, got %d body=%s",
			f.rec.Code, f.rec.Body.String())
	}
}

// TestSuspendV2_HTTPShell_SignaturesCompile — router-swap safety net.
func TestSuspendV2_HTTPShell_SignaturesCompile(t *testing.T) {
	var (
		_ func(*restful.Request, *restful.Response) = (&Handler{}).suspendV2
		_                                           = (&Handler{}).deliverSuspendV2
	)
}

// -----------------------------------------------------------------
// resumeV2
// -----------------------------------------------------------------

// TestResumeV2_DeliversTriggerOp — happy path from Stopped.
// includeSharedServer=true stamps AppResumeAllKey="true".
func TestResumeV2_DeliversTriggerOp(t *testing.T) {
	f := newSuspendResumeV2Fixture(t, appv1alpha1.Stopped)
	seededStatus := f.am.Status.DeepCopy()

	if err := f.handler.deliverResumeV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, true /* includeSharedServer */); err != nil {
		t.Fatalf("deliverResumeV2 returned unexpected error: %v", err)
	}

	if f.rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want 200; body=%s", f.rec.Code, f.rec.Body.String())
	}

	updated := f.currentAM()
	if updated.Spec.OpType != appv1alpha1.ResumeOp {
		t.Errorf("Spec.OpType = %q, want %q", updated.Spec.OpType, appv1alpha1.ResumeOp)
	}
	wantAnn := map[string]string{
		"unrelated.example.com/keep-me":     "yes",
		api.AppResumeAllKey:                 "true",
		domain.AnnotationOperationRequestID: updated.Annotations[domain.AnnotationOperationRequestID],
	}
	if !reflect.DeepEqual(updated.Annotations, wantAnn) {
		t.Errorf("Annotations mismatch:\n got %#v\nwant %#v", updated.Annotations, wantAnn)
	}
	if !reflect.DeepEqual(&updated.Status, seededStatus) {
		t.Errorf("Status must be untouched:\n got %#v\nwant %#v", updated.Status, *seededStatus)
	}
}

// TestResumeV2_ResumeAllAnnotationEncoding — the AppResumeAllKey
// tracks the includeSharedServer bool verbatim ("true"/"false").
func TestResumeV2_ResumeAllAnnotationEncoding(t *testing.T) {
	cases := []struct {
		name                string
		includeSharedServer bool
		want                string
	}{
		{"includeSharedServer true", true, "true"},
		{"includeSharedServer false", false, "false"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newSuspendResumeV2Fixture(t, appv1alpha1.Stopped)

			if err := f.handler.deliverResumeV2(context.Background(), f.req, f.resp,
				f.am.Name, f.am, f.am.Spec.AppName, c.includeSharedServer); err != nil {
				t.Fatalf("deliverResumeV2: %v", err)
			}
			if got := f.currentAM().Annotations[api.AppResumeAllKey]; got != c.want {
				t.Errorf("AppResumeAllKey = %q, want %q", got, c.want)
			}
		})
	}
}

// TestResumeV2_RejectsWhenOpNotAllowed — ResumeOp is only allowed
// from Stopped / ResumeFailed. Running rejects it.
func TestResumeV2_RejectsWhenOpNotAllowed(t *testing.T) {
	f := newSuspendResumeV2Fixture(t, appv1alpha1.Running)

	err := f.handler.deliverResumeV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, false)
	if err == nil {
		t.Fatal("deliverResumeV2 must return an error on state rejection")
	}
	if f.rec.Code != http.StatusBadRequest {
		t.Fatalf("response status = %d, want 400; body=%s", f.rec.Code, f.rec.Body.String())
	}
	if got := f.currentAM().Spec.OpType; got != appv1alpha1.InstallOp {
		t.Errorf("Spec.OpType must be untouched on rejection; got %q", got)
	}
}

// TestResumeV2_PropagatesCtrlClientError — spec write fails → no
// TriggerOp; error is returned to the outer shell so the deferred
// compute cleanup can undo ApplyBindingSelection.
func TestResumeV2_PropagatesCtrlClientError(t *testing.T) {
	f := newSuspendResumeV2Fixture(t, appv1alpha1.Stopped)

	if err := f.handler.ctrlClient.Delete(context.Background(), f.am); err != nil {
		t.Fatalf("seed: delete AM: %v", err)
	}

	err := f.handler.deliverResumeV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, false)
	if err == nil {
		t.Fatal("deliverResumeV2 must surface the Patch error to the caller")
	}
	if f.rec.Code == http.StatusOK {
		t.Fatalf("expected non-200 on ctrlClient failure, got %d body=%s",
			f.rec.Code, f.rec.Body.String())
	}
}

func TestSuspendV2_RejectsPendingApplyEnv(t *testing.T) {
	f := newSuspendResumeV2Fixture(t, appv1alpha1.Running)
	f.am.Spec.OpType = appv1alpha1.ApplyEnvOp
	f.am.Annotations[domain.AnnotationOperationRequestID] = "pending-apply"
	if err := f.handler.ctrlClient.Update(context.Background(), f.am); err != nil {
		t.Fatal(err)
	}

	f.handler.deliverSuspendV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, false)

	if f.rec.Code != http.StatusBadRequest {
		t.Fatalf("response status = %d, want 400; body=%s", f.rec.Code, f.rec.Body.String())
	}
	got := f.currentAM()
	if got.Spec.OpType != appv1alpha1.ApplyEnvOp {
		t.Fatalf("Spec.OpType = %q, want ApplyEnvOp", got.Spec.OpType)
	}
	if got.Annotations[domain.AnnotationOperationRequestID] != "pending-apply" {
		t.Fatalf("request id = %q, want pending-apply", got.Annotations[domain.AnnotationOperationRequestID])
	}
}

// TestResumeV2_HTTPShell_SignaturesCompile — router-swap safety net.
func TestResumeV2_HTTPShell_SignaturesCompile(t *testing.T) {
	var (
		_ func(*restful.Request, *restful.Response) = (&Handler{}).resumeV2
		_                                           = (&Handler{}).deliverResumeV2
	)
}
