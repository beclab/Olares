package callerjwt

import (
	"context"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

const (
	JWKSCAConfigMapName     = "caller-jwt-jwks-ca"
	JWKSCAConfigMapDataKey  = "ca.crt"
	JWKSBackendTLSPolicyName = "caller-jwt-jwks"
	jwksServicePortName     = "https"
)

var backendTLSPolicyGVK = schema.GroupVersionKind{
	Group:   "gateway.networking.k8s.io",
	Version: "v1",
	Kind:    "BackendTLSPolicy",
}

// reconcileJWKSTrust publishes the JWKS TLS CA and BackendTLSPolicy so Envoy
// Gateway can verify HTTPS to authz.olares.system when fetching JWKS.
func (r *IssuerReconciler) reconcileJWKSTrust(ctx context.Context) error {
	if err := r.reconcileJWKSCAConfigMap(ctx); err != nil {
		return err
	}
	return r.reconcileJWKSBackendTLSPolicy(ctx)
}

func (r *IssuerReconciler) reconcileJWKSCAConfigMap(ctx context.Context) error {
	if r == nil || r.Client == nil {
		return nil
	}
	certFile, _ := tlsCertPaths()
	caPath := "/etc/certs/ca.crt"
	if d := dirOf(certFile); d != "" {
		caPath = d + "/ca.crt"
	}

	pemBytes, err := os.ReadFile(caPath)
	if err != nil {
		if os.IsNotExist(err) {
			klog.V(4).Infof("callerjwt: JWKS CA file %s not found, skip ConfigMap", caPath)
			return nil
		}
		klog.Errorf("callerjwt: read JWKS CA %s failed: %v", caPath, err)
		return fmt.Errorf("read JWKS CA: %w", err)
	}
	if len(pemBytes) == 0 {
		klog.Warningf("callerjwt: JWKS CA file %s is empty, skip ConfigMap", caPath)
		return nil
	}

	desired := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      JWKSCAConfigMapName,
			Namespace: JWKSServiceNamespace,
			Labels: map[string]string{
				managedByLabel:          managedByValue,
				managedByComponentLabel: JWKSIngressNPComponentValue,
			},
		},
		Data: map[string]string{
			JWKSCAConfigMapDataKey: string(pemBytes),
		},
	}

	current := &corev1.ConfigMap{}
	key := types.NamespacedName{Name: JWKSCAConfigMapName, Namespace: JWKSServiceNamespace}
	err = r.Client.Get(ctx, key, current)
	switch {
	case apierrors.IsNotFound(err):
		if err := r.Client.Create(ctx, desired); err != nil {
			klog.Errorf("callerjwt: create JWKS CA ConfigMap failed: %v", err)
			return fmt.Errorf("create JWKS CA ConfigMap: %w", err)
		}
		return nil
	case err != nil:
		klog.Errorf("callerjwt: get JWKS CA ConfigMap failed: %v", err)
		return fmt.Errorf("get JWKS CA ConfigMap: %w", err)
	default:
		current.Data = desired.Data
		if current.Labels == nil {
			current.Labels = map[string]string{}
		}
		for k, v := range desired.Labels {
			current.Labels[k] = v
		}
		if err := r.Client.Update(ctx, current); err != nil {
			klog.Errorf("callerjwt: update JWKS CA ConfigMap failed: %v", err)
			return fmt.Errorf("update JWKS CA ConfigMap: %w", err)
		}
		return nil
	}
}

func (r *IssuerReconciler) reconcileJWKSBackendTLSPolicy(ctx context.Context) error {
	if r == nil || r.Client == nil {
		return nil
	}
	desired := desiredJWKSBackendTLSPolicy()

	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(backendTLSPolicyGVK)
	key := types.NamespacedName{Name: JWKSBackendTLSPolicyName, Namespace: JWKSServiceNamespace}
	err := r.Client.Get(ctx, key, current)
	switch {
	case apierrors.IsNotFound(err):
		if err := r.Client.Create(ctx, desired); err != nil {
			klog.Errorf("callerjwt: create JWKS BackendTLSPolicy failed: %v", err)
			return fmt.Errorf("create JWKS BackendTLSPolicy: %w", err)
		}
		return nil
	case err != nil:
		klog.Errorf("callerjwt: get JWKS BackendTLSPolicy failed: %v", err)
		return fmt.Errorf("get JWKS BackendTLSPolicy: %w", err)
	default:
		desired.SetResourceVersion(current.GetResourceVersion())
		if err := r.Client.Update(ctx, desired); err != nil {
			klog.Errorf("callerjwt: update JWKS BackendTLSPolicy failed: %v", err)
			return fmt.Errorf("update JWKS BackendTLSPolicy: %w", err)
		}
		return nil
	}
}

func desiredJWKSBackendTLSPolicy() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "BackendTLSPolicy",
		"metadata": map[string]any{
			"name":      JWKSBackendTLSPolicyName,
			"namespace": JWKSServiceNamespace,
			"labels": map[string]any{
				managedByLabel:          managedByValue,
				managedByComponentLabel: JWKSIngressNPComponentValue,
			},
		},
		"spec": map[string]any{
			"targetRefs": []any{
				map[string]any{
					"group":       "",
					"kind":        "Service",
					"name":        JWKSServiceName,
					"sectionName": jwksServicePortName,
				},
			},
			"validation": map[string]any{
				"hostname": IssuerHost,
				"caCertificateRefs": []any{
					map[string]any{
						"group": "",
						"kind":  "ConfigMap",
						"name":  JWKSCAConfigMapName,
					},
				},
			},
		},
	}}
	obj.SetGroupVersionKind(backendTLSPolicyGVK)
	return obj
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return ""
}
