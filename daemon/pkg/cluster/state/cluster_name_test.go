package state

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestRefreshClusterNameUsesMachineConfiguration(t *testing.T) {
	restoreState(t)
	for _, name := range []string{"lab", ""} {
		client := fake.NewSimpleDynamicClient(runtime.NewScheme())
		client.PrependReactor("get", "terminus", func(action clienttesting.Action) (bool, runtime.Object, error) {
			return true, &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": "sys.bytetrade.io/v1alpha1", "kind": "Terminus",
				"metadata": map[string]interface{}{"name": "terminus"},
				"spec":     map[string]interface{}{"localName": name},
			}}, nil
		})
		username := "alice@olares.local"
		TerminusStateMu.Lock()
		CurrentState.TerminusName = &username
		err := refreshClusterName(context.Background(), client)
		got := CurrentState.ClusterName
		TerminusStateMu.Unlock()
		if err != nil {
			t.Fatal(err)
		}
		if got != name {
			t.Fatalf("clusterName = %q, want machine name %q", got, name)
		}
	}
}
