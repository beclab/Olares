package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	lifecycleapi "github.com/beclab/Olares/framework/app-service/pkg/applifecycle/apiserver"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// applyEnvV2Fixture mirrors uninstallV2Fixture but seeds an AM with
// (a) a non-nil Annotations map because deliverApplyEnvV2 rejects
// nil-annotations on purpose (legacy-preserved defence), and (b) a
// rich Status so an accidental clobber surfaces.
type applyEnvV2Fixture struct {
	t          *testing.T
	handler    *Handler
	dispatcher *lifecycleapi.FakeDispatcher
	am         *appv1alpha1.ApplicationManager

	req  *restful.Request
	resp *restful.Response
	rec  *httptest.ResponseRecorder
}

// newApplyEnvV2Fixture defaults to Annotations={"unrelated": "keep"}
// so the happy path can prove the MergeFrom patch preserves what it
// didn't touch. Tests that want nil annotations call
// f.am.Annotations = nil AND persist that back via Update before
// exercising the handler.
func newApplyEnvV2Fixture(t *testing.T, state appv1alpha1.ApplicationManagerState) *applyEnvV2Fixture {
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
	dispatcher := &lifecycleapi.FakeDispatcher{}

	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/apps/nginx/applyenv", nil)
	req := restful.NewRequest(httpReq)
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)

	return &applyEnvV2Fixture{
		t:          t,
		handler:    &Handler{ctrlClient: ctrlClient, opDispatcher: dispatcher},
		dispatcher: dispatcher,
		am:         am,
		req:        req,
		resp:       resp,
		rec:        rec,
	}
}

func (f *applyEnvV2Fixture) currentAM() *appv1alpha1.ApplicationManager {
	f.t.Helper()
	var got appv1alpha1.ApplicationManager
	if err := f.handler.ctrlClient.Get(context.Background(), client.ObjectKey{Name: f.am.Name}, &got); err != nil {
		f.t.Fatalf("re-read AM: %v", err)
	}
	return &got
}

// TestApplyEnvV2_DeliversTriggerOp is the happy path from Running:
// spec flips to ApplyEnvOp, token annotation lands, unrelated
// annotation survives, status untouched, exactly one TriggerOp
// dispatched, response is api.Response{Code: 200} (no UID — legacy
// shape).
func TestApplyEnvV2_DeliversTriggerOp(t *testing.T) {
	f := newApplyEnvV2Fixture(t, appv1alpha1.Running)
	seededStatus := f.am.Status.DeepCopy()

	f.handler.deliverApplyEnvV2(context.Background(), f.req, f.resp, f.am.Name, f.am, "sa-token-123")

	if f.rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want 200; body=%s", f.rec.Code, f.rec.Body.String())
	}

	updated := f.currentAM()
	if updated.Spec.OpType != appv1alpha1.ApplyEnvOp {
		t.Errorf("Spec.OpType = %q, want %q", updated.Spec.OpType, appv1alpha1.ApplyEnvOp)
	}
	wantAnn := map[string]string{
		"unrelated.example.com/keep-me":     "yes",
		api.AppTokenKey:                     "sa-token-123",
		domain.AnnotationOperationRequestID: updated.Annotations[domain.AnnotationOperationRequestID],
	}
	if !reflect.DeepEqual(updated.Annotations, wantAnn) {
		t.Errorf("Annotations mismatch:\n got %#v\nwant %#v", updated.Annotations, wantAnn)
	}
	if !reflect.DeepEqual(&updated.Status, seededStatus) {
		t.Errorf("Status must be untouched:\n got %#v\nwant %#v", updated.Status, *seededStatus)
	}

	calls := f.dispatcher.Calls()
	want := []lifecycleapi.DispatchCall{
		{Name: f.am.Name, Action: domain.TriggerOp{Op: appv1alpha1.ApplyEnvOp}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("dispatcher calls mismatch:\n got %#v\nwant %#v", calls, want)
	}

	// Legacy applyenv responds with the plain api.Response envelope
	// (no InstallationResponseData) — preserve that so existing
	// callers keep parsing.
	var body api.Response
	if err := json.Unmarshal(f.rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != 200 {
		t.Errorf("response body Code = %d, want 200", body.Code)
	}
}

// TestApplyEnvV2_RejectsWhenOpNotAllowed pins the state-guard case.
// ApplyEnvOp is not allowed from Uninstalled (see state_transition.go),
// so the handler must respond 400 without side effects.
func TestApplyEnvV2_RejectsWhenOpNotAllowed(t *testing.T) {
	f := newApplyEnvV2Fixture(t, appv1alpha1.Uninstalled)

	f.handler.deliverApplyEnvV2(context.Background(), f.req, f.resp, f.am.Name, f.am, "sa-token-123")

	if f.rec.Code != http.StatusBadRequest {
		t.Fatalf("response status = %d, want 400; body=%s", f.rec.Code, f.rec.Body.String())
	}
	if got := f.currentAM().Spec.OpType; got != appv1alpha1.InstallOp {
		t.Errorf("Spec.OpType must be untouched on rejection; got %q", got)
	}
	if got := f.dispatcher.Calls(); len(got) != 0 {
		t.Errorf("dispatcher must not be triggered on rejection; got %#v", got)
	}
}

// TestApplyEnvV2_RejectsNilAnnotations preserves the legacy defensive
// rejection: an AM with no annotation map at all is not a valid
// ApplyEnv target. The legacy handler returns via api.HandleError
// (a 5xx-shaped response), so v2 must too — otherwise clients that
// key on the specific error code break.
func TestApplyEnvV2_RejectsNilAnnotations(t *testing.T) {
	f := newApplyEnvV2Fixture(t, appv1alpha1.Running)
	// Persist the nil-annotations AM to the fake store so the
	// handler observes the state it's asserting on.
	f.am.Annotations = nil
	if err := f.handler.ctrlClient.Update(context.Background(), f.am); err != nil {
		t.Fatalf("seed: nil-annotations AM: %v", err)
	}

	f.handler.deliverApplyEnvV2(context.Background(), f.req, f.resp, f.am.Name, f.am, "sa-token-123")

	if f.rec.Code == http.StatusOK {
		t.Fatalf("expected non-200 on nil-annotations rejection, got %d body=%s",
			f.rec.Code, f.rec.Body.String())
	}
	if got := f.currentAM().Spec.OpType; got != appv1alpha1.InstallOp {
		t.Errorf("Spec.OpType must be untouched on rejection; got %q", got)
	}
	if got := f.dispatcher.Calls(); len(got) != 0 {
		t.Errorf("dispatcher must not be triggered on rejection; got %#v", got)
	}
}

// TestApplyEnvV2_PropagatesCtrlClientError — same invariant as the
// cancel and uninstall counterparts: a mid-write spec failure must
// leave the dispatcher untouched.
func TestApplyEnvV2_PropagatesCtrlClientError(t *testing.T) {
	f := newApplyEnvV2Fixture(t, appv1alpha1.Running)

	if err := f.handler.ctrlClient.Delete(context.Background(), f.am); err != nil {
		t.Fatalf("seed: delete AM: %v", err)
	}

	f.handler.deliverApplyEnvV2(context.Background(), f.req, f.resp, f.am.Name, f.am, "sa-token-123")

	if f.rec.Code == http.StatusOK {
		t.Fatalf("expected non-200 on ctrlClient failure, got %d body=%s",
			f.rec.Code, f.rec.Body.String())
	}
	if got := f.dispatcher.Calls(); len(got) != 0 {
		t.Errorf("dispatcher must not be triggered when spec write fails; got %#v", got)
	}
}

func TestApplyEnvV2_RejectsPendingStop(t *testing.T) {
	f := newApplyEnvV2Fixture(t, appv1alpha1.Running)
	f.am.Spec.OpType = appv1alpha1.StopOp
	f.am.Annotations[domain.AnnotationOperationRequestID] = "pending-stop"
	if err := f.handler.ctrlClient.Update(context.Background(), f.am); err != nil {
		t.Fatal(err)
	}

	f.handler.deliverApplyEnvV2(context.Background(), f.req, f.resp, f.am.Name, f.am, "sa-token-123")

	if f.rec.Code != http.StatusBadRequest {
		t.Fatalf("response status = %d, want 400; body=%s", f.rec.Code, f.rec.Body.String())
	}
	got := f.currentAM()
	if got.Spec.OpType != appv1alpha1.StopOp {
		t.Fatalf("Spec.OpType = %q, want StopOp", got.Spec.OpType)
	}
	if got.Annotations[domain.AnnotationOperationRequestID] != "pending-stop" {
		t.Fatalf("request id = %q, want pending-stop", got.Annotations[domain.AnnotationOperationRequestID])
	}
	if calls := f.dispatcher.Calls(); len(calls) != 0 {
		t.Fatalf("dispatcher calls = %#v, want none", calls)
	}
}

// Compile-time router-swap safety net.
func TestApplyEnvV2_HTTPShell_SignaturesCompile(t *testing.T) {
	var (
		_ func(*restful.Request, *restful.Response) = (&Handler{}).appApplyEnvV2
		_                                           = (&Handler{}).deliverApplyEnvV2
	)
}
