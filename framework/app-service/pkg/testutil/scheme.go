// Package testutil provides shared helpers for app-service unit tests:
// a scheme builder, a fake controller-runtime client, fixtures for the
// custom resources and workloads, and a fake HelmOps implementation.
package testutil

import (
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// NewScheme returns a scheme registered with the core kubernetes types
// (corev1, appsv1, ...) plus the app-service custom resources.
func NewScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(appv1alpha1.AddToScheme(s))
	utilruntime.Must(sysv1alpha1.AddToScheme(s))
	return s
}
