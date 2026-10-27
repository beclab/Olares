package mesh

import "testing"

func TestIsEntranceEGPEPObject(t *testing.T) {
	cases := []struct {
		name   string
		obj    string
		labels map[string]string
		want   bool
	}{
		{"ext-auth suffix", "app-demo-web-entrance-ext-auth", nil, true},
		{"cookie suffix", "app-demo-web-entrance-cookie", nil, true},
		{"probe route suffix", "app-demo-web-probe-bypass", nil, true},
		{"probe policy suffix", "app-demo-web-entrance-probe", nil, true},
		{"auth-kind ext-auth", "x", map[string]string{AuthKindLabel: AuthKindEntranceExtAuth}, true},
		{"auth-kind cookie", "x", map[string]string{AuthKindLabel: AuthKindEntranceCookie}, true},
		{"jwt-authn retained by name", "shared-demo-jwt-authn", nil, false},
		{"unrelated", "shared-demo-httproute", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsEntranceEGPEPObject(tc.obj, tc.labels)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestEvaluateNoEntranceEGExtAuth(t *testing.T) {
	if !EvaluateNoEntranceEGExtAuth(0) {
		t.Fatal("expected true for zero leftovers")
	}
	if EvaluateNoEntranceEGExtAuth(1) {
		t.Fatal("expected false when leftovers remain")
	}
}
