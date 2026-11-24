package access

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	admissionregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func catalogWebhookFixture() *admissionregv1.ValidatingWebhookConfiguration {
	return &admissionregv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: catalogConfig},
		Webhooks: []admissionregv1.ValidatingWebhook{{
			FailurePolicy: ptr.To(admissionregv1.Fail),
			ClientConfig: admissionregv1.WebhookClientConfig{Service: &admissionregv1.ServiceReference{
				Namespace: "os-framework", Name: "app-service", Port: ptr.To(int32(8433)), Path: ptr.To(CatalogWebhookPath),
			}},
			NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"kubernetes.io/metadata.name": PolicyNamespace}},
			Rules: []admissionregv1.RuleWithOperations{{
				Operations: []admissionregv1.OperationType{admissionregv1.Create, admissionregv1.Update},
				Rule:       admissionregv1.Rule{APIGroups: []string{"access.olares.io"}, APIVersions: []string{"v1alpha1"}, Resources: []string{"appcapabilitysets"}},
			}},
		}},
	}
}

func TestCatalogTrustChangesOnlyAfterAuthenticatedProbe(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	_ = admissionregv1.AddToScheme(scheme)
	hook := catalogWebhookFixture()
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(hook).Build()
	ca := []byte("verified serving CA")
	failed := func(context.Context, []byte) error { return errors.New("TLS endpoint unavailable") }
	if err := ensureCatalogWebhook(ctx, c, ca, failed); err == nil {
		t.Fatal("unreachable endpoint approved")
	}
	if err := c.Get(ctx, client.ObjectKey{Name: catalogConfig}, hook); err != nil {
		t.Fatal(err)
	}
	if len(hook.Webhooks[0].ClientConfig.CABundle) != 0 {
		t.Fatal("failed probe published a CA bundle")
	}
	verified := func(_ context.Context, bundle []byte) error {
		if !bytes.Equal(bundle, ca) {
			return errors.New("unexpected trust bundle")
		}
		return nil
	}
	if err := ensureCatalogWebhook(ctx, c, ca, verified); err != nil {
		t.Fatal(err)
	}
	if err := ensureCatalogWebhook(ctx, c, ca, verified); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(ctx, client.ObjectKey{Name: catalogConfig}, hook); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(hook.Webhooks[0].ClientConfig.CABundle, ca) || *hook.Webhooks[0].FailurePolicy != admissionregv1.Fail {
		t.Fatal("verified trust was not published with fail-closed admission")
	}
	hook.Webhooks[0].FailurePolicy = ptr.To(admissionregv1.Ignore)
	if err := c.Update(ctx, hook); err != nil {
		t.Fatal(err)
	}
	if err := ensureCatalogWebhook(ctx, c, ca, verified); err == nil {
		t.Fatal("fail-open webhook passed the publication gate")
	}
}

func TestCatalogWebhookObjectSelector(t *testing.T) {
	for _, tc := range []struct {
		name     string
		selector *metav1.LabelSelector
		allowed  bool
	}{
		{name: "omitted", allowed: true},
		{name: "API server default", selector: &metav1.LabelSelector{}, allowed: true},
		{name: "match labels", selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"validate": "true"},
		}},
		{name: "match expressions", selector: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "validate", Operator: metav1.LabelSelectorOpExists}},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			if err := admissionregv1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			hook := catalogWebhookFixture()
			hook.Webhooks[0].ObjectSelector = tc.selector
			oldCA := []byte("previous serving CA")
			newCA := []byte("replacement serving CA")
			hook.Webhooks[0].ClientConfig.CABundle = oldCA
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(hook).Build()

			probed := false
			err := ensureCatalogWebhook(ctx, c, newCA, func(context.Context, []byte) error {
				probed = true
				return nil
			})
			if (err == nil) != tc.allowed || probed != tc.allowed {
				t.Fatalf("unexpected result: probe called=%v, err=%v", probed, err)
			}
			if err := c.Get(ctx, client.ObjectKey{Name: catalogConfig}, hook); err != nil {
				t.Fatal(err)
			}
			expectedCA := oldCA
			if tc.allowed {
				expectedCA = newCA
			}
			if !bytes.Equal(hook.Webhooks[0].ClientConfig.CABundle, expectedCA) ||
				*hook.Webhooks[0].FailurePolicy != admissionregv1.Fail {
				t.Fatal("unexpected admission trust or failure policy")
			}
		})
	}
}

type probeTransport func(*http.Request) (*http.Response, error)

func (f probeTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCatalogProbeRequiresMatchingSuccessfulAdmissionResponse(t *testing.T) {
	for _, tc := range []struct {
		body string
		ok   bool
	}{
		{`{"response":{"uid":"app-access-readiness","allowed":true}}`, true},
		{`{"response":{"uid":"app-access-readiness","allowed":false}}`, false},
		{`{"response":{"uid":"another-review","allowed":true}}`, false},
		{`{}`, false},
	} {
		hc := &http.Client{Transport: probeTransport(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost {
				t.Fatal("probe did not exercise admission")
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(tc.body))}, nil
		})}
		if err := checkCatalogEndpoint(context.Background(), hc, "https://catalog.invalid"); (err == nil) != tc.ok {
			t.Fatalf("unexpected probe outcome for %s: %v", tc.body, err)
		}
	}
}
