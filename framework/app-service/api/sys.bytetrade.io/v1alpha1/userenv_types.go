package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Namespaced, shortName={userenv}, categories={all}
//+kubebuilder:printcolumn:JSONPath=.envName, name=env name, type=string
//+kubebuilder:printcolumn:JSONPath=.value, name=value, type=string
//+kubebuilder:printcolumn:JSONPath=.editable, name=editable, type=boolean
//+kubebuilder:printcolumn:JSONPath=.required, name=required, type=boolean
//+kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp, name=age, type=date

// UserEnv is the Schema for the user environment variables API
type UserEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	EnvVarSpec `json:",inline"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UserEnvList contains a list of UserEnv
type UserEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserEnv `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UserEnv{}, &UserEnvList{})
}
