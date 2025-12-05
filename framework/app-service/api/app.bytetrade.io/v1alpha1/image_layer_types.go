package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+genclient
//+genclient:nonNamespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster, shortName={appimage}, categories={all}
//+kubebuilder:printcolumn:JSONPath=.spec.appName, name=application name, type=string
//+kubebuilder:printcolumn:JSONPath=.status.state, name=state, type=string
//+kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp, name=age, type=date
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppImage is the Schema for the image managers API
type AppImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageSpec   `json:"spec,omitempty"`
	Status ImageStatus `json:"status,omitempty"`
}

type ImageSpec struct {
	AppName string   `json:"appName"`
	Nodes   []string `json:"nodes"`
	Refs    []string `json:"refs"`
}

type ImageStatus struct {
	// processing, completed, failed

	State      string       `json:"state"`
	Images     []ImageInfo  `json:"images,omitempty"`
	StatueTime *metav1.Time `json:"statueTime"`
	Message    string       `json:"message,omitempty"`
	Conditions []Condition  `json:"conditions,omitempty"`
}

type Condition struct {
	Node      string `json:"node"`
	Completed bool   `json:"completed"`
}

type ImageInfo struct {
	Node         string       `json:"node"`
	Name         string       `json:"name"`
	Architecture string       `json:"architecture,omitempty"`
	Variant      string       `json:"variant,omitempty"`
	Os           string       `json:"os,omitempty"`
	LayersData   []ImageLayer `json:"layersData"`
}

type ImageLayer struct {
	MediaType   string            `json:"mediaType"`
	Digest      string            `json:"digest"`
	Offset      int64             `json:"offset"`
	Size        int64             `json:"size"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

//+kubebuilder:object:root=true
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppImageList contains a list of AppImage
type AppImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppImage{}, &AppImageList{})
}
