package users

import (
	"context"

	"bytetrade.io/web3os/bfl/pkg/client/dynamic_client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "sys.bytetrade.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}
)

var terminusGvr = schema.GroupVersionResource{
	Group:    GroupVersion.Group,
	Version:  GroupVersion.Version,
	Resource: "terminus",
}

// Terminus is the Schema for the terminuses API
type Terminus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerminusSpec   `json:"spec,omitempty"`
	Status TerminusStatus `json:"status,omitempty"`
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

const SettingsDomainNameKey = "domainName"
const SettingsSelfhostedKey = "selfhosted"
const SettingsTerminusdKey = "terminusd"

type ResourceTerminusClient struct {
	c *dynamic_client.ResourceClient[Terminus]
}

func NewResourceTerminusClient() (*ResourceTerminusClient, error) {
	ri, err := dynamic_client.NewResourceClient[Terminus](terminusGvr)
	if err != nil {
		return nil, err
	}
	return &ResourceTerminusClient{c: ri}, nil
}

func (u *ResourceTerminusClient) Get(ctx context.Context, name string, options metav1.GetOptions) (*Terminus, error) {
	return u.c.Get(ctx, name, options)
}
