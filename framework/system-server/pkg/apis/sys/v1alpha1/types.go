package v1alpha1

import (
	"reflect"
	"strings"

	"bytetrade.io/web3os/system-server/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Get    = "Get"
	List   = "List"
	Create = "Create"
	Update = "Update"
	Delete = "Delete"
	Watch  = "Watch"

	Provider = "provider"
	Watcher  = "watcher"

	Event     = "event"
	Calendar  = "calendar"
	Key       = "key"
	Contact   = "contact"
	LegacyAPI = "legacy_api"
	Token     = "token"
	Message   = "message"
	Intent    = "intent"

	Active    = "active"
	Suspended = "suspended"
)

var (
	WatcherSupportedOPs = []string{
		Create,
		Update,
		Delete,
	}
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProviderRegistry is the Schema for the ProviderRegistry API
type ProviderRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderRegistrySpec   `json:"spec,omitempty"`
	Status ProviderRegistryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ProviderRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ProviderRegistry `json:"items"`
}

type ProviderRegistryStatus struct {
	// the state of the application: draft, submitted, passed, rejected, suspended, active
	State      string       `json:"state"`
	UpdateTime *metav1.Time `json:"updateTime,omitempty"`
	StatusTime *metav1.Time `json:"statusTime,omitempty"`
}

type OpApisItem struct {
	Name string `json:"name,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type Callback struct {
	Op      string              `json:"op,omitempty"`
	URI     string              `json:"uri,omitempty"`
	Filters map[string][]string `json:"filters,omitempty"`
}

type With2FA struct {
	URI string `json:"uri,omitempty"`
}

type Permission struct {
	ACL     []string `json:"acl,omitempty"`
	With2FA With2FA  `json:"with2FA,omitempty"`
}

type ProviderRegistrySpec struct {
	Description string       `json:"description,omitempty"`
	Group       string       `json:"group"`
	Kind        string       `json:"kind"`
	DataType    string       `json:"dataType"`
	OpApis      []OpApisItem `json:"opApis,omitempty"`
	Callbacks   []Callback   `json:"callbacks,omitempty"`
	Permission  Permission   `json:"permission,omitempty"`
	Version     string       `json:"version"`
	Deployment  string       `json:"deployment"`
	Namespace   string       `json:"namespace"`
	Endpoint    string       `json:"endpoint"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApplicationPermission ProviderRegistry is the Schema for the ProviderRegistry API
type ApplicationPermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationPermissionSpec   `json:"spec,omitempty"`
	Status ApplicationPermissionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ApplicationPermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ApplicationPermission `json:"items"`
}

type ApplicationPermissionStatus struct {
	// the state of the application: draft, submitted, passed, rejected, suspended, active
	State      string       `json:"state"`
	UpdateTime *metav1.Time `json:"updateTime,omitempty"`
	StatusTime *metav1.Time `json:"statusTime,omitempty"`
}

type ApplicationPermissionSpec struct {
	Description string              `json:"description,omitempty"`
	App         string              `json:"app,omitempty"`
	Appid       string              `json:"appid,omitempty"`
	Key         string              `json:"key,omitempty"`
	Secret      string              `json:"secret,omitempty"`
	Permission  []PermissionRequire `json:"permissions,omitempty"`
}

type PermissionRequire struct {
	Group    string   `json:"group"`
	DataType string   `json:"dataType"`
	Version  string   `json:"version"`
	Ops      []string `json:"ops"`
	AppKey   string
}

type RequiredOp struct {
	Op     string
	Params map[string]string
}

func (p *PermissionRequire) CompareTo(other *PermissionRequire) bool {
	return p.Group == other.Group &&
		p.DataType == other.DataType &&
		p.Version == other.Version &&
		reflect.DeepEqual(utils.UniqAndSort(p.Ops), utils.UniqAndSort(other.Ops))
}

func (p *PermissionRequire) Include(other *PermissionRequire, fullMatch bool) bool {
	if p.Group == other.Group &&
		p.DataType == other.DataType &&
		p.Version == other.Version {
		for _, o := range other.Ops {
			op := o

			if !utils.ListContains(p.Ops, op) {
				if !fullMatch {
					requiredOp := DecodeOps(o)
					if !utils.ListContains(p.Ops, requiredOp.Op) {
						return false
					}
				} else {
					return false
				}
			}
		}

		return true
	}

	return false
}

func DecodeOps(op string) *RequiredOp {
	opTokenizer := strings.Split(op, "?")

	rop := &RequiredOp{
		Op:     opTokenizer[0],
		Params: make(map[string]string),
	}

	if len(opTokenizer) > 1 {
		params := strings.Split(opTokenizer[1], "&")
		for _, p := range params {
			if p != "" {
				kv := strings.Split(p, "=")
				v := ""
				if len(kv) > 1 {
					v = kv[1]
				}

				rop.Params[kv[0]] = v
			}
		}
	}

	return rop
}
