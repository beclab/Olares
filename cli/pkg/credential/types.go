// Package credential is the orchestration layer that turns a
// cliconfig.ProfileConfig + a stored token into a fully-resolved view that
// command code can consume without touching disk directly.
//
// There are two kinds of profile and one Provider for each: DefaultProvider
// for an account somebody logged into on this machine, and ManagedProvider
// for one the platform issued by mounting a credential into an application
// container. CredentialProvider picks between them by reading the selected
// profile, never by trying both.
package credential

import (
	"context"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// ResolvedProfile is the "ready to make an API call" view of a profile —
// analogous to lark-cli's CliConfig. Command code interacts only with this
// struct, so where the token came from is not something a verb has to know.
type ResolvedProfile struct {
	Name     string // alias, falls back to OlaresID
	OlaresID string
	UserUID  string

	AuthURL      string
	VaultURL     string
	DesktopURL   string
	SettingsURL  string
	FilesURL     string
	MarketURL    string
	DashboardURL string
	// ControlHubURL is the per-user ControlHub BFF base URL
	// ("https://control-hub.<terminus>"). The `cluster` command tree uses
	// this — see pkg/olares/id.go::ControlHubURL for the full description
	// of which path prefixes ride this origin.
	ControlHubURL string

	AccessToken string
	// ExpiresAt is the unix-seconds expiry decoded from AccessToken's `exp`
	// claim. Zero means "no exp claim found" and is treated as "trust the
	// token until the server says otherwise".
	ExpiresAt int64

	// Source identifies which Provider produced this ResolvedProfile (for
	// diagnostics: "default" or "managed").
	Source string

	// InsecureSkipVerify is forwarded from the underlying ProfileConfig so
	// HTTP clients constructed against this profile honor the dev override.
	InsecureSkipVerify bool

	// Managed and AppName mirror the same fields on ProfileConfig, so error
	// formatters can tell a user whose grant came from an application
	// install that logging in is not the way out of a failure.
	Managed bool
	AppName string
}

// Provider is implemented by anything that can turn a ProfileConfig into a
// ResolvedProfile. Returning (nil, nil) means "this is not a profile I own",
// which the orchestrating CredentialProvider surfaces as ErrNoProfile.
//
// The `profile` argument is the currently-selected ProfileConfig from
// cliconfig, supplied by CredentialProvider.
type Provider interface {
	Name() string
	Resolve(ctx context.Context, profile *cliconfig.ProfileConfig) (*ResolvedProfile, error)
}
