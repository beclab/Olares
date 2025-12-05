package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Namespaced, shortName={appenv}, categories={all}
//+kubebuilder:printcolumn:JSONPath=.appName, name=app name, type=string
//+kubebuilder:printcolumn:JSONPath=.appOwner, name=owner, type=string
//+kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp, name=age, type=date

// AppEnv is the Schema for the application environment variables API
type AppEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	AppName   string      `json:"appName" yaml:"appName" validate:"required"`
	AppOwner  string      `json:"appOwner" yaml:"appOwner" validate:"required"`
	Envs      []AppEnvVar `json:"envs,omitempty" yaml:"envs,omitempty"`
	NeedApply bool        `json:"needApply,omitempty" yaml:"needApply,omitempty"`
}

type AppEnvVar struct {
	EnvVarSpec    `json:",inline" yaml:",inline"`
	ApplyOnChange bool       `json:"applyOnChange,omitempty" yaml:"applyOnChange,omitempty"`
	ValueFrom     *ValueFrom `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

// ValueFrom defines a reference to an environment variable (UserEnv or SystemEnv)
type ValueFrom struct {
	EnvName string `json:"envName" validate:"required"`
	Status  string `json:"status,omitempty"`
}

type EnvValueOptionItem struct {
	Title string `json:"title" yaml:"title"`
	Value string `json:"value" yaml:"value"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppEnvList contains a list of AppEnv
type AppEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppEnv `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppEnv{}, &AppEnvList{})
}
