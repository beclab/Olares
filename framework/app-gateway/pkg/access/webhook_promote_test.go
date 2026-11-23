package access

import (
	"context"
	"testing"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPromoteAdmissionFailurePolicyEndState(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = admissionregistrationv1.AddToScheme(scheme)

	ignore := admissionregistrationv1.Ignore
	wh := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: AdmissionWebhookConfigName},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			Name:           "access.olares.io",
			FailurePolicy:  &ignore,
			ClientConfig:   admissionregistrationv1.WebhookClientConfig{},
			SideEffects:    func() *admissionregistrationv1.SideEffectClass { s := admissionregistrationv1.SideEffectClassNone; return &s }(),
			AdmissionReviewVersions: []string{"v1"},
		}},
	}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: AdmissionTLSNamespace}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns, wh).Build()

	if err := PromoteAdmissionFailurePolicy(context.Background(), c, true); err != nil {
		t.Fatal(err)
	}
	var got admissionregistrationv1.ValidatingWebhookConfiguration
	if err := c.Get(context.Background(), types.NamespacedName{Name: AdmissionWebhookConfigName}, &got); err != nil {
		t.Fatal(err)
	}
	if got.Webhooks[0].FailurePolicy == nil || *got.Webhooks[0].FailurePolicy != admissionregistrationv1.Fail {
		t.Fatal("expected Fail")
	}
	if len(got.Webhooks[0].ClientConfig.CABundle) == 0 {
		t.Fatal("expected caBundle")
	}
	var sec corev1.Secret
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: AdmissionTLSNamespace, Name: AdmissionTLSSecretName}, &sec); err != nil {
		t.Fatal(err)
	}
}

func TestPromoteAdmissionFailurePolicyRollbackIgnore(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = admissionregistrationv1.AddToScheme(scheme)
	fail := admissionregistrationv1.Fail
	wh := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: AdmissionWebhookConfigName},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{{
			Name:          "access.olares.io",
			FailurePolicy: &fail,
			ClientConfig:  admissionregistrationv1.WebhookClientConfig{},
			SideEffects: func() *admissionregistrationv1.SideEffectClass {
				s := admissionregistrationv1.SideEffectClassNone
				return &s
			}(),
			AdmissionReviewVersions: []string{"v1"},
		}},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(wh).Build()
	if err := PromoteAdmissionFailurePolicy(context.Background(), c, false); err != nil {
		t.Fatal(err)
	}
	var got admissionregistrationv1.ValidatingWebhookConfiguration
	_ = c.Get(context.Background(), types.NamespacedName{Name: AdmissionWebhookConfigName}, &got)
	if got.Webhooks[0].FailurePolicy == nil || *got.Webhooks[0].FailurePolicy != admissionregistrationv1.Ignore {
		t.Fatal("expected Ignore when end-state off")
	}
}
