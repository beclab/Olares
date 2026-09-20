package clusterop

import (
	"fmt"
	"strings"
	"sync"
)

// SignatureRequirementRegistry holds the operation types that the master's
// POST /cluster/operations route must admit only with an owner signature.
// Types not registered here may be admitted with an access token instead;
// the handler's binding check is separate and unchanged.
type SignatureRequirementRegistry struct {
	mu     sync.RWMutex
	types  map[Type]struct{}
	frozen bool
}

var defaultSignatureRequirements = NewSignatureRequirementRegistry()

func NewSignatureRequirementRegistry() *SignatureRequirementRegistry {
	return &SignatureRequirementRegistry{
		types: make(map[Type]struct{}),
	}
}

func DefaultSignatureRequirementRegistry() *SignatureRequirementRegistry {
	return defaultSignatureRequirements
}

// MustRequireSignature records that typ must present an owner signature on
// the master's create route. It panics when the registration is refused,
// matching MustRegisterModule.
func MustRequireSignature(typ Type) {
	if err := defaultSignatureRequirements.Register(typ); err != nil {
		panic(err)
	}
}

// RequiresSignature reports whether the default registry requires an owner
// signature for typ on the master's create route.
func RequiresSignature(typ Type) bool {
	return defaultSignatureRequirements.Requires(typ)
}

func (r *SignatureRequirementRegistry) Register(typ Type) error {
	typ = Type(strings.TrimSpace(string(typ)))
	if typ == "" {
		return fmt.Errorf("cluster operation signature requirement type is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.frozen {
		return fmt.Errorf("cluster operation signature requirement registry is frozen")
	}
	if _, exists := r.types[typ]; exists {
		return fmt.Errorf("cluster operation signature requirement %q already registered", typ)
	}
	r.types[typ] = struct{}{}
	return nil
}

func (r *SignatureRequirementRegistry) Requires(typ Type) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.types[Type(strings.TrimSpace(string(typ)))]
	return ok
}

func (r *SignatureRequirementRegistry) Freeze() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.frozen = true
}
