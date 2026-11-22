package access

import (
	"encoding/json"
	"testing"
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
