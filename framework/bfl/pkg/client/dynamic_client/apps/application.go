package apps

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "app.bytetrade.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}
)

var ApplicationGvr = schema.GroupVersionResource{
	Group:    GroupVersion.Group,
	Version:  GroupVersion.Version,
	Resource: "applications",
}

// ApplicationSpec
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

	IsSysApp bool `json:"isSysApp"`

	// the namespace of the application
	Namespace string `json:"namespace,omitempty"`

	// the deployment of the application
	DeploymentName string `json:"deployment,omitempty"`

	// the owner of the application
	Owner string `json:"owner,omitempty"`

	//Entrances []Entrance `json:"entrances,omitempty"`
	Entrances []Entrance `json:"entrances,omitempty"`

	Ports         []ServicePort `json:"ports,omitempty"`
	TailScale     TailScale     `json:"tailscale,omitempty"`
	TailScaleACLs []ACL         `json:"tailscaleAcls,omitempty"`

	// the extend settings of the application
	Settings map[string]string `json:"settings,omitempty"`
}

type ACL struct {
	Action string   `json:"action,omitempty"`
	Src    []string `json:"src,omitempty"`
	Proto  string   `json:"proto"`
	Dst    []string `json:"dst"`
}

type TailScale struct {
	ACLs      []ACL    `json:"acls,omitempty"`
	SubRoutes []string `json:"subRoutes,omitempty"`
}

type EntranceState string

const (
	EntranceRunning EntranceState = "running"
	EntranceCrash   EntranceState = "crash"
	EntranceSuspend EntranceState = "suspend"
)

func (e EntranceState) String() string {
	return string(e)
}

// Entrance contains details for application entrance
type Entrance struct {
	Name string `yaml:"name" json:"name"`
	Host string `yaml:"host" json:"host"`
	Port int32  `yaml:"port" json:"port"`
	// Optional. if invisible=true.
	Icon string `yaml:"icon,omitempty" json:"icon,omitempty"`
	// Optional. if invisible=true.
	Title     string `yaml:"title" json:"title,omitempty"`
	AuthLevel string `yaml:"authLevel,omitempty" json:"authLevel,omitempty"`
	Invisible bool   `yaml:"invisible,omitempty" json:"invisible,omitempty"`

	// openMethod has three choices default, iframe, window
	// Optional. if invisible=true.
	OpenMethod string `yaml:"openMethod,omitempty" json:"openMethod,omitempty"`

	WindowPushState bool `yaml:"windowPushState,omitempty" json:"windowPushState,omitempty"`
}

type ServicePort struct {
	Name string `json:"name" yaml:"name"`
	Host string `yaml:"host" json:"host"`
	Port int32  `yaml:"port" json:"port"`

	ExposePort int32 `yaml:"exposePort" json:"exposePort,omitempty"`

	// The protocol for this entrance. Supports "tcp" and "udp".
	// Default is tcp.
	// +default="tcp"
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
	// StartedTime is the time that app first to running state
	StartedTime      *metav1.Time     `json:"startedTime,omitempty"`
	EntranceStatuses []EntranceStatus `json:"entranceStatuses,omitempty"`
}

type EntranceStatus struct {
	Name       string        `json:"name"`
	State      EntranceState `json:"state"`
	StatusTime *metav1.Time  `json:"statusTime"`
	Reason     string        `json:"reason"`
	Message    string        `json:"message,omitempty"`
}

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

// AppResourceName return application name
func AppResourceName(name, namespace string) string {
	return fmt.Sprintf("%s-%s", namespace, name)
}

func (a ApplicationList) GetItems() []Application {
	return a.Items
}

// ACLProto If the ACL proto field is empty, it allows ICMPv4, ICMPv6, TCP, and UDP as per Tailscale behaviour
var ACLProto = sets.NewString("", "igmp", "ipv4", "ip-in-ip", "tcp", "egp", "igp", "udp", "gre", "esp", "ah", "sctp", "icmp")

const expectedTokenItems = 2

var (
	ErrInvalidAction     = errors.New("invalid action")
	ErrInvalidPortFormat = errors.New("invalid port format")
)

func CheckTailScaleACLs(acls []ACL) error {
	if len(acls) == 0 {
		return nil
	}
	var err error
	// fill default value fro ACL
	for i := range acls {
		acls[i].Action = "accept"
		acls[i].Src = []string{"*"}
	}
	for _, acl := range acls {
		err = parseProtocol(acl.Proto)
		if err != nil {
			return err
		}
		for _, dest := range acl.Dst {
			_, _, err = parseDestination(dest)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parseProtocol(protocol string) error {
	if ACLProto.Has(protocol) {
		return nil
	}
	return fmt.Errorf("unsupported protocol: %v", protocol)
}

// parseDestination from
// https://github.com/juanfont/headscale/blob/770f3dcb9334adac650276dcec90cd980af53c6e/hscontrol/policy/acls.go#L475
func parseDestination(dest string) (string, string, error) {
	var tokens []string

	// Check if there is a IPv4/6:Port combination, IPv6 has more than
	// three ":".
	tokens = strings.Split(dest, ":")
	if len(tokens) < expectedTokenItems || len(tokens) > 3 {
		port := tokens[len(tokens)-1]

		maybeIPv6Str := strings.TrimSuffix(dest, ":"+port)

		filteredMaybeIPv6Str := maybeIPv6Str
		if strings.Contains(maybeIPv6Str, "/") {
			networkParts := strings.Split(maybeIPv6Str, "/")
			filteredMaybeIPv6Str = networkParts[0]
		}

		if maybeIPv6, err := netip.ParseAddr(filteredMaybeIPv6Str); err != nil && !maybeIPv6.Is6() {

			return "", "", fmt.Errorf(
				"failed to parse destination, tokens %v: %w",
				tokens,
				ErrInvalidPortFormat,
			)
		} else {
			tokens = []string{maybeIPv6Str, port}
		}
	}

	var alias string
	// We can have here stuff like:
	// git-server:*
	// 192.168.1.0/24:22
	// fd7a:115c:a1e0::2:22
	// fd7a:115c:a1e0::2/128:22
	// tag:montreal-webserver:80,443
	// tag:api-server:443
	// example-host-1:*
	if len(tokens) == expectedTokenItems {
		alias = tokens[0]
	} else {
		alias = fmt.Sprintf("%s:%s", tokens[0], tokens[1])
	}

	return alias, tokens[len(tokens)-1], nil
}
