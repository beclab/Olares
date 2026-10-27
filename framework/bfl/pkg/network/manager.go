package network

import (
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	iamv1 "github.com/beclab/api/iam/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func NewManager(config *rest.Config) (manager.Manager, error) {
	scheme := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{clientgoscheme.AddToScheme, appv1.AddToScheme, sysv1.AddToScheme, iamv1.AddToScheme} {
		if err := add(scheme); err != nil {
			return nil, err
		}
	}
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme: scheme,
		// Restrict network resources without excluding applications in personal
		// namespaces or cluster-scoped users, nodes and machine settings.
		Cache: cache.Options{ByObject: map[client.Object]cache.ByObject{
			&corev1.Secret{}:     {Namespaces: map[string]cache.Config{Namespace: {}}},
			&corev1.ConfigMap{}:  {Namespaces: map[string]cache.Config{Namespace: {}}},
			&corev1.Pod{}:        {Namespaces: map[string]cache.Config{Namespace: {}}},
			&appsv1.Deployment{}: {Namespaces: map[string]cache.Config{Namespace: {}}},
		}},
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
	})
	if err != nil {
		return nil, err
	}
	// Reads immediately after a write must not see an old informer snapshot.
	direct, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	controller := &Controller{Store: &Store{Client: direct}}
	if err := controller.SetupWithManager(mgr); err != nil {
		return nil, err
	}
	if err := (&ProxyController{Controller: controller}).SetupWithManager(mgr); err != nil {
		return nil, err
	}
	if err := (&EntranceController{Store: controller.Store}).SetupWithManager(mgr); err != nil {
		return nil, err
	}
	return mgr, nil
}
