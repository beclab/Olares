package dids

import (
	"context"
	"sync"

	"olares-cli/pkg/web5/dids/did"

	"olares-cli/pkg/web5/dids/didcore"
)

// Resolve resolves the provided DID URI. This function is capable of resolving
// the DID methods implemented in web5-go
func Resolve(uri string) (didcore.ResolutionResult, error) {
	return getDefaultResolver().Resolve(uri)
}

// ResolveWithContext resolves the provided DID URI. This function is capable of resolving
// the DID methods implemented in web5-go
func ResolveWithContext(ctx context.Context, uri string) (didcore.ResolutionResult, error) {
	return getDefaultResolver().ResolveWithContext(ctx, uri)
}

var instance *didResolver
var once sync.Once

func getDefaultResolver() *didResolver {
	once.Do(func() {
		instance = &didResolver{
			resolvers: map[string]didcore.MethodResolver{
				// "dht": diddht.DefaultResolver(),
				// "jwk": didjwk.Resolver{},
				// "web": didweb.Resolver{},
			},
		}
	})

	return instance
}

type didResolver struct {
	resolvers map[string]didcore.MethodResolver
}

func (r *didResolver) Resolve(uri string) (didcore.ResolutionResult, error) {
	did, err := did.Parse(uri)
	if err != nil {
		return didcore.ResolutionResultWithError("invalidDid"), didcore.ResolutionError{Code: "invalidDid"}
	}

	resolver := r.resolvers[did.Method]
	if resolver == nil {
		return didcore.ResolutionResultWithError("methodNotSupported"), didcore.ResolutionError{Code: "methodNotSupported"}
	}

	return resolver.Resolve(uri)
}

func (r *didResolver) ResolveWithContext(ctx context.Context, uri string) (didcore.ResolutionResult, error) {
	did, err := did.Parse(uri)
	if err != nil {
		return didcore.ResolutionResultWithError("invalidDid"), didcore.ResolutionError{Code: "invalidDid"}
	}

	resolver := r.resolvers[did.Method]
	if resolver == nil {
		return didcore.ResolutionResultWithError("methodNotSupported"), didcore.ResolutionError{Code: "methodNotSupported"}
	}

	return resolver.ResolveWithContext(ctx, uri)
}
