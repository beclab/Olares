package credential

import (
	"context"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// EnvProvider is the placeholder for the in-cluster ("sandbox") scenario:
// when olares-cli runs inside an application container, the user-service
// will inject access_token / scope / olaresId via environment variables and
// this provider will surface them as a ResolvedProfile.
//
// Phase 1 ships an inert implementation that always declines (returns
// (nil, nil)) so the chain falls through to DefaultProvider. Phase 3 will fill
// it in once the user-service env-var contract is finalized.
type EnvProvider struct{}

// NewEnvProvider returns the Phase-1 stub. The real implementation will accept
// a config struct here.
func NewEnvProvider() Provider { return &EnvProvider{} }

// Name implements Provider.
func (e *EnvProvider) Name() string { return "env" }

// Resolve implements Provider. Always declines in Phase 1.
func (e *EnvProvider) Resolve(_ context.Context, _ *cliconfig.ProfileConfig) (*ResolvedProfile, error) {
	return nil, nil
}
