package syncer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testNamespace = "default"
	testLabel     = "bytetrade.io/cm-sidecar"
)

func configMap(name string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      name,
			Labels:    map[string]string{testLabel: "true"},
		},
		Data: data,
	}
}

func unlabeledConfigMap(name string, data map[string]string) *corev1.ConfigMap {
	cm := configMap(name, data)
	cm.Labels = nil
	return cm
}

func newReconciler(t *testing.T, dir string, objects ...client.Object) *Reconciler {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	// An existence selector, matching how the default is parsed in main.
	selector, err := labels.Parse(testLabel)
	if err != nil {
		t.Fatalf("parse selector: %v", err)
	}

	return &Reconciler{
		Client:    fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build(),
		Namespace: testNamespace,
		Selector:  selector,
		Dir:       dir,
	}
}

// reconcile runs a sync. The request is ignored by the reconciler, which always
// converges the whole directory.
func reconcile(t *testing.T, r *Reconciler) {
	t.Helper()
	if _, err := r.Reconcile(context.Background(), ctrl.Request{}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}
}

func deleteConfigMap(t *testing.T, r *Reconciler, name string) {
	t.Helper()
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: testNamespace, Name: name}}
	if err := r.Delete(context.Background(), cm); err != nil {
		t.Fatalf("delete %s: %v", name, err)
	}
}

func updateConfigMap(t *testing.T, r *Reconciler, name string, mutate func(*corev1.ConfigMap)) {
	t.Helper()

	ctx := context.Background()
	var cm corev1.ConfigMap
	if err := r.Get(ctx, types.NamespacedName{Namespace: testNamespace, Name: name}, &cm); err != nil {
		t.Fatalf("get %s: %v", name, err)
	}
	mutate(&cm)
	if err := r.Update(ctx, &cm); err != nil {
		t.Fatalf("update %s: %v", name, err)
	}
}

func TestReconcileWritesKeys(t *testing.T) {
	dir := t.TempDir()
	cm := configMap("sample", map[string]string{"app.conf": "listen = 8080"})
	cm.BinaryData = map[string][]byte{"cert.p12": {0x00, 0x01}}

	r := newReconciler(t, dir, cm)
	reconcile(t, r)

	if got := readFile(t, filepath.Join(dir, "app.conf")); got != "listen = 8080" {
		t.Errorf("app.conf = %q", got)
	}
	if got := readFile(t, filepath.Join(dir, "cert.p12")); got != "\x00\x01" {
		t.Errorf("cert.p12 = %q", got)
	}
}

func TestReconcileMergesMultipleConfigMaps(t *testing.T) {
	dir := t.TempDir()
	r := newReconciler(t, dir,
		configMap("first", map[string]string{"a.conf": "a"}),
		configMap("second", map[string]string{"b.conf": "b"}),
	)
	reconcile(t, r)

	if got := readFile(t, filepath.Join(dir, "a.conf")); got != "a" {
		t.Errorf("a.conf = %q", got)
	}
	if got := readFile(t, filepath.Join(dir, "b.conf")); got != "b" {
		t.Errorf("b.conf = %q", got)
	}
}

func TestReconcileIgnoresConfigMapWithoutLabel(t *testing.T) {
	dir := t.TempDir()
	r := newReconciler(t, dir,
		configMap("watched", map[string]string{"watched.conf": "watched"}),
		unlabeledConfigMap("ignored", map[string]string{"ignored.conf": "ignored"}),
	)
	reconcile(t, r)

	if got := readFile(t, filepath.Join(dir, "watched.conf")); got != "watched" {
		t.Errorf("watched.conf = %q", got)
	}
	assertMissing(t, filepath.Join(dir, "ignored.conf"))
}

func TestReconcileRemovesFilesWhenLabelRemoved(t *testing.T) {
	dir := t.TempDir()
	r := newReconciler(t, dir, configMap("sample", map[string]string{"app.conf": "app"}))
	reconcile(t, r)

	updateConfigMap(t, r, "sample", func(cm *corev1.ConfigMap) {
		delete(cm.Labels, testLabel)
	})
	reconcile(t, r)

	assertMissing(t, filepath.Join(dir, "app.conf"))
}

func TestReconcileUpdatesChangedValue(t *testing.T) {
	dir := t.TempDir()
	r := newReconciler(t, dir, configMap("sample", map[string]string{"app.conf": "old"}))
	reconcile(t, r)

	updateConfigMap(t, r, "sample", func(cm *corev1.ConfigMap) {
		cm.Data["app.conf"] = "new"
	})
	reconcile(t, r)

	if got := readFile(t, filepath.Join(dir, "app.conf")); got != "new" {
		t.Errorf("app.conf = %q, want %q", got, "new")
	}
}

func TestReconcileRemovesFileOfDroppedKey(t *testing.T) {
	dir := t.TempDir()
	r := newReconciler(t, dir, configMap("sample", map[string]string{
		"keep.conf": "keep",
		"drop.conf": "drop",
	}))
	reconcile(t, r)

	updateConfigMap(t, r, "sample", func(cm *corev1.ConfigMap) {
		delete(cm.Data, "drop.conf")
	})
	reconcile(t, r)

	assertMissing(t, filepath.Join(dir, "drop.conf"))
	if got := readFile(t, filepath.Join(dir, "keep.conf")); got != "keep" {
		t.Errorf("keep.conf = %q", got)
	}
}

func TestReconcileRemovesFilesWhenConfigMapDeleted(t *testing.T) {
	dir := t.TempDir()
	r := newReconciler(t, dir,
		configMap("sample", map[string]string{"a.conf": "a", "b.conf": "b"}),
		configMap("other", map[string]string{"c.conf": "c"}),
	)
	reconcile(t, r)

	deleteConfigMap(t, r, "sample")
	reconcile(t, r)

	assertMissing(t, filepath.Join(dir, "a.conf"))
	assertMissing(t, filepath.Join(dir, "b.conf"))
	if got := readFile(t, filepath.Join(dir, "c.conf")); got != "c" {
		t.Errorf("c.conf = %q, want the surviving configmap's file to stay", got)
	}
}

func TestReconcileKeepsKeySharedWithAnotherConfigMap(t *testing.T) {
	dir := t.TempDir()
	r := newReconciler(t, dir,
		configMap("first", map[string]string{"shared.conf": "shared"}),
		configMap("second", map[string]string{"shared.conf": "shared"}),
	)
	reconcile(t, r)

	deleteConfigMap(t, r, "first")
	reconcile(t, r)

	if got := readFile(t, filepath.Join(dir, "shared.conf")); got != "shared" {
		t.Errorf("shared.conf = %q, want it kept by the surviving configmap", got)
	}
}

// The target directory belongs to the sidecar, so a file that does not come
// from a ConfigMap is removed. This is what allows a deleted ConfigMap to be
// cleaned up without tracking which file came from where.
func TestReconcileRemovesFileNotBackedByConfigMap(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "stale.conf")
	if err := os.WriteFile(stale, []byte("stale"), fileMode); err != nil {
		t.Fatalf("write: %v", err)
	}

	r := newReconciler(t, dir, configMap("sample", map[string]string{"app.conf": "app"}))
	reconcile(t, r)

	assertMissing(t, stale)
	if got := readFile(t, filepath.Join(dir, "app.conf")); got != "app" {
		t.Errorf("app.conf = %q", got)
	}
}

func TestReconcileWithNoConfigMapsEmptiesDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "stale.conf"), []byte("stale"), fileMode); err != nil {
		t.Fatalf("write: %v", err)
	}

	r := newReconciler(t, dir)
	if err := r.SyncOnce(context.Background()); err != nil {
		t.Fatalf("SyncOnce: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("dir has %d entries, want 0", len(entries))
	}
}
