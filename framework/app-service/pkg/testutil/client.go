package testutil

import (
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// NewFakeClient builds a fake controller-runtime client with the app-service
// scheme and the status subresource enabled for ApplicationManager and
// Application, which the appstate code patches via client.Status-style merges.
func NewFakeClient(objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithScheme(NewScheme()).
		WithStatusSubresource(&appv1alpha1.ApplicationManager{}, &appv1alpha1.Application{}).
		WithObjects(objs...).
		Build()
}
