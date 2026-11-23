package access

import (
	"encoding/json"
	"testing"
)

func TestAdmitProtectedConfigMap(t *testing.T) {
	if !AdmitProtectedConfigMap("system:serviceaccount:default:app", "other-cm").Allowed {
		t.Fatal("unprotected cm")
	}
	deny := AdmitProtectedConfigMap("system:serviceaccount:wordpress-alice:default", "app-access-runtime")
	if deny.Allowed {
		t.Fatal("app must not mutate posture cm")
	}
	allow := AdmitProtectedConfigMap("system:serviceaccount:os-app-access:access-admission", "app-access-runtime")
	if !allow.Allowed {
		t.Fatal("platform sa allowed")
	}
	if !AdmitProtectedConfigMap("system:masters", "olares-http-mesh-wave").Allowed {
		t.Fatal("masters allowed")
	}
}

func TestAdmitAdmissionReviewConfigMap(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{
		"apiVersion": "admission.k8s.io/v1",
		"kind":       "AdmissionReview",
		"request": map[string]any{
			"uid": "1",
			"resource": map[string]any{"resource": "configmaps"},
			"userInfo": map[string]any{"username": "system:serviceaccount:default:evil"},
			"object": map[string]any{
				"kind":     "ConfigMap",
				"metadata": map[string]any{"name": "app-access-runtime", "namespace": "os-app-access"},
				"data":     map[string]any{"endState": "false"},
			},
		},
	})
	out, err := AdmitAdmissionReview(raw)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	_ = json.Unmarshal(out, &m)
	resp := m["response"].(map[string]any)
	if resp["allowed"].(bool) {
		t.Fatal("expected deny")
	}
}
