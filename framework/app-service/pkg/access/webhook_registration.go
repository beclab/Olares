package access

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ensureCatalogWebhook(ctx context.Context, c client.Client, ca []byte, probe func(context.Context, []byte) error) error {
	hook, err := catalogWebhook(ctx, c)
	if err != nil {
		return err
	}
	if err := probe(ctx, ca); err != nil {
		return err
	}
	if bytes.Equal(hook.Webhooks[0].ClientConfig.CABundle, ca) {
		return nil
	}
	original := hook.DeepCopy()
	hook.Webhooks[0].ClientConfig.CABundle = ca
	return c.Patch(ctx, hook, client.MergeFromWithOptions(original, client.MergeFromWithOptimisticLock{}))
}

func catalogWebhook(ctx context.Context, c client.Client) (*admissionregv1.ValidatingWebhookConfiguration, error) {
	var hook admissionregv1.ValidatingWebhookConfiguration
	if err := c.Get(ctx, client.ObjectKey{Name: catalogConfig}, &hook); err != nil {
		return nil, err
	}
	if len(hook.Webhooks) != 1 {
		return nil, fmt.Errorf("expected one catalog webhook")
	}
	w := hook.Webhooks[0]
	s := w.ClientConfig.Service
	if w.FailurePolicy == nil || *w.FailurePolicy != admissionregv1.Fail ||
		s == nil || s.Name != "app-service" || s.Namespace != "os-framework" ||
		s.Port == nil || *s.Port != 8433 || s.Path == nil || *s.Path != CatalogWebhookPath ||
		w.ClientConfig.URL != nil || w.NamespaceSelector == nil ||
		w.NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != PolicyNamespace ||
		len(w.NamespaceSelector.MatchLabels) != 1 || len(w.NamespaceSelector.MatchExpressions) != 0 ||
		len(w.MatchConditions) != 0 || w.ObjectSelector != nil ||
		len(w.Rules) != 1 {
		return nil, fmt.Errorf("catalog webhook must use the fail-closed app-service endpoint")
	}
	r := w.Rules[0]
	if !slices.Equal(r.APIGroups, []string{"access.olares.io"}) || !slices.Equal(r.APIVersions, []string{"v1alpha1"}) ||
		!slices.Equal(r.Resources, []string{"appcapabilitysets"}) || len(r.Operations) != 2 ||
		!slices.Contains(r.Operations, admissionregv1.Create) || !slices.Contains(r.Operations, admissionregv1.Update) {
		return nil, fmt.Errorf("catalog webhook rules must cover catalog writes and leave revocation available")
	}
	return &hook, nil
}

// This probe exercises the same Service, TLS trust and handler that the API
// server uses. It sends an AdmissionReview without creating any API object.
func probeCatalogWebhook(ctx context.Context, ca []byte) error {
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(ca) {
		return fmt.Errorf("invalid app-service webhook CA")
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}}
	defer transport.CloseIdleConnections()
	httpClient := &http.Client{Transport: transport, Timeout: 3 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	return checkCatalogEndpoint(ctx, httpClient, "https://app-service.os-framework.svc:8433"+CatalogWebhookPath)
}

func checkCatalogEndpoint(ctx context.Context, httpClient *http.Client, endpoint string) error {
	review := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{Kind: "AdmissionReview", APIVersion: "admission.k8s.io/v1"},
		Request: &admissionv1.AdmissionRequest{
			UID: "app-access-readiness", Operation: admissionv1.Create, Namespace: PolicyNamespace,
			Resource: metav1.GroupVersionResource{Group: "access.olares.io", Version: "v1alpha1", Resource: "appcapabilitysets"},
			Object:   runtime.RawExtension{Raw: []byte(`{"spec":{"capabilities":[{"rules":[{"method":"GET","path":"/readiness"}]}]}}`)},
		},
	}
	body, err := json.Marshal(review)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var answer admissionv1.AdmissionReview
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&answer) != nil ||
		answer.Response == nil || !answer.Response.Allowed || answer.Response.UID != review.Request.UID {
		return fmt.Errorf("catalog admission did not validate the readiness probe")
	}
	return nil
}
