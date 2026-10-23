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

// uninstallV2Fixture is uninstall's flavour of cancelV2Fixture. It
// pre-populates the AM with a rich Status so any accidental clobber
// on Status (via a stray Update rather than Patch) shows up as a
// test failure, and it seeds a couple of pre-existing annotations
// so we can prove the MergeFrom patch preserves what it didn't
// touch.
type uninstallV2Fixture struct {
	t          *testing.T
	handler    *Handler
	dispatcher *lifecycleapi.FakeDispatcher
	am         *appv1alpha1.ApplicationManager

	req  *restful.Request
	resp *restful.Response
	rec  *httptest.ResponseRecorder
}

func newUninstallV2Fixture(t *testing.T, state appv1alpha1.ApplicationManagerState) *uninstallV2Fixture {
	t.Helper()

	am := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "nginx-alice",
			ResourceVersion: "1",
			Annotations: map[string]string{
				// Pre-existing annotation — must survive the
				// MergeFrom patch so the state-machine handler
				// still has the install-time metadata.
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
	httpReq := httptest.NewRequest(http.MethodPost, "/apps/nginx/uninstall", nil)
	req := restful.NewRequest(httpReq)
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)

	return &uninstallV2Fixture{
		t:          t,
		handler:    &Handler{ctrlClient: ctrlClient, opDispatcher: dispatcher},
		dispatcher: dispatcher,
		am:         am,
		req:        req,
		resp:       resp,
		rec:        rec,
	}
}

func (f *uninstallV2Fixture) currentAM() *appv1alpha1.ApplicationManager {
	f.t.Helper()
	var got appv1alpha1.ApplicationManager
	if err := f.handler.ctrlClient.Get(context.Background(), client.ObjectKey{Name: f.am.Name}, &got); err != nil {
		f.t.Fatalf("re-read AM: %v", err)
	}
	return &got
}

func defaultUninstallOpts() uninstallV2Options {
	return uninstallV2Options{
		Force:        false,
		Token:        "sa-token-123",
		UninstallAll: false,
		DeleteData:   false,
	}
}

// TestUninstallV2_DeliversTriggerOp is the happy path from Running:
// spec.OpType flips to UninstallOp, all three uninstall annotations
// land, unrelated annotations survive, status is untouched, and the
// dispatcher sees exactly one TriggerOp{UninstallOp}.
func TestUninstallV2_DeliversTriggerOp(t *testing.T) {
	f := newUninstallV2Fixture(t, appv1alpha1.Running)
	opts := defaultUninstallOpts()
	opts.UninstallAll = true
	opts.DeleteData = true

	seededStatus := f.am.Status.DeepCopy()

	f.handler.deliverUninstallV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, opts)

	if f.rec.Code != http.StatusOK {
		t.Fatalf("response status = %d, want 200; body=%s", f.rec.Code, f.rec.Body.String())
	}

	updated := f.currentAM()
	if updated.Spec.OpType != appv1alpha1.UninstallOp {
		t.Errorf("Spec.OpType = %q, want %q", updated.Spec.OpType, appv1alpha1.UninstallOp)
	}
	wantAnn := map[string]string{
		"unrelated.example.com/keep-me":     "yes",
		api.AppTokenKey:                     "sa-token-123",
		api.AppUninstallAllKey:              "true",
		api.AppDeleteDataKey:                "true",
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
		{Name: f.am.Name, Action: domain.TriggerOp{Op: appv1alpha1.UninstallOp}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("dispatcher calls mismatch:\n got %#v\nwant %#v", calls, want)
	}

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

// TestUninstallV2_BoolAnnotationEncoding pins the "%t" formatter
// choice — legacy code stamps "true"/"false" (not "1"/"0", not "yes"/
// "no") and the downstream state-machine handlers currently parse
// via strconv.ParseBool. If someone ever swaps the format string
// this test snaps.
func TestUninstallV2_BoolAnnotationEncoding(t *testing.T) {
	cases := []struct {
		name         string
		uninstallAll bool
		deleteData   bool
		wantAll      string
		wantDelete   string
	}{
		{"both false", false, false, "false", "false"},
		{"both true", true, true, "true", "true"},
		{"all only", true, false, "true", "false"},
		{"delete only", false, true, "false", "true"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newUninstallV2Fixture(t, appv1alpha1.Running)
			opts := defaultUninstallOpts()
			opts.UninstallAll = c.uninstallAll
			opts.DeleteData = c.deleteData

			f.handler.deliverUninstallV2(context.Background(), f.req, f.resp,
				f.am.Name, f.am, f.am.Spec.AppName, opts)

			if f.rec.Code != http.StatusOK {
				t.Fatalf("response status = %d, want 200", f.rec.Code)
			}
			updated := f.currentAM()
			if got := updated.Annotations[api.AppUninstallAllKey]; got != c.wantAll {
				t.Errorf("AppUninstallAllKey = %q, want %q", got, c.wantAll)
			}
			if got := updated.Annotations[api.AppDeleteDataKey]; got != c.wantDelete {
				t.Errorf("AppDeleteDataKey = %q, want %q", got, c.wantDelete)
			}
		})
	}
}

// TestUninstallV2_RejectsWhenOpNotAllowed pins the default (no
// force) rejection path: from a state where UninstallOp is not
// allowed by the TransitionTable, the handler must respond 400
// and MUST NOT deliver anything.
func TestUninstallV2_RejectsWhenOpNotAllowed(t *testing.T) {
	// PendingCanceling doesn't allow UninstallOp — user must let the
	// cancel resolve first.
	f := newUninstallV2Fixture(t, appv1alpha1.PendingCanceling)
	opts := defaultUninstallOpts()

	f.handler.deliverUninstallV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, opts)

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

// TestUninstallV2_ForceBypassesGate exercises the operator-escape-
// hatch path: force=true skips AllowsOp. Same PendingCanceling
// starting state as the rejection test above, so the pair pins the
// force flag as the ONLY thing that flips the outcome.
func TestUninstallV2_ForceBypassesGate(t *testing.T) {
	f := newUninstallV2Fixture(t, appv1alpha1.PendingCanceling)
	opts := defaultUninstallOpts()
	opts.Force = true

	f.handler.deliverUninstallV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, opts)

	if f.rec.Code != http.StatusOK {
		t.Fatalf("force=true response status = %d, want 200; body=%s", f.rec.Code, f.rec.Body.String())
	}
	if got := f.currentAM().Spec.OpType; got != appv1alpha1.UninstallOp {
		t.Errorf("Spec.OpType = %q, want %q", got, appv1alpha1.UninstallOp)
	}
	if got := len(f.dispatcher.Calls()); got != 1 {
		t.Errorf("dispatcher calls = %d, want 1", got)
	}
}

// TestUninstallV2_PropagatesCtrlClientError mirrors the cancel test:
// if the spec patch fails, no TriggerOp goes out. State-machine
// consistency invariant: the Runner MUST NEVER see a TriggerOp for
// an AM whose durable spec change didn't land.
func TestUninstallV2_PropagatesCtrlClientError(t *testing.T) {
	f := newUninstallV2Fixture(t, appv1alpha1.Running)

	if err := f.handler.ctrlClient.Delete(context.Background(), f.am); err != nil {
		t.Fatalf("seed: delete AM: %v", err)
	}

	f.handler.deliverUninstallV2(context.Background(), f.req, f.resp,
		f.am.Name, f.am, f.am.Spec.AppName, defaultUninstallOpts())

	if f.rec.Code == http.StatusOK {
		t.Fatalf("expected non-200 on ctrlClient failure, got %d body=%s",
			f.rec.Code, f.rec.Body.String())
	}
	if got := f.dispatcher.Calls(); len(got) != 0 {
		t.Errorf("dispatcher must not be triggered when spec write fails; got %#v", got)
	}
}

// TestUninstallV2_HTTPShell_SignaturesCompile is a compile-time
// safety net that both HTTP-shell and deliver-helper signatures
// remain stable so router wire-up is a one-liner in the follow-on
// PR.
func TestUninstallV2_HTTPShell_SignaturesCompile(t *testing.T) {
	var (
		_ func(*restful.Request, *restful.Response) = (&Handler{}).uninstallV2
		_                                           = (&Handler{}).deliverUninstallV2
	)
}
