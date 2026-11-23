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

func TestPromoteAdmissionFailurePolicyAlwaysFail(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = admissionregistrationv1.AddToScheme(scheme)

	ignore := admissionregistrationv1.Ignore
	side := admissionregistrationv1.SideEffectClassNone
	wh := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: AdmissionWebhookConfigName},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{
			{
				Name:                    "access.olares.io",
				FailurePolicy:           &ignore,
				ClientConfig:            admissionregistrationv1.WebhookClientConfig{},
				SideEffects:             &side,
				AdmissionReviewVersions: []string{"v1"},
			},
			{
				Name:                    "posture-config.access.olares.io",
				FailurePolicy:           &ignore,
				ClientConfig:            admissionregistrationv1.WebhookClientConfig{},
				SideEffects:             &side,
				AdmissionReviewVersions: []string{"v1"},
			},
		},
	}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: AdmissionTLSNamespace}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns, wh).Build()

	if err := PromoteAdmissionFailurePolicy(context.Background(), c); err != nil {
		t.Fatal(err)
	}
	var got admissionregistrationv1.ValidatingWebhookConfiguration
	if err := c.Get(context.Background(), types.NamespacedName{Name: AdmissionWebhookConfigName}, &got); err != nil {
		t.Fatal(err)
	}
	for i := range got.Webhooks {
		if got.Webhooks[i].FailurePolicy == nil || *got.Webhooks[i].FailurePolicy != admissionregistrationv1.Fail {
			t.Fatalf("webhook %d expected Fail", i)
		}
		if len(got.Webhooks[i].ClientConfig.CABundle) == 0 {
			t.Fatalf("webhook %d expected caBundle", i)
		}
	}
	var sec corev1.Secret
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: AdmissionTLSNamespace, Name: AdmissionTLSSecretName}, &sec); err != nil {
		t.Fatal(err)
	}
	if len(sec.Data["tls.crt"]) == 0 || len(sec.Data["tls.key"]) == 0 {
		t.Fatal("expected tls material")
	}
}
