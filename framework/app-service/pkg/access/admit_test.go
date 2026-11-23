package access

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
)

func TestAdmitHTTPRules(t *testing.T) {
	deny := AdmitHTTPRules(AdmissionHTTPRulesRequest{})
	if deny.Allowed {
		t.Fatal("empty rules must be denied")
	}
	allow := AdmitHTTPRules(AdmissionHTTPRulesRequest{
		Rules: []HTTPRule{{Method: "GET", Path: "/state", PathType: PathTypeExact}},
	})
	if !allow.Allowed {
		t.Fatalf("valid rules denied: %s", allow.Message)
	}
}

func TestAdmitHTTPRulesJSON(t *testing.T) {
	raw, err := json.Marshal(AdmissionHTTPRulesRequest{
		Rules: []HTTPRule{{Method: "GET", Path: "^/x$", PathType: PathTypeRegularExpression}},
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := AdmitHTTPRulesJSON(raw)
	if err != nil || !resp.Allowed {
		t.Fatalf("json admit: err=%v allowed=%v msg=%s", err, resp.Allowed, resp.Message)
	}
	resp, err = AdmitHTTPRulesJSON(nil)
	if err == nil || resp.Allowed {
		t.Fatal("empty payload must fail closed")
	}
	resp, err = AdmitHTTPRulesJSON([]byte(`{`))
	if err == nil || resp.Allowed {
		t.Fatal("invalid json must fail closed")
	}
}

func TestCatalogWebhookUsesFormalRulesAndRejectsInvalidRegex(t *testing.T) {
	for _, tc := range []struct {
		name, namespace, rules string
		allowed                bool
	}{
		{"formal", PolicyNamespace, `"rules":[{"method":"GET","path":"/state"}]`, true},
		{"invalid regex", PolicyNamespace, `"rules":[{"method":"GET","path":"^/state/[$","pathType":"RegularExpression"}]`, false},
		{"legacy input", PolicyNamespace, `"http":[{"method":"GET","path":"/state"}]`, false},
		{"wrong namespace", "tenant", `"rules":[{"method":"GET","path":"/state"}]`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"apiVersion":"admission.k8s.io/v1","kind":"AdmissionReview","request":{"uid":"request-1","operation":"CREATE","namespace":%q,"resource":{"group":"access.olares.io","version":"v1alpha1","resource":"appcapabilitysets"},"object":{"spec":{"capabilities":[{%s}]}}}}`, tc.namespace, tc.rules)
			req := httptest.NewRequest(http.MethodPost, CatalogWebhookPath, strings.NewReader(body))
			res := httptest.NewRecorder()
			CatalogAdmissionHandler(res, req)
			var got admissionv1.AdmissionReview
			if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil ||
				got.Response == nil || got.Response.UID != "request-1" || got.Response.Allowed != tc.allowed {
				t.Fatalf("unexpected admission decision: %d %s", res.Code, res.Body.String())
			}
		})
	}
}
