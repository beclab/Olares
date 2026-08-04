package validation

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PreflightSnapshotCollector struct {
	client         client.Client
	nodes          func(context.Context, client.Client) ([]compute.Node, error)
	pressure       func(context.Context) (compute.PressureSnapshot, error)
	clusterMetrics func(string) (*prometheus.ClusterMetrics, []string, error)
	userMetrics    func(context.Context, string) (*prometheus.ClusterMetrics, error)
	k8sAvailable   func() (resource.Quantity, resource.Quantity, error)
}

func NewPreflightSnapshotCollector(ctrlClient client.Client) *PreflightSnapshotCollector {
	return &PreflightSnapshotCollector{
		client:         ctrlClient,
		nodes:          compute.FetchSchedulableNodeComputeAllocations,
		pressure:       compute.FetchPressureSnapshot,
		clusterMetrics: apputils.GetClusterResource,
		userMetrics:    prometheus.GetCurUserResource,
		k8sAvailable:   apputils.GetClusterAvailableResourceQuantities,
	}
}

func (c *PreflightSnapshotCollector) Collect(ctx context.Context, token string, owners []string) (compute.PreflightSnapshot, error) {
	nodes, err := c.nodes(ctx, c.client)
	if err != nil {
		return compute.PreflightSnapshot{}, fmt.Errorf("collect nodes and allocations: %w", err)
	}
	pressure, err := c.pressure(ctx)
	if err != nil {
		return compute.PreflightSnapshot{}, fmt.Errorf("collect node pressure: %w", err)
	}
	if err := compute.ValidatePressureSnapshot(nodes, pressure); err != nil {
		return compute.PreflightSnapshot{}, fmt.Errorf("validate node pressure: %w", err)
	}
	cluster, _, err := c.clusterMetrics(token)
	if err != nil {
		return compute.PreflightSnapshot{}, fmt.Errorf("collect cluster metrics: %w", err)
	}
	if _, err := apputils.EvaluateClusterPressure(apputils.ResourceState{}, cluster, apputils.AllResourceDimensions); err != nil {
		return compute.PreflightSnapshot{}, fmt.Errorf("validate cluster metrics: %w", err)
	}

	ownerMetrics := make(map[string]*prometheus.ClusterMetrics)
	for _, owner := range owners {
		if owner == compute.PreflightSharedOwner {
			continue
		}
		if _, exists := ownerMetrics[owner]; exists {
			continue
		}
		metrics, err := c.userMetrics(ctx, owner)
		if err != nil {
			return compute.PreflightSnapshot{}, fmt.Errorf("collect metrics for owner %q: %w", owner, err)
		}
		if _, err := apputils.EvaluateOwnerPressure(apputils.ResourceState{}, metrics, apputils.AllResourceDimensions); err != nil {
			return compute.PreflightSnapshot{}, fmt.Errorf("validate metrics for owner %q: %w", owner, err)
		}
		ownerMetrics[owner] = metrics
	}

	cpu, memory, err := c.k8sAvailable()
	if err != nil {
		return compute.PreflightSnapshot{}, fmt.Errorf("collect kubernetes request availability: %w", err)
	}
	if _, err := apputils.EvaluateK8sRequest(apputils.ResourceState{}, apputils.AllResourceDimensions, cpu, memory); err != nil {
		return compute.PreflightSnapshot{}, fmt.Errorf("validate kubernetes request availability: %w", err)
	}
	return compute.PreflightSnapshot{
		Nodes: nodes, Pressure: pressure, Cluster: cluster, Owners: ownerMetrics,
		K8sAvailable: apputils.ResourceState{CPU: cpu.MilliValue(), Memory: memory.Value()},
	}, nil
}
