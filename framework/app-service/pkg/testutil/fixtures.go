package testutil

import (
	"encoding/json"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AMOption mutates an ApplicationManager fixture.
type AMOption func(*appv1alpha1.ApplicationManager)

// NewAppManager builds an ApplicationManager fixture. By default it represents
// a market app owned by "alice" with an Install op pending.
func NewAppManager(name string, opts ...AMOption) *appv1alpha1.ApplicationManager {
	am := &appv1alpha1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: appv1alpha1.ApplicationManagerSpec{
			AppName:      name,
			AppNamespace: name,
			AppOwner:     "alice",
			Source:       "market",
			Type:         appv1alpha1.App,
			OpType:       appv1alpha1.InstallOp,
		},
	}
	for _, o := range opts {
		o(am)
	}
	return am
}

func WithNamespace(ns string) AMOption {
	return func(am *appv1alpha1.ApplicationManager) { am.Spec.AppNamespace = ns }
}

func WithOwner(owner string) AMOption {
	return func(am *appv1alpha1.ApplicationManager) { am.Spec.AppOwner = owner }
}

func WithSource(source string) AMOption {
	return func(am *appv1alpha1.ApplicationManager) { am.Spec.Source = source }
}

func WithOpType(op appv1alpha1.OpType) AMOption {
	return func(am *appv1alpha1.ApplicationManager) {
		am.Spec.OpType = op
		am.Status.OpType = op
	}
}

func WithState(state appv1alpha1.ApplicationManagerState) AMOption {
	return func(am *appv1alpha1.ApplicationManager) {
		am.Status.State = state
		now := metav1.Now()
		am.Status.StatusTime = &now
	}
}

func WithConfigJSON(cfg string) AMOption {
	return func(am *appv1alpha1.ApplicationManager) { am.Spec.Config = cfg }
}

// WithConfig marshals cfg into the manager's Spec.Config.
func WithConfig(t *testing.T, cfg *appcfg.ApplicationConfig) AMOption {
	t.Helper()
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal app config: %v", err)
	}
	return WithConfigJSON(string(b))
}

func WithToken(token string) AMOption {
	return func(am *appv1alpha1.ApplicationManager) { am.Annotations[api.AppTokenKey] = token }
}

func WithAnnotation(k, v string) AMOption {
	return func(am *appv1alpha1.ApplicationManager) { am.Annotations[k] = v }
}

func WithOpID(id string) AMOption {
	return func(am *appv1alpha1.ApplicationManager) { am.Status.OpID = id }
}

// NewDeployment builds a Deployment fixture with the given replica count.
func NewDeployment(name, namespace string, replicas int32) *appsv1.Deployment {
	labels := map[string]string{"app": name}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: name, Image: "test:latest"}},
				},
			},
		},
	}
}

// NewStatefulSet builds a StatefulSet fixture with the given replica count.
func NewStatefulSet(name, namespace string, replicas int32) *appsv1.StatefulSet {
	labels := map[string]string{"app": name}
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: name, Image: "test:latest"}},
				},
			},
		},
	}
}

// NewReadyPod builds a Pod whose single container is Started and Ready.
func NewReadyPod(name, namespace string, labels map[string]string) *corev1.Pod {
	started := true
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "main", Image: "test:latest"}},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "main", Ready: true, Started: &started},
			},
		},
	}
}
