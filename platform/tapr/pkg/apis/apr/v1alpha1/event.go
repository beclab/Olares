package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=".spec.type",description="Action type, publisher / subscriber"
// +kubebuilder:printcolumn:name="Callback",type=date,JSONPath=".spec.callback",description="The callback url of ubscriber"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=".status.state",description="Status"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp",description="Created time"
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced, shortName={ser}, categories={all}
// SysEventRegistry is the Schema for the Sys Event publisher and subscriber
type SysEventRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SysEventRegistrySpec   `json:"spec,omitempty"`
	Status SysEventRegistryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SysEventRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SysEventRegistry `json:"items"`
}

type SysEventRegistrySpec struct {
	Type     ActionType `json:"type"`
	Event    EventType  `json:"event"`
	Callback string     `json:"callback"`
}

type ActionType string

const (
	Subscriber ActionType = "subscriber"
)

type EventType string

const (
	UserCreate         EventType = "user.create"
	UserDelete         EventType = "user.delete"
	UserActive         EventType = "user.active"
	UserLogin          EventType = "user.login"
	AppInstall         EventType = "app.install"
	AppUninstall       EventType = "app.uninstall"
	AppSuspend         EventType = "app.suspend"
	MemoryHigh         EventType = "metrics.memory.high"
	CPUHigh            EventType = "metrics.cpu.high"
	UserMemoryHigh     EventType = "metrics.user.memory.high"
	UserCPUHigh        EventType = "metrics.user.cpu.high"
	RecommendInstall   EventType = "recommend.install"
	RecommendUninstall EventType = "recommend.uninstall"
)

// SysEventRegistryStatus defines the observed state of SysEventRegistry
type SysEventRegistryStatus struct {
	// the state of the application: draft, submitted, passed, rejected, suspended, active
	State      string       `json:"state,omitempty"`
	UpdateTime *metav1.Time `json:"updateTime"`
	StatusTime *metav1.Time `json:"statusTime"`
}
