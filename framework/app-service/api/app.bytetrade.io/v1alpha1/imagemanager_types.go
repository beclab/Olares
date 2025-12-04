package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+genclient
//+genclient:nonNamespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster, shortName={im}, categories={all}
//+kubebuilder:printcolumn:JSONPath=.spec.appName, name=application name, type=string
//+kubebuilder:printcolumn:JSONPath=.spec.appNamespace, name=namespace, type=string
//+kubebuilder:printcolumn:JSONPath=.status.state, name=state, type=string
//+kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp, name=age, type=date
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ImageManager is the Schema for the image managers API
type ImageManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageManagerSpec   `json:"spec,omitempty"`
	Status ImageManagerStatus `json:"status,omitempty"`
}

// ImageManagerStatus defines the observed state of ApplicationManager
type ImageManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions map[string]map[string]map[string]string `json:"conditions,omitempty"`
	Message    string                                  `json:"message,omitempty"`
	State      string                                  `json:"state"`
	UpdateTime *metav1.Time                            `json:"updateTime"`
	StatusTime *metav1.Time                            `json:"statusTime"`
}

// ImageManagerSpec defines the desired state of ImageManager
type ImageManagerSpec struct {
	AppName      string   `json:"appName"`
	AppNamespace string   `json:"appNamespace,omitempty"`
	AppOwner     string   `json:"appOwner,omitempty"`
	Refs         []Ref    `json:"refs"`
	Nodes        []string `json:"nodes"`
}

type Ref struct {
	Name            string            `json:"name"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
}

type ImageProgress struct {
	NodeName string `json:"nodeName"`
	ImageRef string `json:"imageRef"`
	Progress string `json:"progress"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ImageManagerList contains a list of ApplicationManager
type ImageManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImageManager{}, &ImageManagerList{})
}
