package notification

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ApplicationPermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ApplicationPermissionSpec `json:"spec,omitempty"`
}

type ApplicationPermissionSpec struct {
	Description string `json:"description,omitempty"`
	App         string `json:"app,omitempty"`
	Appid       string `json:"appid,omitempty"`
	Key         string `json:"key,omitempty"`
	Secret      string `json:"secret,omitempty"`
}

var AppPermGVR = schema.GroupVersionResource{
	Group:    "sys.bytetrade.io",
	Version:  "v1alpha1",
	Resource: "applicationpermissions",
}
