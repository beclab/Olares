package testutil

import (
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// NewFakeClient builds a fake controller-runtime client with the app-service
// scheme.
//
// Only Application is registered as a status subresource: its CRD declares
// subresources.status, so its status must be written via Status().Update/Patch.
// ApplicationManager's CRD declares an empty subresources block (no /status),
// so appstate.updateStatus persists status through a plain Patch; registering
// it as a status subresource here would make the fake client silently drop
// those status writes.
func NewFakeClient(objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithScheme(NewScheme()).
		WithStatusSubresource(&appv1alpha1.Application{}).
		WithObjects(objs...).
		Build()
}
