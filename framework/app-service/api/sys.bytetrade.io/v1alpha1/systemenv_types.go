package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+genclient
//+genclient:nonNamespaced
//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster, shortName={sysenv}, categories={all}
//+kubebuilder:printcolumn:JSONPath=.envName, name=env name, type=string
//+kubebuilder:printcolumn:JSONPath=.value, name=value, type=string
//+kubebuilder:printcolumn:JSONPath=.editable, name=editable, type=boolean
//+kubebuilder:printcolumn:JSONPath=.required, name=required, type=boolean
//+kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp, name=age, type=date

// SystemEnv is the Schema for the system environment variables API
type SystemEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	EnvVarSpec `json:",inline"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SystemEnvList contains a list of SystemEnv
type SystemEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SystemEnv `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SystemEnv{}, &SystemEnvList{})
}
