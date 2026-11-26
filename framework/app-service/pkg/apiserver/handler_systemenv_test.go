package apiserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestBatchUpdateSystemEnvsPreservesAdministratorBoundary(t *testing.T) {
	roles := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/apis/iam.kubesphere.io/v1alpha2/users" {
			t.Errorf("unexpected role lookup: %s %s", req.Method, req.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiVersion":"iam.kubesphere.io/v1alpha2","kind":"UserList","items":[
			{"metadata":{"name":"alice","annotations":{"bytetrade.io/owner-role":"owner"}}},
			{"metadata":{"name":"bob","annotations":{"bytetrade.io/owner-role":"normal"}}}
		]}`))
	}))
	defer roles.Close()

	for _, tc := range []struct {
		user   string
		status int
		value  string
	}{
		{"alice", http.StatusOK, "updated"},
		{"bob", http.StatusBadRequest, "previous"},
	} {
		t.Run(tc.user, func(t *testing.T) {
			env := &sysv1alpha1.SystemEnv{
				ObjectMeta: metav1.ObjectMeta{Name: "test-setting"},
				EnvVarSpec: sysv1alpha1.EnvVarSpec{EnvName: "TEST_SETTING", Value: "previous", Editable: true},
			}
			ctrlClient := testutil.NewFakeClient(env)
			handler := &Handler{ctrlClient: ctrlClient, kubeConfig: &rest.Config{Host: roles.URL}}
			httpReq := httptest.NewRequest(http.MethodPut, "/app-service/v1/systemenvs",
				strings.NewReader(`[{"envName":"TEST_SETTING","value":"updated"}]`))
			httpReq.Header.Set("Content-Type", restful.MIME_JSON)
			req := restful.NewRequest(httpReq)
			req.SetAttribute(constants.UserContextAttribute, tc.user)
			rec := httptest.NewRecorder()
			resp := restful.NewResponse(rec)
			resp.SetRequestAccepts(restful.MIME_JSON)
			handler.batchUpdateSystemEnvs(req, resp)
			if rec.Code != tc.status {
				t.Fatalf("save returned HTTP %d, want %d", rec.Code, tc.status)
			}
			var saved sysv1alpha1.SystemEnv
			if err := ctrlClient.Get(t.Context(), client.ObjectKeyFromObject(env), &saved); err != nil {
				t.Fatal(err)
			}
			if saved.Value != tc.value {
				t.Fatal("save did not preserve the expected administrator boundary")
			}
		})
	}
}
