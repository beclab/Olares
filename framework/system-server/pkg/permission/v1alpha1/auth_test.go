package permission

import (
	"encoding/json"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bytetrade.io/web3os/system-server/pkg/constants"
	"github.com/emicklei/go-restful/v3"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestNonceWorkloadAuthorization(t *testing.T) {
	valid := authv1.TokenReviewStatus{Authenticated: true,
		User: authv1.UserInfo{Username: nonceCaller}, Audiences: []string{nonceAudience}}
	cases := []struct {
		name     string
		headers  []string
		peer     string
		status   authv1.TokenReviewStatus
		apiError error
		want     int
		review   bool
	}{
		{"authelia through local proxy", []string{"Bearer projected-token"}, "10.233.1.4:5000", valid, nil, 200, true},
		{"authelia after pod re-IP", []string{"Bearer projected-token"}, "10.233.2.9:6000", valid, nil, 200, true},
		{"anonymous direct Pod access", nil, "10.233.1.4:5000", valid, nil, 403, false},
		{"anonymous loopback bypass", nil, "127.0.0.1:5000", valid, nil, 403, false},
		{"forged bearer", []string{"Bearer forged"}, "10.233.1.4:5000", authv1.TokenReviewStatus{}, nil, 403, true},
		{"shared old account", []string{"Bearer token"}, "10.233.1.4:5000", authv1.TokenReviewStatus{Authenticated: true, User: authv1.UserInfo{Username: "system:serviceaccount:os-framework:os-internal"}, Audiences: []string{nonceAudience}}, nil, 403, true},
		{"cross-tenant same account name", []string{"Bearer token"}, "10.233.1.4:5000", authv1.TokenReviewStatus{Authenticated: true, User: authv1.UserInfo{Username: "system:serviceaccount:user-system-bob:authelia-backend"}, Audiences: []string{nonceAudience}}, nil, 403, true},
		{"API audience token replay", []string{"Bearer token"}, "10.233.1.4:5000", authv1.TokenReviewStatus{Authenticated: true, User: valid.User, Audiences: []string{"https://kubernetes.default.svc"}}, nil, 403, true},
		{"audience unaware review", []string{"Bearer token"}, "10.233.1.4:5000", authv1.TokenReviewStatus{Authenticated: true, User: valid.User}, nil, 403, true},
		{"review error status", []string{"Bearer token"}, "10.233.1.4:5000", authv1.TokenReviewStatus{Authenticated: true, User: valid.User, Audiences: valid.Audiences, Error: "invalid"}, nil, 403, true},
		{"revoked or expired token", []string{"Bearer token"}, "10.233.1.4:5000", authv1.TokenReviewStatus{Error: "expired"}, nil, 403, true},
		{"API unavailable", []string{"Bearer token"}, "10.233.1.4:5000", valid, errors.New("sensitive-api-error"), 503, true},
		{"duplicate credentials", []string{"Bearer token", "Bearer token"}, "10.233.1.4:5000", valid, nil, 403, false},
		{"comma joined credentials", []string{"Bearer token, Bearer other"}, "10.233.1.4:5000", valid, nil, 403, false},
		{"missing token", []string{"Bearer"}, "10.233.1.4:5000", valid, nil, 403, false},
		{"oversized token", []string{"Bearer " + strings.Repeat("a", 16385)}, "10.233.1.4:5000", valid, nil, 403, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			client.PrependReactor("create", "tokenreviews", func(action ktesting.Action) (bool, runtime.Object, error) {
				review := action.(ktesting.CreateAction).GetObject().(*authv1.TokenReview)
				if len(review.Spec.Audiences) != 1 || review.Spec.Audiences[0] != nonceAudience || review.Spec.Token == "" {
					t.Fatalf("unscoped review: %+v", review.Spec.Audiences)
				}
				return true, &authv1.TokenReview{Status: tc.status}, tc.apiError
			})
			handler := Handler{kubeClientSet: client}
			req := httptest.NewRequest(http.MethodGet, "http://system-server/permission/v1alpha1/nonce", nil)
			req.RemoteAddr = tc.peer
			for _, v := range tc.headers {
				req.Header.Add("Authorization", v)
			}
			// Neither a forged service identity, tenant, nor source IP is evidence.
			req.Header.Set("l5d-client-id", "authelia-backend.os-framework.serviceaccount.identity.linkerd.cluster.local")
			req.Header.Set("Remote-User", "alice")
			req.Header.Set("X-BFL-USER", "alice")
			req.Header.Set("X-Forwarded-For", "10.233.98.111")
			req.Header.Set("X-Real-IP", "10.233.98.111")
			response := httptest.NewRecorder()
			handler.nonce(restful.NewRequest(req), restful.NewResponse(response))
			if response.Code != tc.want {
				t.Fatalf("HTTP %d, want %d", response.Code, tc.want)
			}
			if (len(client.Actions()) != 0) != tc.review {
				t.Fatal("unexpected TokenReview invocation")
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("nonce response may be cached")
			}
			if tc.want == 200 {
				if response.Body.String() != constants.Nonce {
					t.Fatal("nonce not returned to authorized caller")
				}
			} else if strings.Contains(response.Body.String(), constants.Nonce) || strings.Contains(response.Body.String(), "sensitive-api-error") {
				t.Fatal("denial leaked credential or API error")
			}
		})
	}
}

// Exercise the actual route registration and HTTP client rather than invoking
// the handler alone: method/path changes and direct Pod-style requests must not
// bypass the workload credential check.
func TestNonceRegisteredRouteAndReviewDeadline(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var review authv1.TokenReview
		if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		if r.URL.Path != "/apis/authentication.k8s.io/v1/tokenreviews" || len(review.Spec.Audiences) != 1 || review.Spec.Audiences[0] != nonceAudience {
			t.Error("review endpoint or audience changed")
		}
		if review.Spec.Token == "wait" {
			<-r.Context().Done()
			return
		}
		review.TypeMeta = metav1.TypeMeta{APIVersion: "authentication.k8s.io/v1", Kind: "TokenReview"}
		review.Spec = authv1.TokenReviewSpec{}
		review.Status = authv1.TokenReviewStatus{Authenticated: true, User: authv1.UserInfo{Username: nonceCaller}, Audiences: []string{nonceAudience}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&review)
	}))
	defer api.Close()
	container := restful.NewContainer()
	if err := AddPermissionControlToContainer(container, &PermissionControlSet{}, &rest.Config{Host: api.URL, ContentConfig: rest.ContentConfig{ContentType: "application/json", AcceptContentTypes: "application/json"}}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method, path, token string
		want                int
	}{
		{"GET", "/permission/v1alpha1/nonce", "valid", 200},
		{"POST", "/permission/v1alpha1/nonce", "valid", 405},
		{"GET", "/permission/v1alpha1/nonce", "", 403},
		{"GET", "/permission/v1alpha1/nonce", "wait", 503},
	} {
		request := httptest.NewRequest(tc.method, tc.path, nil)
		request.RemoteAddr = "127.0.0.1:1234"
		if tc.token != "" {
			request.Header.Set("Authorization", "Bearer "+tc.token)
		}
		response := httptest.NewRecorder()
		started := time.Now()
		container.ServeHTTP(response, request)
		if response.Code != tc.want {
			t.Fatalf("%s %s: HTTP %d, want %d", tc.method, tc.path, response.Code, tc.want)
		}
		if tc.token == "wait" && time.Since(started) > 2*time.Second {
			t.Fatal("TokenReview exceeded bounded deadline")
		}
	}
	for _, path := range []string{"/permission/v1alpha1/nonce/", "/permission/v1alpha1/%6eonce", "/permission/v1alpha1/nonce/extra", "/permission/v1alpha1/../v1alpha1/nonce"} {
		request := httptest.NewRequest("GET", path, nil)
		request.Header.Set("l5d-client-id", nonceCaller)
		response := httptest.NewRecorder()
		container.ServeHTTP(response, request)
		if response.Code == 200 || strings.Contains(response.Body.String(), constants.Nonce) {
			t.Fatalf("uncredentialed path bypass: %s", path)
		}
	}
}
