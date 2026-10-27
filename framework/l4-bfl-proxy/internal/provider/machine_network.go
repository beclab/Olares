package provider

import (
	"context"
	"strings"

	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/beclab/l4-bfl-proxy/internal/tlsutil"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultDomainTLSName = "default-domain-tls"

// The pinned API client predates LocalName. An unstructured informer preserves
// both fields from the machine's Terminus rather than discarding unknown data.
func machineTerminus() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "sys.bytetrade.io", Version: "v1alpha1", Kind: "Terminus"})
	return obj
}

func (p *Provider) readMachineNetwork(ctx context.Context, res *message.Resources) error {
	res.LocalZone = "olares.local"
	terminus := machineTerminus()
	if err := p.cache.Get(ctx, client.ObjectKey{Name: "terminus"}, terminus); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	} else {
		domain, _, err := unstructured.NestedString(terminus.Object, "spec", "settings", "domainName")
		if err != nil {
			return err
		}
		res.DefaultDomain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
		localName, _, err := unstructured.NestedString(terminus.Object, "spec", "localName")
		if err != nil {
			return err
		}
		if localName != "" {
			res.LocalZone = localName + ".olares.local"
		}
	}
	if res.DefaultDomain == "" {
		return nil
	}
	var secret corev1.Secret
	if err := p.cache.Get(ctx, client.ObjectKey{Namespace: "os-network", Name: defaultDomainTLSName}, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	cert, key := string(secret.Data[corev1.TLSCertKey]), string(secret.Data[corev1.TLSPrivateKeyKey])
	if err := tlsutil.ValidateKeyPair(cert, key); err != nil {
		klog.Warningf("provider: default domain TLS pair is not ready: %v", err)
		return nil
	}
	res.SSL = &message.SSLConfig{Zone: res.DefaultDomain, CertData: cert, KeyData: key}
	return nil
}
