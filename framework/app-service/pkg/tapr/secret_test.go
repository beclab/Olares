package tapr

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateOrUpdateSecret(t *testing.T) {
	var (
		appName                   = "app"
		Namespace                 = "namespace"
		middleware MiddlewareType = "postgres"
	)
	clientset := fake.NewSimpleClientset()

	// createOrUpdateSecret should create a new Secret
	err := createOrUpdateSecret(clientset, appName, Namespace, middleware)
	assert.NoError(t, err)

	// ensure the Secret has been created
	name := fmt.Sprintf("%s-%s-%s-password", appName, Namespace, middleware)
	secret, err := clientset.CoreV1().Secrets(Namespace).Get(context.Background(), name, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, "Opaque", string(secret.Type))
	assert.Equal(t, map[string][]byte{"password": secret.Data["password"]}, secret.Data)

	// createOrUpdateSecret should update an existing Secret
	originalPassword := secret.Data["password"]
	err = createOrUpdateSecret(clientset, appName, Namespace, middleware)
	assert.NoError(t, err)

	// make sure existing Secret does not change password
	secret, err = clientset.CoreV1().Secrets(Namespace).Get(context.Background(), name, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, originalPassword, secret.Data["password"])
}
