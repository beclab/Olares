package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SyncSpec defines the desired state of Sync
type SyncSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	LLdap *LLdapProvider `json:"lldap"`
}

//
//type Provider struct {
//	Name          string `json:"name"`
//	*ProviderType `json:",inline"`
//}
//
//type ProviderType struct {
//	LLdap *LdapProvider `json:"ldap"`
//}

type LLdapProvider struct {
	Name              string     `json:"name"`
	URL               string     `json:"url"`
	CredentialsSecret *ObjectRef `json:"credentialsSecret"`
	GroupWhitelist    []string   `json:"groupWhitelist,omitempty"`
	UserBlacklist     []string   `json:"userBlacklist,omitempty"`
}

type ObjectRefKind string

const (
	SecretObjectRefKind = "Secret"
)

type ObjectRef struct {
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Kind      ObjectRefKind `json:"kind"`
}

//+genclient
//+genclient:nonNamespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster, shortName={sync}, categories={all}
//+kubebuilder:printcolumn:JSONPath=.spec.name, name=sync name, type=string
//+kubebuilder:printcolumn:JSONPath=.spec.namespace, name=namespace, type=string
//+kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp, name=age, type=date
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Sync is the Schema for the sync API
type Sync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SyncSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SyncList contains a list of Sync
type SyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sync `json:"items"`
}
