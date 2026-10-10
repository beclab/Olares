package onboarding

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestResetUsesAuthenticatedAuthProvider(t *testing.T) {
	responseCode := 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/cli/api/reset/alice/password" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		for _, header := range []string{"Authorization", "Olares-CLI-Authorization"} {
			if r.Header.Get(header) != "Bearer test-token" {
				t.Errorf("missing %s", header)
			}
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["password"] != passwordWire("new-password") {
			t.Error("incorrect password wire format")
		}
		w.WriteHeader(responseCode)
	}))
	defer server.Close()
	host, portStr, _ := net.SplitHostPort(server.Listener.Addr().String())
	port, _ := strconv.Atoi(portStr)
	kube := fake.NewSimpleClientset(&corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: Namespace, Name: "auth-provider-svc"},
		Spec: corev1.ServiceSpec{ClusterIP: host, Ports: []corev1.ServicePort{{Name: "server", Port: int32(port)}}}})
	kube.PrependReactor("create", "serviceaccounts", func(a ktesting.Action) (bool, runtime.Object, error) {
		if a.GetSubresource() != "token" {
			t.Fatal("expected token request")
		}
		return true, &authv1.TokenRequest{Status: authv1.TokenRequestStatus{Token: "test-token"}}, nil
	})
	reset := ResetWithAuthelia(kube)
	if err := reset(context.Background(), "alice", "new-password"); err != nil {
		t.Fatal(err)
	}
	responseCode = 500
	if err := reset(context.Background(), "alice", "new-password"); err == nil {
		t.Fatal("backend failure was ignored")
	}
}
