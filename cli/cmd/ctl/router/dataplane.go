package router

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/beclab/Olares/cli/internal/keychain"
)

// Reaching the data plane.
//
// The management plane runs on the Olares session this whole tree carries: the
// edge injects X-BFL-USER and Router trusts it. The data plane does not accept
// that header. It takes an `sk-*` key, or a platform-injected caller identity
// when the request came from inside the cluster — and nothing else, on purpose:
// a browser session that could also spend budget would make every open Router
// tab a spending credential.
//
// So a call has to answer "which of the two do I have" before it can be made.
// The order below is capability-first rather than configuration-first, because
// the same binary runs in both places and asking the user which one they are in
// is a question the process can answer itself:
//
//  1. A key given explicitly — flag or environment — wins. Someone naming a
//     credential is not asking to be second-guessed.
//  2. A key this machine minted before, from the OS keychain. Reusing it keeps
//     one key per machine instead of one per invocation.
//  3. Inside a container: try with no credential at all. If the platform is
//     injecting a caller identity, that succeeds and no key needs to exist.
//  4. Otherwise mint one through the management plane and keep it.
//
// Step 3 is skipped outside a container. It cannot succeed there — no platform
// component is injecting anything into a request from a laptop — and spending a
// round trip to be told so would slow down the common case.

// dataPlaneKeyEnv names a key without storing one. It is read before the
// keychain so a scripted run can pin a credential with a budget of its own.
const dataPlaneKeyEnv = "OLARES_ROUTER_API_KEY"

// keychainAccountSuffix separates the data-plane key from the profile's own
// access token, which lives under the bare Olares ID in the same service.
const keychainAccountSuffix = "#router-api-key"

// authMode records which of the two credentials a call ended up using, so the
// verbs can say so when it matters and stay quiet when it does not.
type authMode string

const (
	authPlatform authMode = "platform"
	authKey      authMode = "key"
)

type dataPlaneAuth struct {
	Mode authMode
	Key  string
	// Minted is true when this run created the key, which is the one time the
	// user should hear about it: a new credential now exists on the server.
	Minted bool
}

// dataPlane returns a client for /v1 with whatever credential this machine can
// present, and a note to show the user when a key was created for them.
func dataPlane(ctx context.Context, pc *preparedClient, explicitKey string) (*routerClient, *dataPlaneAuth, error) {
	auth, err := resolveDataPlaneAuth(ctx, pc, explicitKey)
	if err != nil {
		return nil, nil, err
	}
	if auth.Mode == authPlatform {
		return pc.router, auth, nil
	}
	return pc.router.withHeader("Authorization", "Bearer "+auth.Key), auth, nil
}

func resolveDataPlaneAuth(ctx context.Context, pc *preparedClient, explicitKey string) (*dataPlaneAuth, error) {
	if k := strings.TrimSpace(explicitKey); k != "" {
		return &dataPlaneAuth{Mode: authKey, Key: k}, nil
	}
	if k := strings.TrimSpace(os.Getenv(dataPlaneKeyEnv)); k != "" {
		return &dataPlaneAuth{Mode: authKey, Key: k}, nil
	}
	if k, err := cachedDataPlaneKey(pc.profile.OlaresID); err == nil && k != "" {
		return &dataPlaneAuth{Mode: authKey, Key: k}, nil
	}
	if inContainer() && platformIdentityWorks(ctx, pc) {
		return &dataPlaneAuth{Mode: authPlatform}, nil
	}
	key, err := mintDataPlaneKey(ctx, pc)
	if err != nil {
		return nil, err
	}
	// A keychain that will not store is not worth failing the call over: the
	// key works for this run, and the next run mints another one.
	if serr := storeDataPlaneKey(pc.profile.OlaresID, key); serr != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save the new key locally (%v); "+
			"the next call will issue another one\n", serr)
	}
	return &dataPlaneAuth{Mode: authKey, Key: key, Minted: true}, nil
}

// inContainer decides whether a platform-injected identity is even possible.
// The Kubernetes service-account mount is the reliable signal; the environment
// variables are set for pods in a cluster and absent from a laptop.
func inContainer() bool {
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		return true
	}
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

// platformIdentityWorks asks the cheapest data-plane route whether this request
// arrives with an identity Router accepts. /v1/models reads a projection of the
// catalogue, so a probe costs no upstream call and no spend.
func platformIdentityWorks(ctx context.Context, pc *preparedClient) bool {
	var sink struct {
		Object string `json:"object"`
	}
	return pc.router.doJSON(ctx, "GET", epDataPlaneModels, nil, &sink) == nil
}

// mintDataPlaneKey issues a key through the management plane. The name carries
// the machine so a key list read months later says where the credential lives,
// which is the question that matters when deciding whether to revoke one.
func mintDataPlaneKey(ctx context.Context, pc *preparedClient) (string, error) {
	name := "olares-cli"
	if host, err := os.Hostname(); err == nil && strings.TrimSpace(host) != "" {
		name += " on " + strings.TrimSpace(host)
	}
	var created createdKey
	err := pc.router.doJSON(ctx, "POST", epAPIKeys, map[string]any{"name": name}, &created)
	if err != nil {
		var re *RouterError
		if errors.As(err, &re) && re.Status == 404 {
			return "", fmt.Errorf("this Router serves no API keys, so the data plane cannot be reached with one. " +
				"A call from inside the cluster still works, where the platform supplies the caller identity")
		}
		return "", fmt.Errorf("issue a key for this machine: %w", err)
	}
	if strings.TrimSpace(created.Key) == "" {
		return "", fmt.Errorf("Router created a key but returned no secret; " +
			"`olares-cli router key list` will show it, and it has to be revoked and re-issued to be usable")
	}
	fmt.Fprintf(os.Stderr, "issued a data-plane key for this machine (%q, %s) and saved it locally; "+
		"`olares-cli router key list` shows it and `router key revoke` ends it\n", name, created.KeyPrefix)
	return created.Key, nil
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

func storeDataPlaneKey(olaresID, key string) error {
	if strings.TrimSpace(olaresID) == "" {
		return fmt.Errorf("no Olares ID to file the key under")
	}
	return keychain.Set(keychain.OlaresCliService, keychainAccount(olaresID), key)
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
	case ours && (re.Code == "invalid_api_key" || re.Code == "key_disabled" || re.Code == "key_expired"):
		return fmt.Errorf("%w\nThe key this machine saved is no longer usable. "+
			"`olares-cli router key local --forget` drops it, and the next call issues a fresh one", err)
	case ours && (re.Code == "owner_disabled" || re.Code == "app_archived" || re.Code == "app_suspended"):
		return fmt.Errorf("%w\nThis is about who or what the key belongs to rather than the key itself. "+
			"For a person, `olares-cli settings users get <name>` shows their account; for an "+
			"application, `olares-cli market list --mine` says whether it is still here", err)
	case re.Status == 401 || re.Status == 403:
		return fmt.Errorf("%w\nThis came back from the provider, not from Router: the credential Router "+
			"holds for it was refused. `olares-cli router provider validate <provider>` checks it against "+
			"the upstream, and `router provider update` replaces it", err)
	case re.Code == "no_default_model":
		return fmt.Errorf("%w\nNothing installed can serve that category. `olares-cli router default show` "+
			"says where each one stands, and it fills once a model of that kind exists — a category is "+
			"maintained against what is configured rather than pointed at by hand. --model names one "+
			"directly in the meantime", err)
	case re.Code == "model_route_disabled":
		return fmt.Errorf("%w\nThe name resolves, but the route serving it is switched off. "+
			"`olares-cli router route get <name>` shows it, and `route enable <name>` puts it back", err)
	case re.Code == "model_not_allowed":
		return fmt.Errorf("%w\nThe credential is restricted to a list this model is not on. "+
			"`olares-cli router models` shows what it may call, and `router key update` changes the list", err)
	case re.Code == "model_not_found":
		return fmt.Errorf("%w\n`olares-cli router models` shows the names this credential may send", err)
	case re.Code == "ambiguous_model":
		return fmt.Errorf("%w\nQualify it as <provider>/<model>; `olares-cli router models` shows the "+
			"qualified names", err)
	case re.Type == "quota_exceeded_error":
		return fmt.Errorf("%w\n`olares-cli router quota list` shows the limits, and `router usage summary` "+
			"what has been spent against them", err)
	case re.Code == "stream_unsupported_on_fallback_provider":
		// Router only reaches its legacy single-upstream path when no provider
		// claims the model, and complains about streaming from there. The name
		// is what was actually wrong.
		return fmt.Errorf("%w\nNo provider serves this model, which is what put the request on Router's "+
			"fallback path. `olares-cli router models` shows the names it does serve", err)
	case strings.HasSuffix(re.Code, "_mode_mismatch"):
		return fmt.Errorf("%w\nThe model is configured for a different kind of work than this verb asks "+
			"for. `olares-cli router list` shows each model's mode, and a route can only serve the mode "+
			"it was created with", err)
	case strings.HasSuffix(re.Code, "_unsupported_for_provider") || re.Code == "audio_path_unsupported":
		return fmt.Errorf("%w\nThis model's provider does not serve that route at all. For a model "+
			"running on this Olares that usually means a different engine image does this job: "+
			"`olares-cli router list --mode audio` and `router provider get <provider>` show which "+
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
