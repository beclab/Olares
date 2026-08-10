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

	module, ok := r.modules[Type(strings.TrimSpace(string(typ)))]
	return module, ok
}

func (r *ModuleRegistry) Parse(value string) (Type, error) {
	typ := Type(strings.TrimSpace(value))
	if _, ok := r.Lookup(typ); ok {
		return typ, nil
	}
	return "", fmt.Errorf("unsupported cluster operation type %q", value)
}

func (r *ModuleRegistry) Freeze() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.frozen = true
}
