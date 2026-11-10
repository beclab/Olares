package apiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/emicklei/go-restful/v3"
)

func requireDurableRequestID(t *testing.T, annotations map[string]string) string {
	t.Helper()
	id := annotations[domain.AnnotationOperationRequestID]
	if id == "" {
		t.Fatalf("annotation %q is empty", domain.AnnotationOperationRequestID)
	}
	if _, err := uuid.Parse(id); err != nil {
		t.Fatalf("annotation %q = %q, want UUID: %v", domain.AnnotationOperationRequestID, id, err)
	}
	return id
}

func TestStampDurableOperationRequestInitializesNilAnnotations(t *testing.T) {
	annotations, requestID := stampDurableOperationRequest(nil)
	if got := requireDurableRequestID(t, annotations); got != requestID {
		t.Fatalf("returned request ID = %q, stamped %q", requestID, got)
	}
}

func TestV2OperationsPersistDurableRequestID(t *testing.T) {
	tests := []struct {
		name   string
		invoke func(*testing.T) map[string]string
	}{
		{
			name: "install",
			invoke: func(t *testing.T) map[string]string {
				f := newInstallV2Fixture(t)
				f.handler.deliverInstallV2(context.Background(), f.req, f.resp,
					"nginx-alice-nginx", baseInstallAppCfg(), baseInstallReq(), baseInstallOpts())
				if f.rec.Code != http.StatusOK {
					t.Fatalf("status = %d, body=%s", f.rec.Code, f.rec.Body.String())
				}
				return f.currentAM("nginx-alice-nginx").Annotations
			},
		},
		{
			name: "upgrade",
			invoke: func(t *testing.T) map[string]string {
				f := newUpgradeV2Fixture(t, appv1alpha1.Running)
				f.handler.deliverUpgradeV2(context.Background(), f.req, f.resp,
					f.am.Name, f.am, f.am.Spec.AppName, fullUpgradeOpts())
				return f.currentAM().Annotations
			},
		},
		{
			name: "uninstall",
			invoke: func(t *testing.T) map[string]string {
				f := newUninstallV2Fixture(t, appv1alpha1.Running)
				f.handler.deliverUninstallV2(context.Background(), f.req, f.resp,
					f.am.Name, f.am, f.am.Spec.AppName, defaultUninstallOpts())
				return f.currentAM().Annotations
			},
		},
		{
			name: "cancel",
			invoke: func(t *testing.T) map[string]string {
				f := newCancelV2Fixture(t, appv1alpha1.Installing)
				f.handler.deliverCancelV2(context.Background(), f.req, f.resp,
					f.am.Name, f.am, f.am.Spec.AppName)
				return f.currentAM().Annotations
			},
		},
		{
			name: "stop",
			invoke: func(t *testing.T) map[string]string {
				f := newSuspendResumeV2Fixture(t, appv1alpha1.Running)
				f.handler.deliverSuspendV2(context.Background(), f.req, f.resp,
					f.am.Name, f.am, f.am.Spec.AppName, false)
				return f.currentAM().Annotations
			},
		},
		{
			name: "resume",
			invoke: func(t *testing.T) map[string]string {
				f := newSuspendResumeV2Fixture(t, appv1alpha1.Stopped)
				if err := f.handler.deliverResumeV2(context.Background(), f.req, f.resp,
					f.am.Name, f.am, f.am.Spec.AppName, false); err != nil {
					t.Fatalf("deliverResumeV2: %v", err)
				}
				return f.currentAM().Annotations
			},
		},
		{
			name: "apply-env",
			invoke: func(t *testing.T) map[string]string {
				f := newApplyEnvV2Fixture(t, appv1alpha1.Running)
				f.handler.deliverApplyEnvV2(context.Background(), f.req, f.resp,
					f.am.Name, f.am, "token")
				return f.currentAM().Annotations
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requireDurableRequestID(t, tc.invoke(t))
		})
	}
}

func TestV2OperationInvocationMintsFreshRequestID(t *testing.T) {
	const name = "nginx-alice-nginx"
	f := newInstallV2Fixture(t)
	invoke := func() string {
		rec := httptest.NewRecorder()
		req := restful.NewRequest(httptest.NewRequest(http.MethodPost, "/apps/nginx/install", nil))
		resp := restful.NewResponse(rec)
		resp.SetRequestAccepts(restful.MIME_JSON)

		f.handler.deliverInstallV2(context.Background(), req, resp,
			name, baseInstallAppCfg(), baseInstallReq(), baseInstallOpts())
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
		}
		return requireDurableRequestID(t, f.currentAM(name).Annotations)
	}

	first := invoke()
	rec := httptest.NewRecorder()
	req := restful.NewRequest(httptest.NewRequest(http.MethodPost, "/apps/nginx/install", nil))
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)
	f.handler.deliverInstallV2(context.Background(), req, resp,
		name, baseInstallAppCfg(), baseInstallReq(), baseInstallOpts())
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("second install status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
	if got := requireDurableRequestID(t, f.currentAM(name).Annotations); got != first {
		t.Fatalf("request id = %q, want the first id %q", got, first)
	}
}

func TestSuspendV2_PreSideEffectFailureClearsPriorDurableRequestID(t *testing.T) {
	f := newSuspendResumeV2Fixture(t, appv1alpha1.Running)
	// The seeded id matches Status.OpID, so it is already consumed.
	// Suspend still deletes it before the compute side effect; a bad
	// config then fails before a replacement id is stamped.
	f.am.Annotations[domain.AnnotationOperationRequestID] = "seed-op-id"
	f.am.Spec.Config = "{not json"
	if err := f.handler.ctrlClient.Update(context.Background(), f.am); err != nil {
		t.Fatalf("seed prior request and malformed config: %v", err)
	}

	rec := httptest.NewRecorder()
	req := restful.NewRequest(httptest.NewRequest(http.MethodPost, "/apps/nginx/suspend", nil))
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)
	f.handler.deliverSuspendV2(context.Background(), req, resp,
		f.am.Name, f.am, f.am.Spec.AppName, false)

	if got := f.currentAM().Annotations[domain.AnnotationOperationRequestID]; got != "" {
		t.Fatalf("durable request ID = %q after pre-side-effect failure, want empty", got)
	}
}
