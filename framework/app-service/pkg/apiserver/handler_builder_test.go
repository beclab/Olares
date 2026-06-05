package apiserver

import (
	"testing"

	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Regression for #3225: when WithAppInformer fails to build a clientset the
// error must be carried into Build() and returned, not swallowed into a
// nil-pointer panic.
func TestHandlerBuilderPropagatesInformerError(t *testing.T) {
	// exec and auth providers are mutually exclusive, so clientset
	// construction fails deterministically without any cluster.
	cfg := &rest.Config{
		Host:         "http://localhost:8080",
		ExecProvider: &clientcmdapi.ExecConfig{},
		AuthProvider: &clientcmdapi.AuthProviderConfig{Name: "x"},
	}
	b := (&handlerBuilder{kubeConfig: cfg}).WithAppInformer()
	if b.err == nil {
		t.Fatal("WithAppInformer should record an error for conflicting exec/auth providers")
	}

	h, err := b.Build()
	if err == nil {
		t.Fatal("Build should return the accumulated informer error")
	}
	if h != nil {
		t.Errorf("Build should return a nil handler on error, got %#v", h)
	}
}
