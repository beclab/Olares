package v1alpha1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// the entrance of the application
	Index string `json:"index,omitempty"`

	// description from app's description or frontend
	Description string `json:"description,omitempty"`

	// The url of the icon
	Icon string `json:"icon,omitempty"`

	// the name of the application
	Name string `json:"name"`

	// the unique id of the application
	// for sys application appid equal name otherwise appid equal md5(name)[:8]
	Appid string `json:"appid"`

	// application is system app
	IsSysApp bool `json:"isSysApp"`

	// the namespace of the application
	Namespace string `json:"namespace,omitempty"`

	// the deployment of the application
	DeploymentName string `json:"deployment,omitempty"`

	// the owner of the application
	Owner string `json:"owner,omitempty"`

	// the service address of the application
	Entrances []Entrance `json:"entrances,omitempty"`

	Ports []ServicePort `json:"ports,omitempty"`

	// the extend settings of the application
	Settings map[string]string `json:"settings,omitempty"`
}

type Entrance struct {
	Name            string `yaml:"name" json:"name"`
	Host            string `yaml:"host" json:"host"`
	Port            int32  `yaml:"port" json:"port"`
	Icon            string `yaml:"icon" json:"icon,omitempty"`
	Title           string `yaml:"title" json:"title"`
	AuthLevel       string `yaml:"authLevel" json:"authLevel,omitempty"`
	Invisible       bool   `yaml:"invisible,omitempty" json:"invisible,omitempty"`
	WindowPushState bool   `yaml:"windowPushState,omitempty" json:"windowPushState,omitempty"`
}

type ServicePort struct {
	Name string `json:"name" yaml:"name"`
	Host string `yaml:"host" json:"host"`
	Port int32  `yaml:"port" json:"port"`

	ExposePort int32 `yaml:"exposePort" json:"exposePort,omitempty"`

	// The protocol for this entrance. Supports "tcp" and "udp".
	// Default is tcp.
	// +default="udp"
	// +optional
	Protocol string `yaml:"protocol" json:"protocol,omitempty"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// the state of the application: draft, submitted, passed, rejected, suspended, active
	State      string       `json:"state,omitempty"`
	UpdateTime *metav1.Time `json:"updateTime"`
	StatusTime *metav1.Time `json:"statusTime"`
}

//+genclient
//+genclient:nonNamespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster, shortName={app}, categories={all}
//+kubebuilder:printcolumn:JSONPath=.spec.name, name=application name, type=string
//+kubebuilder:printcolumn:JSONPath=.spec.namespace, name=namespace, type=string
//+kubebuilder:printcolumn:JSONPath=.status.state, name=state, type=string
//+kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp, name=age, type=date
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

var (
	AppEntranceCertConfigMapLabel   = fmt.Sprintf("%s/%s", GroupVersion.Group, "custom-domain-cert")
	AppEntranceCertConfigMapNameTpl = "%s-ssl-config"
	AppEntranceCertConfigMapCertKey = "cert"
	AppEntranceCertConfigMapKeyKey  = "key"
	AppEntranceCertConfigMapZoneKey = "zone"
)

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
