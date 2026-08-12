package nodestatus

import (
	"fmt"
	"strings"
	"sync"
)

type namedProbe struct {
	name  string
	probe Probe
}

// ProbeRegistry holds the capability probes that Detect walks. Names are the
// capability keys; registration order is the walk order.
type ProbeRegistry struct {
	mu     sync.RWMutex
	probes []namedProbe
	byName map[string]int
	frozen bool
}

var defaultProbeRegistry = NewProbeRegistry()

func NewProbeRegistry() *ProbeRegistry {
	return &ProbeRegistry{
		byName: make(map[string]int),
	}
}

func DefaultProbeRegistry() *ProbeRegistry { return defaultProbeRegistry }

// MustRegisterProbe adds a probe to the package default registry. It panics
// when the registration is refused, matching MustRegisterModule.
func MustRegisterProbe(name string, p Probe) {
	if err := defaultProbeRegistry.Register(name, p); err != nil {
		panic(err)
	}
}

func (r *ProbeRegistry) Register(name string, p Probe) error {
	if p == nil {
		return fmt.Errorf("capability probe is nil")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("capability probe name is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.frozen {
		return fmt.Errorf("capability probe registry is frozen")
	}
	if _, exists := r.byName[name]; exists {
		return fmt.Errorf("capability probe %q already registered", name)
	}
	r.byName[name] = len(r.probes)
	r.probes = append(r.probes, namedProbe{name: name, probe: p})
	return nil
}

// Probes returns the registered probes in registration order.
func (r *ProbeRegistry) Probes() []Probe {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Probe, len(r.probes))
	for i, np := range r.probes {
		out[i] = np.probe
	}
	return out
}

func (r *ProbeRegistry) Freeze() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.frozen = true
}
