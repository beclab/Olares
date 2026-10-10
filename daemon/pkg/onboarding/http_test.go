package onboarding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPOnboarding(t *testing.T) {
	s := testService()
	h := Handler(func(context.Context) (*Service, error) { return s, nil })
	call := func(method, path, body, password string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if password != "" {
			r.SetBasicAuth("alice", password)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := call("GET", "/onboarding", "", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "waiting-for-user") {
		t.Fatalf("initial status: %s", w.Body.String())
	}
	for _, body := range []string{`{}`, `{"username":"alice","password":"short","did":"did:key:alice"}`, `{"username":"alice","password":"initial-password","did":"did:key:alice","domain":"example.com"}`} {
		if w := call("POST", "/onboarding", body, ""); w.Code != http.StatusBadRequest {
			t.Fatalf("invalid body accepted: %d", w.Code)
		}
	}
	data, _ := json.Marshal(request())
	if w := call("POST", "/onboarding", string(data), ""); w.Code != 200 {
		t.Fatalf("create: %s", w.Body.String())
	}
	for _, password := range []string{"", "wrong-password"} {
		if w := call("GET", "/onboarding", "", password); w.Code != 401 {
			t.Fatalf("status exposed: %d", w.Code)
		}
		if w := call("PUT", "/onboarding/password", `{"password":"new-password"}`, password); w.Code != 401 {
			t.Fatalf("reset admitted: %d", w.Code)
		}
	}
	if w := call("GET", "/onboarding", "", request().Password); w.Code != 200 || !strings.Contains(w.Body.String(), "creating") {
		t.Fatalf("status: %s", w.Body.String())
	}
	ready(t, s)
	if w := call("PUT", "/onboarding/password", `{"password":"new-password"}`, request().Password); w.Code != 200 || !strings.Contains(w.Body.String(), "completed") {
		t.Fatalf("reset: %s", w.Body.String())
	}
	if w := call("POST", "/onboarding", string(data), ""); w.Code != 410 {
		t.Fatalf("creation reopened: %d", w.Code)
	}
	if w := call("PUT", "/onboarding/password", `{"password":"another-password"}`, request().Password); w.Code != 410 {
		t.Fatalf("reset remained available: %d", w.Code)
	}
	if w := call("POST", "/command/install", "", ""); w.Code != 404 {
		t.Fatalf("management route exposed: %d", w.Code)
	}
}
