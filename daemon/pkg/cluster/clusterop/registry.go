package clusterop

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type ModuleRegistry struct {
	mu      sync.RWMutex
	modules map[Type]OperationModule
	frozen  bool
}

var defaultRegistry = NewRegistry()

func NewRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		modules: make(map[Type]OperationModule),
	}
}

func DefaultRegistry() *ModuleRegistry { return defaultRegistry }

func MustRegisterModule(module OperationModule) {
	if err := defaultRegistry.Register(module); err != nil {
		panic(err)
	}
}

func (r *ModuleRegistry) Register(module OperationModule) error {
	if module == nil {
		return fmt.Errorf("cluster operation module is nil")
	}
	value := reflect.ValueOf(module)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return fmt.Errorf("cluster operation module is nil")
	}

	typ := Type(strings.TrimSpace(string(module.Type())))
	if typ == "" {
		return fmt.Errorf("cluster operation type is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.frozen {
		return fmt.Errorf("cluster operation registry is frozen")
	}
	if _, exists := r.modules[typ]; exists {
		return fmt.Errorf("cluster operation module %q already registered", typ)
	}
	r.modules[typ] = module
	return nil
}

func (r *ModuleRegistry) Lookup(typ Type) (OperationModule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, ok := r.modules[typ]
	return module, ok
}

func (r *ModuleRegistry) Parse(value string) (Type, error) {
	typ := Type(value)
	if _, ok := r.Lookup(typ); ok {
		return typ, nil
	}
	return "", fmt.Errorf("unsupported cluster operation type %q", value)
}

// Types lists what this registry holds. The order is not defined: it is for
// asking a question about the whole set, not for iterating it in sequence.
func (r *ModuleRegistry) Types() []Type {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]Type, 0, len(r.modules))
	for typ := range r.modules {
		types = append(types, typ)
	}
	return types
}

func (r *ModuleRegistry) Freeze() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.frozen = true
}
