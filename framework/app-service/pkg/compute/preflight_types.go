package compute

import (
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
)

const (
	// PreflightSharedOwner identifies cluster-scoped shared applications.
	PreflightSharedOwner           = "shared"
	PreflightReasonClusterCapacity = "cluster_capacity"
	PreflightReasonClusterPressure = "cluster_pressure"
	PreflightReasonOwnerQuota      = "owner_quota"
	PreflightReasonK8sRequest      = "k8s_request"
	PreflightReasonNodePressure    = "node_pressure"
	PreflightReasonUnschedulable   = "unschedulable"
)

type PreflightDemand struct {
	ID          string      `json:"id"`
	Owner       string      `json:"owner"`
	Application string      `json:"application"`
	AppID       string      `json:"appId,omitempty"`
	Requirement Requirement `json:"requirement"`
}

type PreflightSnapshot struct {
	Nodes        []Node
	Pressure     PressureSnapshot
	Cluster      *prometheus.ClusterMetrics
	Owners       map[string]*prometheus.ClusterMetrics
	K8sAvailable apputils.ResourceState
}

type PreflightPlacement struct {
	NodeNames   []string
	Allocations []Allocation
}

type PreflightPressure struct {
	Source     string              `json:"source"`
	Dimensions []DimensionPressure `json:"dimensions"`
}

type PreflightReport struct {
	Installable    bool                `json:"installable"`
	FailedDemandID string              `json:"failedDemandId,omitempty"`
	Reason         string              `json:"reason,omitempty"`
	Pressure       []PreflightPressure `json:"pressure,omitempty"`
}

func clonePreflightSnapshot(snapshot PreflightSnapshot) PreflightSnapshot {
	cloned := snapshot
	cloned.Nodes = make([]Node, len(snapshot.Nodes))
	for i, node := range snapshot.Nodes {
		cloned.Nodes[i] = node
		cloned.Nodes[i].GPUTypes = append([]string(nil), node.GPUTypes...)
		cloned.Nodes[i].Devices = make([]Device, len(node.Devices))
		for j, device := range node.Devices {
			cloned.Nodes[i].Devices[j] = device
			cloned.Nodes[i].Devices[j].AvailableSupportTypes = append([]string(nil), device.AvailableSupportTypes...)
			cloned.Nodes[i].Devices[j].Bindings = append([]Allocation(nil), device.Bindings...)
		}
	}
	cloned.Pressure.UsageByNode = make(map[string]prometheus.NodeResourceUsage, len(snapshot.Pressure.UsageByNode))
	for nodeName, usage := range snapshot.Pressure.UsageByNode {
		cloned.Pressure.UsageByNode[nodeName] = usage
	}
	if snapshot.Cluster != nil {
		cluster := *snapshot.Cluster
		cloned.Cluster = &cluster
	}
	cloned.Owners = make(map[string]*prometheus.ClusterMetrics, len(snapshot.Owners))
	for owner, metrics := range snapshot.Owners {
		if metrics != nil {
			clonedMetrics := *metrics
			metrics = &clonedMetrics
		}
		cloned.Owners[owner] = metrics
	}
	return cloned
}
