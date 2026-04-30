package node

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Node is the minimal corev1.Node + status subset rendered by the
// cluster node verbs. Mirrors the projection KubeSphere returns from
// /kapis/resources.kubesphere.io/v1alpha3/nodes.
type Node struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	Metadata   struct {
		Name              string            `json:"name"`
		UID               string            `json:"uid,omitempty"`
		CreationTimestamp string            `json:"creationTimestamp,omitempty"`
		Labels            map[string]string `json:"labels,omitempty"`
		Annotations       map[string]string `json:"annotations,omitempty"`
	} `json:"metadata"`
	Spec struct {
		Unschedulable bool      `json:"unschedulable,omitempty"`
		Taints        []NodeTaint `json:"taints,omitempty"`
	} `json:"spec,omitempty"`
	Status struct {
		Conditions  []NodeCondition `json:"conditions,omitempty"`
		Addresses   []NodeAddress   `json:"addresses,omitempty"`
		NodeInfo    NodeInfo        `json:"nodeInfo,omitempty"`
		Capacity    map[string]string `json:"capacity,omitempty"`
		Allocatable map[string]string `json:"allocatable,omitempty"`
	} `json:"status,omitempty"`
}

type NodeTaint struct {
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect"`
}

type NodeCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

type NodeAddress struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

type NodeInfo struct {
	OSImage         string `json:"osImage,omitempty"`
	KernelVersion   string `json:"kernelVersion,omitempty"`
	Architecture    string `json:"architecture,omitempty"`
	OperatingSystem string `json:"operatingSystem,omitempty"`
	KubeletVersion  string `json:"kubeletVersion,omitempty"`
	ContainerRuntimeVersion string `json:"containerRuntimeVersion,omitempty"`
}

// StatusLabel returns the kubectl-style STATUS column.
//   - "Ready"     when conditions[type=Ready].status == "True"
//     and the node is schedulable.
//   - "Ready,SchedulingDisabled" when Ready=True but
//     spec.unschedulable=true (cordoned).
//   - "NotReady"  when Ready=False or Unknown.
//   - "Unknown"   when no Ready condition is present at all.
//
// Named with the "Label" suffix because Status() would shadow the
// `Status` field on the same struct.
func (n Node) StatusLabel() string {
	ready := ""
	for _, c := range n.Status.Conditions {
		if c.Type == "Ready" {
			ready = c.Status
			break
		}
	}
	switch ready {
	case "True":
		if n.Spec.Unschedulable {
			return "Ready,SchedulingDisabled"
		}
		return "Ready"
	case "False":
		return "NotReady"
	case "":
		return "Unknown"
	default:
		return "NotReady"
	}
}

// Roles returns the comma-joined list of node roles, derived from
// the standard `node-role.kubernetes.io/<role>` label key prefix.
// Empty (no role labels) → "<none>" (matches `kubectl get nodes`).
func (n Node) Roles() string {
	prefix := "node-role.kubernetes.io/"
	roles := []string{}
	for k := range n.Metadata.Labels {
		if strings.HasPrefix(k, prefix) {
			r := strings.TrimPrefix(k, prefix)
			if r != "" {
				roles = append(roles, r)
			}
		}
	}
	if len(roles) == 0 {
		return "<none>"
	}
	sort.Strings(roles)
	return strings.Join(roles, ",")
}

// Age returns the AGE column. Mirrors pod.ageOf semantics.
func (n Node) Age(now time.Time) string {
	return ageOf(n.Metadata.CreationTimestamp, now)
}

// InternalIP returns the first InternalIP address (matches kubectl
// behavior). Falls back to ExternalIP, then "-".
func (n Node) InternalIP() string {
	for _, a := range n.Status.Addresses {
		if a.Type == "InternalIP" {
			return a.Address
		}
	}
	for _, a := range n.Status.Addresses {
		if a.Type == "ExternalIP" {
			return a.Address
		}
	}
	return "-"
}

// KubeletVersion is a small accessor — keeps the rendering code in
// list.go / get.go from poking deep into Status.NodeInfo.
func (n Node) KubeletVersion() string {
	return n.Status.NodeInfo.KubeletVersion
}

// ageOf / dashIfEmpty mirror the helpers in cmd/ctl/cluster/pod/types.go.
func ageOf(ts string, now time.Time) string {
	if ts == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return "-"
	}
	d := now.Sub(t)
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
