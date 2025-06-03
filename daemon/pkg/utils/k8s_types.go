package utils

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	UserSchemeGroupVersion = schema.GroupVersion{Group: "iam.kubesphere.io", Version: "v1alpha2"}

	UserGVR = schema.GroupVersionResource{
		Group:    UserSchemeGroupVersion.Group,
		Version:  UserSchemeGroupVersion.Version,
		Resource: "users",
	}
)
