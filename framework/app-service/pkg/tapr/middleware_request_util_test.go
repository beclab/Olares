package tapr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func TestCreateOrUpdateMiddlewareRequest(t *testing.T) {

	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(schema.GroupVersion{
		Group:   "apr.bytetrade.io",
		Version: "v1alpha1",
	}, &unstructured.Unstructured{})
	client := fake.NewSimpleDynamicClient(scheme)

	namespce := "default"
	request := []byte(`
apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: postgres
  namespace: default
spec:
  app: filebrowser
  appNamespace: default
  middleware: "postgres"
  postgreSQL:
    databases:
      - "db0"
      - "db1"
    password:
      value: "password"
    user: "user"
`)
	// Create a new MiddlewareRequest
	obj, err := createOrUpdateMiddlewareRequest(client, namespce, request)
	if err != nil {
		t.Fatalf("Failed to create or update middleware request: %v", err)
	}
	assert.NotNil(t, obj)
	assert.Equal(t, "postgres", obj.GetName())

	// Update the MiddlewareRequest
	request = []byte(`
apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: postgres
  namespace: default
spec:
  app: filebrowser
  appNamespace: default
  middleware: "postgres"
  postgreSQL:
    databases:
      - "db0"
      - "db1"
    password:
      value: "password-new"
    user: "user"
`)
	obj, err = createOrUpdateMiddlewareRequest(client, namespce, request)
	if err != nil {
		t.Fatalf("Failed to create or update middleware request: %v", err)
	}
	assert.NotNil(t, obj)
	assert.Equal(t, "password-new", obj.Object["spec"].(map[string]interface{})["postgreSQL"].(map[string]interface{})["password"].(map[string]interface{})["value"])
}
