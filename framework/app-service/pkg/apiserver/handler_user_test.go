package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/users"
	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestUsersListScopesRecordsToAuthenticatedRole(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := iamv1alpha2.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	user := func(name, role string) *iamv1alpha2.User {
		return &iamv1alpha2.User{ObjectMeta: metav1.ObjectMeta{
			Name: name, Annotations: map[string]string{users.UserAnnotationOwnerRole: role},
		}}
	}
	h := &Handler{ctrlClient: fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(user("alice", "normal"), user("bob", "normal"), user("admin", "admin"), user("owner", "owner")).
		WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				if key.Name == "unavailable" {
					return errors.New("user store unavailable")
				}
				return c.Get(ctx, key, obj, opts...)
			},
		}).Build()}
	container := restful.NewContainer()
	ws := new(restful.WebService).Path("/app-service/v1").Produces(restful.MIME_JSON)
	ws.Filter(h.authenticate)
	ws.Route(ws.GET("/users").To(h.handleUsers))
	container.Add(ws)

	for _, tc := range []struct {
		name    string
		headers []string
		status  int
		visible []string
	}{
		{"ordinary user", []string{"alice"}, http.StatusOK, []string{"alice"}},
		{"another tenant", []string{"bob"}, http.StatusOK, []string{"bob"}},
		{"instance admin", []string{"admin"}, http.StatusOK, []string{"admin", "alice", "bob", "owner"}},
		{"instance owner", []string{"owner"}, http.StatusOK, []string{"admin", "alice", "bob", "owner"}},
		{"missing identity", nil, http.StatusUnauthorized, nil},
		{"unknown identity", []string{"unknown"}, http.StatusUnauthorized, nil},
		{"user store unavailable", []string{"unavailable"}, http.StatusInternalServerError, nil},
		{"duplicate identity", []string{"alice", "owner"}, http.StatusUnauthorized, nil},
		{"joined identities", []string{"alice,owner"}, http.StatusUnauthorized, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/app-service/v1/users", nil)
			for _, header := range tc.headers {
				req.Header.Add(constants.BflUserKey, header)
			}
			resp := httptest.NewRecorder()
			container.ServeHTTP(resp, req)
			if resp.Code != tc.status {
				t.Fatalf("HTTP status = %d, want %d", resp.Code, tc.status)
			}
			if tc.status != http.StatusOK {
				return
			}
			var result struct {
				Code   int        `json:"code"`
				Data   []UserInfo `json:"data"`
				Totals int        `json:"totals"`
			}
			if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			var names []string
			for _, entry := range result.Data {
				names = append(names, entry.Name)
			}
			slices.Sort(names)
			if result.Code != http.StatusOK || result.Totals != len(tc.visible) || !slices.Equal(names, tc.visible) {
				t.Fatalf("visible users = %v, total %d, code %d; want %v", names, result.Totals, result.Code, tc.visible)
			}
		})
	}
}
