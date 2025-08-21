package utils

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	UserSchemeGroupVersion = schema.GroupVersion{Group: "iam.kubesphere.io", Version: "v1alpha2"}

	UserGVR = schema.GroupVersionResource{
		Group:    UserSchemeGroupVersion.Group,
		Version:  UserSchemeGroupVersion.Version,
		Resource: "users",
	}
)

type NodePressure struct {
	Type    corev1.NodeConditionType `json:"type"`
	Message string                   `json:"message"`
}
