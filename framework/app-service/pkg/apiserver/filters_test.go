package apiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/emicklei/go-restful/v3"
)

func TestAuthenticateRequiresSingleCanonicalTenantIdentity(t *testing.T) {
	c := restful.NewContainer()
	c.Filter((&Handler{}).authenticate)
	ws := (&restful.WebService{}).Path("/app-service/v1")
	tenant := func(req *restful.Request, resp *restful.Response) {
		if req.Attribute(constants.UserContextAttribute) != "alice" {
			t.Error("authenticated tenant was not passed to the handler")
		}
		resp.WriteHeader(http.StatusNoContent)
	}
	ws.Route(ws.GET("/apps").To(tenant))
	ws.Route(ws.GET("/compute-resources").To(tenant))
	ws.Route(ws.GET("/applications/{name}/attachments").To(tenant))
	ws.Route(ws.POST("/apps/{name}/install").To(tenant))
	ws.Route(ws.GET("/terminus/version").To(func(_ *restful.Request, resp *restful.Response) {
		resp.WriteHeader(http.StatusNoContent)
	}))
	c.Add(ws)

	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/apps"},
		{http.MethodGet, "/compute-resources"},
		{http.MethodGet, "/applications/example/attachments"},
		{http.MethodPost, "/apps/example/install"},
	} {
		for _, tc := range []struct {
			name   string
			values []string
			want   int
		}{
			{"authenticated", []string{"alice"}, http.StatusNoContent},
			{"anonymous", nil, http.StatusUnauthorized},
			{"empty", []string{""}, http.StatusUnauthorized},
			{"multiple tenants", []string{"alice", "bob"}, http.StatusUnauthorized},
			{"merged tenants", []string{"alice,bob"}, http.StatusUnauthorized},
			{"noncanonical identity", []string{" alice"}, http.StatusUnauthorized},
		} {
			t.Run(route.path+"/"+tc.name, func(t *testing.T) {
				req := httptest.NewRequest(route.method, "/app-service/v1"+route.path, nil)
				for _, value := range tc.values {
					req.Header.Add(constants.BflUserKey, value)
				}
				resp := httptest.NewRecorder()
				c.ServeHTTP(resp, req)
				if resp.Code != tc.want {
					t.Fatalf("got status %d, want %d", resp.Code, tc.want)
				}
			})
		}
	}
	resp := httptest.NewRecorder()
	c.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/app-service/v1/terminus/version", nil))
	if resp.Code != http.StatusNoContent {
		t.Fatalf("existing version endpoint behavior changed to %d", resp.Code)
	}
}
