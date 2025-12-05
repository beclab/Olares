package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+genclient
//+genclient:nonNamespaced
//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster, shortName={term}, categories={all}
//+kubebuilder:printcolumn:JSONPath=.spec.name, name=version name, type=string
//+kubebuilder:printcolumn:JSONPath=.spec.version, name=version, type=string
//+kubebuilder:printcolumn:JSONPath=.status.state, name=state, type=string
//+kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp, name=age, type=date
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Terminus is the Schema for the terminuses API
type Terminus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerminusSpec   `json:"spec,omitempty"`
	Status TerminusStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TerminusList contains a list of Terminus
type TerminusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Terminus `json:"items"`
}

// TerminusStatus defines the observed state of Terminus
type TerminusStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// the state of the terminus: draft, submitted, passed, rejected, suspended, active
	State      string       `json:"state"`
	UpdateTime *metav1.Time `json:"updateTime,omitempty"`
	StatusTime *metav1.Time `json:"statusTime,omitempty"`
}

// TerminusSpec defines the desired state of Terminus
type TerminusSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// description from terminus
	Description string `json:"description,omitempty"`

	// the version name of the terminus os
	Name string `json:"name"`

	// the DisplayName of the terminus
	DisplayName string `json:"display,omitempty"`

	// the version of the terminus
	Version string `json:"version"`

	// the release server of the terminus
	ReleaseServer ReleaseServer `json:"releaseServer"`

	// the extend settings of the terminus
	Settings map[string]string `json:"settings,omitempty"`
}

// ReleaseServer defines the Terminus new version release server
type ReleaseServer struct {

	// serverType: github or others
	ServerType string `json:"serverType"`

	// github defines github repo where the terminus released
	Github GithubRepository `json:"github,omitempty"`
}

// GithubRepository defines github repo info
type GithubRepository struct {

	// github repository owner
	Owner string `json:"owner"`

	// github repository name
	Repo string `json:"repo"`
}

func init() {
	SchemeBuilder.Register(&Terminus{}, &TerminusList{})
}
