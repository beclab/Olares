package permission

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bytetrade.io/web3os/system-server/pkg/constants"
	"github.com/emicklei/go-restful/v3"
)

func TestNonceRequiresExplicitMeshEnrollment(t *testing.T) {
	for _, mode := range []string{"", "ip", "token", "linkerd"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("NONCE_AUTHORIZATION_MODE", mode)
			h := newHandler(&PermissionControlSet{}, nil)
			req := httptest.NewRequest("GET", noncePath, nil)
			req.RemoteAddr = "10.233.1.4:5000"
			response := httptest.NewRecorder()
			h.nonce(restful.NewRequest(req), restful.NewResponse(response))
			if mode == "linkerd" {
				if response.Code != 200 || response.Body.String() != constants.Nonce {
					t.Fatal("enrolled endpoint incompatible with existing Authelia")
				}
			} else {
				if response.Code != 503 || strings.Contains(response.Body.String(), constants.Nonce) {
					t.Fatal("unenrolled image exposed nonce")
				}
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("nonce may be cached")
			}
		})
	}
}

func TestNonceRejectsPathNormalizationBypass(t *testing.T) {
	h := Handler{nonceMeshRequired: true}
	for _, path := range []string{
		noncePath + "/", noncePath + "?tenant=bob", "/permission/v1alpha1/%6eonce",
		"/permission/v2alpha1/../v1alpha1/nonce", "/provider/v2alpha1/../../permission/v1alpha1/nonce",
		"/permission/v1alpha1//nonce", "http://system-server" + noncePath,
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			// Model a backend router resolving an alternate path to the nonce.
			// The wire URI must still be canonical, independent of decoded Path.
			req.URL.Path = noncePath
			req.Header.Set("Remote-User", "alice")
			req.Header.Set("X-BFL-USER", "bob")
			req.Header.Set("l5d-client-id", "authelia-backend.os-framework.serviceaccount.identity.linkerd.cluster.local")
			req.Header.Set("Authorization", "Bearer forged")
			response := httptest.NewRecorder()
			h.nonce(restful.NewRequest(req), restful.NewResponse(response))
			if response.Code != 404 || strings.Contains(response.Body.String(), constants.Nonce) {
				t.Fatal("noncanonical path obtained nonce")
			}
		})
	}
	req := httptest.NewRequest("POST", noncePath, nil)
	response := httptest.NewRecorder()
	h.nonce(restful.NewRequest(req), restful.NewResponse(response))
	if response.Code != 404 {
		t.Fatal("wrong method accepted")
	}
	req = httptest.NewRequest("GET", noncePath, nil)
	req.URL.Path = "/permission/v2alpha1/register"
	response = httptest.NewRecorder()
	h.nonce(restful.NewRequest(req), restful.NewResponse(response))
	if response.Code != 404 {
		t.Fatal("decoded path mismatch accepted")
	}
}

func TestNonceRegisteredRouteRemainsTokenFree(t *testing.T) {
	t.Setenv("NONCE_AUTHORIZATION_MODE", "linkerd")
	c := restful.NewContainer()
	if err := AddPermissionControlToContainer(c, &PermissionControlSet{}, nil); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, noncePath, nil)
	response := httptest.NewRecorder()
	c.ServeHTTP(response, req)
	if response.Code != 200 || response.Body.String() != constants.Nonce {
		t.Fatal("existing Authelia requires an added token")
	}
}
