package routecontrol

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const labelTLSCustomReplica = "gateway.olares.io/tls-custom-replica"

var customTLSSyncTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "olares_mesh_in_custom_tls_sync_total",
		Help: "mesh-in custom-domain TLS aggregate sync by result",
	},
	[]string{"result"},
)

func init() { prometheus.MustRegister(customTLSSyncTotal) }

type customTLSMaterial struct {
	Cert, Key []byte
	Hash      string
}

// syncMeshInCustomTLSReplica projects CustomDomainTLS into callerNS Secret
// olares-mesh-in-custom-tls (keys <fqdn>.crt/.key). In-place Update so kubelet
// refreshes files without restarting mesh-in; viewer cert volume is untouched.
func syncMeshInCustomTLSReplica(ctx context.Context, c client.Client, callerNS string, tlsHosts []string) error {
	if c == nil || callerNS == "" {
		return nil
	}
	material, err := loadCustomDomainTLSMaterial(ctx, c)
	if err != nil {
		return customTLSErr(callerNS, "list source", err)
	}

	data, aggHash := buildCustomTLSData(tlsHosts, material)
	if len(data) == 0 {
		return deleteMeshInCustomTLSReplica(ctx, c, callerNS)
	}

	nsObj := &corev1.Namespace{}
	if err := c.Get(ctx, types.NamespacedName{Name: callerNS}, nsObj); err != nil {
		return customTLSErr(callerNS, "get namespace", err)
	}

	dst := &corev1.Secret{}
	err = c.Get(ctx, types.NamespacedName{Namespace: callerNS, Name: constants.MeshInCustomTLSSecretName}, dst)
	switch {
	case apierrors.IsNotFound(err):
		if err := c.Create(ctx, desiredMeshInCustomTLSSecret(callerNS, nsObj, data, aggHash)); err != nil {
			return customTLSErr(callerNS, "create", err)
		}
		return customTLSOK(callerNS, "created", len(data)/2)
	case err != nil:
		return customTLSErr(callerNS, "get", err)
	case dst.Annotations != nil && dst.Annotations[annotationTLSContentHash] == aggHash:
		customTLSSyncTotal.WithLabelValues("noop").Inc()
		return nil
	}

	if dst.Labels == nil {
		dst.Labels = map[string]string{}
	}
	dst.Labels[ManagedByLabel] = ManagedByValue
	dst.Labels[labelTLSCustomReplica] = "true"
	if dst.Annotations == nil {
		dst.Annotations = map[string]string{}
	}
	dst.Annotations[annotationTLSContentHash] = aggHash
	dst.Type = corev1.SecretTypeOpaque
	dst.StringData = nil
	dst.Data = data
	ensureNamespaceOwnerRef(dst, nsObj)
	if err := c.Update(ctx, dst); err != nil {
		return customTLSErr(callerNS, "update", err)
	}
	return customTLSOK(callerNS, "updated", len(data)/2)
}

func deleteMeshInCustomTLSReplica(ctx context.Context, c client.Client, callerNS string) error {
	if c == nil || callerNS == "" {
		return nil
	}
	sec := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{Namespace: callerNS, Name: constants.MeshInCustomTLSSecretName}, sec)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return customTLSErr(callerNS, "get for delete", err)
	}
	if sec.Labels[ManagedByLabel] != ManagedByValue && sec.Labels[labelTLSCustomReplica] != "true" {
		return nil
	}
	if err := c.Delete(ctx, sec); err != nil && !apierrors.IsNotFound(err) {
		return customTLSErr(callerNS, "delete", err)
	}
	return customTLSOK(callerNS, "deleted", 0)
}

func buildCustomTLSData(tlsHosts []string, material map[string]customTLSMaterial) (map[string][]byte, string) {
	seen := map[string]struct{}{}
	var hosts, parts []string
	for _, h := range tlsHosts {
		h = strings.ToLower(strings.TrimSpace(h))
		if h == "" {
			continue
		}
		m, ok := material[h]
		if !ok {
			continue
		}
		if _, dup := seen[h]; dup {
			continue
		}
		seen[h] = struct{}{}
		hosts = append(hosts, h)
		parts = append(parts, h+"="+m.Hash)
	}
	sort.Strings(hosts)
	sort.Strings(parts)
	if len(hosts) == 0 {
		return nil, ""
	}
	data := make(map[string][]byte, len(hosts)*2)
	for _, h := range hosts {
		m := material[h]
		data[h+".crt"] = append([]byte(nil), m.Cert...)
		data[h+".key"] = append([]byte(nil), m.Key...)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return data, hex.EncodeToString(sum[:])
}

func loadCustomDomainTLSMaterial(ctx context.Context, c client.Client) (map[string]customTLSMaterial, error) {
	var list corev1.SecretList
	if err := c.List(ctx, &list, client.InNamespace(defaultGatewayNS), client.MatchingLabels{
		ManagedByLabel: ManagedByValue,
	}); err != nil {
		return nil, err
	}
	out := make(map[string]customTLSMaterial)
	for i := range list.Items {
		sec := &list.Items[i]
		if !strings.HasPrefix(sec.Name, customDomainTLSPrefix) && strings.TrimSpace(sec.Labels[labelTLSCustomDomain]) == "" {
			continue
		}
		domain := strings.ToLower(strings.TrimSpace(sec.Labels[labelTLSCustomDomain]))
		if domain == "" {
			domain = strings.ToLower(strings.TrimSpace(sec.Annotations[annotationTLSHostname]))
		}
		if domain == "" {
			continue
		}
		cert, key := secretTLSBytes(sec)
		if len(cert) == 0 || len(key) == 0 {
			continue
		}
		hash := strings.TrimSpace(sec.Annotations[annotationTLSContentHash])
		if hash == "" {
			hash = tlsMaterialHash(string(cert), string(key))
		}
		out[domain] = customTLSMaterial{Cert: cert, Key: key, Hash: hash}
	}
	return out, nil
}

func secretTLSBytes(sec *corev1.Secret) (cert, key []byte) {
	if sec == nil {
		return nil, nil
	}
	cert, key = sec.Data[corev1.TLSCertKey], sec.Data[corev1.TLSPrivateKeyKey]
	if len(cert) == 0 && sec.StringData != nil {
		cert = []byte(sec.StringData[corev1.TLSCertKey])
	}
	if len(key) == 0 && sec.StringData != nil {
		key = []byte(sec.StringData[corev1.TLSPrivateKeyKey])
	}
	return cert, key
}

func desiredMeshInCustomTLSSecret(callerNS string, ns *corev1.Namespace, data map[string][]byte, aggHash string) *corev1.Secret {
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.MeshInCustomTLSSecretName,
			Namespace: callerNS,
			Labels: map[string]string{
				ManagedByLabel:        ManagedByValue,
				labelTLSCustomReplica: "true",
			},
			Annotations: map[string]string{annotationTLSContentHash: aggHash},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}
	ensureNamespaceOwnerRef(sec, ns)
	return sec
}

func ensureNamespaceOwnerRef(sec *corev1.Secret, ns *corev1.Namespace) {
	if sec == nil || ns == nil || ns.UID == "" {
		return
	}
	for _, o := range sec.OwnerReferences {
		if o.UID == ns.UID {
			return
		}
	}
	f := false
	sec.OwnerReferences = append(sec.OwnerReferences, metav1.OwnerReference{
		APIVersion: "v1", Kind: "Namespace", Name: ns.Name, UID: ns.UID,
		Controller: &f, BlockOwnerDeletion: &f,
	})
}

func collectTLSHosts(targets []SharedHostsTarget) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, t := range targets {
		for _, h := range t.TLSHosts {
			h = strings.ToLower(strings.TrimSpace(h))
			if h == "" {
				continue
			}
			if _, ok := seen[h]; ok {
				continue
			}
			seen[h] = struct{}{}
			out = append(out, h)
		}
	}
	sort.Strings(out)
	return out
}

func customTLSErr(ns, op string, err error) error {
	customTLSSyncTotal.WithLabelValues("error").Inc()
	klog.Errorf("mesh-in-custom-tls: %s failed ns=%s: %v", op, hashCallerNS(ns), err)
	return err
}

func customTLSOK(ns, result string, domains int) error {
	customTLSSyncTotal.WithLabelValues(result).Inc()
	klog.Infof("mesh-in-custom-tls: sync ns=%s domains=%d result=%s", hashCallerNS(ns), domains, result)
	return nil
}
