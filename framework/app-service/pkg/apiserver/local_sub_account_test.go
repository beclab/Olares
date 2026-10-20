package apiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/users"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateSubAccountWithoutDID(t *testing.T) {
	for _, tc := range []struct{ domain, wantName, ephemeral string }{
		{"", "member", "false"},
		{"olares.local", "member", "false"},
		{"olares.com", "member@olares.com", "true"},
	} {
		t.Run(tc.domain, func(t *testing.T) {
			scheme := runtime.NewScheme()
			if err := sysv1alpha1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			if err := iamv1alpha2.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			terminus := &sysv1alpha1.Terminus{ObjectMeta: metav1.ObjectMeta{Name: "terminus"}}
			terminus.Spec.Settings = map[string]string{"domainName": tc.domain}
			ctrlClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(terminus).Build()
			handler := &Handler{ctrlClient: ctrlClient}
			body, err := json.Marshal(UserCreateRequest{Name: "member", OwnerRole: "normal", Password: "encoded-final-password", CpuLimit: "1", MemoryLimit: "4G"})
			if err != nil {
				t.Fatal(err)
			}
			raw := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			raw.Header.Set("Content-Type", "application/json")
			req := restful.NewRequest(raw)
			req.SetAttribute(constants.UserContextAttribute, "owner")
			recorder := httptest.NewRecorder()
			handler.createUser(req, restful.NewResponse(recorder))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			var user iamv1alpha2.User
			if err := ctrlClient.Get(context.Background(), types.NamespacedName{Name: "member"}, &user); err != nil {
				t.Fatal(err)
			}
			if got := user.Annotations[users.UserAnnotationTerminusNameKey]; got != tc.wantName {
				t.Fatalf("identity=%q, want %q", got, tc.wantName)
			}
			if got := user.Annotations[users.UserAnnotationIsEphemeral]; got != tc.ephemeral {
				t.Fatalf("ephemeral=%q, want %q", got, tc.ephemeral)
			}
			if user.Spec.InitialPassword != "encoded-final-password" {
				t.Fatal("creation did not preserve the supplied final password")
			}
			if user.Annotations["bytetrade.io/user-did"] != "" {
				t.Fatal("DID must remain unbound until scanning")
			}
		})
	}
}
