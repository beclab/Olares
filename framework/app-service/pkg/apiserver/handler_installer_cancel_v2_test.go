package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// cancelV2Fixture wires the minimum surface deliverCancelV2 needs:
// a controller-runtime client with a seeded AM, and a
// *restful.Response + httptest recorder for HTTP assertions.
//
// The outer HTTP shell (cancelV2) is skipped because its
// loadAuthorizedLifecycleAM prelude reaches into utils.GetClient()
// which reads a real kubeconfig — orthogonal to what these tests
// pin, which is the translator step.
type cancelV2Fixture struct {
	t       *testing.T
	handler *Handler
	am      *appv1alpha1.ApplicationManager

	req  *restful.Request
	resp *restful.Response
	rec  *httptest.ResponseRecorder
}

func newCancelV2Fixture(t *testing.T, state appv1alpha1.ApplicationManagerState) *cancelV2Fixture {
	t.Helper()

	am := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "nginx-alice",
			ResourceVersion: "1",
		},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName:      "nginx",
			AppNamespace: "nginx-alice",
			AppOwner:     "alice",
			OpType:       appv1alpha1.InstallOp,
		},
		Status: appv1alpha1.ApplicationManagerStatus{State: state},
	}

	ctrlClient := testutil.NewFakeClient(am)
	rec := httptest.NewRecorder()
	httpReq := httptest.NewRequest(http.MethodPost, "/apps/nginx/cancel", nil)
	req := restful.NewRequest(httpReq)
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)

	return &cancelV2Fixture{
		t:       t,
		handler: &Handler{ctrlClient: ctrlClient},
		am:      am,
		req:     req,
		resp:    resp,
		rec:     rec,
	}
}

func (f *cancelV2Fixture) currentAM() *appv1alpha1.ApplicationManager {
	f.t.Helper()
	var got appv1alpha1.ApplicationManager
	if err := f.handler.ctrlClient.Get(context.Background(), client.ObjectKey{Name: f.am.Name}, &got); err != nil {
		f.t.Fatalf("re-read AM after cancelV2: %v", err)
	}
	return &got
}

// TestCancelV2_DeliversTriggerOp is the "happy path": from an
// operating state (Installing), cancelV2 must (a) leave
// Spec.OpType=CancelOp and a new request id on the AM, (b) leave
// Status.State untouched (the state machine, not this handler,
// writes *Canceling), and (c) respond 200 with the app UID.
func TestCancelV2_DeliversTriggerOp(t *testing.T) {
	f := newCancelV2Fixture(t, appv1alpha1.Installing)

	f.handler.deliverCancelV2(context.Background(), f.req, f.resp, f.am.Name, f.am, f.am.Spec.AppName)

	if f.rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want 200; body=%s", f.rec.Code, f.rec.Body.String())
	}

	updated := f.currentAM()
	if updated.Spec.OpType != appv1alpha1.CancelOp {
		t.Errorf("Spec.OpType = %q, want %q", updated.Spec.OpType, appv1alpha1.CancelOp)
	}
	if updated.Status.State != appv1alpha1.Installing {
		t.Errorf("Status.State must be untouched by the HTTP layer; got %q, want %q",
			updated.Status.State, appv1alpha1.Installing)
	}

	// Response body carries the UID (the AM's Spec.AppName) so the
	// legacy client contract keeps working.
	var body struct {
		Data struct {
			UID string `json:"uid"`
		} `json:"data"`
	}
	if err := json.Unmarshal(f.rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.Data.UID != f.am.Spec.AppName {
		t.Errorf("response uid = %q, want %q", body.Data.UID, f.am.Spec.AppName)
	}
}

// TestCancelV2_RejectsWhenOpNotAllowed pins the "cancel from a
// terminal state" case. From Uninstalled, CancelOp is not in the
// transition table; the v2 handler must reject with 400 and leave
// Spec.OpType untouched.
func TestCancelV2_RejectsWhenOpNotAllowed(t *testing.T) {
	f := newCancelV2Fixture(t, appv1alpha1.Uninstalled)

	f.handler.deliverCancelV2(context.Background(), f.req, f.resp, f.am.Name, f.am, f.am.Spec.AppName)

	if f.rec.Code != http.StatusBadRequest {
		t.Fatalf("response status = %d, want %d; body=%s",
			f.rec.Code, http.StatusBadRequest, f.rec.Body.String())
	}
	if got := f.currentAM().Spec.OpType; got != appv1alpha1.InstallOp {
		t.Errorf("Spec.OpType must be untouched on rejection; got %q, want %q",
			got, appv1alpha1.InstallOp)
	}
}

// TestCancelV2_PropagatesCtrlClientError pins the mid-write failure
// case. When the ctrlClient rejects the spec Update (e.g. NotFound
// on a race with a concurrent delete), the handler must surface an
// HTTP error. The state machine must not see a CancelOp for an AM
// whose spec mutation didn't land.
func TestCancelV2_PropagatesCtrlClientError(t *testing.T) {
	f := newCancelV2Fixture(t, appv1alpha1.Installing)

	// Delete the AM from the fake store so the subsequent Update
	// returns NotFound.
	if err := f.handler.ctrlClient.Delete(context.Background(), f.am); err != nil {
		t.Fatalf("seed: delete AM: %v", err)
	}

	f.handler.deliverCancelV2(context.Background(), f.req, f.resp, f.am.Name, f.am, f.am.Spec.AppName)

	if f.rec.Code == http.StatusOK {
		t.Fatalf("expected non-200 on ctrlClient failure, got %d body=%s",
			f.rec.Code, f.rec.Body.String())
	}
}

// Absence assertion: cancelV2 (and its deliverCancelV2 helper) must
// NEVER touch AM.Status directly. This test enforces the invariant
// by inspecting the AM after a successful cancel and verifying the
// Status subtree is byte-for-byte the same as the seeded fixture —
// only the ResourceVersion is allowed to differ (the Spec write
// bumps it).
func TestCancelV2_DoesNotTouchStatus(t *testing.T) {
	f := newCancelV2Fixture(t, appv1alpha1.Installing)
	// Populate a rich status so any accidental clobber shows up.
	seededStatus := appv1alpha1.ApplicationManagerStatus{
		State:    appv1alpha1.Installing,
		OpType:   appv1alpha1.InstallOp,
		OpID:     "seed-op-id",
		Message:  "installing helm chart",
		Reason:   "InstallInFlight",
		Progress: "42.00",
	}
	f.am.Status = seededStatus
	if err := f.handler.ctrlClient.Update(context.Background(), f.am); err != nil {
		t.Fatalf("seed: re-persist rich status: %v", err)
	}

	f.handler.deliverCancelV2(context.Background(), f.req, f.resp, f.am.Name, f.am, f.am.Spec.AppName)

	if f.rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want 200", f.rec.Code)
	}

	updated := f.currentAM()
	if !reflect.DeepEqual(updated.Status, seededStatus) {
		t.Errorf("Status must be byte-for-byte unchanged:\n got %#v\nwant %#v",
			updated.Status, seededStatus)
	}
}

// TestCancelV2_HTTPShell_FallbackGuardCompiles pins the router
// signatures so cancelV2 stays directly wireable into go-restful.
func TestCancelV2_HTTPShell_FallbackGuardCompiles(t *testing.T) {
	// Static assertions: cancelV2 must exist on *Handler with the
	// (req, resp) signature so it is directly wireable into the
	// go-restful router without an adapter.
	var (
		_ func(*restful.Request, *restful.Response) = (&Handler{}).cancelV2
	)

	// deliverCancelV2 must exist as a separate method so unit
	// tests can bypass the auth prelude.
	var (
		_ = (&Handler{}).deliverCancelV2
	)
}
