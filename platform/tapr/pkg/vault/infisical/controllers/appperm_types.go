package controllers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProviderRegistry is the Schema for the ProviderRegistry API
type ApplicationPermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationPermissionSpec   `json:"spec,omitempty"`
	Status ApplicationPermissionStatus `json:"status,omitempty"`
}

type ApplicationPermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ApplicationPermission `json:"items"`
}

type ApplicationPermissionStatus struct {
	// the state of the application: draft, submitted, passed, rejected, suspended, active
	State      string       `json:"state"`
	UpdateTime *metav1.Time `json:"updateTime,omitempty"`
	StatusTime *metav1.Time `json:"statusTime,omitempty"`
}

type ApplicationPermissionSpec struct {
	Description string              `json:"description,omitempty"`
	App         string              `json:"app,omitempty"`
	Appid       string              `json:"appid,omitempty"`
	Key         string              `json:"key,omitempty"`
	Secret      string              `json:"secret,omitempty"`
	Permission  []PermissionRequire `json:"permissions,omitempty"`
}

type PermissionRequire struct {
	Group    string   `json:"group"`
	DataType string   `json:"dataType"`
	Version  string   `json:"version"`
	Ops      []string `json:"ops"`
}
