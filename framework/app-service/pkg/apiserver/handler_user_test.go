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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func TestDeleteUserAuthorizesCurrentActorBeforeMutation(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := iamv1alpha2.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name, caller, target string
		status               int
	}{
		{"owner cleans failed tenant", "alice", "bob", http.StatusOK},
		{"admin deletes ordinary tenant", "olivia", "bob", http.StatusOK},
		{"owner can manage admin", "alice", "olivia", http.StatusOK},
		{"ordinary cross-tenant caller", "bob", "charlie", http.StatusForbidden},
		{"admin cannot delete peer", "olivia", "charlie-admin", http.StatusForbidden},
		{"admin cannot delete owner", "olivia", "alice", http.StatusForbidden},
		{"owner cannot delete self", "alice", "alice", http.StatusForbidden},
		{"unknown caller", "unknown", "bob", http.StatusUnauthorized},
		{"target was removed", "alice", "missing", http.StatusBadRequest},
		{"caller is being removed", "deleting", "bob", http.StatusForbidden},
		{"caller lookup fails", "unavailable", "bob", http.StatusInternalServerError},
		{"target changed during authorization", "alice", "changes", http.StatusConflict},
		{"anonymous caller", "", "bob", http.StatusUnauthorized},
		{"creation still in progress", "alice", "charlie", http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			user := func(name, role, state string) *iamv1alpha2.User {
				return &iamv1alpha2.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: name, Annotations: map[string]string{users.UserAnnotationOwnerRole: role},
						Finalizers: []string{"fixture.example/cleanup"},
					},
					Status: iamv1alpha2.UserStatus{State: iamv1alpha2.UserState(state)},
				}
			}
			deleting := user("deleting", "admin", "Created")
			now := metav1.Now()
			deleting.DeletionTimestamp = &now
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(deleting,
				user("alice", "owner", "Created"), user("bob", "normal", "Failed"),
				user("charlie", "normal", "Creating"), user("olivia", "admin", "Created"),
				user("charlie-admin", "admin", "Created"), user("changes", "normal", "Failed"),
			).WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if key.Name == "unavailable" {
						return errors.New("user store unavailable")
					}
					return c.Get(ctx, key, obj, opts...)
				},
				Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					if obj.GetName() == "changes" {
						return apierrors.NewConflict(schema.GroupResource{Group: "iam.kubesphere.io", Resource: "users"}, "changes", errors.New("target modified"))
					}
					return c.Patch(ctx, obj, patch, opts...)
				},
			}).Build()
			h := &Handler{ctrlClient: c}
			ws := new(restful.WebService).Path("/app-service/v1").Produces(restful.MIME_JSON)
			ws.Filter(h.authenticate)
			ws.Route(ws.DELETE("/users/{user}").To(h.deleteUser))
			container := restful.NewContainer()
			container.Add(ws)
			req := httptest.NewRequest(http.MethodDelete, "/app-service/v1/users/"+tc.target, nil)
			if tc.caller != "" {
				req.Header.Set(constants.BflUserKey, tc.caller)
			}
			resp := httptest.NewRecorder()
			container.ServeHTTP(resp, req)
			if resp.Code != tc.status {
				t.Fatalf("delete returned %d, want %d: %s", resp.Code, tc.status, resp.Body)
			}
			if tc.target == "missing" {
				return
			}
			var target iamv1alpha2.User
			if err := c.Get(t.Context(), client.ObjectKey{Name: tc.target}, &target); err != nil {
				t.Fatal(err)
			}
			deleter := target.Annotations[users.AnnotationUserDeleter]
			if tc.status == http.StatusOK {
				if target.DeletionTimestamp.IsZero() || deleter != tc.caller {
					t.Fatal("approved deletion did not run with an attributed finalizer")
				}
			} else if !target.DeletionTimestamp.IsZero() || deleter != "" {
				t.Fatal("rejected request changed the target")
			}
		})
	}
}

func TestUserStatusReportsDeletionWithoutActivationMetadata(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := iamv1alpha2.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	now := metav1.Now()
	target := &iamv1alpha2.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bob", DeletionTimestamp: &now, Finalizers: []string{"fixture.example/cleanup"},
		},
		Status: iamv1alpha2.UserStatus{State: "Failed"},
	}
	h := &Handler{ctrlClient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(target).Build()}
	ws := new(restful.WebService).Path("/app-service/v1").Produces(restful.MIME_JSON)
	ws.Filter(h.authenticate)
	ws.Route(ws.GET("/users/{user}/status").To(h.userStatus))
	container := restful.NewContainer().Add(ws)

	for user, want := range map[string]string{"bob": "Deleting", "already-removed": "Deleted"} {
		t.Run(want, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/app-service/v1/users/"+user+"/status", nil)
			req.Header.Set(constants.BflUserKey, "alice")
			resp := httptest.NewRecorder()
			container.ServeHTTP(resp, req)
			var body struct {
				Data struct{ Name, Status string }
			}
			if resp.Code != http.StatusOK || json.Unmarshal(resp.Body.Bytes(), &body) != nil ||
				body.Data.Name != user || body.Data.Status != want {
				t.Fatalf("deletion status returned %d: %s, want %s", resp.Code, resp.Body, want)
			}
		})
	}
}
