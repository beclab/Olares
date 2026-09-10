package syncer

import (
	"context"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// Reconciler mirrors the ConfigMaps matching Selector in Namespace into Dir.
// Every event triggers a full resync of the directory rather than an
// incremental update.
//
// A watch event for a deleted ConfigMap no longer carries the object, so its
// keys are unknown by then. Comparing the whole directory against the keys of
// all remaining ConfigMaps sidesteps that: key removal, ConfigMap deletion,
// label removal and files left behind by a previous run all resolve on the same
// code path, with no state to keep. In exchange Dir belongs exclusively to this
// process.
type Reconciler struct {
	client.Client

	Namespace string
	Selector  labels.Selector
	Dir       string

	// mu serialises the directory sync, which SyncOnce may enter concurrently
	// with the controller's own reconciles.
	mu sync.Mutex
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("configmap-file-syncer").
		For(&corev1.ConfigMap{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}

// SyncOnce runs a single sync regardless of any watch event. Without it a
// directory holding files from a previous run would never be cleaned up when no
// ConfigMap matches the selector, because no event would ever arrive.
func (r *Reconciler) SyncOnce(ctx context.Context) error {
	_, err := r.Reconcile(ctx, ctrl.Request{})
	return err
}

func (r *Reconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// The manager's cache is configured with the same namespace and selector, so
	// these options are redundant at runtime. They are passed anyway to keep the
	// set of ConfigMaps this reconciler acts on visible here, instead of leaving
	// it as an implicit consequence of how the cache was set up.
	var list corev1.ConfigMapList
	if err := r.List(ctx, &list,
		client.InNamespace(r.Namespace),
		client.MatchingLabelsSelector{Selector: r.Selector},
	); err != nil {
		return ctrl.Result{}, err
	}

	desired := make(map[string][]byte)
	for i := range list.Items {
		cm := &list.Items[i]
		for name, value := range cm.Data {
			put(desired, cm, name, []byte(value))
		}
		for name, value := range cm.BinaryData {
			put(desired, cm, name, value)
		}
	}

	if err := syncDir(r.Dir, desired); err != nil {
		return ctrl.Result{}, err
	}

	klog.V(4).Infof("synced %d file(s) from %d configmap(s) into %s", len(desired), len(list.Items), r.Dir)
	return ctrl.Result{}, nil
}

func put(desired map[string][]byte, cm *corev1.ConfigMap, name string, value []byte) {
	if _, dup := desired[name]; dup {
		klog.Warningf("key %q of configmap %s/%s also exists in another matched configmap, overwriting",
			name, cm.Namespace, cm.Name)
	}
	desired[name] = value
}
