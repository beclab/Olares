package watchers

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

var clusterConfigGVR = schema.GroupVersionResource{
	Group:    "cluster.olares.io",
	Version:  "v1alpha1",
	Resource: "clusterconfigs",
}

// inClusterGatewayEnabled reads ClusterConfig.spec.inClusterGatewayEnabled.
// Absent CR, field, or API errors default to true (legacy clusters keep Shared DNS).
func inClusterGatewayEnabled(ctx context.Context, dynamicClient dynamic.Interface) bool {
	if dynamicClient == nil {
		return true
	}
	u, err := dynamicClient.Resource(clusterConfigGVR).Get(ctx, "cluster", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return true
		}
		klog.V(2).Infof("clusterconfig get failed, default inClusterGatewayEnabled=true: %v", err)
		return true
	}
	spec, found, err := unstructured.NestedMap(u.Object, "spec")
	if err != nil || !found {
		return true
	}
	v, ok := spec["inClusterGatewayEnabled"]
	if !ok {
		return true
	}
	b, ok := v.(bool)
	if !ok {
		klog.Warningf("ClusterConfig.spec.inClusterGatewayEnabled non-bool, default true")
		return true
	}
	return b
}
