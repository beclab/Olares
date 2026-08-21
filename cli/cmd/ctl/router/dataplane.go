package router

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/beclab/Olares/cli/internal/keychain"
)

// Reaching the data plane.
//
// Router's /v1 accepts three identities and tries them in this order: an `sk-*`
// Bearer, the platform-injected `x-caller-appid` of a calling application, and
// the platform-injected `X-BFL-USER` of a person. The third arrived in Router
// v2.2.1 (ADR-49), and it is what makes this file short.
//
// Both injected headers are stamped by the Olares edge on the way in — the same
// edge, the same host and the same profile session that every management verb
// in this tree already travels on, since the console plane works precisely
// because that session becomes X-BFL-USER. So a /v1 request sent from here
// carrying no credential of its own is not anonymous: it arrives as whoever the
// platform says is calling. In a pod that is the application; from a laptop it
// is the person holding the profile.
//
// Hence two steps rather than five:
//
//  1. A key named explicitly — flag, then environment — wins. Someone naming a
//     credential is not asking to be second-guessed, and a key remains the way
//     to call with a model allowlist, a budget of its own, or from outside
//     Olares entirely.
//  2. Otherwise send nothing, and let the platform say who this is.
//
// What stood here before was a keychain lookup, a probe gated on running inside
// a container, and — for every laptop, on the first call of any kind — minting
// a key through the management plane and saving it. That key had no expiry, no
// quota and no model restriction, and `router call models` was enough to create
// one. Nothing in this tree issues a credential as a side effect of using one
// any more, which is also why a key saved by an older build is now ignored
// rather than preferred: it still works in Router, and `key current` says so.
//
// A Router older than v2.2.1 refuses the keyless request with
// `missing_credentials`. That is a message to answer rather than a case to mint
// around, and callErr answers it by naming the upgrade and --api-key.

// dataPlaneKeyEnv names a key for one run without saving it anywhere.
const dataPlaneKeyEnv = "OLARES_ROUTER_API_KEY"

// keychainAccountSuffix separates a data-plane key saved by an older build from
// the profile's own access token, which lives under the bare Olares ID in the
// same keychain service. Nothing writes this entry now; `key current` reads it
// so a key left behind can still be found and dealt with.
const keychainAccountSuffix = "#router-api-key"

// authMode records which identity a call ended up presenting.
type authMode string

const (
	authPlatform authMode = "platform"
	authKey      authMode = "key"
)

type dataPlaneAuth struct {
	Mode authMode
	Key  string
}

// dataPlane returns a client for /v1. Choosing the credential reads a flag and
// an environment variable and speaks to nothing, so there is no context to pass
// and no failure to report: without a key the request carries no Authorization
// at all and the platform vouches for the caller.
func dataPlane(pc *preparedClient, explicitKey string) *routerClient {
	auth := resolveDataPlaneAuth(explicitKey)
	if auth.Mode == authPlatform {
		return pc.router
	}
	return pc.router.withHeader("Authorization", "Bearer "+auth.Key)
}

func resolveDataPlaneAuth(explicitKey string) *dataPlaneAuth {
	if k := strings.TrimSpace(explicitKey); k != "" {
		return &dataPlaneAuth{Mode: authKey, Key: k}
	}
	if k := strings.TrimSpace(os.Getenv(dataPlaneKeyEnv)); k != "" {
		return &dataPlaneAuth{Mode: authKey, Key: k}
	}
	return &dataPlaneAuth{Mode: authPlatform}
}

func keychainAccount(olaresID string) string {
	return olaresID + keychainAccountSuffix
}

func cachedDataPlaneKey(olaresID string) (string, error) {
	if strings.TrimSpace(olaresID) == "" {
		return "", nil
	}
	return keychain.Get(keychain.OlaresCliService, keychainAccount(olaresID))
}

func forgetDataPlaneKey(olaresID string) error {
	if strings.TrimSpace(olaresID) == "" {
		return nil
	}
	return keychain.Remove(keychain.OlaresCliService, keychainAccount(olaresID))
}

// dataPlaneKeyFlagUsage is shared by every calling verb so the sentence a user
// reads about credentials is the same one wherever they meet it.
const dataPlaneKeyFlagUsage = "call with this `sk-*` key instead of the one this machine keeps"

// callErr adds what a refused call needs next. The data plane refuses for
// reasons that live in three different places — the credential, the catalogue,
// and the budget — and the message alone rarely says which.
//
// Two credentials can fail one call, and they read almost identically: ours to
// Router, and Router's to the provider. Both come back 401 `invalid_api_key`,
// because Router forwards what the upstream said. The `type` is the only thing
// that separates them — Router's own refusals are `authentication_error`, an
// upstream's carry whatever the upstream calls it — and pointing at the wrong
// key sends someone to re-issue a credential that was never the problem.
func callErr(err error) error {
	if err == nil {
		return nil
	}
	var re *RouterError
	if !errors.As(err, &re) {
		return err
	}
	ours := re.Type == "authentication_error"
	switch {
	case ours && re.Code == "missing_credentials":
		// Router saw no Bearer, no app header and no X-BFL-USER. From here the
		// third one is always sent by the edge, so the Router that says this is
		// one that predates reading it.
		return fmt.Errorf("%w\nThis Router does not yet accept a call without a key: taking the caller's "+
			"identity from the platform arrived in Router v2.2.1, and `olares-cli market upgrade` on the "+
			"Router application is what closes the gap. Until then pass --api-key or set %s; "+
			"`olares-cli router key list` shows the keys that exist and `router key issue <name>` "+
			"creates one", err, dataPlaneKeyEnv)
	case ours && re.Code == "unknown_bfl_user":
		// The platform vouched for a person Router has no row for. Router
		// deliberately does not create one from the data plane, and neither
		// does this: the console plane is where a person comes into existence.
		return fmt.Errorf("%w\nThe platform knows you, Router does not have you yet. It records a person "+
			"the first time they use the console plane, so any management verb — `olares-cli router model "+
			"list` will do — creates the row, and the call then works", err)
	case ours && (re.Code == "invalid_api_key" || re.Code == "key_disabled" || re.Code == "key_expired"):
		return fmt.Errorf("%w\nThe key this call presented is no longer usable. Dropping it is enough to "+
			"keep working: without a key the call goes as you, vouched for by the platform. Unset %s "+
			"if that is where it came from, leave --api-key off, and `olares-cli router key current "+
			"--forget` discards a copy saved by an older build", err, dataPlaneKeyEnv)
	case ours && (re.Code == "owner_disabled" || re.Code == "app_archived" || re.Code == "app_suspended"):
		return fmt.Errorf("%w\nThis is about who or what the key belongs to rather than the key itself. "+
			"For a person, `olares-cli settings users get <name>` shows their account; for an "+
			"application, `olares-cli market list --mine` says whether it is still here", err)
	case re.Status == 401 || re.Status == 403:
		return fmt.Errorf("%w\nThis came back from the provider, not from Router: the credential Router "+
			"holds for it was refused. `olares-cli router provider validate <provider>` checks it against "+
			"the upstream, and `router provider update` replaces it", err)
	case re.Code == "no_default_model":
		return fmt.Errorf("%w\nNothing installed can serve that category. `olares-cli router route list --kind default` "+
			"says where each one stands, and it fills once a model of that kind exists — a category is "+
			"maintained against what is configured rather than pointed at by hand. --model names one "+
			"directly in the meantime", err)
	case re.Code == "model_route_disabled":
		return fmt.Errorf("%w\nThe name resolves, but the route serving it is switched off. "+
			"`olares-cli router route get <name>` shows it, and `route enable <name>` puts it back", err)
	case re.Code == "model_not_allowed":
		return fmt.Errorf("%w\nThe credential is restricted to a list this model is not on. "+
			"`olares-cli router call models` shows what it may call, and `router key update` changes the list", err)
	case re.Code == "model_not_found":
		return fmt.Errorf("%w\n`olares-cli router call models` shows the names this credential may send", err)
	case re.Code == "ambiguous_model":
		return fmt.Errorf("%w\nQualify it as <provider>/<model>; `olares-cli router call models` shows the "+
			"qualified names", err)
	case re.Type == "quota_exceeded_error":
		return fmt.Errorf("%w\n`olares-cli router quota list` shows the limits, and `router usage summary` "+
			"what has been spent against them", err)
	case re.Code == "stream_unsupported_on_fallback_provider":
		// Router only reaches its legacy single-upstream path when no provider
		// claims the model, and complains about streaming from there. The name
		// is what was actually wrong.
		return fmt.Errorf("%w\nNo provider serves this model, which is what put the request on Router's "+
			"fallback path. `olares-cli router call models` shows the names it does serve", err)
	case strings.HasSuffix(re.Code, "_mode_mismatch"):
		return fmt.Errorf("%w\nThe model is configured for a different kind of work than this verb asks "+
			"for. `olares-cli router model list` shows each model's mode, and a route can only serve the mode "+
			"it was created with", err)
	case strings.HasSuffix(re.Code, "_unsupported_for_provider") || re.Code == "audio_path_unsupported":
		return fmt.Errorf("%w\nThis model's provider does not serve that route at all. For a model "+
			"running on this Olares that usually means a different engine image does this job: "+
			"`olares-cli router model list --mode audio` and `router provider get <provider>` show which "+
			"capability each one declares", err)
	case re.Code == "capability_not_supported" || re.Code == "stream_unsupported_for_model":
		return fmt.Errorf("%w\n`olares-cli router provider get <provider>` lists what the model supports", err)
	case re.Status == 404 && re.Code == "" && re.Type == "":
		// A 404 with no envelope on a route Router does mount. Audio, OCR and
		// translate are forwarded to the engine unchanged, and an engine serves
		// the capability it was built for and nothing else, so this is usually
		// the wrong engine rather than a wrong URL.
		return fmt.Errorf("%w\nRouter forwarded this and the model's own endpoint has no such route. "+
			"For audio that means the wrong engine: recognition, synthesis, voice activity, "+
			"diarization, enhancement and alignment are separate images, and each answers 404 for "+
			"the others. Leaving --model off picks a model that declares the capability", err)
	case re.Status == 502 || re.Status == 503 || re.Status == 504:
		return fmt.Errorf("%w\nThe model's own endpoint did not answer. For a model running on this Olares "+
			"that usually means the application is not serving yet: `olares-cli router provider get <provider>` "+
			"shows its Olares status, and a degraded one is a matter for `olares-cli market` rather than "+
			"anything here", err)
	}
	return err
}
