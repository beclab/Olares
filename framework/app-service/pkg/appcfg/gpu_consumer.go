package appcfg

import (
	"sort"
	"strings"

	"github.com/beclab/api/manifest"
)

const (
	PerPodOne   = manifest.PerPodOne
	PerPodAll   = manifest.PerPodAll
	PerPodSplit = manifest.PerPodSplit
)

// GPUConsumer is one workload that holds GPU cards. A legacy manifest
// (spec.accelerator) yields a single consumer with an empty Name; a manifest
// using workloadOptions.<workload>.accelerator yields one consumer per such
// workload, named after it.
type GPUConsumer struct {
	Name       string
	Containers []string
	PerPod     string
	Replicas   int32
	Modes      []ResourceMode
}

// HasGPUConsumers reports whether the manifest declares accelerators per
// workload. Such apps never read SelectedGpuType or spec.accelerator.
func (c *ApplicationConfig) HasGPUConsumers() bool {
	return c != nil && c.WorkloadOptions.HasAccelerator()
}

// GPUConsumers returns the per-workload consumers sorted by name, the order
// in which cards are placed.
func (c *ApplicationConfig) GPUConsumers() []GPUConsumer {
	if !c.HasGPUConsumers() {
		return nil
	}
	out := make([]GPUConsumer, 0, len(c.WorkloadOptions))
	for name, option := range c.WorkloadOptions {
		if option.Accelerator == nil {
			continue
		}
		replicas := int32(1)
		if option.Replicas != nil {
			replicas = *option.Replicas
		}
		out = append(out, GPUConsumer{
			Name:       name,
			Containers: option.Accelerator.Containers,
			PerPod:     normalizePerPod(option.Accelerator.PerPod),
			Replicas:   replicas,
			Modes:      option.Accelerator.Modes,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// GPUConsumer looks a consumer up by workload name.
func (c *ApplicationConfig) GPUConsumer(name string) (GPUConsumer, bool) {
	for _, consumer := range c.GPUConsumers() {
		if consumer.Name == name {
			return consumer, true
		}
	}
	return GPUConsumer{}, false
}

// SelectedConsumerMode returns the mode chosen for a consumer. A consumer
// declaring exactly one mode needs no explicit choice.
func (c *ApplicationConfig) SelectedConsumerMode(consumer GPUConsumer) (ResourceMode, bool) {
	if selected := c.SelectedGpuTypes[consumer.Name]; selected != "" {
		if mode := findResourceMode(consumer.Modes, selected); mode != nil {
			return *mode, true
		}
		return ResourceMode{}, false
	}
	if len(consumer.Modes) == 1 {
		return consumer.Modes[0], true
	}
	return ResourceMode{}, false
}

func normalizePerPod(perPod string) string {
	switch p := strings.TrimSpace(perPod); p {
	case PerPodAll, PerPodSplit:
		return p
	default:
		return PerPodOne
	}
}

// CardsPerPod returns how many cards each pod of the consumer requests given
// the number of cards bound to it. For "all" the count is decided per node by
// HAMi, so 0 is returned.
func (consumer GPUConsumer) CardsPerPod(boundCards int) int {
	switch consumer.PerPod {
	case PerPodAll:
		return 0
	case PerPodSplit:
		if consumer.Replicas <= 0 || boundCards <= 0 {
			return 0
		}
		return boundCards / int(consumer.Replicas)
	default:
		return 1
	}
}

// PodReplicas returns the replica count the consumer's workload runs with
// given the cards and distinct nodes bound to it.
func (consumer GPUConsumer) PodReplicas(boundCards, boundNodes int) int32 {
	switch consumer.PerPod {
	case PerPodAll:
		return int32(boundNodes)
	case PerPodSplit:
		return consumer.Replicas
	default:
		return int32(boundCards)
	}
}
